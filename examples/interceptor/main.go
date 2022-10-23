package main

import (
	"context"
	"log"

	"github.com/erni27/mob"
	"github.com/erni27/mob/examples"
)

func LoggingInterceptor(ctx context.Context, req interface{}, invoker mob.SendInvoker) (interface{}, error) {
	log.Printf("Starting. Request: %v\n", req)
	res, err := invoker(ctx, req)
	if err != nil {
		log.Printf("Error occured. Error: %v", err)
	}
	log.Printf("Ending. Response: %v\n", res)
	return res, nil
}

func main() {
	// Add LogginInterceptor to the global mob instance.
	mob.AddInterceptor(LoggingInterceptor)
	// Register EchoRequestHandler to the global mob instance.
	if err := mob.RegisterRequestHandler[examples.EchoRequest, examples.EchoResponse](examples.EchoRequestHandler{}); err != nil {
		log.Fatal(err)
	}
	res, err := mob.Send[examples.EchoRequest, examples.EchoResponse](context.Background(), "Hello. I'm looking for an Interceptor.")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
}
