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
	"bytes"
	"github.com/lastbackend/engine/protoc-gen-engine/descriptor"
	"strings"
	"text/template"
)

type tplOptions struct {
	*descriptor.File
	Imports          []descriptor.GoPackage
	Plugins          map[string]map[string]*Plugin
	ProtocVersion    string
	GeneratorVersion string
}

type Plugin struct {
	Prefix string
	Plugin string
}

type contentParams struct {
	Plugins  map[string]map[string]*Plugin
	Services []*descriptor.Service
}

func applyTemplate(to tplOptions) (string, error) {
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

	tp := contentParams{
		Plugins:  to.Plugins,
		Services: targetServices,
	}

	if err := contentTemplate.Execute(w, tp); err != nil {
		return "", err
	}

	return w.String(), nil
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
// Code generated by protoc-gen-engine. DO NOT EDIT.
// source: {{ .GetName }}
// versions:
// - protoc            {{ .ProtocVersion }}
// - protoc-gen-engine {{ .GeneratorVersion }}

package {{ .GoPkg.Name }}

import (
	{{range $i := .Imports}}{{if $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}

	{{range $i := .Imports}}{{if not $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}
)

// Suppress "imported and not used" errors
var _ context.Context
var _ logger.Logger
var _ plugin.Plugin
var _ server.Server
`))

	funcMap = template.FuncMap{
		"ToUpper":      strings.ToUpper,
		"ToLower":      strings.ToLower,
		"ToCapitalize": strings.Title,
	}

	contentTemplate = template.Must(template.New("content").Funcs(funcMap).Parse(`
var props = map[string]map[string]engine.ServiceProps{
{{range $type, $plugins := .Plugins}}
	"{{$type}}": {
		{{range $name, $plugin := $plugins}}
			"{{$name | ToCapitalize}}": engine.ServiceProps{
				Func: {{$plugin.Plugin}}.Register,
				Options: plugin.Option{
					Prefix: "{{$plugin.Prefix | ToLower}}",
				},
			},
		{{end}}
	},
{{end}}
}

type Service interface {
	Register(interface{}) error
	Logger() logger.Logger
	Meta() engine.Meta
	Init() error
	Run() error
}

func NewService(name string) Service {
	return &service{engine.NewService(name)}
}

type service struct {
	base engine.Service
}

func (s *service) Register(i interface{}) error {

	if err := s.base.Register(i, props); err != nil {
		return err
	}

{{range $svc := .Services}}
	// Install grpc server
	type {{$svc.GetName}}GrpcHandler struct {
		{{$svc.GetName}}Server
	}

  h := &{{$svc.GetName | ToLower}}GrpcHandler{i.({{$svc.GetName}}Handler)}
  g := server.NewServer("grpc")
  if err := g.Register(&{{$svc.GetName}}_ServiceDesc, &{{$svc.GetName}}GrpcHandler{h}); err != nil {
  	return err
  }
  
  if err := s.base.Transport(g); err != nil {
  	return err
  }
{{end}}

	return nil
}

func (s *service) Meta() engine.Meta {
	return s.base.Meta()
}

func (s *service) Run() error {
	return s.base.Run()
}

func (s *service) Init() error {
	return s.base.Init()
}

func (s *service) Logger() logger.Logger {
	return s.base.Logger()
}

{{range $svc := .Services}}
	// Server API for Api service
	type {{$svc.GetName}}Handler interface {
	{{range $m := $svc.Methods}}
		{{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}) (*{{$m.ResponseType.GoName}}, error)
	{{end}}
	}
{{end}}

{{range $svc := .Services}}
	type {{$svc.GetName | ToLower}}GrpcHandler struct {
		{{$svc.GetName}}Handler
	}

	{{range $m := $svc.Methods}}
		func (h *{{$svc.GetName | ToLower}}GrpcHandler) {{$m.GetName}}(ctx context.Context, in *{{$m.RequestType.GoName}}) (*{{$m.ResponseType.GoName}}, error) {
			return  h.{{$svc.GetName}}Handler.{{$m.GetName}}(ctx, in)
		}
	{{end}}
	func ({{$svc.GetName | ToLower}}GrpcHandler) mustEmbedUnimplemented{{$svc.GetName}}Server() {}
{{end}}


{{range $svc := .Services}}
// Client methods for {{$svc.GetName}} service
const (
{{range $m := $svc.Methods}}
	{{$m.GetName}}Method = "/{{$svc.GetName | ToLower}}.{{$svc.GetName}}/{{$m.GetName}}"
{{end}}
)
{{end}}

`))
)