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

package logger

import (
	"go.uber.org/fx/fxevent"
	"io"
)

type Fields map[string]interface{}

type Level int8

const (
	FatalLevel Level = iota
	PanicLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

type Options struct {
	Verbose         Level `env:"VERBOSE"`
	JSONFormat      bool  `env:"JSON_FORMAT"`
	CallerSkipCount int
	Fields          Fields
	Out             io.Writer
	Tags            map[string]string
}

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	V(Level) Logger
	Inject(fn func(level Level))
	Fx() fxevent.Logger
}
