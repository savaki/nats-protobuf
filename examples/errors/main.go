package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/nats-io/go-nats"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *In) (*Out, error) {
	return nil, errors.New("boom!")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	done, _ := SubscribeRPC(ctx, nc, "subject", "id", Service{})
	defer cancel()

	// Client
	client := NewRPC(nc, "subject")

	_, err := client.InOut(ctx, &In{Input: "Joe"})
	fmt.Println(err)

	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}
