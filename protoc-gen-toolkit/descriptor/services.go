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

package descriptor

import (
	"fmt"
	"net/http"
	"strings"

	toolkit_annotattions "github.com/lastbackend/toolkit/protoc-gen-toolkit/toolkit/options"
	options "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func (d *Descriptor) loadServices(file *File) error {
	var services []*Service
	for _, service := range file.GetService() {
		svc := &Service{
			File:                   file,
			ServiceDescriptorProto: service,
		}

		for _, md := range service.GetMethod() {
			method, err := d.newMethod(svc, md)
			if err != nil {
				return err
			}
			svc.Methods = append(svc.Methods, method)
		}

		services = append(services, svc)
	}

	file.Services = services

	return nil
}

func (d *Descriptor) newMethod(svc *Service, md *descriptorpb.MethodDescriptorProto) (*Method, error) {
	requestType, err := d.findMessage(svc.File.GetPackage(), md.GetInputType())
	if err != nil {
		return nil, err
	}
	responseType, err := d.findMessage(svc.File.GetPackage(), md.GetOutputType())
	if err != nil {
		return nil, err
	}

	method := &Method{
		Service:               svc,
		MethodDescriptorProto: md,
		RequestType:           requestType,
		ResponseType:          responseType,
	}

	if method.Options != nil && proto.HasExtension(method.Options, toolkit_annotattions.E_Server) {
		err = setBindingsToMethod(method)
		if err != nil {
			return nil, err
		}
	}

	return method, nil
}

func (d *Descriptor) findMessage(location, name string) (*Message, error) {
	if strings.HasPrefix(name, ".") {
		method, ok := d.messageMap[name]
		if !ok {
			return nil, fmt.Errorf("no message found: %s", name)
		}
		return method, nil
	}

	if !strings.HasPrefix(location, ".") {
		location = fmt.Sprintf(".%s", location)
	}

	parts := strings.Split(location, ".")
	for len(parts) > 0 {
		messageName := strings.Join(append(parts, name), ".")
		if method, ok := d.messageMap[messageName]; ok {
			return method, nil
		}
		parts = parts[:len(parts)-1]
	}

	return nil, fmt.Errorf("no message found: %s", name)
}

func setBindingsToMethod(method *Method) error {
	pOpts, err := getProxyOptions(method)
	if err != nil {
		return err
	}

	if pOpts != nil {
		switch true {
		case pOpts.GetWebsocket():
			opts, err := getHTTPOptions(method)
			if err != nil {
				return err
			}

			method.IsWebsocket = true

			binding := &Binding{
				Method:       method,
				Index:        len(method.Bindings),
				RpcMethod:    method.GetName(),
				HttpMethod:   http.MethodGet,
				HttpPath:     opts.GetGet(),
				HttpParams:   getVariablesFromPath(opts.GetGet()),
				RequestType:  method.RequestType,
				ResponseType: method.ResponseType,
				Websocket:    true,
			}
			method.Bindings = append(method.Bindings, binding)

		case pOpts.GetWebsocketProxy() != nil:
			rOpts := pOpts.GetWebsocketProxy()

			method.IsWebsocket = true

			binding := &Binding{
				Method:         method,
				Index:          len(method.Bindings),
				Service:        rOpts.GetService(),
				RpcPath:        rOpts.GetMethod(),
				RpcMethod:      method.GetName(),
				HttpMethod:     http.MethodGet,
				RequestType:    method.RequestType,
				ResponseType:   method.ResponseType,
				WebsocketProxy: true,
			}

			method.Bindings = append(method.Bindings, binding)

		case pOpts.GetHttpProxy() != nil && proto.HasExtension(method.Options, options.E_Http):
			rOpts := pOpts.GetHttpProxy()

			opts, err := getHTTPOptions(method)
			if err != nil {
				return err
			}

			if opts != nil {
				var (
					httpMethod string
					httpPath   string
				)
				switch {
				case opts.GetGet() != "":
					httpMethod = "http.MethodGet"
					httpPath = opts.GetGet()
				case opts.GetPut() != "":
					httpMethod = "http.MethodPut"
					httpPath = opts.GetPut()
				case opts.GetPost() != "":
					httpMethod = "http.MethodPost"
					httpPath = opts.GetPost()
				case opts.GetDelete() != "":
					httpMethod = "http.MethodDelete"
					httpPath = opts.GetDelete()
				case opts.GetPatch() != "":
					httpMethod = "http.MethodPatch"
					httpPath = opts.GetPatch()
				default:
					return fmt.Errorf("unknown http method: %v", opts.GetGet())
				}

				binding := &Binding{
					Method:       method,
					Index:        len(method.Bindings),
					Service:      rOpts.GetService(),
					RpcPath:      rOpts.GetMethod(),
					RpcMethod:    method.GetName(),
					HttpMethod:   httpMethod,
					HttpPath:     httpPath,
					HttpParams:   getVariablesFromPath(httpPath),
					RequestType:  method.RequestType,
					ResponseType: method.ResponseType,
					Stream:       method.GetClientStreaming(),
				}

				binding, err := newBinding(method, opts, rOpts)
				if err != nil {
					return err
				}

				method.Bindings = append(method.Bindings, binding)

				for _, additional := range opts.GetAdditionalBindings() {
					if len(additional.AdditionalBindings) > 0 {
						continue
					}
					b, err := newBinding(method, additional, rOpts)
					if err != nil {
						continue
					}
					method.Bindings = append(method.Bindings, b)
				}
			}
		}

	}

	return nil
}

func getHTTPOptions(method *Method) (*options.HttpRule, error) {
	if method.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(method.Options, options.E_Http) {
		return nil, nil
	}
	ext := proto.GetExtension(method.Options, options.E_Http)
	opts, ok := ext.(*options.HttpRule)
	if !ok {
		return nil, fmt.Errorf("extension is not an HttpRule")
	}
	return opts, nil
}

func getProxyOptions(m *Method) (*toolkit_annotattions.Server, error) {
	if m.Options == nil {
		return nil, nil
	}
	if !proto.HasExtension(m.Options, toolkit_annotattions.E_Server) {
		return nil, nil
	}
	ext := proto.GetExtension(m.Options, toolkit_annotattions.E_Server)
	opts, ok := ext.(*toolkit_annotattions.Server)
	if !ok {
		return nil, fmt.Errorf("extension is not an Proxy")
	}
	return opts, nil
}

func newBinding(method *Method, opts *options.HttpRule, rOpts *toolkit_annotattions.HttpProxy) (*Binding, error) {
	var (
		httpMethod string
		httpPath   string
	)
	switch {
	case opts.GetGet() != "":
		httpMethod = "http.MethodGet"
		httpPath = opts.GetGet()
	case opts.GetPut() != "":
		httpMethod = "http.MethodPut"
		httpPath = opts.GetPut()
		if opts.Body == "" {
			opts.Body = "*"
		}
	case opts.GetPost() != "":
		httpMethod = "http.MethodPost"
		httpPath = opts.GetPost()
		if opts.Body == "" {
			opts.Body = "*"
		}
	case opts.GetDelete() != "":
		httpMethod = "http.MethodDelete"
		httpPath = opts.GetDelete()
	case opts.GetPatch() != "":
		httpMethod = "http.MethodPatch"
		httpPath = opts.GetPatch()
		if opts.Body == "" {
			opts.Body = "*"
		}
	default:
		return nil, fmt.Errorf("not fount method")
	}

	return &Binding{
		Method:       method,
		Index:        len(method.Bindings),
		Service:      rOpts.GetService(),
		RpcPath:      rOpts.GetMethod(),
		RpcMethod:    method.GetName(),
		HttpMethod:   httpMethod,
		HttpPath:     httpPath,
		HttpParams:   getVariablesFromPath(httpPath),
		RequestType:  method.RequestType,
		ResponseType: method.ResponseType,
		Stream:       method.GetClientStreaming(),
		RawBody:      opts.Body,
	}, nil
}

func getVariablesFromPath(path string) (variables []string) {
	if path == "" {
		return make([]string, 0)
	}

	for path != "" {
		firstIndex := -1
		lastIndex := -1

		firstIndex = strings.IndexAny(path, "{")
		lastIndex = strings.IndexAny(path, "}")

		if firstIndex > -1 && lastIndex > -1 {
			field := path[firstIndex+1 : lastIndex]
			if len(strings.TrimSpace(field)) > 0 {
				variables = append(variables, field)
			}
			path = path[lastIndex+1:]
		}

		if firstIndex == -1 || lastIndex == -1 {
			path = path[1:]
		}
	}

	return variables
}
