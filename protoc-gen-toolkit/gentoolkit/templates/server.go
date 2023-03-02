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

// ServerGRPCDefineTpl is the server GRPC define template used for new services.
var ServerGRPCDefineTpl = `
{{ range $svc := .Services }}
{{ if $svc.Methods }}
	// Define GRPC services for {{ $svc.GetName }} GRPC server
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
{{ end }}

{{ range $svc := .Services }}
{{ if $svc.Methods }}
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
{{ end }}
`

// ServerHTTPDefineTpl is the server HTTP define template used for new services.
var ServerHTTPDefineTpl = `
func (s *service) registerRouter() {
	{{ range $svc := .Services }}
		{{ range $m := $svc.Methods }}
			{{ range $binding := $m.Bindings }}
				
				{{ if $binding.Websocket }}
				s.toolkit.Router().Handle(http.MethodGet, "{{ $binding.HttpPath }}", s.Router().ServerWS,
					router.HandleOptions{Middlewares: middlewares.getMiddleware("{{ $binding.RpcMethod }}")})
				{{ end }}
			
				{{ if or (not $binding.Websocket) ($binding.WebsocketProxy) }}
				s.toolkit.Router().Subscribe("{{ $binding.RpcMethod }}", func(ctx context.Context, event ws.Event, c *ws.Client) error {
					ctx, cancel := context.WithCancel(ctx)
					defer cancel()
			
					var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
					var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}
			
					if err := json.Unmarshal(event.Payload, &protoRequest); err != nil {
						return err
					}
			
					callOpts := make([]grpc.CallOption, 0)
		
					if headers := ctx.Value(ws.RequestHeaders); headers != nil {
						if v, ok := headers.(map[string]string); ok {
							callOpts = append(callOpts, grpc.Headers(v))
						}
					}
			
					if err := s.toolkit.Client().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
						return err	
					}
			
					return c.WriteJSON(protoResponse)
				})
		
				{{ if not $binding.WebsocketProxy }}
					s.toolkit.Router().Handle({{ $binding.HttpMethod }}, "{{ $binding.HttpPath }}", func(w http.ResponseWriter, r *http.Request) {
						ctx, cancel := context.WithCancel(r.Context())
						defer cancel()
			
						var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
						var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}
				
						{{ if or (eq $binding.HttpMethod "http.MethodPost") (eq $binding.HttpMethod "http.MethodPut") (eq $binding.HttpMethod "http.MethodPatch") }}
							{{ if eq $binding.RawBody "*" }}
								im, om := router.GetMarshaler(s.toolkit.Router(), r)
			
								reader, err := router.NewReader(r.Body)
								if err != nil {
									errors.HTTP.InternalServerError(w)
									return
								}
								
								if err := im.NewDecoder(reader).Decode(&protoRequest); err != nil && err != io.EOF {
									errors.HTTP.InternalServerError(w)
									return
								}
							{{ else }}
								_, om := router.GetMarshaler(s.toolkit.Router(), r)
			
								if err := router.SetRawBodyToProto(r, &protoRequest, "{{ $binding.RawBody }}"); err != nil {
									errors.HTTP.InternalServerError(w)
									return
								}
			
							{{ end }}
						{{ else }}
							_, om := router.GetMarshaler(s.toolkit.Router(), r)
			
							if err := r.ParseForm(); err != nil {
								errors.HTTP.InternalServerError(w)
								return
							}
				
							if err := router.ParseRequestQueryParametersToProto(&protoRequest, r.Form); err != nil {
								errors.HTTP.InternalServerError(w)
								return
							}
						{{ end }}
			
						{{ range $param := $binding.HttpParams }}
						if err := router.ParseRequestUrlParametersToProto(r, &protoRequest, "{{ $param | ToTrimRegexFromQueryParameter }}"); err != nil {
							errors.HTTP.InternalServerError(w)
							return
						}
						{{ end }}
				
						headers, err := router.PrepareHeaderFromRequest(r)
						if err != nil {
							errors.HTTP.InternalServerError(w)
							return
						}
			
						callOpts := make([]grpc.CallOption, 0)
						callOpts = append(callOpts, grpc.Headers(headers))
			
						if err := s.toolkit.Client().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
							errors.GrpcErrorHandlerFunc(w, err)
							return			
						}
				
						buf, err := om.Marshal(protoResponse)
						if err != nil {
							errors.HTTP.InternalServerError(w)
							return
						}
			
						w.Header().Set("Content-Type", om.ContentType())
			
						w.WriteHeader(http.StatusOK)
						if _, err = w.Write(buf); err != nil {
							s.toolkit.Logger().Infof("Failed to write response: %v", err)
							return
						}
			
					}, router.HandleOptions{Middlewares: middlewares.getMiddleware("{{ $binding.RpcMethod }}")})
					{{ end }}
				{{ end }} 
			{{ end }} 
		{{ end }}
	{{ end }}
}
`
