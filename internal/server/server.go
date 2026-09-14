package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/guillermoriv/httpfromtcp/internal/request"
	"github.com/guillermoriv/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	headers := response.GetDefaultHeaders(len(he.Message))
	response.WriteHeaders(w, headers)
	w.Write([]byte(he.Message)) // writting the body back for the message
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	srv := &Server{
		listener: listener,
		handler:  handler,
	}
	go srv.listen()
	return srv, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		hErr := &HandlerError{
			StatusCode: response.StatusBadRequest,
			Message:    err.Error(),
		}
		hErr.Write(conn)
		return
	}

	buf := bytes.NewBuffer([]byte{})

	hErr := s.handler(buf, req)
	if hErr != nil {
		hErr.Write(conn)
		return
	}

	b := buf.Bytes()
	response.WriteStatusLine(conn, response.StatusOK)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
}
