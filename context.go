package zerowrap

import (
	"context"

	"github.com/rs/zerolog"
)

// FromCtx extracts the logger from context.
// If no logger is found, returns a disabled (no-op) logger.
func FromCtx(ctx context.Context) zerolog.Logger {
	return *zerolog.Ctx(ctx)
}

// Ctx returns a pointer to the logger in context.
// This is for compatibility with zerolog's Ctx pattern.
// If no logger is found, returns a pointer to a disabled logger.
func Ctx(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}

// WithCtx attaches the logger to the context and returns the new context.
func WithCtx(ctx context.Context, log zerolog.Logger) context.Context {
	return log.WithContext(ctx)
}
