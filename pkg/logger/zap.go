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
	sentry_core "github.com/lastbackend/toolkit/pkg/logger/sentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"os"
	"sync"
)

func init() {
	lvl, err := GetLevel(os.Getenv("TOOLKIT_LOG_LEVEL"))
	if err != nil {
		lvl = InfoLevel
	}
	verbose, err := GetLevel(os.Getenv("TOOLKIT_VERBOSE_LOG_LEVEL"))
	if err != nil {
		lvl = ErrorLevel
	}

	DefaultLogger = newZapLogger(Options{Level: lvl, VerboseLevel: verbose})
}

type zapLogger struct {
	sync.RWMutex
	logger *zap.SugaredLogger
	opts   Options
}

func newZapLogger(opts Options) *zapLogger {
	cores := make([]zapcore.Core, 0)
	combinedCore := zapcore.NewTee(cores...)
	l := &zapLogger{
		logger: zap.New(combinedCore).Sugar(),
		opts: Options{
			Level:           InfoLevel,
			VerboseLevel:    ErrorLevel,
			Fields:          make(map[string]interface{}, 0),
			Out:             os.Stdout,
			CallerSkipCount: 1,
			JSONFormat:      false,
		},
	}
	l.init(opts)
	return l
}

func (l *zapLogger) Init(opts Options) Logger {
	l.opts.JSONFormat = opts.JSONFormat
	l.opts.SentryDNS = opts.SentryDNS
	if opts.Level != InfoLevel {
		l.opts.Level = opts.Level
	}
	if opts.VerboseLevel != InfoLevel {
		l.opts.VerboseLevel = opts.VerboseLevel
	}
	if opts.Fields != nil {
		l.opts.Fields = opts.Fields
	}
	if opts.Out != nil {
		l.opts.Out = opts.Out
	}
	if opts.CallerSkipCount > 0 {
		l.opts.CallerSkipCount = opts.CallerSkipCount
	}
	return l.init(opts)
}

func (l *zapLogger) WithFields(fields Fields) Logger {
	return &zapLogger{logger: l.logger.With(l.getFields(fields)...)}
}

func (l *zapLogger) Options() Options {
	return l.opts
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

func (l *zapLogger) init(opts Options) Logger {
	zapCores := make([]zapcore.Core, 0)
	zapOptions := make([]zap.Option, 0)

	encoderConfig := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevelAt(getZapLevel(l.opts.Level))

	writer := zapcore.Lock(zapcore.NewMultiWriteSyncer(zapcore.AddSync(l.opts.Out)))
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), writer, level)

	if opts.JSONFormat {
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), writer, level)
	}

	zapCores = append(zapCores, core)
	zapOptions = append(zapOptions, zap.AddCallerSkip(l.opts.CallerSkipCount))
	zapOptions = append(zapOptions, zap.AddCaller())

	log := zap.New(zapcore.NewTee(zapCores...), zapOptions...)

	// Enable sentry
	if l.opts.SentryDNS != "" {
		cfg := sentry_core.Configuration{
			Level:             zapcore.ErrorLevel,
			EnableBreadcrumbs: true,
		}
		if opts.Tags != nil {
			cfg.Tags = opts.Tags
		}
		sentryCore, err := sentry_core.NewCore(cfg, sentry_core.NewSentryClientFromDSN(l.opts.SentryDNS))
		if err != nil {
			panic(err)
		}
		log = sentry_core.AttachCoreToLogger(sentryCore, log)
		log = log.With(sentry_core.NewScope())
	}

	l.logger = log.Sugar()

	return l
}

func getZapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case PanicLevel:
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func (l *zapLogger) getFields(fields Fields) []interface{} {
	fields = fieldsMerge(l.opts.Fields, fields)
	var f = make([]interface{}, 0)
	for k, v := range fields {
		f = append(f, k)
		f = append(f, v)
	}
	return f
}
