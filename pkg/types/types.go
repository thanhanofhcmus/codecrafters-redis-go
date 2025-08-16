package types

import "fmt"

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

var symbolStrings = map[Sym]string{
	SymString:         "String",
	SymError:          "Error",
	SymInteger:        "Integer",
	SymNull:           "Null",
	SymBoolean:        "Boolean",
	SymDouble:         "Double",
	SymBigNumber:      "BigNumber",
	SymBulkString:     "BulkString",
	SymArray:          "Array",
	SymBulkError:      "BulkError",
	SymVerbatimString: "VerbatimString",
	SymMap:            "Map",
	SymAttribute:      "Attribute",
	SymSet:            "Set",
	SymPush:           "Push",
}

func IsSymbolValid(sym Sym) bool {
	_, exists := symbolStrings[sym]
	return exists
}

func GetSymString(sym Sym) string {
	if value, exists := symbolStrings[sym]; exists {
		return value
	}
	return fmt.Sprintf("<unknown|%b>", sym)
}

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

func NewNullCommand() Command {
	return Command{
		Sym: SymNull,
	}
}

func NewStringCommand(value string) Command {
	return Command{
		Sym:    SymString,
		String: value,
	}
}

func NewBulkStringCommand(value string) Command {
	return Command{
		Sym:        SymBulkString,
		BulkString: value,
	}
}

func NewErrorCommand(value string) Command {
	return Command{
		Sym:   SymError,
		Error: value,
	}
}
