package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProviderPrecedenceFlagConfigEnvDefault(t *testing.T) {
	t.Setenv("UATIARI_PROVIDER", "kimi")
	cfg := Config{Provider: "gemini"}

	resolved := ResolveProvider("codex", cfg)
	if resolved != "codex" {
		t.Fatalf("resolved = %q, want codex", resolved)
	}

	resolved = ResolveProvider("", cfg)
	if resolved != "gemini" {
		t.Fatalf("resolved = %q, want gemini", resolved)
	}

	resolved = ResolveProvider("", Config{})
	if resolved != "kimi" {
		t.Fatalf("resolved = %q, want kimi", resolved)
	}

	os.Unsetenv("UATIARI_PROVIDER")
	resolved = ResolveProvider("", Config{})
	if resolved != DefaultProvider {
		t.Fatalf("resolved = %q, want %q", resolved, DefaultProvider)
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := Config{Provider: "codex"}

	if err := Save(path, cfg); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if loaded.Provider != "codex" {
		t.Fatalf("provider = %q, want codex", loaded.Provider)
	}
}
