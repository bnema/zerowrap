package zerowrap

import (
	"fmt"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with additional convenience methods.
// All zerolog.Logger methods are available via embedding.
type Logger struct {
	zerolog.Logger
}

// WrapErr logs the error and returns a wrapped error with the message.
// Uses fmt.Errorf with %w for unwrapping support.
//
//	log := zerowrap.FromCtx(ctx)
//	if err != nil {
//	    return log.WrapErr(err, "failed to connect")
//	}
func (l Logger) WrapErr(err error, msg string) error {
	if err == nil {
		return nil
	}
	l.Error().Err(err).Msg(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErrWithFields logs with fields and returns a wrapped error.
//
//	if err != nil {
//	    return log.WrapErrWithFields(err, "query failed", map[string]any{"id": id})
//	}
func (l Logger) WrapErrWithFields(err error, msg string, fields map[string]any) error {
	if err == nil {
		return nil
	}
	c := l.With()
	for k, v := range fields {
		c = addToContext(c, k, v)
	}
	logger := c.Logger()
	logger.Error().Err(err).Msg(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

// WrapErrf logs the error and returns a wrapped error with a formatted message.
//
//	if err != nil {
//	    return log.WrapErrf(err, "failed to connect to %s", host)
//	}
func (l Logger) WrapErrf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	l.Error().Err(err).Msg(msg)
	return fmt.Errorf("%s: %w", msg, err)
}

// WithField returns a new Logger with the field added.
func (l Logger) WithField(key string, value any) Logger {
	return Logger{addToContext(l.With(), key, value).Logger()}
}

// WithFields returns a new Logger with the fields added.
func (l Logger) WithFields(fields map[string]any) Logger {
	c := l.With()
	for k, v := range fields {
		c = addToContext(c, k, v)
	}
	return Logger{c.Logger()}
}

// WithStruct returns a new Logger with fields extracted from struct tags.
func (l Logger) WithStruct(s any) Logger {
	return l.WithFields(extractFields(s))
}
