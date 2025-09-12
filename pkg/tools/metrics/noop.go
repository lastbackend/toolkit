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

package metrics

import (
	"context"
	"net/http"
)

// NoopMetrics implements Metrics interface with no-op behavior
type NoopMetrics struct{}

// NewNoopMetrics creates a new no-op metrics instance
func NewNoopMetrics() Metrics {
	return &NoopMetrics{}
}

func (n *NoopMetrics) Counter(name, help string, labelKeys ...string) Counter {
	return &noopCounter{}
}

func (n *NoopMetrics) Gauge(name, help string, labelKeys ...string) Gauge {
	return &noopGauge{}
}

func (n *NoopMetrics) Histogram(name, help string, labelKeys []string, buckets ...float64) Histogram {
	return &noopHistogram{}
}

func (n *NoopMetrics) Timer(name, help string, labelKeys []string, buckets ...float64) Timer {
	return &noopTimer{}
}

func (n *NoopMetrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# No metrics enabled\n"))
	})
}

func (n *NoopMetrics) Start(ctx context.Context) error {
	return nil // No server to start
}

// Noop implementations
type noopCounter struct{}

func (c *noopCounter) Inc(labels ...Labels)                {}
func (c *noopCounter) Add(value float64, labels ...Labels) {}
func (c *noopCounter) With(labels Labels) Counter          { return c }

type noopGauge struct{}

func (g *noopGauge) Set(value float64, labels ...Labels) {}
func (g *noopGauge) Inc(labels ...Labels)                {}
func (g *noopGauge) Dec(labels ...Labels)                {}
func (g *noopGauge) Add(value float64, labels ...Labels) {}
func (g *noopGauge) Sub(value float64, labels ...Labels) {}
func (g *noopGauge) With(labels Labels) Gauge            { return g }

type noopHistogram struct{}

func (h *noopHistogram) Observe(value float64, labels ...Labels) {}
func (h *noopHistogram) With(labels Labels) Histogram            { return h }

type noopTimer struct{}

func (t *noopTimer) Start(labels ...Labels) func() {
	return func() {} // Return empty function
}
func (t *noopTimer) Observe(seconds float64, labels ...Labels) {}
func (t *noopTimer) With(labels Labels) Timer                  { return t }
