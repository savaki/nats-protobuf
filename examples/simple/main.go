package main

import (
	"context"
	"fmt"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-protobuf/examples"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *examples.In) (*examples.Out, error) {
	return &examples.Out{Output: "Hello " + in.Input}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	done, _ := examples.SubscribeRPC(ctx, nc, "subject", "id", Service{})

	// Client
	client := examples.NewRPC(nc, "subject")

	out, _ := client.InOut(ctx, &examples.In{Input: "Joe"})
	fmt.Println(out.Output)

	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}
