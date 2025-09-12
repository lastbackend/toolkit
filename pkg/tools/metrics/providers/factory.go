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

package providers

import (
	"fmt"

	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/tools/metrics"
	"github.com/lastbackend/toolkit/pkg/tools/metrics/prometheus"
)

// Provider types
const (
	ProviderPrometheus = "prometheus"
	ProviderInfluxDB   = "influxdb"
	ProviderStatsD     = "statsd"
	ProviderNoop       = "noop"
)

// MetricsFactory creates metrics instances based on provider
type MetricsFactory interface {
	Create(provider string, runtime runtime.Runtime) (metrics.Metrics, error)
}

type defaultFactory struct{}

// NewMetricsFactory returns a new metrics factory
func NewMetricsFactory() MetricsFactory {
	return &defaultFactory{}
}

// Create creates a metrics instance based on the provider
func (f *defaultFactory) Create(provider string, runtime runtime.Runtime) (metrics.Metrics, error) {
	switch provider {
	case ProviderPrometheus:
		return f.createPrometheus(runtime)
	case ProviderInfluxDB:
		return f.createInfluxDB(runtime)
	case ProviderStatsD:
		return f.createStatsD(runtime)
	case ProviderNoop:
		return f.createNoop(runtime)
	default:
		return nil, fmt.Errorf("unknown metrics provider: %s", provider)
	}
}

// createPrometheus creates a Prometheus metrics instance
func (f *defaultFactory) createPrometheus(runtime runtime.Runtime) (metrics.Metrics, error) {
	return prometheus.NewPrometheusMetrics(runtime)
}

// createInfluxDB creates an InfluxDB metrics instance
func (f *defaultFactory) createInfluxDB(runtime runtime.Runtime) (metrics.Metrics, error) {
	// TODO: Implement InfluxDB provider
	return nil, fmt.Errorf("InfluxDB metrics provider not implemented yet")
}

// createStatsD creates a StatsD metrics instance
func (f *defaultFactory) createStatsD(runtime runtime.Runtime) (metrics.Metrics, error) {
	// TODO: Implement StatsD provider
	return nil, fmt.Errorf("StatsD metrics provider not implemented yet")
}

// createNoop creates a no-op metrics instance
func (f *defaultFactory) createNoop(runtime runtime.Runtime) (metrics.Metrics, error) {
	return metrics.NewNoopMetrics(), nil
}

// NewMetrics creates a metrics instance using the default factory
// If provider is empty, defaults to prometheus
func NewMetrics(provider string, runtime runtime.Runtime) (metrics.Metrics, error) {
	factory := NewMetricsFactory()

	if provider == "" {
		provider = ProviderPrometheus
	}

	return factory.Create(provider, runtime)
}
