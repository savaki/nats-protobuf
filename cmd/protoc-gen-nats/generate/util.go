package generate

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gogo/protobuf/gogoproto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

// String returns a pointer to the provide string or nil if the zero value was passed in
func String(in string) *string {
	if in == "" {
		return nil
	}

	return &in
}

func fileName(in *descriptor.FileDescriptorProto, kind string) *string {
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

func id(typeName string, messages []*descriptor.DescriptorProto) string {
	name := base(typeName)
	for _, message := range messages {
		if name == *message.Name {
			for _, field := range message.Field {
				fieldName := *field.Name
				if fieldName != "id" {
					continue
				}

				name := gogoproto.GetCustomName(field)
				if name != "" {
					return name
				}

				return "Id"
			}
		}
	}

	return "Id"
}

func camel(in string) string {
	segments := strings.Split(in, "_")
	capped := make([]string, 0, len(segments))

	for _, segment := range segments {
		if segment == "" {
			continue
		}
		capped = append(capped, strings.ToUpper(segment[0:1])+segment[1:])
	}
	return strings.Join(capped, "")
}

func name(field *descriptor.FieldDescriptorProto) string {
	name := gogoproto.GetCustomName(field)
	if name != "" {
		return name
	}

	return camel(*field.Name)
}

func newTemplate(content string) (*template.Template, error) {
	fn := map[string]interface{}{
		"base":  base,
		"lower": lower,
		"name":  name,
	}

	return template.New("page").Funcs(fn).Parse(content)
}
