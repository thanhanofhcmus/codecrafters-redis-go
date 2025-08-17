package encoding

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

type EncodingError struct {
	cmd           types.RawCmd
	encondingType string // Marshal | Unmarshal
	step          string
	inner         error
}

func (err *EncodingError) Error() string {
	return fmt.Sprintf("(%s)[%s|%s]: %s", err.encondingType, types.GetSymString(err.cmd.Sym), err.step, err.inner)
}

func (err *EncodingError) Is(target error) bool {
	return errors.Is(err.inner, target)
}

func newMarshalError(cmd types.RawCmd, step string, inner error) *EncodingError {
	return &EncodingError{
		cmd:           cmd,
		encondingType: "Marshal",
		step:          step,
		inner:         inner,
	}
}

func newUnmarshalError(cmd types.RawCmd, step string, inner error) *EncodingError {
	return &EncodingError{
		cmd:           cmd,
		encondingType: "Unmarshal",
		step:          step,
		inner:         inner,
	}
}
