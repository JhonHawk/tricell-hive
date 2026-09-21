package target

import "testing"

func TestCatalogSeparatesSkillsAndAgents(t *testing.T) {
	sources := []string{"content/agents/review/review-code.md", "integrations/agent-profiles.json", "content/skills/report/SKILL.md", "content/skills/report/assets/layout.html"}
	base := []Target{{Kind: "skill", Path: "/shared/skills/workspace-conventions/SKILL.md"}}
	got := ExpandSkills(base, sources)
	if len(got) != 2 || got[1].Path != "/shared/skills/report/assets/layout.html" {
		t.Fatalf("skill expansion: %#v", got)
	}
	got = ExpandAgents(got, sources, "/native/agents", ".toml", Target{Host: "codex", Scope: "user", Context: "/home"})
	if len(got) != 3 || got[2].Path != "/native/agents/review-code.toml" || got[2].Kind != "agent" || got[2].Host != "codex" {
		t.Fatalf("agent expansion: %#v", got)
	}
}
