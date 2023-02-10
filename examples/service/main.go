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
	"os"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/examples/service/internal/controller"
	"github.com/lastbackend/toolkit/examples/service/internal/repository"
	"github.com/lastbackend/toolkit/examples/service/internal/server"
	"github.com/lastbackend/toolkit/pkg/logger"
)

func main() {

	svc := servicepb.NewService("example")

	log := svc.Logger()
	log.WithFields(logger.Fields{
		"microservice": "example",
	})
	log.Info("Run microservice")

	svc.Meta().
		SetVersion("0.1.0").
		SetEnvPrefix("LB").
		SetShortDescription("Example microservice").
		SetLongDescription("Microservice for development needs")

	cfg := config.New()
	setFlags(svc.CLI(), cfg)

	svc.SetConfig(cfg)
	svc.AddPackage(repository.NewRepository)
	svc.AddPackage(controller.NewController)
	svc.SetServer(server.NewServer)

	svc.Invoke(func(c *controller.Controller) error {

		//if err := c.PreStart(); err != nil {
		//	return err
		//}

		return nil
	})

	if err := svc.Run(context.Background()); err != nil {
		log.Errorf("could not run the service %v", err)
		os.Exit(1)
		return
	}
}

func setFlags(cli toolkit.CLI, cfg *config.Config) {

	// describe flag: "Set access token"
	cli.AddStringFlag("demo-flag", &cfg.Nested.Demo).
		Env("DEMO_KEY").
		Usage("Set demo key for example").
		Required()

}
