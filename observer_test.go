package mob

import (
	"context"
	"errors"
	"strconv"
	"strings"
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

func (h *DummyEventHandler3) Handle(ctx context.Context, ev DummyEvent1) error {
	h.calls++
	return h.handleFunc(ctx, ev)
}

func (h *DummyEventHandler3) Calls() int {
	return h.calls
}

type DummyEventHandler5 struct {
	doSomethingFunc func(context.Context, DummyEvent1) error
	calls           int
}

func (h *DummyEventHandler5) DoSomething(ctx context.Context, ev DummyEvent1) error {
	h.calls++
	return h.doSomethingFunc(ctx, ev)
}

func (h *DummyEventHandler5) Calls() int {
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
			for i, hn := range tt.handlers {
				if err := RegisterEventHandler[DummyEvent1](hn, WithName("Handler"+strconv.Itoa(i+1))); err != nil {
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

func TestNotify_Named(t *testing.T) {
	var handler DummyEventHandler5
	dummyErr := errors.New("dummy err")
	tests := []struct {
		name  string
		arg   DummyEvent1
		setup func(*DummyEventHandler5)
		want  error
	}{
		{
			name: "success",
			arg:  DummyEvent1{String: "String", Int: 997},
			setup: func(h *DummyEventHandler5) {
				h.doSomethingFunc = func(_ context.Context, _ DummyEvent1) error {
					return nil
				}
			},
			want: nil,
		},
		{
			name: "error",
			arg:  DummyEvent1{String: "String", Int: 997},
			setup: func(h *DummyEventHandler5) {
				h.doSomethingFunc = func(_ context.Context, _ DummyEvent1) error {
					return dummyErr
				}
			},
			want: dummyErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clear()
			tt.setup(&handler)
			var handlerf EventHandlerFunc[DummyEvent1] = handler.DoSomething
			if err := RegisterEventHandler[DummyEvent1](handlerf, WithName("DummyEventHandler5")); err != nil {
				t.Fatalf("want success, got %v", err)
			}
			err := Notify(context.Background(), tt.arg)
			if !errors.Is(err, tt.want) {
				t.Errorf("want %v, got %v", tt.want, err)
			}
			if tt.want != nil && !strings.HasPrefix(err.Error(), "DummyEventHandler5: ") {
				t.Errorf("want named err, got %v", err.Error())
			}
		})
	}
}
