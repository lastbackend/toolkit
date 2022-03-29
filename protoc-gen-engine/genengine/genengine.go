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
	"go/format"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"

	"fmt"
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
	StoragePluginType = "Storage"
	CachePluginType   = "Cache"
	BrokerPluginType  = "Broker"
)

type Generator interface {
	Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}

type generator struct {
	desc *descriptor.Descriptor
	opts *Options
}

type Options struct {
	SourcePackage string
}

func New(desc *descriptor.Descriptor, opts *Options) Generator {

	g := &generator{}

	if opts == nil {
		opts = new(Options)
	}
	g.desc = desc
	g.opts = opts

	return g
}

func (g *generator) Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	var (
		files []*descriptor.ResponseFile
	)

	for _, file := range targets {
		if len(file.Services) == 0 {
			continue
		}

		// Generate service
		lbServiceCode, err := g.generateService(file)
		if err != nil {
			return nil, err
		}
		lbServiceFormatted, err := format.Source([]byte(lbServiceCode))
		if err != nil {
			return nil, err
		}

		dir := filepath.Dir(file.GeneratedFilenamePrefix)
		name := filepath.Base(file.GeneratedFilenamePrefix)

		files = append(files, &descriptor.ResponseFile{
			GoPkg: file.GoPkg,
			CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
				Name:    proto.String(filepath.Join(dir, "service", name+".pb.lb.service.go")),
				Content: proto.String(string(lbServiceFormatted)),
			},
		})

		if g.hasServiceMethods(file) {
			// Generate rpc client
			lbClientCode, err := g.generateClient(file)
			if err != nil {
				return nil, err
			}
			lbClientFormatted, err := format.Source([]byte(lbClientCode))
			if err != nil {
				return nil, err
			}

			files = append(files, &descriptor.ResponseFile{
				GoPkg: file.GoPkg,
				CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
					Name:    proto.String(filepath.Join(dir, "client", name+".pb.lb.rpc.go")),
					Content: proto.String(string(lbClientFormatted)),
				},
			})
		} else {
			if err := os.RemoveAll(filepath.Join(dir, "client")); err != nil {
				return nil, err
			}
		}

		// Generate mockery
		if proto.HasExtension(file.Options, engine_annotattions.E_TestsSpec) {
			code, err := g.generateTestStubs(file)
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
					Name:    proto.String(filepath.Join(dir, "tests", name+".pb.lb.mockery.go")),
					Content: proto.String(string(formatted)),
				},
			})
		}
	}

	return files, nil
}

func (g *generator) generateService(file *descriptor.File) (string, error) {

	var pluginImportsExists = make(map[string]bool, 0)
	var clientImportsExists = make(map[string]bool, 0)
	var plugins = make(map[string]map[string]*Plugin, 0)
	var clients = make(map[string]*Client, 0)
	var imports = g.prepareImports([]string{
		"engine github.com/lastbackend/engine",
		"logger github.com/lastbackend/engine/logger",
		"fx go.uber.org/fx",
		"context",
		"os",
		"os/signal",
		"syscall",
	})

	// Add imports for server
	if g.hasServiceMethods(file) {
		imports = append(imports, g.prepareImports([]string{
			"server github.com/lastbackend/engine/server",
			"github.com/lastbackend/engine/client/grpc",
			fmt.Sprintf("proto %s/apis/proto", g.opts.SourcePackage),
		})...)
	}

	for _, svc := range file.Services {
		if svc.Options != nil && proto.HasExtension(svc.Options, engine_annotattions.E_Clients) {
			eClients := proto.GetExtension(svc.Options, engine_annotattions.E_Clients)
			if eClients != nil {
				clnts := eClients.(*engine_annotattions.Clients)
				for _, value := range clnts.Client {
					if _, ok := clientImportsExists[value.Service]; !ok {
						imports = append(imports, descriptor.GoPackage{
							Alias: strings.ToLower(value.Service),
							Path:  filepath.Join(value.Package, "client"),
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

	if file.Options != nil && proto.HasExtension(file.Options, engine_annotattions.E_Plugins) {
		ePlugins := proto.GetExtension(file.Options, engine_annotattions.E_Plugins)
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

	to := tplServiceOptions{
		HasNotServer: !g.hasServiceMethods(file),
		File:         file,
		Imports:      imports,
		Clients:      clients,
		Plugins:      plugins,
	}

	return applyServiceTemplate(to)
}

func (g *generator) generateClient(file *descriptor.File) (string, error) {

	pkgImports := []string{
		"context context",
		fmt.Sprintf("proto %s/apis/proto", g.opts.SourcePackage),
		"github.com/lastbackend/engine/client/grpc",
	}

	var clients = make(map[string]*Client, 0)
	var imports = g.prepareImports(pkgImports)

	for _, svc := range file.Services {
		if svc.Options != nil && proto.HasExtension(svc.Options, engine_annotattions.E_Clients) {
			eClients := proto.GetExtension(svc.Options, engine_annotattions.E_Clients)
			if eClients != nil {
				clnts := eClients.(*engine_annotattions.Clients)
				for _, value := range clnts.Client {
					clients[value.Service] = &Client{
						Service: value.Service,
						Pkg:     value.Package,
					}
				}
			}
		}
	}

	to := tplClientOptions{
		File:    file,
		Imports: imports,
		Clients: clients,
	}

	return applyClientTemplate(to)
}

func (g *generator) generateTestStubs(file *descriptor.File) (string, error) {
	ext := proto.GetExtension(file.Options, engine_annotattions.E_TestsSpec)
	opts, ok := ext.(*engine_annotattions.TestSpec)
	if ok {

		baseImports := []string{
			"context context",
			"grpc github.com/lastbackend/engine/client/grpc",
			fmt.Sprintf("proto %s", filepath.Join(g.opts.SourcePackage, filepath.Dir(file.GeneratedFilenamePrefix))),
			fmt.Sprintf("client %s/apis/proto/client", g.opts.SourcePackage),
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

func (g *generator) hasServiceMethods(file *descriptor.File) bool {
	for _, service := range file.Services {
		if len(service.Methods) > 0 {
			return true
		}
	}
	return false
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
