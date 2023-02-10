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
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/lastbackend/toolkit/protoc-gen-toolkit/descriptor"
	toolkit_annotattions "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options"
	options "google.golang.org/genproto/googleapis/api/annotations"
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
	opts        *Options
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
				Name:    proto.String(filepath.Join(dir, name+"_service.pb.toolkit.go")),
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
					Name:    proto.String(filepath.Join(dir, "client", name+".pb.toolkit.rpc.go")),
					Content: proto.String(string(lbClientFormatted)),
				},
			})

			// Generate mockery
			if proto.HasExtension(file.Options, toolkit_annotattions.E_TestsSpec) {
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
						Name:    proto.String(filepath.Join(dir, "tests", name+".pb.toolkit.mockery.go")),
						Content: proto.String(string(formatted)),
					},
				})
			}
		}

	}

	return files, nil
}

func (g *generator) generateService(file *descriptor.File) (string, error) {

	var pluginImportsExists = make(map[string]bool, 0)
	var clientImportsExists = make(map[string]bool, 0)
	var plugins = make(map[string][]*Plugin, 0)
	var clients = make(map[string]*Client, 0)
	var pkgExists = make(map[string]bool, 0)
	var imports = g.prepareImports([]string{
		"context",
		"io",
		"os",
		"os/signal",
		"net/http",
		"syscall",
		"encoding/json",
		"toolkit github.com/lastbackend/toolkit",
		"server github.com/lastbackend/toolkit/pkg/server",
		"router github.com/lastbackend/toolkit/pkg/router",
		"logger github.com/lastbackend/toolkit/pkg/logger",
		"grpc github.com/lastbackend/toolkit/pkg/client/grpc",
		"errors github.com/lastbackend/toolkit/pkg/router/errors",
		"ws github.com/lastbackend/toolkit/pkg/router/ws",
		"emptypb google.golang.org/protobuf/types/known/emptypb",
		"fx go.uber.org/fx",
	})

	for _, pkg := range g.baseImports {
		pkgExists[pkg.Path] = true
		imports = append(imports, pkg)
	}

	// Add imports for server
	if g.hasServiceMethods(file) {

		for _, svc := range file.Services {
			for _, m := range svc.RPCMethods {
				imports = append(imports, m.RequestType.File.GoPkg)
			}
		}
	}

	for _, svc := range file.Services {
		for _, m := range svc.ProxyMethods {
			if m.Options != nil && proto.HasExtension(m.Options, toolkit_annotattions.E_Proxy) {

				pOpts, err := getProxyOptions(m)
				if err != nil {
					return "", err
				}

				if pOpts != nil {
					switch true {
					case pOpts.GetWs() != nil:
						wsOpts := pOpts.GetWs()
						b := &descriptor.Binding{
							RpcMethod:    m.GetName(),
							HttpMethod:   http.MethodGet,
							HttpPath:     wsOpts.Path,
							HttpParams:   getVariablesFromPath(wsOpts.Path),
							RequestType:  m.RequestType,
							ResponseType: m.ResponseType,
							Websocket:    true,
						}

						svc.Bindings = append(svc.Bindings, b)
					case pOpts.GetHttp() != nil && proto.HasExtension(m.Options, options.E_Http):
						rOpts := pOpts.GetHttp()

						opts, err := getHTTPOptions(m)
						if err != nil {
							return "", err
						}

						if opts != nil {
							var (
								httpMethod string
								httpPath   string
							)
							switch {
							case opts.GetGet() != "":
								httpMethod = "http.MethodGet"
								httpPath = opts.GetGet()
							case opts.GetPut() != "":
								httpMethod = "http.MethodPut"
								httpPath = opts.GetPut()
							case opts.GetPost() != "":
								httpMethod = "http.MethodPost"
								httpPath = opts.GetPost()
							case opts.GetDelete() != "":
								httpMethod = "http.MethodDelete"
								httpPath = opts.GetDelete()
							case opts.GetPatch() != "":
								httpMethod = "http.MethodPatch"
								httpPath = opts.GetPatch()
							default:
								continue
							}

							b := &descriptor.Binding{
								Service:      rOpts.GetService(),
								RpcPath:      rOpts.GetMethod(),
								RpcMethod:    m.GetName(),
								HttpMethod:   httpMethod,
								HttpPath:     httpPath,
								HttpParams:   getVariablesFromPath(httpPath),
								RequestType:  m.RequestType,
								ResponseType: m.ResponseType,
								Stream:       m.GetClientStreaming(),
							}

							b, err := newBinding(m, opts, rOpts)
							if err != nil {
								continue
							}
							svc.Bindings = append(svc.Bindings, b)

							for _, additional := range opts.GetAdditionalBindings() {
								if len(additional.AdditionalBindings) > 0 {
									continue
								}
								b, err := newBinding(m, additional, rOpts)
								if err != nil {
									continue
								}
								svc.Bindings = append(svc.Bindings, b)
							}
						}
					}

				}
			}

			pkg := m.RequestType.File.GoPkg
			if pkg == file.GoPkg || pkgExists[pkg.Path] {
				continue
			}

			pkgExists[pkg.Path] = true
			imports = append(imports, pkg)
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

	if file.Options != nil && proto.HasExtension(file.Options, toolkit_annotattions.E_Plugins) {
		ePlugins := proto.GetExtension(file.Options, toolkit_annotattions.E_Plugins)
		if ePlugins != nil {
			plgs := ePlugins.([]*toolkit_annotattions.Plugin)
			for _, props := range plgs {
				if _, ok := plugins[props.Plugin]; !ok {
					plugins[props.Plugin] = make([]*Plugin, 0)
				}
				if _, ok := pluginImportsExists[props.Plugin]; !ok {
					imports = append(imports, descriptor.GoPackage{
						Path: fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin)),
						Name: path.Base(fmt.Sprintf("%s/plugin/%s", defaultRepoRootPath, strings.ToLower(props.Plugin))),
					})
				}
				plugins[props.Plugin] = append(plugins[props.Plugin], &Plugin{
					Plugin: props.Plugin,
					Prefix: props.Prefix,
					Pkg:    strings.ToLower(props.Plugin),
				})
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
		"grpc github.com/lastbackend/toolkit/pkg/client/grpc",
		"emptypb google.golang.org/protobuf/types/known/emptypb",
	}

	var imports = g.prepareImports(pkgImports)

	for _, svc := range file.Services {
		for _, m := range svc.RPCMethods {
			imports = append(imports, m.RequestType.File.GoPkg)
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

	return applyClientTemplate(to)
}

func (g *generator) generateTestStubs(file *descriptor.File) (string, error) {
	ext := proto.GetExtension(file.Options, toolkit_annotattions.E_TestsSpec)
	opts, ok := ext.(*toolkit_annotattions.TestSpec)
	if ok {

		baseImports := []string{
			"context context",
			"grpc github.com/lastbackend/toolkit/pkg/client/grpc",
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

		for _, svc := range file.Services {
			for _, m := range svc.RPCMethods {
				imports = append(imports, m.RequestType.File.GoPkg)
			}
		}

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
		if len(service.RPCMethods) > 0 {
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

func newBinding(m *descriptor.Method, opts *options.HttpRule, rOpts *toolkit_annotattions.HttpProxy) (*descriptor.Binding, error) {
	var (
		httpMethod string
		httpPath   string
	)
	switch {
	case opts.GetGet() != "":
		httpMethod = "http.MethodGet"
		httpPath = opts.GetGet()
	case opts.GetPut() != "":
		httpMethod = "http.MethodPut"
		httpPath = opts.GetPut()
		if opts.Body == "" {
			opts.Body = "*"
		}
	case opts.GetPost() != "":
		httpMethod = "http.MethodPost"
		httpPath = opts.GetPost()
		if opts.Body == "" {
			opts.Body = "*"
		}
	case opts.GetDelete() != "":
		httpMethod = "http.MethodDelete"
		httpPath = opts.GetDelete()
	case opts.GetPatch() != "":
		httpMethod = "http.MethodPatch"
		httpPath = opts.GetPatch()
		if opts.Body == "" {
			opts.Body = "*"
		}
	default:
		return nil, fmt.Errorf("not fount method")
	}

	return &descriptor.Binding{
		Service:      rOpts.GetService(),
		RpcPath:      rOpts.GetMethod(),
		RpcMethod:    m.GetName(),
		HttpMethod:   httpMethod,
		HttpPath:     httpPath,
		HttpParams:   getVariablesFromPath(httpPath),
		RequestType:  m.RequestType,
		ResponseType: m.ResponseType,
		Stream:       m.GetClientStreaming(),
		RawBody:      opts.Body,
	}, nil
}

func getHTTPOptions(m *descriptor.Method) (*options.HttpRule, error) {
	if m.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(m.Options, options.E_Http) {
		return nil, nil
	}
	ext := proto.GetExtension(m.Options, options.E_Http)
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, fmt.Errorf("extension is not an HttpRule")
	}
	return opts, nil
}

func getProxyOptions(m *descriptor.Method) (*toolkit_annotattions.Proxy, error) {
	if m.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(m.Options, toolkit_annotattions.E_Proxy) {
		return nil, nil
	}
	ext := proto.GetExtension(m.Options, toolkit_annotattions.E_Proxy)
	opts, ok := ext.(*toolkit_annotattions.Proxy)
	if !ok {
		return nil, fmt.Errorf("extension is not an Proxy")
	}
	return opts, nil
}

func getVariablesFromPath(path string) (variables []string) {
	if path == "" {
		return []string{}
	}

	for path != "" {
		firstIndex := -1
		lastIndex := -1

		firstIndex = strings.IndexAny(path, "{")
		lastIndex = strings.IndexAny(path, "}")

		if firstIndex > -1 && lastIndex > -1 {
			field := path[firstIndex+1 : lastIndex]
			if len(strings.TrimSpace(field)) > 0 {
				variables = append(variables, field)
			}
			path = path[lastIndex+1:]
		}

		if firstIndex == -1 || lastIndex == -1 {
			path = path[1:]
		}
	}

	return variables
}
