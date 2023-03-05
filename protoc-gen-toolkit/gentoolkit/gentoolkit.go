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

package gentoolkit

import (
	"fmt"
	"go/format"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lastbackend/toolkit/protoc-gen-toolkit/descriptor"
	toolkit_annotattions "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	defaultRepoRootPath = "github.com/lastbackend/toolkit"
)

type Generator interface {
	Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}

type generator struct {
	desc        *descriptor.Descriptor
	baseImports []descriptor.GoPackage
}

func New(desc *descriptor.Descriptor) Generator {
	return &generator{
		desc: desc,
	}
}

func (g *generator) Generate(files []*descriptor.File) ([]*descriptor.ResponseFile, error) {
	contentFiles := make([]*descriptor.ResponseFile, 0)

	for _, file := range files {
		if len(file.Services) == 0 {
			continue
		}

		dir := filepath.Dir(file.GeneratedFilenamePrefix)
		name := filepath.Base(file.GeneratedFilenamePrefix)

		// Generate service
		filename := filepath.Join(dir, name+"_service.pb.toolkit.go")
		genFiles, err := g.generate(filename, file, g.generateService)
		if err != nil {
			return nil, err
		}
		if files != nil {
			contentFiles = append(contentFiles, genFiles...)
		}

		if g.hasServiceMethods(file) {
			// Generate rpc client
			filename = filepath.Join(dir, "client", name+".pb.toolkit.rpc.go")
			genFiles, err = g.generate(filename, file, g.generateClient)
			if err != nil {
				return nil, err
			}
			if files != nil {
				contentFiles = append(contentFiles, genFiles...)
			}

			// Generate mockery
			if proto.HasExtension(file.Options, toolkit_annotattions.E_TestsSpec) {
				filename = filepath.Join(dir, "tests", name+".pb.toolkit.mockery.go")
				genFiles, err = g.generate(filename, file, g.generateTestStubs)
				if err != nil {
					return nil, err
				}
				if files != nil {
					contentFiles = append(contentFiles, genFiles...)
				}
			}
		}
	}

	return contentFiles, nil
}

type genFunc func(file *descriptor.File) ([]byte, error)

func (g *generator) generate(filename string, file *descriptor.File, fn genFunc) ([]*descriptor.ResponseFile, error) {
	files := make([]*descriptor.ResponseFile, 0)
	content, err := fn(file)
	if err != nil {
		return nil, err
	}
	if content != nil {
		files = append(files, &descriptor.ResponseFile{
			GoPkg: file.GoPkg,
			CodeGeneratorResponse_File: &pluginpb.CodeGeneratorResponse_File{
				Name:    proto.String(filename),
				Content: proto.String(string(content)),
			},
		})
	}
	return files, nil
}

func (g *generator) generateService(file *descriptor.File) ([]byte, error) {

	var pluginImportsExists = make(map[string]bool, 0)
	var clientImportsExists = make(map[string]bool, 0)
	var plugins = make(map[string][]*descriptor.Plugin, 0)
	var clients = make(map[string]*Client, 0)
	var pkgExists = make(map[string]bool, 0)
	var plgExists = make(map[string]bool, 0)
	var imports = g.prepareImports([]string{
		"context",
		"encoding/json",
		"io",
		"net/http",
		"runtime github.com/lastbackend/toolkit/pkg/runtime",
		"controller github.com/lastbackend/toolkit/pkg/runtime/controller",
		"tk_http github.com/lastbackend/toolkit/pkg/server/http",
		"tk_ws github.com/lastbackend/toolkit/pkg/server/http/websockets",
		"toolkit github.com/lastbackend/toolkit",
		"errors github.com/lastbackend/toolkit/pkg/server/http/errors",
		"emptypb google.golang.org/protobuf/types/known/emptypb",
	})

	for _, pkg := range g.baseImports {
		pkgExists[pkg.Path] = true
		imports = append(imports, pkg)
	}

	if file.Options != nil && proto.HasExtension(file.Options, toolkit_annotattions.E_Plugins) {
		ePlugins := proto.GetExtension(file.Options, toolkit_annotattions.E_Plugins)
		if ePlugins != nil {
			plgs := ePlugins.([]*toolkit_annotattions.Plugin)
			for _, props := range plgs {
				if _, ok := plugins[props.Plugin]; !ok {
					plugins[props.Plugin] = make([]*descriptor.Plugin, 0)
				}
				key := fmt.Sprintf("%s/%s", props.Plugin, props.Prefix)
				if _, ok := plgExists[key]; ok {
					continue
				}
				if _, ok := pluginImportsExists[props.Plugin]; !ok {
					imports = append(imports, descriptor.GoPackage{
						Path: fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
						Name: path.Base(fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
					})
				}
				plugins[props.Plugin] = append(plugins[props.Plugin], &descriptor.Plugin{
					Plugin:   props.Plugin,
					Prefix:   props.Prefix,
					Pkg:      strings.ToLower(props.Plugin),
					IsGlobal: true,
				})

				plgExists[key] = true
			}
		}
	}

	for _, svc := range file.Services {
		for _, m := range svc.Methods {
			pkg := m.RequestType.File.GoPkg
			if pkg == file.GoPkg || pkgExists[pkg.Path] {
				continue
			}
			pkgExists[pkg.Path] = true
			imports = append(imports, pkg)
		}

		svc.Plugins = make(map[string][]*descriptor.Plugin, 0)

		if svc.Options != nil && proto.HasExtension(svc.Options, toolkit_annotattions.E_Service) {
			eService := proto.GetExtension(svc.Options, toolkit_annotattions.E_Service)
			if eService != nil {
				ss := eService.(*toolkit_annotattions.Service)
				if ss.Plugins != nil {
					for _, props := range ss.Plugins {
						if _, ok := svc.Plugins[props.Plugin]; !ok {
							svc.Plugins[props.Plugin] = make([]*descriptor.Plugin, 0)
						}
						key := fmt.Sprintf("%s/%s", props.Plugin, props.Prefix)
						if _, ok := plgExists[key]; ok {
							continue
						}
						if _, ok := pluginImportsExists[props.Plugin]; !ok {
							imports = append(imports, descriptor.GoPackage{
								Path: fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
								Name: path.Base(fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
							})
						}

						svc.Plugins[props.Plugin] = append(svc.Plugins[props.Plugin], &descriptor.Plugin{
							Plugin: props.Plugin,
							Prefix: props.Prefix,
							Pkg:    strings.ToLower(props.Plugin),
						})

						plgExists[key] = true
					}
				}
			}
		}
	}

	if file.Options != nil && proto.HasExtension(file.Options, toolkit_annotattions.E_Clients) {
		eClients := proto.GetExtension(file.Options, toolkit_annotattions.E_Clients)
		if eClients != nil {
			clnts := eClients.([]*toolkit_annotattions.Client)
			for _, value := range clnts {
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

	to := tplServiceOptions{
		File:    file,
		Imports: imports,
		Clients: clients,
		Plugins: plugins,
	}

	content, err := applyServiceTemplate(to)
	if err != nil {
		return nil, err
	}

	return format.Source([]byte(content))
}

func (g *generator) generateClient(file *descriptor.File) ([]byte, error) {

	pkgImports := []string{
		"context context",
		"client github.com/lastbackend/toolkit/pkg/client",
		"emptypb google.golang.org/protobuf/types/known/emptypb",
	}

	var imports = g.prepareImports(pkgImports)

	for _, svc := range file.Services {
		for _, m := range svc.Methods {
			if m.IsWebsocket || m.IsWebsocketProxy {
				continue
			}

			pkg := m.RequestType.File.GoPkg
			if strings.HasPrefix(m.RequestType.File.GoPkg.Path, "./") {
				pkg.Path = filepath.Join(file.GoPkg.Name, m.RequestType.File.GoPkg.Path)
			}

			imports = append(imports, pkg)
		}
	}

	var clients = make(map[string]*Client, 0)

	if file.Options != nil && proto.HasExtension(file.Options, toolkit_annotattions.E_Clients) {
		eClients := proto.GetExtension(file.Options, toolkit_annotattions.E_Clients)
		if eClients != nil {
			clnts := eClients.([]*toolkit_annotattions.Client)
			for _, value := range clnts {
				clients[value.Service] = &Client{
					Service: value.Service,
					Pkg:     value.Package,
				}
			}
		}
	}

	to := tplClientOptions{
		File:    file,
		Imports: imports,
		Clients: clients,
	}

	content, err := applyClientTemplate(to)
	if err != nil {
		return nil, err
	}

	return format.Source([]byte(content))
}

func (g *generator) generateTestStubs(file *descriptor.File) ([]byte, error) {
	ext := proto.GetExtension(file.Options, toolkit_annotattions.E_TestsSpec)
	opts, ok := ext.(*toolkit_annotattions.TestSpec)
	if !ok {
		return nil, nil
	}

	baseImports := []string{
		"context context",
		"client github.com/lastbackend/toolkit/pkg/client",
		"emptypb google.golang.org/protobuf/types/known/emptypb",
		fmt.Sprintf("servicepb %s/client", filepath.Dir(file.GeneratedFilenamePrefix)),
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
		content, err := applyTemplateWithMessage(tplMessageOptions{
			File:    file,
			Message: "Warning: You have no mock in provided directory. Please check mockery docs for mocks generation.",
		})
		if err != nil {
			return nil, err
		}

		return format.Source([]byte(content))
	}

	var imports = g.prepareImports(baseImports)

	imports = append(imports, descriptor.GoPackage{
		Path:  fmt.Sprintf(opts.Mockery.Package),
		Name:  path.Base(opts.Mockery.Package),
		Alias: "service_mocks",
	})

	for _, svc := range file.Services {
		for _, m := range svc.Methods {
			imports = append(imports, m.RequestType.File.GoPkg)
		}
	}

	content, err := applyTestTemplate(tplMockeryTestOptions{
		File:    file,
		Imports: imports,
	})
	if err != nil {
		return nil, err
	}

	return format.Source([]byte(content))
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
