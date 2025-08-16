package encoding

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

func UnmarshalCommand(bufReader *bufio.Reader) (types.Command, error) {
	// TODO: move error to other step
	cmd, err := parseElement(bufReader)
	if err != nil {
		return types.Command{}, newUnmarshalError(cmd, "unmarshal", err)
	}
	return cmd, nil
}

func readUntilCRLF(bufReader *bufio.Reader) ([]byte, error) {
	data, err := bufReader.ReadBytes(CR)
	if err != nil {
		return nil, err
	}
	lastByte, err := bufReader.ReadByte()
	if err != nil {
		return nil, err
	}
	if lastByte != LF {
		return nil, fmt.Errorf("expected ending byte to be LF, got `%b`", lastByte)
	}
	return data[:len(data)-1], nil
}

func readNumUntilCRLF(bufReader *bufio.Reader) (int64, error) {
	sizeBytes, err := readUntilCRLF(bufReader)
	if err != nil {
		return 0, fmt.Errorf("read number bytes failed: %w", err)
	}
	sizeStr := string(sizeBytes)
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("cannot convert `%s` as number: %w", sizeStr, err)
	}
	return size, nil
}

func parseSymArray(bufReader *bufio.Reader) (types.Command, error) {
	size, err := readNumUntilCRLF(bufReader)
	if err != nil {
		return types.Command{}, fmt.Errorf("read array size failed: %w", err)
	}

	var cmd types.Command
	cmd.Sym = types.SymArray
	cmd.Array = make([]types.Command, 0, size)

	for i := range size {
		elemCmd, err := parseElement(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("parse element at position %d failed: %w", i, err)
		}
		cmd.Array = append(cmd.Array, elemCmd)
	}

	return cmd, nil
}

func readLengthAndStringUntilCRLF(bufReader *bufio.Reader) (string, error) {
	size, err := readNumUntilCRLF(bufReader)
	if err != nil {
		return "", fmt.Errorf("read string length failed: %w", err)
	}
	buffer := make([]byte, size)
	_, err = io.ReadFull(bufReader, buffer)
	if err != nil {
		return "", fmt.Errorf("read string data failed: %w", err)
	}
	crlf, err := readUntilCRLF(bufReader)
	if err != nil {
		return "", err
	}
	if len(crlf) != 0 {
		return "", fmt.Errorf("found more data at the end of bulk string")
	}
	return string(buffer), nil
}

func parseElement(bufReader *bufio.Reader) (types.Command, error) {
	symByte, err := bufReader.ReadByte()
	if err != nil {
		return types.Command{}, fmt.Errorf("failed to read symbol byte: %w", err)
	}

	sym := types.Sym(symByte)
	if !types.IsSymbolValid(sym) {
		return types.Command{}, fmt.Errorf("unknown symbol byte `%c`", symByte)
	}

	var cmd types.Command
	cmd.Sym = sym

	switch sym {
	case types.SymNull:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("read simple string value failed: %w", err)
		}
		if len(data) != 0 {
			return types.Command{}, fmt.Errorf("null cannot have other data")
		}
	case types.SymString:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("read simple string value failed: %w", err)
		}
		cmd.String = string(data)
	case types.SymError:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("read simple error value failed: %w", err)
		}
		cmd.Error = string(data)
	case types.SymBulkString:
		data, err := readLengthAndStringUntilCRLF(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("read bulk string value failed: %w", err)
		}
		cmd.BulkString = data
	case types.SymBulkError:
		data, err := readLengthAndStringUntilCRLF(bufReader)
		if err != nil {
			return types.Command{}, fmt.Errorf("read bulk string value failed: %w", err)
		}
		cmd.BulkError = data
	case types.SymArray:
		cmd, err = parseSymArray(bufReader)
		if err != nil {
			return types.Command{}, err
		}
	default:
		return types.Command{}, fmt.Errorf("symbol type `%c` is currently not supported", sym)
	}

	return cmd, nil
}
