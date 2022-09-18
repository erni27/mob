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

// RequestHandler provides an interface for a request handler.
type RequestHandler[T any, U any] interface {
	Named
	Handle(context.Context, T) (U, error)
}

// RequestSender is the interface that wraps the mob's Send method.
type RequestSender[T any, U any] interface {
	// Send sends a given request T to an appropriate handler and returns a response U.
	//
	// If the appropriate handler does not exist in the sender's Mob instance, ErrHandlerNotFound is returned.
	Send(context.Context, T) (U, error)
}

// NewRequestSender returns a request sender which uses a given Mob instance.
func NewRequestSender[T any, U any](m *Mob) RequestSender[T, U] {
	return &sender[T, U]{m: m}
}

// A sender is a facilitator for a given request-response type pair.
type sender[T any, U any] struct {
	m *Mob
}

func (s *sender[T, U]) Send(ctx context.Context, req T) (U, error) {
	var res U
	k := reqHnKey{reqt: reflect.TypeOf(req), rest: reflect.TypeOf(res)}
	hn, ok := m.rhandlers[k]
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

// RegisterRequestHandlerTo adds a given request handler to the given Mob instance.
// Returns nil if the handler added successfully, an error otherwise.
//
// An only one handler for a given request-response pair can be registered.
// If support for multiple handlers for the same request-response pairs is needed within the same Mob instance,
// introduce type aliasing to avoid handlers' collision.
func RegisterRequestHandlerTo[T any, U any](m *Mob, hn RequestHandler[T, U]) error {
	if !isValid(hn) {
		return ErrInvalidHandler
	}
	var req T
	var res U
	k := reqHnKey{reqt: reflect.TypeOf(req), rest: reflect.TypeOf(res)}
	if _, ok := m.rhandlers[k]; ok {
		return ErrDuplicateHandler
	}
	m.rhandlers[k] = hn
	return nil
}

// RegisterRequestHandler adds a given request handler to the global Mob instance.
// Returns nil if the handler added successfully, an error otherwise.
//
// An only one handler for a given request-response pair can be registered.
// If support for multiple handlers for the same request-response pairs is needed within the Mob global instance,
// introduce type aliasing to avoid handlers' collision.
func RegisterRequestHandler[T any, U any](hn RequestHandler[T, U]) error {
	return RegisterRequestHandlerTo(m, hn)
}

// Send sends a given request T to an appropriate handler and returns a response U.
//
// If the appropriate handler does not exist in the global Mob instance, ErrHandlerNotFound is returned.
func Send[T any, U any](ctx context.Context, req T) (U, error) {
	return NewRequestSender[T, U](m).Send(ctx, req)
}
