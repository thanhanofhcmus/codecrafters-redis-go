package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/argsparser"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
	"github.com/codecrafters-io/redis-starter-go/pkg/types/cmd"
)

type mType int

const (
	mTypeSimple mType = iota
	mTypeList
)

type value struct {
	value        string
	listValues   []string
	mType        mType
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
		result, err = app.handlePING(args)
	case "ECHO":
		result, err = app.handleECHO(args)
	case "SET":
		result, err = app.handleSET(args)
	case "GET":
		result, err = app.handleGET(args)
	case "RPUSH":
		result, err = app.handleRPUSH(args)
	case "LRANGE":
		result, err = app.handleLRANGE(args)
	default:
		err = fmt.Errorf("unknown command `%s`", command)
	}

	if err != nil {
		err = NewHandleCommandError(command, err)
	}

	return
}

func (app *App) handlePING(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.PING](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return types.NewStringRawCmd(c.Message), nil
}

func (app *App) handleECHO(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.ECHO](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return types.NewStringRawCmd(c.Message), nil
}

func (app *App) handleSET(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.SET](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	oldValue, oldValueExists := app.m[c.Key]

	if c.SetKey.Key != "" {
		if c.SetKey.NX && oldValueExists {
			return types.NewNullRawCmd(), nil
		}
		if c.SetKey.XX && !oldValueExists {
			return types.NewNullRawCmd(), nil
		}
	}

	newValue := value{value: c.Value, mType: mTypeSimple}

	if c.Expire.Key != "" {
		now := time.Now()
		newValue.shouldExpire = true
		switch c.Expire.Key {
		case "EX":
			newValue.expireAt = now.Add(time.Second * time.Duration(c.Expire.EX))
		case "PX":
			newValue.expireAt = now.Add(time.Millisecond * time.Duration(c.Expire.PX))
		case "EXAT":
			newValue.expireAt = time.Unix(int64(c.Expire.EXAT), 0)
		case "PXAT":
			newValue.expireAt = time.UnixMilli(int64(c.Expire.PXAT))
		case "KEEPTTL":
			newValue.shouldExpire = oldValue.shouldExpire
			newValue.expireAt = oldValue.expireAt
		default:
			panic("should not get to here")
		}
	}

	app.m[c.Key] = newValue

	if c.GET {
		return types.NewStringRawCmd(c.Value), nil
	}
	return types.NewStringRawCmd("OK"), nil
}

func (app *App) handleGET(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.GET](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	cmd := types.NewNullRawCmd()
	if value, exists := app.m[c.Key]; exists {
		if value.shouldExpire && time.Now().After(value.expireAt) {
			delete(app.m, args[1])
		} else if value.mType != mTypeSimple {
			return types.RawCmd{}, NewWrongTypeError(mTypeSimple, value.mType)
		} else {
			cmd = types.NewBulkStringRawCmd(value.value)
		}
	}
	return cmd, nil
}

func (app *App) handleRPUSH(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.RPUSH](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.m[c.Key]
	if exists && value.mType != mTypeList {
		return types.RawCmd{}, NewWrongTypeError(mTypeSimple, value.mType)
	}
	value.mType = mTypeList
	value.listValues = append(value.listValues, c.Values...)

	app.m[c.Key] = value

	return types.NewIntegerRawCmd(int64(len(value.listValues))), nil
}

func (app *App) handleLRANGE(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LRANGE](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.m[c.Key]
	length := len(value.listValues)
	start, stop := c.Start, min(len(value.listValues), c.Stop+1)

	if !exists || start > length || start >= stop {
		return types.NewBulkArrayBulkString(nil), nil
	}

	return types.NewBulkArrayBulkString(value.listValues[start:stop]), nil
}
