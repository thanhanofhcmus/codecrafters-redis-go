package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

type value struct {
	value        string
	shouldExpire bool
	expireAt     time.Time
}

type App struct {
	m map[string]value
}

func convertArgsCmdToString(cmd types.Command) ([]string, error) {
	if cmd.Sym != types.SymArray {
		return nil, NewInvalidTypeError(types.SymArray, cmd.Sym)
	}
	args := cmd.Array
	if len(args) == 0 {
		return nil, NewExpectArgumentError("<command>")
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
		m: map[string]value{},
	}
}

func (app *App) HandleCommand(cmd types.Command) (result types.Command, err error) {
	args, err := convertArgsCmdToString(cmd)
	if err != nil {
		return types.Command{}, err
	}

	command := args[0]

	switch strings.ToUpper(command) {
	case "PING":
		result = types.NewStringCommand("PONG")
	case "ECHO":
		if len(args) < 2 {
			err = NewExpectArgumentError("ECHO")
			break
		}
		result = types.NewBulkStringCommand(args[1])
	case "SET":
		result, err = app.handleSET(args, cmd)
	case "GET":
		result, err = app.handleGET(args, cmd)
	default:
		err = fmt.Errorf("unknown command `%s`", command)
	}

	if err != nil {
		err = NewHandleCommandError(command, err)
	}

	return
}

func (app *App) handleSET(args []string, _ types.Command) (types.Command, error) {
	argsLen := len(args)
	if argsLen != 3 && argsLen != 5 {
		return types.Command{}, fmt.Errorf("must have at least 3 arguments or 5 with PX")
	}
	value := value{
		value: args[2],
	}

	if argsLen == 3 {
		app.m[args[1]] = value
		return types.NewStringCommand("OK"), nil
	}

	// args length now is 5

	if strings.ToUpper(args[3]) != "PX" {
		return types.Command{}, NewInvalidOptionError(args[3])
	}

	duration, parseErr := strconv.Atoi(args[4])
	if parseErr != nil {
		return types.Command{}, fmt.Errorf("parse PX time failed: %w", parseErr)
	}

	value.shouldExpire = true
	value.expireAt = time.Now().Add(time.Millisecond * time.Duration(duration))

	app.m[args[1]] = value
	return types.NewStringCommand("OK"), nil
}

func (app *App) handleGET(args []string, _ types.Command) (types.Command, error) {
	if len(args) != 2 {
		return types.Command{}, fmt.Errorf("invalid number of arguments")
	}
	cmd := types.NewNullCommand()
	if value, exists := app.m[args[1]]; exists {
		if value.shouldExpire && time.Now().After(value.expireAt) {
			delete(app.m, args[1])
		} else {
			cmd = types.NewBulkStringCommand(value.value)
		}
	}
	return cmd, nil
}
