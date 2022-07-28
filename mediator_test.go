package mob

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

type DummyRequest1 struct {
	String string
}

type DummyResponse1 struct {
	String string
	Int    int
	Bool   bool
	Time   time.Time
}

type DummyRequestHandler1 struct {
	handleFunc func(context.Context, DummyRequest1) (DummyResponse1, error)
}

func (DummyRequestHandler1) Name() string {
	return "DummyHandler1"
}

func (h DummyRequestHandler1) Handle(ctx context.Context, req DummyRequest1) (DummyResponse1, error) {
	return h.handleFunc(ctx, req)
}

type DummyDuplicateRequestHandler1 struct{}

func (*DummyDuplicateRequestHandler1) Name() string {
	return "DummyDuplicate1"
}

func (*DummyDuplicateRequestHandler1) Handle(_ context.Context, _ DummyRequest1) (DummyResponse1, error) {
	return DummyResponse1{}, nil
}

type DummyRequest2 struct {
	Int int
}

type DummyResponse2 struct {
	String string
	Int    int
	Bool   bool
	Time   time.Time
	Float  float32
}

type DummyRequestHandler2 struct{}

func (*DummyRequestHandler2) Name() string {
	return "DummyHandler2"
}

func (*DummyRequestHandler2) Handle(_ context.Context, _ DummyRequest2) (DummyResponse2, error) {
	return DummyResponse2{}, nil
}

type DummyRequestHandler3 struct{}

func (*DummyRequestHandler3) Name() string {
	return "DummyHandler3"
}

func (*DummyRequestHandler3) Handle(_ context.Context, _ DummyRequest1) (DummyResponse2, error) {
	return DummyResponse2{}, nil
}

type DummyRequestHandler4 struct{}

func (DummyRequestHandler4) Name() string {
	return "DummyHandler4"
}

func (DummyRequestHandler4) Handle(_ context.Context, _ DummyRequest2) (DummyResponse1, error) {
	return DummyResponse1{}, nil
}

func TestRegisterRequestHandler_DuplicateHandler(t *testing.T) {
	defer clearRequestHandlers()
	tests := []struct {
		name string
		arg  RequestHandler[DummyRequest1, DummyResponse1]
		want error
	}{
		{
			name: "dummy handler",
			arg:  DummyRequestHandler1{},
			want: ErrDuplicateHandler,
		},
		{
			name: "dummy duplicate",
			arg:  &DummyDuplicateRequestHandler1{},
			want: ErrDuplicateHandler,
		},
	}
	if err := RegisterRequestHandler[DummyRequest1, DummyResponse1](DummyRequestHandler1{}); err != nil {
		t.Fatalf("register first handler: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterRequestHandler(tt.arg); err != tt.want {
				t.Errorf("want %v, got error %v", tt.want, err)
			}
		})
	}
}

func TestRegisterRequestHandler_InvalidHandler(t *testing.T) {
	defer clearRequestHandlers()
	tests := []struct {
		name string
		arg  RequestHandler[DummyRequest1, DummyResponse1]
		want error
	}{
		{
			name: "nil interface",
			arg:  nil,
			want: ErrInvalidHandler,
		},
		{
			name: "nil value",
			arg:  (*DummyDuplicateRequestHandler1)(nil),
			want: ErrInvalidHandler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RegisterRequestHandler(tt.arg); err != tt.want {
				t.Errorf("want %v, got error %v", tt.want, err)
			}
		})
	}
}

func TestRegisterRequestHandler(t *testing.T) {
	defer clearRequestHandlers()
	t.Run("dummy handler 1", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest1, DummyResponse1](DummyRequestHandler1{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 2", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest2, DummyResponse2](&DummyRequestHandler2{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 3", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest1, DummyResponse2](&DummyRequestHandler3{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 4", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest2, DummyResponse1](&DummyRequestHandler4{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
}

func TestSend_HandlerNotFound(t *testing.T) {
	if _, err := Send[DummyRequest1, DummyResponse1](context.Background(), DummyRequest1{}); err != ErrHandlerNotFound {
		t.Errorf("want error %v, got %v", ErrHandlerNotFound, err)
	}
}

func TestSend(t *testing.T) {
	errDummy := errors.New("dummy error")
	now := time.Now()
	tests := []struct {
		name    string
		arg     DummyRequest1
		handle  func(context.Context, DummyRequest1) (DummyResponse1, error)
		want    DummyResponse1
		wantErr error
	}{
		{
			name: "success",
			arg:  DummyRequest1{String: "dummy string"},
			handle: func(_ context.Context, req DummyRequest1) (DummyResponse1, error) {
				return DummyResponse1{
					String: req.String,
					Int:    997,
					Bool:   true,
					Time:   now,
				}, nil
			},
			want: DummyResponse1{
				String: "dummy string",
				Int:    997,
				Bool:   true,
				Time:   now,
			},
			wantErr: nil,
		},
		{
			name: "handler error",
			arg:  DummyRequest1{String: "dummy string"},
			handle: func(_ context.Context, req DummyRequest1) (DummyResponse1, error) {
				return DummyResponse1{}, errDummy
			},
			want:    DummyResponse1{},
			wantErr: errDummy,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer clearRequestHandlers()
			var handler RequestHandler[DummyRequest1, DummyResponse1] = DummyRequestHandler1{handleFunc: tt.handle}
			if err := RegisterRequestHandler(handler); err != nil {
				t.Fatalf("register handler: %v", err)
			}
			got, err := Send[DummyRequest1, DummyResponse1](context.Background(), tt.arg)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want err %v, got %v", tt.wantErr, err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func clearRequestHandlers() {
	rhandlers = map[reqHnKey]interface{}{}
}
