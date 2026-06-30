package skills

import (
	"strings"
	"testing"
)

func TestDetectsLaravelManually(t *testing.T) {
	manager := NewManager()

	active := manager.Detect("laravel", nil, nil)

	if len(active) != 1 || active[0].Name() != "laravel" {
		t.Fatalf("active skills = %#v, want laravel", active)
	}
}

func TestDetectsLaravelAutomatically(t *testing.T) {
	manager := NewManager()

	active := manager.Detect("", []string{"composer.json", "artisan"}, []string{"app/User.php"})

	if len(active) != 1 || active[0].Name() != "laravel" {
		t.Fatalf("active skills = %#v, want laravel", active)
	}
}

func TestDoesNotDetectLaravelWithoutPHPChanges(t *testing.T) {
	manager := NewManager()

	active := manager.Detect("", []string{"composer.json", "artisan"}, []string{"README.md"})

	if len(active) != 0 {
		t.Fatalf("active skills = %#v, want none", active)
	}
}

func TestSystemPromptAppendsSkillAddon(t *testing.T) {
	m := NewManager()
	m.Detect("laravel", nil, nil)
	prompt := m.SystemPrompt("base prompt")
	if !strings.Contains(prompt, "base prompt") {
		t.Fatal("SystemPrompt missing base prompt")
	}
	if !strings.Contains(prompt, "LARAVEL") {
		t.Fatal("SystemPrompt missing laravel addon")
	}
}

func TestMetadataManualDetection(t *testing.T) {
	m := NewManager()
	m.Detect("laravel", nil, nil)
	meta := m.Metadata("laravel")
	if meta == nil {
		t.Fatal("Metadata returned nil")
	}
	if meta["detection_method"] != "manual" {
		t.Fatalf("detection_method = %v, want manual", meta["detection_method"])
	}
}

func TestMetadataAutomaticDetection(t *testing.T) {
	m := NewManager()
	m.Detect("", []string{"artisan"}, []string{"app/User.php"})
	meta := m.Metadata("")
	if meta["detection_method"] != "automatic" {
		t.Fatalf("detection_method = %v, want automatic", meta["detection_method"])
	}
}

func TestMetadataNoActiveSkills(t *testing.T) {
	m := NewManager()
	m.Detect("", nil, nil)
	if m.Metadata("") != nil {
		t.Fatal("Metadata should be nil when no skills active")
	}
}

func TestLaravelPromptAddon(t *testing.T) {
	addon := Laravel{}.PromptAddon()
	if !strings.Contains(addon, "LARAVEL") {
		t.Fatalf("PromptAddon = %q, expected LARAVEL", addon)
	}
}

func TestLaravelMetadata(t *testing.T) {
	meta := Laravel{}.Metadata()
	if meta["name"] != "laravel" {
		t.Fatalf("Metadata name = %v, want laravel", meta["name"])
	}
}
