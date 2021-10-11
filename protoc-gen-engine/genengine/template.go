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

	"bytes"
	"strings"
	"text/template"
)

type tplOptions struct {
	*descriptor.File
	Imports          []descriptor.GoPackage
	Plugins          map[string]map[string]*Plugin
	Routes           []*Route
	ProtocVersion    string
	GeneratorVersion string
}

type Plugin struct {
	Prefix string
	Plugin string
	Pkg    string
}

type Route struct {
	Name   string
	Path   string
	Method string
}

type contentParams struct {
	Plugins  map[string]map[string]*Plugin
	Services []*descriptor.Service
}

func applyTemplate(to tplOptions) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := headerTemplate.Execute(w, to); err != nil {
		return "", err
	}

	var targetServices = make([]*descriptor.Service, 0)
	for _, msg := range to.Messages {
		msgName := camel(*msg.Name)
		msg.Name = &msgName
	}

	for _, svc := range to.Services {
		svcName := camel(*svc.Name)
		svc.Name = &svcName
		targetServices = append(targetServices, svc)
	}

	tp := contentParams{
		Plugins:  to.Plugins,
		Services: targetServices,
	}

	if err := contentTemplate.Execute(w, tp); err != nil {
		return "", err
	}

	return w.String(), nil
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-engine. DO NOT EDIT.
// source: {{ .GetName }}
// versions:
// - protoc            {{ .ProtocVersion }}
// - protoc-gen-engine {{ .GeneratorVersion }}

package {{ .GoPkg.Name }}

import (
	{{range $i := .Imports}}{{if $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}

	{{range $i := .Imports}}{{if not $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}

	"github.com/lastbackend/engine/client/grpc"
)

// Suppress "imported and not used" errors
var _ context.Context
var _ logger.Logger
var _ plugin.Plugin
var _ server.Server
var _ client.Client
`))

	funcMap = template.FuncMap{
		"ToUpper":      strings.ToUpper,
		"ToLower":      strings.ToLower,
		"ToCapitalize": strings.Title,
	}

	_ = template.Must(contentTemplate.New("plugins-content").Parse(`
var props = map[string]map[string]engine.ServiceProps{
	{{range $type, $plugins := .Plugins}}
		"{{$type}}": {
			{{range $name, $plugin := $plugins}}
				"{{$name | ToCapitalize}}": engine.ServiceProps{
					Func: {{$plugin.Plugin}}.Register,
					Options: plugin.Option{
						Prefix: "{{$plugin.Prefix | ToLower}}",
					},
				},
			{{end}}
		},
	{{end}}
	}

	type layer struct {
		RPC *RPC
		{{range $type, $plugins := .Plugins}}{{$type}} *{{$type}}{{end}}
	}

	type RPC struct {
		Grpc grpc.Client
	}

	{{range $type, $plugins := .Plugins}}
		{{ $length := len $plugins }} {{ if eq $length 1 }}
			type {{$type}} struct {
				{{range $name, $plugin := $plugins}}{{$plugin.Pkg}}{{end}}
			}
		{{else}}
			type {{$type}} struct {
				{{range $name, $plugin := $plugins}}
					{{$name | ToCapitalize}} {{$plugin.Pkg}}{{end}}
			} 
		{{end}}
	{{end}}

`))

	_ = template.Must(contentTemplate.New("services-content").Parse(`
type Service interface {
	Logger() logger.Logger
	Meta() engine.Meta
	CLI() engine.CLI
	Run(i interface{}) error

	RPC() *RPC
	{{if .Plugins}}{{range $type, $plugins := .Plugins}}{{$type}}() *{{$type}}{{end}}{{end}}
}

func NewService(name string) Service {
	return &service{
		layer: &layer{
		RPC: &RPC{Grpc: new(struct{ grpc.Client })},
		{{if .Plugins}}{{range $type, $plugins := .Plugins}}{{$type}}: new({{$type}}),{{end}}{{end}}
		},
		base: engine.NewService(name,
	)}
}

type service struct {
	layer *layer
	base engine.Service
}

func (s *service) Meta() engine.Meta {
	return s.base.Meta()
}

func (s *service) CLI() engine.CLI {
	return s.base.CLI()
}

func (s *service) Run(i interface{}) error {
  if err := s.register(i); err != nil {
		return err
	}
	return s.base.Run()
}

func (s *service) Logger() logger.Logger {
	return s.base.Logger()
}

func (s *service) RPC() *RPC {
	return s.layer.RPC
}

{{if .Plugins}}
	{{range $type, $plugins := .Plugins}}
		func (s *service) {{$type}}() *{{$type}} {
			return s.layer.{{$type}}
		}
	{{end}}
{{end}}

func (s *service) register(i interface{}) error {

	// Register layers
	if err := s.base.Register(s.layer, props); err != nil {
		return err
	}

	// Register servers
	{{range $svc := .Services}}
		// Install grpc server
		type {{$svc.GetName}}GrpcRpcServer struct {
			{{$svc.GetName}}Server
		}
	
		h := &{{$svc.GetName | ToLower}}GrpcRpcServer{i.({{$svc.GetName}}RpcServer)}
		grpcServer := server.NewGrpcServer()
		if err := grpcServer.Register(&{{$svc.GetName}}_ServiceDesc, &{{$svc.GetName}}GrpcRpcServer{h}); err != nil {
			return err
		}
		
		if err := s.base.Server(grpcServer); err != nil {
			return err
		}
	{{end}}

	// Register clients
	{{range $svc := .Services}}
		// Install grpc client
		type {{$svc.GetName}}GrpcRpcClient struct {
			{{$svc.GetName}}Client
		}
	
		if err := s.base.Client(s.layer.RPC.Grpc, grpc.Register, client.Option{Prefix: "client-grpc"}); err != nil {
			return err
		}
		//s.layer.RPC.Auth = NewAuthRpcClient("auth", s.layer.RPC.Grpc)
	{{end}}

	return nil
}
`))

	_ = template.Must(contentTemplate.New("server-content").Parse(`
{{range $svc := .Services}}
	// Server API for Api service
	type {{$svc.GetName}}RpcServer interface {
		{{range $m := $svc.Methods}}
    {{if and (not $m.GetServerStreaming) (not $m.GetClientStreaming)}}
			{{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}) (*{{$m.ResponseType.GoName}}, error)
    {{else}}{{if not $m.GetClientStreaming}}
			{{$m.GetName}}(req *{{$m.RequestType.GoName}}, stream {{$svc.GetName}}_{{$m.GetName}}Server) error
    {{else}}
			{{$m.GetName}}(stream {{$svc.GetName}}_{{$m.GetName}}Server) error
    {{end}}{{end}}
	{{end}}
	}
{{end}}

{{range $svc := .Services}}
	type {{$svc.GetName | ToLower}}GrpcRpcServer struct {
		{{$svc.GetName}}RpcServer
	}

	{{range $m := $svc.Methods}}
    {{if and (not $m.GetServerStreaming) (not $m.GetClientStreaming)}}
  		func (h *{{$svc.GetName | ToLower}}GrpcRpcServer) {{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}) (*{{$m.ResponseType.GoName}}, error) {
				return h.{{$svc.GetName}}RpcServer.{{$m.GetName}}(ctx, in)				
			}
    {{else}}{{if not $m.GetClientStreaming}}
			func (h *{{$svc.GetName | ToLower}}GrpcRpcServer) {{$m.GetName}}(req *{{$m.RequestType.GoName}}, stream {{$svc.GetName}}_{{$m.GetName}}Server) error {
				return h.{{$svc.GetName}}RpcServer.{{$m.GetName}}(req, stream)
			}
    {{else}}
			func (h *{{$svc.GetName | ToLower}}GrpcRpcServer) {{$m.GetName}}(stream {{$svc.GetName}}_{{$m.GetName}}Server) error {
				return h.{{$svc.GetName}}RpcServer.{{$m.GetName}}(stream)
			}
    {{end}}{{end}}
	{{end}}
	func ({{$svc.GetName | ToLower}}GrpcRpcServer) mustEmbedUnimplemented{{$svc.GetName}}Server() {}
{{end}}
`))

	_ = template.Must(contentTemplate.New("client-content").Parse(`
{{range $svc := .Services}}
	// Client gRPC API for {{$svc.GetName}} service
	func New{{$svc.GetName}}RpcClient(service string, c grpc.Client) {{$svc.GetName}}RpcClient {
		return &{{$svc.GetName | ToLower}}GrpcRpcClient{service, c}
	}

	// Client gRPC API for {{$svc.GetName}} service
	type {{$svc.GetName}}RpcClient interface {
		{{range $m := $svc.Methods}}
    {{if and (not $m.GetClientStreaming) (not $m.GetClientStreaming)}}
			{{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}, opts ...grpc.CallOption) (*{{$m.ResponseType.GoName}}, error)
    {{else}}{{if not $m.GetClientStreaming}}
			{{$m.GetName}}(req *{{$m.RequestType.GoName}}, stream {{$svc.GetName}}_{{$m.GetName}}Client) error
    {{else}}
			{{$m.GetName}}(stream {{$svc.GetName}}_{{$m.GetName}}Client) error
    {{end}}{{end}}
	{{end}}
	}
{{end}}

{{range $svc := .Services}}
	type {{$svc.GetName | ToLower}}GrpcRpcClient struct {
		service string
		cli     grpc.Client
	}

	{{range $m := $svc.Methods}}
    {{if and (not $m.GetClientStreaming) (not $m.GetClientStreaming)}}
  		func (c *{{$svc.GetName | ToLower}}GrpcRpcClient) {{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}, opts ...grpc.CallOption) (*{{$m.ResponseType.GoName}}, error) {
				resp := new({{$m.ResponseType.GoName}})
				if err := c.cli.Call(ctx, c.service, {{$svc.GetName}}_{{$m.GetName}}Method, in, resp, opts...); err != nil {
					return nil, err
				}
				return resp, nil
			}
    {{else}}{{if not $m.GetClientStreaming}}
			func (c *{{$svc.GetName | ToLower}}GrpcRpcClient) {{$m.GetName}}(req *{{$m.RequestType.GoName}}, stream {{$svc.GetName}}_{{$m.GetName}}Client) error {
				return c.{{$svc.GetName}}RpcClient.{{$m.GetName}}(req, stream)
			}
    {{else}}
			func (c *{{$svc.GetName | ToLower}}GrpcRpcClient) {{$m.GetName}}(stream {{$svc.GetName}}_{{$m.GetName}}Client) error {
				return c.{{$svc.GetName}}RpcClient.{{$m.GetName}}(stream)
			}
    {{end}}{{end}}
	{{end}}
	func ({{$svc.GetName | ToLower}}GrpcRpcClient) mustEmbedUnimplemented{{$svc.GetName}}Client() {}
{{end}}

{{range $svc := .Services}}
	// Client methods for {{$svc.GetName}} service
	const (
		{{range $m := $svc.Methods}}
			{{$svc.GetName}}_{{$m.GetName}}Method = "/{{$svc.GetName | ToLower}}.{{$svc.GetName}}/{{$m.GetName}}"
		{{end}}
	)
{{end}}
`))

	contentTemplate = template.Must(template.New("content").Funcs(funcMap).Parse(`
{{template "plugins-content" .}}
{{template "services-content" .}}
{{template "server-content" .}}
{{template "client-content" .}}
`))
)
