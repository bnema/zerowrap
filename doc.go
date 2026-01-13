// Package zerowrap provides a reusable wrapper around zerolog for context-based logging.
//
// It simplifies common logging patterns by providing convenient functions for:
//   - Storing and retrieving loggers from context
//   - Adding fields to loggers (single, multiple, or from structs)
//   - Creating configured loggers with sensible defaults
//   - File-based logging with rotation support
//
// # Basic Usage
//
//	// Create a logger
//	log := zerowrap.New(zerowrap.Config{
//	    Level:  "debug",
//	    Format: "console",
//	})
//
//	// Attach to context
//	ctx := zerowrap.WithCtx(context.Background(), log)
//
//	// Use throughout your application
//	zerowrap.FromCtx(ctx).Info().Msg("hello world")
//
// # Adding Fields
//
//	// Single field
//	log := zerowrap.FromCtxWithField(ctx, "user_id", 123)
//
//	// Multiple fields
//	log := zerowrap.FromCtxWithFields(ctx, map[string]any{
//	    "user_id": 123,
//	    "request_id": "abc",
//	})
//
//	// From struct with tags
//	type Request struct {
//	    UserID    int    `log:"user_id"`
//	    RequestID string `log:"request_id"`
//	}
//	log := zerowrap.FromCtxWithStruct(ctx, Request{UserID: 123, RequestID: "abc"})
//
// # Field Constants
//
// The package provides common field name constants for consistency:
//
//	zerowrap.FieldComponent  // "component"
//	zerowrap.FieldRequestID  // "request_id"
//	zerowrap.FieldTraceID    // "trace_id"
//	zerowrap.FieldSpanID     // "span_id"
//	zerowrap.FieldUserID     // "user_id"
//	zerowrap.FieldError      // "error"
//	zerowrap.FieldDuration   // "duration_ms"
//
// # OpenTelemetry Integration
//
// For OpenTelemetry log bridging, use the optional otel sub-package:
//
//	import "github.com/bnema/zerowrap/otel"
//
//	log := zerowrap.New(cfg).Hook(otel.NewOTelHook("my-service"))
package zerowrap
