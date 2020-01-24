package jsonrpc

import "time"

type ReconnectConfig struct {
	Reconnect    bool
	WaitInterval time.Duration
	MaxRetries   int
}

type Config struct {
	Reconnect ReconnectConfig
}

var defaultConfig = Config{
	Reconnect: ReconnectConfig{
		Reconnect:    false,
		WaitInterval: time.Second * 5,
		MaxRetries:   5,
	},
}

type Option func(c *Config)

// WithReconnect configures automatic connection setup if lost.
// If maxRetries is zero, will retry forever.
func WithReconnect(enabled bool, wi time.Duration, maxRetries int) func(*Config) {
	return func(c *Config) {
		c.Reconnect.Reconnect = enabled
		c.Reconnect.WaitInterval = wi
		c.Reconnect.MaxRetries = maxRetries
	}
}
