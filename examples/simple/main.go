package main

import (
	"context"
	"fmt"

	"github.com/nats-io/go-nats"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *In) (*Out, error) {
	return &Out{Output: "Hello " + in.Input}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	done, _ := Subscribe(ctx, nc, "subject", "id", Service{})

	// Client
	client := NewRPC(nc, "subject")

	out, _ := client.InOut(ctx, &In{Input: "Joe"})
	fmt.Println(out.Output)

	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}
