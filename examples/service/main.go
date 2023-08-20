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
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/examples/service/internal/controller"
	"github.com/lastbackend/toolkit/examples/service/internal/repository"
	"github.com/lastbackend/toolkit/examples/service/internal/server"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server/http"
	"os"
	"time"
)

func main() {
	// define service with name and options
	app, err := servicepb.NewExampleService("example",
		runtime.WithVersion("0.1.0"),
		runtime.WithDescription("Example microservice"),
		runtime.WithEnvPrefix("LB"),
	)
	if err != nil {
		fmt.Println(err)
	}

	// Config management
	cfg := config.New()

	if err := app.RegisterConfig(cfg); err != nil {
		app.Log().Error(err)
		return
	}

	// Add packages
	app.RegisterPackage(repository.NewRepository, controller.NewController)

	// Add server
	app.Server().GRPC().SetService(server.NewServer)
	app.Server().GRPC().SetInterceptor(server.NewExampleGRPCServerInterceptor)

	app.Server().HTTPNew("", nil)
	app.Server().HTTP().SetMiddleware(server.RegisterExampleHTTPServerMiddleware)
	app.Server().HTTP().AddHandler(http.MethodGet, "/", server.ExampleHTTPServerHandler, http.WithMiddleware(server.MWAuthenticate))

	app.RegisterOnStartHook(func(ctx context.Context) error {
		time.Sleep(3 * time.Second)
		app.Log().Info("call gracefully stop")
		app.Stop(ctx, fmt.Errorf("test error"))
		return nil
	})

	//go func() {
	//	time.Sleep(5 * time.Second)
	//	app.Log().Info("call gracefully stop")
	//	app.Stop(context.Background())
	//}()

	// Service run
	if err := app.Start(context.Background()); err != nil {
		app.Log().Errorf("could not run the service %v", err)
		os.Exit(1)
		return
	}

	// time.Sleep(10 * time.Second)
	// app.Stop(context.Background())

	app.Log().Info("graceful stop")
}
