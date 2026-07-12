package engine

import "github.com/baldaworks/go-actorlayer"

type ActorAddress = actorlayer.ActorAddress
type Envelope = actorlayer.Envelope

// SystemAddress returns an actorlayer system address.
func SystemAddress(key string) ActorAddress {
	return actorlayer.SystemAddress(key)
}

// WildcardAddress returns the normalized wildcard address for a target.
func WildcardAddress(target string) string {
	return actorlayer.WildcardAddress(target)
}

// EncodeEnvelope validates and marshals an envelope as JSON.
func EncodeEnvelope(e Envelope) (string, error) {
	return actorlayer.EncodeEnvelope(e)
}

// EncodeEnvelopeWithRegistry validates and marshals an envelope with the provided
// registry.
func EncodeEnvelopeWithRegistry(e Envelope, reg actorlayer.CodecRegistry) (string, error) {
	return actorlayer.EncodeEnvelopeWithRegistry(e, reg)
}

// DecodeEnvelope unmarshals and validates an envelope JSON string.
func DecodeEnvelope(raw string) (Envelope, error) {
	return actorlayer.DecodeEnvelope(raw)
}

// DecodeEnvelopeWithRegistry unmarshals and validates an envelope string with the
// provided registry.
func DecodeEnvelopeWithRegistry(raw string, reg actorlayer.CodecRegistry) (Envelope, error) {
	return actorlayer.DecodeEnvelopeWithRegistry(raw, reg)
}
