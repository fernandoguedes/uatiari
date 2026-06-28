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
