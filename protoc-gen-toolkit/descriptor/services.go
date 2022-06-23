/*
Copyright [2014] - [2022] The Last.Backend authors.

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
	"google.golang.org/protobuf/types/descriptorpb"

	"fmt"
	"strings"
)

func (d *Descriptor) loadServices(file *File) error {
	var services []*Service
	for _, service := range file.GetService() {
		svc := &Service{
			File:                   file,
			ServiceDescriptorProto: service,
		}
		for _, md := range service.GetMethod() {
			meth, err := d.newMethod(svc, md)
			if err != nil {
				return err
			}
			svc.Methods = append(svc.Methods, meth)
		}
		services = append(services, svc)
	}
	file.Services = services
	return nil
}

func (d *Descriptor) newMethod(svc *Service, md *descriptorpb.MethodDescriptorProto) (*Method, error) {
	requestType, err := d.FindMessage(svc.File.GetPackage(), md.GetInputType())
	if err != nil {
		return nil, err
	}
	responseType, err := d.FindMessage(svc.File.GetPackage(), md.GetOutputType())
	if err != nil {
		return nil, err
	}
	meth := &Method{
		Service:               svc,
		MethodDescriptorProto: md,
		RequestType:           requestType,
		ResponseType:          responseType,
	}
	return meth, nil
}

func (d *Descriptor) FindMessage(location, name string) (*Message, error) {
	if strings.HasPrefix(name, ".") {
		m, ok := d.messageMap[name]
		if !ok {
			return nil, fmt.Errorf("no message found: %s", name)
		}
		return m, nil
	}

	if !strings.HasPrefix(location, ".") {
		location = fmt.Sprintf(".%s", location)
	}
	components := strings.Split(location, ".")
	for len(components) > 0 {
		messageName := strings.Join(append(components, name), ".")
		if m, ok := d.messageMap[messageName]; ok {
			return m, nil
		}
		components = components[:len(components)-1]
	}
	return nil, fmt.Errorf("no message found: %s", name)
}
