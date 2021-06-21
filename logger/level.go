/*
Copyright [2014] - [2021] The Last.Backend authors.

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
	"fmt"
	"os"
)

type Level int8

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	}
	return ""
}

func GetLevel(lvl string) (Level, error) {
	switch lvl {
	case DebugLevel.String():
		return DebugLevel, nil
	case InfoLevel.String():
		return InfoLevel, nil
	case WarnLevel.String():
		return WarnLevel, nil
	case ErrorLevel.String():
		return ErrorLevel, nil
	case PanicLevel.String():
		return PanicLevel, nil
	case FatalLevel.String():
		return FatalLevel, nil
	}
	return InfoLevel, fmt.Errorf("unknown log level: '%s', defaulting to InfoLevel", lvl)
}

func WithFields(fields Fields) Logger {
	return DefaultLogger.WithFields(fields)
}

func Debug(args ...interface{}) {
	DefaultLogger.Debug(args)
}

func Debugf(format string, args ...interface{}) {
	DefaultLogger.Debugf(format, args)
}

func Info(args ...interface{}) {
	DefaultLogger.Info(args)
}

func Infof(format string, args ...interface{}) {
	DefaultLogger.Infof(format, args)
}

func Warn(args ...interface{}) {
	DefaultLogger.Warn(args)
}

func Warnf(format string, args ...interface{}) {
	DefaultLogger.Warnf(format, args)
}

func Error(args ...interface{}) {
	DefaultLogger.Error(args)
}

func Errorf(format string, args ...interface{}) {
	DefaultLogger.Errorf(format, args)
}

func Panic(args ...interface{}) {
	DefaultLogger.Panic(args)
	os.Exit(1)
}

func Panicf(format string, args ...interface{}) {
	DefaultLogger.Panicf(format, args)
	os.Exit(1)
}

func Fatal(args ...interface{}) {
	DefaultLogger.Fatal(args)
	os.Exit(1)
}

func Fatalf(format string, args ...interface{}) {
	DefaultLogger.Fatalf(format, args)
	os.Exit(1)
}

func V(lvl Level, log Logger) bool {
	l := DefaultLogger
	if log != nil {
		l = log
	}
	return l.Options().VerboseLevel <= lvl
}
