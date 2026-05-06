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

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %q: %v", path, err)
	}
	return string(data)
}
