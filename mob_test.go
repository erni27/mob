package mob

import (
	"errors"
	"strings"
	"testing"
)

func TestAggregateHandlerError_Is(t *testing.T) {
	errDummy := errors.New("dummy")
	tests := []struct {
		name   string
		aggr   AggregateHandlerError
		target error
		want   bool
	}{
		{
			name:   "err dummy within aggregate",
			aggr:   []HandlerError{{Handler: "DummyHandler1", Err: errors.New("some error")}, {Handler: "DummyHandler2", Err: errDummy}},
			target: errDummy,
			want:   true,
		},
		{
			name:   "err dummy not in aggregate",
			aggr:   []HandlerError{{Handler: "DummyHandler1", Err: errors.New("some error 1")}, {Handler: "DummyHandler2", Err: errors.New("some error 2")}},
			target: errDummy,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.aggr.Is(tt.target); got != tt.want {
				t.Errorf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestAggregateHandlerError_Error(t *testing.T) {
	var aggr AggregateHandlerError = []HandlerError{
		{Handler: "DummyHandler1", Err: errors.New("error message 1")},
		{Handler: "DummyHandler2", Err: errors.New("error message 2")},
		{Handler: "DummyHandler3", Err: errors.New("error message 3")},
		{Handler: "DummyHandler4", Err: errors.New("error message 4")},
	}
	got := aggr.Error()
	for _, err := range aggr {
		if !strings.Contains(got, err.Error()) {
			t.Errorf("aggregate msg should contain %s, got %s", err.Error(), got)
		}
	}
}
