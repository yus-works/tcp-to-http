package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer f.Close()
		defer close(lines)

		var curr []byte
		var buff []byte

		for {
			buff = make([]byte, 8)

			_, err := f.Read(buff)
			if err == io.EOF {
				break
			}

			parts := bytes.Split(buff, []byte{'\n'})

			curr = append(curr, parts[0]...)

			if len(parts) == 2 {
				lines<-string(curr)
				curr = parts[1]
			}
		}
	}()

	return lines
}

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Println("failed to open messages.txt: ", err)
		os.Exit(1)
	}

	for line := range getLinesChannel(file) {
		fmt.Printf("read: %s\n", line)
	}
}
