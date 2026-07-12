package memory_test

import (
	"context"
	"fmt"

	"github.com/baldaworks/go-actorlayer"
	"github.com/baldaworks/go-actorlayer/engine"
	"github.com/baldaworks/go-actorlayer/transport/memory"
)

func ExampleTransport() {
	bus := memory.New(1)
	payload, _ := actorlayer.MarshalPayload(struct {
		Name string `json:"name"`
	}{Name: "ada"})

	_, _ = bus.Dispatch(context.Background(), actorlayer.Envelope{
		ID:        "env-1",
		Namespace: "example.command",
		Kind:      "welcome",
		From:      actorlayer.SystemAddress("example"),
		To:        actorlayer.ActorAddress{Target: "session", Key: "demo"},
		Payload:   payload,
	})
	_ = bus.Run(context.Background(), func(ctx context.Context, delivery engine.Delivery) error {
		var got struct {
			Name string `json:"name"`
		}
		_ = actorlayer.UnmarshalPayload(delivery.Envelope().Payload, &got)
		fmt.Printf("%s -> %s\n", delivery.Envelope().ID, got.Name)
		_ = delivery.Ack(ctx)
		return bus.Drain(ctx)
	})
	// Output: env-1 -> ada
}
