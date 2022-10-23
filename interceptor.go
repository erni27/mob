package mob

import (
	"context"
)

// SendInvoker is a function called by an Interceptor to invoke
// the next Interceptor in the chain or the underlying invoker.
type SendInvoker func(ctx context.Context, req interface{}) (interface{}, error)

// Interceptor intercepts an invocation of a Send method.
type Interceptor func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error)

// AddInterceptorTo adds an Interceptor to the given Mob instance.
// Interceptors are invoked in order they're added to the chain.
func AddInterceptorTo(m *Mob, interceptor Interceptor) {
	m.interceptors = append(m.interceptors, interceptor)
}

// AddInterceptorTo adds an Interceptor to the global Mob instance.
// Interceptors are invoked in order they're added to the chain.
func AddInterceptor(interceptor Interceptor) {
	AddInterceptorTo(m, interceptor)
}

func chainInterceptors(interceptors []Interceptor) Interceptor {
	if len(interceptors) == 0 {
		return nil
	}
	if len(interceptors) == 1 {
		return interceptors[0]
	}
	return func(ctx context.Context, req interface{}, invoker SendInvoker) (interface{}, error) {
		return interceptors[0](ctx, req, buildInvoker(invoker, interceptors, 0))
	}
}

func buildInvoker(inner SendInvoker, interceptors []Interceptor, depth int) func(context.Context, interface{}) (interface{}, error) {
	if len(interceptors)-1 == depth {
		return inner
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return interceptors[depth+1](ctx, req, buildInvoker(inner, interceptors, depth+1))
	}
}
