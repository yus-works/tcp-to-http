package server

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
		defer func() {
			conn.Close()
			close(lines)

			fmt.Println("channel closed")
		}()

		var curr []byte
		var buff []byte

		for {
			buff = make([]byte, 8)

			n, err := conn.Read(buff)
			if n > 0 {
				buff = buff[:n]

				parts := bytes.Split(buff, []byte{'\n'})

				curr = append(curr, parts[0]...)

				if len(parts) > 1 {
					lines <- string(curr)

					curr = parts[1]
				}
			}

			// must handle EOF after bytes because Read
			// might return EOF with the final bytes ???
			if err != nil {
				if err != io.EOF {
					log.Println("Error reading bytes: ", err)
				}

				// flush whatever is left in curr
				lines <- string(curr)
				break
			}
		}
	}()

	return lines
}

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

		for line := range handleConnection(conn) {
			fmt.Printf("read: %s\n", line)
		}
	}
}
