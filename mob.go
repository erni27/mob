// Package mob is a simple mediator / observer library.
// It supports in-process requests / events processing.
package mob

import (
	"errors"
	"reflect"
)

var m *Mob

func init() {
	m = New()
}

// A Mob is a request / event handlers registry.
type Mob struct {
	interceptors []Interceptor
	rhandlers    map[reqHnKey]*handler
	ehandlers    map[reflect.Type][]*handler
}

// New returns an initialized Mob instance.
func New() *Mob {
	return &Mob{rhandlers: map[reqHnKey]*handler{}, ehandlers: map[reflect.Type][]*handler{}}
}

var (
	// ErrHandlerNotFound indicates that a requested handler is not registered.
	ErrHandlerNotFound = errors.New("mob: handler not found")
	// ErrInvalidHandler indicates that a given handler is not valid.
	ErrInvalidHandler = errors.New("mob: invalid handler")
	// ErrDuplicateHandler indicates that a handler for a given request / response pair is already registered.
	// It applies only to request handlers.
	ErrDuplicateHandler = errors.New("mob: duplicate handler")
	// ErrUnmarshal indicates that a request or a response type is malformed and cannot be
	// unmarshal to a given type.
	// It happens when a request or a response type is modified in the request processing pipeline.
	ErrUnmarshal = errors.New("mob: failed to unmarshal")
)

type handler struct {
	name     string
	embedded interface{}
}

// An AggregateHandlerError is a type alias for a slice of handler errors. It applies only to event handlers.
type AggregateHandlerError []error

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

type token struct{}

func isValid(hn any) bool {
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
