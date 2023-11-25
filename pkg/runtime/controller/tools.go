package controller

import (
	"context"

	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/tools/metrics"
	"github.com/lastbackend/toolkit/pkg/tools/probes"
	"github.com/lastbackend/toolkit/pkg/tools/probes/server"
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

func (t *Tools) OnStart(ctx context.Context) error {
	return t.probes.Start(ctx)
}

func newToolsRegistration(runtime runtime.Runtime) (runtime.Tools, error) {
	var (
		tools = new(Tools)
		err   error
	)
	tools.runtime = runtime

	if tools.probes, err = server.NewProbesServer(runtime); err != nil {
		return nil, err
	}

	return tools, nil
}
