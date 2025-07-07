package config

import (
	"context"
)

// contextKey is the type for context keys
type contextKey string

// configKey is the context key for storing configuration
const configKey contextKey = "config"

// WithConfig adds configuration to context
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, configKey, cfg)
}

// FromContext retrieves configuration from context
func FromContext(ctx context.Context) (*Config, bool) {
	if ctx == nil {
		return nil, false
	}
	cfg, ok := ctx.Value(configKey).(*Config)
	return cfg, ok
}

// MustFromContext retrieves configuration from context or panics
func MustFromContext(ctx context.Context) *Config {
	cfg, ok := FromContext(ctx)
	if !ok {
		panic("configuration not found in context")
	}
	return cfg
}