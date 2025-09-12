package controller

import (
	"context"

	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/tools/metrics"
	"github.com/lastbackend/toolkit/pkg/tools/metrics/providers"
	"github.com/lastbackend/toolkit/pkg/tools/probes"
	probesServer "github.com/lastbackend/toolkit/pkg/tools/probes/server"
	"github.com/lastbackend/toolkit/pkg/tools/traces"
)

type Tools struct {
	runtime.Tools

	runtime runtime.Runtime
	metrics metrics.Metrics
	probes  probes.Probes
	traces  traces.Traces
}

func (t *Tools) Probes() probes.Probes {
	return t.probes
}

func (t *Tools) Traces() traces.Traces {
	return t.traces
}

func (t *Tools) Metrics() metrics.Metrics {
	return t.metrics
}

// SetMetricsProvider sets the metrics provider
func (t *Tools) SetMetricsProvider(provider string) error {
	newMetrics, err := providers.NewMetrics(provider, t.runtime)
	if err != nil {
		return err
	}
	t.metrics = newMetrics
	return nil
}

// OnStart start tools services
func (t *Tools) OnStart(ctx context.Context) error {
	return nil
}

// StartProbes starts the probes server
func (t *Tools) StartProbes(ctx context.Context) error {
	return t.probes.Start(ctx)
}

// StartMetrics starts the metrics server with specified provider
func (t *Tools) StartMetrics(ctx context.Context, provider string) error {
	// Set provider if specified
	if provider != "" {
		if err := t.SetMetricsProvider(provider); err != nil {
			return err
		}
	}

	// Start metrics server
	return t.metrics.Start(ctx)
}

func newToolsRegistration(runtime runtime.Runtime) (runtime.Tools, error) {
	var (
		tools = new(Tools)
		err   error
	)
	tools.runtime = runtime

	// Initialize metrics with default provider (prometheus)
	if tools.metrics, err = providers.NewMetrics("", runtime); err != nil {
		return nil, err
	}

	// Initialize probes
	if tools.probes, err = probesServer.NewProbesServer(runtime); err != nil {
		return nil, err
	}

	return tools, nil
}
