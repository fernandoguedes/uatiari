package review

import (
	"context"
	"testing"

	"github.com/fernandoguedes/uatiari/internal/provider"
	"github.com/fernandoguedes/uatiari/internal/report"
)

type fakeProvider struct {
	content string
}

func (f fakeProvider) Run(context.Context, provider.Request) (provider.Response, error) {
	return provider.Response{Content: f.content}, nil
}

func TestExecuteParsesProviderJSONAndAddsMarkdown(t *testing.T) {
	r := Reviewer{
		Provider: fakeProvider{content: `{"overall":{"verdict":"APPROVE","reason":"ok","confidence":"HIGH"}}`},
		Lang:     "pt_BR",
	}

	result, err := r.Execute(context.Background(), Input{
		WorkingDir:   ".",
		Diff:         "diff --git",
		ChangedFiles: []string{"src/main.go"},
		DiffStats: map[string]report.DiffStat{
			"src/main.go": {Added: 4},
		},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Overall.Verdict != "APPROVE" {
		t.Fatalf("verdict = %q, want APPROVE", result.Overall.Verdict)
	}
	if result.TestAnalysis.Verdict != "MISSING" {
		t.Fatalf("test verdict = %q, want MISSING", result.TestAnalysis.Verdict)
	}
	if result.SummaryMarkdown == "" {
		t.Fatal("summary markdown was not populated")
	}
}
