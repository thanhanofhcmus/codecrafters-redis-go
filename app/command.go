package main

import (
	"bytes"
	"errors"
	"fmt"
)

type Command struct {
	Sym Sym

	Integer    int64
	Boolean    bool
	Double     float64
	String     string
	Error      string
	BulkString string
	BulkError  string

	Array []Command

	Map       map[*Command]Command
	Attribute map[*Command]Command
	Set       map[*Command]bool

	// TODO: VerbatimStrings, Pushes, BigNumber
}

type buildRESPError struct {
	cmd   Command
	step  string
	inner error
}

func (err *buildRESPError) Error() string {
	return fmt.Sprintf("[%c|%s]: %s", err.cmd.Sym, err.step, err.inner)
}

func (err *buildRESPError) Is(target error) bool {
	return errors.Is(err.inner, target)
}

func (cmd Command) buildRESPBytes(buffer *bytes.Buffer) error {
	newErr := func(step string, inner error) error {
		return &buildRESPError{
			cmd:   cmd,
			step:  step,
			inner: inner,
		}
	}

	err := buffer.WriteByte(byte(cmd.Sym))
	if err != nil {
		return newErr("write symbol", err)
	}

	switch cmd.Sym {
	case SymNull:
		// Do nothing
	case SymString:
		if _, err = buffer.WriteString(cmd.String); err != nil {
			return newErr("write string", err)
		}
	case SymError:
		if _, err = buffer.WriteString(cmd.Error); err != nil {
			return newErr("write error", err)
		}
	case SymBulkString:
		if _, err = fmt.Fprint(buffer, len(cmd.BulkString)); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		if _, err = buffer.WriteString(cmd.BulkString); err != nil {
			return newErr("write string", err)
		}
	case SymArray:
		if _, err = fmt.Fprint(buffer, len(cmd.Array)); err != nil {
			return newErr("write length", err)
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return newErr("write length CRLF", err)
		}
		for _, elem := range cmd.Array {
			if err = elem.buildRESPBytes(buffer); err != nil {
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

func (cmd Command) ToRESPBytes() ([]byte, error) {
	var buffer bytes.Buffer
	err := cmd.buildRESPBytes(&buffer)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
