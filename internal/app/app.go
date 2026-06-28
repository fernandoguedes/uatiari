package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fernandoguedes/uatiari/internal/config"
	gitclient "github.com/fernandoguedes/uatiari/internal/git"
	"github.com/fernandoguedes/uatiari/internal/provider"
	"github.com/fernandoguedes/uatiari/internal/report"
	"github.com/fernandoguedes/uatiari/internal/review"
	"github.com/fernandoguedes/uatiari/internal/version"
)

type App struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
	Dir    string
}

func (a App) Run(ctx context.Context, args []string) int {
	if a.Stdout == nil {
		a.Stdout = os.Stdout
	}
	if a.Stderr == nil {
		a.Stderr = os.Stderr
	}
	if a.Stdin == nil {
		a.Stdin = os.Stdin
	}
	if a.Dir == "" {
		wd, _ := os.Getwd()
		a.Dir = wd
	}

	opts, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n\n", err)
		printHelp(a.Stdout)
		return 1
	}

	cfgPath, err := config.DefaultPath()
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error loading config: %v\n", err)
		return 1
	}

	switch opts.Command {
	case "help":
		printHelp(a.Stdout)
		return 0
	case "version":
		fmt.Fprintf(a.Stdout, "uatiari version %s\n", version.Version)
		return 0
	case "providers-doctor":
		return a.runDoctor()
	case "config-set-provider":
		return a.setProvider(cfgPath, cfg, opts.ConfigProvider)
	case "update":
		return a.update(ctx)
	default:
		return a.review(ctx, opts, cfg)
	}
}

func (a App) review(ctx context.Context, opts Options, cfg config.Config) int {
	providerName := config.ResolveProvider(opts.ProviderFlag, cfg)
	if err := config.ValidateProvider(providerName); err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	format := config.ResolveFormat(opts.FormatFlag, cfg)
	lang := config.ResolveLang(opts.LangFlag, cfg)

	git := gitclient.Client{Dir: a.Dir}
	fmt.Fprintf(a.Stderr, "Fetching git context for %s -> %s...\n", opts.Branch, opts.Base)
	diff, err := git.Diff(ctx, opts.Branch, opts.Base)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	changedFiles, err := git.ChangedFiles(ctx, opts.Branch, opts.Base)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	stats, err := git.DiffStats(ctx, opts.Branch, opts.Base)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	repoFiles, _ := git.RepositoryFiles(ctx)

	fmt.Fprintf(a.Stderr, "Found %d changed file(s). Provider: %s\n", len(changedFiles), providerName)
	fmt.Fprintln(a.Stderr, reviewPlan(changedFiles, stats))
	fmt.Fprint(a.Stderr, "Approve execution? (y/n): ")
	if !approved(a.Stdin) {
		fmt.Fprintln(a.Stderr, "Cancelled.")
		return 1
	}

	cli, err := provider.NewCLI(providerName)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	result, err := review.Reviewer{Provider: cli, Lang: lang}.Execute(ctx, review.Input{
		WorkingDir:   a.Dir,
		Branch:       opts.Branch,
		Base:         opts.Base,
		ManualSkill:  opts.Skill,
		Diff:         diff,
		ChangedFiles: changedFiles,
		RepoFiles:    repoFiles,
		DiffStats:    convertStats(stats),
	})
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	output, err := report.Render(result, format)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprint(a.Stdout, output)
	return 0
}

func reviewPlan(changedFiles []string, stats map[string]gitclient.DiffStat) string {
	var b strings.Builder
	b.WriteString("\nReview plan:\n")
	b.WriteString("1. Files to review:\n")
	for _, file := range changedFiles {
		stat := stats[file]
		b.WriteString(fmt.Sprintf("   - %s (%d added, %d deleted)\n", file, stat.Added, stat.Deleted))
	}
	b.WriteString("2. XP aspects to check:\n")
	b.WriteString("   - Business correctness, security, and data integrity risks\n")
	b.WriteString("   - TDD coverage for changed critical paths\n")
	b.WriteString("   - Simple Design, duplication, coupling, and YAGNI\n")
	b.WriteString("3. Estimated review time: 5-15 minutes\n")
	return b.String()
}

func (a App) runDoctor() int {
	for _, item := range provider.Doctor() {
		status := "missing"
		if item.OK {
			status = "ok"
		}
		fmt.Fprintf(a.Stdout, "%-12s %-8s %s\n", item.Name, status, item.Message)
	}
	return 0
}

func (a App) setProvider(path string, cfg config.Config, providerName string) int {
	if err := config.ValidateProvider(providerName); err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	cfg.Provider = providerName
	if err := config.Save(path, cfg); err != nil {
		fmt.Fprintf(a.Stderr, "Error saving config: %v\n", err)
		return 1
	}
	fmt.Fprintf(a.Stdout, "Default provider set to %s\n", providerName)
	return 0
}

func (a App) update(ctx context.Context) int {
	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+version.GitHubRepo+"/releases/latest", nil)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Error: %v\n", err)
		return 1
	}
	req.Header.Set("User-Agent", "uatiari-cli")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(a.Stderr, "Update check failed: %v\n", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(a.Stderr, "Update check failed: HTTP %d\n", resp.StatusCode)
		return 1
	}
	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Fprintf(a.Stderr, "Update check failed: %v\n", err)
		return 1
	}
	if strings.TrimPrefix(release.TagName, "v") == version.Version {
		fmt.Fprintln(a.Stdout, "You are on the latest version.")
		return 0
	}
	fmt.Fprintf(a.Stdout, "New version available: %s\nInstall it with:\n", release.TagName)
	fmt.Fprintf(a.Stdout, "curl -fsSL https://raw.githubusercontent.com/%s/main/install.sh | bash\n", version.GitHubRepo)
	return 0
}

func approved(reader io.Reader) bool {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func convertStats(stats map[string]gitclient.DiffStat) map[string]report.DiffStat {
	converted := make(map[string]report.DiffStat, len(stats))
	for file, stat := range stats {
		converted[file] = report.DiffStat{Added: stat.Added, Deleted: stat.Deleted}
	}
	return converted
}

func printHelp(writer io.Writer) {
	exe := filepath.Base(os.Args[0])
	if exe == "" {
		exe = "uatiari"
	}
	fmt.Fprintf(writer, `uatiari - XP Code Reviewer

USAGE:
  %s <branch-name> [options]
  %s update
  %s providers doctor
  %s config set provider <name>

OPTIONS:
  --base=<branch>       Base branch for comparison (default: main)
  --skill=<name>        Manually specify skill (e.g., laravel)
  --provider=<name>     kimi, gemini, claude, antigravity, codex
  --format=<name>       json, pretty, markdown (default: json)
  --lang=<name>         pt_BR, en_US (default: pt_BR)
  --version             Show version
  --help, -h            Show help
`, exe, exe, exe, exe)
}
