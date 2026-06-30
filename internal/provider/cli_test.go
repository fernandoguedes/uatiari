package provider

import (
	"strings"
	"testing"
	"time"
)

func TestCodexAdapterKeepsPromptSentinelLast(t *testing.T) {
	cli, err := NewCLI("codex")
	if err != nil {
		t.Fatalf("NewCLI returned error: %v", err)
	}

	args := cli.argsFor(Request{WorkingDir: "/tmp/repo"})
	if args[len(args)-1] != "-" {
		t.Fatalf("last arg = %q, want prompt sentinel", args[len(args)-1])
	}
	for i, arg := range args {
		if arg == "-C" && i > len(args)-3 {
			t.Fatalf("-C appears too late in args: %#v", args)
		}
	}
}

func TestCodexAdapterUsesSupportedExecFlags(t *testing.T) {
	cli, err := NewCLI("codex")
	if err != nil {
		t.Fatalf("NewCLI returned error: %v", err)
	}

	for _, arg := range cli.Args {
		if arg == "--ask-for-approval" {
			t.Fatalf("codex args contain unsupported approval flag: %#v", cli.Args)
		}
	}
}

func TestNewCLIAllProviders(t *testing.T) {
	providers := []string{"kimi", "gemini", "claude", "antigravity", "codex"}
	for _, name := range providers {
		cli, err := NewCLI(name)
		if err != nil {
			t.Fatalf("NewCLI(%q) returned error: %v", name, err)
		}
		if cli.Name != name {
			t.Fatalf("NewCLI(%q).Name = %q", name, cli.Name)
		}
	}
}

func TestNewCLIUnsupported(t *testing.T) {
	_, err := NewCLI("unsupported-provider")
	if err == nil {
		t.Fatal("expected error for unsupported provider, got nil")
	}
}

func TestBuildPromptContainsFields(t *testing.T) {
	req := Request{
		SystemPrompt: "sys prompt",
		Lang:         "pt_BR",
		PlanPrompt:   "plan text",
		ReviewPrompt: "review text",
		ChangedFiles: []string{"foo.go", "bar.go"},
		Diff:         "- old\n+ new",
		WorkingDir:   "/tmp",
	}
	prompt := BuildPrompt(req)
	for _, want := range []string{"sys prompt", "pt_BR", "plan text", "review text", "foo.go", "bar.go", "- old"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("BuildPrompt missing %q:\n%s", want, prompt)
		}
	}
}

func TestDoctorReturnsOneResultPerProvider(t *testing.T) {
	results := Doctor()
	if len(results) != 5 {
		t.Fatalf("Doctor returned %d results, want 5", len(results))
	}
}

func TestKimiArgsIncludeWorkDir(t *testing.T) {
	cli, _ := NewCLI("kimi")
	args := cli.argsFor(Request{WorkingDir: "/tmp/repo"})
	for i, arg := range args {
		if arg == "--work-dir" && i+1 < len(args) && args[i+1] == "/tmp/repo" {
			return
		}
	}
	t.Fatalf("kimi args missing --work-dir /tmp/repo: %v", args)
}

func TestValueOrDefaultDuration(t *testing.T) {
	if got := valueOrDefaultDuration(0, 5*time.Second); got != 5*time.Second {
		t.Fatalf("expected fallback, got %v", got)
	}
	if got := valueOrDefaultDuration(3*time.Second, 5*time.Second); got != 3*time.Second {
		t.Fatalf("expected value, got %v", got)
	}
}
