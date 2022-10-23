package mob

import (
	"context"
	"errors"
	"testing"
)

type EmptyRequest struct{}

type EmptyResponse struct{}

func TestUseInterceptor_ContextValuePropagation(t *testing.T) {
	defer clear()

	type dummyContextKey0 struct{}
	type dummyContextKey1 struct{}
	type dummyContextKey2 struct{}
	type dummyContextKey3 struct{}

	const (
		dummyContextValue0 = "dummyValue0"
		dummyContextValue1 = "dummyValue1"
		dummyContextValue2 = "dummyValue2"
		dummyContextValue3 = "dummyValue3"
	)

	checks := make([]bool, 4)

	AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		checks[0] = true
		if got, ok := ctx.Value(dummyContextKey0{}).(string); !ok || got != dummyContextValue0 {
			t.Fatalf("interceptor1 got value in ctx %v, want %s", got, dummyContextValue0)
		}
		if got, ok := ctx.Value(dummyContextKey1{}).(string); ok {
			t.Fatalf("interceptor1 got value in ctx %v, want no value", got)
		}
		if got, ok := ctx.Value(dummyContextKey2{}).(string); ok {
			t.Fatalf("interceptor1 got value in ctx %v, want no value", got)
		}
		if got, ok := ctx.Value(dummyContextKey3{}).(string); ok {
			t.Fatalf("interceptor1 got value in ctx %v, want no value", got)
		}
		return invoker(context.WithValue(ctx, dummyContextKey1{}, dummyContextValue1), req)
	})
	AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		checks[1] = true
		if got, ok := ctx.Value(dummyContextKey0{}).(string); !ok || got != dummyContextValue0 {
			t.Fatalf("interceptor2 got value in ctx %v, want %s", got, dummyContextValue0)
		}
		if got, ok := ctx.Value(dummyContextKey1{}).(string); !ok || got != dummyContextValue1 {
			t.Fatalf("interceptor2 got value in ctx %v, want %s", got, dummyContextValue1)
		}
		if got, ok := ctx.Value(dummyContextKey2{}).(string); ok {
			t.Fatalf("interceptor2 got value in ctx %v, want no value", got)
		}
		if got, ok := ctx.Value(dummyContextKey3{}).(string); ok {
			t.Fatalf("interceptor2 got value in ctx %v, want no value", got)
		}
		return invoker(context.WithValue(ctx, dummyContextKey2{}, dummyContextValue2), req)
	})
	AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		checks[2] = true
		if got, ok := ctx.Value(dummyContextKey0{}).(string); !ok || got != dummyContextValue0 {
			t.Fatalf("interceptor3 got value in ctx %v, want %s", got, dummyContextValue0)
		}
		if got, ok := ctx.Value(dummyContextKey1{}).(string); !ok || got != dummyContextValue1 {
			t.Fatalf("interceptor3 got value in ctx %v, want %s", got, dummyContextValue1)
		}
		if got, ok := ctx.Value(dummyContextKey2{}).(string); !ok || got != dummyContextValue2 {
			t.Fatalf("interceptor3 got value in ctx %v, want %s", got, dummyContextValue2)
		}
		if got, ok := ctx.Value(dummyContextKey3{}).(string); ok {
			t.Fatalf("interceptor3 got value in ctx %v, want no value", got)
		}
		return invoker(context.WithValue(ctx, dummyContextKey3{}, dummyContextValue3), req)
	})

	var handler RequestHandlerFunc[*EmptyRequest, *EmptyResponse] = func(ctx context.Context, _ *EmptyRequest) (*EmptyResponse, error) {
		checks[3] = true
		if got, ok := ctx.Value(dummyContextKey0{}).(string); !ok || got != dummyContextValue0 {
			t.Fatalf("invoker got value in ctx %v, want %s", got, dummyContextValue0)
		}
		if got, ok := ctx.Value(dummyContextKey1{}).(string); !ok || got != dummyContextValue1 {
			t.Fatalf("invoker got value in ctx %v, want %s", got, dummyContextValue1)
		}
		if got, ok := ctx.Value(dummyContextKey2{}).(string); !ok || got != dummyContextValue2 {
			t.Fatalf("invoker got value in ctx %v, want %s", got, dummyContextValue2)
		}
		if got, ok := ctx.Value(dummyContextKey3{}).(string); !ok || got != dummyContextValue3 {
			t.Fatalf("invoker got value in ctx %v, want %s", got, dummyContextValue3)
		}
		return &EmptyResponse{}, nil
	}
	if err := RegisterRequestHandler[*EmptyRequest, *EmptyResponse](handler); err != nil {
		t.Fatalf("unexpected register err %v", err)
	}

	_, err := Send[*EmptyRequest, *EmptyResponse](context.WithValue(context.Background(), dummyContextKey0{}, dummyContextValue0), nil)
	if err != nil {
		t.Fatalf("unexpected send err: %v", err)
	}

	for i, check := range checks[:len(checks)-1] {
		if !check {
			t.Errorf("interceptor%d not called", i+1)
		}
	}
	if !checks[len(checks)-1] {
		t.Errorf("final invoker not called")
	}
}

func TestUseInterceptor_BrokenChain(t *testing.T) {
	defer clear()
	checks := make([]bool, 4)

	AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		checks[0] = true
		return invoker(ctx, req)
	})
	AddInterceptor(func(_ context.Context, _ interface{}, _ SendInvoker) (interface{}, error) {
		checks[1] = true
		return &EmptyResponse{}, nil
	})
	AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		checks[2] = true
		return invoker(ctx, req)
	})

	var handler RequestHandlerFunc[*EmptyRequest, *EmptyResponse] = func(_ context.Context, _ *EmptyRequest) (*EmptyResponse, error) {
		checks[3] = true
		return &EmptyResponse{}, nil
	}
	if err := RegisterRequestHandler[*EmptyRequest, *EmptyResponse](handler); err != nil {
		t.Fatalf("unexpected register err %v", err)
	}

	_, err := Send[*EmptyRequest, *EmptyResponse](context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected send err: %v", err)
	}

	for i, check := range checks[:len(checks)-2] {
		if !check {
			t.Errorf("interceptor%d not called", i+1)
		}
	}
	if checks[len(checks)-2] {
		t.Errorf("unexpected execution of interceptor%d", len(checks)-1)
	}
	if checks[len(checks)-1] {
		t.Errorf("unexpected execution of final invoker")
	}
}

func TestUseInterceptor_MalformedRequest(t *testing.T) {
	tests := []struct {
		name string
		req  interface{}
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "different request type",
			req:  DummyRequest2{},
		},
	}
	var handler RequestHandlerFunc[*EmptyRequest, *EmptyResponse] = func(_ context.Context, _ *EmptyRequest) (*EmptyResponse, error) {
		return &EmptyResponse{}, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clear()
			if err := RegisterRequestHandler[*EmptyRequest, *EmptyResponse](handler); err != nil {
				t.Fatalf("unexpected register err %v", err)
			}
			AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
				return invoker(ctx, tt.req)
			})
			_, err := Send[*EmptyRequest, *EmptyResponse](context.Background(), nil)
			if !errors.Is(err, ErrUnmarshal) {
				t.Fatalf("got err %v, want err %v", err, ErrUnmarshal)
			}
		})
	}
}

func TestUseInterceptor_MalformedResponse(t *testing.T) {
	tests := []struct {
		name string
		res  interface{}
	}{
		{
			name: "nil request",
			res:  nil,
		},
		{
			name: "different request type",
			res:  DummyResponse2{},
		},
	}
	var handler RequestHandlerFunc[*EmptyRequest, *EmptyResponse] = func(_ context.Context, _ *EmptyRequest) (*EmptyResponse, error) {
		return &EmptyResponse{}, nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clear()
			if err := RegisterRequestHandler[*EmptyRequest, *EmptyResponse](handler); err != nil {
				t.Fatalf("unexpected register err %v", err)
			}
			AddInterceptor(func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
				return tt.res, nil
			})
			_, err := Send[*EmptyRequest, *EmptyResponse](context.Background(), nil)
			if !errors.Is(err, ErrUnmarshal) {
				t.Fatalf("got err %v, want err %v", err, ErrUnmarshal)
			}
		})
	}
}
