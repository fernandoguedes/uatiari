package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzePassesSmallFunctions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "small.go", `package sample

func small(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
`)

	findings, err := analyze([]string{dir}, 20, 5)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("findings = %#v, want none", findings)
	}
}

func TestAnalyzeReportsFunctionsOverLimits(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "large.go", `package sample

func large(a int) int {
	if a > 0 {
		a++
	}
	if a > 1 {
		a++
	}
	return a
}
`)

	findings, err := analyze([]string{dir}, 5, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("len(findings) = %d, want 2: %#v", len(findings), findings)
	}
}

func writeFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
