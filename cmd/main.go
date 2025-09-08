package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yus-works/tcp-to-http/internal/request"
	"github.com/yus-works/tcp-to-http/internal/response"
	"github.com/yus-works/tcp-to-http/internal/server"
)

const port = 42069

func main() {
	frfrHandler := func(
		w io.Writer, req *request.Request,
	) *response.HandlerError {
		if req.RequestLine.RequestTarget == "/yourproblem" {
			err := response.NewHandlerErr(response.StatusBadRequest)
			return &err
		}

		if req.RequestLine.RequestTarget == "/myproblem" {
			err := response.NewHandlerErr(response.StatusInternalServerError)
			return &err
		}

		fmt.Fprint(w, "all good frfr\n")
		return nil
	}

	server, err := server.Serve(port, frfrHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)

	// if sigint or sigterm detected, relay it to sigChan
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan // block until sigChan has something to produce
	log.Println("Server gracefully stopped")
}

