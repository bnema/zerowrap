// Package zerowrap provides a reusable wrapper around zerolog for context-based logging.
//
// It simplifies common logging patterns by providing:
//   - Context-based logger storage and retrieval
//   - A Logger type with error wrapping helpers
//   - Field enrichment (single, multiple, or from structs)
//   - Configurable logger creation with sensible defaults
//   - File-based logging with rotation support
//   - OpenTelemetry integration (optional sub-package)
//
// # Logger Type
//
// The Logger type wraps zerolog.Logger and provides additional convenience methods.
// All zerolog.Logger methods are available via embedding:
//
//	type Logger struct {
//	    zerolog.Logger
//	}
//
// Logger methods:
//
//	log.WrapErr(err, msg) error           // Log and wrap error
//	log.WrapErrWithFields(err, msg, fields) error  // Log with fields and wrap
//	log.WrapErrf(err, format, args...) error       // Log and wrap with formatted message
//	log.WithField(key, value) Logger      // Return logger with added field
//	log.WithFields(fields) Logger         // Return logger with added fields
//	log.WithStruct(s) Logger              // Return logger with fields from struct
//
// # Quick Start
//
//	// Create and attach logger to context
//	logger := zerowrap.New(zerowrap.Config{
//	    Level:  "debug",
//	    Format: "console",
//	})
//	ctx := zerowrap.WithCtx(context.Background(), logger)
//
//	// Use throughout your application
//	log := zerowrap.FromCtx(ctx)
//	log.Info().Msg("hello world")
//
// # Context Functions
//
// Store and retrieve loggers from context:
//
//	FromCtx(ctx) Logger                   // Get logger from context (no-op if none)
//	Ctx(ctx) *zerolog.Logger              // Get pointer to underlying zerolog.Logger
//	WithCtx(ctx, log) context.Context     // Attach logger to context
//	WithCtxZerolog(ctx, log) context.Context  // Attach zerolog.Logger to context
//
// # Field Helpers
//
// Get logger with additional fields:
//
//	FromCtxWithField(ctx, key, value) Logger      // One field
//	FromCtxWithFields(ctx, fields) Logger         // Multiple fields
//	FromCtxWithStruct(ctx, s) Logger              // Fields from struct tags
//
// Get new context with enriched logger:
//
//	CtxWithField(ctx, key, value) context.Context
//	CtxWithFields(ctx, fields) context.Context
//	CtxWithStruct(ctx, s) context.Context
//
// # Struct Tags
//
// Extract fields from structs using the `log` tag (falls back to `json`, then field name):
//
//	type Request struct {
//	    UserID    int    `log:"user_id"`
//	    RequestID string `log:"request_id"`
//	    IP        string `json:"ip_address"`
//	    Internal  string `log:"-"`  // skipped
//	}
//
//	log := zerowrap.FromCtxWithStruct(ctx, Request{UserID: 123, RequestID: "abc"})
//
// # Logger Creation
//
// Create loggers with configuration:
//
//	New(cfg Config) Logger                        // Create with config
//	NewFromEnv(prefix string) Logger              // Create from env vars
//	NewWithFile(cfg, fileCfg) (Logger, func(), error)  // Create with file output
//	Default() Logger                              // Default logger (info, console)
//	WithHook(log, hook) Logger                    // Add hook to logger
//
// # Config
//
// Configuration for logger creation:
//
//	type Config struct {
//	    Level      string     // trace, debug, info, warn, error, fatal, panic
//	    Format     string     // json or console
//	    TimeFormat string     // time format (default: time.RFC3339)
//	    Output     io.Writer  // output writer (default: os.Stderr)
//	    Caller     bool       // include caller info (file:line)
//	}
//
// # FileConfig
//
// Configuration for file-based logging with rotation:
//
//	type FileConfig struct {
//	    Enabled    bool    // toggle file logging
//	    Path       string  // log file path
//	    MaxSize    int     // max size in MB before rotation (default: 100)
//	    MaxBackups int     // max old files to retain (default: 3)
//	    MaxAge     int     // max days to retain (default: 28)
//	    Compress   bool    // compress rotated files
//	}
//
// # Error Helpers
//
// Log and return wrapped errors in one line:
//
//	func doSomething(ctx context.Context) error {
//	    log := zerowrap.FromCtx(ctx)
//
//	    // Simple wrap
//	    if err != nil {
//	        return log.WrapErr(err, "failed to connect")
//	    }
//
//	    // With fields
//	    if err != nil {
//	        return log.WrapErrWithFields(err, "query failed", map[string]any{
//	            "table": tableName,
//	        })
//	    }
//
//	    // With formatted message
//	    if err != nil {
//	        return log.WrapErrf(err, "failed to connect to %s", host)
//	    }
//	}
//
// # Field Constants
//
// Common field names for consistency:
//
//	// Identity & Tracing
//	FieldComponent, FieldRequestID, FieldTraceID, FieldSpanID
//	FieldCorrelationID, FieldSessionID, FieldUserID
//
//	// HTTP/API
//	FieldMethod, FieldPath, FieldStatus, FieldClientIP
//
//	// Service/Infra
//	FieldService, FieldVersion, FieldHost, FieldEnv
//
//	// Operations
//	FieldAction, FieldOperation, FieldError, FieldDuration
//
//	// Data
//	FieldCount, FieldSize
//
//	// Clean Architecture - Layers
//	FieldLayer, FieldUseCase
//
//	// Clean Architecture - Adapters
//	FieldAdapter, FieldAdapterType, FieldHandler, FieldRepository, FieldGateway
//
//	// Database/Storage
//	FieldTable, FieldQuery, FieldDatabase
//
//	// Messaging/Events
//	FieldEvent, FieldTopic, FieldQueue, FieldPayload
//
//	// Entity/Resource
//	FieldEntity, FieldEntityID, FieldEntityType
//
// Usage:
//
//	ctx = zerowrap.CtxWithField(ctx, zerowrap.FieldComponent, "database")
//	ctx = zerowrap.CtxWithFields(ctx, map[string]any{
//	    zerowrap.FieldRequestID: requestID,
//	    zerowrap.FieldUserID:    userID,
//	})
//
// # Environment Variables
//
// Create logger from environment variables:
//
//	// Reads MYAPP_LOG_LEVEL and MYAPP_LOG_FORMAT
//	log := zerowrap.NewFromEnv("MYAPP")
//
// # File Logging
//
// Create logger with file output and rotation:
//
//	log, cleanup, err := zerowrap.NewWithFile(
//	    zerowrap.Config{Level: "info", Format: "console"},
//	    zerowrap.FileConfig{
//	        Enabled:    true,
//	        Path:       "/var/log/app.log",
//	        MaxSize:    100,  // MB
//	        MaxBackups: 3,
//	        MaxAge:     28,   // days
//	        Compress:   true,
//	    },
//	)
//	if err != nil {
//	    panic(err)
//	}
//	defer cleanup()
//
// # OpenTelemetry Integration
//
// For OpenTelemetry log bridging, use the optional otel sub-package:
//
//	import "github.com/bnema/zerowrap/otel"
//
//	// Using global provider
//	log := zerowrap.New(cfg).Hook(otel.NewHook("my-service"))
//
//	// Using custom provider
//	log := zerowrap.New(cfg).Hook(otel.NewHookWithProvider(provider, "my-service"))
//
// # Field Propagation Pattern
//
// The key pattern is to enrich the context with fields EARLY (at request entry points),
// so all downstream code automatically includes those fields in logs:
//
//	// Middleware: add request-scoped fields once
//	func RequestMiddleware(next http.Handler) http.Handler {
//	    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        ctx := r.Context()
//
//	        // Define fields early - they propagate to ALL downstream logs
//	        ctx = zerowrap.CtxWithFields(ctx, map[string]any{
//	            zerowrap.FieldRequestID: uuid.New().String(),
//	            zerowrap.FieldMethod:    r.Method,
//	            zerowrap.FieldPath:      r.URL.Path,
//	            zerowrap.FieldClientIP:  r.RemoteAddr,
//	        })
//
//	        next.ServeHTTP(w, r.WithContext(ctx))
//	    })
//	}
//
//	// Handler: add user-specific fields after auth
//	func UserHandler(ctx context.Context, userID string) error {
//	    // Add user field - now ALL downstream logs include user_id
//	    ctx = zerowrap.CtxWithField(ctx, zerowrap.FieldUserID, userID)
//
//	    // This log has: request_id, method, path, client_ip, user_id
//	    zerowrap.FromCtx(ctx).Info().Msg("user authenticated")
//
//	    return processUser(ctx)  // ctx carries all fields downstream
//	}
//
//	// Service layer: just use the context, fields are already there
//	func processUser(ctx context.Context) error {
//	    log := zerowrap.FromCtx(ctx)
//
//	    // This log automatically has ALL parent fields
//	    log.Debug().Msg("processing user")
//
//	    if err := db.Query(ctx); err != nil {
//	        // Error log includes all fields: request_id, method, path, client_ip, user_id
//	        return log.WrapErr(err, "database query failed")
//	    }
//	    return nil
//	}
//
// # Complete Example
//
//	func main() {
//	    // Create logger with service-level fields
//	    logger := zerowrap.New(zerowrap.Config{
//	        Level:  "debug",
//	        Format: "console",
//	    })
//	    ctx := zerowrap.WithCtx(context.Background(), logger)
//
//	    // Service-level fields: present in ALL logs
//	    ctx = zerowrap.CtxWithFields(ctx, map[string]any{
//	        zerowrap.FieldService: "my-api",
//	        zerowrap.FieldVersion: "1.0.0",
//	        zerowrap.FieldEnv:     "production",
//	    })
//
//	    // Start server with enriched context
//	    server.Start(ctx)
//	}
//
//	func HandleRequest(ctx context.Context, req *Request) error {
//	    // Request-level fields: added at entry point
//	    ctx = zerowrap.CtxWithFields(ctx, map[string]any{
//	        zerowrap.FieldRequestID: req.ID,
//	        zerowrap.FieldUserID:    req.UserID,
//	        zerowrap.FieldPath:      req.Path,
//	    })
//
//	    log := zerowrap.FromCtx(ctx)
//	    log.Info().Msg("request started")
//
//	    // All downstream calls use same enriched context
//	    if err := validateRequest(ctx, req); err != nil {
//	        return log.WrapErr(err, "validation failed")
//	    }
//
//	    if err := processRequest(ctx, req); err != nil {
//	        return log.WrapErr(err, "processing failed")
//	    }
//
//	    log.Info().Msg("request completed")
//	    return nil
//	}
//
//	func validateRequest(ctx context.Context, req *Request) error {
//	    log := zerowrap.FromCtx(ctx)
//	    // Log automatically includes: service, version, env, request_id, user_id, path
//	    log.Debug().Msg("validating request")
//	    return nil
//	}
//
//	func processRequest(ctx context.Context, req *Request) error {
//	    log := zerowrap.FromCtx(ctx)
//	    // Same fields propagated here too
//	    log.Debug().Msg("processing request")
//
//	    if err := callExternalAPI(ctx); err != nil {
//	        // Error includes all context fields for easy debugging
//	        return log.WrapErr(err, "external API call failed")
//	    }
//	    return nil
//	}
package zerowrap
