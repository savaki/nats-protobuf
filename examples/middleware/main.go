package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-protobuf"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *In) (*Out, error) {
	return &Out{Output: "Hello " + in.Input}, nil
}

func Timer(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
	return func(ctx context.Context, msg *nats_protobuf.Message) (*nats_protobuf.Message, error) {
		started := time.Now()
		defer func() {
			fmt.Println("elapsed:", time.Now().Sub(started))
		}()
		return fn(ctx, msg)
	}
}

func Logger(label string) nats_protobuf.Filter {
	return func(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
		return func(ctx context.Context, msg *nats_protobuf.Message) (*nats_protobuf.Message, error) {
			fmt.Println(label, msg.Method)
			return fn(ctx, msg)
		}
	}
}

func Interceptor(output string) nats_protobuf.Filter {
	return func(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
		return func(ctx context.Context, msg *nats_protobuf.Message) (*nats_protobuf.Message, error) {
			return nats_protobuf.NewMessage(msg.Method, &Out{Output: output})
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	done, _ := Subscribe(ctx, nc, "subject", "id", Service{}, Logger("server"), Timer, Interceptor("output!"))

	// Client
	client := NewRPC(nc, "subject", Logger("client"))

	out, _ := client.InOut(ctx, &In{Input: "Joe"})
	fmt.Println(out.Output)

	cancel() // stop the service
	<-done   // wait for it to unsubscribe
}