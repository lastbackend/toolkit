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

// ContentTpl is the content template used for new services.
var ContentTpl = `
// This is a compile-time assertion to ensure that this generated file
// is compatible with the toolkit package it is being compiled against and
// suppress "imported and not used" errors
var (
	_ context.Context
	_ logger.Logger
	_ emptypb.Empty
	_ grpc.Client
	_ http.Handler
	_ errors.Err
	_ io.Reader
	_ json.Marshaler
	//_ ws.Client
)

// Definitions
type service struct {
	runtime runtime.Runtime
}

// Plugins define
{{- if .Plugins }}
{{- template "plugin-define" .Plugins }}
{{- end }}
{{- if .Plugins }}
{{ range $svc := .Services }}
{{- template "plugin-define" $svc.Plugins }}
{{- end }}
{{- end }}

{{ range $svc := .Services }}
// Service {{ $svc.GetName }} define
func New{{ $svc.GetName }}Service(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
	app := new(service)

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

{{ if $svc.Methods }}
	// set descriptor to {{ $svc.GetName }} GRPC server
	app.runtime.Server().GRPCNew(name)
	app.runtime.Server().GRPC().SetDescriptor({{ $svc.GetName }}_ServiceDesc)
	app.runtime.Server().GRPC().SetConstructor(register{{ $svc.GetName }}GRPCServer)
{{ end }}

{{ if $svc.Methods }}
	// create new {{ $svc.GetName }} HTTP server
	app.runtime.Server().HTTPNew(name)
	app.runtime.Server().HTTP().AddMiddleware("middleware1", {{ $svc.GetName | ToLower }}HTTPServerMiddleware)
	app.runtime.Server().HTTP().AddHandler(http.MethodPost, "/hello", {{ $svc.GetName | ToLower }}HTTPServerSubscribeHandler, tk_http.WithMiddleware("middleware1"))
{{ end }}

	return app.runtime.Service(), nil
}
{{ end }}

{{ template "grpc-service-define" . }}

{{ template "http-handler-define" . }}
`
