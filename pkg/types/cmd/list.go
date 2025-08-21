package cmd

type RPUSH struct {
	Key    string   `arg:"pos:1"`
	Values []string `arg:"pos:2,variadic"`
}

type LRANGE struct {
	Key   string `arg:"pos:1"`
	Start int    `arg:"pos:2"`
	Stop  int    `arg:"pos:3"`
}
