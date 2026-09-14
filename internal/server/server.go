package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/guillermoriv/httpfromtcp/internal/request"
	"github.com/guillermoriv/httpfromtcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)

// Server is an HTTP 1.1 server
type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
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
	w := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		w.WriteStatusLine(response.StatusBadRequest)
		body := fmt.Appendf(nil, "Error parsing request: %v", err)
		w.WriteHeaders(response.GetDefaultHeaders(len(body)))
		w.WriteBody(body)
		return
	}
	s.handler(w, req)
}
