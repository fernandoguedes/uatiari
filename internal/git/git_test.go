package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeFakeGit creates a fake git binary in a temp dir and returns a Client
// wired to use it. The script receives all args as $* and can branch on them.
func makeFakeGit(t *testing.T, script string) Client {
	t.Helper()
	dir := t.TempDir()
	gitPath := filepath.Join(dir, "git")
	if err := os.WriteFile(gitPath, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return Client{Dir: dir}
}

func TestErrorMessage(t *testing.T) {
	e := Error{Message: "something went wrong"}
	if e.Error() != "something went wrong" {
		t.Fatalf("Error() = %q", e.Error())
	}
}

func TestSplitLinesFiltersEmpty(t *testing.T) {
	got := splitLines("a\n\nb\n  \nc\n")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("splitLines = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("splitLines[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseNumstatIntSpecialCases(t *testing.T) {
	if parseNumstatInt("-") != 0 {
		t.Fatal("expected 0 for binary marker '-'")
	}
	if parseNumstatInt("notanint") != 0 {
		t.Fatal("expected 0 for invalid string")
	}
	if parseNumstatInt("42") != 42 {
		t.Fatal("expected 42")
	}
}

func TestCurrentBranchReturnsName(t *testing.T) {
	client := makeFakeGit(t, `printf 'main\n'`)
	branch, err := client.CurrentBranch(context.Background())
	if err != nil {
		t.Fatalf("CurrentBranch returned error: %v", err)
	}
	if branch != "main" {
		t.Fatalf("branch = %q, want main", branch)
	}
}

func TestCurrentBranchDetachedHEAD(t *testing.T) {
	client := makeFakeGit(t, `printf 'HEAD\n'`)
	_, err := client.CurrentBranch(context.Background())
	if err == nil {
		t.Fatal("expected error for detached HEAD, got nil")
	}
}

func TestCheckRepositoryFails(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	err := client.checkRepository(context.Background())
	if err == nil {
		t.Fatal("expected error when not in a git repo")
	}
}

func TestDiffReturnsDiff(t *testing.T) {
	// rev-parse always succeeds; diff prints content
	client := makeFakeGit(t, `case "$*" in
  *diff*) printf '%s\n' '- old' '+ new' ;;
  *) exit 0 ;;
esac`)
	diff, err := client.Diff(context.Background(), "feature", "main")
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}
	if !strings.Contains(diff, "old") {
		t.Fatalf("Diff output = %q, expected to contain 'old'", diff)
	}
}

func TestDiffEmptyReturnsError(t *testing.T) {
	client := makeFakeGit(t, `case "$*" in
  *diff*) printf '' ;;
  *) exit 0 ;;
esac`)
	_, err := client.Diff(context.Background(), "feature", "main")
	if err == nil {
		t.Fatal("expected error for empty diff, got nil")
	}
}

func TestChangedFilesReturnsList(t *testing.T) {
	client := makeFakeGit(t, `case "$*" in
  *--name-only*) printf 'foo.go\nbar.go\n' ;;
  *) exit 0 ;;
esac`)
	files, err := client.ChangedFiles(context.Background(), "feature", "main")
	if err != nil {
		t.Fatalf("ChangedFiles returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ChangedFiles = %#v, want 2 files", files)
	}
}

func TestChangedFilesEmptyReturnsError(t *testing.T) {
	client := makeFakeGit(t, `case "$*" in
  *--name-only*) printf '' ;;
  *) exit 0 ;;
esac`)
	_, err := client.ChangedFiles(context.Background(), "feature", "main")
	if err == nil {
		t.Fatal("expected error for empty file list, got nil")
	}
}

func TestDiffStatsReturnsParsed(t *testing.T) {
	client := makeFakeGit(t, `case "$*" in
  *--numstat*) printf '5\t3\tfoo.go\n' ;;
  *) exit 0 ;;
esac`)
	stats, err := client.DiffStats(context.Background(), "feature", "main")
	if err != nil {
		t.Fatalf("DiffStats returned error: %v", err)
	}
	if stats["foo.go"].Added != 5 || stats["foo.go"].Deleted != 3 {
		t.Fatalf("stats[foo.go] = %#v, want Added=5 Deleted=3", stats["foo.go"])
	}
}

func TestValidateBranchEmptyString(t *testing.T) {
	client := makeFakeGit(t, `exit 0`)
	err := client.validateBranch(context.Background(), "", "Base branch")
	if err == nil {
		t.Fatal("expected error for empty branch, got nil")
	}
}

func TestValidateBranchNonExistent(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	err := client.validateBranch(context.Background(), "nonexistent", "Branch")
	if err == nil {
		t.Fatal("expected error when branch rev-parse fails")
	}
}

func TestDiffFailsWhenRepoInvalid(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	_, err := client.Diff(context.Background(), "feature", "main")
	if err == nil {
		t.Fatal("expected Diff error when not in a git repo")
	}
}

func TestChangedFilesFailsWhenRepoInvalid(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	_, err := client.ChangedFiles(context.Background(), "feature", "main")
	if err == nil {
		t.Fatal("expected ChangedFiles error when not in a git repo")
	}
}

func TestDiffStatsFailsWhenRepoInvalid(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	_, err := client.DiffStats(context.Background(), "feature", "main")
	if err == nil {
		t.Fatal("expected DiffStats error when not in a git repo")
	}
}

func TestRepositoryFilesFailsWhenRepoInvalid(t *testing.T) {
	client := makeFakeGit(t, `exit 1`)
	_, err := client.RepositoryFiles(context.Background())
	if err == nil {
		t.Fatal("expected RepositoryFiles error when not in a git repo")
	}
}

func TestRepositoryFilesReturnsList(t *testing.T) {
	client := makeFakeGit(t, `case "$*" in
  *ls-tree*) printf 'a.go\nb.go\n' ;;
  *) exit 0 ;;
esac`)
	files, err := client.RepositoryFiles(context.Background())
	if err != nil {
		t.Fatalf("RepositoryFiles returned error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("RepositoryFiles = %#v, want 2 files", files)
	}
}

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
