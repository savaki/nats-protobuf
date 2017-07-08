package nats_protobuf_test

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/savaki/nats-protobuf"
	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	method := "method"
	in := &nats_protobuf.Message{Method: "In"}

	msg, err := nats_protobuf.NewMessage(method, in)
	assert.Nil(t, err)
	assert.Equal(t, method, msg.Method)

	actual := &nats_protobuf.Message{}
	err = proto.UnmarshalMerge(msg.Payload, actual)
	assert.Nil(t, err)
	assert.Equal(t, in, actual)
}
