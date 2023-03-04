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
	log      logger.Logger
	cfg      *config.Config
	repo     *repository.Repository
	services servicepb.Services
}

func (c *Controller) OnStart(ctx context.Context) error {
	c.log.Info("controller: on start")
	c.repo.Meta(ctx)
	//c.repo.Info()

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

	c.log.Info(resp.Name)
	return nil
}

func (c *Controller) OnStop(ctx context.Context) error {
	c.log.Info("controller: on stop")
	return nil
}

func NewController(app toolkit.Service, cfg *config.Config, repo *repository.Repository, services servicepb.Services) *Controller {
	app.Log().Info("---- new controller ----")
	ctrl := &Controller{log: app.Log(), cfg: cfg, repo: repo, services: services}
	app.Package().Register(ctrl)
	return ctrl
}

func Start(ctx context.Context, app toolkit.Service, ctr *Controller) {
	app.Log().Info("---- controller start ----")
	ctr.repo.DB().Raw("SELECT 1")
	ctr.Call(ctx)
}
