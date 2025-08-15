package main

import (
	"bytes"
	"fmt"
)

type Cmd struct {
	Sym Sym

	Integer    int64
	Boolean    bool
	Double     float64
	String     string
	Error      string
	BulkString string
	BulkError  string

	Array []Cmd

	Map       map[*Cmd]Cmd
	Attribute map[*Cmd]Cmd
	Set       map[*Cmd]bool

	// TODO: VerbatimStrings, Pushes, BigNumber
}

func (cmd Cmd) buildRESPBytes(buffer *bytes.Buffer) error {
	err := buffer.WriteByte(byte(cmd.Sym))
	if err != nil {
		return err
	}

	switch cmd.Sym {
	case SymNull:
		// Do nothing
	case SymString:
		if _, err = buffer.WriteString(cmd.String); err != nil {
			return err
		}
	case SymError:
		if _, err = buffer.WriteString(cmd.Error); err != nil {
			return err
		}
	case SymBulkString:
		if _, err = fmt.Fprint(buffer, len(cmd.BulkString)); err != nil {
			return err
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return err
		}
		if _, err = buffer.WriteString(cmd.BulkString); err != nil {
			return err
		}
	case SymArray:
		if _, err = fmt.Fprint(buffer, len(cmd.Array)); err != nil {
			return err
		}
		if _, err = buffer.Write(CRLF); err != nil {
			return err
		}
		for _, elem := range cmd.Array {
			if err = elem.buildRESPBytes(buffer); err != nil {
				return err
			}
		}
	default:
		panic(fmt.Sprintf("unknown symbol type %c", cmd.Sym))
	}

	_, err = buffer.Write(CRLF)
	if err != nil {
		return err
	}

	return nil
}

func (cmd Cmd) ToRESPBytes() ([]byte, error) {
	var buffer bytes.Buffer
	err := cmd.buildRESPBytes(&buffer)
	return buffer.Bytes(), err
}
