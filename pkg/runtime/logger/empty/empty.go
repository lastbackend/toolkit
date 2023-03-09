package empty

import (
	"github.com/lastbackend/toolkit/pkg/runtime/logger"
	"go.uber.org/fx/fxevent"
)

type emptyLogger struct {
	logger.Logger
}

func (l *emptyLogger) WithFields(_ logger.Fields) logger.Logger {
	return l
}

func (l *emptyLogger) Init(_ logger.Options) logger.Logger {
	return l
}

func (l *emptyLogger) Options() logger.Options {
	return logger.Options{}
}

func (l *emptyLogger) Debug(_ ...interface{})            {}
func (l *emptyLogger) Debugf(_ string, _ ...interface{}) {}
func (l *emptyLogger) Info(_ ...interface{})             {}
func (l *emptyLogger) Infof(_ string, _ ...interface{})  {}
func (l *emptyLogger) Warn(_ ...interface{})             {}
func (l *emptyLogger) Warnf(_ string, _ ...interface{})  {}
func (l *emptyLogger) Error(_ ...interface{})            {}
func (l *emptyLogger) Errorf(_ string, _ ...interface{}) {}
func (l *emptyLogger) Panic(_ ...interface{})            {}
func (l *emptyLogger) Panicf(_ string, _ ...interface{}) {}
func (l *emptyLogger) Fatal(_ ...interface{})            {}
func (l *emptyLogger) Fatalf(_ string, _ ...interface{}) {}
func (l *emptyLogger) V(logger.Level) logger.Logger {
	return l
}

func (l *emptyLogger) Inject(_ func(level logger.Level)) {

}

func (l *emptyLogger) Fx() fxevent.Logger {
	return fxevent.NopLogger
}

func NewLogger() logger.Logger {
	return new(emptyLogger)
}
