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
	"github.com/codecrafters-io/redis-starter-go/pkg/ulid"
)

type ctxKey int

const idKey ctxKey = 1

func NewContext(ctx context.Context, id ulid.ID) context.Context {
	return context.WithValue(ctx, idKey, id)
}

func GetIdFromContext(ctx context.Context) ulid.ID {
	id, _ := ctx.Value(idKey).(ulid.ID)
	return id
}

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
		connId := app.idGenerator.MustNew()
		ctx = NewContext(ctx, connId)

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
