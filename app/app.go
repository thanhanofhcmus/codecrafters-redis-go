package main

import "fmt"

type App struct {
	m map[string]string
}

func convertArgsCmdToString(cmd Command) ([]string, error) {
	if cmd.Sym != SymArray {
		return nil, fmt.Errorf("expect command type to be Array, got %c", cmd.Sym)
	}
	args := cmd.Array
	if len(args) == 0 {
		return nil, fmt.Errorf("command cannot have zero argument")
	}
	result := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.Sym != SymBulkString {
			return nil, fmt.Errorf("expect argument to be of the type BulkString, got %c", arg.Sym)
		}
		if len(arg.BulkString) == 0 {
			return nil, fmt.Errorf("argument is empty")
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

func (app *App) HandleCommand(cmd Command) (Command, error) {
	args, err := convertArgsCmdToString(cmd)
	if err != nil {
		return Command{}, err
	}

	var result Command

	switch args[0] {
	case "PING":
		result.Sym = SymString
		result.String = "PONG"
	case "ECHO":
		result.Sym = SymBulkString
		if len(args) >= 2 {
			result.BulkString = args[1]
		}
	case "SET":
		if len(args) < 3 {
			return Command{}, fmt.Errorf("invalid number of arguments")
		}
		app.m[args[1]] = args[2]
		result.Sym = SymString
		result.String = "OK"
	case "GET":
		if len(args) != 2 {
			return Command{}, fmt.Errorf("invalid number of arguments")
		}
		if value, exists := app.m[args[1]]; exists {
			result.Sym = SymBulkString
			result.BulkString = value
		} else {
			result.Sym = SymNull
		}
	default:
		return Command{}, fmt.Errorf("unknown command `%s`", args[0])
	}

	return result, nil
}
