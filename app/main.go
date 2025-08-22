package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/app"
	"github.com/codecrafters-io/redis-starter-go/internal/encoding"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func handleConn(app *app.App, conn net.Conn) {
	// TODO: timeout with SetWriteDeadline
	bufReader := bufio.NewReader(conn)
	for {
		var err error

		defer func() {
			// TODO: only write if not a connection error
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-%s\r\n", err.Error())
			}
		}()

		var res types.RawCmd

		res, err = encoding.UnmarshalCommand(bufReader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				log.Println("Failed to unmarshal data:", err)
			}
			return
		}

		ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*30)
		defer cancelFn()

		resp, err := app.HandleCommand(ctx, res)
		if err != nil {
			return
		}

		respByte, err := encoding.MarshalCommand(resp)
		if err != nil {
			return
		}

		// TODO: timeout with SetWriteDeadline
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
