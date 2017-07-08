package generate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// String returns a pointer to the provide string or nil if the zero value was passed in
func String(in string) *string {
	if in == "" {
		return nil
	}

	return &in
}

func name(in *descriptor.FileDescriptorProto, kind string) *string {
	n := "service"
	if in.Name != nil {
		n = *in.Name
	}

	ext := filepath.Ext(n)
	n = n[0 : len(n)-len(ext)]
	return String(fmt.Sprintf("%v.pb.%v.go", n, kind))
}

func packageName(in *descriptor.FileDescriptorProto) (string, error) {
	if in.Package != nil {
		return *in.Package, nil
	}

	if in.Name != nil {
		name := *in.Name
		ext := filepath.Ext(name)
		return name[0 : len(name)-len(ext)], nil
	}

	return "", errors.New("unable to determine package name")
}

func base(in string) string {
	idx := strings.LastIndex(in, ".")
	if idx == -1 {
		return in
	}
	return in[idx+1:]
}

func lower(in string) string {
	return strings.ToLower(in)
}

func newTemplate(content string) (*template.Template, error) {
	fn := map[string]interface{}{
		"base":  base,
		"lower": lower,
	}

	return template.New("page").Funcs(fn).Parse(content)
}
