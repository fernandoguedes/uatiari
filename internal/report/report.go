package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DiffStat struct {
	Added   int
	Deleted int
}

type Result struct {
	BlockingIssues  []Issue        `json:"blocking_issues,omitempty"`
	Warnings        []Warning      `json:"warnings,omitempty"`
	Suggestions     []Suggestion   `json:"suggestions,omitempty"`
	TestAnalysis    TestAnalysis   `json:"test_analysis,omitempty"`
	BusinessLogic   BusinessLogic  `json:"business_logic,omitempty"`
	Overall         Overall        `json:"overall"`
	Metadata        map[string]any `json:"metadata,omitempty"`
	SummaryMarkdown string         `json:"summary_markdown,omitempty"`
	Comments        Comments       `json:"comments,omitempty"`
}

type Issue struct {
	File        string `json:"file,omitempty"`
	Lines       string `json:"lines,omitempty"`
	Category    string `json:"category,omitempty"`
	Issue       string `json:"issue,omitempty"`
	Action      string `json:"action,omitempty"`
	WhyBlocking string `json:"why_blocking,omitempty"`
}

type Warning struct {
	File        string `json:"file,omitempty"`
	Lines       string `json:"lines,omitempty"`
	Category    string `json:"category,omitempty"`
	Issue       string `json:"issue,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
	Effort      string `json:"effort,omitempty"`
	XPPrinciple string `json:"xp_principle,omitempty"`
}

type Suggestion struct {
	File        string `json:"file,omitempty"`
	Lines       string `json:"lines,omitempty"`
	Improvement string `json:"improvement,omitempty"`
	Benefit     string `json:"benefit,omitempty"`
}

type TestAnalysis struct {
	HasTests        bool     `json:"has_tests,omitempty"`
	TestFiles       []string `json:"test_files,omitempty"`
	MissingTestsFor []string `json:"missing_tests_for,omitempty"`
	Notes           string   `json:"notes,omitempty"`
	ProductionLines int      `json:"production_lines,omitempty"`
	TestLines       int      `json:"test_lines,omitempty"`
	Ratio           float64  `json:"ratio,omitempty"`
	Verdict         string   `json:"verdict,omitempty"`
}

type BusinessLogic struct {
	RulesAffected   []string `json:"rules_affected,omitempty"`
	CriticalPath    bool     `json:"critical_path,omitempty"`
	BreakingChanges bool     `json:"breaking_changes,omitempty"`
}

type Overall struct {
	Verdict    string `json:"verdict,omitempty"`
	Reason     string `json:"reason,omitempty"`
	Confidence string `json:"confidence,omitempty"`
}

type Comments struct {
	Blocking    []string `json:"blocking,omitempty"`
	Warnings    []string `json:"warnings,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

func Parse(raw string) (Result, error) {
	content := extractJSON(strings.TrimSpace(raw))
	var result Result
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return Result{}, fmt.Errorf("provider returned invalid JSON: %w", err)
	}
	return result, nil
}

func ApplyTestAnalysis(result *Result, stats map[string]DiffStat) {
	prodLines := 0
	testLines := 0
	var testFiles []string
	for file, stat := range stats {
		if isTestFile(file) {
			testLines += stat.Added
			testFiles = append(testFiles, file)
		} else {
			prodLines += stat.Added
		}
	}
	sort.Strings(testFiles)
	ratio := 0.0
	switch {
	case prodLines > 0:
		ratio = float64(testLines) / float64(prodLines)
	case testLines > 0:
		ratio = 100
	}
	verdict := "N/A"
	switch {
	case ratio >= 1:
		verdict = "EXCELLENT"
	case ratio >= 0.5:
		verdict = "GOOD"
	case ratio > 0:
		verdict = "ACCEPTABLE"
	case prodLines > 0:
		verdict = "MISSING"
	}
	result.TestAnalysis.ProductionLines = prodLines
	result.TestAnalysis.TestLines = testLines
	result.TestAnalysis.Ratio = ratio
	result.TestAnalysis.Verdict = verdict
	if len(result.TestAnalysis.TestFiles) == 0 {
		result.TestAnalysis.TestFiles = testFiles
	}
	result.TestAnalysis.HasTests = testLines > 0 || len(result.TestAnalysis.TestFiles) > 0
}

func EnsureMarkdown(result *Result, lang string) {
	if result.Overall.Verdict == "" {
		result.Overall.Verdict = "UNKNOWN"
	}
	if result.SummaryMarkdown == "" {
		result.SummaryMarkdown = summaryMarkdown(*result, lang)
	}
	result.Comments.Blocking = result.Comments.Blocking[:0]
	for _, item := range result.BlockingIssues {
		result.Comments.Blocking = append(result.Comments.Blocking, blockingMarkdown(item, lang))
	}
	result.Comments.Warnings = result.Comments.Warnings[:0]
	for _, item := range result.Warnings {
		result.Comments.Warnings = append(result.Comments.Warnings, warningMarkdown(item, lang))
	}
	result.Comments.Suggestions = result.Comments.Suggestions[:0]
	for _, item := range result.Suggestions {
		result.Comments.Suggestions = append(result.Comments.Suggestions, suggestionMarkdown(item, lang))
	}
}

func Render(result Result, format string) (string, error) {
	switch format {
	case "", "json":
		bytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return "", err
		}
		return string(bytes) + "\n", nil
	case "markdown":
		var b strings.Builder
		b.WriteString(result.SummaryMarkdown)
		b.WriteString("\n")
		for _, comment := range result.Comments.Blocking {
			b.WriteString("\n")
			b.WriteString(comment)
			b.WriteString("\n")
		}
		for _, comment := range result.Comments.Warnings {
			b.WriteString("\n")
			b.WriteString(comment)
			b.WriteString("\n")
		}
		for _, comment := range result.Comments.Suggestions {
			b.WriteString("\n")
			b.WriteString(comment)
			b.WriteString("\n")
		}
		return b.String(), nil
	case "pretty":
		return pretty(result), nil
	default:
		return "", fmt.Errorf("format %q is not supported", format)
	}
}

func extractJSON(raw string) string {
	if strings.HasPrefix(raw, "```") {
		lines := strings.Split(raw, "\n")
		if len(lines) >= 3 {
			return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
		}
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(raw[start : end+1])
	}
	return raw
}

func isTestFile(file string) bool {
	lower := strings.ToLower(file)
	return strings.Contains(lower, "test") || strings.HasSuffix(lower, "_spec.go") || strings.HasSuffix(lower, "_test.go")
}

func summaryMarkdown(result Result, lang string) string {
	if lang == "en_US" {
		return fmt.Sprintf("## XP Review\n\n**Verdict:** %s\n\n%s", result.Overall.Verdict, result.Overall.Reason)
	}
	return fmt.Sprintf("## Revisão XP\n\n**Veredito:** %s\n\n%s", result.Overall.Verdict, result.Overall.Reason)
}

func blockingMarkdown(item Issue, lang string) string {
	if lang == "en_US" {
		return fmt.Sprintf("### Blocking issue: %s\n\n**File:** `%s:%s`\n\n**Required action:** %s\n\n%s", item.Issue, item.File, item.Lines, item.Action, item.WhyBlocking)
	}
	return fmt.Sprintf("### Bloqueio: %s\n\n**Arquivo:** `%s:%s`\n\n**Ação necessária:** %s\n\n%s", item.Issue, item.File, item.Lines, item.Action, item.WhyBlocking)
}

func warningMarkdown(item Warning, lang string) string {
	if lang == "en_US" {
		return fmt.Sprintf("### Warning: %s\n\n**File:** `%s:%s`\n\n**Suggestion:** %s\n\n**Effort:** %s", item.Issue, item.File, item.Lines, item.Suggestion, item.Effort)
	}
	return fmt.Sprintf("### Atenção: %s\n\n**Arquivo:** `%s:%s`\n\n**Sugestão:** %s\n\n**Esforço:** %s", item.Issue, item.File, item.Lines, item.Suggestion, item.Effort)
}

func suggestionMarkdown(item Suggestion, lang string) string {
	if lang == "en_US" {
		return fmt.Sprintf("### Suggestion: %s\n\n**File:** `%s:%s`\n\n%s", item.Improvement, item.File, item.Lines, item.Benefit)
	}
	return fmt.Sprintf("### Sugestão: %s\n\n**Arquivo:** `%s:%s`\n\n%s", item.Improvement, item.File, item.Lines, item.Benefit)
}

func pretty(result Result) string {
	var b strings.Builder
	b.WriteString("uatiari - XP Code Reviewer\n\n")
	b.WriteString(fmt.Sprintf("Verdict: %s\n", result.Overall.Verdict))
	if result.Overall.Reason != "" {
		b.WriteString(fmt.Sprintf("Reason: %s\n", result.Overall.Reason))
	}
	writeBlockingIssues(&b, result.BlockingIssues)
	writeWarnings(&b, result.Warnings)
	writeSuggestions(&b, result.Suggestions)
	if result.TestAnalysis.Verdict != "" {
		b.WriteString(fmt.Sprintf("\nTests: %s (prod=%d test=%d ratio=%.2f)\n", result.TestAnalysis.Verdict, result.TestAnalysis.ProductionLines, result.TestAnalysis.TestLines, result.TestAnalysis.Ratio))
	}
	return b.String()
}

func writeBlockingIssues(b *strings.Builder, issues []Issue) {
	if len(issues) == 0 {
		b.WriteString("\nBlocking Issues: none\n")
		return
	}
	b.WriteString("\nBlocking Issues:\n")
	for i, item := range issues {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, location(item.File, item.Lines)))
		writeField(b, "Category", item.Category)
		writeField(b, "Issue", item.Issue)
		writeField(b, "Action", item.Action)
		writeField(b, "Why blocking", item.WhyBlocking)
	}
}

func writeWarnings(b *strings.Builder, warnings []Warning) {
	if len(warnings) == 0 {
		b.WriteString("\nWarnings: none\n")
		return
	}
	b.WriteString("\nWarnings:\n")
	for i, item := range warnings {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, location(item.File, item.Lines)))
		writeField(b, "Category", item.Category)
		writeField(b, "Issue", item.Issue)
		writeField(b, "Suggestion", item.Suggestion)
		writeField(b, "Effort", item.Effort)
		writeField(b, "XP principle", item.XPPrinciple)
	}
}

func writeSuggestions(b *strings.Builder, suggestions []Suggestion) {
	if len(suggestions) == 0 {
		b.WriteString("\nSuggestions: none\n")
		return
	}
	b.WriteString("\nSuggestions:\n")
	for i, item := range suggestions {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, location(item.File, item.Lines)))
		writeField(b, "Improvement", item.Improvement)
		writeField(b, "Benefit", item.Benefit)
	}
}

func location(file, lines string) string {
	if file == "" {
		return "(no file)"
	}
	if lines == "" {
		return file
	}
	return file + ":" + lines
}

func writeField(b *strings.Builder, label, value string) {
	if value == "" {
		return
	}
	b.WriteString(fmt.Sprintf("   %s: %s\n", label, value))
}
