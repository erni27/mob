package mob

import (
	"context"
	"reflect"
	"sync"
)

// ehandlers is a global registry for event handlers.
var ehandlers = map[reflect.Type][]interface{}{}

// EventHandler provides an interface for an event handler.
type EventHandler[T any] interface {
	Named
	Handle(context.Context, T) error
}

// RegisterRequestHandler adds a given event handler to the global registry.
// Returns nil if the handler added successfully, an error otherwise.
//
// Multiple event handlers can be registered for a single event's type.
func RegisterEventHandler[T any](hn EventHandler[T]) error {
	if !isValid(hn) {
		return ErrInvalidHandler
	}
	var ev T
	evt := reflect.TypeOf(ev)
	hns, ok := ehandlers[evt]
	if !ok {
		ehandlers[evt] = []interface{}{hn}
		return nil
	}
	ehandlers[evt] = append(hns, hn)
	return nil
}

// Notify dispatches a given event and execute all handlers registered with a dispatched event's type.
// Handlers are executed concurrently and errors are collected, if any, they're returned to the client.
//
// If there is no appropriate handler in the global registry, ErrHandlerNotFound is returned.
func Notify[T any](ctx context.Context, ev T) error {
	evt := reflect.TypeOf(ev)
	hns, ok := ehandlers[evt]
	if !ok {
		return ErrHandlerNotFound
	}
	n := len(hns)
	c := make(chan HandlerError)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Dispatching result not checked because if a handler is found then it should always satisfy EventHandler[T] interface.
			dhn, _ := hns[i].(EventHandler[T])
			err := dhn.Handle(ctx, ev)
			if err != nil {
				c <- HandlerError{Handler: dhn.Name(), Err: err}
			}
		}(i)
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	var aggr AggregateHandlerError = make([]HandlerError, 0, n)
	for err := range c {
		aggr = append(aggr, err)
	}
	if len(aggr) > 0 {
		return aggr
	}
	return nil
}
