package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func handleConnection(conn net.Conn) <-chan string {
	lines := make(chan string)

	go func() {
		defer conn.Close()
		defer close(lines)

		defer func() { fmt.Println("channel closed") }()

		var curr []byte
		var buff []byte

		for {
			buff = make([]byte, 8)

			n, err := conn.Read(buff)
			if err == io.EOF {
				break
			}

			buff = buff[:n]

			parts := bytes.Split(buff, []byte{'\n'})

			curr = append(curr, parts[0]...)

			if len(parts) == 2 {
				lines <- string(curr)
				curr = parts[1]
			}
		}
	}()

	return lines
}

func main() {
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

		for line := range handleConnection(conn) {
			fmt.Printf("read: %s\n", line)
		}
	}

}
