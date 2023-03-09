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

package zap

import (
	"github.com/lastbackend/toolkit/pkg/runtime"
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"github.com/lastbackend/toolkit/pkg/runtime/logger/empty"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"

	"sync"
)

type zapLogger struct {
	sync.RWMutex
	logger *zap.SugaredLogger
	opts   logger.Options
	empty  logger.Logger
}

func NewLogger(runtime runtime.Runtime, fields logger.Fields) logger.Logger {

	l := &zapLogger{
		empty: empty.NewLogger(),
	}

	runtime.Config().Parse(&l.opts, "")
	if l.opts.CallerSkipCount == 0 {
		l.opts.CallerSkipCount = 1
	}

	// info level enabler
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})

	// error and fatal level enabler
	errorFatalLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.ErrorLevel || level == zapcore.FatalLevel
	})

	// write syncers
	stdoutSyncer := zapcore.Lock(os.Stdout)
	stderrSyncer := zapcore.Lock(os.Stderr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if l.opts.JSONFormat {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	zapOptions := make([]zap.Option, 0)
	zapOptions = append(zapOptions, zap.AddCallerSkip(l.opts.CallerSkipCount))
	zapOptions = append(zapOptions, zap.AddCaller())

	// tee core
	core := zapcore.NewTee(
		zapcore.NewCore(
			encoder,
			stdoutSyncer,
			infoLevel,
		),
		zapcore.NewCore(
			encoder,
			stderrSyncer,
			errorFatalLevel,
		),
	)

	l.logger = zap.New(core, zapOptions...).Sugar().With(fields)

	return l
}

func (l *zapLogger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *zapLogger) Info(args ...interface{}) {
	l.logger.Info(args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *zapLogger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *zapLogger) Error(args ...interface{}) {
	l.logger.Error(args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func (l *zapLogger) Panic(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *zapLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}

func (l *zapLogger) V(level logger.Level) logger.Logger {

	if l.opts.Verbose < level {
		return l.empty
	}

	return l
}

func (l *zapLogger) Inject(fn func(level logger.Level)) {
	return
}

func (l *zapLogger) Fx() fxevent.Logger {

	if l.opts.Verbose < 6 {
		return l.empty.Fx()
	}

	return &fxevent.ZapLogger{
		Logger: l.logger.Desugar(),
	}
}

func getZapLevel(level logger.Level) zapcore.Level {
	switch level {
	case logger.DebugLevel:
		return zapcore.DebugLevel
	case logger.InfoLevel:
		return zapcore.InfoLevel
	case logger.WarnLevel:
		return zapcore.WarnLevel
	case logger.ErrorLevel:
		return zapcore.ErrorLevel
	case logger.FatalLevel:
		return zapcore.FatalLevel
	case logger.PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *zapLogger) getFields(fields logger.Fields) []interface{} {
	fields = l.fieldsMerge(l.opts.Fields, fields)
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	return f
}

func (l *zapLogger) fieldsMerge(parent, src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(parent)+len(src))
	for k, v := range parent {
		dst[k] = v
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
