package http

type RetryConfig struct {
	MaxRetries int
	DelayMs    int
}

type requestOptions struct {
	retryConfig RetryConfig
}

type Options func(client *requestOptions)

func WithRetryConfig(retryCfg RetryConfig) func(options *requestOptions) {
	return func(options *requestOptions) {
		options.retryConfig = retryCfg
	}
}
