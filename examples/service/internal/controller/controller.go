package controller

import (
	"context"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	typespb "github.com/lastbackend/toolkit/examples/service/gen/ptypes"
	"github.com/lastbackend/toolkit/examples/service/internal/repository"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
)

type Controller struct {
	app      toolkit.Service
	log      logger.Logger
	cfg      *config.Config
	repo     *repository.Repository
	services servicepb.ExampleServices
}

func (c *Controller) OnStart(ctx context.Context) error {
	c.log.Info("> service controller: on start")
	c.repo.Meta(ctx)
	c.Call(ctx)

	return nil
}

func (c *Controller) Call(_ context.Context) error {
	resp, err := c.services.Example().HelloWorld(context.Background(), &typespb.HelloWorldRequest{
		Name: "name",
		Type: "type",
		Data: nil,
	})
	if err != nil {
		c.log.Error(err.Error())
		return err
	}

	c.log.Info("> service response from server:> name:", resp.Name)
	return nil
}

func (c *Controller) OnStop(context.Context) error {
	c.log.Info("> service controller: on stop")
	return nil
}

func NewController(app toolkit.Service, cfg *config.Config, repo *repository.Repository, services servicepb.ExampleServices) *Controller {
	app.Log().Info("> service controller ----")
	ctrl := &Controller{app: app, log: app.Log(), cfg: cfg, repo: repo, services: services}
	return ctrl
}
