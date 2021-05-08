package logger

import (
	"context"
)

var (
	DefaultLogger Logger = NewLogger()
)

const (
	InstanceZapLogger int = iota
)

//Fields Type to pass when we want to call WithFields for structured logging
type Fields map[string]interface{}

//Logger is our contract for the logger
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	WithFields(keyValues Fields) Logger
}

type logger struct {
	l Logger
}

//NewLogger returns an instance of logger
func NewLogger(opts ...Option) *logger {

	// Default options
	options := Options{
		EnableConsole:     false,
		ConsoleJSONFormat: false,
		ConsoleLevel:      InfoLevel,
	}

	for _, o := range opts {
		o(&options)
	}

	return &logger{l: newZapLogger(options)}
}

func (l logger) Debug(args ...interface{}) {
	l.l.Debug(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.l.Debugf(format, args...)
}

func (l logger) Info(args ...interface{}) {
	l.l.Info(args...)
}

func (l logger) Infof(format string, args ...interface{}) {
	l.l.Infof(format, args...)
}

func (l logger) Warn(args ...interface{}) {
	l.l.Warn(args...)
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.l.Warnf(format, args...)
}

func (l logger) Error(args ...interface{}) {
	l.l.Error(args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.l.Errorf(format, args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.l.Fatal(args...)
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.l.Fatalf(format, args...)
}

func (l logger) Panic(args ...interface{}) {
	l.l.Panic(args...)
}

func (l logger) Panicf(format string, args ...interface{}) {
	l.l.Panicf(format, args...)
}

func (l logger) WithFields(keyValues Fields) Logger {
	return l.l.WithFields(keyValues)
}

func (l logger) WithContext(ctx context.Context) Logger {

	if ctx == nil {
		return l.l
	}

	if ctxLogger, ok := ctx.Value(InstanceZapLogger).(Logger); ok {
		return ctxLogger
	}

	return l.l
}
