package examples

import (
	"context"
	"errors"
	"log"
)

type EchoRequest string

type EchoResponse string

type EchoRequestHandler struct{}

func (h EchoRequestHandler) Handle(_ context.Context, req EchoRequest) (EchoResponse, error) {
	if req == "" {
		return "", errors.New("invalid request")
	}
	return (EchoResponse)(req), nil
}

type LogEvent string

type LogEventHandler struct{}

func (h LogEventHandler) Handle(_ context.Context, event LogEvent) error {
	if event == "" {
		return errors.New("invalid event")
	}
	log.Println(event)
	return nil
}

type DummyExecutor struct{}

type DummyRequest struct {
	ID string
}

type DummyResponse struct {
	ID   string
	Name string
}

func (e DummyExecutor) Execute(_ context.Context, req DummyRequest) (DummyResponse, error) {
	return DummyResponse{ID: req.ID, Name: "Dummy name."}, nil
}
