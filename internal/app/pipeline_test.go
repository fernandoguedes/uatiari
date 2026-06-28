package app

import (
	"os"
	"strings"
	"testing"
)

func TestCIWorkflowUsesGoToolchain(t *testing.T) {
	content := readWorkflow(t, "../../.github/workflows/ci.yml")

	for _, expected := range []string{"actions/setup-go@v5", "go test ./...", "go vet ./...", "gofmt -l ."} {
		if !strings.Contains(content, expected) {
			t.Fatalf("ci workflow missing %q", expected)
		}
	}
	for _, forbidden := range []string{"setup-python", "poetry", "pytest", "ruff", "black"} {
		if strings.Contains(strings.ToLower(content), forbidden) {
			t.Fatalf("ci workflow still contains %q", forbidden)
		}
	}
}

func TestReleaseWorkflowBuildsGoAssets(t *testing.T) {
	content := readWorkflow(t, "../../.github/workflows/release.yml")

	for _, expected := range []string{"actions/setup-go@v5", "go build", "uatiari-linux-x64", "uatiari-macos-x64", "uatiari-macos-arm64"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("release workflow missing %q", expected)
		}
	}
	for _, forbidden := range []string{"setup-python", "poetry", "pyinstaller"} {
		if strings.Contains(strings.ToLower(content), forbidden) {
			t.Fatalf("release workflow still contains %q", forbidden)
		}
	}
}

func readWorkflow(t *testing.T, path string) string {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	return string(bytes)
}
