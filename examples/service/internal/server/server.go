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

package server

import (
	"context"

	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	typespb "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
)

type Handlers struct {
	svc servicepb.Service
	cfg *config.Config
}

func (h Handlers) HelloWorld(ctx context.Context, req *typespb.HelloWorldRequest) (*typespb.HelloWorldResponse, error) {
	//TODO implement me
	panic("implement me")
}

func NewServer(svc servicepb.Service, cfg *config.Config) servicepb.ExampleRpcServer {
	return &Handlers{
		svc: svc,
		cfg: cfg,
	}
}
