package main

import (
	"context"
	"log"

	"github.com/erni27/mob"
	"github.com/erni27/mob/examples"
)

func main() {
	ctx := context.Background()

	// Register an EchoRequestHandler to the global mob instance.
	if err := mob.RegisterRequestHandler[examples.EchoRequest, examples.EchoResponse](examples.EchoRequestHandler{}); err != nil {
		log.Fatal(err)
	}
	// Send a request.
	eres, err := mob.Send[examples.EchoRequest, examples.EchoResponse](ctx, "Hello world!")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(eres)

	// Register an executor function as a RequestHandler to the global mob instance.
	executor := examples.DummyExecutor{}
	var hf mob.RequestHandlerFunc[examples.DummyRequest, examples.DummyResponse] = executor.Execute
	if err := mob.RegisterRequestHandler[examples.DummyRequest, examples.DummyResponse](hf); err != nil {
		log.Fatal(err)
	}
	// Send a request.
	dres, err := mob.Send[examples.DummyRequest, examples.DummyResponse](ctx, examples.DummyRequest{ID: "997"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println(dres)

	// Register a LogEventHandler to the global mob instance.
	if err := mob.RegisterEventHandler[examples.LogEvent](examples.LogEventHandler{}); err != nil {
		log.Fatal(err)
	}
	// Notify an occurance of an event.
	if err := mob.Notify[examples.LogEvent](ctx, "Hello world!"); err != nil {
		log.Fatal(err)
	}
}
