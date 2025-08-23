package cmd

type APPEND struct {
	Key   string `arg:"pos:1"`
	Value string `arg:"pos:2"`
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
