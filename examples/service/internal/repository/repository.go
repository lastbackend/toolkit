package repository

import (
	"context"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/plugin/postgres_gorm"
)

// The Repository represents Model.
type Repository struct {
	postgres_gorm.Plugin
	cfg *config.Config
	log logger.Logger
}

type Tk struct {
	Id    int
	Count int
}

func (Tk) TableName() string {
	return "toolkit"
}

func (r *Repository) Meta(_ context.Context) {
	val := new(Tk)
	r.DB().First(val)
	r.log.Info(">> meta count:", val.Count)
}

func (r *Repository) Get(_ context.Context) *Tk {
	val := new(Tk)
	r.DB().First(val)
	return val
}

func (r *Repository) OnStart(ctx context.Context) {
	val := new(Tk)
	r.DB().First(val)
	r.log.Info(">> onStart count:", val.Count)
	val.Count = val.Count + 1
	r.DB().Save(val)
}

// NewRepository Model with given configurations.
func NewRepository(app toolkit.Service, psql servicepb.PgsqlPlugin, cfg *config.Config) *Repository {
	repo := &Repository{psql, cfg, app.Log()}
	app.Log().Info("repository init")
	return repo
}
