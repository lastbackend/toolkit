package probes

import (
	"context"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"sync"
)

const prefix = "probes"

type Options struct {
	Host          string
	Post          string
	LivenessPath  string
	ReadinessPath string
}

type HandleFunc func() error

type server struct {
	mtx     sync.RWMutex
	runtime runtime.Runtime

	opts            Options
	livenessProbes  map[string]HandleFunc
	readinessProbes map[string]HandleFunc
}

func NewProbesServer(runtime runtime.Runtime) error {
	srv := new(server)

	srv.runtime = runtime

	srv.readinessProbes = make(map[string]HandleFunc)
	srv.livenessProbes = make(map[string]HandleFunc)

	runtime.Config().Parse(&srv.opts, prefix)

	return nil
}

func (s *server) AddLivenessCheck(name string, fn HandleFunc) error {

	if _, ok := s.livenessProbes[name]; ok {
		return fmt.Errorf("trying to override liveness probe check: %s", name)
	}

	s.mtx.Lock()
	s.livenessProbes[name] = fn
	s.mtx.Unlock()

	return nil
}

func (s *server) AddReadinessCheck(name string, fn HandleFunc) error {

	if _, ok := s.readinessProbes[name]; ok {
		return fmt.Errorf("trying to override redisness probe check: %s", name)
	}

	s.mtx.Lock()
	s.readinessProbes[name] = fn
	s.mtx.Unlock()

	return nil
}

func (s *server) Start(ctx context.Context) error {

	return nil
}
