package report

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseExtractsJSONFromMarkdownFence(t *testing.T) {
	input := "```json\n{\"overall\":{\"verdict\":\"APPROVE\",\"reason\":\"ok\",\"confidence\":\"HIGH\"}}\n```"

	result, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if result.Overall.Verdict != "APPROVE" {
		t.Fatalf("verdict = %q, want APPROVE", result.Overall.Verdict)
	}
}

func TestParseRejectsInvalidJSON(t *testing.T) {
	_, err := Parse("not json")
	if err == nil {
		t.Fatal("Parse returned nil error for invalid JSON")
	}
}

func TestApplyTestAnalysisOverridesLineCounts(t *testing.T) {
	result := Result{
		TestAnalysis: TestAnalysis{
			Notes: "kept",
		},
	}
	stats := map[string]DiffStat{
		"src/payment.go":        {Added: 10, Deleted: 2},
		"tests/payment_test.go": {Added: 8, Deleted: 0},
	}

	ApplyTestAnalysis(&result, stats)

	if result.TestAnalysis.ProductionLines != 10 {
		t.Fatalf("production lines = %d, want 10", result.TestAnalysis.ProductionLines)
	}
	if result.TestAnalysis.TestLines != 8 {
		t.Fatalf("test lines = %d, want 8", result.TestAnalysis.TestLines)
	}
	if result.TestAnalysis.Verdict != "GOOD" {
		t.Fatalf("verdict = %q, want GOOD", result.TestAnalysis.Verdict)
	}
	if result.TestAnalysis.Notes != "kept" {
		t.Fatalf("notes = %q, want kept", result.TestAnalysis.Notes)
	}
}

func TestEnsureMarkdownAddsCopyableComments(t *testing.T) {
	result := Result{
		Overall: Overall{Verdict: "REQUEST_CHANGES", Reason: "needs validation", Confidence: "HIGH"},
		BlockingIssues: []Issue{{
			File:   "src/payment.go",
			Lines:  "10-12",
			Issue:  "Allows negative amount",
			Action: "Validate amount > 0",
		}},
	}

	EnsureMarkdown(&result, "pt_BR")

	if !strings.Contains(result.SummaryMarkdown, "REQUEST_CHANGES") {
		t.Fatalf("summary markdown missing verdict: %q", result.SummaryMarkdown)
	}
	if len(result.Comments.Blocking) != 1 {
		t.Fatalf("blocking comments len = %d, want 1", len(result.Comments.Blocking))
	}
	if !strings.Contains(result.Comments.Blocking[0], "src/payment.go") {
		t.Fatalf("blocking comment missing file: %q", result.Comments.Blocking[0])
	}
}

func TestRenderJSONIsValid(t *testing.T) {
	result := Result{Overall: Overall{Verdict: "APPROVE"}}
	out, err := Render(result, "json")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(out), &decoded); err != nil {
		t.Fatalf("json output is invalid: %v\n%s", err, out)
	}
}

func TestEnsureMarkdownEnglish(t *testing.T) {
	result := Result{
		Overall: Overall{Verdict: "APPROVE", Reason: "all good"},
		BlockingIssues: []Issue{{File: "foo.go", Lines: "1", Issue: "X", Action: "Do Y"}},
		Warnings:       []Warning{{File: "bar.go", Lines: "2", Issue: "W", Suggestion: "S", Effort: "5min"}},
		Suggestions:    []Suggestion{{File: "baz.go", Lines: "3", Improvement: "I", Benefit: "B"}},
	}
	EnsureMarkdown(&result, "en_US")
	if !strings.Contains(result.SummaryMarkdown, "XP Review") {
		t.Fatalf("en_US summary missing 'XP Review': %q", result.SummaryMarkdown)
	}
	if len(result.Comments.Warnings) != 1 || !strings.Contains(result.Comments.Warnings[0], "Warning") {
		t.Fatalf("en_US warning comment = %q", result.Comments.Warnings)
	}
	if len(result.Comments.Suggestions) != 1 || !strings.Contains(result.Comments.Suggestions[0], "Suggestion") {
		t.Fatalf("en_US suggestion comment = %q", result.Comments.Suggestions)
	}
}

func TestEnsureMarkdownSetsDefaultVerdict(t *testing.T) {
	result := Result{}
	EnsureMarkdown(&result, "pt_BR")
	if result.Overall.Verdict != "UNKNOWN" {
		t.Fatalf("verdict = %q, want UNKNOWN", result.Overall.Verdict)
	}
}

func TestRenderMarkdownContainsSummary(t *testing.T) {
	result := Result{
		Overall:        Overall{Verdict: "APPROVE", Reason: "ok"},
		SummaryMarkdown: "## Revisão XP\n\n**Veredito:** APPROVE\n\nok",
		Comments: Comments{
			Blocking:    []string{"block comment"},
			Warnings:    []string{"warning comment"},
			Suggestions: []string{"suggestion comment"},
		},
	}
	out, err := Render(result, "markdown")
	if err != nil {
		t.Fatalf("Render markdown error: %v", err)
	}
	for _, want := range []string{"Revisão XP", "block comment", "warning comment", "suggestion comment"} {
		if !strings.Contains(out, want) {
			t.Fatalf("markdown output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderUnsupportedFormatReturnsError(t *testing.T) {
	_, err := Render(Result{}, "xml")
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}

func TestRenderDefaultFormatIsJSON(t *testing.T) {
	out, err := Render(Result{Overall: Overall{Verdict: "APPROVE"}}, "")
	if err != nil {
		t.Fatalf("Render '' error: %v", err)
	}
	if !strings.Contains(out, "APPROVE") {
		t.Fatalf("default format output missing APPROVE: %q", out)
	}
}

func TestEnsureMarkdownPtBRWarningsAndSuggestions(t *testing.T) {
	result := Result{
		Overall:     Overall{Verdict: "APPROVE"},
		Warnings:    []Warning{{File: "a.go", Lines: "1", Issue: "W", Suggestion: "S", Effort: "5min"}},
		Suggestions: []Suggestion{{File: "b.go", Lines: "2", Improvement: "I", Benefit: "B"}},
	}
	EnsureMarkdown(&result, "pt_BR")
	if len(result.Comments.Warnings) != 1 || !strings.Contains(result.Comments.Warnings[0], "Atenção") {
		t.Fatalf("pt_BR warning = %q, expected 'Atenção'", result.Comments.Warnings)
	}
	if len(result.Comments.Suggestions) != 1 || !strings.Contains(result.Comments.Suggestions[0], "Sugestão") {
		t.Fatalf("pt_BR suggestion = %q, expected 'Sugestão'", result.Comments.Suggestions)
	}
}

func TestParseExtractsRawBraces(t *testing.T) {
	// extractJSON fallback: no ``` prefix, but valid JSON with surrounding text
	input := `some text {"overall":{"verdict":"APPROVE","reason":"ok","confidence":"HIGH"}} more text`
	result, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.Overall.Verdict != "APPROVE" {
		t.Fatalf("verdict = %q, want APPROVE", result.Overall.Verdict)
	}
}

func TestApplyTestAnalysisExcellentRatio(t *testing.T) {
	stats := map[string]DiffStat{
		"src/foo.go":      {Added: 5},
		"src/foo_test.go": {Added: 6},
	}
	result := Result{}
	ApplyTestAnalysis(&result, stats)
	if result.TestAnalysis.Verdict != "EXCELLENT" {
		t.Fatalf("verdict = %q, want EXCELLENT", result.TestAnalysis.Verdict)
	}
}

func TestApplyTestAnalysisMissingTests(t *testing.T) {
	stats := map[string]DiffStat{
		"src/foo.go": {Added: 10},
	}
	result := Result{}
	ApplyTestAnalysis(&result, stats)
	if result.TestAnalysis.Verdict != "MISSING" {
		t.Fatalf("verdict = %q, want MISSING", result.TestAnalysis.Verdict)
	}
}

func TestPrettyWithNoIssues(t *testing.T) {
	result := Result{Overall: Overall{Verdict: "APPROVE"}}
	out, err := Render(result, "pretty")
	if err != nil {
		t.Fatalf("Render pretty error: %v", err)
	}
	if !strings.Contains(out, "none") {
		t.Fatalf("pretty with no issues should show 'none': %q", out)
	}
}

func TestRenderPrettyIncludesActionableDetails(t *testing.T) {
	result := Result{
		Overall: Overall{
			Verdict: "REQUEST_CHANGES",
			Reason:  "needs provider contract fix",
		},
		BlockingIssues: []Issue{{
			File:     "internal/provider/cli.go",
			Lines:    "28",
			Category: "CONTRACT",
			Issue:    "Codex adapter uses an unsupported flag",
			Action:   "Remove --ask-for-approval from codex exec args",
		}},
		Warnings: []Warning{{
			File:       "internal/app/app.go",
			Lines:      "82-90",
			Category:   "VALIDATION",
			Issue:      "Provider auth is only checked after approval",
			Suggestion: "Validate provider readiness before asking for execution approval",
			Effort:     "20min",
		}},
		Suggestions: []Suggestion{{
			File:        "README.md",
			Lines:       "31-45",
			Improvement: "Document provider authentication requirements",
			Benefit:     "Users can configure their CLI before running a review",
		}},
		TestAnalysis: TestAnalysis{
			Verdict:         "ACCEPTABLE",
			ProductionLines: 10,
			TestLines:       4,
			Ratio:           0.4,
		},
	}

	out, err := Render(result, "pretty")
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	for _, want := range []string{
		"Blocking Issues:",
		"internal/provider/cli.go:28",
		"Category: CONTRACT",
		"Action: Remove --ask-for-approval from codex exec args",
		"Warnings:",
		"Suggestion: Validate provider readiness before asking for execution approval",
		"Effort: 20min",
		"Suggestions:",
		"Benefit: Users can configure their CLI before running a review",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("pretty output missing %q:\n%s", want, out)
		}
	}
}
