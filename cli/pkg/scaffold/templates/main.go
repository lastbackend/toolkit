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

package templates

// Main is the main template used for new services.
var Main = `package main

import (
	"context"
	"os"

	servicepb "{{.Vendor}}{{lower .Service}}/gen"
  "{{.Vendor}}{{lower .Service}}/config"
	"{{.Vendor}}{{lower .Service}}/internal/server"
)

func main() {

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

	svc.SetServer(server.NewServer)

	if err := app.Start(context.Background()); err != nil {
		app.Log().Errorf("could not run the service %v", err)
		os.Exit(1)
		return
	}

	// time.Sleep(10 * time.Second)
	// app.Stop(context.Background())

	app.Log().Info("graceful stop")
}
`
