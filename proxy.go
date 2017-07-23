package nats_protobuf

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"
)

type rtFunc func(req *http.Request) (*http.Response, error)

func (fn rtFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func marshal(r io.ReadCloser, name string, in proto.Message) (*Message, error) {
	if err := json.NewDecoder(r).Decode(in); err != nil {
		return nil, err
	}
	defer r.Close()

	data, err := proto.Marshal(in)
	if err != nil {
		return nil, err
	}

	return &Message{
		Method:  name,
		Headers: map[string]string{},
		Payload: data,
	}, nil
}

func newTransport(subject string, h HandlerFunc, mapper Mapper) http.RoundTripper {
	return rtFunc(func(req *http.Request) (*http.Response, error) {
		name := req.URL.Path
		if index := strings.LastIndex(req.URL.Path, "/"); index >= 0 {
			name = name[index+1:]
		}

		in, out, ok := mapper(name)
		if !ok {
			w := httptest.NewRecorder()
			w.WriteHeader(http.StatusNotFound)
			return w.Result(), nil
		}

		msgIn, err := marshal(req.Body, name, in)
		if err != nil {
			return nil, err
		}

		msgOut, err := h(req.Context(), subject, msgIn)
		if err != nil {
			return nil, err
		}

		if msgOut.Error != "" {
			w := httptest.NewRecorder()
			w.HeaderMap.Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, msgOut.Error)
			return w.Result(), nil
		}

		if err := proto.Unmarshal(msgOut.Payload, out); err != nil {
			return nil, err
		}

		w := httptest.NewRecorder()
		w.HeaderMap.Set("Content-Type", "application/json")
		if msgOut.Headers != nil {
			for k, v := range msgOut.Headers {
				w.HeaderMap.Add(k, v)
			}
		}
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(out); err != nil {
			return nil, err
		}

		return w.Result(), nil
	})
}

// Mapper accepts the name of a function and returns a new instance of it's input and output types or false if the name
// provided was not recognized
type Mapper func(string) (proto.Message, proto.Message, bool)

// NewHandler constructs a new http -> nats proxy
func NewProxy(nc *nats.Conn, subject string, mapper Mapper, filters ...Filter) http.Handler {
	fn := NewRequestFunc(nc, subject)

	for i := len(filters) - 1; i >= 0; i-- {
		filter := filters[i]
		fn = filter(fn)
	}

	transport := newTransport(subject, fn, mapper)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		resp, err := transport.RoundTrip(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		if resp.Body != nil {
			io.Copy(w, resp.Body)
			defer resp.Body.Close()
		}
	})
}
