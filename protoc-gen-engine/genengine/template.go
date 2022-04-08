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
}

type RPC struct {
	Grpc grpc.RPCClient
	{{- range $key, $value := .Clients }}
		{{ $value.Service | ToCapitalize }} {{ $key }}.{{ $value.Service | ToCapitalize }}RPCClient
	{{ end }}
}

func NewService(name string) Service {
	return &service{
		engine: engine.NewService(name),
		{{- if not .HasNotServer }}
		srv:     make([]interface{}, 0),
		{{- end }}
		svc:     make([]interface{}, 0),
		ctrl:    make([]interface{}, 0),
		rpc:     new(RPC),
	}
}

type service struct {
	engine engine.Service
	rpc    *RPC
	{{- if not .HasNotServer }}
	srv  []interface{}
	{{- end }}
	svc  []interface{}
	ctrl []interface{}
	cfg  interface{}
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
	{{ range $name, $plugin := $plugins }}
type {{ $plugin.Prefix | ToCapitalize }}{{ $type | ToCapitalize }} interface {
	{{ $plugin.Plugin }}.Plugin
}
	{{ end }}
{{ end }}

func (s *service) Run(ctx context.Context) error {
	
	{{ range $type, $plugins := .Plugins }}
		{{- range $name, $plugin := $plugins }}
			{{ $type | ToLower }}{{ $plugin.Prefix | ToCapitalize }} := {{ $plugin.Plugin }}.NewPlugin(s.engine, &{{ $plugin.Plugin }}.Options{Name: "{{ $plugin.Prefix | ToLower }}"})
		{{- end }}
	{{- end }}

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
		func() *RPC {
			return s.rpc
		},
		{{- range $type, $plugins := .Plugins }}
			{{- range $name, $plugin := $plugins }}
				fx.Annotate(
					func() {{ $plugin.Prefix | ToCapitalize }}{{ $type | ToCapitalize }} {
						return {{ $type | ToLower }}{{ $plugin.Prefix | ToCapitalize }}
					},
				),
			{{- end }}
		{{- end }}
	)

	provide = append(provide, s.svc...)
	{{- if not .HasNotServer }}
	provide = append(provide, s.srv...)
	{{- end }}

	app := fx.New(
		fx.Options(
			fx.Supply(s.cfg),
			fx.Provide(provide...),
			fx.Invoke(s.registerClients),
			{{- if not $.HasNotServer }}
				{{- range $svc := .Services }}			
					fx.Invoke(s.register{{ $svc.GetName }}Server),
				{{- end }}
			{{- end }}
			fx.Invoke(s.ctrl...),
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


func (s *service) registerClients() error {

	// Register clients
	
	s.rpc.Grpc = grpc.NewClient(s.engine, &grpc.ClientOptions{Name: "client-grpc"})

	if err := s.engine.ClientRegister(s.rpc.Grpc); err != nil {
		return err
	}

	{{ range $key, $value := .Clients }}
		s.rpc.{{ $value.Service | ToCapitalize }} = {{ $value.Service | ToLower }}.New{{ $value.Service | ToCapitalize }}RPCClient("{{ $value.Service | ToLower }}", s.rpc.Grpc.Client())
	{{ end }}

	return nil
}

{{ if not .HasNotServer }}
	{{ range $svc := .Services }}
	func (s *service) register{{ $svc.GetName }}Server(srv {{ $svc.GetName }}RpcServer) error {
	
		// Register servers
	
		type {{ $svc.GetName }}GrpcRpcServer struct {
				{{ $svc.GetName }}Server
		}
	
		h := &{{ $svc.GetName | ToLower }}GrpcRpcServer{srv.({{ $svc.GetName }}RpcServer)}
		grpcServer := server.NewServer(s.engine, &server.ServerOptions{Name: "server-{{ $svc.GetName | ToLower }}-grpc"})
		if err := grpcServer.Register(&{{ $svc.GetName }}_ServiceDesc, &{{ $svc.GetName }}GrpcRpcServer{h}); err != nil {
			return err
		}
	
		if err := s.engine.ServerRegister(grpcServer); err != nil {
			return err
		}
	
		return nil
	}
	{{ end }}
{{ end }}

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
			{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}) (*{{ $m.ResponseType.GoName }}, error)
    {{ else }}{{ if not $m.GetClientStreaming }}
			{{ $m.GetName }}(req *{{ $m.RequestType.GoName }}, stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error
    {{ else }}
			{{ $m.GetName }}(stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error
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
  		func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}) (*{{ $m.ResponseType.GoName }}, error) {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(ctx, req)
			}
    {{ else }}{{ if not $m.GetClientStreaming }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(req *{{ $m.RequestType.GoName }}, stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(req, stream)
			}
    {{ else }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error {
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
	func New{{ $svc.GetName }}RPCClient(service string, c grpc.Client) {{ $svc.GetName }}RPCClient {
		return &{{ $svc.GetName | ToLower }}GrpcRPCClient{service, c}
	}

	// Client gRPC API for {{ $svc.GetName }} service
	type {{ $svc.GetName }}RPCClient interface {
		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) (*{{ $m.ResponseType.GoName }}, error)
			{{ else }}
				{{ if not $m.GetClientStreaming }}
					{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
				{{ else }}
					{{ $m.GetName }}(ctx context.Context, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
				{{ end }}
			{{ end }}
		{{ end }}
	}
{{ end }}

{{ range $svc := .Services }}
	type {{ $svc.GetName | ToLower }}GrpcRPCClient struct {
		service string
		cli     grpc.Client
	}

	{{ range $m := $svc.Methods }}
		{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
			func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) (*{{ $m.ResponseType.GoName }}, error) {
				resp := new({{ $m.ResponseType.GoName }})
				if err := c.cli.Call(ctx, c.service, {{ $svc.GetName }}_{{ $m.GetName }}Method, req, resp, opts...); err != nil {
					return nil, err
				}
				return resp, nil
			}
		{{ else }}
			{{ if not $m.GetClientStreaming }}
				func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
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
				func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context,  opts ...grpc.CallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
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
				Recv() (*{{ $m.ResponseType.GoName }}, error)
				{{ if $m.GetClientStreaming }}Send(*{{ $m.RequestType.GoName }}) error{{ end }}
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

			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Recv() (*{{ $m.ResponseType.GoName }}, error) {
				m := new({{ $m.ResponseType.GoName }})
				err := x.stream.RecvMsg(m)
				if err != nil {
					return nil, err
				}
				return m, nil
			}

			{{ if $m.GetClientStreaming }}
			func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Send(m *{{ $m.RequestType.GoName }}) error {
				return x.stream.SendMsg(m)
			}
			{{ end }}

		{{ end }}
	{{ end }}

	func ({{ $svc.GetName | ToLower }}GrpcRPCClient) mustEmbedUnimplemented{{ $svc.GetName }}Client() {}
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

	func With{{ $svc.GetName }}Stubs(stubs *{{ $svc.GetName }}Stubs) servicepb.{{ $svc.GetName }}RPCClient{

		rpc_mock := new(service_mocks.{{ $svc.GetName }}RPCClient)

		{{ range $m := $svc.Methods }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				for _, st := range stubs.{{ $m.GetName }} {
					resp := st.Response
					err := st.Error
					rpc_mock.On("{{ $m.GetName }}", st.Context, st.Request).Return(
						func(ctx context.Context, req *servicepb.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) *servicepb.{{ $m.ResponseType.GoName }} {
							return resp
						},
						func(ctx context.Context, req *servicepb.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) error {
							return err
						},
					)
					rpc_mock.On("{{ $m.GetName }}", st.Context, st.Request, st.CallOptions).Return(
						func(ctx context.Context, req *servicepb.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) *servicepb.{{ $m.ResponseType.GoName }} {
							return resp
						},
						func(ctx context.Context, req *servicepb.{{ $m.RequestType.GoName }}, opts ...grpc.CallOption) error {
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
				Request     *servicepb.{{ $m.RequestType.GoName }}
				Response    *servicepb.{{ $m.ResponseType.GoName }}
				CallOptions []grpc.CallOption				
				Error 	    error
			{{ end }}
			}
		{{ end }}
	{{ end }}
`))
)
