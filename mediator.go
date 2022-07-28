package mob

import (
	"context"
	"reflect"
)

type reqHnKey struct {
	req reflect.Type
	res reflect.Type
}

var rhandlers map[reqHnKey]interface{} = map[reqHnKey]interface{}{}

type RequestHandler[T any, U any] interface {
	Handle(context.Context, T) (U, error)
}

func Send[T any, U any](ctx context.Context, req T) (U, error) {
	var res U
	reqt := reflect.TypeOf(req)
	rest := reflect.TypeOf(res)
	k := reqHnKey{req: reqt, res: rest}
	hn, ok := rhandlers[k]
	if !ok {
		return res, ErrHandlerNotFound
	}
	// Dispatching result not checked because if a handler is found then it should always satisfy RequestHandler[T, U] interface.
	dhn, _ := hn.(RequestHandler[T, U])
	return dhn.Handle(ctx, req)
}
