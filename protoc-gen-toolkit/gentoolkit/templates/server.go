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
var ServerGRPCDefineTpl = `// Define GRPC services for {{ .GetName }} GRPC server

type {{ .GetName }}RpcServer interface {
{{ range $m := .Methods }}
{{ if not $m.IsWebsocket }}
	{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
		{{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }}, error)
	{{ else }}{{ if not $m.GetClientStreaming }}
		{{ $m.GetName }}(req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, stream {{ $.GetName }}_{{ $m.GetName }}Server) error
	{{ else }}
		{{ $m.GetName }}(stream {{ $.GetName }}_{{ $m.GetName }}Server) error
	{{ end }}{{ end }}
{{ end }}
{{ end }}
}

type {{ .GetName | ToLower }}GrpcRpcServer struct {
	{{ .GetName }}RpcServer
}

{{ range $m := .Methods }}
{{ if not $m.IsWebsocket }}
	{{ if and (not $m.GetServerStreaming) (not $m.GetClientStreaming) }}
		func (h *{{ $.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(ctx context.Context, req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}) (*{{ $m.ResponseType.GoType $m.Service.File.GoPkg.Path }}, error) {
			return h.{{ $.GetName }}RpcServer.{{ $m.GetName }}(ctx, req)
		}
	{{ else }}{{ if not $m.GetClientStreaming }}
		func (h *{{ $.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(req *{{ $m.RequestType.GoType $m.Service.File.GoPkg.Path }}, stream {{ $.GetName }}_{{ $m.GetName }}Server) error {
			return h.{{ $.GetName }}RpcServer.{{ $m.GetName }}(req, stream)
		}
	{{ else }}
		func (h *{{ $.GetName | ToLower }}GrpcRpcServer) {{ $m.GetName }}(stream {{ $.GetName }}_{{ $m.GetName }}Server) error {
			return h.{{ $.GetName }}RpcServer.{{ $m.GetName }}(stream)
		}
	{{ end }}{{ end }}
{{ end }}
{{ end }}
func ({{ $.GetName | ToLower }}GrpcRpcServer) mustEmbedUnimplemented{{ $.GetName }}Server() {}

func register{{ $.GetName }}GRPCServer(runtime runtime.Runtime, srv {{ $.GetName }}RpcServer) error {
	runtime.Server().GRPC().RegisterService(&{{ $.GetName | ToLower }}GrpcRpcServer{srv})
	return nil
}
`

// ServerHTTPDefineTpl is the server HTTP define template used for new services.
var ServerHTTPDefineTpl = `// Define HTTP handlers for Router HTTP server
{{- range $m := .Methods }}
{{- range $binding := $m.Bindings }}
{{- if and $.UseWebsocketProxyServer $binding.WebsocketProxy (not $binding.Websocket) }}
func (s *service{{ $.GetName | ToCamel }}) handlerWSProxy{{ $.GetName | ToCamel }}{{ $m.GetName | ToCamel }}(ctx context.Context, event tk_ws.Event, c *tk_ws.Client) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
	var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}

	if err := json.Unmarshal(event.Payload, &protoRequest); err != nil {
		return err
	}

	callOpts := make([]client.GRPCCallOption, 0)

	if headers := ctx.Value(tk_ws.RequestHeaders); headers != nil {
		if v, ok := headers.(map[string]string); ok {
			callOpts = append(callOpts, client.GRPCOptionHeaders(v))
		}
	}

	if err := s.runtime.Client().GRPC().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
		return err	
	}

	return c.WriteJSON(protoResponse)
}
{{- end }}
{{ if and $.UseHTTPProxyServer (not $binding.WebsocketProxy) (not $binding.Websocket) }}
func (s *service{{ $.GetName | ToCamel }}) handlerHTTP{{ $.GetName | ToCamel }}{{ $m.GetName | ToCamel }}{{- if $binding.AdditionalBinding }}_{{ $binding.Index }}{{ end }}(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	var protoRequest {{ $binding.RequestType.GoType $binding.Method.Service.File.GoPkg.Path }}
	var protoResponse {{ $binding.ResponseType.GoType $binding.Method.Service.File.GoPkg.Path }}

	{{ if or (eq $binding.HttpMethod "http.MethodPost") (eq $binding.HttpMethod "http.MethodPut") (eq $binding.HttpMethod "http.MethodPatch") }}
		{{ if eq $binding.RawBody "*" }}
			im, om := tk_http.GetMarshaler(s.runtime.Server().HTTP(), r)

			reader, err := tk_http.NewReader(r.Body)
			if err != nil {
				errors.HTTP.InternalServerError(w)
				return
			}
			
			if err := im.NewDecoder(reader).Decode(&protoRequest); err != nil && err != io.EOF {
				errors.HTTP.InternalServerError(w)
				return
			}
		{{ else }}
			_, om := tk_http.GetMarshaler(s.runtime.Server().HTTP(), r)

			if err := tk_http.SetRawBodyToProto(r, &protoRequest, "{{ $binding.RawBody }}"); err != nil {
				errors.HTTP.InternalServerError(w)
				return
			}

		{{ end }}
	{{ else }}
		_, om := tk_http.GetMarshaler(s.runtime.Server().HTTP(), r)

		if err := r.ParseForm(); err != nil {
			errors.HTTP.InternalServerError(w)
			return
		}

		if err := tk_http.ParseRequestQueryParametersToProto(&protoRequest, r.Form); err != nil {
			errors.HTTP.InternalServerError(w)
			return
		}
	{{ end }}

	{{ range $param := $binding.HttpParams }}
	if err := tk_http.ParseRequestUrlParametersToProto(r, &protoRequest, "{{ $param | ToTrimRegexFromQueryParameter }}"); err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}
	{{ end }}

	headers, err := tk_http.PrepareHeaderFromRequest(r)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}

	callOpts := make([]client.GRPCCallOption, 0)
	callOpts = append(callOpts, client.GRPCOptionHeaders(headers))

	if err := s.runtime.Client().GRPC().Call(ctx, "{{ $binding.Service }}", "{{ $binding.RpcPath }}", &protoRequest, &protoResponse, callOpts...); err != nil {
		errors.GrpcErrorHandlerFunc(w, err)
		return			
	}

	buf, err := om.Marshal(protoResponse)
	if err != nil {
		errors.HTTP.InternalServerError(w)
		return
	}
	
  w.Header().Set("Content-Type", om.ContentType())
	if proceed, err := tk_http.HandleGRPCResponse(w, r, headers); err != nil || !proceed {
		return
	}
	
	if _, err = w.Write(buf); err != nil {
		return
	}
}
{{- end }} 
{{- end }} 
{{- end }} 
`
