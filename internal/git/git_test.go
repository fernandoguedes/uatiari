package git

import "testing"

func TestParseNumstat(t *testing.T) {
	stats := ParseNumstat("10\t5\tsrc/main.go\n-\t-\timage.png\n20\t0\ttests/main_test.go\n")

	if stats["src/main.go"].Added != 10 || stats["src/main.go"].Deleted != 5 {
		t.Fatalf("src/main.go stats = %#v", stats["src/main.go"])
	}
	if stats["image.png"].Added != 0 || stats["image.png"].Deleted != 0 {
		t.Fatalf("image.png stats = %#v", stats["image.png"])
	}
	if stats["tests/main_test.go"].Added != 20 {
		t.Fatalf("tests/main_test.go stats = %#v", stats["tests/main_test.go"])
	}
}
