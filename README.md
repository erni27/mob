# mob

`mob` is a generic-based, simple **m**ediator / **ob**server library.

It supports in-process requests / events processing.

## Motivation

I was a bit tired of managing dependencies between handlers. Reusing them became the existential issue. That's how `mob` has been created. It solves complex dependency management by introducing a single communication point. `mob` acts as a *facade*. The *mediator* part encapsulates request-response communication while the *observer* one acts as a *facade* focused on *observer* relationships.

`mob` supports two types of handlers - request handlers and event handlers.

## Request handlers

A request handler responses to a particular request.

## Event handlers

An event handler executes some logic in response to a dispatched event.