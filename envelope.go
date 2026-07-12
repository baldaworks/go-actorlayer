package actorlayer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const EncodingJSON = "json"

// Codec marshals and unmarshals actor payload values for one encoding.
type Codec interface {
	Encoding() string
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, v any) error
	Validate(data []byte) error
}

// CodecRegistry resolves codecs by encoding name.
type CodecRegistry interface {
	Lookup(encoding string) (Codec, bool)
}

type codecRegistry struct {
	codecs map[string]Codec
}

// NewCodecRegistry builds a registry from the provided codecs.
func NewCodecRegistry(codecs ...Codec) CodecRegistry {
	out := codecRegistry{codecs: make(map[string]Codec, len(codecs))}
	for _, codec := range codecs {
		if codec == nil {
			continue
		}
		encoding := normalizeEncoding(codec.Encoding())
		if encoding == "" {
			continue
		}
		out.codecs[encoding] = codec
	}
	return out
}

// DefaultCodecRegistry returns the built-in codec registry.
func DefaultCodecRegistry() CodecRegistry {
	return defaultCodecRegistry
}

type jsonCodec struct{}

func (jsonCodec) Encoding() string {
	return EncodingJSON
}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Validate(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("payload must be valid json")
	}
	return nil
}

// JSONCodec is the default actorlayer codec and preserves current wire/storage
// compatibility.
var JSONCodec Codec = jsonCodec{}

var defaultCodecRegistry = NewCodecRegistry(JSONCodec)

// ActorAddress identifies an actor target and concrete key.
type ActorAddress struct {
	Target string `json:"target"`
	Key    string `json:"key"`
}

// SystemAddress returns an address in the reserved "system" target.
func SystemAddress(key string) ActorAddress {
	return ActorAddress{Target: "system", Key: key}
}

// WildcardAddress returns the normalized registry address for all keys in a
// target, such as "session:*".
func WildcardAddress(target string) string {
	return strings.ToLower(strings.TrimSpace(target)) + ":*"
}

// String returns the normalized full actor address.
func (a ActorAddress) String() (string, error) {
	target := strings.ToLower(strings.TrimSpace(a.Target))
	key := strings.TrimSpace(a.Key)
	if target == "" {
		return "", fmt.Errorf("actor target is required")
	}
	if key == "" {
		return "", fmt.Errorf("actor key is required")
	}
	return target + ":" + key, nil
}

// Payload is the format-neutral envelope body.
type Payload struct {
	Encoding string
	Data     []byte
}

// Validate verifies the payload shape and codec-level validity.
func (p Payload) Validate() error {
	return p.ValidateWithRegistry(DefaultCodecRegistry())
}

// ValidateWithRegistry verifies the payload using the provided registry.
func (p Payload) ValidateWithRegistry(reg CodecRegistry) error {
	if normalizeEncoding(p.Encoding) == "" {
		return fmt.Errorf("payload encoding is required")
	}
	if len(p.Data) == 0 {
		return fmt.Errorf("payload data is required")
	}
	codec, ok := normalizeRegistry(reg).Lookup(p.Encoding)
	if !ok {
		return fmt.Errorf("payload codec %q is not registered", p.Encoding)
	}
	if err := codec.Validate(p.Data); err != nil {
		return fmt.Errorf("payload %s: %w", normalizeEncoding(p.Encoding), err)
	}
	return nil
}

// String returns payload data as a string for text-based transports and legacy
// compatibility paths.
func (p Payload) String() string {
	return string(p.Data)
}

// Envelope is the durable actor transport unit.
type Envelope struct {
	ID            string
	Namespace     string
	Kind          string
	From          ActorAddress
	To            ActorAddress
	CorrelationID string
	CausationID   string
	Priority      int
	DedupeKey     string
	Attempt       int
	MaxAttempts   int
	NotBefore     time.Time
	ExpiresAt     time.Time
	Payload       Payload
	Meta          map[string]string
	ReportTo      *ActorAddress
}

// Validate verifies the envelope fields required by actorlayer runtimes and
// transports.
func (e Envelope) Validate() error {
	return e.ValidateWithRegistry(DefaultCodecRegistry())
}

// ValidateWithRegistry verifies the envelope with the provided registry.
func (e Envelope) ValidateWithRegistry(reg CodecRegistry) error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("envelope id is required")
	}
	if strings.TrimSpace(e.Namespace) == "" {
		return fmt.Errorf("envelope namespace is required")
	}
	if strings.TrimSpace(e.Kind) == "" {
		return fmt.Errorf("envelope kind is required")
	}
	if _, err := e.From.String(); err != nil {
		return fmt.Errorf("envelope from: %w", err)
	}
	if _, err := e.To.String(); err != nil {
		return fmt.Errorf("envelope to: %w", err)
	}
	if err := e.Payload.ValidateWithRegistry(reg); err != nil {
		return fmt.Errorf("envelope payload: %w", err)
	}
	if e.ReportTo != nil {
		if _, err := e.ReportTo.String(); err != nil {
			return fmt.Errorf("envelope report_to: %w", err)
		}
	}
	return nil
}

type envelopeWire struct {
	ID              string            `json:"id"`
	Namespace       string            `json:"namespace"`
	Kind            string            `json:"kind"`
	From            ActorAddress      `json:"from"`
	To              ActorAddress      `json:"to"`
	CorrelationID   string            `json:"correlation_id,omitempty"`
	CausationID     string            `json:"causation_id,omitempty"`
	Priority        int               `json:"priority,omitempty"`
	DedupeKey       string            `json:"dedupe_key,omitempty"`
	Attempt         int               `json:"attempt,omitempty"`
	MaxAttempts     int               `json:"max_attempts,omitempty"`
	NotBefore       time.Time         `json:"not_before,omitempty"`
	ExpiresAt       time.Time         `json:"expires_at,omitempty"`
	PayloadJSON     string            `json:"payload_json,omitempty"`
	PayloadEncoding string            `json:"payload_encoding,omitempty"`
	PayloadData     string            `json:"payload_data,omitempty"`
	Meta            map[string]string `json:"meta,omitempty"`
	ReportTo        *ActorAddress     `json:"report_to,omitempty"`
}

// EncodeEnvelope validates and marshals an envelope using the default
// JSON-based wire projection.
func EncodeEnvelope(e Envelope) (string, error) {
	return EncodeEnvelopeWithRegistry(e, DefaultCodecRegistry())
}

// EncodeEnvelopeWithRegistry validates and marshals an envelope with the
// provided registry.
func EncodeEnvelopeWithRegistry(e Envelope, reg CodecRegistry) (string, error) {
	if err := e.ValidateWithRegistry(reg); err != nil {
		return "", fmt.Errorf("encode envelope: %w", err)
	}
	wire, err := envelopeToWire(e)
	if err != nil {
		return "", fmt.Errorf("encode envelope: %w", err)
	}
	data, err := json.Marshal(wire)
	if err != nil {
		return "", fmt.Errorf("encode envelope: %w", err)
	}
	return string(data), nil
}

// DecodeEnvelope unmarshals and validates an envelope JSON string.
func DecodeEnvelope(raw string) (Envelope, error) {
	return DecodeEnvelopeWithRegistry(raw, DefaultCodecRegistry())
}

// DecodeEnvelopeWithRegistry unmarshals and validates an envelope string with
// the provided registry.
func DecodeEnvelopeWithRegistry(raw string, reg CodecRegistry) (Envelope, error) {
	var wire envelopeWire
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &wire); err != nil {
		return Envelope{}, DecodeError(fmt.Errorf("decode envelope: %w", err))
	}
	env, err := wireToEnvelope(wire)
	if err != nil {
		return Envelope{}, DecodeError(fmt.Errorf("decode envelope: %w", err))
	}
	if err := env.ValidateWithRegistry(reg); err != nil {
		return Envelope{}, DecodeError(err)
	}
	return env, nil
}

// NewPayload validates and normalizes a payload value.
func NewPayload(encoding string, data []byte) (Payload, error) {
	payload := Payload{
		Encoding: normalizeEncoding(encoding),
		Data:     append([]byte(nil), data...),
	}
	if err := payload.Validate(); err != nil {
		return Payload{}, err
	}
	return payload, nil
}

// MarshalPayload marshals a typed actor payload with the default JSON codec.
func MarshalPayload(v any) (Payload, error) {
	return MarshalPayloadWithCodec(JSONCodec, v)
}

// MarshalPayloadWithCodec marshals a typed actor payload using the provided
// codec.
func MarshalPayloadWithCodec(codec Codec, v any) (Payload, error) {
	normalized := normalizeCodec(codec)
	data, err := normalized.Marshal(v)
	if err != nil {
		return Payload{}, fmt.Errorf("marshal actor payload: %w", err)
	}
	payload := Payload{
		Encoding: normalizeEncoding(normalized.Encoding()),
		Data:     data,
	}
	if err := payload.ValidateWithRegistry(NewCodecRegistry(normalized)); err != nil {
		return Payload{}, fmt.Errorf("marshal actor payload: %w", err)
	}
	return payload, nil
}

// UnmarshalPayload unmarshals an actor payload into dst with the default
// registry.
func UnmarshalPayload(p Payload, dst any) error {
	return UnmarshalPayloadWithRegistry(DefaultCodecRegistry(), p, dst)
}

// UnmarshalPayloadWithRegistry unmarshals payload content using the provided
// registry.
func UnmarshalPayloadWithRegistry(reg CodecRegistry, p Payload, dst any) error {
	if dst == nil {
		return DecodeError(fmt.Errorf("payload destination is required"))
	}
	reg = normalizeRegistry(reg)
	if err := p.ValidateWithRegistry(reg); err != nil {
		return DecodeError(err)
	}
	codec, ok := reg.Lookup(p.Encoding)
	if !ok {
		return DecodeError(fmt.Errorf("payload codec %q is not registered", p.Encoding))
	}
	if err := codec.Unmarshal(p.Data, dst); err != nil {
		return DecodeError(fmt.Errorf("unmarshal actor payload: %w", err))
	}
	return nil
}

func envelopeToWire(e Envelope) (envelopeWire, error) {
	wire := envelopeWire{
		ID:            e.ID,
		Namespace:     e.Namespace,
		Kind:          e.Kind,
		From:          e.From,
		To:            e.To,
		CorrelationID: e.CorrelationID,
		CausationID:   e.CausationID,
		Priority:      e.Priority,
		DedupeKey:     e.DedupeKey,
		Attempt:       e.Attempt,
		MaxAttempts:   e.MaxAttempts,
		NotBefore:     e.NotBefore,
		ExpiresAt:     e.ExpiresAt,
		Meta:          e.Meta,
		ReportTo:      e.ReportTo,
	}
	if normalizeEncoding(e.Payload.Encoding) == EncodingJSON {
		wire.PayloadJSON = string(e.Payload.Data)
		return wire, nil
	}
	wire.PayloadEncoding = normalizeEncoding(e.Payload.Encoding)
	wire.PayloadData = base64.StdEncoding.EncodeToString(e.Payload.Data)
	return wire, nil
}

func wireToEnvelope(w envelopeWire) (Envelope, error) {
	payload, err := payloadFromWire(w)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		ID:            w.ID,
		Namespace:     w.Namespace,
		Kind:          w.Kind,
		From:          w.From,
		To:            w.To,
		CorrelationID: w.CorrelationID,
		CausationID:   w.CausationID,
		Priority:      w.Priority,
		DedupeKey:     w.DedupeKey,
		Attempt:       w.Attempt,
		MaxAttempts:   w.MaxAttempts,
		NotBefore:     w.NotBefore,
		ExpiresAt:     w.ExpiresAt,
		Payload:       payload,
		Meta:          w.Meta,
		ReportTo:      w.ReportTo,
	}, nil
}

func payloadFromWire(w envelopeWire) (Payload, error) {
	if strings.TrimSpace(w.PayloadEncoding) != "" || strings.TrimSpace(w.PayloadData) != "" {
		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(w.PayloadData))
		if err != nil {
			return Payload{}, fmt.Errorf("decode payload_data: %w", err)
		}
		return Payload{
			Encoding: normalizeEncoding(w.PayloadEncoding),
			Data:     data,
		}, nil
	}
	return Payload{
		Encoding: EncodingJSON,
		Data:     []byte(strings.TrimSpace(w.PayloadJSON)),
	}, nil
}

func normalizeCodec(codec Codec) Codec {
	if codec == nil {
		return JSONCodec
	}
	return codec
}

func normalizeRegistry(reg CodecRegistry) CodecRegistry {
	if reg == nil {
		return DefaultCodecRegistry()
	}
	return reg
}

func normalizeEncoding(encoding string) string {
	return strings.ToLower(strings.TrimSpace(encoding))
}

func (r codecRegistry) Lookup(encoding string) (Codec, bool) {
	codec, ok := r.codecs[normalizeEncoding(encoding)]
	return codec, ok
}

// DedupeKeyOrID returns the explicit dedupe key, or the envelope ID when no
// dedupe key is set.
func DedupeKeyOrID(env Envelope) string {
	if trimmed := strings.TrimSpace(env.DedupeKey); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(env.ID)
}
