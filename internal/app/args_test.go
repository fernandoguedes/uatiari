package app

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestParseReviewArgs(t *testing.T) {
	opts, err := ParseArgs([]string{"feature/auth", "--base=develop", "--skill=laravel", "--provider=codex", "--format=markdown", "--lang=en_US"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if opts.Branch != "feature/auth" || opts.Base != "develop" || opts.Skill != "laravel" || opts.ProviderFlag != "codex" || opts.FormatFlag != "markdown" || opts.LangFlag != "en_US" {
		t.Fatalf("unexpected opts: %#v", opts)
	}
}

func TestParseConfigSetProvider(t *testing.T) {
	opts, err := ParseArgs([]string{"config", "set", "provider", "kimi"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if opts.Command != "config-set-provider" || opts.ConfigProvider != "kimi" {
		t.Fatalf("unexpected opts: %#v", opts)
	}
}

func TestParseProvidersDoctor(t *testing.T) {
	opts, err := ParseArgs([]string{"providers", "doctor"})
	if err != nil {
		t.Fatalf("ParseArgs returned error: %v", err)
	}

	if opts.Command != "providers-doctor" {
		t.Fatalf("command = %q, want providers-doctor", opts.Command)
	}
}

func TestParseUnknownOptionFails(t *testing.T) {
	_, err := ParseArgs([]string{"feature/auth", "--wat"})
	if err == nil {
		t.Fatal("ParseArgs returned nil error")
	}
}

func TestHelpDoesNotRequireUserConfigDir(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CONFIG_HOME", "")

	var stdout, stderr bytes.Buffer
	code := App{
		Stdout: &stdout,
		Stderr: &stderr,
		Stdin:  strings.NewReader(""),
	}.Run(context.Background(), []string{"--help"})

	if code != 0 {
		t.Fatalf("Run returned %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "USAGE:") {
		t.Fatalf("help output missing usage: %q", stdout.String())
	}
}
