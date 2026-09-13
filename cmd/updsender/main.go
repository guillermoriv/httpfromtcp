package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const port = ":42069"

func main() {
	address, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatalf("error resolving for UDP address: %s\n", err.Error())
	}

	conn, err := net.DialUDP("udp", nil, address)
	if err != nil {
		log.Fatalf("connection for UDP failed: %s\n", err.Error())
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error while parsing the line: %s", err.Error())
			continue
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Printf("error sending bytes through the conn: %s", err.Error())
			continue
		}
	}
}
