package repository

import (
	"context"

	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit-plugins/postgres_gorm"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
)

// The Repository represents Model.
type Repository struct {
	postgres_gorm.Plugin
	log logger.Logger
}

type DemoStruct struct {
	Id    int
	Count int
}

func (r *Repository) Get(_ context.Context) *DemoStruct {
	val := new(DemoStruct)
	r.DB().First(val)
	return val
}

// NewRepository Model with given configurations.
func NewRepository(app toolkit.Service, psql servicepb.PgsqlPlugin) *Repository {
	repo := &Repository{psql, app.Log()}
	app.Log().Info("repository init")
	return repo
}
