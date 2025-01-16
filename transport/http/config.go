package http

type TransportConfig struct {
	ServiceName         string
	ExternalServiceName string

	MaxRetries      int
	BackoffDelaysMs int
}
