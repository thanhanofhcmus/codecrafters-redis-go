package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

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
		return nil, fmt.Errorf("expected ending byte to be LF, got `%c`", lastByte)
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

func parseSymArray(bufReader *bufio.Reader) (Cmd, error) {
	size, err := readNumUntilCRLF(bufReader)
	if err != nil {
		return Cmd{}, fmt.Errorf("read array size failed: %w", err)
	}

	var cmd Cmd
	cmd.Sym = SymArray
	cmd.Array = make([]Cmd, 0, size)

	for i := range size {
		elemCmd, err := parseElement(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("parse element at position %d failed: %w", i, err)
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

func parseElement(bufReader *bufio.Reader) (Cmd, error) {
	symByte, err := bufReader.ReadByte()
	if err != nil {
		return Cmd{}, fmt.Errorf("failed to read symbol byte: %w", err)
	}

	sym := Sym(symByte)
	if _, exists := ALL_SYMS_SET[sym]; !exists {
		return Cmd{}, fmt.Errorf("unknown symbol byte `%c`", symByte)
	}

	var cmd Cmd
	cmd.Sym = sym

	switch sym {
	case SymNull:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("read simple string value failed: %w", err)
		}
		if len(data) != 0 {
			return Cmd{}, fmt.Errorf("null cannot have other data")
		}
	case SymString:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("read simple string value failed: %w", err)
		}
		cmd.String = string(data)
	case SymError:
		data, err := readUntilCRLF(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("read simple error value failed: %w", err)
		}
		cmd.Error = string(data)
	case SymBulkString:
		data, err := readLengthAndStringUntilCRLF(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("read bulk string value failed: %w", err)
		}
		cmd.BulkString = data
	case SymBulkError:
		data, err := readLengthAndStringUntilCRLF(bufReader)
		if err != nil {
			return Cmd{}, fmt.Errorf("read bulk string value failed: %w", err)
		}
		cmd.BulkError = data
	case SymArray:
		cmd, err = parseSymArray(bufReader)
		if err != nil {
			return Cmd{}, err
		}
	default:
		return Cmd{}, fmt.Errorf("symbol type `%c` is currently not supported", sym)
	}

	return cmd, nil
}

func readAndParseCommand(bufReader *bufio.Reader) (Cmd, error) {
	cmd, err := parseElement(bufReader)
	if err != nil && errors.Is(err, io.EOF) {
		err = nil
	}
	return cmd, err
}
