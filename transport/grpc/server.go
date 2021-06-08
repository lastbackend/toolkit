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
	"github.com/lastbackend/engine/logger"
	"github.com/lastbackend/engine/service/server"
	"github.com/pkg/errors"

	"context"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"
)

var (
	typeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

type methodType struct {
	method      reflect.Method
	ArgType     reflect.Type
	ReplyType   reflect.Type
	ContextType reflect.Type
	stream      bool
}

type service struct {
	name            string
	receiverMethods reflect.Value
	typeReceiver    reflect.Type
	method          map[string]*methodType
}

type rpcServer struct {
	mu         sync.Mutex // protects the serviceMap
	serviceMap map[string]*service
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return isExported(t.Name()) || t.PkgPath() == ""
}

func prepareEndpoint(method reflect.Method) *methodType {
	mtype := method.Type
	mname := method.Name
	var replyType, argType, contextType reflect.Type
	var stream bool

	// Endpoint() must be exported.
	if method.PkgPath != "" {
		return nil
	}

	switch mtype.NumIn() {
	case 3:
		// assuming streaming
		argType = mtype.In(2)
		contextType = mtype.In(1)
		stream = true
	case 4:
		// method that takes a context
		argType = mtype.In(2)
		replyType = mtype.In(3)
		contextType = mtype.In(1)
	default:
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("method %v of %v has wrong number of ins: %v", mname, mtype, mtype.NumIn())
		}
		return nil
	}

	if stream {
		// check stream type
		streamType := reflect.TypeOf((*server.Stream)(nil)).Elem()
		if !argType.Implements(streamType) {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("%v argument does not implement Streamer interface: %v", mname, argType)
			}
			return nil
		}
	} else {
		// if not stream check the replyType

		// First arg need not be a pointer.
		if !isExportedOrBuiltinType(argType) {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("%v argument type not exported: %v", mname, argType)
			}
			return nil
		}

		if replyType.Kind() != reflect.Ptr {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("method %v reply type not a pointer: %v", mname, replyType)
			}
			return nil
		}

		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				logger.Errorf("method %v reply type not exported: %v", mname, replyType)
			}
			return nil
		}
	}

	// Endpoint() needs one out.
	if mtype.NumOut() != 1 {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("method %v has wrong number of outs: %v", mname, mtype.NumOut())
		}
		return nil
	}
	// The return type of the method must be error.
	if returnType := mtype.Out(0); returnType != typeOfError {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Errorf("method %v returns %v not error", mname, returnType.String())
		}
		return nil
	}
	return &methodType{method: method, ArgType: argType, ReplyType: replyType, ContextType: contextType, stream: stream}
}

func (server *rpcServer) register(rcvr interface{}) error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typeReceiver = reflect.TypeOf(rcvr)
	s.receiverMethods = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.receiverMethods).Type().Name()
	if sname == "" {
		logger.Fatalf("rpc: no service name for type %v", s.typeReceiver.String())
	}
	if !isExported(sname) {
		s := "rpc Register: type " + sname + " is not exported"
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Error(s)
		}
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname
	s.method = make(map[string]*methodType)

	// Install the methods
	for m := 0; m < s.typeReceiver.NumMethod(); m++ {
		method := s.typeReceiver.Method(m)
		if mt := prepareEndpoint(method); mt != nil {
			s.method[method.Name] = mt
		}
	}

	if len(s.method) == 0 {
		s := "rpc Register: type " + sname + " has no exported methods of suitable type"
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			logger.Error(s)
		}
		return errors.New(s)
	}
	server.serviceMap[s.name] = s
	return nil
}

func (m *methodType) prepareContext(ctx context.Context) reflect.Value {
	if contextv := reflect.ValueOf(ctx); contextv.IsValid() {
		return contextv
	}
	return reflect.Zero(m.ContextType)
}
