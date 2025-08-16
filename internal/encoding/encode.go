package encoding

import (
	"bytes"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func MarshalCommand(cmd types.Command) ([]byte, error) {
	var buffer bytes.Buffer
	err := buildRESPBytes(cmd, &buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildRESPBytes(cmd types.Command, buffer *bytes.Buffer) *EncodingError {
	newErr := func(step string, inner error) *EncodingError {
		return newMarshalError(cmd, step, inner)
	}

	err := buffer.WriteByte(byte(cmd.Sym))
	if err != nil {
		return newErr("write symbol", err)
	}

	switch cmd.Sym {
	case types.SymNull:
		// Do nothing
	case types.SymString:
		if _, err = buffer.WriteString(cmd.String); err != nil {
			return newErr("write string", err)
		}
	case types.SymError:
		if _, err = buffer.WriteString(cmd.Error); err != nil {
			return newErr("write error", err)
		}
	case types.SymBulkString:
		if _, err = fmt.Fprint(buffer, len(cmd.BulkString)); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		if _, err = buffer.WriteString(cmd.BulkString); err != nil {
			return newErr("write string", err)
		}
	case types.SymArray:
		if _, err = fmt.Fprint(buffer, len(cmd.Array)); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		for _, elem := range cmd.Array {
			// TODO: fix this
			if err = buildRESPBytes(elem, buffer); err != nil {
				return newErr("write element", err)
			}
		}
	default:
		panic(fmt.Sprintf("unknown symbol type %c", cmd.Sym))
	}

	if _, err = buffer.Write(CRLF); err != nil {
		return newErr("write CRLF", err)
	}

	return nil
}
