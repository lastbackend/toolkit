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
	"fmt"
	"github.com/lastbackend/toolkit/examples/service/internal/repository"

	"time"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	typespb "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
)

type Handlers struct {
	servicepb.ExampleRpcServer

	app  toolkit.Service
	cfg  *config.Config
	repo *repository.Repository
}

func (h Handlers) HelloWorld(ctx context.Context, req *typespb.HelloWorldRequest) (*typespb.HelloWorldResponse, error) {

	h.app.Log().Info("ExamplseRpcServer: HelloWorld: call")
	tk := h.repo.Get(ctx)

	resp := typespb.HelloWorldResponse{
		Id:        "",
		Name:      fmt.Sprintf("%s: %d", req.Name, tk.Count),
		Type:      req.Type,
		Data:      nil,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return &resp, nil
}

func NewServer(app toolkit.Service, cfg *config.Config, repo *repository.Repository) servicepb.ExampleRpcServer {
	return &Handlers{
		repo: repo,
		app:  app,
		cfg:  cfg,
	}
}
