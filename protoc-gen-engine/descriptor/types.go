/*
Copyright [2014] - [2021] The Last.Backend authors.

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
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"strings"
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
	components := []string{""}
	if m.File.Package != nil {
		components = append(components, m.File.GetPackage())
	}
	components = append(components, m.Outers...)
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

func (m *Message) GoType() string {
	var components []string
	components = append(components, m.Outers...)
	components = append(components, m.GetName())

	name := strings.Join(components, "_")
	return fmt.Sprintf("%s.%s", m.File.Pkg(), name)
}

func (m *Message) GoName() string {
	var components []string
	components = append(components, m.Outers...)
	components = append(components, m.GetName())
	return strings.Join(components, "_")
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
	components := []string{""}
	if e.File.Package != nil {
		components = append(components, e.File.GetPackage())
	}
	components = append(components, e.Outers...)
	components = append(components, e.GetName())
	return strings.Join(components, ".")
}

type Service struct {
	*descriptorpb.ServiceDescriptorProto
	File    *File
	Methods []*Method
}

func (s *Service) FullyName() string {
	components := []string{""}
	if s.File.Package != nil {
		components = append(components, s.File.GetPackage())
	}
	components = append(components, s.GetName())
	return strings.Join(components, ".")
}

type Method struct {
	*descriptorpb.MethodDescriptorProto
	Service      *Service
	RequestType  *Message
	ResponseType *Message
}

func (m *Method) FullyName() string {
	var components []string
	components = append(components, m.Service.FullyName())
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

type ResponseFile struct {
	*pluginpb.CodeGeneratorResponse_File
	GoPkg GoPackage
}
