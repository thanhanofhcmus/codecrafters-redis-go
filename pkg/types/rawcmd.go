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
