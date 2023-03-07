package controller

import (
	"context"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/runtime/meta"
)

type service struct {
	toolkit.Service
	runtime runtime.Runtime
	start   func(ctx context.Context, invoke ...interface{}) error
}

func (s *service) Meta() *meta.Meta {
	return s.runtime.Meta()
}

func (s *service) Log() logger.Logger {
	return s.runtime.Log()
}

func (s *service) Client() toolkit.Client {
	return s.runtime.Client()
}

func (s *service) Config() toolkit.Config {
	return s.runtime.Config()
}

func (s *service) Server() toolkit.Server {
	return s.runtime.Server()
}

func (s *service) Package() toolkit.Package {
	return s.runtime.Package()
}

func (s *service) Plugin() toolkit.Plugin {
	return s.runtime.Plugin()
}

func (s *service) Run(ctx context.Context, invoke ...interface{}) error {
	return s.runtime.Run(ctx, invoke...)
}

func (s *service) Start(ctx context.Context, invoke ...interface{}) error {
	return s.runtime.Start(ctx, invoke...)
}

func (s *service) Stop(ctx context.Context) error {
	return s.runtime.Stop(ctx)
}
