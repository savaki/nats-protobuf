package main

import (
	"context"
	"fmt"

	"github.com/nats-io/go-nats"
)

//go:generate protoc -I .:$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf --gogo_out=. service.proto
//go:generate protoc -I .:$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *In) (*Out, error) {
	return &Out{
		Output: fmt.Sprintf("%v %v [%v]", in.Input, in.FirstName, in.OrgID),
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	done, _ := SubscribeRPC(ctx, nc, "subject", "id", Service{})
	defer cancel()

	// Client
	client := NewRPC(nc, "subject")

	in := &In{
		Input:     "Joe",
		FirstName: "first",
		OrgID:     "abc",
	}
	out, _ := client.InOut(ctx, in)
	fmt.Println(out.Output)

	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}
