/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/lastbackend/toolkit/pkg/cmd"
	"github.com/lastbackend/toolkit/pkg/probe/types"
)

const (
	defaultPrefix         = "probe"
	defaultLivenessRoute  = "/healthz/live"
	defaultReadinessRoute = "/healthz/ready"
	defaultMetricsRoute   = "/metrics"
	defaultPort           = 9090

	defaultContentType = "application/json; charset=utf-8"
)

type probe struct {
	http.ServeMux

	mtx sync.RWMutex

	livenessProbes  map[string]types.HandleFunc
	readinessProbes map[string]types.HandleFunc
	metrics         map[string]types.HandleFunc

	prefix    string
	envPrefix string

	disable        bool
	host           string
	port           int
	livenessRoute  string
	readinessRoute string
	metricsRoute   string

	server *http.Server
}

func NewProbe() types.Probe {
	return &probe{
		livenessProbes:  make(map[string]types.HandleFunc),
		readinessProbes: make(map[string]types.HandleFunc),
		metrics:         make(map[string]types.HandleFunc),
		prefix:          defaultPrefix,
	}
}

func (p *probe) Init(prefix string, cli cmd.FlagSet) {
	p.envPrefix = prefix

	cli.AddBoolFlag(p.withPrefix("disable"), &p.disable).
		Env(p.generateEnvName("DISABLE")).
		Default(false).
		Usage("Disable probes checker")

	cli.AddStringFlag(p.withPrefix("host"), &p.host).
		Env(p.generateEnvName("HOST")).
		Usage("Sets probe HTTP server port")

	cli.AddIntFlag(p.withPrefix("port"), &p.port).
		Env(p.generateEnvName("PORT")).
		Default(defaultPort).
		Usage("Sets probe HTTP server port")

	cli.AddStringFlag(p.withPrefix("liveness-path"), &p.livenessRoute).
		Env(p.generateEnvName("LIVENESS_URI")).
		Default(defaultLivenessRoute).
		Usage("Sets liveness uri path")

	cli.AddStringFlag(p.withPrefix("readiness-path"), &p.readinessRoute).
		Env(p.generateEnvName("READINESS_URI")).
		Default(defaultReadinessRoute).
		Usage("Sets readiness uri path")

	cli.AddStringFlag(p.withPrefix("metrics-path"), &p.metricsRoute).
		Env(p.generateEnvName("METRICS_URI")).
		Default(defaultMetricsRoute).
		Usage("Sets metrics uri path")
}

func (p *probe) LiveEndpoint(w http.ResponseWriter, r *http.Request) {
	p.probeHandle(w, r, p.livenessProbes)
}

func (p *probe) ReadyEndpoint(w http.ResponseWriter, r *http.Request) {
	p.probeHandle(w, r, p.readinessProbes, p.livenessProbes)
}

func (p *probe) MetricsEndpoint(w http.ResponseWriter, r *http.Request) {
	p.metricsHandle(w, r, p.metrics)
}

func (p *probe) AddLivenessFunc(name string, fn types.HandleFunc) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.livenessProbes[name] = fn
}

func (p *probe) AddReadinessFunc(name string, fn types.HandleFunc) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.readinessProbes[name] = fn
}

func (p *probe) AddMetricsFunc(name string, fn types.HandleFunc) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.metrics[name] = fn
}

func (p *probe) checkProbes(probes map[string]types.HandleFunc, resultsOut map[string]string, statusOut *int) {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	for name, probe := range probes {
		if err := probe(); err != nil {
			*statusOut = http.StatusBadRequest
			resultsOut[name] = err.Error()
		} else {
			resultsOut[name] = "OK"
		}
	}
}

func (p *probe) Start(ctx context.Context) error {

	if p.disable {
		return nil
	}

	p.Handle(p.livenessRoute, http.HandlerFunc(p.LiveEndpoint))
	p.Handle(p.readinessRoute, http.HandlerFunc(p.ReadyEndpoint))
	p.Handle(p.metricsRoute, http.HandlerFunc(p.MetricsEndpoint))

	p.server = &http.Server{Addr: fmt.Sprintf("%s:%d", p.host, p.port), Handler: p}

	go func() {
		<-ctx.Done()
		p.server.Shutdown(ctx)
	}()

	return p.server.ListenAndServe()
}

func (p *probe) Stop() error {
	if p.disable {
		return nil
	}

	return p.server.Shutdown(context.Background())
}

func (p *probe) probeHandle(w http.ResponseWriter, r *http.Request, probes ...map[string]types.HandleFunc) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var (
		result = make(map[string]string, 0)
		status = http.StatusOK
	)

	for _, v := range probes {
		p.checkProbes(v, result, &status)
	}

	w.Header().Set("Content-Type", defaultContentType)
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(result); err != nil {
		return
	}
}

func (p *probe) metricsHandle(w http.ResponseWriter, r *http.Request, metrics ...map[string]types.HandleFunc) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", defaultContentType)
	w.WriteHeader(http.StatusOK)
}

func (p *probe) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", p.prefix, name)
}

func (p *probe) generateEnvName(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.prefix, strings.Replace(name, "-", "_", -1)))
}

func (p *probe) generateWithEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.envPrefix, p.generateEnvName(name)))
}
