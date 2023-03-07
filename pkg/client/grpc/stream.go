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

package grpc

import (
	"github.com/lastbackend/toolkit/pkg/client"
	"google.golang.org/grpc"

	"context"
	"sync"
)

type stream struct {
	sync.RWMutex
	context context.Context

	grpc.ClientStream

	request *client.GRPCRequest
	conn    *poolConn
	close   func(err error)
}

func (s *stream) Context() context.Context {
	return s.context
}

func (s *stream) Headers() map[string]string {
	return s.request.Headers()
}

func (s *stream) SendMsg(msg interface{}) error {
	return s.ClientStream.SendMsg(msg)
}

func (s *stream) RecvMsg(msg interface{}) (err error) {
	return s.ClientStream.RecvMsg(msg)
}

func (s *stream) CloseSend() error {
	return s.ClientStream.CloseSend()
}
