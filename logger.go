package zerowrap

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds logger configuration options.
type Config struct {
	// Level is the minimum log level (trace, debug, info, warn, error, fatal, panic).
	// Defaults to "info" if empty or invalid.
	Level string

	// Format is the output format: "json" or "console".
	// Defaults to "console" if empty or invalid.
	Format string

	// TimeFormat is the time format string.
	// Defaults to time.RFC3339 if empty.
	TimeFormat string

	// Output is the writer for log output.
	// Defaults to os.Stderr if nil.
	Output io.Writer

	// Caller adds caller information (file:line) to log entries.
	Caller bool
}

// FileConfig holds configuration for file-based logging.
type FileConfig struct {
	// Enabled toggles file logging on/off.
	Enabled bool

	// Path is the log file path.
	Path string

	// MaxSize is the maximum size in megabytes before rotation.
	// Defaults to 100 MB if 0.
	MaxSize int

	// MaxBackups is the maximum number of old log files to retain.
	// Defaults to 3 if 0.
	MaxBackups int

	// MaxAge is the maximum number of days to retain old log files.
	// Defaults to 28 if 0.
	MaxAge int

	// Compress determines if rotated files should be compressed.
	Compress bool
}

// New creates a new zerolog.Logger with the given configuration.
func New(cfg Config) zerolog.Logger {
	output := cfg.Output
	if output == nil {
		output = os.Stderr
	}

	timeFormat := cfg.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	format := strings.ToLower(cfg.Format)
	if format == "console" || format == "" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: timeFormat,
		}
	}

	level := parseLevel(cfg.Level)

	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	if cfg.Caller {
		logger = logger.With().Caller().Logger()
	}

	return logger
}

// NewFromEnv creates a logger configured from environment variables.
// Uses {prefix}_LOG_LEVEL and {prefix}_LOG_FORMAT.
// Example: with prefix "MYAPP", reads MYAPP_LOG_LEVEL and MYAPP_LOG_FORMAT.
func NewFromEnv(prefix string) zerolog.Logger {
	level := os.Getenv(prefix + "_LOG_LEVEL")
	format := os.Getenv(prefix + "_LOG_FORMAT")
	return New(Config{
		Level:  level,
		Format: format,
	})
}

// Default returns a sensible default logger writing to stderr with console format.
func Default() zerolog.Logger {
	return New(Config{
		Level:  "info",
		Format: "console",
	})
}

// NewWithFile creates a logger that writes to both stderr and a file.
// Returns the logger, a cleanup function that must be called to close the file,
// and any error encountered.
func NewWithFile(cfg Config, fileCfg FileConfig) (zerolog.Logger, func(), error) {
	if !fileCfg.Enabled || fileCfg.Path == "" {
		return New(cfg), func() {}, nil
	}

	// Set defaults for file config
	maxSize := fileCfg.MaxSize
	if maxSize == 0 {
		maxSize = 100
	}
	maxBackups := fileCfg.MaxBackups
	if maxBackups == 0 {
		maxBackups = 3
	}
	maxAge := fileCfg.MaxAge
	if maxAge == 0 {
		maxAge = 28
	}

	fileWriter := &lumberjack.Logger{
		Filename:   fileCfg.Path,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   fileCfg.Compress,
	}

	cleanup := func() {
		_ = fileWriter.Close()
	}

	// Determine console output
	consoleOutput := cfg.Output
	if consoleOutput == nil {
		consoleOutput = os.Stderr
	}

	timeFormat := cfg.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	// Create multi-writer: console (formatted) + file (JSON)
	var writers []io.Writer

	format := strings.ToLower(cfg.Format)
	if format == "console" || format == "" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        consoleOutput,
			TimeFormat: timeFormat,
		})
	} else {
		writers = append(writers, consoleOutput)
	}

	// File always gets JSON format for easy parsing
	writers = append(writers, fileWriter)

	multiWriter := zerolog.MultiLevelWriter(writers...)

	level := parseLevel(cfg.Level)

	logger := zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Logger()

	if cfg.Caller {
		logger = logger.With().Caller().Logger()
	}

	return logger, cleanup, nil
}

// WithHook returns a new logger with the hook attached.
func WithHook(log zerolog.Logger, hook zerolog.Hook) zerolog.Logger {
	return log.Hook(hook)
}

// parseLevel converts a level string to zerolog.Level.
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info", "":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}
