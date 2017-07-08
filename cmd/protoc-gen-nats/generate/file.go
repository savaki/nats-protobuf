package generate

import (
	"bytes"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
)

const (
	rpc = `// Code generated by github.com/savaki/nats-protobuf. DO NOT EDIT.
// source: {{ .File }}

package {{ .Package }}

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/savaki/nats-protobuf"
)
{{ range .Services }}{{ $Service := .Name }}
type {{ $Service }} interface {
{{ range .Method }}	{{ .Name }}(ctx context.Context, in *{{ .InputType | base }}) (*{{ .OutputType | base}}, error)
{{ end }}}

type {{ $Service | lower }}Client struct {
	subject string
	fn      nats_protobuf.HandlerFunc
}

{{ range .Method }}func (c *{{ $Service | lower}}Client) {{ .Name }}(ctx context.Context, in *{{ .InputType | base }}) (*{{ .OutputType | base}}, error) {
	message, err := nats_protobuf.NewMessage("{{ .Name }}", in)
	if err != nil {
		return nil, err
	}

	reply, err := c.fn(ctx, message)
	if err != nil {
		return nil, err
	}

	result := &{{ .OutputType | base }}{}
	err = proto.Unmarshal(reply.Payload, result)
	if err != nil {
		return nil, err
	}

	return result, nil

}
{{ end }}

func New{{ .Name }}(nc *nats.Conn, subject string, filters ...nats_protobuf.Filter) {{ .Name }} {
	fn := nats_protobuf.NewRequestFunc(nc, subject)

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]
		fn = filter(fn)
	}

	return &{{ $Service | lower }}Client{
		subject: subject,
		fn:      fn,
	}
}

func Subscribe(ctx context.Context, nc *nats.Conn, subject, queue string, service {{ .Name }}, filters ...nats_protobuf.Filter) (<-chan struct{}, error) {
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
{{ range .Method }}
		case "{{ .Name }}":
			in := &{{ .InputType | base }}{}
			if err := proto.Unmarshal(m.Payload, in); err != nil {
				return nil, err
			}

			v, err := service.{{ .Name }}(ctx, in)
			if err != nil {
				return nil, err
			}

			return nats_protobuf.NewMessage(m.Method, v)
{{ end }}
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
{{ end }}
`
)

// File accepts the proto file definition and returns the response for this file
func RPC(in *descriptor.FileDescriptorProto) (*plugin_go.CodeGeneratorResponse_File, error) {
	pkg, err := packageName(in)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	t, err := newTemplate(rpc)
	if err != nil {
		return nil, err
	}

	t.Execute(buf, map[string]interface{}{
		"File":     *in.Name,
		"Package":  pkg,
		"Services": in.Service,
	})

	return &plugin_go.CodeGeneratorResponse_File{
		Name:    name(in, "nats"),
		Content: String(buf.String()),
	}, nil
}