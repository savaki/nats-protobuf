// Code generated by github.com/savaki/nats-protobuf. DO NOT EDIT.
// source: service.proto

package main

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-protobuf"
)

type RPC interface {
	InOut(ctx context.Context, in *In) (*Out, error)
}

type rpcClient struct {
	subject string
	fn      nats_protobuf.HandlerFunc
}

func (c *rpcClient) InOut(ctx context.Context, in *In) (*Out, error) {
	message, err := nats_protobuf.NewMessage("InOut", in)
	if err != nil {
		return nil, err
	}

	reply, err := c.fn(ctx, message)
	if err != nil {
		return nil, err
	}

	result := &Out{}
	err = proto.Unmarshal(reply.Payload, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}


func NewRPC(nc *nats.Conn, subject string, filters ...nats_protobuf.Filter) RPC {
	fn := nats_protobuf.NewRequestFunc(nc, subject)

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]
		fn = filter(fn)
	}

	return &rpcClient{
		subject: subject,
		fn:      fn,
	}
}

func Subscribe(ctx context.Context, nc *nats.Conn, subject, queue string, service RPC, filters ...nats_protobuf.Filter) (<-chan struct{}, error) {
	publishErr := func(reply string, err error) {
		raw := &nats_protobuf.Message{Error: err.Error()}
		if data, err := proto.Marshal(raw); err == nil {
			if reply != "" {
				nc.Publish(reply, data)
			}
		}
	}

	fn := func(ctx context.Context, m *nats_protobuf.Message) (*nats_protobuf.Message, error) {
		switch m.Method {

		case "InOut":
			in := &In{}
			if err := proto.Unmarshal(m.Payload, in); err != nil {
				return nil, err
			}

			v, err := service.InOut(ctx, in)
			if err != nil {
				return nil, err
			}

			return nats_protobuf.NewMessage(m.Method, v)

		default:
			return nil, fmt.Errorf("unhandled method, %v", m.Method)
		}
	}

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]
		fn = filter(fn)
	}

	subscription, err := nc.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		go func() {
			req := &nats_protobuf.Message{}
			if err := proto.Unmarshal(msg.Data, req); err != nil {
				publishErr(msg.Reply, err)
				return
			}

			if msg.Reply == "" {
				return
			}

			v, err := fn(ctx, req)
			if err != nil {
				publishErr(msg.Reply, err)
				return
			}

			data, err := proto.Marshal(v)
			if err != nil {
				publishErr(msg.Reply, err)
				return
			}

			nc.Publish(msg.Reply, data)
		}()
	})
	if err != nil {
		return nil, err
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		select {
		case <-ctx.Done():
			subscription.Unsubscribe()
		}
	}()

	return done, nil
}

