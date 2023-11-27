package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/tools/probes"
	"net/http"
	"sync"
)

const prefix = "probes"

const (
	defaultProbesHttpServerName string = "probes"
	defaultContentType                 = "application/json; charset=utf-8"
)

type Options struct {
	Enabled bool   `env:"SERVER_ENABLED" envDefault:"true" comment:"Enable or disable probes server"`
	Host    string `env:"SERVER_LISTEN" envDefault:"0.0.0.0" comment:"Set probes listen host"`
	Port    int    `env:"SERVER_PORT" envDefault:"8080" comment:"Set probes listen port"`

	LivenessPath  string `env:"LIVENESS_PATH" envDefault:"/_healthz/liveness" comment:"Set liveness probe path"`
	ReadinessPath string `env:"READINESS_PATH" envDefault:"/_healthz/readiness" comment:"Set readiness probe path"`
}

type probe struct {
	mtx     sync.RWMutex
	runtime runtime.Runtime

	opts            Options
	readinessProbes map[string]probes.HandleFunc
	livenessProbes  map[string]probes.HandleFunc
}

func NewProbesServer(runtime runtime.Runtime) (probes.Probes, error) {
	srv := new(probe)

	srv.runtime = runtime
	srv.opts = Options{}

	srv.readinessProbes = make(map[string]probes.HandleFunc, 0)
	srv.livenessProbes = make(map[string]probes.HandleFunc, 0)

	return srv, runtime.Config().Parse(&srv.opts, prefix)
}

func (p *probe) RegisterCheck(name string, kind probes.ProbeKind, fn probes.HandleFunc) error {

	p.mtx.Lock()
	defer p.mtx.Unlock()

	switch kind {
	case probes.LivenessProbe:
		if _, ok := p.livenessProbes[name]; ok {
			return fmt.Errorf("trying to override liveness probe check: %s", name)
		}
		p.livenessProbes[name] = fn
	case probes.ReadinessProbe:
		if _, ok := p.readinessProbes[name]; ok {
			return fmt.Errorf("trying to override liveness probe check: %s", name)
		}
		p.readinessProbes[name] = fn
	}

	return nil
}

func (p *probe) livenessProbeHandler(w http.ResponseWriter, r *http.Request) {
	p.probeHandler(w, r, p.livenessProbes)
}

func (p *probe) readinessProbeHandler(w http.ResponseWriter, r *http.Request) {
	p.probeHandler(w, r, p.readinessProbes)
}

func (p *probe) probeHandler(w http.ResponseWriter, _ *http.Request, probes map[string]probes.HandleFunc) {

	var (
		result = make(map[string]string, 0)
		status = http.StatusOK
	)

	for name, probe := range probes {
		if err := probe(); err != nil {

			status = http.StatusInternalServerError

			result[name] = err.Error()
		} else {
			result[name] = "OK"
		}
	}

	w.Header().Set("Content-Type", defaultContentType)
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(result); err != nil {
		return
	}
}

func (p *probe) Start(_ context.Context) error {

	var (
		s server.HTTPServer
	)

	if !p.opts.Enabled {
		return nil
	}

	if p.opts.Port > 0 {
		// check if provided port is used by grpc server
		grpcServers := p.runtime.Server().GRPCList()
		for _, srv := range grpcServers {
			if srv.Info().Port == p.opts.Port {
				return fmt.Errorf("can not bind probes handlers to grpc server. please change porbes port to different port")
			}
		}

		// check if provided port is used on provided http server
		httpServers := p.runtime.Server().HTTPList()
		for _, srv := range httpServers {
			if srv.Info().Port == p.opts.Port {
				s = srv
				break
			}
		}
	}

	if s == nil {
		s = p.runtime.Server().HTTPNew(defaultProbesHttpServerName, &server.HTTPServerOptions{
			Host:      p.opts.Host,
			Port:      p.opts.Port,
			TLSConfig: nil,
		})
	}

	s.AddHandler(http.MethodGet, p.opts.LivenessPath, p.livenessProbeHandler)
	s.AddHandler(http.MethodGet, p.opts.ReadinessPath, p.readinessProbeHandler)

	return nil
}
