package cmd

type LLEN struct {
	Key string `arg:"pos:1"`
}

type RPUSH struct {
	Key    string   `arg:"pos:1"`
	Values []string `arg:"pos:2,variadic"`
}

type LPUSH struct {
	Key    string   `arg:"pos:1"`
	Values []string `arg:"pos:2,variadic"`
}

type LRANGE struct {
	Key   string `arg:"pos:1"`
	Start int    `arg:"pos:2"`
	Stop  int    `arg:"pos:3"`
}

type LPOP struct {
	Key   string `arg:"pos:1"`
	Count *int   `arg:"pos:2,optional"`
}

type BLPOP struct {
	Key           string   `arg:"pos:1"`
	KeyRest       []string `arg:"pos:2,variadic"`
	TimeoutSecond float64  `arg:"pos:3"`
}
