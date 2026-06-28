package skills

import "testing"

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
