// Package transport defines actorlayer transport interfaces.
//
// Implementations dispatch durable envelopes, publish or consume envelope
// events, and expose graceful drain behavior.
//
// Transport adapters are expected to validate envelopes before sending them
// across a broker or other process boundary.
package transport
