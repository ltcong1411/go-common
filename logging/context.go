package logging

import (
	"context"
	"sync"
)

type contextKey struct{}

var (
	mu            = sync.Mutex{}
	defaultLogger = NewNopLogger()
)

// NewContext returns a new Context with given logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns the logger associated with a context.
// It returns the default Logger if no Logger exists.
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(contextKey{}).(Logger); ok {
		return logger
	}
	return defaultLogger
}

// SetDefaultLogger sets the default logger.
func SetDefaultLogger(logger Logger) {
	mu.Lock()
	defer mu.Unlock()

	defaultLogger = logger
}

// With returns a new contextual Logger with additional context args.
func With(logger Logger, args ...interface{}) Logger {
	return logger.With(args...)
}
