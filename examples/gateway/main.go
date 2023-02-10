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
	"io"
	"net/http"
	"os"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/gateway/config"
	pb "github.com/lastbackend/toolkit/examples/gateway/gen/server"
	"github.com/lastbackend/toolkit/examples/gateway/middleware"
	"github.com/lastbackend/toolkit/pkg/logger"
	"github.com/lastbackend/toolkit/pkg/router"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, `{"alive": true}`); err != nil {
		fmt.Println(err)
	}
}

func main() {
	log := logger.DefaultLogger
	log = log.WithFields(logger.Fields{
		"service": "gateway",
	})

	log.Infof("Start process")

	svc := pb.NewService("gateway")
	svc.Meta().SetEnvPrefix("GTW")

	cfg := config.New()

	setFlags(svc.CLI(), cfg)

	svc.SetConfig(cfg)
	svc.AddMiddleware(middleware.New)

	svc.Router().Handle(http.MethodGet, "/health", HealthCheckHandler, router.HandleOptions{})

	if err := svc.Run(context.Background()); err != nil {
		log.Errorf("Failed run service: %v", err)
		os.Exit(1)
		return
	}
}

func setFlags(cli toolkit.CLI, cfg *config.Config) {

	// describe flag: "Set demo-flag flag"
	cli.AddStringFlag("demo-flag", &cfg.Demo).
		Env("DEMO_FLAG").
		Usage("Set demo flag")

}
