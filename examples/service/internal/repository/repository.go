package repository

import (
	"github.com/lastbackend/toolkit/examples/service/config"
	servicepb "github.com/lastbackend/toolkit/examples/service/gen"
	"github.com/lastbackend/toolkit/plugin/postgres_gorm"
)

// The Repository represents Model.
type Repository struct {
	postgres_gorm.Plugin
	cfg *config.Config
}

// NewRepository Model with given configurations.
func NewRepository(psql servicepb.PgsqlPlugin, cfg *config.Config) *Repository {
	return &Repository{psql, cfg}
}
