// Package dispatch provides actor registration and address resolution helpers.
//
// MemoryRegistry stores process-local actors, normalizes addresses before
// lookup, resolves exact matches first, and then falls back to target wildcard
// addresses such as "session:*".
package dispatch
