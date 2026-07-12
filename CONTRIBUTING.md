# Contributing to go-actorlayer

## Scope

`go-actorlayer` is a generic library.

Keep product and infrastructure policy out of this module. In particular, do
not add:

- app-specific concepts such as users, sessions, goals, channels, or tenants;
- broker subjects, headers, stream names, or queue naming policy;
- persistence schemas, read models, or database migrations;
- config loading, process lifecycle wiring, or observability policy.

This repository owns generic:

- envelopes;
- payload codecs;
- actor addressing;
- dispatch and wildcard resolution;
- runtime execution and retry hooks;
- transport-facing interfaces;
- in-memory transport for tests and examples.

## Development setup

The main validation command is:

```bash
go test ./...
```

Run it from the module root before sending a change.

## API changes

This is a library. Public API changes need extra care.

Before changing exported behavior, check:

- does the change alter wire compatibility?
- does it change retry or settlement semantics?
- does it change how addresses are normalized or resolved?
- does it change payload encoding behavior?
- can the same goal be achieved without expanding the generic surface area?

If a breaking change is intentional, make it explicit in:

- the code;
- tests;
- `README.md`;
- `CHANGELOG.md`.

## Wire compatibility

Envelope encoding and decoding are part of the library contract.

Today:

- JSON payloads keep the legacy `payload_json` wire projection;
- non-JSON payloads use `payload_encoding` and base64-encoded `payload_data`.

Do not change wire behavior casually. If you touch encoding or decoding, add or
update tests for compatibility behavior.

## Examples and docs

Examples are part of the product surface.

When changing docs or examples:

- keep `README.md`, package docs, and `example_test.go` aligned;
- prefer examples that compile and run under `go test`;
- use the same terminology across README, package docs, and examples;
- optimize for a reader who may be new to actor-style systems.

If you add a feature that users should discover quickly, show it in one of:

- `README.md`
- `example_test.go`
- a package-level example in the relevant subpackage

## Tests

Prefer tests that verify behavior, not implementation details.

Important areas to cover:

- envelope validation and codec behavior;
- encode/decode compatibility;
- retry classification;
- actor resolution;
- runtime settlement behavior;
- transport behavior in `transport/memory`.

If a change affects examples, make sure example tests still pass.

## Style

Keep the package small and unsurprising.

Prefer:

- explicit contracts;
- narrow interfaces;
- transport-agnostic naming;
- generic terminology over product terminology;
- documentation that explains practical use before deep details.

Avoid adding abstraction layers unless they simplify the public model.

## Release notes

If your change affects users, update `CHANGELOG.md`.

For release mechanics, see [RELEASING.md](RELEASING.md).
