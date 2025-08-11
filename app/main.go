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
		var data []byte
		var err error

		// a hack since a simple PING command changed from +PING to *1\r\n$4\r\nPING\r\n
		data, err = bufReader.ReadBytes('\n')
		log.Printf("READ 1: %d, `%s`, %s", len(data), string(data), err)
		data, err = bufReader.ReadBytes('\n')
		log.Printf("READ 2: %d, `%s`, %s", len(data), string(data), err)
		data, err = bufReader.ReadBytes('\n')
		log.Printf("READ 3: %d, `%s`, %s", len(data), string(data), err)

		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("Failed to read data", err)
			return
		}

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
