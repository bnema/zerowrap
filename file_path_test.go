package zerowrap

import "testing"

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
}
