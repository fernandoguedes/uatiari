package provider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/fernandoguedes/uatiari/internal/provider/clirunner"
)

type CLIProvider struct {
	Name    string
	Command string
	Args    []string
	Timeout time.Duration
}

func NewCLI(name string) (CLIProvider, error) {
	switch name {
	case "kimi":
		return CLIProvider{Name: name, Command: "kimi", Args: []string{"--quiet"}}, nil
	case "gemini":
		return CLIProvider{Name: name, Command: "gemini", Args: []string{"--prompt", "Execute the prompt from stdin and return only the requested JSON.", "--approval-mode", "plan"}}, nil
	case "claude":
		return CLIProvider{Name: name, Command: "claude", Args: []string{"--bare", "-p", "Execute the prompt from stdin and return only the requested JSON.", "--output-format", "text"}}, nil
	case "antigravity":
		return CLIProvider{Name: name, Command: "agy", Args: []string{}}, nil
	case "codex":
		return CLIProvider{Name: name, Command: "codex", Args: []string{"exec", "--sandbox", "read-only", "--color", "never"}}, nil
	default:
		return CLIProvider{}, fmt.Errorf("provider %q is not supported", name)
	}
}

func (p CLIProvider) Run(ctx context.Context, request Request) (Response, error) {
	runner := clirunner.Runner{
		Command: p.Command,
		Args:    p.argsFor(request),
		Dir:     request.WorkingDir,
		Timeout: valueOrDefaultDuration(p.Timeout, 10*time.Minute),
	}
	content, err := runner.Run(ctx, BuildPrompt(request))
	if err != nil {
		return Response{}, err
	}
	return Response{Content: content}, nil
}

func (p CLIProvider) argsFor(request Request) []string {
	args := append([]string{}, p.Args...)
	if p.Name == "kimi" {
		args = append(args, "--work-dir", request.WorkingDir)
	}
	if p.Name == "codex" {
		args = append(args, "-C", request.WorkingDir)
		args = append(args, "-")
	}
	return args
}

func BuildPrompt(request Request) string {
	return strings.TrimSpace(fmt.Sprintf(`%s

Idioma da resposta: %s.

Retorne somente JSON válido no schema solicitado. Não use markdown fora dos campos JSON.

Plano de revisão:
%s

Tarefa:
%s

Arquivos alterados:
%s

Diff:
%s
`, request.SystemPrompt, request.Lang, request.PlanPrompt, request.ReviewPrompt, strings.Join(request.ChangedFiles, "\n"), request.Diff))
}

func Doctor() []DoctorResult {
	providers := []string{"kimi", "gemini", "claude", "antigravity", "codex"}
	results := make([]DoctorResult, 0, len(providers))
	for _, name := range providers {
		cli, err := NewCLI(name)
		if err != nil {
			results = append(results, DoctorResult{Name: name, OK: false, Message: err.Error()})
			continue
		}
		path, err := exec.LookPath(cli.Command)
		if err != nil {
			results = append(results, DoctorResult{Name: name, OK: false, Message: fmt.Sprintf("%s not found in PATH", cli.Command)})
			continue
		}
		results = append(results, DoctorResult{Name: name, OK: true, Message: path})
	}
	return results
}

type DoctorResult struct {
	Name    string
	OK      bool
	Message string
}

func valueOrDefaultDuration(value, fallback time.Duration) time.Duration {
	if value == 0 {
		return fallback
	}
	return value
}
