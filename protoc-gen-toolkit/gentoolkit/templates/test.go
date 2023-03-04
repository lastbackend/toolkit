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

// TestTpl is the test template used for new services.
var TestTpl = `// Suppress "imported and not used" errors
var _ context.Context
var _ emptypb.Empty

{{ range $svc := .Services }}
	// Mock server API for Api service
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
						func(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, opts ...grpc.CallOption) *{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }} {
							return resp
						},
						func(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, opts ...grpc.CallOption) error {
							return err
						},
					)
					rpc_mock.On("{{ $m.GetName }}", st.Context, st.Request, st.CallOptions).Return(
						func(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, opts ...grpc.CallOption) *{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }} {
							return resp
						},
						func(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, opts ...grpc.CallOption) error {
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
				Request     *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}
				Response    *{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }}
				CallOptions []grpc.CallOption				
				Error 	    error
			{{ end }}
			}
		{{ end }}
	{{ end }}
`
