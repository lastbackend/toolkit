package controller

import (
	"github.com/lastbackend/toolkit/example/config"
	"github.com/lastbackend/toolkit/example/internal/repository"
)

type Controller struct {
	cfg  *config.Config
	repo *repository.Repository
}

//func (c *Controller) PreStart() error {
//	fmt.Println("pre start")
//	c.repo.DB().Select("1")
//	return nil
//}

func NewController(cfg *config.Config, repo *repository.Repository) *Controller {
	return &Controller{cfg: cfg, repo: repo}
}
