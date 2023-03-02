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

package main

import (
	"context"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime/types"
	"io"
	"net/http"
	"os"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/gateway/config"
	pb "github.com/lastbackend/toolkit/examples/gateway/gen/server"
	"github.com/lastbackend/toolkit/examples/gateway/middleware"
	"github.com/lastbackend/toolkit/pkg/http"
	"github.com/lastbackend/toolkit/pkg/logger"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, `{"alive": true}`); err != nil {
		fmt.Println(err)
	}
}

func main() {


	svc := pb.NewService("gateway", types.WithEnvPrefix("GTW"))
	svc.Log().Infof("Start process")

	cfg := config.New()

	svc.Config.Provide(cfg)
	svc.Server().HTTP()..AddMiddleware(middleware.New)

	svc.Server().HTTP().
		AddHandler(http.MethodGet, "/health", HealthCheckHandler)

	if err := svc.Run(context.Background()); err != nil {
		svc.Log().Errorf("Failed run service: %v", err)
		os.Exit(1)
		return
	}
}
