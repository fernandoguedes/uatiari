package clirunner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Runner struct {
	Command string
	Args    []string
	Dir     string
	Timeout time.Duration
}

func (r Runner) Run(ctx context.Context, input string) (string, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.Command, r.Args...)
	cmd.Dir = r.Dir
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("provider CLI %q not found", r.Command)
		}
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("provider CLI %q timed out after %s", r.Command, timeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("provider CLI %q failed: %s", r.Command, msg)
	}
	return stdout.String(), nil
}
