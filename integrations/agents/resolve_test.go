package agents

import (
	"encoding/json"
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
	profiles := syntheticProfiles(t)
	cases := []struct {
		host, profile string
		want          Model
	}{
		{"claude", "execution", Model{"syn-claude-exec", "medium"}},
		{"claude", "reasoning", Model{"syn-claude-reason", "high"}},
		{"claude", "inherit", Model{"inherit", "low"}},
		{"codex", "reasoning", Model{"syn-codex-reason", "high"}},
		{"codex", "inherit", Model{"", "low"}},
		{"grok", "execution", Model{}},
		{"opencode", "execution", Model{"syn-oc/exec#medium", ""}},
		{"opencode", "inherit", Model{"syn-oc/inherit#max", ""}},
		{"cursor", "reasoning", Model{"inherit", ""}},
	}
	for _, c := range cases {
		profile, m, err := Resolve(resolveSource, resolveRole(c.profile, ""), profiles, c.host, nil)
		if err != nil {
			t.Fatalf("Resolve(%s/%s): %v", c.host, c.profile, err)
		}
		if profile != c.profile || m != c.want {
			t.Fatalf("Resolve(%s/%s) = %q %+v, want %q %+v", c.host, c.profile, profile, m, c.profile, c.want)
		}
	}
}

func TestResolveRoleEffortWinsOnlyWhereHostAcceptsEffort(t *testing.T) {
	profiles := syntheticProfiles(t)
	data := resolveRole("execution", "low")
	for _, host := range []string{"claude", "codex", "pi"} {
		_, m, err := Resolve(resolveSource, data, profiles, host, nil)
		if err != nil || m.Effort != "low" {
			t.Fatalf("Resolve(%s) effort = %q, %v; want the role's low", host, m.Effort, err)
		}
	}
	for _, host := range []string{"grok", "cursor"} {
		_, m, err := Resolve(resolveSource, data, profiles, host, nil)
		if err != nil || m.Effort != "" {
			t.Fatalf("Resolve(%s) effort = %q, %v; want none", host, m.Effort, err)
		}
	}
	// OpenCode carries the role's effort as the model variant.
	if _, m, err := Resolve(resolveSource, data, profiles, "opencode", nil); err != nil || m != (Model{"syn-oc/exec#low", ""}) {
		t.Fatalf("Resolve(opencode) = %+v, %v; want the role's low variant", m, err)
	}
}

func TestResolveRejectsBadInput(t *testing.T) {
	profiles := syntheticProfiles(t)
	if _, _, err := Resolve(resolveSource, resolveRole("execution", ""), profiles, "vim", nil); err == nil || !strings.Contains(err.Error(), "unsupported agent host") {
		t.Fatalf("unknown host error = %v", err)
	}
	if _, _, err := Resolve("not/a/source.md", resolveRole("execution", ""), profiles, "claude", nil); err == nil {
		t.Fatal("invalid source accepted")
	}
	if _, _, err := Resolve(resolveSource, resolveRole("execution", ""), []byte("{}"), "claude", nil); err == nil {
		t.Fatal("invalid profiles accepted")
	}
}

// editProfiles applies edit to the models object of every host in the
// synthetic profiles and returns the re-encoded JSON.
func editProfiles(t *testing.T, edit func(host string, models map[string]any)) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(syntheticProfiles(t), &doc); err != nil {
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
	for _, host := range allHosts {
		_, _, err := Resolve(resolveSource, resolveRole("verifier", ""), profiles, host, nil)
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

func TestResolveVerifierRoleAgainstSyntheticProfiles(t *testing.T) {
	source, data := syntheticRole(t, "synthetic-verifier")
	profiles := syntheticProfiles(t)
	cases := []struct{ host, model, effort string }{
		{"claude", "syn-claude-verify", "xhigh"},
		{"codex", "syn-codex-verify", "xhigh"},
		{"pi", "syn-pi/verify", "xhigh"},
		{"opencode", "syn-oc/verify#xhigh", ""},
		{"grok", "syn-grok-verify", ""},
		{"cursor", "inherit", ""},
	}
	for _, c := range cases {
		profile, m, err := Resolve(source, data, profiles, c.host, nil)
		if err != nil || profile != "verifier" || m.Model != c.model || m.Effort != c.effort {
			t.Errorf("Resolve(%s) = %q, %+v, %v; want verifier, %s/%q", c.host, profile, m, err, c.model, c.effort)
		}
	}
}
