package udpsender

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func UDPdemo() {
	raddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Println("Error resolving address: ", err)
		os.Exit(1)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Println("Error dialing udp connection: ", err)
		os.Exit(1)
	}

	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")

		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println("Error reading from stdio: ", err)
		}

		_, err = conn.Write(line)
		if err != nil {
			log.Println("Error writing to udp conn: ", err)
		}
	}
}
