package cmd

type RPUSH struct {
	Key    string   `arg:"pos:1"`
	Values []string `arg:"pos:2,variadic"`
}
