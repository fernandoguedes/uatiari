package review

import (
	"context"

	"github.com/fernandoguedes/uatiari/internal/provider"
	"github.com/fernandoguedes/uatiari/internal/report"
	"github.com/fernandoguedes/uatiari/internal/skills"
)

type Input struct {
	WorkingDir   string
	Branch       string
	Base         string
	ManualSkill  string
	Diff         string
	ChangedFiles []string
	RepoFiles    []string
	DiffStats    map[string]report.DiffStat
}

type Reviewer struct {
	Provider provider.Provider
	Lang     string
}

func (r Reviewer) Execute(ctx context.Context, input Input) (report.Result, error) {
	manager := skills.NewManager()
	manager.Detect(input.ManualSkill, input.RepoFiles, input.ChangedFiles)
	systemPrompt := manager.SystemPrompt(XPSystemPrompt)

	response, err := r.Provider.Run(ctx, provider.Request{
		WorkingDir:   input.WorkingDir,
		SystemPrompt: systemPrompt,
		PlanPrompt:   PlanPrompt,
		ReviewPrompt: ReviewPrompt,
		Diff:         input.Diff,
		ChangedFiles: input.ChangedFiles,
		Lang:         r.Lang,
	})
	if err != nil {
		return report.Result{}, err
	}

	result, err := report.Parse(response.Content)
	if err != nil {
		return report.Result{}, err
	}
	report.ApplyTestAnalysis(&result, input.DiffStats)
	if metadata := manager.Metadata(input.ManualSkill); metadata != nil {
		result.Metadata = metadata
	}
	report.EnsureMarkdown(&result, r.Lang)
	return result, nil
}
