package app

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fernandoguedes/uatiari/internal/config"
	gitclient "github.com/fernandoguedes/uatiari/internal/git"
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

func TestApprovedAcceptsY(t *testing.T) {
	for _, input := range []string{"y\n", "Y\n", "yes\n", "YES\n"} {
		if !approved(strings.NewReader(input)) {
			t.Fatalf("approved(%q) = false, want true", input)
		}
	}
}

func TestApprovedRejectsN(t *testing.T) {
	for _, input := range []string{"n\n", "no\n", "\n", ""} {
		if approved(strings.NewReader(input)) {
			t.Fatalf("approved(%q) = true, want false", input)
		}
	}
}

func TestConvertStats(t *testing.T) {
	in := map[string]gitclient.DiffStat{
		"foo.go": {Added: 5, Deleted: 3},
	}
	out := convertStats(in)
	if out["foo.go"].Added != 5 || out["foo.go"].Deleted != 3 {
		t.Fatalf("convertStats = %#v", out)
	}
}

func TestReviewPlanContainsFiles(t *testing.T) {
	files := []string{"foo.go", "bar.go"}
	stats := map[string]gitclient.DiffStat{
		"foo.go": {Added: 10, Deleted: 2},
	}
	plan := reviewPlan(files, stats)
	for _, want := range []string{"foo.go", "bar.go", "added"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("reviewPlan missing %q:\n%s", want, plan)
		}
	}
}

func TestRunDoctorReturnsZero(t *testing.T) {
	var stdout bytes.Buffer
	code := App{Stdout: &stdout, Stderr: &bytes.Buffer{}}.runDoctor()
	if code != 0 {
		t.Fatalf("runDoctor returned %d", code)
	}
	if stdout.Len() == 0 {
		t.Fatal("runDoctor produced no output")
	}
}

func TestSetProviderSavesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	var stdout, stderr bytes.Buffer
	a := App{Stdout: &stdout, Stderr: &stderr}

	code := a.setProvider(path, config.Config{}, "claude")
	if code != 0 {
		t.Fatalf("setProvider returned %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "claude") {
		t.Fatalf("setProvider output = %q", stdout.String())
	}
}

func TestSetProviderRejectsUnknown(t *testing.T) {
	var stderr bytes.Buffer
	a := App{Stdout: &bytes.Buffer{}, Stderr: &stderr}
	code := a.setProvider("/tmp/unused.toml", config.Config{}, "bad-provider")
	if code == 0 {
		t.Fatal("expected non-zero exit for unknown provider")
	}
}

func TestRunProvidersDoctorReturnsZero(t *testing.T) {
	var stdout bytes.Buffer
	code := App{Stdout: &stdout, Stderr: &bytes.Buffer{}}.
		Run(context.Background(), []string{"providers", "doctor"})
	if code != 0 {
		t.Fatalf("Run providers doctor returned %d", code)
	}
}

func TestRunConfigSetProviderSaves(t *testing.T) {
	dir := t.TempDir()
	// Point XDG_CONFIG_HOME so DefaultPath() lands in the temp dir
	t.Setenv("XDG_CONFIG_HOME", dir)
	var stdout bytes.Buffer
	code := App{Stdout: &stdout, Stderr: &bytes.Buffer{}}.
		Run(context.Background(), []string{"config", "set", "provider", "claude"})
	if code != 0 {
		t.Fatalf("Run config set provider returned %d", code)
	}
	if !strings.Contains(stdout.String(), "claude") {
		t.Fatalf("output = %q, expected claude", stdout.String())
	}
}

func TestRunVersionReturnsZero(t *testing.T) {
	var stdout bytes.Buffer
	code := App{Stdout: &stdout, Stderr: &bytes.Buffer{}}.Run(context.Background(), []string{"--version"})
	if code != 0 {
		t.Fatalf("Run --version returned %d", code)
	}
	if !strings.Contains(stdout.String(), "uatiari") {
		t.Fatalf("version output = %q", stdout.String())
	}
}

func TestUpdateWithCancelledContextFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stderr bytes.Buffer
	code := App{Stdout: &bytes.Buffer{}, Stderr: &stderr}.update(ctx)
	if code != 1 {
		t.Fatalf("update with cancelled context returned %d, want 1", code)
	}
}

func TestRunInvalidArgsReturnsOne(t *testing.T) {
	var stderr bytes.Buffer
	code := App{Stdout: &bytes.Buffer{}, Stderr: &stderr, Stdin: strings.NewReader("")}.
		Run(context.Background(), []string{"--invalid-flag-xyz"})
	if code != 1 {
		t.Fatalf("Run returned %d, want 1", code)
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
