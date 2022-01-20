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

package genscripts

import (
	"github.com/lastbackend/engine/protoc-gen-engine/descriptor"
	annotations "github.com/lastbackend/engine/protoc-gen-engine/engine/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type Generator interface {
	GenerateDockerfile(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}

type generator struct {
}

func New() Generator {
	return &generator{}
}

func (g *generator) GenerateDockerfile(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile
	for _, f := range targets {
		if proto.HasExtension(f.Options, annotations.E_DockerfileSpec) {
			ext := proto.GetExtension(f.Options, annotations.E_DockerfileSpec)
			opts, ok := ext.(*annotations.DockerfileSpec)
			if ok {
				if len(opts.Package) == 0 {
					opts.Package = "github.com/dummy/dummy"
				}
				dockerfileContent, err := applyDockerfileTemplate(tplDockerfileOptions{
					Package:         opts.Package,
					RewriteIfExists: opts.RewriteIfExists,
					Expose:          opts.Expose,
					Commands:        opts.Commands,
				})
				if err != nil {
					return nil, err
				}
				files = append(files, &descriptor.ResponseFile{
					Rewrite: opts.RewriteIfExists,
					CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
						Name:    proto.String("Dockerfile"),
						Content: proto.String(dockerfileContent),
					},
				})

				makefileContent, err := applyMakefileTemplate(tplMakefileOptions{})
				if err != nil {
					return nil, err
				}
				files = append(files, &descriptor.ResponseFile{
					Rewrite: opts.RewriteIfExists,
					CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
						Name:    proto.String("Makefile"),
						Content: proto.String(makefileContent),
					},
				})
			}
			break
		}
	}
	return files, nil
}
