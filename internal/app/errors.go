package app

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

type InvalidTypeError struct {
	Expected types.Sym
	Actual   types.Sym
}

func (e InvalidTypeError) Error() string {
	return fmt.Sprintf("expect symbol of type `%s`, got `%s`", types.GetSymString(e.Expected), types.GetSymString(e.Actual))
}

func NewInvalidTypeError(execpted, actual types.Sym) InvalidTypeError {
	return InvalidTypeError{
		Expected: execpted,
		Actual:   actual,
	}
}

type ArrayElementError struct {
	Index int
	Err   error
}

func (e ArrayElementError) Error() string {
	return fmt.Sprintf("error at at index `%d`: %s", e.Index, e.Err)
}

func (e ArrayElementError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

func NewArrayElementError(index int, err error) ArrayElementError {
	return ArrayElementError{
		Index: index,
		Err:   err,
	}
}
