package statsd

import (
	"strings"
	"time"
)

// Option represents a functional option for configuring the StatsD client.
type Option func(*options)

// Host sets the StatsD server hostname or IP address in the client configuration.
func Host(host string) Option {
	return func(o *options) {
		o.host = host
	}
}

// Port sets the UDP port for connecting to the StatsD server.
func Port(port int) Option {
	return func(o *options) {
		o.port = port
	}
}

// MaxBufferSize sets the maximum buffer size for metrics before triggering a flush.
func MaxBufferSize(maxBufferSize int) Option {
	return func(o *options) {
		o.maxBufferSize = maxBufferSize
	}
}

// FlushInterval sets the time interval between automatic flushes of buffered metrics to StatsD.
func FlushInterval(flushInterval time.Duration) Option {
	return func(o *options) {
		o.flushInterval = flushInterval
	}
}

// ErrorHandler sets a custom error handling function, which is called when there are errors in sending metrics.
func ErrorHandler(errorHandler func(error)) Option {
	return func(o *options) {
		o.errorHandler = errorHandler
	}
}

// Prefix sets an optional prefix for all metric names to distinguish them or group them logically.
// A trailing dot is added to the prefix if it does not already exist.
func Prefix(prefix string) Option {
	return func(o *options) {
		o.prefix = strings.TrimSuffix(prefix, ".") + "."
	}
}

// Tags sets a default set of tags to be included with every metric sent by the client.
func Tags(tags []Tag) Option {
	return func(o *options) {
		o.tags = tags
	}
}
