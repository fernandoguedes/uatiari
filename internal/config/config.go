package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultProvider = "gemini"
	DefaultFormat   = "json"
	DefaultLang     = "pt_BR"
)

var ValidProviders = map[string]bool{
	"kimi":        true,
	"gemini":      true,
	"claude":      true,
	"antigravity": true,
	"codex":       true,
}

type Config struct {
	Provider string
	Format   string
	Lang     string
}

func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "uatiari", "config.toml"), nil
}

func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	cfg := Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		switch key {
		case "provider":
			cfg.Provider = value
		case "format":
			cfg.Format = value
		case "lang":
			cfg.Lang = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf("provider = %q\nformat = %q\nlang = %q\n", cfg.Provider, valueOrDefault(cfg.Format, DefaultFormat), valueOrDefault(cfg.Lang, DefaultLang))
	return os.WriteFile(path, []byte(content), 0o644)
}

func ResolveProvider(flag string, cfg Config) string {
	if flag != "" {
		return flag
	}
	if cfg.Provider != "" {
		return cfg.Provider
	}
	if env := os.Getenv("UATIARI_PROVIDER"); env != "" {
		return env
	}
	return DefaultProvider
}

func ResolveFormat(flag string, cfg Config) string {
	if flag != "" {
		return flag
	}
	if cfg.Format != "" {
		return cfg.Format
	}
	if env := os.Getenv("UATIARI_FORMAT"); env != "" {
		return env
	}
	return DefaultFormat
}

func ResolveLang(flag string, cfg Config) string {
	if flag != "" {
		return flag
	}
	if cfg.Lang != "" {
		return cfg.Lang
	}
	if env := os.Getenv("UATIARI_LANG"); env != "" {
		return env
	}
	return DefaultLang
}

func ValidateProvider(provider string) error {
	if !ValidProviders[provider] {
		return fmt.Errorf("provider %q is not supported", provider)
	}
	return nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
