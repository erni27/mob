package mob

import (
	"context"
	"fmt"
	"reflect"
)

// A reqHnKey is a request handler key consists of request and response types.
type reqHnKey struct {
	reqt reflect.Type
	rest reflect.Type
}

// RequestHandler provides an interface for a request handler.
type RequestHandler[T any, U any] interface {
	Handle(ctx context.Context, req T) (U, error)
}

// RequestHandlerFunc type is an adapter to allow the use of ordinary functions as request handlers.
type RequestHandlerFunc[T any, U any] func(ctx context.Context, req T) (U, error)

func (f RequestHandlerFunc[T, U]) Handle(ctx context.Context, req T) (U, error) {
	return f(ctx, req)
}

// RequestSender is the interface that wraps the mob's Send method.
type RequestSender[T any, U any] interface {
	// Send sends a given request T to an appropriate handler and returns a response U.
	//
	// If the appropriate handler does not exist in the sender's Mob instance, ErrHandlerNotFound is returned.
	Send(ctx context.Context, req T) (U, error)
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
	var err error
	k := reqHnKey{reqt: reflect.TypeOf(req), rest: reflect.TypeOf(res)}
	hn, ok := s.m.rhandlers[k]
	if !ok {
		return res, ErrHandlerNotFound
	}
	// Dispatching result not checked because if a handler is found then it should always satisfy RequestHandler[T, U] interface.
	dhn, _ := hn.embedded.(RequestHandler[T, U])
	if len(s.m.interceptors) != 0 {
		invoker := func(ctx context.Context, creq interface{}) (interface{}, error) {
			req, ok := creq.(T)
			if !ok {
				return nil, fmt.Errorf("%w: request is %T, want %T", ErrUnmarshal, creq, req)
			}
			return dhn.Handle(ctx, req)
		}
		chained := chainInterceptors(s.m.interceptors)
		cres, cerr := chained(ctx, req, invoker)
		if cerr == nil {
			res, ok = cres.(U)
			if !ok {
				err = fmt.Errorf("%w: response is %T, want %T", ErrUnmarshal, cres, res)
			}
		} else {
			err = cerr
		}
	} else {
		res, err = dhn.Handle(ctx, req)
	}
	if err != nil {
		if hn.name != "" {
			return res, fmt.Errorf("%s: %w", hn.name, err)
		}
		return res, err
	}
	return res, nil
}

// RegisterRequestHandlerTo adds a given request handler to the given Mob instance.
// Returns nil if the handler added successfully, an error otherwise.
//
// An only one handler for a given request-response pair can be registered.
// If support for multiple handlers for the same request-response pairs is needed within the same Mob instance,
// introduce type aliasing to avoid handlers' collision.
func RegisterRequestHandlerTo[T any, U any](m *Mob, rhn RequestHandler[T, U], opts ...Option) error {
	if !isValid(rhn) {
		return ErrInvalidHandler
	}
	var req T
	var res U
	k := reqHnKey{reqt: reflect.TypeOf(req), rest: reflect.TypeOf(res)}
	if _, ok := m.rhandlers[k]; ok {
		return ErrDuplicateHandler
	}
	hn := &handler{embedded: rhn}
	for _, opt := range opts {
		opt.apply(hn)
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
func RegisterRequestHandler[T any, U any](hn RequestHandler[T, U], opts ...Option) error {
	return RegisterRequestHandlerTo(m, hn, opts...)
}

// Send sends a given request T to an appropriate handler and returns a response U.
//
// If the appropriate handler does not exist in the global Mob instance, ErrHandlerNotFound is returned.
func Send[T any, U any](ctx context.Context, req T) (U, error) {
	return NewRequestSender[T, U](m).Send(ctx, req)
}
