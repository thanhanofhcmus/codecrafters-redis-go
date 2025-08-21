package encoding

import (
	"bytes"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func MarshalCommand(cmd types.RawCmd) ([]byte, error) {
	var buffer bytes.Buffer
	err := buildRESPBytes(cmd, &buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func buildRESPBytes(cmd types.RawCmd, buffer *bytes.Buffer) error {
	newErr := func(step string, inner error) *EncodingError {
		return newMarshalError(cmd, step, inner)
	}

	var err error

	// Trick to handle Null on RESP2
	if cmd.Sym == types.SymNull {
		err = buffer.WriteByte(byte(types.SymBulkString))
		if err != nil {
			return newErr("write symbol", err)
		}
		if _, err = fmt.Fprint(buffer, -1); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		return nil
	}

	err = buffer.WriteByte(byte(cmd.Sym))
	if err != nil {
		return newErr("write symbol", err)
	}

	switch cmd.Sym {
	case types.SymNull:
		// RESP3: Do nothing
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write CRLF", err)
		}
	case types.SymInteger:
		if _, err = fmt.Fprint(buffer, cmd.Integer); err != nil {
			return newErr("write integer", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write CRLF", err)
		}
	case types.SymString:
		if _, err = buffer.WriteString(cmd.String); err != nil {
			return newErr("write string", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write CRLF", err)
		}
	case types.SymError:
		if _, err = buffer.WriteString(cmd.Error); err != nil {
			return newErr("write error", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write CRLF", err)
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
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write CRLF", err)
		}
	case types.SymArray:
		if _, err = fmt.Fprint(buffer, len(cmd.Array)); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		for _, elem := range cmd.Array {
			if err = buildRESPBytes(elem, buffer); err != nil {
				return newErr("write element", err)
			}
		}
	default:
		panic(fmt.Sprintf("unknown symbol type %c", cmd.Sym))
	}

	return nil
}
