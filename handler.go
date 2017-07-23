package nats_protobuf

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
)

// HandlerFunc provides an abstraction over the call to nats.  Useful for defining middleware.
type HandlerFunc func(ctx context.Context, subject string, m *Message) (*Message, error)

// Filter defines the shape of the middleware.  Both ```Subscribe``` and ```New*``` accept an optional list of middlewares
// that will be applied in FIFO order
type Filter func(requestFunc HandlerFunc) HandlerFunc

// NewRequestFunc generates a new HandlerFunc that uses nats.Request
func NewRequestFunc(nc *nats.Conn, subject string) HandlerFunc {
	return func(ctx context.Context, _ string, m *Message) (*Message, error) {
		data, err := proto.Marshal(m)
		if err != nil {
			return nil, err
		}

		msg, err := nc.RequestWithContext(ctx, subject, data)
		if err != nil {
			return nil, err
		}

		reply := &Message{}
		if err := proto.Unmarshal(msg.Data, reply); err != nil {
			return nil, err
		}

		if reply.Error != "" {
			return nil, errors.New(reply.Error)
		}

		return reply, nil
	}
}

// PublishOnly sends the message using nats.Publish rather than nats.Request allowing for multiple
// receivers.  Note that when using PublishOnly, the response object will always be nil.
func PublishOnly(nc *nats.Conn, subject string) Filter {
	return func(fn HandlerFunc) HandlerFunc {
		return func(ctx context.Context, _ string, m *Message) (*Message, error) {
			data, err := proto.Marshal(m)
			if err != nil {
				return nil, err
			}

			return nil, nc.Publish(subject, data)
		}
	}
}

// NewPublishFunc generates a new HandlerFunc that uses nats.Publish
func NewPublishFunc(nc *nats.Conn, subject string) HandlerFunc {
	return func(ctx context.Context, _ string, m *Message) (*Message, error) {
		data, err := proto.Marshal(m)
		if err != nil {
			return nil, err
		}

		err = nc.Publish(subject, data)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

func NewMessage(method string, in proto.Message) (*Message, error) {
	message := &Message{
		Method:  method,
		Headers: map[string]string{},
	}

	if in != nil {
		v, err := proto.Marshal(in)
		if err != nil {
			return nil, err
		}
		message.Payload = v
	}

	return message, nil
}
