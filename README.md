# zerowrap

A reusable Go package that wraps [zerolog](https://github.com/rs/zerolog) with context-based logger access and maximum abstraction.

## Installation

```bash
go get github.com/bnema/zerowrap
```

For OpenTelemetry integration (optional):

```bash
go get github.com/bnema/zerowrap/otel
```

## Quick Start

```go
package main

import (
    "context"
    "github.com/bnema/zerowrap"
)

func main() {
    // Create a logger
    log := zerowrap.New(zerowrap.Config{
        Level:  "debug",
        Format: "console",
    })

    // Attach to context
    ctx := zerowrap.WithCtx(context.Background(), log)

    // Use throughout your application
    zerowrap.FromCtx(ctx).Info().Msg("hello world")

    // Pass context to other functions
    doSomething(ctx)
}

func doSomething(ctx context.Context) {
    zerowrap.FromCtx(ctx).Debug().Msg("doing something")
}
```

## Features

- Context-based logger storage and retrieval
- Add fields to loggers (single, multiple, or from structs)
- Configurable logger creation with sensible defaults
- File-based logging with rotation support (via lumberjack)
- OpenTelemetry log bridging (optional sub-package)
- Common field name constants for consistency

## API Reference

### Context Functions

| Function | Description |
|----------|-------------|
| `FromCtx(ctx)` | Extract logger from context (returns no-op if none) |
| `Ctx(ctx)` | Get pointer to logger in context |
| `WithCtx(ctx, log)` | Attach logger to context |

### Field Helpers

| Function | Description |
|----------|-------------|
| `FromCtxWithField(ctx, key, value)` | Get logger with one additional field |
| `FromCtxWithFields(ctx, fields)` | Get logger with multiple fields |
| `FromCtxWithStruct(ctx, struct)` | Get logger with fields from struct tags |
| `CtxWithField(ctx, key, value)` | Get new context with enriched logger |
| `CtxWithFields(ctx, fields)` | Get new context with enriched logger |
| `CtxWithStruct(ctx, struct)` | Get new context with enriched logger |

### Logger Creation

| Function | Description |
|----------|-------------|
| `New(cfg)` | Create logger with configuration |
| `NewFromEnv(prefix)` | Create logger from environment variables |
| `NewWithFile(cfg, fileCfg)` | Create logger with file output |
| `Default()` | Create default logger (info level, console format) |
| `WithHook(log, hook)` | Add hook to logger |

## Usage Examples

### Basic Logging with Context

```go
log := zerowrap.Default()
ctx := zerowrap.WithCtx(context.Background(), log)

zerowrap.FromCtx(ctx).Info().Str("action", "start").Msg("application started")
```

### Adding Fields

```go
// Single field
log := zerowrap.FromCtxWithField(ctx, "user_id", 123)
log.Info().Msg("user action")

// Multiple fields
log := zerowrap.FromCtxWithFields(ctx, map[string]any{
    "user_id":    123,
    "request_id": "abc-123",
    "ip":         "192.168.1.1",
})
log.Info().Msg("request received")

// Enrich context for downstream use
ctx = zerowrap.CtxWithField(ctx, zerowrap.FieldComponent, "auth")
zerowrap.FromCtx(ctx).Info().Msg("authenticating") // includes component=auth
```

### Struct Tags

Extract fields from structs using the `log` tag (falls back to `json` tag, then field name):

```go
type RequestInfo struct {
    UserID    int    `log:"user_id"`
    RequestID string `log:"request_id"`
    IP        string `json:"ip_address"`
    Internal  string `log:"-"` // skipped
}

info := RequestInfo{
    UserID:    123,
    RequestID: "abc-123",
    IP:        "192.168.1.1",
    Internal:  "secret",
}

log := zerowrap.FromCtxWithStruct(ctx, info)
log.Info().Msg("request info")
// Output includes: user_id=123 request_id=abc-123 ip_address=192.168.1.1
```

### Logger Configuration

```go
log := zerowrap.New(zerowrap.Config{
    Level:      "debug",           // trace, debug, info, warn, error, fatal, panic
    Format:     "console",         // console or json
    TimeFormat: time.RFC3339,      // custom time format
    Output:     os.Stdout,         // custom output writer
    Caller:     true,              // include caller info (file:line)
})
```

### Environment Variables

```go
// Reads MYAPP_LOG_LEVEL and MYAPP_LOG_FORMAT
log := zerowrap.NewFromEnv("MYAPP")
```

### File Logging

```go
log, cleanup, err := zerowrap.NewWithFile(
    zerowrap.Config{
        Level:  "info",
        Format: "console",
    },
    zerowrap.FileConfig{
        Enabled:    true,
        Path:       "/var/log/myapp/app.log",
        MaxSize:    100,  // MB
        MaxBackups: 3,
        MaxAge:     28,   // days
        Compress:   true,
    },
)
if err != nil {
    panic(err)
}
defer cleanup()

ctx := zerowrap.WithCtx(context.Background(), log)
// Logs go to both console (formatted) and file (JSON)
```

### OpenTelemetry Integration

```go
import (
    "github.com/bnema/zerowrap"
    "github.com/bnema/zerowrap/otel"
)

// Using global provider
log := zerowrap.New(cfg).Hook(otel.NewHook("my-service"))

// Using custom provider
provider := // your OTel logger provider
log := zerowrap.New(cfg).Hook(otel.NewHookWithProvider(provider, "my-service"))

ctx := zerowrap.WithCtx(context.Background(), log)
// Logs now flow to both zerolog output AND OpenTelemetry
```

## Field Constants

Common field names for consistency across your application:

```go
zerowrap.FieldComponent  // "component"
zerowrap.FieldRequestID  // "request_id"
zerowrap.FieldTraceID    // "trace_id"
zerowrap.FieldSpanID     // "span_id"
zerowrap.FieldUserID     // "user_id"
zerowrap.FieldError      // "error"
zerowrap.FieldDuration   // "duration_ms"
```

Usage:

```go
ctx = zerowrap.CtxWithField(ctx, zerowrap.FieldComponent, "database")
ctx = zerowrap.CtxWithField(ctx, zerowrap.FieldRequestID, requestID)
```

## License

MIT License - see [LICENSE](LICENSE) file.
