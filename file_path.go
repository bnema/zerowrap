package zerowrap

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	sessionID     string
	sessionIDOnce sync.Once
)

func getSessionID() string {
	sessionIDOnce.Do(func() {
		ts := time.Now().UnixMilli()
		pid := os.Getpid()
		base := fmt.Sprintf("%d-%d", ts, pid)
		randBytes := make([]byte, 4)
		if _, err := rand.Read(randBytes); err == nil {
			sessionID = base + "-" + hex.EncodeToString(randBytes)
		} else {
			sessionID = base
		}
	})
	return sessionID
}

// ResolveLogPath resolves the log file path from the given FileConfig.
//
// If Path is set, it is returned directly and no other fields are validated.
// If Path is empty and AppName is empty, an empty path is returned with no error.
// Otherwise, an app-managed path is built from AppName, defaulting Name to "app",
// Mode to FileModeSingle, and FileFormat to FileFormatJSON.
func ResolveLogPath(fileCfg FileConfig) (string, error) {
	if fileCfg.Path != "" {
		return fileCfg.Path, nil
	}
	if fileCfg.AppName == "" {
		return "", nil
	}

	if err := validateComponent(fileCfg.AppName, "AppName"); err != nil {
		return "", err
	}

	name := fileCfg.Name
	if name == "" {
		name = "app"
	}
	if err := validateComponent(name, "Name"); err != nil {
		return "", err
	}
	if !strings.HasSuffix(name, ".log") {
		name += ".log"
	}

	mode := fileCfg.Mode
	if mode == "" {
		mode = FileModeSingle
	}
	switch mode {
	case FileModeSingle, FileModeSession:
	default:
		return "", fmt.Errorf("invalid FileMode: %q (must be %q or %q)", mode, FileModeSingle, FileModeSession)
	}

	_, err := normalizeFileFormat(fileCfg.FileFormat)
	if err != nil {
		return "", err
	}

	root, err := resolveLogRoot(fileCfg)
	if err != nil {
		return "", err
	}

	switch mode {
	case FileModeSingle:
		return filepath.Join(root, name), nil
	case FileModeSession:
		return filepath.Join(root, "sessions", getSessionID(), name), nil
	default:
		return "", fmt.Errorf("unexpected mode: %q", mode)
	}
}

func resolveLogRoot(fileCfg FileConfig) (string, error) {
	if fileCfg.BaseDir != "" {
		return filepath.Join(fileCfg.BaseDir, fileCfg.AppName, "logs"), nil
	}
	return defaultLogRoot(fileCfg.AppName)
}

func defaultLogRoot(appName string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Logs", appName), nil
	case "windows":
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("resolve home directory: %w", err)
			}
			dir = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(dir, appName, "Logs"), nil
	default: // linux and others
		xdg := os.Getenv("XDG_STATE_HOME")
		if xdg != "" {
			return filepath.Join(xdg, appName, "logs"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, ".local", "state", appName, "logs"), nil
	}
}

func validateComponent(s, fieldName string) error {
	if s == "." || s == ".." {
		return fmt.Errorf("%s must not be %q", fieldName, s)
	}
	if strings.ContainsAny(s, "/\\") {
		return fmt.Errorf("%s must not contain path separators, got %q", fieldName, s)
	}
	return nil
}

// normalizeFileFormat validates and defaults the FileFormat value.
// Returns the normalized format and an error if the value is invalid.
func normalizeFileFormat(f FileFormat) (FileFormat, error) {
	if f == "" {
		return FileFormatJSON, nil
	}
	switch f {
	case FileFormatJSON, FileFormatConsole:
		return f, nil
	default:
		return "", fmt.Errorf("invalid FileFormat: %q (must be %q or %q)", f, FileFormatJSON, FileFormatConsole)
	}
}
