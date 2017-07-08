[![GoDoc](https://godoc.org/github.com/savaki/nats-protobuf?status.svg)](https://godoc.org/github.com/savaki/nats-protobuf)

# nats-protobuf

```nats-protobuf``` is a protoc plugin that generates client, server, and json over 
http implementations of the services defined in the .proto file

The library is functional, but given how new it is, expect breaking changes.

## Installation

```
go get github.com/savaki/nats-protobuf/...
```

## Usage
 
The simplest usage is to use nats-protobuf along side your protoc go:generate line.
  
```text
//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto
```

## Example

For example, given this protoc file, service.proto:

```proto
syntax = "proto3";

package simple;

message In {
    string input = 1;
}

message Out {
    string output = 1;
}

service RPC {
    rpc InOut (In) returns (Out);
}
```

The following is a complete example that illustrates a round trip using client and server transports
generated by the nats-rpc plugin.

```go
package main

import (
	"context"
	"fmt"

	"github.com/nats-io/nats"
)

//go:generate protoc --go_out=. service.proto
//go:generate protoc --nats_out=. service.proto

type Service struct {
}

func (s Service) InOut(ctx context.Context, in *In) (*Out, error) {
	return &Out{Output: "Hello " + in.Input}, nil
}

func main() {
	ctx := context.Background()
	nc, _ := nats.Connect(nats.DefaultURL)

	// Server
	cancel, _ := SubscribeRPC(ctx, nc, "subject", "id", Service{})
	defer cancel()

	// Client
	client := NewRPC(nc, "subject")

	out, _ := client.InOut(ctx, &In{Input: "Joe"})
	fmt.Println(out.Output)
}
```

Additional examples can be found in the examples directory.

## Middleware

Both ```Listen``` and ```New*``` accept an optional list of middleware
filters to apply.  The filters need to implement:

```go
func(fn func(context.Context, *Message) error) func(context.Context, *Message) error
```

For example, if we want to print out the elapsed execution time, we could write a timer as follows: 

```go
func Timer(fn nats_protobuf.HandlerFunc) nats_protobuf.HandlerFunc {
	return func(ctx context.Context, msg *nats_protobuf.Message) error {
		started := time.Now()
		defer func() {
			fmt.Println("elapsed:", time.Now().Sub(started))
		}()
		return fn(ctx, msg)
	}
}
```

For a more detailed example of middleware in use see the ```middleware``` package in the examples directory.

####  Enjoy!
