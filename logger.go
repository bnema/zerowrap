package zerowrap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// FileMode controls how app-managed log file paths are laid out.
type FileMode string

const (
	// FileModeSingle writes to one stable file under the app log directory.
	FileModeSingle FileMode = "single"

	// FileModeSession writes under a process-level session directory.
	FileModeSession FileMode = "session"
)

// FileFormat controls the format used for the file writer.
type FileFormat string

const (
	// FileFormatJSON writes JSON lines to the log file.
	FileFormatJSON FileFormat = "json"

	// FileFormatConsole writes human-readable console output to the log file.
	FileFormatConsole FileFormat = "console"
)

// FileConfig holds configuration for file-based logging.
// Use keyed fields when constructing FileConfig values; unkeyed composite
// literals are not supported as this exported configuration struct may grow.
type FileConfig struct {
	// Enabled toggles file logging on/off.
	Enabled bool

	// Path is the full log file path. If set, it takes priority over app-managed path fields.
	Path string

	// AppName enables app-managed log paths when Path is empty.
	AppName string

	// BaseDir overrides the app-managed root directory before AppName.
	BaseDir string

	// Name is the log file name. Defaults to "app". The .log extension is added if missing.
	Name string

	// Mode controls app-managed file layout. Defaults to FileModeSingle.
	Mode FileMode

	// FileFormat controls the file output format. Defaults to FileFormatJSON.
	FileFormat FileFormat

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

// New creates a new Logger with the given configuration.
func New(cfg Config) Logger {
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

	return Logger{logger}
}

// NewFromEnv creates a logger configured from environment variables.
// Uses {prefix}_LOG_LEVEL and {prefix}_LOG_FORMAT.
// Example: with prefix "MYAPP", reads MYAPP_LOG_LEVEL and MYAPP_LOG_FORMAT.
func NewFromEnv(prefix string) Logger {
	level := os.Getenv(prefix + "_LOG_LEVEL")
	format := os.Getenv(prefix + "_LOG_FORMAT")
	return New(Config{
		Level:  level,
		Format: format,
	})
}

// Default returns a sensible default logger writing to stderr with console format.
func Default() Logger {
	return New(Config{
		Level:  "info",
		Format: "console",
	})
}

// NewWithFile creates a logger that writes to both stderr and a file.
// Returns the logger, a cleanup function that must be called to close the file,
// and any error encountered.
func NewWithFile(cfg Config, fileCfg FileConfig) (Logger, func(), error) {
	if !fileCfg.Enabled || (fileCfg.Path == "" && fileCfg.AppName == "") {
		return New(cfg), func() {}, nil
	}

	fileFormat, err := normalizeFileFormat(fileCfg.FileFormat)
	if err != nil {
		return Logger{}, func() {}, fmt.Errorf("file format: %w", err)
	}

	appManaged := fileCfg.Path == "" && fileCfg.AppName != ""

	filePath, err := ResolveLogPath(fileCfg)
	if err != nil {
		return Logger{}, func() {}, fmt.Errorf("resolve log path: %w", err)
	}

	if appManaged {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o750); err != nil {
			return Logger{}, func() {}, fmt.Errorf("create log directory: %w", err)
		}
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
		Filename:   filePath,
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

	// Create multi-writer: console (formatted) + file (respects FileFormat)
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

	// File output respects FileFormat: JSON by default, or console (human-readable, no-color).
	fileOutput := io.Writer(fileWriter)
	if fileFormat == FileFormatConsole {
		fileOutput = zerolog.ConsoleWriter{
			Out:        fileWriter,
			TimeFormat: timeFormat,
			NoColor:    true,
		}
	}
	writers = append(writers, fileOutput)

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

	return Logger{logger}, cleanup, nil
}

// WithHook returns a new logger with the hook attached.
func WithHook(log Logger, hook zerolog.Hook) Logger {
	return Logger{log.Hook(hook)}
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
