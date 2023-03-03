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

// ServiceContentTpl is the content template used for new services.
var ServiceContentTpl = `
// This is a compile-time assertion to ensure that this generated file
// is compatible with the toolkit package it is being compiled against and
// suppress "imported and not used" errors
var (
	_ context.Context
	_ emptypb.Empty
	_ grpc.Client
	_ http.Handler
	_ errors.Err
	_ io.Reader
	_ json.Marshaler
	_ tk_ws.Client
	_ tk_http.Handler
)

// Definitions
{{ if .Plugins }}
{{- template "plugin-define" .Plugins }}
{{- end }}
{{ range $svc := .Services }}
{{- template "plugin-define" $svc.Plugins }}
{{- end }}

{{ range $svc := .Services }}
// Service {{ $svc.GetName }} define
type service{{ $svc.GetName | ToCamel }} struct {
	runtime runtime.Runtime
}

func New{{ $svc.GetName }}Service(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
	app := new(service{{ $svc.GetName | ToCamel }})

	app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
	if err != nil {
		return nil, err
	}

{{ if or .Plugins $svc.Plugins }}
	// loop over plugins and initialize plugin instance
	{{- template "plugin-init" $.Plugins }}
	{{- template "plugin-init" $svc.Plugins }}
{{ end }}

{{ if or .Plugins $svc.Plugins }}
	// loop over plugins and register plugin in toolkit
	{{- template "plugin-register" $.Plugins }}
	{{- template "plugin-register" $svc.Plugins }}
{{ end }}

{{ if and $svc.UseGRPCServer $svc.Methods }}
	// set descriptor to {{ $svc.GetName }} GRPC server
	app.runtime.Server().GRPCNew(name)
	app.runtime.Server().GRPC().SetDescriptor({{ $svc.GetName }}_ServiceDesc)
	app.runtime.Server().GRPC().SetConstructor(register{{ $svc.GetName }}GRPCServer)
{{ end }}

{{ if and (or $svc.UseHTTPProxyServer $svc.UseWebsocketProxyServer $svc.UseWebsocketServer) $svc.Methods }}
	// create new {{ $svc.GetName }} HTTP server
	app.runtime.Server().HTTPNew(name)
	{{ if and (or $svc.UseHTTPProxyServer $svc.UseWebsocketProxyServer) $svc.HTTPMiddlewares }}app.runtime.Server().HTTP().UseMiddleware({{ range $index, $mdw := $svc.HTTPMiddlewares }}{{ if lt 0 $index }}, {{ end }}"{{ $mdw }}"{{ end }}){{ end }}
	{{- range $m := $svc.Methods }}
	{{- range $binding := $m.Bindings }}
		{{- if and $binding.Websocket $svc.UseWebsocketServer }}
			app.runtime.Server().HTTP().AddHandler(http.MethodGet, "{{ $binding.HttpPath }}", app.runtime.Server().HTTP().ServerWS)
		{{- end }} 
		{{- if and $svc.UseWebsocketProxyServer $binding.WebsocketProxy (not $binding.Websocket) }}
		app.runtime.Server().HTTP().Subscribe("{{ $binding.RpcMethod }}", app.handlerWSProxy{{ $svc.GetName | ToCamel }}{{ $m.GetName | ToCamel }})
		{{- end }}
		{{- if and $svc.UseHTTPProxyServer (not $binding.WebsocketProxy) (not $binding.Websocket) }}
		app.runtime.Server().HTTP().AddHandler({{ $binding.HttpMethod }}, "{{ $binding.HttpPath }}", app.handlerHTTP{{ $svc.GetName | ToCamel }}{{ $m.GetName | ToCamel }}{{ if $binding.Middlewares }}, 
			{{ range $index, $mdw := $binding.Middlewares }}{{ if lt 0 $index }}, {{ end }}tk_http.WithMiddleware("{{ $mdw }}"){{ end }}{{ end }})
		{{- end }}
	{{- end }} 
	{{- end }} 
{{ end }}

	return app.runtime.Service(), nil
}

{{ if and $svc.UseGRPCServer .Methods }}
{{- template "grpc-service-define" . }}
{{ end }}

{{ if and (or $svc.UseHTTPProxyServer $svc.UseWebsocketProxyServer $svc.UseWebsocketServer) .Methods }}
{{- template "http-handler-define" . }}
{{- end }}

{{ end }}
`
