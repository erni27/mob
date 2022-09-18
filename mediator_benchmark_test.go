package mob

import (
	"context"
	"testing"
)

//nolint:errcheck
func BenchmarkSend(b *testing.B) {
	defer clear()
	if err := RegisterRequestHandler[DummyRequest1, DummyResponse1](&DummyRequestHandler1{}); err != nil {
		b.Fatalf("register request handler: %v", err)
	}
	if err := RegisterRequestHandler[DummyRequest2, DummyResponse2](&DummyRequestHandler2{}); err != nil {
		b.Fatalf("register request handler: %v", err)
	}
	if err := RegisterRequestHandler[DummyRequest1, DummyResponse2](&DummyRequestHandler3{}); err != nil {
		b.Fatalf("register request handler: %v", err)
	}
	ctx := context.Background()
	req := DummyRequest2{Int: 997}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		res, err = Send[DummyRequest2, DummyResponse2](ctx, req)
		if err != nil {
			b.Fatalf("want no err, got %v", err)
		}
	}
}

var err error
var res DummyResponse2
