package zerowrap

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewWithFileDisabledReturnsUsableLogger(t *testing.T) {
	var out bytes.Buffer
	log, cleanup, err := NewWithFile(
		Config{Output: &out, Format: "json"},
		FileConfig{Enabled: false},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// cleanup must be callable (no-op is fine).
	cleanup()

	msg := "disabled-test-message"
	log.Info().Msg(msg)
	if !strings.Contains(out.String(), msg) {
		t.Fatalf("output did not contain %q, got: %s", msg, out.String())
	}
}

func TestNewWithFileEnabledWithoutPathOrAppNameIsNoopCompatible(t *testing.T) {
	var out bytes.Buffer
	log, cleanup, err := NewWithFile(
		Config{Output: &out, Format: "console"},
		FileConfig{Enabled: true},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanup()

	msg := "noop-compatible-message"
	log.Info().Msg(msg)
	if !strings.Contains(out.String(), msg) {
		t.Fatalf("output did not contain %q, got: %s", msg, out.String())
	}
}

func TestNewWithFileExplicitPathStillWritesJSONByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.log")

	var out bytes.Buffer
	log, cleanup, err := NewWithFile(
		Config{Output: &out, Format: "json"},
		FileConfig{Enabled: true, Path: path},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := "explicit-path-message"
	log.Info().Msg(msg)

	// Message also goes to the configured console output.
	if !strings.Contains(out.String(), msg) {
		t.Fatalf("console output did not contain %q, got: %s", msg, out.String())
	}

	cleanup()

	content := readFileForTest(t, path)
	if !strings.Contains(content, msg) {
		t.Fatalf("file %q did not contain %q, got: %s", path, msg, content)
	}
}

func TestNewWithFileExplicitPathWinsOverAppName(t *testing.T) {
	base := t.TempDir()
	explicitPath := filepath.Join(base, "explicit.log")

	var out bytes.Buffer
	log, cleanup, err := NewWithFile(
		Config{Output: &out, Format: "json"},
		FileConfig{
			Enabled: true,
			Path:    explicitPath,
			AppName: "IgnoredApp",
			BaseDir: base,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := "explicit-wins-message"
	log.Info().Msg(msg)

	cleanup()

	content := readFileForTest(t, explicitPath)
	if !strings.Contains(content, msg) {
		t.Fatalf("explicit file %q did not contain %q, got: %s", explicitPath, msg, content)
	}

	// Prove no app-managed directories were created.
	appDir := filepath.Join(base, "IgnoredApp")
	if _, statErr := os.Stat(appDir); statErr == nil {
		t.Fatalf("app-managed directory %q exists, expected it not to", appDir)
	}
}

func TestNewWithFileAppManagedSingleCreatesDirectoryAndWritesFile(t *testing.T) {
	base := t.TempDir()
	fileCfg := FileConfig{
		Enabled: true,
		AppName: "MyApp",
		BaseDir: base,
		Name:    "app",
		Mode:    FileModeSingle,
	}

	expectedPath, err := ResolveLogPath(fileCfg)
	if err != nil {
		t.Fatalf("ResolveLogPath returned error: %v", err)
	}

	log, cleanup, err := NewWithFile(
		Config{Format: "json"},
		fileCfg,
	)
	if err != nil {
		t.Fatalf("NewWithFile returned error: %v", err)
	}

	log.Info().Msg("managed")
	cleanup()

	content := readFileForTest(t, expectedPath)
	if !strings.Contains(content, `"message":"managed"`) {
		t.Fatalf("file %q did not contain expected message, got: %s", expectedPath, content)
	}
}

func TestNewWithFileAppManagedSessionCreatesSharedSessionDirectory(t *testing.T) {
	base := t.TempDir()

	appCfg := FileConfig{
		Enabled: true,
		AppName: "MyApp",
		BaseDir: base,
		Name:    "app",
		Mode:    FileModeSession,
	}
	dbCfg := FileConfig{
		Enabled: true,
		AppName: "MyApp",
		BaseDir: base,
		Name:    "database",
		Mode:    FileModeSession,
	}

	appPath, err := ResolveLogPath(appCfg)
	if err != nil {
		t.Fatalf("ResolveLogPath for app returned error: %v", err)
	}
	dbPath, err := ResolveLogPath(dbCfg)
	if err != nil {
		t.Fatalf("ResolveLogPath for database returned error: %v", err)
	}

	if filepath.Dir(appPath) != filepath.Dir(dbPath) {
		t.Fatalf("expected same session directory, got app=%q db=%q", filepath.Dir(appPath), filepath.Dir(dbPath))
	}

	appLog, appCleanup, err := NewWithFile(Config{Format: "json"}, appCfg)
	if err != nil {
		t.Fatalf("NewWithFile for app returned error: %v", err)
	}
	dbLog, dbCleanup, err := NewWithFile(Config{Format: "json"}, dbCfg)
	if err != nil {
		t.Fatalf("NewWithFile for database returned error: %v", err)
	}

	appLog.Info().Msg("app-message")
	dbLog.Info().Msg("database-message")

	appCleanup()
	dbCleanup()

	appContent := readFileForTest(t, appPath)
	if !strings.Contains(appContent, `"message":"app-message"`) {
		t.Fatalf("app file %q did not contain expected message, got: %s", appPath, appContent)
	}

	dbContent := readFileForTest(t, dbPath)
	if !strings.Contains(dbContent, `"message":"database-message"`) {
		t.Fatalf("database file %q did not contain expected message, got: %s", dbPath, dbContent)
	}
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", path, err)
	}
	return string(data)
}
