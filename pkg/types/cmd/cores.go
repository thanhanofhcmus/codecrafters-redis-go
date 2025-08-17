package cmd

type PING struct {
	Message string `arg:"pos:1,default:PONG,optional"`
}

type ECHO struct {
	Message string `arg:"pos:1"`
}

type SET struct {
	Key   string `arg:"pos:1"`
	Value string `arg:"pos:2"`

	GET bool

	SetKey struct {
		Key string `arg:"enum-key"`
		NX  bool
		XX  bool
	} `arg:"enum"`

	Expire struct {
		Key     string `arg:"enum-key"`
		EX      int
		PX      int
		EXAT    int
		PXAT    int
		KEEPTTL bool
	} `arg:"enum"`
}

type GET struct {
	Key string `arg:"pos:1"`
}
