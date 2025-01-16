package logging

import (
	"fmt"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ltcong1411/go-common/config"
)

type zapDriverLogger struct {
	*zap.SugaredLogger
}

// NewZapDriverLogger returns a new instance of Logger that uses Zap & Stack Driver.
func NewZapDriverLogger(cp config.Provider) Logger {
	conf := GetConfig(cp)

	var zapConfig zap.Config
	if conf.IsDevelopment {
		// running in development mode we will use a human-readable output
		zapConfig = zapdriver.NewDevelopmentConfig()
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zapdriver.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	var l zapcore.Level
	if err := l.UnmarshalText([]byte(conf.Level)); err != nil {
		l = zap.DebugLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(l)

	clientLogger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("zap.config.Build(): %v", err))
	}

	return &zapDriverLogger{
		clientLogger.Sugar(),
	}
}

// With returns a new logger with given arguments.
func (l *zapDriverLogger) With(args ...interface{}) Logger {
	return &zapDriverLogger{
		l.SugaredLogger.With(args...),
	}
}

// AsZapLogger converts a Logger to zap.Logger if possible.
func AsZapLogger(l Logger) (*zap.Logger, bool) {
	if zl, ok := l.(*zapDriverLogger); ok {
		return zl.SugaredLogger.Desugar(), true
	} else {
		return zap.NewNop(), false
	}
}
