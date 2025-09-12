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

package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/server"
	"github.com/lastbackend/toolkit/pkg/tools/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const prefix = "metrics"

const (
	defaultMetricsHttpServerName string = "metrics"
)

type Options struct {
	Enabled bool   `env:"SERVER_ENABLED" envDefault:"true" comment:"Enable or disable metrics server"`
	Host    string `env:"SERVER_LISTEN" envDefault:"0.0.0.0" comment:"Set metrics server listen host"`
	Port    int    `env:"SERVER_PORT" envDefault:"8080" comment:"Set metrics server listen port"`
	Path    string `env:"SERVER_PATH" envDefault:"/metrics" comment:"Set metrics endpoint path"`
}

type metricsServer struct {
	runtime  runtime.Runtime
	registry *prometheus.Registry
	opts     Options

	// Cache for created metrics
	counters   map[string]*prometheusCounter
	gauges     map[string]*prometheusGauge
	histograms map[string]*prometheusHistogram
	timers     map[string]*prometheusTimer
}

func NewPrometheusMetrics(runtime runtime.Runtime) (metrics.Metrics, error) {
	srv := &metricsServer{
		runtime:    runtime,
		registry:   prometheus.NewRegistry(),
		opts:       Options{},
		counters:   make(map[string]*prometheusCounter),
		gauges:     make(map[string]*prometheusGauge),
		histograms: make(map[string]*prometheusHistogram),
		timers:     make(map[string]*prometheusTimer),
	}

	return srv, runtime.Config().Parse(&srv.opts, prefix)
}

// Implement MetricsServer interface
func (m *metricsServer) Start(ctx context.Context) error {
	if !m.opts.Enabled {
		m.runtime.Log().V(5).Info("metrics server disabled")
		return nil
	}

	var s server.HTTPServer

	if m.opts.Port > 0 {
		// check if provided port is used by grpc server
		grpcServers := m.runtime.Server().GRPCList()
		for _, srv := range grpcServers {
			if srv.Info().Port == m.opts.Port {
				return fmt.Errorf("can not bind metrics handlers to grpc server. please change metrics port to different port")
			}
		}

		// check if provided port is used on provided http server
		httpServers := m.runtime.Server().HTTPList()
		for _, srv := range httpServers {
			if srv.Info().Port == m.opts.Port {
				s = srv
				m.runtime.Log().V(5).Infof("metrics reusing existing HTTP server on port %d", s.Info().Port)
				break
			}
		}
	}

	// Create new HTTP server if not found existing one
	if s == nil {
		if m.opts.Port <= 0 {
			return fmt.Errorf("metrics server: no port specified")
		}
		s = m.runtime.Server().HTTPNew(defaultMetricsHttpServerName, &server.HTTPServerOptions{
			Host:      m.opts.Host,
			Port:      m.opts.Port,
			TLSConfig: nil,
		})
		m.runtime.Log().V(5).Infof("metrics created new HTTP server on port %d", m.opts.Port)
	}

	// Add metrics endpoint
	s.AddHandler(http.MethodGet, m.opts.Path, func(w http.ResponseWriter, r *http.Request) {
		handler := promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
		handler.ServeHTTP(w, r)
	})

	m.runtime.Log().V(5).Infof("metrics endpoint added at %s", m.opts.Path)

	return nil
}

// Implement MetricsServer methods
func (m *metricsServer) Counter(name, help string, labelKeys ...string) metrics.Counter {
	if counter, exists := m.counters[name]; exists {
		return counter
	}

	promCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labelKeys)

	m.registry.MustRegister(promCounter)
	counter := &prometheusCounter{
		counter:    promCounter,
		labelKeys:  labelKeys,
		baseLabels: prometheus.Labels{},
	}
	m.counters[name] = counter
	return counter
}

func (m *metricsServer) Gauge(name, help string, labelKeys ...string) metrics.Gauge {
	if gauge, exists := m.gauges[name]; exists {
		return gauge
	}

	promGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labelKeys)

	m.registry.MustRegister(promGauge)
	gauge := &prometheusGauge{
		gauge:      promGauge,
		labelKeys:  labelKeys,
		baseLabels: prometheus.Labels{},
	}
	m.gauges[name] = gauge
	return gauge
}

func (m *metricsServer) Histogram(name, help string, labelKeys []string, buckets ...float64) metrics.Histogram {
	if histogram, exists := m.histograms[name]; exists {
		return histogram
	}

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	promHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelKeys)

	m.registry.MustRegister(promHistogram)
	histogram := &prometheusHistogram{
		histogram:  promHistogram,
		labelKeys:  labelKeys,
		baseLabels: prometheus.Labels{},
	}
	m.histograms[name] = histogram
	return histogram
}

func (m *metricsServer) Timer(name, help string, labelKeys []string, buckets ...float64) metrics.Timer {
	if timer, exists := m.timers[name]; exists {
		return timer
	}

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	promHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labelKeys)

	m.registry.MustRegister(promHistogram)
	timer := &prometheusTimer{
		histogram:  promHistogram,
		labelKeys:  labelKeys,
		baseLabels: prometheus.Labels{},
	}
	m.timers[name] = timer
	return timer
}

func (m *metricsServer) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Prometheus metric wrappers
type prometheusCounter struct {
	counter   *prometheus.CounterVec
	labelKeys []string
	baseLabels prometheus.Labels
}

// mergeLabels combines base labels with provided labels
func (c *prometheusCounter) mergeLabels(labels ...metrics.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	// Copy base labels first
	for k, v := range c.baseLabels {
		merged[k] = v
	}
	// Override with provided labels
	if len(labels) > 0 {
		for k, v := range labels[0] {
			merged[k] = v
		}
	}
	return merged
}

func (c *prometheusCounter) Inc(labels ...metrics.Labels) {
	if len(c.labelKeys) == 0 {
		// No labels defined, use simple counter
		c.counter.WithLabelValues().Inc()
	} else {
		merged := c.mergeLabels(labels...)
		c.counter.With(merged).Inc()
	}
}

func (c *prometheusCounter) Add(value float64, labels ...metrics.Labels) {
	if len(c.labelKeys) == 0 {
		// No labels defined, use simple counter
		c.counter.WithLabelValues().Add(value)
	} else {
		merged := c.mergeLabels(labels...)
		c.counter.With(merged).Add(value)
	}
}

func (c *prometheusCounter) With(labels metrics.Labels) metrics.Counter {
	newBaseLabels := prometheus.Labels{}
	// Copy existing base labels
	for k, v := range c.baseLabels {
		newBaseLabels[k] = v
	}
	// Override with provided labels
	for k, v := range labels {
		newBaseLabels[k] = v
	}
	return &prometheusCounter{
		counter:    c.counter,
		labelKeys:  c.labelKeys,
		baseLabels: newBaseLabels,
	}
}

type prometheusGauge struct {
	gauge      *prometheus.GaugeVec
	labelKeys  []string
	baseLabels prometheus.Labels
}

// mergeLabels combines base labels with provided labels
func (g *prometheusGauge) mergeLabels(labels ...metrics.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	// Copy base labels first
	for k, v := range g.baseLabels {
		merged[k] = v
	}
	// Override with provided labels
	if len(labels) > 0 {
		for k, v := range labels[0] {
			merged[k] = v
		}
	}
	return merged
}

func (g *prometheusGauge) Set(value float64, labels ...metrics.Labels) {
	if len(g.labelKeys) == 0 {
		g.gauge.WithLabelValues().Set(value)
	} else {
		merged := g.mergeLabels(labels...)
		g.gauge.With(merged).Set(value)
	}
}

func (g *prometheusGauge) Inc(labels ...metrics.Labels) {
	if len(g.labelKeys) == 0 {
		g.gauge.WithLabelValues().Inc()
	} else {
		merged := g.mergeLabels(labels...)
		g.gauge.With(merged).Inc()
	}
}

func (g *prometheusGauge) Dec(labels ...metrics.Labels) {
	if len(g.labelKeys) == 0 {
		g.gauge.WithLabelValues().Dec()
	} else {
		merged := g.mergeLabels(labels...)
		g.gauge.With(merged).Dec()
	}
}

func (g *prometheusGauge) Add(value float64, labels ...metrics.Labels) {
	if len(g.labelKeys) == 0 {
		g.gauge.WithLabelValues().Add(value)
	} else {
		merged := g.mergeLabels(labels...)
		g.gauge.With(merged).Add(value)
	}
}

func (g *prometheusGauge) Sub(value float64, labels ...metrics.Labels) {
	if len(g.labelKeys) == 0 {
		g.gauge.WithLabelValues().Sub(value)
	} else {
		merged := g.mergeLabels(labels...)
		g.gauge.With(merged).Sub(value)
	}
}

func (g *prometheusGauge) With(labels metrics.Labels) metrics.Gauge {
	newBaseLabels := prometheus.Labels{}
	// Copy existing base labels
	for k, v := range g.baseLabels {
		newBaseLabels[k] = v
	}
	// Override with provided labels
	for k, v := range labels {
		newBaseLabels[k] = v
	}
	return &prometheusGauge{
		gauge:      g.gauge,
		labelKeys:  g.labelKeys,
		baseLabels: newBaseLabels,
	}
}

type prometheusHistogram struct {
	histogram  *prometheus.HistogramVec
	labelKeys  []string
	baseLabels prometheus.Labels
}

// mergeLabels combines base labels with provided labels
func (h *prometheusHistogram) mergeLabels(labels ...metrics.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	// Copy base labels first
	for k, v := range h.baseLabels {
		merged[k] = v
	}
	// Override with provided labels
	if len(labels) > 0 {
		for k, v := range labels[0] {
			merged[k] = v
		}
	}
	return merged
}

func (h *prometheusHistogram) Observe(value float64, labels ...metrics.Labels) {
	if len(h.labelKeys) == 0 {
		h.histogram.WithLabelValues().Observe(value)
	} else {
		merged := h.mergeLabels(labels...)
		h.histogram.With(merged).Observe(value)
	}
}

func (h *prometheusHistogram) With(labels metrics.Labels) metrics.Histogram {
	newBaseLabels := prometheus.Labels{}
	// Copy existing base labels
	for k, v := range h.baseLabels {
		newBaseLabels[k] = v
	}
	// Override with provided labels
	for k, v := range labels {
		newBaseLabels[k] = v
	}
	return &prometheusHistogram{
		histogram:  h.histogram,
		labelKeys:  h.labelKeys,
		baseLabels: newBaseLabels,
	}
}

type prometheusTimer struct {
	histogram  *prometheus.HistogramVec
	labelKeys  []string
	baseLabels prometheus.Labels
}

// mergeLabels combines base labels with provided labels
func (t *prometheusTimer) mergeLabels(labels ...metrics.Labels) prometheus.Labels {
	merged := prometheus.Labels{}
	// Copy base labels first
	for k, v := range t.baseLabels {
		merged[k] = v
	}
	// Override with provided labels
	if len(labels) > 0 {
		for k, v := range labels[0] {
			merged[k] = v
		}
	}
	return merged
}

func (t *prometheusTimer) Start(labels ...metrics.Labels) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start).Seconds()
		if len(t.labelKeys) == 0 {
			t.histogram.WithLabelValues().Observe(duration)
		} else {
			merged := t.mergeLabels(labels...)
			t.histogram.With(merged).Observe(duration)
		}
	}
}

func (t *prometheusTimer) Observe(seconds float64, labels ...metrics.Labels) {
	if len(t.labelKeys) == 0 {
		t.histogram.WithLabelValues().Observe(seconds)
	} else {
		merged := t.mergeLabels(labels...)
		t.histogram.With(merged).Observe(seconds)
	}
}

func (t *prometheusTimer) With(labels metrics.Labels) metrics.Timer {
	newBaseLabels := prometheus.Labels{}
	// Copy existing base labels
	for k, v := range t.baseLabels {
		newBaseLabels[k] = v
	}
	// Override with provided labels
	for k, v := range labels {
		newBaseLabels[k] = v
	}
	return &prometheusTimer{
		histogram:  t.histogram,
		labelKeys:  t.labelKeys,
		baseLabels: newBaseLabels,
	}
}
