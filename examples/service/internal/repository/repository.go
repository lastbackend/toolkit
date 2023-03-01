package repository

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/pkg/logger"
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
	var val = new(Tk)
	r.DB().First(val)
	r.log.Info(">> meta count:", val.Count)
}

func (r *Repository) PreStart(ctx context.Context) {
	var val = new(Tk)
	r.DB().First(val)
	r.log.Info(">> pre start count:", val.Count)
	val.Count = val.Count + 1
	r.DB().Save(val)

}

// NewRepository Model with given configurations.
func NewRepository(app toolkit.Service, psql servicepb.PgsqlPlugin, cfg *config.Config) *Repository {
	repo := &Repository{psql, cfg, app.Log()}
	app.Log().Info("repository init")
	app.Package().Register(repo)
	return repo
}
