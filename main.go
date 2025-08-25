package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Println("failed to open messages.txt: ", err)
		os.Exit(1)
	}

	defer file.Close()

	var curr []byte
	var chars []byte

	for {
		chars = make([]byte, 8)

		_, err := file.Read(chars)
		if err == io.EOF {
			break
		}

		parts := bytes.Split(chars, []byte{'\n'})

		curr = append(curr, parts[0]...)

		// will not handle lines shorter than 8 bytes
		if len(parts) == 2 {
			fmt.Printf("read: %s\n", curr)
			curr = parts[1]
		}
	}
}
