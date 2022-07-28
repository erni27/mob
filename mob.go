// Package mob is a simple mediator / observer library.
// It supports in-process requests / events processing.
package mob

import (
	"errors"
	"reflect"
)

// Named is an interface that wraps Name() string method.
type Named interface {
	Name() string
}

var (
	// ErrHandlerNotFound indicates that a requested handler is not registerd.
	ErrHandlerNotFound = errors.New("mob: handler not found")
	// ErrInvalidHandler indicates that a given handler is not valid.
	ErrInvalidHandler = errors.New("mob: invalid handler")
	// ErrDuplicateHandler indicates that a handler for a given req / res pair is already registered.
	// It applies only to request handlers.
	ErrDuplicateHandler = errors.New("mob: duplicate handler")
)

// A HandlerError is an error wrapper which identifies the error source (handler) by its name.
type HandlerError struct {
	Handler string
	Err     error
}

func (e HandlerError) Error() string {
	return e.Handler + ": " + e.Err.Error()
}

// An AggregateHandlerError is a type alias for a slice of handler errors. It applies only to event handlers.
type AggregateHandlerError []HandlerError

func (e AggregateHandlerError) Error() string {
	var msg string
	for _, err := range e {
		msg += err.Error() + ";"
	}
	return msg[:len(msg)-1]
}

func (e AggregateHandlerError) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (e HandlerError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

type token struct{}

func isValid(hn Named) bool {
	if hn == nil {
		return false
	}
	if _, ok := nilable[reflect.TypeOf(hn).Kind()]; ok {
		return !reflect.ValueOf(hn).IsNil()
	}
	return true
}

var nilable map[reflect.Kind]token = map[reflect.Kind]token{
	reflect.Ptr:   {},
	reflect.Map:   {},
	reflect.Array: {},
	reflect.Chan:  {},
	reflect.Slice: {},
}
