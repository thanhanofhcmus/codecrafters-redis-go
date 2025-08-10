package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind", err)
		os.Exit(0)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Failed to accept conn", err)
		os.Exit(0)
	}

	_ = conn

}
