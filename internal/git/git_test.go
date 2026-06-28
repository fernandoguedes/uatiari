package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseNumstat(t *testing.T) {
	stats := ParseNumstat("10\t5\tsrc/main.go\n-\t-\timage.png\n20\t0\ttests/main_test.go\n")

	if stats["src/main.go"].Added != 10 || stats["src/main.go"].Deleted != 5 {
		t.Fatalf("src/main.go stats = %#v", stats["src/main.go"])
	}
	if stats["image.png"].Added != 0 || stats["image.png"].Deleted != 0 {
		t.Fatalf("image.png stats = %#v", stats["image.png"])
	}
	if stats["tests/main_test.go"].Added != 20 {
		t.Fatalf("tests/main_test.go stats = %#v", stats["tests/main_test.go"])
	}
}

func TestValidateBranchSeparatesRefFromOptions(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "git-args.log")
	gitPath := filepath.Join(dir, "git")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"" + logPath + "\"\nexit 0\n"
	if err := os.WriteFile(gitPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	err := Client{Dir: dir}.validateBranch(context.Background(), "--bad-ref", "Branch")
	if err != nil {
		t.Fatalf("validateBranch returned error: %v", err)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read git args log: %v", err)
	}
	if !strings.Contains(string(content), "rev-parse --verify -- --bad-ref") {
		t.Fatalf("git args did not separate ref from options:\n%s", content)
	}
}
