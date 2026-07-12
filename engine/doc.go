// Package engine provides transport-agnostic actor delivery execution.
//
// A Runtime validates envelopes, selects a lane, invokes the target actor, and
// settles each delivery through transport-owned Ack, Retry, or DeadLetter
// hooks.
//
// The engine owns execution flow. The transport owns broker behavior,
// redelivery, delay semantics, and dead-letter storage.
//
// InProgress is a delivery hook for hosts that own heartbeat cadence. Runtime
// exposes EmitInProgress for event publication but does not start a heartbeat
// loop on its own.
package engine
