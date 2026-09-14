package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/guillermoriv/httpfromtcp/internal/headers"
	"github.com/guillermoriv/httpfromtcp/internal/request"
	"github.com/guillermoriv/httpfromtcp/internal/response"
	"github.com/guillermoriv/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHandler(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/video" {
		handlerVideo(w, req)
		return
	}

	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	handler200(w, req)
}

func handlerVideo(w *response.Writer, req *request.Request) {
	w.WriteStatusLine(response.StatusOK)

	data, err := os.ReadFile("./assets/vim.mp4") // we just dump everything to memory for he sake of simplicity
	if err != nil {
		handler500(w, req)
		return
	}

	h := response.GetDefaultHeaders(len(data))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)
	w.WriteBody(data)
}

func proxyHandler(w *response.Writer, req *request.Request) {
	target := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	url := "https://httpbingo.org/" + target
	fmt.Println("Proxying to", url)
	res, err := http.Get(url)
	if err != nil {
		handler500(w, req)
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-Sha256, X-Content-Length")
	h.Delete("Content-Length") // so we can add Transfer-Encoding: chunked
	w.WriteHeaders(h)

	fullBody := make([]byte, 0)

	const maxChunkSize = 1024
	buffer := make([]byte, maxChunkSize)
	for {
		n, err := res.Body.Read(buffer)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buffer[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			fullBody = append(fullBody, buffer[:n]...)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			fmt.Println("Error reading response body:", err)
			break
		}
	}

	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	trailers := headers.NewHeaders()
	sha256 := fmt.Sprintf("%x", sha256.Sum256(fullBody))
	trailers.Override("X-Content-Sha256", sha256)
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(fullBody)))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Println("Error writing trailers:", err)
	}
	fmt.Println("Wrote trailers")
}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusBadRequest)
	body := []byte(`
		<html>
			<head>
				<title>400 Bad Request</title>
			</head>
			<body>
				<h1>Bad Request</h1>
				<p>Your request honestly kinda sucked.</p>
			</body>
		</html>
	`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusServerError)
	body := []byte(`
		<html>
			<head>
				<title>500 Internal Server Error</title>
			</head>
			<body>
				<h1>Internal Server Error</h1>
				<p>Okay, you know what? This one is on me.</p>
			</body>
		</html>
	`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusOK)
	body := []byte(`
		<html>
			<head>
				<title>200 OK</title>
			</head>
			<body>
				<h1>Success!</h1>
				<p>Your request was an absolute banger.</p>
			</body>
		</html>
	`)
	h := response.GetDefaultHeaders(len(body))
	h.Override("Content-Type", "text/html")
	w.WriteHeaders(h)
	w.WriteBody(body)
}
