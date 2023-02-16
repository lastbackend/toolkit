/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package descriptor

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

type GoPackage struct {
	Path  string
	Name  string
	Alias string
}

func (p GoPackage) Standard() bool {
	return !strings.Contains(p.Path, ".")
}

func (p GoPackage) String() string {
	if p.Alias == "" {
		return fmt.Sprintf("%q", p.Path)
	}
	return fmt.Sprintf("%s %q", p.Alias, p.Path)
}

type File struct {
	*descriptorpb.FileDescriptorProto
	GeneratedFilenamePrefix string
	GoPkg                   GoPackage
	Messages                []*Message
	Enums                   []*Enum
	Services                []*Service
}

func (f *File) Pkg() string {
	pkg := f.GoPkg.Name
	if alias := f.GoPkg.Alias; alias != "" {
		pkg = alias
	}
	return pkg
}

type Message struct {
	*descriptorpb.DescriptorProto
	File   *File
	Fields []*Field
	Outers []string
	Index  int
}

func (m *Message) FullyName() string {
	parts := []string{""}
	if m.File.Package != nil {
		parts = append(parts, m.File.GetPackage())
	}
	parts = append(parts, m.Outers...)
	parts = append(parts, m.GetName())
	return strings.Join(parts, ".")
}

func (m *Message) GoType(currentPackage string) string {
	var parts []string
	parts = append(parts, m.Outers...)
	parts = append(parts, m.GetName())

	name := strings.Join(parts, "_")
	if m.File.GoPkg.Path == currentPackage {
		return name
	}
	return fmt.Sprintf("%s.%s", m.File.Pkg(), name)
}

func (m *Message) GoName() string {
	var parts []string
	parts = append(parts, m.Outers...)
	parts = append(parts, m.GetName())
	return strings.Join(parts, "_")
}

type Field struct {
	*descriptorpb.FieldDescriptorProto
	Message      *Message
	FieldMessage *Message
}

type Enum struct {
	*descriptorpb.EnumDescriptorProto
	File   *File
	Outers []string
	Index  int
}

func (e *Enum) FullyName() string {
	parts := []string{""}
	if e.File.Package != nil {
		parts = append(parts, e.File.GetPackage())
	}
	parts = append(parts, e.Outers...)
	parts = append(parts, e.GetName())
	return strings.Join(parts, ".")
}

type Service struct {
	*descriptorpb.ServiceDescriptorProto
	File    *File
	Methods []*Method
}

func (s *Service) FullyName() string {
	var parts []string
	if s.File.Package != nil {
		parts = append(parts, s.File.GetPackage())
	}
	parts = append(parts, s.GetName())
	return strings.Join(parts, ".")
}

type Method struct {
	*descriptorpb.MethodDescriptorProto
	Service      *Service
	RequestType  *Message
	ResponseType *Message
	Name         string
	IsWebsocket  bool
	Bindings     []*Binding
}

func (m *Method) FullyName() string {
	var parts []string
	parts = append(parts, m.Service.FullyName())
	parts = append(parts, m.GetName())
	return strings.Join(parts, ".")
}

type ResponseFile struct {
	*pluginpb.CodeGeneratorResponse_File
	GoPkg   GoPackage
	Rewrite bool
}

type Binding struct {
	Method       *Method
	Index        int
	RpcMethod    string
	RpcPath      string
	Service      string
	HttpMethod   string
	HttpPath     string
	RawBody      string
	HttpParams   []string
	RequestType  *Message
	ResponseType *Message
	Stream       bool
	Websocket    bool
	Subscribe    bool
}
