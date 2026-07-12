// Package memory provides an in-memory actorlayer transport implementation.
//
// The transport is intended for tests, examples, and local standalone use. It
// implements command dispatch and source behavior, event publish and consume
// behavior, and graceful drain without introducing broker dependencies.
package memory
