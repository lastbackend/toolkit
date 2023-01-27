package repository

import (
	"github.com/lastbackend/toolkit/example/config"
	servicepb "github.com/lastbackend/toolkit/example/gen"
	"github.com/lastbackend/toolkit/plugin/storage/postgres_gorm"
)

// The Repository represents Model.
type Repository struct {
	postgres_gorm.Plugin
	cfg *config.Config
}

// NewRepository Model with given configurations.
func NewRepository(psql servicepb.PgsqlStorage, cfg *config.Config) *Repository {
	return &Repository{psql, cfg}
}
