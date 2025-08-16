package main

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

type App struct {
	m map[string]string
}

func convertArgsCmdToString(cmd types.Command) ([]string, error) {
	if cmd.Sym != types.SymArray {
		return nil, NewInvalidTypeError(types.SymArray, cmd.Sym)
	}
	args := cmd.Array
	if len(args) == 0 {
		return nil, fmt.Errorf("command cannot have zero argument")
	}
	result := make([]string, 0, len(args))
	for idx, arg := range args {
		if arg.Sym != types.SymBulkString {
			return nil, NewArrayElementError(idx, NewInvalidTypeError(types.SymBulkString, arg.Sym))
		}
		result = append(result, arg.BulkString)
	}
	return result, nil
}

func NewApp() *App {
	return &App{
		m: map[string]string{},
	}
}

func (app *App) HandleCommand(cmd types.Command) (result types.Command, err error) {
	args, err := convertArgsCmdToString(cmd)
	if err != nil {
		return types.Command{}, err
	}

	switch strings.ToUpper(args[0]) {
	case "PING":
		result.Sym = types.SymString
		result.String = "PONG"
	case "ECHO":
		result.Sym = types.SymBulkString
		if len(args) >= 2 {
			result.BulkString = args[1]
		}
	case "SET":
		if len(args) < 3 {
			return types.Command{}, fmt.Errorf("invalid number of arguments")
		}
		app.m[args[1]] = args[2]
		result.Sym = types.SymString
		result.String = "OK"
	case "GET":
		if len(args) != 2 {
			return types.Command{}, fmt.Errorf("invalid number of arguments")
		}
		if value, exists := app.m[args[1]]; exists {
			result.Sym = types.SymBulkString
			result.BulkString = value
		} else {
			result.Sym = types.SymNull
		}
	default:
		return types.Command{}, fmt.Errorf("unknown command `%s`", args[0])
	}

	return result, nil
}
