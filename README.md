# mob

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/erni27/mob/ci?style=flat-square)](https://github.com/erni27/mob/actions?query=workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/erni27/mob)](https://goreportcard.com/report/github.com/erni27/mob)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![GoDoc](https://pkg.go.dev/badge/mod/github.com/erni27/mob)](https://pkg.go.dev/mod/github.com/erni27/mob)
[![Coverage Status](https://codecov.io/gh/erni27/mob/branch/master/graph/badge.svg)](https://codecov.io/gh/erni27/mob)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)

`mob` is a generic-based, simple **m**ediator / **ob**server (event aggregator) library.

It supports in-process requests / events processing.

## Motivation

I was a bit tired of managing dependencies between handlers. Reusing them became the existential issue. That's how `mob` has been created. It solves complex dependency management by introducing a single communication point. The *mediator* part encapsulates request-response communication while the *observer* one acts as a *facade* focused on *observer* relationships. `mob` is conceptually similiar to [Event aggregator](https://martinfowler.com/eaaDev/EventAggregator.html) described by Martin Fowler.

`mob` supports two types of handlers - request handlers and event handlers.

## Request handlers

A request handler responses to a particular request.

Request handlers can be registered through the `RegisterRequestHandler` method.

```go
type DummyHandler struct{}

func (DummyHandler) Handle(ctx context.Context, req DummyRequest) (DummyResponse, error) {
    // Logic.
}

...

func main() {
    handler := DummyHandler{}
    if err := mob.RegisterRequestHandler[DummyRequest, DummyResponse](handler); err != nil {
        log.Fatalf("register handler: %v", err)
    }
}
```

A handler to register must satisfy the `RequestHandler` interface. Both request and response can have arbitrary data types.

Only one handler for a particular request-response pair can be registered. To avoid handlers conflicts use type alias declarations.

To send a request and get a response simply call the `Send` method.

```go
// Somewhere in your code.
response, err := mob.Send[DummyRequest, DummyResponse](ctx, req)
```

If a handler does not exist for a given request - response pair - `ErrHandlerNotFound` is returned.

## Interceptors

The processing can get complex, especially when building large, enterprise systems. It's necessary to add many cross-cutting concerns like logging, monitoring, validations or security. To make it simple, `mob` supports `Interceptor`s. `Interceptor`s allow to intercept an invocation of `Send` method so they offer a way to enrich the request-response processing pipeline (basically apply decorators).

`Interceptor`s can be added to `mob` by calling `AddInterceptor` method.

```go
mob.AddInterceptor(LoggingInterceptor)
```

`Interceptor`s are invoked in order they're added to the chain.

For more information on how to create and use `Interceptor`s, see the [example](https://github.com/erni27/mob/blob/master/examples/interceptor/main.go).

## Event handlers

An event handler executes some logic in response to a dispatched event.

Event handlers can be registered through the `RegisterEventHandler` method.

```go
type DummyHandler struct{}

func (DummyHandler) Handle(ctx context.Context, req DummyRequest) error {
    // Logic.
}

...

func main() {
    handler := DummyHandler{}
    if err := mob.RegisterEventHandler[DummyRequest](handler); err != nil {
        log.Fatalf("register handler: %v", err)
    }
}
```

A handler to register must satisfy the `EventHandler` interface. A request can have an arbitrary data type.

Event handlers are almost identical to the request ones. There are a few subtle differences though. An event handler does not return a response, only an error in case of failure. Unlike request ones, multiple handlers for a given request type can be registered. Be careful, `mob` doesn't check if a concrete handler is registered multiple times. Type alias declarations solves handler conflicts.

To notify all registered handlers about a certain event call the `Notify` method.

```go
// Somewhere in your code.
err := mob.Notify(ctx, event)
```

`mob` executes all registered handlers concurrently. If at least one of them fails, an aggregate error containing all errors is returned.

## Named handlers

It's recommended to register a handler with a meaningful name. `WithName` is used to return an `Option` that associates a given name with a handler.

```go
err := mob.RegisterEventHandler[LogEvent](LogEventHandler{}, mob.WithName("LogEventHandler"));
```

It helps debugging potential issues. Extremely useful when multiple event handlers are registered to the specific subject and there is a need to communicate which handler fails. `mob` prefixes all errors by a handler's name if configured.

## Register ordinary functions as handlers

`mob` exports both `RequestHandlerFunc` and `EventHandlerFunc` that act as adapters to allow the use of ordinary functions (and structs' methods) as request and event handlers.

```go
var hf mob.RequestHandlerFunc[DummyRequest, DummyResponse] = func(ctx context.Context, req DummyRequest) (DummyResponse, error) {
    // Your logic goes here.
}
err := mob.RegisterRequestHandler[DummyRequest, DummyResponse](hf)
```

## Concurrency

`mob` is a concurrent-safe library for multiple requests and events processing. But you shouldn't mix handlers' registration  with requests or events processing. `mob` assumes that clients register their handlers during the initialization process and after first request or event is processed - no handler is registered.

## Use cases

There are many use cases for `mob`. Everytime when there is a burden of dependency management, `mob` can become a useful friend.

There are two cases where I find `mob` extremely useful.

The first one is to slim the application layer API handlers. `mob` centralizes control so there is no need to use DI. It makes the components more portable.

The following example shows one of the most popular kind of the application layers handlers - HTTP handlers.

*Classic way*

```go
func GetUserHandler(u UserGetter) http.HandlerFunc {
    return func(rw http.ResponseWriter, req *http.Request) {
        var dureq DummyUserRequest
        _ = json.NewDecoder(req.Body).Decode(&dureq)
        res, _ := u.Get(req.Context(), dureq)
        rw.Header().Set("content-type", "application/json")
        rw.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(rw).Encode(res)
    }
}
```

*`mob` way*

```go
func GetUser(rw http.ResponseWriter, req *http.Request) {
    var dureq DummyUserRequest
    _ = json.NewDecoder(req.Body).Decode(&dureq)
    res, _ := mob.Send[DummyUserRequest, DummyUserResponse](req.Context(), dureq)
    rw.Header().Set("content-type", "application/json")
    rw.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(rw).Encode(res)
}
```


`mob` is a convenient tool for applying *CQS* and *CQRS*.

`mob` also makes it easier to take advantage of any kind of in-process, event-based communication. A domain event processing is a great example.

*Classic way*

```go
func (s *UserService) UpdateEmail(ctx context.Context, id string, email string) error {
    u, _ := s.Repository.GetUser(ctx, id)
    u.Email = email
    _ = s.Repository.UpdateUser(ctx, u)
    _ = s.ContactBookService.RefreshContactBook(ctx)
    _ = s.NewsletterService.RefreshNewsletterContactInformation(ctx)
    // Do more side-effect actions in response to the email changed event.
    return nil
}
```

*`mob` way*

```go
func (s *UserService) UpdateEmail(ctx context.Context, id string, email string) error {
    u, _ := s.Repository.GetUser(ctx, id)
    u.Email = email
    _ = s.Repository.UpdateUser(ctx, u)
    _ = mob.Notify(ctx, EmailChanged{UserID: id, Email: email})
    return nil
}
```

For more information on how to use the global mob instance, see the [example](https://github.com/erni27/mob/blob/master/examples/global/main.go).

## Multiple mobs

All previous examples correspond to the global mob (singleton based approach).

Although, `mob` itself acts as a global handlers registry. It is possible to configure as many as mobs (so multiple mob instances) as you want. Each mob instance acts as a separate handlers registry. `mob` package uses slightly different API to support multiple mob instances (mostly due to currently supported generic model which doesn't allow method type parameters).

To initialise a new, standalone mob instance use the `New` method.

```go
m := mob.New()
```

`RegisterRequestHandlerTo` is used to register a request handler to the standalone mob instance. Pass the mob instance as a first function parameter followed by a handler to register.

```go
err := mob.RegisterRequestHandlerTo[EchoRequest, EchoResponse](m, EchoRequestHandler{})
```

Because current `Go` design doesn't support the method having type parameters, `mob` uses facilitators to get advantage of mob's generic behaviour. Creating a `RequestSender` tied to a standalone mob instance must precede sending a request through the `Send` method.

```go
res, err := mob.NewRequestSender[EchoRequest, EchoResponse](m).Send(ctx, "Hello world!")
```

Working with event handlers is similiar.

To register an event handler call the `RegisterEventHandlerTo`.

```go
err := mob.RegisterEventHandlerTo[LogEvent](m, LogEventHandler{});
```

In order to notify an occurance of an event create an `EventNotifier` tied to a standalone mob instance and then call the `Notify` method.

```go
err := mob.NewEventNotifier[LogEvent](m).Notify(ctx, "Hello world!")
```

To add an `Interceptor` to a standalone mob instance call the `AddInterceptorTo` method.

```go
mob.AddInterceptorTo(m, LoggingInterceptor)
```

`mob` package keep track only of the global mob instance. It means that users are responsible for keeping track of the multiple, standalone mob instances.

For more information on how to create and use a standalone mob instance, see the [example](https://github.com/erni27/mob/blob/master/examples/standalone/main.go).

## Conclusion

Although `mob` can be exteremely useful. It has some drawbacks. It makes an explicit communication implicit - in many cases a direct communication is much better than an indirect one. Especially when it obscures your domain.
