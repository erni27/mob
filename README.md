# mob

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/erni27/mob/CI?style=flat-square)](https://github.com/erni27/mob/actions?query=workflow%3ACI)
[![Go Report Card](https://goreportcard.com/badge/github.com/erni27/mob?style=flat-square)](https://goreportcard.com/report/github.com/erni27/mob)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.18-61CFDD.svg?style=flat-square)
[![GoDoc](https://pkg.go.dev/badge/mod/github.com/erni27/mob)](https://pkg.go.dev/mod/github.com/erni27/mob)

`mob` is a generic-based, simple **m**ediator / **ob**server library.

It supports in-process requests / events processing.

## Motivation

I was a bit tired of managing dependencies between handlers. Reusing them became the existential issue. That's how `mob` has been created. It solves complex dependency management by introducing a single communication point. The *mediator* part encapsulates request-response communication while the *observer* one acts as a *facade* focused on *observer* relationships. `mob` is conceptually similiar to [Event aggregator](https://martinfowler.com/eaaDev/EventAggregator.html) described by Martin Fowler.

`mob` supports two types of handlers - request handlers and event handlers.

## Request handlers

A request handler responses to a particular request.

Request handlers can be registered through the `RegisterRequestHandler` method.

```go
type DummyHandler struct{}

func(DummyHandler) Name() string {
    return "DummyHandler"
}

func (DummyHandler) Handle(ctx context.Context, req DummyRequest) (DummyResponse, error) {
    // Logic.
}

...

func main() {
    handler := DummyHandler{}
    if err := mob.RegisterRequestHandler[DummyRequest, DummyResponse](handler); err != nil {
        log.Fatalf("register handler %s: %v", handler.Name(), err)
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

## Event handlers

An event handler executes some logic in response to a dispatched event.

Event handlers can be registered through the `RegisterEventHandler` method.

```go
type DummyHandler struct{}

func(DummyHandler) Name() string {
    return "DummyHandler"
}

func (DummyHandler) Handle(ctx context.Context, req DummyRequest) error {
    // Logic.
}

...

func main() {
    handler := DummyHandler{}
    if err := mob.RegisterEventHandler[DummyRequest](handler); err != nil {
        log.Fatalf("register handler %s: %v", handler.Name(), err)
    }
}
```

A handler to register must satisfy the `EventHandler` interface. A request can have an arbitrary data type.

Event handlers are almost identical to the request ones. There are a few subtle differences though. An event handler does not return a response, only an error in case of failure. Unlike request ones, multiple handlers for a given request type can be registered. Be careful, `mob` doesn't check if a concrete handler is registered multiple times. Type alias declarations solves handler conflicts.

To notify all registered handlers about a certain event, call the `Notify` method.

```go
// Somewhere in your code.
err := mob.Notify(ctx, event)
```

`mob` executes all registered handlers concurrently. If at least one of them fails, an aggregate error containing all errors is returned.

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


`mob` is a convenient tool for applying *CQS*.

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

`mob` has some drawbacks. It makes an explicit communication implicit - in many cases a direct communication is much better than an indirect one. Also, where performance is a critical factor, you'd rather go with the explicit communication - it's always faster to call a handler directly.
