package server

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/yus-works/tcp-to-http/internal/request"
)

func Start() {
	ln, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Println("Error listening: ", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Server listening on :42069")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting: ", err)
		}

		fmt.Println("Connection has been accepted")

		rq, err := request.RequestFromReader(conn)
		if err != nil {
			log.Println("Failed to read from conn: ", err)
			os.Exit(1)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %s\n", rq.RequestLine.Method)
		fmt.Printf("- Target: %s\n", rq.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", rq.RequestLine.HttpVersion)

		fmt.Printf("Headers:\n")
		for k, v := range *rq.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		
		fmt.Printf("Body:\n")
		fmt.Print(string(rq.Body))
	}
}
