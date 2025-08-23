package main

import (
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/app"
)

func main() {
	app := app.NewApp()
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Failed to accept new connection", err)
		}

		go app.HandleConnection(conn)
	}
}
