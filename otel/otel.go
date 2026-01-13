package otel

import (
	"context"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

// Hook is a zerolog.Hook that bridges logs to OpenTelemetry.
type Hook struct {
	logger log.Logger
}

// NewHook creates a hook that forwards zerolog events to OpenTelemetry.
// Uses the global logger provider.
func NewHook(serviceName string) *Hook {
	return &Hook{
		logger: global.GetLoggerProvider().Logger(serviceName),
	}
}

// NewHookWithProvider creates a hook with a specific logger provider.
func NewHookWithProvider(provider log.LoggerProvider, serviceName string) *Hook {
	return &Hook{
		logger: provider.Logger(serviceName),
	}
}

// Run implements zerolog.Hook interface.
// It forwards log events to the OpenTelemetry logger.
func (h *Hook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if h.logger == nil {
		return
	}

	ctx := e.GetCtx()
	if ctx == nil {
		ctx = context.Background()
	}

	var record log.Record
	record.SetBody(log.StringValue(msg))
	record.SetSeverity(levelToOTel(level))
	record.SetSeverityText(level.String())

	h.logger.Emit(ctx, record)
}

// levelToOTel converts zerolog.Level to OpenTelemetry log.Severity.
func levelToOTel(level zerolog.Level) log.Severity {
	switch level {
	case zerolog.TraceLevel:
		return log.SeverityTrace
	case zerolog.DebugLevel:
		return log.SeverityDebug
	case zerolog.InfoLevel:
		return log.SeverityInfo
	case zerolog.WarnLevel:
		return log.SeverityWarn
	case zerolog.ErrorLevel:
		return log.SeverityError
	case zerolog.FatalLevel:
		return log.SeverityFatal
	case zerolog.PanicLevel:
		return log.SeverityFatal4
	default:
		return log.SeverityInfo
	}
}
