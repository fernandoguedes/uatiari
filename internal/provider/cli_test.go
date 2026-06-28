package provider

import "testing"

func TestCodexAdapterKeepsPromptSentinelLast(t *testing.T) {
	cli, err := NewCLI("codex")
	if err != nil {
		t.Fatalf("NewCLI returned error: %v", err)
	}

	args := cli.argsFor(Request{WorkingDir: "/tmp/repo"})
	if args[len(args)-1] != "-" {
		t.Fatalf("last arg = %q, want prompt sentinel", args[len(args)-1])
	}
	for i, arg := range args {
		if arg == "-C" && i > len(args)-3 {
			t.Fatalf("-C appears too late in args: %#v", args)
		}
	}
}
