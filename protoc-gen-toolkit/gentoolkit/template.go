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

package gentoolkit

import (
	"bytes"
	"text/template"

	"github.com/lastbackend/toolkit/pkg/util/converter"
	"github.com/lastbackend/toolkit/pkg/util/strings"
	"github.com/lastbackend/toolkit/protoc-gen-toolkit/descriptor"
	"github.com/lastbackend/toolkit/protoc-gen-toolkit/gentoolkit/templates"
)

var (
	funcMap = template.FuncMap{
		"ToUpper":                       strings.ToUpper,
		"ToLower":                       strings.ToLower,
		"ToCapitalize":                  strings.Title,
		"ToCamel":                       strings.ToCamel,
		"ToTrimRegexFromQueryParameter": converter.ToTrimRegexFromQueryParameter,
		"dehyphen":                      strings.DeHyphenFunc,
		"lowerhyphen":                   strings.LowerHyphenFunc,
		"tohyphen":                      strings.ToHyphen,
		"inc": func(n int) int {
			return n + 1
		},
	}
)

var (
	messageTemplate  = template.Must(template.New("message").Parse(templates.MessageTpl))
	headerTemplate   = template.Must(template.New("header").Parse(templates.HeaderTpl))
	clientTemplate   = template.Must(template.New("client-content").Funcs(funcMap).Parse(templates.ClientTpl))
	testStubTemplate = template.Must(template.New("stub-content-mockery").Parse(templates.TestTpl))

	serviceTemplate = template.Must(template.New("content").Funcs(funcMap).Parse(templates.ContentTpl))
	_               = template.Must(serviceTemplate.New("service-grpc-define").Parse(templates.ServerGRPCDefineTpl))
	_               = template.Must(serviceTemplate.New("service-http-define").Parse(templates.ServerHTTPDefineTpl))
	_               = template.Must(serviceTemplate.New("plugin-define").Parse(templates.PluginDefineTpl))
	_               = template.Must(serviceTemplate.New("plugin-init").Parse(templates.PluginInitializeTpl))
	_               = template.Must(serviceTemplate.New("plugin-register").Parse(templates.PluginRegisterTpl))
)

type Client struct {
	Service string
	Pkg     string
}

type contentServiceParams struct {
	Plugins  map[string][]*descriptor.Plugin
	Services []*descriptor.Service
	Clients  map[string]*Client
}

type tplServiceOptions struct {
	*descriptor.File
	Imports []descriptor.GoPackage
	Plugins map[string][]*descriptor.Plugin
	Clients map[string]*Client
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
		Plugins:  to.Plugins,
		Clients:  to.Clients,
		Services: targetServices,
	}

	if err := serviceTemplate.Execute(w, tp); err != nil {
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

	if err := clientTemplate.Execute(w, tp); err != nil {
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

	if err := testStubTemplate.Execute(w, to); err != nil {
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

	if err := messageTemplate.Execute(w, to); err != nil {
		return "", err
	}

	return w.String(), nil
}
