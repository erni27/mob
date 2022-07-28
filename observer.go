package mob

import (
	"context"
	"reflect"
)

var ehandlers map[reflect.Type][]interface{} = map[reflect.Type][]interface{}{}

type EventHandler[T any] interface {
	Named
	Handle(context.Context, T) error
}

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
	hns = append(hns, hn)
	ehandlers[evt] = hns
	return nil
}

func Dispatch[T any](ctx context.Context, ev T) error {
	evt := reflect.TypeOf(ev)
	hns, ok := ehandlers[evt]
	if !ok {
		return ErrHandlerNotFound
	}
	n := len(hns)
	wc := make(chan token, n)
	ec := make(chan HandlerError)
	for i := 0; i < n; i++ {
		wc <- token{}
		go func(i int) {
			defer func() { <-wc }()
			// Dispatching result not checked because if a handler is found then it should always satisfy EventHandler[T] interface.
			dhn, _ := hns[i].(EventHandler[T])
			err := dhn.Handle(ctx, ev)
			if err != nil {
				ec <- HandlerError{Handler: dhn.Name(), Err: err}
			}
		}(i)
	}
	go func() {
		for i := 0; i < n; i++ {
			wc <- token{}
		}
		close(ec)
	}()
	errors := make([]HandlerError, 0, n)
	for err := range ec {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		var aggr AggregateHandlerError = errors
		return aggr
	}
	return nil
}
