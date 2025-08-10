package main

import (
	"log"
	"net"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind", err)
	}

	conn, err := l.Accept()
	if err != nil {
		log.Fatalln("Failed to accept new connection", err)
	}

	_, err = conn.Write([]byte("+PONG\r\n"))
	if err != nil {
		log.Fatalln("Failed to PONG", err)
	}
}
