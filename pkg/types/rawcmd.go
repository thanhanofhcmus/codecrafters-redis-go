package types

type RawCmd struct {
	Sym Sym

	Integer    int64
	Boolean    bool
	Double     float64
	String     string
	Error      string
	BulkString string
	BulkError  string

	Array []RawCmd

	Map       map[*RawCmd]RawCmd
	Attribute map[*RawCmd]RawCmd
	Set       map[*RawCmd]bool

	// TODO: VerbatimStrings, Pushes, BigNumber
}

func NewNullRawCmd() RawCmd {
	return RawCmd{
		Sym: SymNull,
	}
}

func NewStringRawCmd(value string) RawCmd {
	return RawCmd{
		Sym:    SymString,
		String: value,
	}
}

func NewBulkStringRawCmd(value string) RawCmd {
	return RawCmd{
		Sym:        SymBulkString,
		BulkString: value,
	}
}

func NewErrorRawCmd(value string) RawCmd {
	return RawCmd{
		Sym:   SymError,
		Error: value,
	}
}

func NewIntegerRawCmd(value int64) RawCmd {
	return RawCmd{
		Sym:     SymInteger,
		Integer: value,
	}
}

func NewBulkArrayBulkString(values []string) RawCmd {
	array := make([]RawCmd, 0, len(values))

	for _, value := range values {
		array = append(array, RawCmd{
			Sym:        SymBulkString,
			BulkString: value,
		})
	}

	return RawCmd{
		Sym:   SymArray,
		Array: array,
	}
}
