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
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Descriptor struct {
	packageAliasesMap map[string]string
	fileMap           map[string]*File
	messageMap        map[string]*Message
	enumMap           map[string]*Enum
	packageMap        map[string]string
}

func NewDescriptor() *Descriptor {
	return &Descriptor{
		packageAliasesMap: make(map[string]string),
		fileMap:           make(map[string]*File),
		messageMap:        make(map[string]*Message),
		enumMap:           make(map[string]*Enum),
		packageMap:        make(map[string]string),
	}
}

func (d *Descriptor) LoadFromPlugin(gen *protogen.Plugin) error {
	for filePath, f := range gen.FilesByPath {
		d.loadFile(filePath, f)
	}

	for filePath, f := range gen.FilesByPath {
		if !f.Generate {
			continue
		}

		file := d.fileMap[filePath]

		if err := d.loadServices(file); err != nil {
			return err
		}
	}

	return nil
}

func (d *Descriptor) loadFile(filePath string, file *protogen.File) {
	pkg := GoPackage{
		Path: string(file.GoImportPath),
		Name: string(file.GoPackageName),
	}

	if err := d.ReserveGoPackageAlias(pkg.Name, pkg.Path); err != nil {
		for i := 0; ; i++ {
			alias := fmt.Sprintf("%s_%d", pkg.Name, i)
			if err := d.ReserveGoPackageAlias(alias, pkg.Path); err == nil {
				pkg.Alias = alias
				break
			}
		}
	}

	f := &File{
		FileDescriptorProto:     file.Proto,
		GoPkg:                   pkg,
		GeneratedFilenamePrefix: file.GeneratedFilenamePrefix,
	}

	d.fileMap[filePath] = f

	d.registerMsg(f, nil, file.Proto.MessageType)
	d.registerEnum(f, nil, file.Proto.EnumType)
}

func (d *Descriptor) ReserveGoPackageAlias(alias, pkgpath string) error {
	if taken, ok := d.packageAliasesMap[alias]; ok {
		if taken == pkgpath {
			return nil
		}
		return fmt.Errorf("package name %s is already taken. Use another alias", alias)
	}
	d.packageAliasesMap[alias] = pkgpath
	return nil
}

func (d *Descriptor) registerMsg(file *File, outerPath []string, messages []*descriptorpb.DescriptorProto) {
	for index, message := range messages {
		m := &Message{
			File:            file,
			Outers:          outerPath,
			DescriptorProto: message,
			Index:           index,
		}

		for _, fd := range message.GetField() {
			m.Fields = append(m.Fields, &Field{
				Message:              m,
				FieldDescriptorProto: fd,
			})
		}

		file.Messages = append(file.Messages, m)
		d.messageMap[m.FullyName()] = m

		var outers []string
		outers = append(outers, outerPath...)
		outers = append(outers, m.GetName())

		d.registerMsg(file, outers, m.GetNestedType())
		d.registerEnum(file, outers, m.GetEnumType())
	}
}

func (d *Descriptor) registerEnum(file *File, outerPath []string, enums []*descriptorpb.EnumDescriptorProto) {
	for index, enum := range enums {
		e := &Enum{
			File:                file,
			Outers:              outerPath,
			EnumDescriptorProto: enum,
			Index:               index,
		}

		file.Enums = append(file.Enums, e)
		d.enumMap[e.FullyName()] = e
	}
}

func (d *Descriptor) FindFile(name string) (*File, error) {
	file, ok := d.fileMap[name]
	if !ok {
		return nil, fmt.Errorf("no such file: %s", name)
	}

	return file, nil
}
