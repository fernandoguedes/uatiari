package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestDefaultPath(t *testing.T) {
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath returned error: %v", err)
	}
	if path == "" {
		t.Fatal("DefaultPath returned empty string")
	}
}

func TestValueOrDefault(t *testing.T) {
	if got := valueOrDefault("", "fallback"); got != "fallback" {
		t.Fatalf("valueOrDefault('', fallback) = %q", got)
	}
	if got := valueOrDefault("value", "fallback"); got != "value" {
		t.Fatalf("valueOrDefault('value', fallback) = %q", got)
	}
}

func TestResolveFormatPrecedence(t *testing.T) {
	t.Setenv("UATIARI_FORMAT", "pretty")
	cfg := Config{Format: "markdown"}

	if got := ResolveFormat("json", cfg); got != "json" {
		t.Fatalf("flag not honoured: got %q", got)
	}
	if got := ResolveFormat("", cfg); got != "markdown" {
		t.Fatalf("config not honoured: got %q", got)
	}
	if got := ResolveFormat("", Config{}); got != "pretty" {
		t.Fatalf("env not honoured: got %q", got)
	}
	t.Setenv("UATIARI_FORMAT", "")
	if got := ResolveFormat("", Config{}); got != DefaultFormat {
		t.Fatalf("default not used: got %q", got)
	}
}

func TestResolveLangPrecedence(t *testing.T) {
	t.Setenv("UATIARI_LANG", "en_US")
	cfg := Config{Lang: "es_ES"}

	if got := ResolveLang("fr_FR", cfg); got != "fr_FR" {
		t.Fatalf("flag not honoured: got %q", got)
	}
	if got := ResolveLang("", cfg); got != "es_ES" {
		t.Fatalf("config not honoured: got %q", got)
	}
	if got := ResolveLang("", Config{}); got != "en_US" {
		t.Fatalf("env not honoured: got %q", got)
	}
	t.Setenv("UATIARI_LANG", "")
	if got := ResolveLang("", Config{}); got != DefaultLang {
		t.Fatalf("default not used: got %q", got)
	}
}

func TestValidateProviderAcceptsValid(t *testing.T) {
	for name := range ValidProviders {
		if err := ValidateProvider(name); err != nil {
			t.Fatalf("ValidateProvider(%q) returned error: %v", name, err)
		}
	}
}

func TestValidateProviderRejectsUnknown(t *testing.T) {
	if err := ValidateProvider("unknown-provider"); err == nil {
		t.Fatal("expected error for unknown provider, got nil")
	}
}

func TestLoadNonExistentReturnsEmpty(t *testing.T) {
	cfg, err := Load("/does/not/exist/config.toml")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Provider != "" || cfg.Format != "" || cfg.Lang != "" {
		t.Fatalf("expected empty Config, got %#v", cfg)
	}
}

func TestLoadIgnoresCommentsAndBlanks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	content := "# this is a comment\n\nprovider = \"claude\"\n\n# another comment\nformat = \"pretty\"\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.Provider != "claude" {
		t.Fatalf("Provider = %q, want claude", cfg.Provider)
	}
	if cfg.Format != "pretty" {
		t.Fatalf("Format = %q, want pretty", cfg.Format)
	}
}

func TestSaveUsesDefaultFormatAndLang(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := Save(path, Config{Provider: "kimi"}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	if !strings.Contains(content, DefaultFormat) {
		t.Fatalf("saved config missing default format: %s", content)
	}
	if !strings.Contains(content, DefaultLang) {
		t.Fatalf("saved config missing default lang: %s", content)
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
