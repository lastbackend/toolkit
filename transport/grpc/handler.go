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

package grpc

import (
	"fmt"
	"reflect"
)

type HandlerOption func(*HandlerOptions)

type Handler struct {
	name    string
	handler interface{}
	//endpoints []*registry.Endpoint
	//opts      server.HandlerOptions
}

type HandlerOptions struct {
	Internal bool
	Metadata map[string]map[string]string
}

func newHandler(h interface{}, opts ...HandlerOption) *Handler {
	options := HandlerOptions{
		Metadata: make(map[string]map[string]string),
	}

	for _, o := range opts {
		o(&options)
	}

	typ := reflect.TypeOf(h)
	hdlr := reflect.ValueOf(h)
	name := reflect.Indirect(hdlr).Type().Name()

	fmt.Println(">>>>>>>>>>>>>>>>>>>>")
	fmt.Println(typ, hdlr, name)
	fmt.Println(">>>>>>>>>>>>>>>>>>>>")

	//var endpoints []*registry.Endpoint
	//
	//for m := 0; m < typ.NumMethod(); m++ {
	//	if e := extractEndpoint(typ.Method(m)); e != nil {
	//		e.Name = name + "." + e.Name
	//
	//		for k, v := range options.Metadata[e.Name] {
	//			e.Metadata[k] = v
	//		}
	//
	//		endpoints = append(endpoints, e)
	//	}
	//}

	return &Handler{
		//name:      name,
		//handler:   handler,
		//endpoints: endpoints,
		//opts:      options,
	}
}

func (r *Handler) Name() string {
	return r.name
}

func (r *Handler) Handler() interface{} {
	return r.handler
}

//func extractValue(v reflect.Type, d int) *registry.Value {
//	if d == 6 {
//		return nil
//	}
//	if v == nil {
//		return nil
//	}
//
//	if v.Kind() == reflect.Ptr {
//		v = v.Elem()
//	}
//
//	// slices and maps don't have a defined name
//	if v.Kind() != reflect.Slice && v.Kind() != reflect.Map && len(v.Name()) == 0 {
//		return nil
//	}
//
//	arg := &registry.Value{
//		Name: v.Name(),
//		Type: v.Name(),
//	}
//
//	switch v.Kind() {
//	case reflect.Struct:
//		for i := 0; i < v.NumField(); i++ {
//			f := v.Field(i)
//			val := extractValue(f.Type, d+1)
//			if val == nil {
//				continue
//			}
//
//			// if we can find a json tag use it
//			if tags := f.Tag.Get("json"); len(tags) > 0 {
//				parts := strings.Split(tags, ",")
//				if parts[0] == "-" || parts[0] == "omitempty" {
//					continue
//				}
//				val.Name = parts[0]
//			} else {
//				continue
//			}
//
//			// if there's no name default it
//			if len(val.Name) == 0 {
//				continue
//			}
//
//			arg.Values = append(arg.Values, val)
//		}
//	case reflect.Slice:
//		p := v.Elem()
//		if p.Kind() == reflect.Ptr {
//			p = p.Elem()
//		}
//		arg.Type = "[]" + p.Name()
//	case reflect.Map:
//		p := v.Elem()
//		if p.Kind() == reflect.Ptr {
//			p = p.Elem()
//		}
//		key := v.Key()
//		if key.Kind() == reflect.Ptr {
//			key = key.Elem()
//		}
//
//		arg.Type = fmt.Sprintf("map[%s]%s", key.Name(), p.Name())
//
//	}
//
//	return arg
//}
//
//func extractEndpoint(method reflect.Method) *registry.Endpoint {
//	if method.PkgPath != "" {
//		return nil
//	}
//
//	var rspType, reqType reflect.Type
//	var stream bool
//	mt := method.Type
//
//	switch mt.NumIn() {
//	case 3:
//		reqType = mt.In(1)
//		rspType = mt.In(2)
//	case 4:
//		reqType = mt.In(2)
//		rspType = mt.In(3)
//	default:
//		return nil
//	}
//
//	// are we dealing with a stream?
//	switch rspType.Kind() {
//	case reflect.Func, reflect.Interface:
//		stream = true
//	}
//
//	request := extractValue(reqType, 0)
//	response := extractValue(rspType, 0)
//
//	ep := &registry.Endpoint{
//		Name:     method.Name,
//		Request:  request,
//		Response: response,
//		Metadata: make(map[string]string),
//	}
//
//	if stream {
//		ep.Metadata = map[string]string{
//			"stream": fmt.Sprintf("%v", stream),
//		}
//	}
//
//	return ep
//}
//
//func extractSubValue(typ reflect.Type) *registry.Value {
//	var reqType reflect.Type
//	switch typ.NumIn() {
//	case 1:
//		reqType = typ.In(0)
//	case 2:
//		reqType = typ.In(1)
//	case 3:
//		reqType = typ.In(2)
//	default:
//		return nil
//	}
//	return extractValue(reqType, 0)
//}
