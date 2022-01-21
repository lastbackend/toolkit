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
	"fmt"
	"github.com/lastbackend/engine/protoc-gen-engine/descriptor"
	engine_annotattions "github.com/lastbackend/engine/protoc-gen-engine/engine/options"
	"go/format"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultRepoRootPath = "github.com/lastbackend/engine"
)

const (
	StoragePluginType string = "Storage"
	CachePluginType          = "Cache"
	BrokerPluginType         = "Broker"
)

type Generator interface {
	Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}

type generator struct {
	desc *descriptor.Descriptor
}

func New(desc *descriptor.Descriptor) Generator {
	return &generator{
		desc: desc,
	}
}

func (g *generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var (
		files []*descriptor.ResponseFile
	)
	for _, file := range targets {
		if len(file.Services) == 0 {
			continue
		}

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
				Name:    proto.String(file.GeneratedFilenamePrefix + ".pb.lb.go"),
				Content: proto.String(string(formatted)),
			},
		})

		if proto.HasExtension(file.Options, engine_annotattions.E_TestsSpec) {
			code, err := g.generateTestStubs(file)
			if err != nil {
				return nil, err
			}

			formatted, err := format.Source([]byte(code))
			if err != nil {
				return nil, err
			}

			dir := filepath.Dir(file.GeneratedFilenamePrefix)
			name := filepath.Base(file.GeneratedFilenamePrefix)

			files = append(files, &descriptor.ResponseFile{
				GoPkg: file.GoPkg,
				CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
					Name:    proto.String(filepath.Join(dir, "tests", name+".pb.lb.mockery.go")),
					Content: proto.String(string(formatted)),
				},
			})
		}
	}

	return files, nil
}

func (g *generator) generate(file *descriptor.File) (string, error) {

	var pluginImportsExists map[string]bool
	var clientImportsExists map[string]bool
	var imports = g.prepareImports([]string{
		"context context",
		"engine github.com/lastbackend/engine",
		"logger github.com/lastbackend/engine/logger",
		"server github.com/lastbackend/engine/server",
		"github.com/lastbackend/engine/client/grpc",
		"fx go.uber.org/fx",
	})

	for _, pkg := range imports {
		imports = append(imports, pkg)
	}

	plugins := make(map[string]map[string]*Plugin, 0)
	clients := make(map[string]*Client, 0)

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
							Pkg:    fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), StoragePluginType),
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
							Pkg:    fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), CachePluginType),
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
							Pkg:    fmt.Sprintf("%s.%s", strings.ToLower(props.Plugin), BrokerPluginType),
						}
					}
				}
			}
		}

		if svc.Options != nil && proto.HasExtension(svc.Options, engine_annotattions.E_Clients) {
			eClients := proto.GetExtension(svc.Options, engine_annotattions.E_Clients)
			if eClients != nil {
				clnts := eClients.(*engine_annotattions.Clients)
				for _, value := range clnts.Client {
					if _, ok := clientImportsExists[value.Service]; !ok {
						imports = append(imports, descriptor.GoPackage{
							Alias: strings.ToLower(value.Service),
							Path:  value.Package,
						})
					}
					clients[value.Service] = &Client{
						Service: value.Service,
						Pkg:     value.Package,
					}
				}
			}
		}
	}

	to := tplOptions{
		File:    file,
		Imports: imports,
		Clients: clients,
		Plugins: plugins,
	}

	return applyTemplate(to)
}

func (g *generator) generateTestStubs(file *descriptor.File) (string, error) {
	ext := proto.GetExtension(file.Options, engine_annotattions.E_TestsSpec)
	opts, ok := ext.(*engine_annotattions.TestSpec)
	if ok {

		baseImports := []string{
			"context context",
			"grpc github.com/lastbackend/engine/client/grpc",
			"mock github.com/stretchr/testify/mock",
			fmt.Sprintf("engine %s", filepath.Join(*file.Package, filepath.Dir(file.GeneratedFilenamePrefix))),
		}

		if len(opts.Mockery.Package) == 0 {
			opts.Mockery.Package = "github.com/dummy/dummy"
		}

		var dirErr error
		dir := filepath.Join(os.Getenv("GOPATH"), "src", opts.Mockery.Package)
		if ok, _ := existsFileOrDir(dir); !ok {
			dirErr = fmt.Errorf("directory %s does not exist", dir)
		}
		if ok, _ := dirIsEmpty(dir); ok {
			dirErr = fmt.Errorf("directory %s is empty", dir)
		}
		if dirErr != nil {
			return applyTemplateWithMessage(tplMessageOptions{
				File:    file,
				Message: "Warning: You have no mock in provided directory. Please check mockery docs for mocks generation.",
			})
		}

		var imports = g.prepareImports(baseImports)

		imports = append(imports, descriptor.GoPackage{
			Path:  fmt.Sprintf(opts.Mockery.Package),
			Name:  path.Base(opts.Mockery.Package),
			Alias: "service_mocks",
		})

		content, err := applyTestTemplate(tplMockeryTestOptions{
			File:    file,
			Imports: imports,
		})
		if err != nil {
			return "", err
		}

		return content, nil
	}

	return "", nil
}

func (g *generator) prepareImports(importList []string) []descriptor.GoPackage {
	var imports []descriptor.GoPackage
	for _, pkgPath := range importList {
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
			if err := g.desc.ReserveGoPackageAlias(pkg.Name, pkg.Path); err != nil {
				for i := 0; ; i++ {
					alias := fmt.Sprintf("%s_%d", pkg.Name, i)
					if err := g.desc.ReserveGoPackageAlias(alias, pkg.Path); err != nil {
						continue
					}
					pkg.Alias = alias
					break
				}
			}
		}
		imports = append(imports, pkg)
	}

	return imports
}

func existsFileOrDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func dirIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
