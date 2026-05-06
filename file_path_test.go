package zerowrap

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileConfigExposesAppManagedFields(t *testing.T) {
	cfg := FileConfig{
		Enabled:    true,
		AppName:    "MyApp",
		BaseDir:    t.TempDir(),
		Name:       "app",
		Mode:       FileModeSingle,
		FileFormat: FileFormatConsole,
	}

	if cfg.Mode != FileModeSingle {
		t.Fatalf("Mode = %q, want %q", cfg.Mode, FileModeSingle)
	}
	if cfg.FileFormat != FileFormatConsole {
		t.Fatalf("FileFormat = %q, want %q", cfg.FileFormat, FileFormatConsole)
	}
	// Suppress unused field diagnostics.
	_, _, _ = cfg.Enabled, cfg.AppName, cfg.BaseDir
	_ = cfg.Name
}

func TestResolveLogPathReturnsExplicitPath(t *testing.T) {
	p, err := ResolveLogPath(FileConfig{Path: "/custom/path.log"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "/custom/path.log" {
		t.Fatalf("path = %q, want %q", p, "/custom/path.log")
	}

	// Path wins even when AppName is set.
	p, err = ResolveLogPath(FileConfig{
		Path:    "/custom/path.log",
		AppName: "MyApp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "/custom/path.log" {
		t.Fatalf("path = %q, want %q", p, "/custom/path.log")
	}

	// Empty Path with empty AppName returns empty path.
	p, err = ResolveLogPath(FileConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p != "" {
		t.Fatalf("path = %q, want empty", p)
	}
}

func TestResolveLogPathSingleWithBaseDir(t *testing.T) {
	tmp := t.TempDir()
	p, err := ResolveLogPath(FileConfig{
		AppName: "MyApp",
		BaseDir: tmp,
		Name:    "svc",
		Mode:    FileModeSingle,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(tmp, "MyApp", "logs", "svc.log")
	if p != want {
		t.Fatalf("path = %q, want %q", p, want)
	}
}

func TestResolveLogPathDefaultsNameModeAndExtension(t *testing.T) {
	tmp := t.TempDir()
	// Name empty → defaults to "app".
	// Mode empty → defaults to FileModeSingle.
	// FileFormat empty → defaults to FileFormatJSON (no error).
	p, err := ResolveLogPath(FileConfig{
		AppName: "MyApp",
		BaseDir: tmp,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(tmp, "MyApp", "logs", "app.log")
	if p != want {
		t.Fatalf("path = %q, want %q", p, want)
	}
}

func TestResolveLogPathDoesNotDoubleAppendLogExtension(t *testing.T) {
	tmp := t.TempDir()
	p, err := ResolveLogPath(FileConfig{
		AppName: "MyApp",
		BaseDir: tmp,
		Name:    "app.log",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(tmp, "MyApp", "logs", "app.log")
	if p != want {
		t.Fatalf("path = %q, want %q", p, want)
	}
}

func TestResolveLogPathRejectsPathSeparators(t *testing.T) {
	_, err := ResolveLogPath(FileConfig{AppName: "foo/bar"})
	if err == nil {
		t.Fatal("expected error for AppName with path separator")
	}
	if !strings.Contains(err.Error(), "AppName") {
		t.Fatalf("error should mention AppName, got: %v", err)
	}

	_, err = ResolveLogPath(FileConfig{AppName: `foo\bar`})
	if err == nil {
		t.Fatal("expected error for AppName with backslash")
	}
	if !strings.Contains(err.Error(), "AppName") {
		t.Fatalf("error should mention AppName, got: %v", err)
	}

	_, err = ResolveLogPath(FileConfig{AppName: "MyApp", Name: "a/b"})
	if err == nil {
		t.Fatal("expected error for Name with path separator")
	}
	if !strings.Contains(err.Error(), "Name") {
		t.Fatalf("error should mention Name, got: %v", err)
	}

	_, err = ResolveLogPath(FileConfig{AppName: "."})
	if err == nil {
		t.Fatal("expected error for AppName '.'")
	}
	if !strings.Contains(err.Error(), "AppName") {
		t.Fatalf("error should mention AppName, got: %v", err)
	}

	_, err = ResolveLogPath(FileConfig{AppName: ".."})
	if err == nil {
		t.Fatal("expected error for AppName '..'")
	}
	if !strings.Contains(err.Error(), "AppName") {
		t.Fatalf("error should mention AppName, got: %v", err)
	}
}

func TestResolveLogPathRejectsInvalidModeAndFormat(t *testing.T) {
	_, err := ResolveLogPath(FileConfig{
		AppName: "MyApp",
		Mode:    "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid Mode")
	}

	_, err = ResolveLogPath(FileConfig{
		AppName:    "MyApp",
		FileFormat: "invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid FileFormat")
	}
}

func TestResolveLogPathSessionMode(t *testing.T) {
	tmp := t.TempDir()
	p, err := ResolveLogPath(FileConfig{
		AppName: "MyApp",
		BaseDir: tmp,
		Name:    "svc",
		Mode:    FileModeSession,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	dir := filepath.Dir(p)
	if !strings.Contains(dir, "sessions") {
		t.Fatalf("session path should contain 'sessions', got: %q", p)
	}
	// Verify the file name is svc.log under the session dir.
	base := filepath.Base(p)
	if base != "svc.log" {
		t.Fatalf("base = %q, want %q", base, "svc.log")
	}
}

func TestResolveLogPathLinuxUsesXDGStateHome(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}
	xdgDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", xdgDir)

	p, err := ResolveLogPath(FileConfig{AppName: "MyApp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(xdgDir, "MyApp", "logs", "app.log")
	if p != want {
		t.Fatalf("path = %q, want %q", p, want)
	}
}

func TestResolveLogPathLinuxFallsBackToDotLocal(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}
	t.Setenv("XDG_STATE_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)

	p, err := ResolveLogPath(FileConfig{AppName: "MyApp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(home, ".local", "state", "MyApp", "logs", "app.log")
	if p != want {
		t.Fatalf("path = %q, want %q", p, want)
	}
}

func TestResolveLogPathDefaultRootIsNotEmpty(t *testing.T) {
	// App-managed without BaseDir must resolve to a non-empty root on every OS.
	p, err := ResolveLogPath(FileConfig{AppName: "MyApp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == "" {
		t.Fatal("expected non-empty path")
	}
	if !strings.HasSuffix(p, ".log") {
		t.Fatalf("path should end with .log, got %q", p)
	}
	_ = os.Getenv // ensure os is used
}
