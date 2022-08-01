package mob

import (
	"context"
	"testing"
)

//nolint:errcheck
func BenchmarkSend(b *testing.B) {
	defer clearRequestHandlers()
	handler := &DummyRequestHandler2{}
	err := RegisterRequestHandler[DummyRequest2, DummyResponse2](handler)
	if err != nil {
		b.Fatalf("register request handler: %v", err)
	}
	b.ResetTimer()
	ctx := context.Background()
	req := DummyRequest2{Int: 997}
	for n := 0; n < b.N; n++ {
		Send[DummyRequest2, DummyResponse2](ctx, req)
	}
}
