package mob

import (
	"context"
	"reflect"
	"sync"
)

// EventHandler provides an interface for an event handler.
type EventHandler[T any] interface {
	Named
	Handle(context.Context, T) error
}

// EventNotifier is the interface that wraps the mob's Notify method.
type EventNotifier[T any] interface {
	// Notify dispatches a given event and execute all handlers registered with a dispatched event's type.
	// Handlers are executed concurrently and errors are collected, if any, they're returned to the client.
	//
	// If there is no appropriate handler in the notifier's Mob instance, ErrHandlerNotFound is returned.
	Notify(context.Context, T) error
}

// NewEventNotifier returns an event notifier which uses a given Mob instance.
func NewEventNotifier[T any](m *Mob) EventNotifier[T] {
	return &notifier[T]{m: m}
}

// A notifier is a facilitator for a given event type.
type notifier[T any] struct {
	m *Mob
}

func (nf *notifier[T]) Notify(ctx context.Context, event T) error {
	hns, ok := nf.m.ehandlers[reflect.TypeOf(event)]
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
			if err := dhn.Handle(ctx, event); err != nil {
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

// RegisterEventHandlerTo adds a given event handler to the given Mob instance.
// Returns nil if the handler added successfully, an error otherwise.
//
// Multiple event handlers can be registered for a single event's type.
func RegisterEventHandlerTo[T any](m *Mob, hn EventHandler[T]) error {
	if !isValid(hn) {
		return ErrInvalidHandler
	}
	var ev T
	k := reflect.TypeOf(ev)
	m.ehandlers[k] = append(m.ehandlers[k], hn)
	return nil
}

// RegisterEventHandler adds a given event handler to the global Mob instance.
// Returns nil if the handler added successfully, an error otherwise.
//
// Multiple event handlers can be registered for a single event's type.
func RegisterEventHandler[T any](hn EventHandler[T]) error {
	return RegisterEventHandlerTo(m, hn)
}

// Notify dispatches a given event and execute all handlers registered with a dispatched event's type.
// Handlers are executed concurrently and errors are collected, if any, they're returned to the client.
//
// If there is no appropriate handler in the global Mob instance, ErrHandlerNotFound is returned.
func Notify[T any](ctx context.Context, event T) error {
	return NewEventNotifier[T](m).Notify(ctx, event)
}
