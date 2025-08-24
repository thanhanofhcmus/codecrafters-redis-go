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
	case "LPUSH":
		result, err = app.handleLPUSH(args)
	case "RPUSH":
		result, err = app.handleRPUSH(args)
	case "LRANGE":
		result, err = app.handleLRANGE(args)
	case "LLEN":
		result, err = app.handleLLEN(args)
	case "LPOP":
		result, err = app.handleLPOP(args)
	case "RPOP":
		result, err = app.handleRPOP(args)
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

	// TODO: notify keyspace, check expiry and check type

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

	value := Value{
		Key:       c.Key,
		String:    c.Value,
		ValueType: ValueTypeSimple,
	}
	app.dict[c.Key] = value
	app.NotifyKeySpace(value)

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

func (app *App) handleGenericPUSH(key string, newValues []string, fromLeft bool) (types.RawCmd, error) {
	value, exists := app.dict[key]
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}
	if expireTime, expireExists := app.expiry[key]; expireExists && time.Now().After(expireTime) {
		delete(app.dict, key)
		delete(app.expiry, key)
		return types.NewNullRawCmd(), nil
	}
	value.Key = key
	value.ValueType = ValueTypeList
	if fromLeft {
		value.List = append(value.List, newValues...)

	} else {
		slices.Reverse(newValues)
		value.List = append(newValues, value.List...)
	}
	app.dict[key] = value
	app.NotifyKeySpace(value)

	return types.NewIntegerRawCmd(int64(len(value.List))), nil
}

func (app *App) handleLPUSH(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LPUSH](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return app.handleGenericPUSH(c.Key, c.Values, true)
}

func (app *App) handleRPUSH(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.RPUSH](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return app.handleGenericPUSH(c.Key, c.Values, false)
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

func (app *App) handleGenricPOP(key string, fromLeft bool, count *int) (types.RawCmd, error) {
	value, exists := app.dict[key]
	if !exists {
		return types.NewNullRawCmd(), nil
	}
	if exists && value.ValueType != ValueTypeList {
		return types.RawCmd{}, NewWrongTypeError(ValueTypeSimple, value.ValueType)
	}
	if expireTime, expireExists := app.expiry[key]; expireExists && time.Now().After(expireTime) {
		delete(app.dict, key)
		delete(app.expiry, key)
		return types.NewNullRawCmd(), nil
	}

	length := len(value.List)

	if count == nil {
		if length == 0 {
			return types.NewNullRawCmd(), nil
		}
		v := ""
		v, value.List = splitListOne(value.List, fromLeft)
		app.dict[key] = value
		return types.NewBulkStringRawCmd(v), nil
	}

	var vs []string
	vs, value.List = splitList(value.List, fromLeft, *count)
	app.dict[key] = value

	return types.NewBulkArrayBulkString(vs), nil
}

func (app *App) handleLPOP(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.LPOP](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return app.handleGenricPOP(c.Key, true, c.Count)
}

func (app *App) handleRPOP(args []string) (types.RawCmd, error) {
	c, err := argsparser.Parse[cmd.RPOP](args)
	if err != nil {
		return types.RawCmd{}, err
	}
	return app.handleGenricPOP(c.Key, false, c.Count)
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
		v, value.List = splitListOne(value.List, true)
		app.dict[c.Key] = value
		return types.NewBulkArrayBulkString([]string{c.Key, v}), nil
	}

	// TODO: implement true timeout infinite or some kind of max timeout of a transaction
	// TODO: move to use transaction
	timeoutDuration := time.Hour * 100
	if c.TimeoutSecond != 0 {
		timeoutDuration = time.Second * time.Duration(c.TimeoutSecond)
	}
	connId := GetIdFromContext(ctx)

	keySpaceConsumer := app.SubscribeKeySpace(connId)
	defer app.UnsubscribeKeySpace(connId)

	for {
		select {
		case <-ctx.Done():
			// TODO: handle timeout error
			return types.RawCmd{}, ctx.Err()
		case <-time.After(timeoutDuration):
			return types.NewNullRawCmd(), nil
		case v := <-keySpaceConsumer:
			if v.Key != c.Key {
				continue
			}
			cmd, err := app.handleGenricPOP(v.Key, true, nil)
			if err != nil {
				return types.RawCmd{}, err
			}
			return types.NewBulkArrayBulkString([]string{c.Key, cmd.BulkString}), nil
		}
	}
}

func splitList[T any](l []T, fromLeft bool, count int) ([]T, []T) {
	n := len(l)
	if n == 0 {
		return nil, nil
	}
	if fromLeft {
		count = max(count, 0)
		count = min(count, n)
		return l[:count], l[count:]
	} else {
		count = max(count, 0)
		count = min(count, n)
		return l[:n-count], l[n-count:]
	}
}

func splitListOne[T any](l []T, fromLeft bool) (T, []T) {
	a, b := splitList(l, fromLeft, 1)
	return a[0], b
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
