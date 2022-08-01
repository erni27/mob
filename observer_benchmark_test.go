package mob

import (
	"context"
	"fmt"
	"testing"
)

type DummyEventHandler4 struct{}

func (*DummyEventHandler4) Name() string {
	return "DummyEventHandler4"
}

func (*DummyEventHandler4) Handle(_ context.Context, _ DummyEvent1) error {
	return nil
}

func BenchmarkNotify(b *testing.B) {
	tests := [][]EventHandler[DummyEvent1]{
		{&DummyEventHandler4{}},
		{&DummyEventHandler4{}, &DummyEventHandler4{}},
		{&DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}},
		{&DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}},
		{&DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}, &DummyEventHandler4{}},
	}
	for _, handlers := range tests {
		b.Run(fmt.Sprintf("number of handlers %d", len(handlers)), func(b *testing.B) {
			defer clearEventHandlers()
			for _, handler := range handlers {
				if err := RegisterEventHandler(handler); err != nil {
					b.Fatalf("register event handler: %v", err)
				}
			}
			ctx := context.Background()
			ev := DummyEvent1{String: "string", Int: 997}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Notify(ctx, ev)
			}
		})
	}
}
