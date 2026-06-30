package clirunner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunnerPassesPromptOnStdin(t *testing.T) {
	dir := t.TempDir()
	name := "fake-provider"
	if runtime.GOOS == "windows" {
		name += ".bat"
	}
	path := filepath.Join(dir, name)
	script := "#!/bin/sh\ncat | sed 's/^/seen:/'\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := Runner{Command: path, Timeout: 5 * time.Second}
	out, err := runner.Run(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if strings.TrimSpace(out) != "seen:hello" {
		t.Fatalf("out = %q, want seen:hello", out)
	}
}

func TestRunnerReportsExitCodeAndStderr(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-provider")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho boom >&2\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	runner := Runner{Command: path, Timeout: 5 * time.Second}
	_, err := runner.Run(context.Background(), "hello")
	if err == nil {
		t.Fatal("Run returned nil error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %q, want stderr", err.Error())
	}
}
