package actorlayer_test

import (
	"context"
	"fmt"

	"github.com/baldaworks/go-actorlayer"
	"github.com/baldaworks/go-actorlayer/dispatch"
	"github.com/baldaworks/go-actorlayer/engine"
	"github.com/baldaworks/go-actorlayer/transport/memory"
)

type welcomePayload struct {
	Name string `json:"name"`
}

type welcomeActor struct {
	done chan struct{}
}

func (a welcomeActor) Address() string { return actorlayer.WildcardAddress("session") }

func (a welcomeActor) Handle(_ context.Context, env actorlayer.Envelope) error {
	var payload welcomePayload
	if err := actorlayer.UnmarshalPayload(env.Payload, &payload); err != nil {
		return err
	}
	fmt.Printf("welcome %s\n", payload.Name)
	close(a.done)
	return nil
}

func Example_endToEnd() {
	bus := memory.New(1)
	registry := dispatch.NewMemoryRegistry()
	done := make(chan struct{})
	_ = registry.Register(welcomeActor{done: done})

	runtime, _ := engine.NewDispatchRuntime(engine.RuntimeConfig{
		Registry:  registry,
		AddressOf: func(env engine.Envelope) (string, error) { return env.To.String() },
		Retry: engine.RetryPolicy{
			IsRetryable: actorlayer.IsRetryableError,
			Backoff:     actorlayer.RetryDelay,
		},
	})

	ctx := context.Background()
	runDone := make(chan error, 1)
	go func() {
		runDone <- runtime.Run(ctx, bus)
	}()

	payload, _ := actorlayer.MarshalPayload(welcomePayload{Name: "ada"})
	_, _ = bus.Dispatch(ctx, actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "example.command",
		Kind:      "welcome",
		From:      actorlayer.SystemAddress("example"),
		To:        actorlayer.ActorAddress{Target: "session", Key: "demo"},
		Payload:   payload,
	})

	<-done
	_ = bus.Drain(ctx)
	<-runDone

	// Output: welcome ada
}
