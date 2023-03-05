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

// ClientTpl is the client template used for new services.
var ClientTpl = `// Suppress "imported and not used" errors
var _ context.Context
var _ emptypb.Empty

{{ range $svc := .Services }}
	// Client gRPC API for {{ $svc.GetName }} service
	func New{{ $svc.GetName }}RPCClient(service string, c client.GRPCClient) {{ $svc.GetName }}RPCClient {
		return &{{ $svc.GetName | ToLower }}GrpcRPCClient{service, c}
	}

	// Client gRPC API for {{ $svc.GetName }} service
	type {{ $svc.GetName }}RPCClient interface {
		{{ range $m := $svc.Methods }}
			{{ if not $m.IsWebsocket }}
				{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
					{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType "" }}, opts ...client.GRPCCallOption) (*{{ $m.ResponseType.GoType "" }}, error)
				{{ else }}
					{{ if not $m.GetClientStreaming }}
						{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType "" }}, opts ...client.GRPCCallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
					{{ else }}
						{{ $m.GetName }}(ctx context.Context, opts ...client.GRPCCallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error)
					{{ end }}
				{{ end }}
			{{ end }}
		{{ end }}
	}
{{ end }}

{{ range $svc := .Services }}
	type {{ $svc.GetName | ToLower }}GrpcRPCClient struct {
		service string
		cli     client.GRPCClient
	}

	{{ range $m := $svc.Methods }}
		{{ if not $m.IsWebsocket }}
			{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
				func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType "" }}, opts ...client.GRPCCallOption) (*{{ $m.ResponseType.GoType "" }}, error) {
					resp := new({{ $m.ResponseType.GoType "" }})
					if err := c.cli.Call(ctx, c.service, {{ $svc.GetName }}_{{ $m.GetName }}Method, req, resp, opts...); err != nil {
						return nil, err
					}
					return resp, nil
				}
			{{ else }}
				{{ if not $m.GetClientStreaming }}
					func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType "" }}, opts ...client.GRPCCallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
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
					func (c *{{ $svc.GetName | ToLower }}GrpcRPCClient) {{ $m.GetName }}(ctx context.Context,  opts ...client.GRPCCallOption) ({{ $svc.GetName }}_{{ $m.GetName }}Service, error) {
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
					Recv() (*{{ $m.ResponseType.GoType "" }}, error)
					{{ if $m.GetClientStreaming }}Send(*{{ $m.RequestType.GoType "" }}) error{{ end }}
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
	
				func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Recv() (*{{ $m.ResponseType.GoType "" }}, error) {
					m := new({{ $m.ResponseType.GoType "" }})
					err := x.stream.RecvMsg(m)
					if err != nil {
						return nil, err
					}
					return m, nil
				}
	
				{{ if $m.GetClientStreaming }}
				func (x *{{ $svc.GetName | ToLower }}{{ $m.GetName }}Service) Send(m *{{ $m.RequestType.GoType "" }}) error {
					return x.stream.SendMsg(m)
				}
				{{ end }}
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
			{{ $svc.GetName }}_{{ $m.GetName }}Method = "/{{ $svc.FullyName }}/{{ $m.GetName }}"
		{{ end }}
		)
	{{ end }}
{{ end }}
`
