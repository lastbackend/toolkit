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

// Labels represents metric labels
type Labels map[string]string

// Counter represents a monotonically increasing metric
type Counter interface {
	// Inc increments the counter by 1
	Inc(labels ...Labels)
	// Add increments the counter by the given value
	Add(value float64, labels ...Labels)
	// With returns a counter with the given labels
	With(labels Labels) Counter
}

// Gauge represents a metric that can go up and down
type Gauge interface {
	// Set sets the gauge to the given value
	Set(value float64, labels ...Labels)
	// Inc increments the gauge by 1
	Inc(labels ...Labels)
	// Dec decrements the gauge by 1
	Dec(labels ...Labels)
	// Add adds the given value to the gauge
	Add(value float64, labels ...Labels)
	// Sub subtracts the given value from the gauge
	Sub(value float64, labels ...Labels)
	// With returns a gauge with the given labels
	With(labels Labels) Gauge
}

// Histogram represents a metric for observing distributions
type Histogram interface {
	// Observe adds an observation to the histogram
	Observe(value float64, labels ...Labels)
	// With returns a histogram with the given labels
	With(labels Labels) Histogram
}

// Timer provides convenient time measurement methods
type Timer interface {
	// Start starts the timer and returns a function to record the duration
	Start(labels ...Labels) func()
	// Observe records a duration in seconds
	Observe(seconds float64, labels ...Labels)
	// With returns a timer with the given labels
	With(labels Labels) Timer
}

// MetricsServer extends metrics with server functionality for tools
type Metrics interface {
	// Counter creates or retrieves a counter metric
	Counter(name, help string, labelKeys ...string) Counter

	// Gauge creates or retrieves a gauge metric
	Gauge(name, help string, labelKeys ...string) Gauge

	// Histogram creates or retrieves a histogram metric
	Histogram(name, help string, labelKeys []string, buckets ...float64) Histogram

	// Timer creates or retrieves a timer (histogram) metric for measuring durations
	Timer(name, help string, labelKeys []string, buckets ...float64) Timer

	// Handler returns HTTP handler for exposing metrics
	Handler() http.Handler

	// Start starts the metrics HTTP server
	Start(ctx context.Context) error
}
