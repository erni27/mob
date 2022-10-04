package mob

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAggregateHandlerError_Is(t *testing.T) {
	dummyErr := errors.New("dummy")
	tests := []struct {
		name   string
		aggr   AggregateHandlerError
		target error
		want   bool
	}{
		{
			name:   "err dummy within aggregate",
			aggr:   []error{fmt.Errorf("DummyHandler1: some error"), fmt.Errorf("DummyHandler2: %w", dummyErr)},
			target: dummyErr,
			want:   true,
		},
		{
			name:   "err dummy not in aggregate",
			aggr:   []error{fmt.Errorf("DummyHandler1: some error 1"), fmt.Errorf("DummyHandler2: some error 2")},
			target: dummyErr,
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
	var aggr AggregateHandlerError = []error{
		errors.New("error message 1"),
		errors.New("DummyHandler2: error message 2"),
		errors.New("DummyHandler3: error message 3"),
		errors.New("DummyHandler4: error message 4"),
	}
	got := aggr.Error()
	for _, err := range aggr {
		if !strings.Contains(got, err.Error()) {
			t.Errorf("aggregate msg should contain %s, got %s", err.Error(), got)
		}
	}
}

func clear() {
	m = New()
}
