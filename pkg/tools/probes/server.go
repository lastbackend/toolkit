package probes

import (
	"context"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"net/http"
	"sync"
)

const prefix = "probes"

type ProbeKind int

const (
	ReadinessProbe ProbeKind = iota
	LivenessProbe
)

type Probes interface {
	RegisterCheck(name string, kind ProbeKind, fn HandleFunc)
}

type Options struct {
	Host string
	Post string

	LivenessPath  string
	ReadinessPath string
}

type HandleFunc func() error

type server struct {
	mtx     sync.RWMutex
	runtime runtime.Runtime

	opts            Options
	readinessProbes map[string]HandleFunc
	livenessProbes  map[string]HandleFunc
}

func NewProbesServer(runtime runtime.Runtime) Probes {
	srv := new(server)

	srv.runtime = runtime

	srv.readinessProbes = make(map[string]HandleFunc, 0)
	srv.livenessProbes = make(map[string]HandleFunc, 0)

	runtime.Config().Parse(&srv.opts, prefix)

	return nil
}

func (s *server) RegisterCheck(name string, kind ProbeKind, fn HandleFunc) error {

	s.mtx.Lock()
	defer s.mtx.Unlock()

	switch kind {
	case LivenessProbe:
		if _, ok := s.livenessProbes[name]; ok {
			return fmt.Errorf("trying to override liveness probe check: %s", name)
		}
		s.livenessProbes[name] = fn
	case ReadinessProbe:
		if _, ok := s.readinessProbes[name]; ok {
			return fmt.Errorf("trying to override liveness probe check: %s", name)
		}
		s.readinessProbes[name] = fn
	}

	return nil
}

func (s *server) probeHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) Start(ctx context.Context) error {
	//
	//var srv = s.runtime.Server().HTTP()
	//for _, probe := range s.readinessProbes {
	//	srv.AddHandler(http.MethodGet, s.opts.LivenessPath)
	//}

	return nil
}
