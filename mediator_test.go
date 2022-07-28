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

type DummyHandler1 struct {
	handleFunc func(context.Context, DummyRequest1) (DummyResponse1, error)
}

func (DummyHandler1) Name() string {
	return "DummyHandler1"
}

func (h DummyHandler1) Handle(ctx context.Context, req DummyRequest1) (DummyResponse1, error) {
	return h.handleFunc(ctx, req)
}

type DummyDuplicate1 struct{}

func (*DummyDuplicate1) Name() string {
	return "DummyDuplicate1"
}

func (*DummyDuplicate1) Handle(_ context.Context, _ DummyRequest1) (DummyResponse1, error) {
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

type DummyHandler2 struct{}

func (*DummyHandler2) Name() string {
	return "DummyHandler2"
}

func (*DummyHandler2) Handle(_ context.Context, _ DummyRequest2) (DummyResponse2, error) {
	return DummyResponse2{}, nil
}

type DummyHandler3 struct{}

func (*DummyHandler3) Name() string {
	return "DummyHandler3"
}

func (*DummyHandler3) Handle(_ context.Context, _ DummyRequest1) (DummyResponse2, error) {
	return DummyResponse2{}, nil
}

type DummyHandler4 struct{}

func (DummyHandler4) Name() string {
	return "DummyHandler4"
}

func (DummyHandler4) Handle(_ context.Context, _ DummyRequest2) (DummyResponse1, error) {
	return DummyResponse1{}, nil
}

func TestRegisterRequestHandler_DuplicateHandler(t *testing.T) {
	defer clean()
	tests := []struct {
		name string
		arg  RequestHandler[DummyRequest1, DummyResponse1]
		want error
	}{
		{
			name: "dummy handler",
			arg:  DummyHandler1{},
			want: ErrDuplicateHandler,
		},
		{
			name: "dummy duplicate",
			arg:  &DummyDuplicate1{},
			want: ErrDuplicateHandler,
		},
	}
	if err := RegisterRequestHandler[DummyRequest1, DummyResponse1](DummyHandler1{}); err != nil {
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
	defer clean()
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
			arg:  (*DummyDuplicate1)(nil),
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
	defer clean()
	t.Run("dummy handler 1", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest1, DummyResponse1](DummyHandler1{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 2", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest2, DummyResponse2](&DummyHandler2{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 3", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest1, DummyResponse2](&DummyHandler3{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
	t.Run("dummy handler 4", func(t *testing.T) {
		if err := RegisterRequestHandler[DummyRequest2, DummyResponse1](&DummyHandler4{}); err != nil {
			t.Errorf("want success, got error %v", err)
		}
	})
}

func clean() {
	rhandlers = map[reqHnKey]interface{}{}
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
			defer clean()
			var handler RequestHandler[DummyRequest1, DummyResponse1] = DummyHandler1{handleFunc: tt.handle}
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

func TestSend_HandlerNotFound(t *testing.T) {
	if _, err := Send[DummyRequest1, DummyResponse1](context.Background(), DummyRequest1{}); err != ErrHandlerNotFound {
		t.Errorf("want error %v, got %v", ErrHandlerNotFound, err)
	}
}
