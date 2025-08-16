package app

import (
	"errors"
	"fmt"

	"github.com/codecrafters-io/redis-starter-go/pkg/types"
)

type HandleCommandError struct {
	Command string
	Err     error
}

func (e HandleCommandError) Error() string {
	return fmt.Sprintf("error while handle command `%s`: %s", e.Command, e.Err)
}

func (e HandleCommandError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

func NewHandleCommandError(command string, err error) HandleCommandError {
	return HandleCommandError{
		Command: command,
		Err:     err,
	}
}

type ExpectArgumentError struct {
	Base string
}

func (e ExpectArgumentError) Error() string {
	return fmt.Sprintf("expect argument for `%s` but got nothing", e.Base)
}

func NewExpectArgumentError(base string) ExpectArgumentError {
	return ExpectArgumentError{
		Base: base,
	}
}

type InvalidOptionError struct {
	Option string
}

func (e InvalidOptionError) Error() string {
	return fmt.Sprintf("option `%s` is not valid", e.Option)
}

func NewInvalidOptionError(option string) InvalidOptionError {
	return InvalidOptionError{
		Option: option,
	}
}

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
