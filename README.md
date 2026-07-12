# go-actorlayer

`go-actorlayer` is a transport-agnostic Go actor runtime library extracted from
Balda and published as a standalone module:

```text
github.com/baldaworks/go-actorlayer
```

It owns generic envelopes, actor addressing, retry/error helpers, runtime lane
execution, registry dispatch, and transport-facing contracts. It does not own
Balda product policy, broker naming, durable storage, or channel/provider
semantics.

## Install

```bash
go get github.com/baldaworks/go-actorlayer@latest
```

## Packages

- `actorlayer`: envelope, actor address, payload, retry, and error
  classification primitives.
- `dispatch`: actor registration and exact/wildcard address resolution.
- `engine`: transport-agnostic delivery execution, lane serialization, retry
  handling, lifecycle events, and settlement.
- `transport`: small interfaces for dispatching commands, publishing events,
  consuming events, and draining transports.
- `transport/memory`: in-memory transport implementation for tests, examples,
  and lightweight standalone use.

Dependency direction:

```text
transport/memory -> engine -> dispatch -> actorlayer
transport/memory -> transport -> actorlayer
```

## Runtime model

```text
Envelope -> Delivery -> Runtime -> Registry -> Actor.Handle -> Ack/Retry/DeadLetter
```

An `Envelope` carries identity, addressing, idempotency, metadata, and an
encoded JSON payload. Transport adapters expose incoming messages as
`engine.Delivery` values. A runtime validates the envelope, resolves a lane key
for per-lane serialization, invokes the registered actor, then settles the
delivery through transport-owned `Ack`, `Retry`, or `DeadLetter` hooks.

The engine does not own broker behavior. Redelivery, durable storage, publish
semantics, and dead-letter persistence belong to the concrete transport
implementation.

## Extension points

Implement these interfaces to connect actorlayer to another queue or broker:

- `transport.Dispatcher` to send command envelopes.
- `engine.Source` to stream command deliveries into a runtime.
- `engine.Delivery` to expose message metadata and settlement operations.
- `transport.EventPublisher` and `transport.EventConsumer` for lifecycle or
  domain event streams.
- `transport.Drainer` for graceful transport shutdown.

Use `dispatch.Registry` when actor lookup needs a custom implementation. The
provided `dispatch.MemoryRegistry` is suitable for process-local actors and
supports wildcard fallback addresses such as `session:*`.

## Error and retry semantics

Actor errors are classified as transient, permanent, decode, policy,
actor-not-found, or external delivery failures. `actorlayer.IsRetryableError`
treats transient, actor-not-found, and external-delivery errors as retryable by
default; permanent, decode, and policy errors are terminal unless a host
supplies a different retry policy.

Runtime retry behavior is configured through `engine.RetryPolicy`. The policy
decides whether an error is retryable, how long to delay before retrying, and
when attempts are exhausted.

## Out of scope

Keep product and infrastructure policy outside this module:

- broker subjects, stream names, queue names, and headers;
- database schemas, task projections, and read models;
- channel/provider concepts such as Telegram, Slack, sessions, goals, or Balda
  users;
- app lifecycle wiring, observability policy, and configuration loading.

## Release

- CI: `.github/workflows/test.yml`
- GitHub release tags: `.github/workflows/release.yml`
- Maintainer checklist: [RELEASING.md](RELEASING.md)
