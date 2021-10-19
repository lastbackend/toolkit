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

package generator

import (
	"github.com/lastbackend/engine/protoc-gen-engine/descriptor"
	"github.com/lastbackend/engine/protoc-gen-engine/genengine"
	"github.com/lastbackend/engine/protoc-gen-engine/genscripts"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"flag"
	"io"
)

var (
	DefaultVersion = "1.0.0"
	DefaultName    = "engine"
)

type Generator struct {
	targets []*descriptor.File

	out io.Writer

	// TODO: Implement use version
	version string
	// TODO: Implement use name
	name string
	// TODO: Implement debug logs
	debug bool

	skipDockerfile bool
	skipHelm       bool
}

func Init(opts ...Option) *Generator {
	g := new(Generator)

	for _, opt := range opts {
		opt(g)
	}

	if g.name == "" {
		g.name = DefaultName
	}

	g.targets = make([]*descriptor.File, 0)

	return g
}

func (g *Generator) Run() error {
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {

		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		desc := descriptor.NewDescriptor()

		// Generate engine files
		generatorEngine := genengine.New(desc)

		if err := desc.LoadFromPlugin(gen); err != nil {
			return err
		}

		for _, target := range gen.Request.FileToGenerate {
			f, err := desc.FindFile(target)
			if err != nil {
				return err
			}
			g.targets = append(g.targets, f)
		}

		engineFiles, err := generatorEngine.Generate(g.targets)
		if err != nil {
			return err
		}
		for _, f := range engineFiles {
			genFile := gen.NewGeneratedFile(f.GetName(), protogen.GoImportPath(f.GoPkg.Path))
			if _, err := genFile.Write([]byte(f.GetContent())); err != nil {
				return err
			}
		}

		// Generate scripts scripts
		generatorScripts := genscripts.New()

		scriptFiles, err := generatorScripts.GenerateDockerfile(g.targets)
		if err != nil {
			return err
		}
		for _, f := range scriptFiles {
			if f.Rewrite {
				genFile := gen.NewGeneratedFile(f.GetName(), protogen.GoImportPath(f.GoPkg.Path))
				if _, err := genFile.Write([]byte(f.GetContent())); err != nil {
					return err
				}
			}
		}

		return nil
	})
	return nil
}
