package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/argsparser"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
	"github.com/codecrafters-io/redis-starter-go/pkg/types/cmd"
)

type value struct {
	value        string
	shouldExpire bool
	expireAt     time.Time
}

type App struct {
	m map[string]value
}

func convertArgsCmdToString(cmd types.RawCmd) ([]string, error) {
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

func (app *App) HandleCommand(cmd types.RawCmd) (result types.RawCmd, err error) {
	args, err := convertArgsCmdToString(cmd)
	if err != nil {
		return types.RawCmd{}, err
	}

	command := args[0]

	switch strings.ToUpper(command) {
	case "PING":
		result = types.NewStringRawCmd("PONG")
	case "ECHO":
		if len(args) < 2 {
			err = NewExpectArgumentError("ECHO")
			break
		}
		result = types.NewBulkStringRawCmd(args[1])
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

func (app *App) handleSET(args []string, _ types.RawCmd) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.SET](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value := value{
		value: c.Value,
	}

	/// TODO: handle other args

	if !c.GET {
		return types.NewStringRawCmd("OK"), nil
	}

	if argsLen == 3 {
		app.m[args[1]] = value
	}

	// args length now is 5

	if strings.ToUpper(args[3]) != "PX" {
		return types.RawCmd{}, NewInvalidOptionError(args[3])
	}

	duration, parseErr := strconv.Atoi(args[4])
	if parseErr != nil {
		return types.RawCmd{}, fmt.Errorf("parse PX time failed: %w", parseErr)
	}

	value.shouldExpire = true
	value.expireAt = time.Now().Add(time.Millisecond * time.Duration(duration))

	app.m[args[1]] = value
	return types.NewStringRawCmd("OK"), nil
}

func (app *App) handleGET(args []string, _ types.RawCmd) (types.RawCmd, error) {
	if len(args) != 2 {
		return types.RawCmd{}, fmt.Errorf("invalid number of arguments")
	}
	cmd := types.NewNullRawCmd()
	if value, exists := app.m[args[1]]; exists {
		if value.shouldExpire && time.Now().After(value.expireAt) {
			delete(app.m, args[1])
		} else {
			cmd = types.NewBulkStringRawCmd(value.value)
		}
	}
	return cmd, nil
}
