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

package templates

// ServerGRPCTpl is the server template used for new services.
var ServerGRPCTpl = `
{{ range $svc := .Services }}
	// Server API for Api service
	type {{ $svc.GetName }}RpcServer interface {
	{{ range $m := $svc.Methods }}
		{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
			{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }}, error)
		{{ else }}{{ if not $m.GetClientStreaming }}
			{{ $m.GetName }}(req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error
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
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }}, error) {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(ctx, req)
			}
		{{ else }}{{ if not $m.GetClientStreaming }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(req, stream)
			}
		{{ else }}
			func (h *{{ $svc.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(stream {{ $svc.GetName }}_{{ $m.GetName }}Server) error {
				return h.{{ $svc.GetName }}RpcServer.{{ $m.GetName }}(stream)
			}
		{{ end }}{{ end }}
	{{ end }}
	func ({{ $svc.GetName | ToLower }}GrpcRpcServer) mustEmbedUnimplemented{{ $svc.GetName }}Server() {}

	func register{{ $svc.GetName }}GRPCServer(runtime runtime.Runtime, srv {{ $svc.GetName }}RpcServer) error {
		runtime.Server().GRPC().RegisterService(&{{ $svc.GetName | ToLower }}GrpcRpcServer{srv})
		return nil
	}
{{ end }}
`

var ServerGRPCRegisterTpl = `
{{ range $svc := .Services }}
	// set descriptor to {{ $svc.GetName }} grpc server
	app.runtime.Server().GRPCNew(name, nil)

	app.runtime.Server().GRPC().SetDescriptor({{ $svc.GetName }}_ServiceDesc)
	app.runtime.Server().GRPC().SetConstructor(register{{ $svc.GetName }}GRPCServer)
{{ end }}
`
