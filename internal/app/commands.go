package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/internal/argsparser"
	"github.com/codecrafters-io/redis-starter-go/pkg/types"
	"github.com/codecrafters-io/redis-starter-go/pkg/types/cmd"
)

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

func (app *App) HandleCommand(ctx context.Context, cmd types.RawCmd) (result types.RawCmd, err error) {
	args, err := convertArgsCmdToString(cmd)
	if err != nil {
		return types.RawCmd{}, err
	}

	command := args[0]
	switch strings.ToUpper(command) {
	// connections
	case "PING":
		result, err = app.handlePING(args)
	case "ECHO":
		result, err = app.handleECHO(args)

	// strings
	case "SET":
		result, err = app.handleSET(args)
	case "GET":
		result, err = app.handleGET(args)
	case "APPEND":
		result, err = app.handleAPPEND(args)

	// list
	case "RPUSH":
		result, err = app.handleRPUSH(args)
	case "LPUSH":
		result, err = app.handleLPUSH(args)
	case "LRANGE":
		result, err = app.handleLRANGE(args)
	case "LLEN":
		result, err = app.handleLLEN(args)
	case "LPOP":
		result, err = app.handleLPOP(args)
	case "BLPOP":
		result, err = app.handleBLPOP(ctx, args)
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

func (app *App) handleAPPEND(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.APPEND](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value := app.dict[c.Key]
	value.String += c.Value

	return types.NewIntegerRawCmd(int64(len(value.String))), nil
}

func (app *App) handleSET(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.SET](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	_, oldValueExists := app.dict[c.Key]

	if c.SetKey.Key != "" {
		if c.SetKey.NX && oldValueExists {
			return types.NewNullRawCmd(), nil
		}
		if c.SetKey.XX && !oldValueExists {
			return types.NewNullRawCmd(), nil
		}
	}

	app.dict[c.Key] = value{
		Key:       c.Key,
		String:    c.Value,
		ValueType: ValueTypeSimple,
	}

	if c.Expire.Key != "" {
		now := time.Now()
		switch c.Expire.Key {
		case "EX":
			expireAt := now.Add(time.Second * time.Duration(c.Expire.EX))
			app.expiry[c.Key] = expireAt
		case "PX":
			expireAt := now.Add(time.Millisecond * time.Duration(c.Expire.PX))
			app.expiry[c.Key] = expireAt
		case "EXAT":
			expireAt := time.Unix(int64(c.Expire.EXAT), 0)
			app.expiry[c.Key] = expireAt
		case "PXAT":
			expireAt := time.UnixMilli(int64(c.Expire.PXAT))
			app.expiry[c.Key] = expireAt
		case "KEEPTTL":
			// DO nothing
		default:
			panic("should not get to here")
		}
	}

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

	value, exists := app.dict[c.Key]
	if !exists {
		return types.NewNullRawCmd(), nil
	}
	if value.ValueType != ValueTypeSimple {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}
	if expireTime, expireExists := app.expiry[c.Key]; expireExists && time.Now().After(expireTime) {
		delete(app.dict, c.Key)
		delete(app.expiry, c.Key)
		return types.NewNullRawCmd(), nil
	}

	return types.NewBulkStringRawCmd(value.String), nil
}

func (app *App) handleRPUSH(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.RPUSH](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.dict[c.Key]
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}
	value.Key = c.Key
	value.ValueType = ValueTypeList
	value.List = append(value.List, c.Values...)

	app.dict[c.Key] = value

	return types.NewIntegerRawCmd(int64(len(value.List))), nil
}

func (app *App) handleLPUSH(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LPUSH](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.dict[c.Key]
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}
	value.Key = c.Key
	value.ValueType = ValueTypeList

	slices.Reverse(c.Values)
	value.List = append(c.Values, value.List...)

	app.dict[c.Key] = value

	return types.NewIntegerRawCmd(int64(len(value.List))), nil
}

func (app *App) handleLRANGE(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LRANGE](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.dict[c.Key]
	length := len(value.List)

	start := c.Start
	if start < 0 {
		start = max(0, length+start)
	}

	stop := c.Stop
	if stop < 0 {
		stop = max(0, length+stop)
	}
	stop = min(stop+1, length)

	if !exists || start > length || start >= stop {
		return types.NewBulkArrayBulkString(nil), nil
	}

	return types.NewBulkArrayBulkString(value.List[start:stop]), nil
}

func (app *App) handleLLEN(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LLEN](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.dict[c.Key]
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}

	return types.NewIntegerRawCmd(int64(len(value.List))), nil

}

func (app *App) handleLPOP(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LPOP](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	value, exists := app.dict[c.Key]
	if !exists {
		return types.NewNullRawCmd(), nil
	}
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}

	length := len(value.List)

	if c.Count == nil {
		if length == 0 {
			return types.NewNullRawCmd(), nil
		}
		v := ""
		v, value.List = splitListOne(value.List)
		app.dict[c.Key] = value
		return types.NewBulkStringRawCmd(v), nil
	}

	var vs []string
	vs, value.List = splitList(value.List, *c.Count)
	app.dict[c.Key] = value

	return types.NewBulkArrayBulkString(vs), nil
}

func (app *App) handleBLPOP(ctx context.Context, args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.BLPOP](args)
	if err != nil {
		return types.RawCmd{}, err
	}

	// TODO: support multi list
	if len(c.KeyRest) != 0 {
		return types.NewNullRawCmd(), NewHandleCommandError("BLPOP", fmt.Errorf("multi keys is not supported"))
	}

	v := ""

	// non blocking
	value, exists := app.dict[c.Key]
	if exists && len(value.List) > 0 {
		v, value.List = splitListOne(value.List)
		app.dict[c.Key] = value
		return types.NewBulkArrayBulkString([]string{c.Key, v}), nil
	}

	// TODO: implement true timeout infinite
	timeoutDuration := time.Hour * 100
	if c.TimeoutSecond != 0 {
		timeoutDuration = time.Second * time.Duration(c.TimeoutSecond)
	}
	select {
	case <-ctx.Done():
		// TODO: handle timeout error
		return types.RawCmd{}, ctx.Err()
	case <-time.After(timeoutDuration):
		return types.NewNullRawCmd(), nil
	}
}

func splitList[T any](l []T, count int) ([]T, []T) {
	count = max(count, 0)
	count = min(count, len(l)-1)
	return l[:count], l[count:]
}

func splitListOne[T any](l []T) (T, []T) {
	a, b := splitList(l, 1)
	return a[0], b
}
