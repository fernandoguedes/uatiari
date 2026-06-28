package skills

import "strings"

type Skill interface {
	Name() string
	Detect(repoFiles, changedFiles []string) bool
	PromptAddon() string
	Metadata() map[string]any
}

type Manager struct {
	available []Skill
	active    []Skill
}

func NewManager() *Manager {
	return &Manager{available: []Skill{Laravel{}}}
}

func (m *Manager) Detect(manual string, repoFiles, changedFiles []string) []Skill {
	m.active = nil
	for _, skill := range m.available {
		isActive := false
		if manual != "" && strings.EqualFold(manual, skill.Name()) {
			isActive = true
		} else if manual == "" && skill.Detect(repoFiles, changedFiles) {
			isActive = true
		}
		if isActive {
			m.active = append(m.active, skill)
		}
	}
	return append([]Skill{}, m.active...)
}

func (m *Manager) SystemPrompt(base string) string {
	prompt := base
	for _, skill := range m.active {
		prompt += "\n\n" + skill.PromptAddon()
	}
	return prompt
}

func (m *Manager) Metadata(manual string) map[string]any {
	if len(m.active) == 0 {
		return nil
	}
	details := make([]map[string]any, 0, len(m.active))
	names := make([]string, 0, len(m.active))
	for _, skill := range m.active {
		names = append(names, skill.Name())
		details = append(details, skill.Metadata())
	}
	method := "automatic"
	if manual != "" {
		method = "manual"
	}
	return map[string]any{
		"framework_detected": names[0],
		"skills_applied":     names,
		"detection_method":   method,
		"skill_details":      details,
	}
}

type Laravel struct{}

func (Laravel) Name() string { return "laravel" }

func (Laravel) Detect(repoFiles, changedFiles []string) bool {
	hasPHPChanges := false
	for _, file := range changedFiles {
		if strings.HasSuffix(file, ".php") {
			hasPHPChanges = true
			break
		}
	}
	if !hasPHPChanges {
		return false
	}
	for _, path := range repoFiles {
		if path == "artisan" || path == "composer.json" {
			return true
		}
		if strings.HasPrefix(path, "app/") || strings.HasPrefix(path, "routes/") || strings.HasPrefix(path, "config/") || strings.HasPrefix(path, "database/") {
			return true
		}
	}
	return false
}

func (Laravel) PromptAddon() string {
	return `## LARAVEL SPECIFIC CHECKS

In addition to standard XP rules, enforce Laravel best practices:
- Block SQL injection risks, missing CSRF on forms, and critical N+1 queries in loops.
- Warn on missing eager loading, unindexed query columns, unsafe mass assignment, and weak migrations.
- Suggest small improvements only when ROI is clear.`
}

func (Laravel) Metadata() map[string]any {
	return map[string]any{"name": "laravel", "focus_areas": []string{"performance", "mysql", "security"}}
}
