package cmd

type PING struct {
	Message string `arg:"pos:1,default:PONG,optional"`
}

type ECHO struct {
	Message string `arg:"pos:1"`
}
