package engine_test

import (
	"context"
	"fmt"
	"time"

	"github.com/baldaworks/go-actorlayer"
	"github.com/baldaworks/go-actorlayer/dispatch"
	"github.com/baldaworks/go-actorlayer/engine"
)

type welcomePayload struct {
	Name string `json:"name"`
}

type welcomeActor struct{}

func (welcomeActor) Address() string { return actorlayer.WildcardAddress("session") }

func (welcomeActor) Handle(_ context.Context, env actorlayer.Envelope) error {
	var payload welcomePayload
	if err := actorlayer.UnmarshalPayload(env.Payload, &payload); err != nil {
		return err
	}
	fmt.Printf("welcome %s\n", payload.Name)
	return nil
}

type inMemoryDelivery struct {
	env actorlayer.Envelope
}

func (d inMemoryDelivery) Envelope() engine.Envelope { return d.env }
func (inMemoryDelivery) Attempt() int                { return 1 }
func (inMemoryDelivery) MaxAttempts() int            { return 1 }
func (inMemoryDelivery) InProgress(context.Context) error {
	return nil
}
func (inMemoryDelivery) Ack(context.Context) error { return nil }
func (inMemoryDelivery) Retry(context.Context, time.Duration, string) error {
	return nil
}
func (inMemoryDelivery) DeadLetter(context.Context, string) error { return nil }

func ExampleDispatchRuntime_Handle() {
	registry := dispatch.NewMemoryRegistry()
	_ = registry.Register(welcomeActor{})
	runtime, _ := engine.NewDispatchRuntime(engine.RuntimeConfig{
		Registry:  registry,
		AddressOf: func(env engine.Envelope) (string, error) { return env.To.String() },
		Retry: engine.RetryPolicy{
			IsRetryable: actorlayer.IsRetryableError,
			Backoff:     actorlayer.RetryDelay,
		},
	})

	payload, _ := actorlayer.MarshalPayload(welcomePayload{Name: "ada"})
	_ = runtime.Handle(context.Background(), inMemoryDelivery{env: actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "example.command",
		Kind:      "welcome",
		From:      actorlayer.SystemAddress("example"),
		To:        actorlayer.ActorAddress{Target: "session", Key: "demo"},
		Payload:   payload,
	}})
	// Output: welcome ada
}
