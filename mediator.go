package mob

import (
	"context"
	"reflect"
)

// A reqHnKey is a request handler key consists of request and response types.
type reqHnKey struct {
	reqt reflect.Type
	rest reflect.Type
}

// rhandlers is a global registry for request handlers.
var rhandlers = map[reqHnKey]interface{}{}

// RequestHandler provides an interface for a request handler.
type RequestHandler[T any, U any] interface {
	Named
	Handle(context.Context, T) (U, error)
}

// RegisterRequestHandler adds a given request handler to the global registry.
// Returns nil if the handler added successfully, an error otherwise.
//
// An only one handler for a given request / response pair can be registered.
// If support for multiple handlers for same request / response pairs is needed, introduce type aliasing
// to avoid handlers' collision.
func RegisterRequestHandler[T any, U any](hn RequestHandler[T, U]) error {
	if !isValid(hn) {
		return ErrInvalidHandler
	}
	var req T
	var res U
	reqt := reflect.TypeOf(req)
	rest := reflect.TypeOf(res)
	k := reqHnKey{reqt: reqt, rest: rest}
	_, ok := rhandlers[k]
	if ok {
		return ErrDuplicateHandler
	}
	rhandlers[k] = hn
	return nil
}

// Send sends a given request T to an appropriate handler and returns a response U.
//
// If the appropriate handler does not exist in the global registry, ErrHandlerNotFound is returned.
func Send[T any, U any](ctx context.Context, req T) (U, error) {
	var res U
	reqt := reflect.TypeOf(req)
	rest := reflect.TypeOf(res)
	k := reqHnKey{reqt: reqt, rest: rest}
	hn, ok := rhandlers[k]
	if !ok {
		return res, ErrHandlerNotFound
	}
	// Dispatching result not checked because if a handler is found then it should always satisfy RequestHandler[T, U] interface.
	dhn, _ := hn.(RequestHandler[T, U])
	res, err := dhn.Handle(ctx, req)
	if err != nil {
		return res, HandlerError{Handler: dhn.Name(), Err: err}
	}
	return res, nil
}
