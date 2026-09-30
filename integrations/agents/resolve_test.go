package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		{"codex", "reasoning", Model{"gpt-6.1-sol", "high"}},
		{"grok", "execution", Model{}},
		{"opencode", "execution", Model{"github-copilot/gpt-6.1-sol#medium", ""}},
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
	for _, host := range []string{"grok", "cursor"} {
		_, m, err := Resolve(resolveSource, data, profiles, host)
		if err != nil || m.Effort != "" {
			t.Fatalf("Resolve(%s) effort = %q, %v; want none", host, m.Effort, err)
		}
	}
	// OpenCode carries the role's effort as the model variant.
	if _, m, err := Resolve(resolveSource, data, profiles, "opencode"); err != nil || m != (Model{"github-copilot/gpt-6.1-sol#low", ""}) {
		t.Fatalf("Resolve(opencode) = %+v, %v; want the role's low variant", m, err)
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

// editProfiles applies edit to the models object of every host in the
// repository profiles and returns the re-encoded JSON.
func editProfiles(t *testing.T, edit func(host string, models map[string]any)) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(repositoryProfiles(t), &doc); err != nil {
		t.Fatal(err)
	}
	for host, h := range doc["hosts"].(map[string]any) {
		edit(host, h.(map[string]any)["models"].(map[string]any))
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func threeProfiles(t *testing.T) []byte {
	return editProfiles(t, func(_ string, models map[string]any) { delete(models, "verifier") })
}

func TestReadProfilesValidatesVerifierLikeOtherProfiles(t *testing.T) {
	ok := editProfiles(t, func(host string, models map[string]any) {
		models["verifier"] = map[string]any{"model": "inherit"}
	})
	if _, err := ReadProfiles(ok); err != nil {
		t.Fatalf("four-profile hosts rejected: %v", err)
	}
	badEffort := editProfiles(t, func(host string, models map[string]any) {
		models["verifier"] = map[string]any{"model": "m", "effort": "bogus"}
	})
	if _, err := ReadProfiles(badEffort); err == nil {
		t.Fatal("verifier with invalid effort accepted")
	}
	grokEffort := editProfiles(t, func(host string, models map[string]any) {
		if host == "grok" {
			models["verifier"] = map[string]any{"model": "grok-4.6", "effort": "high"}
		} else {
			models["verifier"] = map[string]any{"model": "inherit"}
		}
	})
	if _, err := ReadProfiles(grokEffort); err == nil {
		t.Fatal("verifier effort on grok accepted")
	}
	multiline := editProfiles(t, func(host string, models map[string]any) {
		models["verifier"] = map[string]any{"model": "a\nb"}
	})
	if _, err := ReadProfiles(multiline); err == nil {
		t.Fatal("multiline verifier model accepted")
	}
}

func TestReadProfilesStillAcceptsThreeProfileHosts(t *testing.T) {
	if _, err := ReadProfiles(threeProfiles(t)); err != nil {
		t.Fatalf("three-profile snapshot rejected: %v", err)
	}
}

func TestParseAcceptsVerifierModelProfile(t *testing.T) {
	r, err := Parse(resolveSource, resolveRole("verifier", ""))
	if err != nil || r.ModelProfile != "verifier" {
		t.Fatalf("Parse(verifier) = %+v, %v", r, err)
	}
}

func TestResolveVerifierAgainstThreeProfilesNamesHostAndProfile(t *testing.T) {
	profiles := threeProfiles(t)
	for _, host := range []string{"claude", "codex", "grok", "pi", "opencode", "cursor"} {
		_, _, err := Resolve(resolveSource, resolveRole("verifier", ""), profiles, host)
		if err == nil || !strings.Contains(err.Error(), host) || !strings.Contains(err.Error(), "verifier") {
			t.Fatalf("Resolve(%s) error = %v, want one naming host and verifier", host, err)
		}
	}
}

func TestRepositoryProfilesDeclareVerifierOnAllHosts(t *testing.T) {
	p, err := ReadProfiles(repositoryProfiles(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Hosts) != 6 {
		t.Fatalf("hosts = %d, want 6", len(p.Hosts))
	}
	for host, h := range p.Hosts {
		if _, ok := h.Models["verifier"]; !ok {
			t.Errorf("%s lacks a verifier profile", host)
		}
	}
}

func TestResolveVerifyTaskRoleAgainstRepositoryProfiles(t *testing.T) {
	source := "content/agents/quality/hive-verify-task.md"
	data, err := os.ReadFile(filepath.Join("..", "..", source))
	if err != nil {
		t.Fatal(err)
	}
	profiles := repositoryProfiles(t)
	cases := []struct{ host, model, effort string }{
		{"claude", "opus", "high"},
		{"codex", "gpt-6-astra", "high"},
		{"pi", "xai/grok-4.7", "high"},
		{"opencode", "github-copilot/claude-opus-5.5#high", ""},
		{"grok", "grok-4.6", ""},
		{"cursor", "inherit", ""},
	}
	for _, c := range cases {
		profile, m, err := Resolve(source, data, profiles, c.host)
		if err != nil || profile != "verifier" || m.Model != c.model || m.Effort != c.effort {
			t.Errorf("Resolve(%s) = %q, %+v, %v; want verifier, %s/%q", c.host, profile, m, err, c.model, c.effort)
		}
	}
}
