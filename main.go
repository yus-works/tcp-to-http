package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		log.Println("failed to open messages.txt: ", err)
		os.Exit(1)
	}

	defer file.Close()

	var data []byte

	for {
		chars := make([]byte, 8)
		_, err := file.Read(chars)
		if err == io.EOF {
			break
		}

		data = append(data, chars...)
	}

	lines := strings.SplitSeq(string(data), "\n")

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
