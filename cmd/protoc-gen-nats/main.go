package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"github.com/savaki/nats-protobuf/cmd/protoc-gen-nats/generate"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	check(err)

	req := plugin_go.CodeGeneratorRequest{}
	err = proto.Unmarshal(data, &req)
	check(err)

	results := make([]*plugin_go.CodeGeneratorResponse_File, 0, len(req.ProtoFile))
	for _, file := range req.ProtoFile {
		v, err := generate.RPC(file)
		check(err)

		results = append(results, v)
	}

	res := &plugin_go.CodeGeneratorResponse{
		File: results,
	}
	data, err = proto.Marshal(res)
	check(err)

	os.Stdout.Write(data)
}
