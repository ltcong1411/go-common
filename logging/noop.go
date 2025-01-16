package logging

import "go.uber.org/zap"

// nopLogger is a no-op logger.
type nopLogger struct {
	*zap.SugaredLogger
}

// NewNopLogger returns a new no-op logger.
func NewNopLogger() Logger {
	return &nopLogger{
		zap.NewNop().Sugar(),
	}
}

// With returns a new no-op logger.
func (l *nopLogger) With(args ...interface{}) Logger {
	return l
}
