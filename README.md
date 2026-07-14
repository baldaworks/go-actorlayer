# go-actorlayer

`go-actorlayer` is a small Go library for message-driven application code.

Use it when your system has things like:

- chat or user sessions;
- background jobs;
- multi-step workflows;
- retries, delays, and dead letters;
- work that should continue after the original request is gone.

It gives you a small execution core without taking ownership of your broker,
headers, stream names, storage, or application semantics.

```text
github.com/baldaworks/go-actorlayer
```

## How to think about it

You send work as messages.

Each message has:

- an address, like `session:demo` or `job:task-42`;
- a payload;
- delivery behavior controlled by the transport.

A handler is registered for an address pattern. The runtime receives a message,
finds the right handler, runs it, and lets the transport decide whether that
message is acknowledged, retried, or dead-lettered.

You do not need to know actor theory to use it. In this library, an actor is
just a message handler with an address.

## When to use it

`go-actorlayer` is a good fit when you want:

- durable envelopes for work;
- explicit routing to handlers;
- exact or wildcard handler resolution;
- per-lane serialization;
- transport-owned `Ack`, `Retry`, and `DeadLetter`;
- a runtime core that you can adapt to your own queue or broker.

It works especially well for AI systems and agents, where one conversation,
task, or workflow may span multiple steps, retries, tool calls, cancellations,
or background runs.

It is not trying to be:

- a full application framework;
- a workflow product;
- a queue or broker;
- a persistence layer;
- a ready-made agent framework.

## Install

```bash
go get github.com/baldaworks/go-actorlayer@latest
```

## Quick start

This is the smallest end-to-end path:

1. define a handler;
2. register it;
3. start the runtime;
4. dispatch a message.

```go
package main

import (
	"context"
	"fmt"

	"github.com/baldaworks/go-actorlayer"
	"github.com/baldaworks/go-actorlayer/dispatch"
	"github.com/baldaworks/go-actorlayer/engine"
	"github.com/baldaworks/go-actorlayer/transport/memory"
)

type welcomePayload struct {
	Name string `json:"name"`
}

type welcomeActor struct {
	done chan struct{}
}

func (a welcomeActor) Address() string { return actorlayer.WildcardAddress("session") }

func (a welcomeActor) Handle(_ context.Context, env actorlayer.Envelope) error {
	var payload welcomePayload
	if err := actorlayer.UnmarshalPayload(env.Payload, &payload); err != nil {
		return err
	}
	fmt.Printf("welcome %s\n", payload.Name)
	close(a.done)
	return nil
}

func main() {
	bus := memory.New(1)
	registry := dispatch.NewMemoryRegistry()
	done := make(chan struct{})
	_ = registry.Register(welcomeActor{done: done})

	runtime, _ := engine.NewDispatchRuntime(engine.RuntimeConfig{
		Registry:  registry,
		AddressOf: func(env engine.Envelope) (string, error) { return env.To.String() },
		Retry: engine.RetryPolicy{
			IsRetryable: actorlayer.IsRetryableError,
			Backoff:     actorlayer.RetryDelay,
		},
	})

	ctx := context.Background()
	runDone := make(chan error, 1)
	go func() {
		runDone <- runtime.Run(ctx, bus)
	}()

	payload, _ := actorlayer.MarshalPayload(welcomePayload{Name: "ada"})
	_, _ = bus.Dispatch(ctx, actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "example.command",
		Kind:      "welcome",
		From:      actorlayer.SystemAddress("example"),
		To:        actorlayer.ActorAddress{Target: "session", Key: "demo"},
		Payload:   payload,
	})

	<-done
	_ = bus.Drain(ctx)
	<-runDone
}
```

Output:

```text
welcome ada
```

The same example is covered by [example_test.go](example_test.go).

## Packages

- `actorlayer`: envelope types, addresses, payload helpers, codecs, errors,
  retry helpers.
- `dispatch`: registration and lookup for handlers.
- `engine`: the runtime that executes deliveries.
- `transport`: interfaces for dispatching, consuming, publishing events, and
  draining.
- `transport/memory`: an in-memory transport for tests and examples.

Dependency direction:

```text
transport/memory -> engine -> dispatch -> actorlayer
transport/memory -> transport -> actorlayer
```

## Mental model

```text
Envelope -> Delivery -> Runtime -> Registry -> Actor.Handle -> Ack/Retry/DeadLetter
```

- an `Envelope` is the unit of work;
- a `Delivery` is the transport wrapper around that work;
- the `Runtime` validates the message and decides where it runs;
- the `Registry` finds the right handler;
- the transport decides what `Ack`, `Retry`, and `DeadLetter` mean.

That split is deliberate:

- the runtime owns execution flow;
- the transport owns broker behavior;
- the application owns business logic, headers, persistence, and policy.

## Envelope and payloads

An envelope carries:

- identity: `ID`, `CorrelationID`, `CausationID`, `DedupeKey`;
- addressing: `From`, `To`, optional `ReportTo`;
- execution hints: `Priority`, `Attempt`, `MaxAttempts`, `NotBefore`,
  `ExpiresAt`;
- payload: `Payload`;
- generic metadata: `Meta`.

`Payload` is format-neutral:

- `Encoding` identifies the codec, for example `json`;
- `Data` contains the raw bytes.

In most code you should use:

- `MarshalPayload(v)` to build a payload with the default JSON codec;
- `UnmarshalPayload(payload, &dst)` to decode it.

If you need custom encodings, use:

- `Codec`
- `CodecRegistry`
- `NewCodecRegistry`
- `MarshalPayloadWithCodec`
- `UnmarshalPayloadWithRegistry`
- `EncodeEnvelopeWithRegistry`
- `DecodeEnvelopeWithRegistry`

The built-in default is `JSONCodec`. On the wire, JSON payloads keep the legacy
`payload_json` projection for compatibility, while non-JSON payloads use
`payload_encoding` plus base64-encoded `payload_data`.

Custom encodings are first-class in the codec-aware marshal, unmarshal, and
envelope encode/decode paths above. The default helper path still uses the
built-in registry, so raw payload construction through `NewPayload` only accepts
encodings registered there.

## Addressing

Addresses look like `target:key`.

Examples:

- `session:demo`
- `job:task-42`
- `delivery:user-42`

Helpful rules:

- `ActorAddress.String()` validates concrete addresses.
- `WildcardAddress("session")` produces `session:*`.
- `dispatch.MemoryRegistry` resolves exact matches first, then target wildcard.

Registry and engine dispatch paths normalize addresses case-insensitively.

## Errors and retries

The library distinguishes:

- `transient`
- `policy`
- `permanent`
- `decode`
- `external_delivery`

Use:

- `actorlayer.TransientError(err)`
- `actorlayer.PolicyError(err)`
- `actorlayer.PermanentError(err)`
- `actorlayer.DecodeError(err)`
- `actorlayer.ExternalDeliveryError(err)`

Most hosts should use `actorlayer.IsRetryableError` as the default rule:

- retryable: transient, external delivery;
- terminal: policy, permanent, decode, actor not found.

`engine.RetryPolicy` lets you define:

- which errors retry;
- how backoff works;
- when retries are exhausted.

## Extension points

To integrate with a real queue or broker, implement some or all of:

- `transport.Dispatcher`
- `engine.Source`
- `engine.Delivery`
- `transport.EventPublisher`
- `transport.EventConsumer`
- `transport.Drainer`

You can also replace `dispatch.Registry` if process-local in-memory registration
is not enough.

## Out of scope

This module intentionally does not define:

- broker subjects, headers, streams, or queue names;
- database schemas or read models;
- product concepts such as users, sessions, goals, or channels;
- config loading or application lifecycle wiring;
- observability policy.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Release

- CI: `.github/workflows/test.yml`
- GitHub release tags: `.github/workflows/release.yml`
- Maintainer checklist: [RELEASING.md](RELEASING.md)
