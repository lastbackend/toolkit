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

package generator

import (
  "flag"

  "github.com/lastbackend/toolkit/protoc-gen-toolkit/descriptor"
  "github.com/lastbackend/toolkit/protoc-gen-toolkit/gentoolkit"
  "google.golang.org/protobuf/compiler/protogen"
  "google.golang.org/protobuf/types/pluginpb"
)

type Generator struct {
  files []*descriptor.File
}

func Init() *Generator {
  g := new(Generator)
  g.files = make([]*descriptor.File, 0)
  return g
}

func (g *Generator) Run() error {
  var flagSet flag.FlagSet

  protogen.Options{
    ParamFunc: flagSet.Set,
  }.Run(func(gen *protogen.Plugin) (err error) {

    gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

    desc := descriptor.NewDescriptor()

    if err := desc.LoadFromPlugin(gen); err != nil {
      return err
    }

    if err := g.LoadFiles(gen, desc); err != nil {
      return err
    }

    // Generate generates a *.pb.toolkit.go file containing Toolkit service definitions.
    contentFiles, err := gentoolkit.New(desc).Generate(g.files)
    if err != nil {
      return err
    }

    // Write generated files to disk
    for _, f := range contentFiles {
      genFile := gen.NewGeneratedFile(f.GetName(), protogen.GoImportPath(f.GoPkg.Path))
      if _, err := genFile.Write([]byte(f.GetContent())); err != nil {
        return err
      }
    }

    return nil
  })

  return nil
}

func (g *Generator) LoadFiles(gen *protogen.Plugin, desc *descriptor.Descriptor) (err error) {
  g.files = make([]*descriptor.File, 0)
  for _, f := range gen.Request.FileToGenerate {
    file, err := desc.FindFile(f)
    if err != nil {
      return err
    }
    g.files = append(g.files, file)
  }
  return err
}
