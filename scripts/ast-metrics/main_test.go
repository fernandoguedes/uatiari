package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
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

// TestMainSubprocess is the subprocess entry-point: when GO_TEST_MAIN=1 is set,
// the test binary calls main() directly so coverage is recorded.
func TestMainSubprocess(t *testing.T) {
	if os.Getenv("GO_TEST_MAIN") != "1" {
		t.Skip("subprocess only")
	}
	main()
}

func TestShouldSkipDir(t *testing.T) {
	cases := []struct {
		path string
		name string
		want bool
	}{
		{".", ".", false},
		{".git", ".git", true},
		{"dist", "dist", true},
		{"vendor", "vendor", true},
		{"internal", "internal", false},
		{"scripts/ast-metrics", "ast-metrics", false},
	}
	for _, tc := range cases {
		got := shouldSkipDir(tc.path, tc.name)
		if got != tc.want {
			t.Fatalf("shouldSkipDir(%q, %q) = %v, want %v", tc.path, tc.name, got, tc.want)
		}
	}
}

func TestFuncNameNoReceiver(t *testing.T) {
	src := `package p; func Hello() {}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	fn := f.Decls[0].(*ast.FuncDecl)
	if got := funcName(fn); got != "Hello" {
		t.Fatalf("funcName = %q, want Hello", got)
	}
}

func TestFuncNameWithPointerReceiver(t *testing.T) {
	src := `package p; type T struct{}; func (t *T) Method() {}`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The method is the second decl (first is type)
	fn := f.Decls[1].(*ast.FuncDecl)
	got := funcName(fn)
	if !strings.Contains(got, "T") || !strings.Contains(got, "Method") {
		t.Fatalf("funcName = %q, expected T.Method", got)
	}
}

func TestExprNameIdent(t *testing.T) {
	expr := &ast.Ident{Name: "MyType"}
	if got := exprName(expr); got != "MyType" {
		t.Fatalf("exprName(Ident) = %q, want MyType", got)
	}
}

func TestExprNameStarExpr(t *testing.T) {
	expr := &ast.StarExpr{X: &ast.Ident{Name: "Ptr"}}
	if got := exprName(expr); got != "Ptr" {
		t.Fatalf("exprName(StarExpr) = %q, want Ptr", got)
	}
}

func TestExprNameIndexExpr(t *testing.T) {
	// Generics: Container[int] → IndexExpr{X: Ident("Container")}
	expr := &ast.IndexExpr{X: &ast.Ident{Name: "Container"}}
	if got := exprName(expr); got != "Container" {
		t.Fatalf("exprName(IndexExpr) = %q, want Container", got)
	}
}

func TestExprNameIndexListExpr(t *testing.T) {
	// Generics: Map[K, V] → IndexListExpr{X: Ident("Map")}
	expr := &ast.IndexListExpr{X: &ast.Ident{Name: "Map"}}
	if got := exprName(expr); got != "Map" {
		t.Fatalf("exprName(IndexListExpr) = %q, want Map", got)
	}
}

func TestExprNameUnknown(t *testing.T) {
	// A type that is not handled — should return its Go type string
	expr := &ast.BasicLit{}
	got := exprName(expr)
	if got == "" {
		t.Fatal("exprName(unknown) returned empty string")
	}
}

func TestBranchCountStmts(t *testing.T) {
	src := `package p
func f(x int) {
	if x > 0 {
		for i := 0; i < x; i++ {}
	}
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, 0)
	fn := f.Decls[0].(*ast.FuncDecl)
	count := branchCount(fn.Body)
	if count < 2 {
		t.Fatalf("branchCount = %d, expected at least 2 (if + for)", count)
	}
}

func TestBranchCountBinaryOps(t *testing.T) {
	src := `package p
func f(a, b, c bool) bool {
	return a && b || c
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, 0)
	fn := f.Decls[0].(*ast.FuncDecl)
	count := branchCount(fn.Body)
	if count < 2 {
		t.Fatalf("branchCount for && and || = %d, expected at least 2", count)
	}
}

func TestBranchCountRangeGoDefer(t *testing.T) {
	src := `package p
func f(items []int) {
	ch := make(chan int)
	defer close(ch)
	for _, v := range items {
		go func(n int) { ch <- n }(v)
	}
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, 0)
	fn := f.Decls[0].(*ast.FuncDecl)
	count := branchCount(fn.Body)
	// range(1) + go(1) + defer(1) = at least 3
	if count < 3 {
		t.Fatalf("branchCount for range/go/defer = %d, expected at least 3", count)
	}
}

func TestBranchCountSwitchCases(t *testing.T) {
	src := `package p
func f(x int) string {
	switch x {
	case 1:
		return "one"
	case 2, 3:
		return "two-three"
	default:
		return "other"
	}
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, 0)
	fn := f.Decls[0].(*ast.FuncDecl)
	count := branchCount(fn.Body)
	// case 1 (1) + case 2,3 (2) + default (1) = 4
	if count < 3 {
		t.Fatalf("branchCount for switch cases = %d, expected at least 3", count)
	}
}

func TestBranchCountSelectComm(t *testing.T) {
	src := `package p
func f(ch chan int) {
	select {
	case v := <-ch:
		_ = v
	default:
	}
}`
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "", src, 0)
	fn := f.Decls[0].(*ast.FuncDecl)
	count := branchCount(fn.Body)
	// CommClause for case + default
	if count < 2 {
		t.Fatalf("branchCount for select/comm = %d, expected at least 2", count)
	}
}

func TestAnalyzeSkipsTestFiles(t *testing.T) {
	dir := t.TempDir()
	// A _test.go file with a huge function — should be ignored
	writeFile(t, dir, "big_test.go", `package sample
func TestHuge(t interface{}) {
	_ = 1; _ = 2; _ = 3; _ = 4; _ = 5; _ = 6; _ = 7; _ = 8; _ = 9; _ = 10
}
`)
	findings, err := analyze([]string{dir}, 1, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for test file, got %#v", findings)
	}
}

func TestAnalyzeSkipsNonGoFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "notes.txt", "this is not go code at all")
	findings, err := analyze([]string{dir}, 1, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for .txt file, got %#v", findings)
	}
}

func TestAnalyzeSkipsVendorDir(t *testing.T) {
	dir := t.TempDir()
	vendorDir := filepath.Join(dir, "vendor")
	if err := os.MkdirAll(vendorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, vendorDir, "big.go", `package vendor
func BigFunc(x int) int {
	if x > 0 { x++ }
	if x > 1 { x++ }
	if x > 2 { x++ }
	if x > 3 { x++ }
	return x
}
`)
	findings, err := analyze([]string{dir}, 2, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected vendor to be skipped, got %#v", findings)
	}
}

func TestAnalyzeReturnsErrorOnBadGoFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "broken.go", `this is not valid go code !!!`)
	_, err := analyze([]string{dir}, 80, 20)
	if err == nil {
		t.Fatal("expected error for invalid Go file, got nil")
	}
}

func TestAnalyzeSkipsFuncWithoutBody(t *testing.T) {
	// External function declarations (no body) must not panic
	dir := t.TempDir()
	writeFile(t, dir, "extern.go", `package sample

func external(x int) int  // no body
`)
	findings, err := analyze([]string{dir}, 1, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings for extern func, got %#v", findings)
	}
}

func TestAnalyzeHandlesReceiverFunctions(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "recv.go", `package sample

type S struct{}

func (s *S) BigMethod(x int) int {
	if x > 0 { x++ }
	if x > 1 { x++ }
	if x > 2 { x++ }
	return x
}
`)
	findings, err := analyze([]string{dir}, 3, 1)
	if err != nil {
		t.Fatalf("analyze returned error: %v", err)
	}
	// Should find the method since it has 3 branches (> limit 1)
	if len(findings) == 0 {
		t.Fatal("expected findings for method with many branches, got none")
	}
	if !strings.Contains(findings[0].name, "BigMethod") {
		t.Fatalf("finding name = %q, expected BigMethod", findings[0].name)
	}
}
