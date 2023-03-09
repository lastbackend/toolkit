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

func (s *service) Server() toolkit.Server {
	return s.runtime.Server()
}

func (s *service) RegisterConfig(items ...any) error {
	return s.runtime.Config().Provide(items...)
}

func (s *service) RegisterPackage(items ...any) {
	s.runtime.Package().Provide(items...)
}

func (s *service) RegisterPlugin(items ...any) {
	s.runtime.Plugin().Provide(items...)
}

func (s *service) Start(ctx context.Context) error {
	return s.runtime.Start(ctx)
}

func (s *service) Stop(ctx context.Context, err error) {
	s.runtime.Stop(ctx, err)
}

func (s *service) RegisterOnStartHook(fn ...func(ctx context.Context) error) {
	s.runtime.RegisterOnStartHook(fn...)
}

func (s *service) RegisterOnStopHook(fn ...func(ctx context.Context) error) {
	s.runtime.RegisterOnStopHook(fn...)
}

func (s *service) RegisterOnStartSyncHook(fn ...func(ctx context.Context) error) {
	s.runtime.RegisterOnStartSyncHook(fn...)
}

func (s *service) RegisterOnStopSyncHook(fn ...func(ctx context.Context) error) {
	s.runtime.RegisterOnStopSyncHook(fn...)
}
