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

func (h EchoRequestHandler) Name() string {
	return "EchoRequestHandler"
}

func (h EchoRequestHandler) Handle(_ context.Context, req EchoRequest) (EchoResponse, error) {
	if req == "" {
		return "", errors.New("invalid request")
	}
	return (EchoResponse)(req), nil
}

type LogEvent string

type LogEventHandler struct{}

func (h LogEventHandler) Name() string {
	return "LogEventHandler"
}

func (h LogEventHandler) Handle(_ context.Context, event LogEvent) error {
	if event == "" {
		return errors.New("invalid event")
	}
	log.Println(event)
	return nil
}

func main() {
	// Initialize a new mob instance.
	m := mob.New()

	ctx := context.Background()

	// Register an EchoRequestHandler to the created mob instance.
	if err := mob.RegisterRequestHandlerTo[EchoRequest, EchoResponse](m, EchoRequestHandler{}); err != nil {
		log.Fatal(err)
	}
	// Initialize a RequestSender with the created mob instance and send a request.
	res, err := mob.NewRequestSender[EchoRequest, EchoResponse](m).Send(ctx, "Hello world!")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)

	// Register a LogEventHandler to the created mob instance.
	if err := mob.RegisterEventHandlerTo[LogEvent](m, LogEventHandler{}); err != nil {
		log.Fatal(err)
	}
	// Initialize an EventNotifier with the created mob instance and notify occurance of an event.
	if err := mob.NewEventNotifier[LogEvent](m).Notify(ctx, "Hello world!"); err != nil {
		log.Fatal(err)
	}
}
