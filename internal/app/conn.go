package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/encoding"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func (app *App) HandleConnection(conn net.Conn) {
	// TODO: timeout with SetWriteDeadline
	bufReader := bufio.NewReader(conn)
	for {
		var err error
		var res types.RawCmd

		defer func() {
			// TODO: only write if not a connection error
			if err != nil {
				_, _ = fmt.Fprintf(conn, "-%s\r\n", err.Error())
			}
		}()

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
