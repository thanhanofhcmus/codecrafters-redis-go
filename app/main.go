package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func handleConn(app *App, conn net.Conn) {
	bufReader := bufio.NewReader(conn)
	for {
		var err error

		defer func() {
			// TODO: only write if not a connection error
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-%s\r\n", err.Error())
			}
		}()

		var res Cmd

		res, err = readAndParseCommand(bufReader)
		if err != nil {
			log.Println("Failed to read and parse data:", err)
			return
		}

		resp, err := app.HandleCommand(res)
		if err != nil {
			return
		}

		respByte, err := resp.ToRESPBytes()
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
	app := NewApp()
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
