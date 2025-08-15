package main

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

const CR byte = '\r'
const LF byte = '\n'

var CRLF []byte = []byte("\r\n")
