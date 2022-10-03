package main

import (
	"context"
	"errors"
	"log"

	"github.com/erni27/mob"
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

func main() {
	ctx := context.Background()

	// Register an EchoRequestHandler to the global mob instance.
	if err := mob.RegisterRequestHandler[EchoRequest, EchoResponse](EchoRequestHandler{}); err != nil {
		log.Fatal(err)
	}
	// Send a request.
	res, err := mob.Send[EchoRequest, EchoResponse](ctx, "Hello world!")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)

	// Register a LogEventHandler to the global mob instance.
	if err := mob.RegisterEventHandler[LogEvent](LogEventHandler{}); err != nil {
		log.Fatal(err)
	}
	// Notify an occurance of an event.
	if err := mob.Notify[LogEvent](ctx, "Hello world!"); err != nil {
		log.Fatal(err)
	}
}
