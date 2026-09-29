package agents

import (
	"strings"
	"testing"
)

const resolveSource = "content/agents/design/test-agent.md"

func resolveRole(profile, effort string) []byte {
	front := "---\nname: \"test-agent\"\ndescription: \"A role\"\nmodel_profile: \"" + profile + "\"\naccess_profile: \"observe\"\n"
	if effort != "" {
		front += "effort: \"" + effort + "\"\n"
	}
	return []byte(front + "---\nBody\n")
}

func TestResolveReturnsProfileModelAndProfileEffort(t *testing.T) {
	profiles := repositoryProfiles(t)
	cases := []struct {
		host, profile string
		want          Model
	}{
		{"claude", "execution", Model{"sonnet", "high"}},
		{"claude", "reasoning", Model{"opus", "high"}},
		{"claude", "inherit", Model{"inherit", "high"}},
		{"codex", "reasoning", Model{"gpt-6-astra", "medium"}},
		{"grok", "execution", Model{}},
		{"opencode", "execution", Model{"opencode-go/deepseek-v4.1-flash#max", ""}},
		{"cursor", "reasoning", Model{"inherit", ""}},
	}
	for _, c := range cases {
		profile, m, err := Resolve(resolveSource, resolveRole(c.profile, ""), profiles, c.host)
		if err != nil {
			t.Fatalf("Resolve(%s/%s): %v", c.host, c.profile, err)
		}
		if profile != c.profile || m != c.want {
			t.Fatalf("Resolve(%s/%s) = %q %+v, want %q %+v", c.host, c.profile, profile, m, c.profile, c.want)
		}
	}
}

func TestResolveRoleEffortWinsOnlyWhereHostAcceptsEffort(t *testing.T) {
	profiles := repositoryProfiles(t)
	data := resolveRole("execution", "low")
	for _, host := range []string{"claude", "codex", "pi"} {
		_, m, err := Resolve(resolveSource, data, profiles, host)
		if err != nil || m.Effort != "low" {
			t.Fatalf("Resolve(%s) effort = %q, %v; want the role's low", host, m.Effort, err)
		}
	}
	for _, host := range []string{"grok", "opencode", "cursor"} {
		_, m, err := Resolve(resolveSource, data, profiles, host)
		if err != nil || m.Effort != "" {
			t.Fatalf("Resolve(%s) effort = %q, %v; want none", host, m.Effort, err)
		}
	}
}

func TestResolveRejectsBadInput(t *testing.T) {
	profiles := repositoryProfiles(t)
	if _, _, err := Resolve(resolveSource, resolveRole("execution", ""), profiles, "vim"); err == nil || !strings.Contains(err.Error(), "unsupported agent host") {
		t.Fatalf("unknown host error = %v", err)
	}
	if _, _, err := Resolve("not/a/source.md", resolveRole("execution", ""), profiles, "claude"); err == nil {
		t.Fatal("invalid source accepted")
	}
	if _, _, err := Resolve(resolveSource, resolveRole("execution", ""), []byte("{}"), "claude"); err == nil {
		t.Fatal("invalid profiles accepted")
	}
}
