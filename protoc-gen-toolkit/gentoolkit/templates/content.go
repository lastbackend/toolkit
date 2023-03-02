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

{{- if .Plugins }}
// Plugins define
{{- range $type, $plugins := .Plugins }}
{{- range $index, $plugin := $plugins }}
type {{ $plugin.Prefix | ToCamel }}Plugin interface {
	{{ $type }}.Plugin
}
{{ end }}
{{ end }}
{{ end }}

{{- if not .HasNotServer }}
	{{ template "server-content" . }}
{{ end }}

func NewService(name string, opts ...runtime.Option) (_ toolkit.Service, err error) {
	app := new(service)

	app.runtime, err = controller.NewRuntime(context.Background(), name, opts...)
	if err != nil {
		return nil, err
	}

{{ if .Plugins }}{{ $count := 0 }}
	// loop over plugins and initialize plugin instance
	{{- range $type, $plugins := .Plugins }}
	{{- range $index, $plugin := $plugins }}
	plugin_{{ $count }} := postgres_gorm.NewPlugin(app.runtime, &postgres_gorm.Options{Name: "{{ $plugin.Prefix | ToLower }}"})
	{{- $count = inc $count }}
	{{- end }}
	{{- end }}
{{ end }}

{{ if .Plugins }}{{ $count := 0 }}
	// loop over plugins and register plugin in toolkit
	{{- range $type, $plugins := .Plugins }}
	{{- range $index, $plugin := $plugins }}
	app.runtime.Plugin().Provide(func() {{ $plugin.Prefix | ToCamel }}Plugin { return plugin_{{ $count }} })
	{{- $count = inc $count }}
	{{- end }}
	{{- end }}
{{ end }}

{{- if not .HasNotServer }}
{{ template "server-register-content" . }}
{{ end }}

	return app.runtime.SVC(), nil
}
`
