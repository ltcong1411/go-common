package http

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/ltcong1411/go-common/logging"
)

// doFunc is an executable function which will return http status code and the error
type doFunc func(ctx context.Context) (int, error)

func (t *httpClientImpl) retry(ctx context.Context, endpoint, method string, doFunc doFunc, opts ...Options) error {
	options := &requestOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
	}
	if options.retryConfig == (RetryConfig{}) { // empty
		options.retryConfig.MaxRetries = t.cfg.MaxRetries
		options.retryConfig.DelayMs = t.cfg.BackoffDelaysMs
	}

	err := retry.Do(func() error {
		var err error
		_, err = doFunc(ctx)
		if err != nil {
			logger := logging.FromContext(ctx)
			logger.Warnw(fmt.Sprintf("[%s] %s: inner attempt failed", method, endpoint), "error", err)
		}
		return err
	},
		retry.Attempts(uint(options.retryConfig.MaxRetries)),
		retry.Delay(time.Duration(options.retryConfig.DelayMs)*time.Millisecond),
		retry.LastErrorOnly(true),
	)
	return err
}
