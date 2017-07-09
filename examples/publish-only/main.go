package main

import (
	"context"
	"fmt"

	"time"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-protobuf"
	"github.com/savaki/nats-protobuf/examples"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *examples.In) (*examples.Out, error) {
	fmt.Printf("InOut(%v)\n", in.Input)
	return &examples.Out{Output: "Hello " + in.Input}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()

	// Server
	subject := "subject"
	done, _ := examples.SubscribeRPC(ctx, nc, subject, "id", Service{})

	// Client
	client := examples.NewRPC(nc, subject,
		nats_protobuf.PublishOnly(nc, subject),
	)

	client.InOut(ctx, &examples.In{Input: "Joe"})

	time.Sleep(time.Millisecond * 250)
	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}
