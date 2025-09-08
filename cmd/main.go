package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yus-works/tcp-to-http/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port)
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

