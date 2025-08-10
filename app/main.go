package main

import (
	"bufio"
	"io"
	"log"
	"net"
)

func handleConn(conn net.Conn) {
	bufReader := bufio.NewReader(conn)
	for {
		data, err := bufReader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			log.Println("Failed to read data", err)
			return
		}

		log.Println("READ: ", string(data))

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			log.Println("Failed to PONG", err)
			return
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Failed to accept new connection", err)
		}

		go handleConn(conn)
	}
}
