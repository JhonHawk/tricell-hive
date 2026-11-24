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
		{"antigravity", "reasoning", Model{"inherit", ""}},
		{"claude", "verifier", Model{"syn-claude-verify", "xhigh"}},
		{"codex", "verifier", Model{"syn-codex-verify", "xhigh"}},
		{"pi", "verifier", Model{"syn-pi/verify", "xhigh"}},
		{"opencode", "verifier", Model{"syn-oc/verify#xhigh", ""}},
		{"grok", "verifier", Model{"syn-grok-verify", ""}},
		{"cursor", "verifier", Model{"inherit", ""}},
		{"antigravity", "verifier", Model{"inherit", ""}},
	}
	for _, c := range cases {
		source, data := resolveSource, resolveRole(c.profile, "")
		if c.profile == "verifier" {
			// The shipped verifier role file exercises the parse path too.
			source, data = syntheticRole(t, "synthetic-verifier")
		}
		profile, m, err := Resolve(source, data, profiles, c.host, nil)
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
	for _, host := range []string{"grok", "cursor", "antigravity"} {
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

// The focused profile resolves like execution on every host, except that
// Codex has its own model. Expected values are literals, so a fault that
// treats both profiles alike on one host cannot cancel itself out.
func TestResolveFocusedProfilePreservesHostAndRoleEffort(t *testing.T) {
	profiles := editProfiles(t, func(host string, models map[string]any) {
		models["focused"] = models["execution"]
		if host == "codex" {
			models["focused"] = map[string]any{"model": "syn-codex-focused", "effort": "medium"}
		}
	})
	cases := []struct {
		host, effort string
		want         Model
	}{
		{"claude", "", Model{"syn-claude-exec", "medium"}},
		{"claude", "low", Model{"syn-claude-exec", "low"}},
		{"claude", "medium", Model{"syn-claude-exec", "medium"}},
		{"codex", "", Model{"syn-codex-focused", "medium"}},
		{"codex", "low", Model{"syn-codex-focused", "low"}},
		{"codex", "medium", Model{"syn-codex-focused", "medium"}},
		{"grok", "", Model{}},
		{"grok", "low", Model{}},
		{"grok", "medium", Model{}},
		{"pi", "", Model{"syn-pi/exec", "medium"}},
		{"pi", "low", Model{"syn-pi/exec", "low"}},
		{"pi", "medium", Model{"syn-pi/exec", "medium"}},
		{"opencode", "", Model{"syn-oc/exec#medium", ""}},
		{"opencode", "low", Model{"syn-oc/exec#low", ""}},
		{"opencode", "medium", Model{"syn-oc/exec#medium", ""}},
		{"cursor", "", Model{"inherit", ""}},
		{"cursor", "low", Model{"inherit", ""}},
		{"cursor", "medium", Model{"inherit", ""}},
		{"antigravity", "", Model{"inherit", ""}},
		{"antigravity", "low", Model{"inherit", ""}},
		{"antigravity", "medium", Model{"inherit", ""}},
	}
	for _, c := range cases {
		profile, got, err := Resolve(resolveSource, resolveRole("focused", c.effort), profiles, c.host, nil)
		if err != nil || profile != "focused" || got != c.want {
			t.Fatalf("Resolve(%s/focused, %q) = %q %+v, %v; want focused %+v", c.host, c.effort, profile, got, err, c.want)
		}
	}
}

func TestReadProfilesValidatesVerifierLikeOtherProfiles(t *testing.T) {
	ok := editProfiles(t, func(host string, models map[string]any) {
		models["verifier"] = map[string]any{"model": "inherit"}
	})
	if _, err := ReadProfiles(ok); err != nil {
		t.Fatalf("four-profile hosts rejected: %v", err)
	}
	// Only Claude carries the bogus effort: on a host that inherits effort,
	// the "inherited effort" rule would reject it and hide the validity check.
	badEffort := editProfiles(t, func(host string, models map[string]any) {
		if host == "claude" {
			models["verifier"] = map[string]any{"model": "m", "effort": "bogus"}
		} else {
			models["verifier"] = map[string]any{"model": "inherit"}
		}
	})
	if _, err := ReadProfiles(badEffort); err == nil || !strings.Contains(err.Error(), "invalid effort for claude") {
		t.Fatalf("verifier with invalid effort: err = %v, want invalid effort for claude", err)
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

func TestResolveVerifierAgainstThreeProfilesNamesHostAndProfile(t *testing.T) {
	profiles := threeProfiles(t)
	if _, err := ReadProfiles(profiles); err != nil {
		t.Fatalf("three-profile snapshot rejected: %v", err)
	}
	for _, host := range allHosts {
		_, _, err := Resolve(resolveSource, resolveRole("verifier", ""), profiles, host, nil)
		if err == nil || !strings.Contains(err.Error(), host) || !strings.Contains(err.Error(), "verifier") {
			t.Fatalf("Resolve(%s) error = %v, want one naming host and verifier", host, err)
		}
	}
}
