package mob

import (
	"context"
	"errors"
	"testing"
)

type EventHandlerMock[T any] interface {
	EventHandler[T]
	Calls() int
}

type DummyEvent1 struct {
	String string
	Int    int
}

type DummyEventHandler1 struct {
	handleFunc func(context.Context, DummyEvent1) error
	calls      int
}

func (*DummyEventHandler1) Name() string {
	return "DummyEventHandler1"
}

func (h *DummyEventHandler1) Handle(ctx context.Context, ev DummyEvent1) error {
	h.calls++
	return h.handleFunc(ctx, ev)
}

func (h *DummyEventHandler1) Calls() int {
	return h.calls
}

type DummyEventHandler2 struct {
	calls      int
	handleFunc func(context.Context, DummyEvent1) error
}

func (*DummyEventHandler2) Name() string {
	return "DummyEventHandler1"
}

func (h *DummyEventHandler2) Handle(ctx context.Context, ev DummyEvent1) error {
	h.calls++
	return h.handleFunc(ctx, ev)
}

func (h *DummyEventHandler2) Calls() int {
	return h.calls
}

type DummyEventHandler3 struct {
	handleFunc func(context.Context, DummyEvent1) error
	calls      int
}

func (*DummyEventHandler3) Name() string {
	return "DummyEventHandler1"
}

func (h *DummyEventHandler3) Handle(ctx context.Context, ev DummyEvent1) error {
	h.calls++
	return h.handleFunc(ctx, ev)
}

func (h *DummyEventHandler3) Calls() int {
	return h.calls
}

func TestRegisterEventHandler_InvalidHandler(t *testing.T) {
	tests := []struct {
		name string
		arg  EventHandler[DummyEvent1]
		want error
	}{
		{
			name: "nil interface",
			arg:  nil,
			want: ErrInvalidHandler,
		},
		{
			name: "nil value",
			arg:  (*DummyEventHandler1)(nil),
			want: ErrInvalidHandler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterEventHandler(tt.arg); err != tt.want {
				t.Errorf("want %v, got error %v", tt.want, err)
			}
		})
	}
}

func TestRegisterEventHandler(t *testing.T) {
	defer clear()
	tests := []struct {
		name string
		arg  EventHandler[DummyEvent1]
	}{
		{
			name: "dummy handler 1",
			arg:  &DummyEventHandler1{},
		},
		{
			name: "dummy handler 2",
			arg:  &DummyEventHandler2{},
		},
		{
			name: "dummy handler 3",
			arg:  &DummyEventHandler3{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterEventHandler(tt.arg); err != nil {
				t.Errorf("want success, got error %v", err)
			}
		})
	}
}

func TestNotify_HandlerNotFound(t *testing.T) {
	if err := Notify(context.Background(), DummyEvent1{}); err != ErrHandlerNotFound {
		t.Errorf("want error %v, got %v", ErrHandlerNotFound, err)
	}
}

func TestNotify(t *testing.T) {
	errFirst := errors.New("first")
	errSecond := errors.New("second")
	errThird := errors.New("third")
	tests := []struct {
		name     string
		arg      DummyEvent1
		handlers []EventHandlerMock[DummyEvent1]
		want     []error
	}{
		{
			name: "single handler",
			arg:  DummyEvent1{},
			handlers: []EventHandlerMock[DummyEvent1]{
				&DummyEventHandler1{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
			},
			want: nil,
		},
		{
			name: "single handler failed",
			arg:  DummyEvent1{},
			handlers: []EventHandlerMock[DummyEvent1]{
				&DummyEventHandler1{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return errFirst },
				},
			},
			want: []error{errFirst},
		},
		{
			name: "multiple handlers",
			arg:  DummyEvent1{},
			handlers: []EventHandlerMock[DummyEvent1]{
				&DummyEventHandler1{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
				&DummyEventHandler2{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
				&DummyEventHandler3{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
			},
			want: nil,
		},
		{
			name: "multiple handlers, one failed",
			arg:  DummyEvent1{},
			handlers: []EventHandlerMock[DummyEvent1]{
				&DummyEventHandler1{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return errFirst },
				},
				&DummyEventHandler2{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
				&DummyEventHandler3{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return nil },
				},
			},
			want: []error{errFirst},
		},
		{
			name: "multiple handlers, all failed",
			arg:  DummyEvent1{},
			handlers: []EventHandlerMock[DummyEvent1]{
				&DummyEventHandler1{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return errFirst },
				},
				&DummyEventHandler2{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return errSecond },
				},
				&DummyEventHandler3{
					handleFunc: func(ctx context.Context, de DummyEvent1) error { return errThird },
				},
			},
			want: []error{errFirst, errSecond, errThird},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clear()
			for _, hn := range tt.handlers {
				if err := RegisterEventHandler[DummyEvent1](hn); err != nil {
					t.Fatalf("want success, got %v", err)
				}
			}
			err := Notify(context.Background(), tt.arg)
			for i, hn := range tt.handlers {
				if calls := hn.Calls(); calls != 1 {
					t.Fatalf("want handler %d called exactly 1, got %d", i+1, calls)
				}
			}
			for _, wantErr := range tt.want {
				if !errors.Is(err, wantErr) {
					t.Errorf("want %v, got %v", wantErr, err)
				}
			}
		})
	}
}
