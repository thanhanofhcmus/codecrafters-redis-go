package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
)

type Sym = rune

const (
	// Simple
	SymString    Sym = '+'
	SymError     Sym = '-'
	SymInteger   Sym = ':'
	SymNull      Sym = '_'
	SymBoolean   Sym = '#'
	SymDouble    Sym = ','
	SymBigNumber Sym = '('

	// Aggregate like
	SymBulkString Sym = '$'
	SymBulkError  Sym = '!'

	// Aggregate
	SymArray          Sym = '*'
	SymVerbatimString Sym = '='
	SymMap            Sym = '%'
	SymAttribute      Sym = '|'
	SymSet            Sym = '~'
	SymPush           Sym = '>'
)

var ALL_SYMS_SET = map[Sym]bool{
	SymString:         true,
	SymError:          true,
	SymInteger:        true,
	SymNull:           true,
	SymBoolean:        true,
	SymDouble:         true,
	SymBigNumber:      true,
	SymBulkString:     true,
	SymArray:          true,
	SymBulkError:      true,
	SymVerbatimString: true,
	SymMap:            true,
	SymAttribute:      true,
	SymSet:            true,
	SymPush:           true,
}

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

const CR byte = '\r'
const LF byte = '\n'

var CRLF []byte = []byte("\r\n")

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

func handleCmd(cmd Cmd) (Cmd, error) {
	if cmd.Sym != SymArray {
		return Cmd{}, fmt.Errorf("cannot handle command that is not of array type")
	}
	args := cmd.Array
	if len(args) == 0 {
		return Cmd{}, fmt.Errorf("command cannot have zero size")
	}
	fArg := args[0]
	if fArg.Sym != SymBulkString {
		return Cmd{}, fmt.Errorf("first argument is not a BulkString")
	}

	var result Cmd

	switch fArg.BulkString {
	case "PING":
		result.Sym = SymString
		result.String = "PONG"
	case "ECHO":
		result.Sym = SymBulkString
		if len(args) >= 2 {
			result.BulkString = args[1].BulkString
		}
	default:
		return Cmd{}, fmt.Errorf("unknown command `%s`", fArg.BulkString)
	}

	return result, nil
}

func handleConn(conn net.Conn) {
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

		resp, err := handleCmd(res)
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
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind", err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("Failed to accept new connection", err)
		}

		go handleConn(conn)
	}
}
