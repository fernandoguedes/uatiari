package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type finding struct {
	file    string
	name    string
	metric  string
	actual  int
	maximum int
}

func main() {
	maxFuncLines := flag.Int("max-func-lines", 80, "maximum physical lines per function")
	maxBranches := flag.Int("max-branches", 20, "maximum branch points per function")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	findings, err := analyze(roots, *maxFuncLines, *maxBranches)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ast-metrics: %v\n", err)
		os.Exit(2)
	}
	if len(findings) > 0 {
		for _, finding := range findings {
			fmt.Fprintf(
				os.Stderr,
				"%s: %s has %s %d, max %d\n",
				finding.file,
				finding.name,
				finding.metric,
				finding.actual,
				finding.maximum,
			)
		}
		os.Exit(1)
	}
}

func analyze(roots []string, maxFuncLines int, maxBranches int) ([]finding, error) {
	fset := token.NewFileSet()
	var findings []finding

	for _, root := range roots {
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if shouldSkipDir(path, entry.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			file, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				name := funcName(fn)
				lines := fset.Position(fn.End()).Line - fset.Position(fn.Pos()).Line + 1
				if lines > maxFuncLines {
					findings = append(findings, finding{
						file: path, name: name, metric: "lines", actual: lines, maximum: maxFuncLines,
					})
				}
				branches := branchCount(fn.Body)
				if branches > maxBranches {
					findings = append(findings, finding{
						file: path, name: name, metric: "branches", actual: branches, maximum: maxBranches,
					})
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return findings, nil
}

func shouldSkipDir(path string, name string) bool {
	if path == "." {
		return false
	}
	switch name {
	case ".git", "dist", "vendor":
		return true
	default:
		return false
	}
}

func funcName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return fmt.Sprintf("%s.%s", exprName(fn.Recv.List[0].Type), fn.Name.Name)
}

func exprName(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.StarExpr:
		return exprName(expr.X)
	case *ast.IndexExpr:
		return exprName(expr.X)
	case *ast.IndexListExpr:
		return exprName(expr.X)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func branchCount(node ast.Node) int {
	count := 0
	ast.Inspect(node, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.GoStmt, *ast.DeferStmt:
			count++
		case *ast.CaseClause:
			count += len(node.List)
			if len(node.List) == 0 {
				count++
			}
		case *ast.CommClause:
			count++
		case *ast.BinaryExpr:
			if node.Op.String() == "&&" || node.Op.String() == "||" {
				count++
			}
		}
		return true
	})
	return count
}
