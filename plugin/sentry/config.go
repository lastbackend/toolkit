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

package sentry

import (
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap/zapcore"
)

// Configuration is a minimal set of parameters for Sentry integration.
type Configuration struct {
	// Tags are passed as is to the corresponding sentry.Event field.
	Tags map[string]string

	// LoggerNameKey is the key for zap logger name.
	// If not empty, the name is added to the rest of zapcore.Field(s),
	// so that be careful with key duplicates.
	// Leave LoggerNameKey empty to disable the feature.
	LoggerNameKey string

	// DisableStacktrace disables adding stacktrace to sentry.Event, if set.
	DisableStacktrace bool

	// Level is the minimal level of sentry.Event(s).
	Level zapcore.Level

	// EnableBreadcrumbs enables use of sentry.Breadcrumb(s).
	// This feature works only when you explicitly passed new scope.
	EnableBreadcrumbs bool

	// BreadcrumbLevel is the minimal level of sentry.Breadcrumb(s).
	// Breadcrumb specifies an application event that occurred before a Sentry event.
	// NewCore fails if BreadcrumbLevel is greater than Level.
	// The field is ignored, if EnableBreadcrumbs is not set.
	BreadcrumbLevel zapcore.Level

	// FlushTimeout is the timeout for flushing events to Sentry.
	FlushTimeout time.Duration

	// Hub overrides the sentry.CurrentHub value.
	// See sentry.Hub docs for more detail.
	Hub *sentry.Hub
}
