package main

import (
	"context"
	"log"

	"github.com/erni27/mob"
	"github.com/erni27/mob/examples"
)

func main() {
	// Initialize a new, standalone mob instance.
	m := mob.New()

	ctx := context.Background()

	// Register an EchoRequestHandler to the created mob instance.
	if err := mob.RegisterRequestHandlerTo[examples.EchoRequest, examples.EchoResponse](m, examples.EchoRequestHandler{}); err != nil {
		log.Fatal(err)
	}
	// Initialize a RequestSender with the created mob instance and send a request.
	eres, err := mob.NewRequestSender[examples.EchoRequest, examples.EchoResponse](m).Send(ctx, "Hello world!")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(eres)

	// Register an executor function as a RequestHandler to the created mob instance.
	executor := examples.DummyExecutor{}
	var hf mob.RequestHandlerFunc[examples.DummyRequest, examples.DummyResponse] = executor.Execute
	if err := mob.RegisterRequestHandlerTo[examples.DummyRequest, examples.DummyResponse](m, hf); err != nil {
		log.Fatal(err)
	}
	// Initialize a RequestSender with the created mob instance and send a request.
	dres, err := mob.NewRequestSender[examples.DummyRequest, examples.DummyResponse](m).Send(ctx, examples.DummyRequest{ID: "997"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(dres)

	// Register a LogEventHandler to the created mob instance.
	if err := mob.RegisterEventHandlerTo[examples.LogEvent](m, examples.LogEventHandler{}); err != nil {
		log.Fatal(err)
	}
	// Initialize an EventNotifier with the created mob instance and notify occurance of an event.
	if err := mob.NewEventNotifier[examples.LogEvent](m).Notify(ctx, "Hello world!"); err != nil {
		log.Fatal(err)
	}
}
