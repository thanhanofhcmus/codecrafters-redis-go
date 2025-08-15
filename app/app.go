package main

import "fmt"

type App struct {
	m map[string]string
}

func NewApp() *App {
	return &App{
		m: map[string]string{},
	}
}

func (app *App) HandleCommand(cmd Cmd) (Cmd, error) {
	if cmd.Sym != SymArray {
		return Cmd{}, fmt.Errorf("cannot handle command that is not of array type")
	}
	args := cmd.Array
	if len(args) == 0 {
		return Cmd{}, fmt.Errorf("command cannot have zero size")
	}
	fArg := args[0]
	if fArg.Sym != SymBulkString {
		return Cmd{}, fmt.Errorf("first argument is not a BulkString")
	}

	var result Cmd

	switch fArg.BulkString {
	case "PING":
		result.Sym = SymString
		result.String = "PONG"
	case "ECHO":
		result.Sym = SymBulkString
		if len(args) >= 2 {
			result.BulkString = args[1].BulkString
		}
	default:
		return Cmd{}, fmt.Errorf("unknown command `%s`", fArg.BulkString)
	}

	return result, nil
}
