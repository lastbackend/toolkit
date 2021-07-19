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

package genengine

import (
	"github.com/lastbackend/engine/protoc-gen-engine/descriptor"
	engine_annotattions "github.com/lastbackend/engine/protoc-gen-engine/engine/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"fmt"
	"go/format"
	"path"
	"strings"
)

const (
	defaultRepoRootPath = "github.com/lastbackend/engine"
)

const (
	StoragePluginType string = "Storage"
	CachePluginType          = "Cache"
	BrokerPluginType         = "Broker"
	ClientPluginType         = "Client"
)

type Generator interface {
	Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}

type generator struct {
	desc        *descriptor.Descriptor
	baseImports []descriptor.GoPackage
}

func New(desc *descriptor.Descriptor) Generator {
	var imports []descriptor.GoPackage
	for _, pkgPath := range []string{
		"context context",
		"engine github.com/lastbackend/engine",
		"logger github.com/lastbackend/engine/logger",
		"plugin github.com/lastbackend/engine/plugin",
		"server github.com/lastbackend/engine/server/grpc",
	} {
		var pkg descriptor.GoPackage

		match := strings.Split(pkgPath, " ")
		if len(match) == 2 {
			pkg = descriptor.GoPackage{
				Path:  match[1],
				Name:  path.Base(match[1]),
				Alias: match[0],
			}
		} else {
			pkg = descriptor.GoPackage{
				Path: pkgPath,
				Name: path.Base(pkgPath),
			}
		}

		if len(pkg.Alias) == 0 {
			if err := desc.ReserveGoPackageAlias(pkg.Name, pkg.Path); err != nil {
				for i := 0; ; i++ {
					alias := fmt.Sprintf("%s_%d", pkg.Name, i)
					if err := desc.ReserveGoPackageAlias(alias, pkg.Path); err != nil {
						continue
					}
					pkg.Alias = alias
					break
				}
			}
		}
		imports = append(imports, pkg)
	}

	return &generator{
		desc:        desc,
		baseImports: imports,
	}
}

func (g *generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var files []*descriptor.ResponseFile
	for _, file := range targets {
		code, err := g.generate(file)
		if err != nil {
			return nil, err
		}
		formatted, err := format.Source([]byte(code))
		if err != nil {
			return nil, err
		}
		files = append(files, &descriptor.ResponseFile{
			GoPkg: file.GoPkg,
			CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
				Name:    proto.String(file.GeneratedFilenamePrefix + ".pb.engine.go"),
				Content: proto.String(string(formatted)),
			},
		})
	}
	return files, nil
}

func (g *generator) generate(file *descriptor.File) (string, error) {
	pkgExists := make(map[string]bool)

	var imports []descriptor.GoPackage
	var pluginImportsExists map[string]bool
	for _, pkg := range g.baseImports {
		pkgExists[pkg.Path] = true
		imports = append(imports, pkg)
	}

	plugins := make(map[string]map[string]*Plugin, 0)
	for _, svc := range file.Services {
		if svc.Options != nil && proto.HasExtension(svc.Options, engine_annotattions.E_Plugins) {
			ePlugins := proto.GetExtension(svc.Options, engine_annotattions.E_Plugins)
			if ePlugins != nil {
				plgs := ePlugins.(*engine_annotattions.Plugins)
				if len(plgs.Storage) > 0 {
					plugins[StoragePluginType] = make(map[string]*Plugin, 0)
					for name, props := range plgs.Storage {
						if _, ok := pluginImportsExists[props.Plugin]; !ok {
							imports = append(imports, descriptor.GoPackage{
								Path: fmt.Sprintf("%s/plugin/storage/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
								Name: path.Base(fmt.Sprintf("%s/plugin/storage/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
							})
						}
						plugins[StoragePluginType][name] = &Plugin{
							Plugin: props.Plugin,
							Prefix: props.Prefix,
							Pkg: fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), StoragePluginType),
						}
					}
				}
				if len(plgs.Cache) > 0 {
					plugins[CachePluginType] = make(map[string]*Plugin, 0)
					for name, props := range plgs.Cache {
						if _, ok := pluginImportsExists[props.Plugin]; !ok {
							imports = append(imports, descriptor.GoPackage{
								Path: fmt.Sprintf("%s/plugin/cache/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
								Name: path.Base(fmt.Sprintf("%s/plugin/cache/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
							})
						}
						plugins[CachePluginType][name] = &Plugin{
							Plugin: props.Plugin,
							Prefix: props.Prefix,
							Pkg: fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), CachePluginType),
						}
					}
				}
				if len(plgs.Broker) > 0 {
					plugins[BrokerPluginType] = make(map[string]*Plugin, 0)
					for name, props := range plgs.Broker {
						if _, ok := pluginImportsExists[props.Plugin]; !ok {
							imports = append(imports, descriptor.GoPackage{
								Path: fmt.Sprintf("%s/plugin/broker/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
								Name: path.Base(fmt.Sprintf("%s/plugin/broker/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
							})
						}
						plugins[BrokerPluginType][name] = &Plugin{
							Plugin: props.Plugin,
							Prefix: props.Prefix,
							Pkg: fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), BrokerPluginType),
						}
					}
				}
				if len(plgs.Client) > 0 {
					plugins[ClientPluginType] = make(map[string]*Plugin, 0)
					for name, props := range plgs.Client {
						if _, ok := pluginImportsExists[props.Plugin]; !ok {
							imports = append(imports, descriptor.GoPackage{
								Path: fmt.Sprintf("%s/plugin/rpc/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
								Name: path.Base(fmt.Sprintf("%s/plugin/rpc/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
							})
						}
						plugins[ClientPluginType][name] = &Plugin{
							Plugin: props.Plugin,
							Prefix: props.Prefix,
							Pkg: fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), ClientPluginType),
						}
					}
				}
			}
		}

		for _, m := range svc.Methods {
			pkg := m.RequestType.File.GoPkg
			if pkg == file.GoPkg || pkgExists[pkg.Path] {
				continue
			}
			pkgExists[pkg.Path] = true
			imports = append(imports, pkg)
		}
	}

	to := tplOptions{
		File:    file,
		Imports: imports,
		Plugins: plugins,
	}

	return applyTemplate(to)
}
