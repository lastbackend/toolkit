package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	"github.com/lastbackend/toolkit/examples/service/internal/repository"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
)

type Controller struct {
	log  logger.Logger
	cfg  *config.Config
	repo *repository.Repository
}

func (c *Controller) OnStart(ctx context.Context) error {
	c.log.Info("controller: on start")
	c.repo.Meta(ctx)
	c.repo.Info()
	return nil
}

func (c *Controller) OnStop(ctx context.Context) error {
	c.log.Info("controller: on stop")
	return nil
}

func NewController(app toolkit.Service, cfg *config.Config, repo *repository.Repository) *Controller {
	app.Log().Info("---- new controller ----")
	ctrl := &Controller{log: app.Log(), cfg: cfg, repo: repo}
	app.Package().Register(ctrl)
	return ctrl
}

func Start(ctx context.Context, app toolkit.Service, ctr *Controller) {
	app.Log().Info("---- controller start ----")
	ctr.repo.DB().Raw("SELECT 1")
	return
}
