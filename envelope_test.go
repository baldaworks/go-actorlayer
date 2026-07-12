package actorlayer_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/baldaworks/go-actorlayer"
)

type wrapperCodec struct{}

func (wrapperCodec) Encoding() string {
	return "wrapped-json"
}

func (wrapperCodec) Marshal(v any) ([]byte, error) {
	raw, err := actorlayer.JSONCodec.Marshal(v)
	if err != nil {
		return nil, err
	}
	return []byte(`{"wrapped":` + string(raw) + `}`), nil
}

func (wrapperCodec) Unmarshal(data []byte, v any) error {
	var wrapped struct {
		Wrapped json.RawMessage `json:"wrapped"`
	}
	if err := actorlayer.JSONCodec.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	return actorlayer.JSONCodec.Unmarshal(wrapped.Wrapped, v)
}

func (wrapperCodec) Validate(data []byte) error {
	var wrapped struct {
		Wrapped json.RawMessage `json:"wrapped"`
	}
	if err := actorlayer.JSONCodec.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	if len(wrapped.Wrapped) == 0 {
		return errors.New("wrapped payload is required")
	}
	return nil
}

func TestEnvelopeRoundTrip(t *testing.T) {
	t.Parallel()

	env := actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "test.command",
		Kind:      "message",
		From:      actorlayer.ActorAddress{Target: "system", Key: "source"},
		To:        actorlayer.ActorAddress{Target: "session", Key: "1"},
		DedupeKey: "dedupe-1",
		Payload: actorlayer.Payload{
			Encoding: actorlayer.EncodingJSON,
			Data:     []byte(`{"ok":true}`),
		},
	}

	raw, err := actorlayer.EncodeEnvelope(env)
	if err != nil {
		t.Fatalf("EncodeEnvelope() error = %v", err)
	}
	got, err := actorlayer.DecodeEnvelope(raw)
	if err != nil {
		t.Fatalf("DecodeEnvelope() error = %v", err)
	}
	if got.ID != env.ID || got.Namespace != env.Namespace || got.Kind != env.Kind || got.From != env.From || got.To != env.To || got.DedupeKey != env.DedupeKey || got.Payload.Encoding != env.Payload.Encoding || string(got.Payload.Data) != string(env.Payload.Data) {
		t.Fatalf("DecodeEnvelope(EncodeEnvelope()) = %#v, want %#v", got, env)
	}
	if key := actorlayer.DedupeKeyOrID(got); key != env.DedupeKey {
		t.Fatalf("DedupeKeyOrID() = %q, want %q", key, env.DedupeKey)
	}
}

func TestDecodeEnvelopeClassifiesInvalidEnvelopeAsDecodeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "invalid json", raw: `{not-json`},
		{name: "invalid envelope", raw: `{"id":"env-1"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := actorlayer.DecodeEnvelope(tt.raw)
			if err == nil {
				t.Fatal("DecodeEnvelope() error = nil, want decode error")
			}
			var actorErr *actorlayer.ActorError
			if !errors.As(err, &actorErr) || actorErr.Kind != actorlayer.ErrorKindDecode {
				t.Fatalf("DecodeEnvelope() error = %v, want decode actor error", err)
			}
			if actorlayer.IsRetryableError(err) {
				t.Fatal("IsRetryableError(decode error) = true, want false")
			}
		})
	}
}

func TestEnvelopeValidateRejectsInvalidJSONPayload(t *testing.T) {
	t.Parallel()

	env := validEnvelopeForTest()
	env.Payload = actorlayer.Payload{Encoding: actorlayer.EncodingJSON, Data: []byte(`{not-json`)}
	if err := env.Validate(); err == nil || !strings.Contains(err.Error(), "payload json: payload must be valid json") {
		t.Fatalf("Validate() error = %v, want invalid json payload error", err)
	}
}

func TestEnvelopeValidateRejectsMissingEncoding(t *testing.T) {
	t.Parallel()

	env := validEnvelopeForTest()
	env.Payload = actorlayer.Payload{Data: []byte(`{"ok":true}`)}
	if err := env.Validate(); err == nil || !strings.Contains(err.Error(), "payload encoding is required") {
		t.Fatalf("Validate() error = %v, want missing encoding error", err)
	}
}

func TestEnvelopeValidateRejectsInvalidReportTo(t *testing.T) {
	t.Parallel()

	env := validEnvelopeForTest()
	env.ReportTo = &actorlayer.ActorAddress{Target: "session"}
	if err := env.Validate(); err == nil || !strings.Contains(err.Error(), "envelope report_to") {
		t.Fatalf("Validate() error = %v, want invalid report_to error", err)
	}
}

func TestEncodeEnvelopeRejectsInvalidEnvelope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  actorlayer.Envelope
		want string
	}{
		{
			name: "missing id",
			env: func() actorlayer.Envelope {
				env := validEnvelopeForTest()
				env.ID = ""
				return env
			}(),
			want: "envelope id is required",
		},
		{
			name: "invalid payload",
			env: func() actorlayer.Envelope {
				env := validEnvelopeForTest()
				env.Payload = actorlayer.Payload{Encoding: actorlayer.EncodingJSON, Data: []byte("{not-json")}
				return env
			}(),
			want: "payload must be valid json",
		},
		{
			name: "invalid report to",
			env: func() actorlayer.Envelope {
				env := validEnvelopeForTest()
				env.ReportTo = &actorlayer.ActorAddress{Target: "session"}
				return env
			}(),
			want: "envelope report_to",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := actorlayer.EncodeEnvelope(tt.env); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("EncodeEnvelope() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestPayloadHelpers(t *testing.T) {
	t.Parallel()

	type payload struct {
		OK   bool   `json:"ok"`
		Name string `json:"name"`
	}
	raw, err := actorlayer.MarshalPayload(payload{OK: true, Name: "test"})
	if err != nil {
		t.Fatalf("MarshalPayload() error = %v", err)
	}
	var got payload
	if err := actorlayer.UnmarshalPayload(raw, &got); err != nil {
		t.Fatalf("UnmarshalPayload() error = %v", err)
	}
	if got != (payload{OK: true, Name: "test"}) {
		t.Fatalf("UnmarshalPayload() = %#v, want ok/test", got)
	}
	if err := actorlayer.UnmarshalPayload(actorlayer.Payload{Encoding: actorlayer.EncodingJSON, Data: []byte(`{"ok":`)}, &got); err == nil || actorlayer.ClassifyError(err) != actorlayer.ErrorKindDecode {
		t.Fatalf("UnmarshalPayload(invalid) error = %v, want decode error", err)
	}
	if err := actorlayer.UnmarshalPayload(raw, nil); err == nil || actorlayer.ClassifyError(err) != actorlayer.ErrorKindDecode {
		t.Fatalf("UnmarshalPayload(nil) error = %v, want decode error", err)
	}
}

func TestPayloadHelpersWithCodec(t *testing.T) {
	t.Parallel()

	type payload struct {
		Value string `json:"value"`
	}

	reg := actorlayer.NewCodecRegistry(actorlayer.JSONCodec, wrapperCodec{})
	raw, err := actorlayer.MarshalPayloadWithCodec(wrapperCodec{}, payload{Value: "ok"})
	if err != nil {
		t.Fatalf("MarshalPayloadWithCodec() error = %v", err)
	}
	if raw.Encoding != "wrapped-json" || string(raw.Data) != `{"wrapped":{"value":"ok"}}` {
		t.Fatalf("MarshalPayloadWithCodec() = %#v", raw)
	}
	var got payload
	if err := actorlayer.UnmarshalPayloadWithRegistry(reg, raw, &got); err != nil {
		t.Fatalf("UnmarshalPayloadWithRegistry() error = %v", err)
	}
	if got.Value != "ok" {
		t.Fatalf("UnmarshalPayloadWithRegistry() = %#v", got)
	}
}

func TestEnvelopeRoundTripLegacyJSONWire(t *testing.T) {
	t.Parallel()

	got, err := actorlayer.DecodeEnvelope(`{"id":"env-1","namespace":"test.command","kind":"message","from":{"target":"system","key":"source"},"to":{"target":"session","key":"1"},"payload_json":"{\"ok\":true}"}`)
	if err != nil {
		t.Fatalf("DecodeEnvelope(legacy) error = %v", err)
	}
	if got.Payload.Encoding != actorlayer.EncodingJSON || string(got.Payload.Data) != `{"ok":true}` {
		t.Fatalf("legacy payload = %#v", got.Payload)
	}
}

func validEnvelopeForTest() actorlayer.Envelope {
	return actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "test.command",
		Kind:      "message",
		From:      actorlayer.ActorAddress{Target: "system", Key: "source"},
		To:        actorlayer.ActorAddress{Target: "session", Key: "1"},
		Payload: actorlayer.Payload{
			Encoding: actorlayer.EncodingJSON,
			Data:     []byte(`{"ok":true}`),
		},
	}
}
