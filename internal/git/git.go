package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Error struct {
	Message string
}

func (e Error) Error() string { return e.Message }

type DiffStat struct {
	Added   int `json:"added"`
	Deleted int `json:"deleted"`
}

type Client struct {
	Dir string
}

func (c Client) Diff(ctx context.Context, branch, base string) (string, error) {
	if err := c.checkRepository(ctx); err != nil {
		return "", err
	}
	if err := c.validateBranch(ctx, base, "Base branch"); err != nil {
		return "", err
	}
	if err := c.validateBranch(ctx, branch, "Branch"); err != nil {
		return "", err
	}
	out, err := c.run(ctx, "diff", base+"..."+branch)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(out) == "" {
		return "", Error{Message: fmt.Sprintf("No differences found between %q and %q.", base, branch)}
	}
	return out, nil
}

func (c Client) ChangedFiles(ctx context.Context, branch, base string) ([]string, error) {
	if err := c.checkRepository(ctx); err != nil {
		return nil, err
	}
	if err := c.validateBranch(ctx, base, "Base branch"); err != nil {
		return nil, err
	}
	if err := c.validateBranch(ctx, branch, "Branch"); err != nil {
		return nil, err
	}
	out, err := c.run(ctx, "diff", "--name-only", base+"..."+branch)
	if err != nil {
		return nil, err
	}
	files := splitLines(out)
	if len(files) == 0 {
		return nil, Error{Message: fmt.Sprintf("No files changed between %q and %q.", base, branch)}
	}
	return files, nil
}

func (c Client) DiffStats(ctx context.Context, branch, base string) (map[string]DiffStat, error) {
	if err := c.checkRepository(ctx); err != nil {
		return nil, err
	}
	if err := c.validateBranch(ctx, base, "Base branch"); err != nil {
		return nil, err
	}
	if err := c.validateBranch(ctx, branch, "Branch"); err != nil {
		return nil, err
	}
	out, err := c.run(ctx, "diff", "--numstat", base+"..."+branch)
	if err != nil {
		return nil, err
	}
	return ParseNumstat(out), nil
}

func (c Client) RepositoryFiles(ctx context.Context) ([]string, error) {
	if err := c.checkRepository(ctx); err != nil {
		return nil, err
	}
	out, err := c.run(ctx, "ls-tree", "-r", "--name-only", "HEAD")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

func (c Client) CurrentBranch(ctx context.Context) (string, error) {
	out, err := c.run(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(out)
	if branch == "HEAD" {
		return "", Error{Message: "Currently in detached HEAD state. Please checkout a branch."}
	}
	return branch, nil
}

func (c Client) checkRepository(ctx context.Context) error {
	_, err := c.run(ctx, "rev-parse", "--git-dir")
	if err != nil {
		return Error{Message: "Not in a git repository. Please run from within a git repo."}
	}
	return nil
}

func (c Client) validateBranch(ctx context.Context, branch, label string) error {
	if branch == "" {
		return Error{Message: label + " is required."}
	}
	_, err := c.run(ctx, "rev-parse", "--verify", "--", branch)
	if err != nil {
		return Error{Message: fmt.Sprintf("%s %q does not exist.", label, branch)}
	}
	return nil
}

func (c Client) run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = c.Dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", Error{Message: "git command not found. Please install git."}
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", Error{Message: "Git command failed: " + msg}
	}
	return stdout.String(), nil
}

func ParseNumstat(output string) map[string]DiffStat {
	stats := map[string]DiffStat{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		stats[parts[2]] = DiffStat{
			Added:   parseNumstatInt(parts[0]),
			Deleted: parseNumstatInt(parts[1]),
		}
	}
	return stats
}

func parseNumstatInt(value string) int {
	if value == "-" {
		return 0
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func splitLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
