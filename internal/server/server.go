package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/yus-works/tcp-to-http/internal/request"
	"github.com/yus-works/tcp-to-http/internal/response"
)

type Server struct {
	port int

	closed   atomic.Bool
	listener net.Listener
	handler response.Handler
}

func Serve(port int, handler response.Handler) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("Error starting server on %d: %w\n", port, err)
	}

	s := Server{
		port:     port,
		listener: ln,
		handler: handler,
	}

	go s.listen()

	return &s, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println("Error accepting connection: ", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Failed to parse/read request: ", err)

		response.WriteError(conn, response.StatusBadRequest)
		return
	}

	buf := bytes.Buffer{}

	handlerErr := s.handler(&buf, req)
	if handlerErr != nil {
		response.WriteError(conn, handlerErr.StatusCode)
	}

	msg := buf.Bytes()

	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(len(msg)))
	conn.Write(msg)
}
