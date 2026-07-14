// Package actorlayer provides the core types for a small transport-agnostic
// actor runtime.
//
// The package defines the durable Envelope model, actor addresses, payload
// helpers, error classification, and retry helpers used by the higher-level
// dispatch, engine, and transport packages.
//
// Package layout:
//   - actorlayer: envelope model, addresses, payload helpers, errors, retry
//   - dispatch: actor registration and address resolution
//   - engine: delivery execution, lane serialization, settlement hooks
//   - transport: dispatch, event, and drain contracts
//   - transport/memory: in-memory transport for tests, examples, and local use
//
// Payload carries an explicit encoding plus raw bytes. Use MarshalPayload and
// UnmarshalPayload to encode and decode typed payloads. The default helpers use
// the built-in JSON codec unless a codec-aware helper is used explicitly.
// ReportTo is optional, but when present it must be a valid actor address.
package actorlayer
