package server

import (
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
}

func Serve(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Printf("Error starting listener on %d: %s\n", port, err)
		return nil, err
	}

	s := Server{
		port:     port,
		listener: ln,
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

	_, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("Failed to parse/read request: ", err)

        conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
        return
	}

	response.WriteStatusLine(conn, response.StatusOK)
	response.WriteHeaders(conn, response.GetDefaultHeaders(0))
}
