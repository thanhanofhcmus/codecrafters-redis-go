package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/app"
	"github.com/codecrafters-io/redis-starter-go/internal/encoding"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func handleConn(app *app.App, conn net.Conn) {
	bufReader := bufio.NewReader(conn)
	for {
		var err error

		defer func() {
			// TODO: only write if not a connection error
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-%s\r\n", err.Error())
			}
		}()

		var res types.Command

		res, err = encoding.UnmarshalCommand(bufReader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				log.Println("Failed to unmarshal data:", err)
			}
			return
		}

		resp, err := app.HandleCommand(res)
		if err != nil {
			return
		}

		respByte, err := encoding.MarshalCommand(resp)
		if err != nil {
			return
		}

		_, err = conn.Write(respByte)
		if err != nil {
			log.Println("Failed to response", err)
			return
		}
	}
}

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

		go handleConn(app, conn)
	}
}
