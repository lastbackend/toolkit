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

type Plugin struct {
	Prefix string
	Plugin string
	Pkg    string
}

type Client struct {
	Service string
	Pkg     string
}

type contentServiceParams struct {
	HasNotServer bool
	Plugins      map[string]map[string]*Plugin
	Services     []*descriptor.Service
	Clients      map[string]*Client
}

type tplServiceOptions struct {
	*descriptor.File
	HasNotServer bool
	Imports      []descriptor.GoPackage
	Plugins      map[string]map[string]*Plugin
	Clients      map[string]*Client
}

func applyServiceTemplate(to tplServiceOptions) (string, error) {
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

	tp := contentServiceParams{
		HasNotServer: to.HasNotServer,
		Plugins:      to.Plugins,
		Clients:      to.Clients,
		Services:     targetServices,
	}

	if err := contentServiceTemplate.Execute(w, tp); err != nil {
		return "", err
	}

	return w.String(), nil
}

type tplClientOptions struct {
	*descriptor.File
	Imports []descriptor.GoPackage
	Clients map[string]*Client
}

type contentClientParams struct {
	HasNotServiceGenerate bool
	Clients               map[string]*Client
	Services              []*descriptor.Service
}

func applyClientTemplate(to tplClientOptions) (string, error) {
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

	tp := contentClientParams{
		Clients:  to.Clients,
		Services: targetServices,
	}

	if err := contentClientTemplate.Execute(w, tp); err != nil {
		return "", err
	}

	return w.String(), nil
}

type tplMockeryTestOptions struct {
	*descriptor.File
	Imports []descriptor.GoPackage
}

func applyTestTemplate(to tplMockeryTestOptions) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := headerTemplate.Execute(w, to); err != nil {
		return "", err
	}

	if err := contentTestStubTemplate.Execute(w, to); err != nil {
		return "", err
	}

	return w.String(), nil
}

type tplMessageOptions struct {
	*descriptor.File
	Message string
}

func applyTemplateWithMessage(to tplMessageOptions) (string, error) {
	w := bytes.NewBuffer(nil)

	if err := templateWithMessage.Execute(w, to); err != nil {
		return "", err
	}

	return w.String(), nil
}

var (
	templateWithMessage = template.Must(template.New("message").Parse(`
// Code generated by protoc-gen-engine. DO NOT EDIT.
// source: {{ .GetName }}

package {{ .GoPkg.Name }}

// {{ .Message }}

`))

	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-engine. DO NOT EDIT.
// source: {{ .GetName }}

package {{ .GoPkg.Name }}

import (
	{{ range $i := .Imports }}{{ if $i.Standard }}{{ $i | printf "%s\n" }}{{ end }}{{ end }}

	{{ range $i := .Imports }}{{ if not $i.Standard }}{{ $i | printf "%s\n" }}{{ end }}{{ end }}
)

`))

	funcMap = template.FuncMap{
		"ToUpper":      strings.ToUpper,
		"ToLower":      strings.ToLower,
		"ToCapitalize": strings.Title,
	}

	_ = template.Must(contentServiceTemplate.New("services-content").Parse(`
type Service interface {
	Logger() logger.Logger
	Meta() engine.Meta
	CLI() engine.CLI
	Run(ctx context.Context) error

	SetConfig(cfg interface{})
	{{- if not .HasNotServer }}
	SetServer(srv interface{})
	{{ end }}

	AddPackage(pkg interface{})
	AddController(ctrl interface{})

	{{ range $type, $plugins := .Plugins }}
		{{- range $name, $plugin := $plugins }} 
			Set{{ $name | ToCapitalize }}({{ $plugin.Prefix | ToLower }} interface{})
		{{ end }}
	{{ end }}
}

{{ range $svc := .Services }}
type {{ $svc.GetName }}RPC struct {
	{{ if not $.HasNotServer }}
	Grpc grpc.RpcClient
	{{ end }}
	{{ range $key, $value := $.Clients -}}
		{{ $value.Service | ToCapitalize }} {{ $key }}.{{ $value.Service | ToCapitalize }}RpcClient
	{{ end }}
}
{{ end }}

func NewService(name string) Service {
	return &service{
		engine: engine.NewService(name),
		{{ if not .HasNotServer }}
		srv:     make([]interface{}, 0),
		{{ end }}
		svc:     make([]interface{}, 0),
		ctrl:    make([]interface{}, 0),
		{{ range $svc := .Services -}}
		rpc{{ $svc.GetName }}:    new({{ $svc.GetName }}RPC),
		{{ end }}
		{{- range $type, $plugins := .Plugins }}
			{{- range $name, $plugin := $plugins }} 
				{{ $plugin.Prefix | ToLower }}: {{ $plugin.Plugin }}.NewPlugin,
			{{ end }}
		{{ end }}
	}
}

type service struct {
	engine engine.Service

	{{- range $svc := .Services }}
	rpc{{ $svc.GetName }} *{{ $svc.GetName }}RPC
	{{ end }}
	
	{{- if not .HasNotServer }}
	srv  []interface{}
	{{ end }}
	svc  []interface{}
	ctrl []interface{}
	cfg  interface{}

	{{- range $type, $plugins := .Plugins }}
		{{- range $name, $plugin := $plugins }} 
			{{ $plugin.Prefix | ToLower }} interface{}
		{{ end }}
	{{ end }}
}

func (s *service) Meta() engine.Meta {
	return s.engine.Meta()
}

func (s *service) CLI() engine.CLI {
	return s.engine.CLI()
}

func (s *service) Logger() logger.Logger {
	return s.engine.Logger()
}

func (s *service) SetConfig(cfg interface{}) {
	s.cfg = cfg
}

{{ if not .HasNotServer }}
func (s *service) SetServer(srv interface{}) {
	s.srv = append(s.srv, srv)
}
{{ end }}

func (s *service) AddPackage(pkg interface{}) {
	s.svc = append(s.svc, pkg)
}

func (s *service) AddController(ctrl interface{}) {
	s.ctrl = append(s.ctrl, ctrl)
}

{{ range $type, $plugins := .Plugins }}
	{{- range $name, $plugin := $plugins }} 
		func (s *service) Set{{ $name | ToCapitalize }}({{ $plugin.Prefix | ToLower }} interface{}) {
			s.{{ $plugin.Prefix | ToLower }} = {{ $plugin.Prefix | ToLower }}
		}
	{{ end }}
{{ end }}

func (s *service) Run(ctx context.Context) error {
	provide := make([]interface{}, 0)
	provide = append(provide,
		fx.Annotate(
			func() engine.Service {
				return s.engine
			},
		),
		func() Service {
			return s
		},
		{{- range $svc := .Services }}
		func() *{{ $svc.GetName }}RPC {
			return s.rpc{{ $svc.GetName }}
		},
		{{ end }}
		{{- range $type, $plugins := .Plugins }}
			{{- range $name, $plugin := $plugins }}
				fx.Annotate(
					func() {{ $plugin.Plugin }}.Plugin {
						return {{ $plugin.Plugin }}.NewPlugin(s.engine, &{{ $plugin.Plugin }}.Options{Name: "{{ $plugin.Prefix | ToLower }}"})
					},
				),
			{{ end }}
		{{ end }}
	)

	{{ if not .HasNotServer }}
	provide = append(provide, s.srv...)
	{{ end }}
	provide = append(provide, s.svc...)
	provide = append(provide, s.ctrl...)
	{{- range $type, $plugins := .Plugins }}
		{{- range $name, $plugin := $plugins }}
			provide = append(provide, s.{{ $plugin.Prefix | ToLower }})
		{{ end }}
	{{ end }}

	app := fx.New(
		fx.Options(
			fx.Supply(s.cfg),
			fx.Provide(provide...),

			{{ if not $.HasNotServer }}
				{{ range $svc := .Services }}			
					fx.Invoke(s.register{{ $svc.GetName }}Client),
					fx.Invoke(s.register{{ $svc.GetName }}Server),
				{{ end -}}
			{{ end }}
			fx.Invoke(s.registerController),
			fx.Invoke(s.runService),
		),
		fx.NopLogger,
	)

	if err := app.Start(ctx); err != nil {
		return err
	}
	
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, shutdownSignals...)

	select {
	// wait on kill signal
	case <-signalCh:
	// wait on context cancel
	case <-ctx.Done():
	}

	return app.Stop(ctx)
}

{{ if not .HasNotServer }}
	{{ range $svc := .Services }}
	func (s *service) register{{ $svc.GetName }}Client() error {
	
		// Register clients
		
		s.rpc{{ $svc.GetName }}.Grpc = grpc.NewClient(s.engine, &grpc.ClientOptions{Name: "client-{{ $svc.GetName | ToLower }}-grpc"})
	
		if err := s.engine.ClientRegister(s.rpc{{ $svc.GetName }}.Grpc); err != nil {
			return err
		}
	
		{{ range $key, $value := $.Clients }}
			s.rpc{{ $svc.GetName }}.{{ $value.Service | ToCapitalize }} = {{ $value.Service | ToLower }}.New{{ $value.Service | ToCapitalize }}RpcClient("{{ $value.Service | ToLower }}", s.rpc{{ $svc.GetName }}.Grpc.Client())
		{{ end }}
	
		return nil
	}
	
	func (s *service) register{{ $svc.GetName }}Server(srv {{ $svc.GetName }}RpcServer) error {
	
		// Register servers
	
		type {{ $svc.GetName }}GrpcRpcServer struct {
				proto.{{ $svc.GetName }}Server
		}
	
		h := &{{ $svc.GetName | ToLower }}GrpcRpcServer{srv.({{ $svc.GetName }}RpcServer)}
		grpcServer := server.NewServer(s.engine, &server.ServerOptions{Name: "server-{{ $svc.GetName | ToLower }}-grpc"})
		if err := grpcServer.Register(&proto.{{ $svc.GetName }}_ServiceDesc, &{{ $svc.GetName }}GrpcRpcServer{h}); err != nil {
			return err
		}
	
		if err := s.engine.ServerRegister(grpcServer); err != nil {
			return err
		}
	
		return nil
	}
	{{ end }}
{{ end }}

func (s *service) registerController(ctrl engine.Controller) error {
	// Register controllers
	if err := s.engine.ControllerRegister(ctrl); err != nil {
		return err
	}
	return nil
}

func (s *service) runService(lc fx.Lifecycle) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return s.engine.Start()
		},
		OnStop: func(ctx context.Context) error {
			return s.engine.Stop()
		},
	})
	return nil
}

var shutdownSignals = []os.Signal{
	syscall.SIGTERM,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGKILL,
}
`))

	_ = template.Must(contentServiceTemplate.New("server-content").Parse(`
{{ range $svc := .Services }}
	// Server API for Api service
	type {{ $svc.GetName }}RpcServer interface {
		{{ range $m := $svc.Methods }}
    {{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
			{{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}) (*proto.{{ $m.ResponseType.GoName }}, error)
    {{ else }}{{ if not $m.GetClientStreaming }}
			{{ $m.GetName }}(req *proto.{{ $m.RequestType.GoName }}, stream proto.{{ $svc.GetName }}_{{ $m.GetName }}Server) error
    {{ else }}
			{{ $m.GetName }}(stream proto.{{ $svc.GetName }}_{{ $m.GetName }}Server) error
    {{ end }}{{ end }}
	{{ end }}
	}
{{ end }}

{{ range $svc := .Services }}
	type {{ $svc.GetName | ToLower }}GrpcRpcServer struct {
		{{ $svc.GetName }}RpcServer
	}

	{{ range $m := $svc.Methods }}
    {{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
  		func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}) (*proto.{{ $m.ResponseType.GoName }}, error) {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(ctx, req)
			}
    {{ else }}{{ if not $m.GetClientStreaming }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(req *proto.{{ $m.RequestType.GoName }}, stream proto.{{ $svc.GetName }}_{{ $m.GetName }}Server) error {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(req, stream)
			}
    {{ else }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(stream proto.{{ $svc.GetName }}_{{ $m.GetName }}Server) error {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(stream)
			}
    {{ end }}{{ end }}
	{{ end }}
	func ({{ $svc.GetName | ToLower }}GrpcRpcServer) mustEmbedUnimplemented{{ $svc.GetName }}Server() {}
{{ end }}
`))

	contentServiceTemplate = template.Must(template.New("content").Funcs(funcMap).Parse(`
// Suppress "imported and not used" errors
var _ context.Context
var _ logger.Logger
{{- if not .HasNotServer }}
var _ server.Server
{{ end }}

{{ template "services-content" . }}
{{ if not .HasNotServer }}
	{{ template "server-content" . }}
{{ end }}
`))

	contentClientTemplate = template.Must(template.New("client-content").Funcs(funcMap).Parse(`
// Suppress "imported and not used" errors
var _ context.Context

{{ range $svc := .Services }}
	// Client gRPC API for {{ $svc.GetName }} service
	func New{{ $svc.GetName }}RpcClient(service string, c grpc.Client) {{ $svc.GetName }}RpcClient {
		return &{{ $svc.GetName | ToLower }}GrpcRpcClient{service, c}
	}

	// Client gRPC API for {{ $svc.GetName }} service
	type {{ $svc.GetName }}RpcClient interface {
		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				{{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) (*proto.{{ $m.ResponseType.GoName }}, error)
			{{ else }}
				{{ if not $m.GetClientStreaming }}
					{{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
				{{ else }}
					{{ $m.GetName }}(ctx context.Context, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
				{{ end }}
			{{ end }}
		{{ end }}
	}
{{ end }}

{{ range $svc := .Services }}
	type {{ $svc.GetName | ToLower }}GrpcRpcClient struct {
		service string
		cli     grpc.Client
	}

	{{ range $m := $svc.Methods }}
		{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
			func (c *{{ $svc.GetName | ToLower }}GrpcRpcClient) {{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) (*proto.{{ $m.ResponseType.GoName }}, error) {
				resp := new(proto.{{ $m.ResponseType.GoName }})
				if err := c.cli.Call(ctx, c.service, {{ $svc.GetName }}_{{ $m.GetName }}Method, req, resp, opts...); err != nil {
					return nil, err
				}
				return resp, nil
			}
		{{ else }}
			{{ if not $m.GetClientStreaming }}
				func (c *{{ $svc.GetName | ToLower }}GrpcRpcClient) {{ $m.GetName }}(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
					stream, err := c.cli.Stream(ctx, c.service, {{ $svc.GetName }}_{{ $m.GetName }}Method, req, opts...)
					if err != nil {
						return nil, err
					}
					if err := stream.SendMsg(req); err != nil {
						return nil, err
					}
					return &{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service{stream}, nil
				}
			{{ else }}
				func (c *{{ $svc.GetName | ToLower }}GrpcRpcClient) {{ $m.GetName }}(ctx context.Context,  opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
					stream, err := c.cli.Stream(ctx, c.service, {{ $svc.GetName }}_{{ $m.GetName }}Method, nil, opts...)
					if err != nil {
						return nil, err
					}
					return &{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service{stream}, nil
				}
			{{ end }}

			type {{ $svc.GetName }}_{{ $m.GetName }}Service interface {
				SendMsg(interface{}) error
				RecvMsg(interface{}) error
				Close() error
				Recv() (*proto.{{ $m.ResponseType.GoName }}, error)
				{{ if $m.GetClientStreaming }}Send(*proto.{{ $m.RequestType.GoName }}) error{{ end }}
			}

			type {{ $svc.GetName | ToLower }}{{ $m.GetName }}Service struct {
				stream grpc.Stream
			}

			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Close() error {
				return x.stream.CloseSend()
			}

			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) SendMsg(m interface{}) error {
				return x.stream.SendMsg(m)
			}

			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) RecvMsg(m interface{}) error {
				return x.stream.RecvMsg(m)
			}

			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Recv() (*proto.{{ $m.ResponseType.GoName }}, error) {
				m := new(proto.{{ $m.ResponseType.GoName }})
				err := x.stream.RecvMsg(m)
				if err != nil {
					return nil, err
				}
				return m, nil
			}

			{{ if $m.GetClientStreaming }}
			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Send(m *proto.{{ $m.RequestType.GoName }}) error {
				return x.stream.SendMsg(m)
			}
			{{ end }}

		{{ end }}
	{{ end }}

	func ({{ $svc.GetName | ToLower }}GrpcRpcClient) mustEmbedUnimplemented{{ $svc.GetName }}Client() {}
{{ end }}

{{ range $svc := .Services }}
	{{ if $svc.Methods }}
		// Client methods for {{ $svc.GetName }} service
		const (
			{{ range $m := $svc.Methods }}
				{{ $svc.GetName }}_{{ $m.GetName }}Method = "/{{ $svc.GetName | ToLower }}.{{ $svc.GetName }}/{{ $m.GetName }}"
			{{ end }}
		)
	{{ end }}
{{ end }}
`))

	contentTestStubTemplate = template.Must(template.New("stub-content-mockery").Parse(`
// Suppress "imported and not used" errors
var _ context.Context

{{ range $svc := .Services }}
	// Server API for Api service
	type {{ $svc.GetName }}Stubs struct {
		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				{{ $m.GetName }} []{{ $m.GetName }}Stub
			{{ else }}
		{{ end }}
	{{ end }}
	}

	func New{{ $svc.GetName }}Stubs() *{{ $svc.GetName }}Stubs {
		stubs:= new({{ $svc.GetName }}Stubs)
		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				stubs.{{ $m.GetName }} = make([]{{ $m.GetName }}Stub,0)
			{{ end }}
		{{ end }}
		return stubs
	}

	func With{{ $svc.GetName }}Stubs(stubs *{{ $svc.GetName }}Stubs) client.{{ $svc.GetName }}RpcClient{

		rpc_mock := new(service_mocks.{{ $svc.GetName }}RpcClient)

		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				for _, st := range stubs.{{ $m.GetName }} {
					resp := st.Response
					err := st.Error
					rpc_mock.On("{{ $m.GetName }}", st.Context, st.Request).Return(
						func(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) *proto.{{ $m.ResponseType.GoName }} {
							return resp
						},
						func(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) error {
							return err
						},
					)
					rpc_mock.On("{{ $m.GetName }}", st.Context, st.Request, st.CallOptions).Return(
						func(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) *proto.{{ $m.ResponseType.GoName }} {
							return resp
						},
						func(ctx context.Context, req *proto.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) error {
							return err
						},
					)
				}
			{{ end }}
		{{ end }}

		return rpc_mock
	}
{{ end }}

	{{ range $svc := .Services }}
		{{ range $m := $svc.Methods }}
			type {{ $m.GetName }}Stub struct {
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				Context     context.Context
				Request     *proto.{{ $m.RequestType.GoName }}
				Response    *proto.{{ $m.ResponseType.GoName }}
				CallOptions []grpc.CallOption				
				Error 	    error
			{{ end }}
			}
		{{ end }}
	{{ end }}
`))
)
