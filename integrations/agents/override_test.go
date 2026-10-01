package agents

import (
	"strings"
	"testing"
)

func TestValidateOverrideTable(t *testing.T) {
	long := strings.Repeat("a", 200)
	cases := []struct {
		name string
		host string
		o    ModelOverride
		want string // substring of the error; empty means accepted
	}{
		{"model and effort", "claude", ModelOverride{"opus", "max"}, ""},
		{"effort only keeps the release model", "claude", ModelOverride{Effort: "high"}, ""},
		{"model only", "grok", ModelOverride{Model: "grok-4"}, ""},
		{"provider slash model", "opencode", ModelOverride{Model: "anthropic/claude-opus-4-6"}, ""},
		{"bracketed effort in the name", "claude", ModelOverride{Model: "claude-opus[effort=high]"}, ""},
		{"version punctuation", "codex", ModelOverride{Model: "gpt-5.5:v1@x+y=z"}, ""},
		{"200 characters", "claude", ModelOverride{Model: long}, ""},
		{"ultra on claude is kept", "claude", ModelOverride{Effort: "ultra"}, ""},
		{"ultra on codex is kept", "codex", ModelOverride{Effort: "ultra"}, ""},
		{"ultra on opencode is kept", "opencode", ModelOverride{Effort: "ultra"}, ""},
		{"ultra on pi is not a pi-subagents level", "pi", ModelOverride{Effort: "ultra"}, `"ultra" is not accepted by pi`},
		{"xhigh on pi", "pi", ModelOverride{Effort: "xhigh"}, ""},
		{"empty override", "claude", ModelOverride{}, "needs a model or an effort"},
		{"201 characters", "claude", ModelOverride{Model: long + "a"}, "at most 200"},
		{"space", "claude", ModelOverride{Model: "claude opus"}, "model"},
		{"double quote", "claude", ModelOverride{Model: `a"b`}, "model"},
		{"single quote", "claude", ModelOverride{Model: "a'b"}, "model"},
		{"hash", "opencode", ModelOverride{Model: "a/b#high"}, "model"},
		{"backslash", "claude", ModelOverride{Model: `a\b`}, "model"},
		{"newline", "claude", ModelOverride{Model: "a\nb"}, "model"},
		{"NUL", "claude", ModelOverride{Model: "a\x00b"}, "model"},
		{"right-to-left override U+202E", "claude", ModelOverride{Model: "a\u202eb"}, "model"},
		{"non-ASCII letter", "claude", ModelOverride{Model: "modèle"}, "model"},
		{"effort on grok", "grok", ModelOverride{Effort: "high"}, "grok does not support effort"},
		{"effort on cursor", "cursor", ModelOverride{Model: "m", Effort: "high"}, "cursor does not support effort"},
		{"unknown effort", "claude", ModelOverride{Effort: "extreme"}, "effort"},
		{"effort casing", "claude", ModelOverride{Effort: "High"}, "effort"},
		{"unknown host", "vim", ModelOverride{Model: "m"}, "unsupported agent host"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateOverride(c.host, c.o)
			if c.want == "" {
				if err != nil {
					t.Fatalf("ValidateOverride(%s, %+v) = %v, want accepted", c.host, c.o, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("ValidateOverride(%s, %+v) = %v, want an error containing %q", c.host, c.o, err, c.want)
			}
		})
	}
}

func TestResolveOverrideReplacesModelAndEffort(t *testing.T) {
	profiles := syntheticProfiles(t)
	cases := []struct {
		name, host, profile string
		role                string // role effort
		o                   ModelOverride
		want                Model
	}{
		{"claude model and effort", "claude", "execution", "", ModelOverride{"opus", "max"}, Model{"opus", "max"}},
		{"claude effort only keeps the release model", "claude", "reasoning", "", ModelOverride{Effort: "low"}, Model{"syn-claude-reason", "low"}},
		{"claude model only keeps the release effort", "claude", "reasoning", "", ModelOverride{Model: "sonnet"}, Model{"sonnet", "high"}},
		{"override beats the role effort", "claude", "execution", "low", ModelOverride{Effort: "xhigh"}, Model{"syn-claude-exec", "xhigh"}},
		{"codex effort only", "codex", "inherit", "", ModelOverride{Effort: "high"}, Model{"", "high"}},
		{"pi effort", "pi", "execution", "", ModelOverride{Effort: "xhigh"}, Model{"syn-pi/exec", "xhigh"}},
		{"grok model", "grok", "execution", "", ModelOverride{Model: "grok-4"}, Model{"grok-4", ""}},
		{"cursor model", "cursor", "execution", "", ModelOverride{Model: "composer-2"}, Model{"composer-2", ""}},
		{"opencode effort replaces the variant", "opencode", "execution", "", ModelOverride{Effort: "high"}, Model{"syn-oc/exec#high", ""}},
		{"opencode model keeps the release variant", "opencode", "execution", "", ModelOverride{Model: "x/y"}, Model{"x/y#medium", ""}},
		{"opencode model and effort", "opencode", "execution", "", ModelOverride{"x/y", "low"}, Model{"x/y#low", ""}},
		{"opencode inherit variant is stripped, not stacked", "opencode", "inherit", "", ModelOverride{Effort: "high"}, Model{"syn-oc/inherit#high", ""}},
		{"opencode inherit model only carries the release variant", "opencode", "inherit", "", ModelOverride{Model: "x/y"}, Model{"x/y#max", ""}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			o := c.o
			_, m, err := Resolve(resolveSource, resolveRole(c.profile, c.role), profiles, c.host, &o)
			if err != nil || m != c.want {
				t.Fatalf("Resolve(%s) = %+v, %v; want %+v", c.host, m, err, c.want)
			}
		})
	}
}

func TestResolveOverrideNilAndEmptyAreDifferent(t *testing.T) {
	profiles := syntheticProfiles(t)
	_, base, err := Resolve(resolveSource, resolveRole("reasoning", ""), profiles, "claude", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := Resolve(resolveSource, resolveRole("reasoning", ""), profiles, "claude", &ModelOverride{}); err == nil {
		t.Fatalf("an empty override was accepted; nil gives %+v", base)
	}
}

func TestResolveOverrideRejectsInvalidValues(t *testing.T) {
	profiles := syntheticProfiles(t)
	for _, c := range []struct {
		host string
		o    ModelOverride
	}{
		{"grok", ModelOverride{Effort: "high"}},
		{"claude", ModelOverride{Model: "a b"}},
		{"claude", ModelOverride{Effort: "bogus"}},
	} {
		if _, _, err := Resolve(resolveSource, resolveRole("execution", ""), profiles, c.host, &c.o); err == nil {
			t.Errorf("Resolve(%s, %+v) accepted an invalid override", c.host, c.o)
		}
	}
}

func TestResolveOverrideOpenCodeEffortNeedsAModel(t *testing.T) {
	noModel := editProfiles(t, func(host string, models map[string]any) {
		if host == "opencode" {
			models["execution"] = map[string]any{}
		}
	})
	data := resolveRole("execution", "")
	_, _, err := Resolve(resolveSource, data, noModel, "opencode", &ModelOverride{Effort: "high"})
	if err == nil || !strings.Contains(err.Error(), "OpenCode effort requires a model") {
		t.Fatalf("effort without any model = %v, want the OpenCode model error", err)
	}
	_, m, err := Resolve(resolveSource, data, noModel, "opencode", &ModelOverride{Model: "x/y", Effort: "high"})
	if err != nil || m != (Model{"x/y#high", ""}) {
		t.Fatalf("model and effort on a profile without a model = %+v, %v", m, err)
	}
	_, m, err = Resolve(resolveSource, data, noModel, "opencode", &ModelOverride{Model: "x/y"})
	if err != nil || m != (Model{"x/y", ""}) {
		t.Fatalf("model only on a profile without a model = %+v, %v", m, err)
	}
}

func TestRenderOverrideWritesEachHostsFields(t *testing.T) {
	profiles := syntheticProfiles(t)
	data := resolveRole("execution", "")
	for _, c := range []struct {
		host string
		o    ModelOverride
		want []string
	}{
		{"claude", ModelOverride{"opus", "max"}, []string{`model: "opus"`, `effort: "max"`}},
		{"codex", ModelOverride{"gpt-x", "high"}, []string{`model = "gpt-x"`, `model_reasoning_effort = "high"`}},
		{"pi", ModelOverride{"p/m", "xhigh"}, []string{`model: "p/m"`, `thinking: "xhigh"`}},
		{"opencode", ModelOverride{"o/m", "low"}, []string{`model: "o/m#low"`}},
		{"opencode", ModelOverride{Effort: "high"}, []string{`model: "syn-oc/exec#high"`}},
		{"grok", ModelOverride{Model: "g"}, []string{`model: "g"`}},
		{"cursor", ModelOverride{Model: "c"}, []string{`model: "c"`}},
	} {
		o := c.o
		out, err := Render(resolveSource, data, profiles, c.host, "/skills", &o)
		if err != nil {
			t.Fatalf("Render(%s): %v", c.host, err)
		}
		for _, want := range c.want {
			if !strings.Contains(string(out), want+"\n") {
				t.Errorf("Render(%s, %+v) has no line %q:\n%s", c.host, c.o, want, out)
			}
		}
		if c.host == "opencode" && strings.Count(string(out), "#") != 1 {
			t.Errorf("Render(opencode) stacked variants:\n%s", out)
		}
	}
}

func TestRenderOverrideChangesOnlyTheOverriddenLines(t *testing.T) {
	profiles := syntheticProfiles(t)
	data := resolveRole("execution", "")
	base, err := Render(resolveSource, data, profiles, "claude", "/skills", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := Render(resolveSource, data, profiles, "claude", "/skills", &ModelOverride{"opus", "max"})
	if err != nil {
		t.Fatal(err)
	}
	want := strings.NewReplacer(`model: "syn-claude-exec"`, `model: "opus"`, `effort: "medium"`, `effort: "max"`).Replace(string(base))
	if string(out) != want {
		t.Fatalf("override changed more than model and effort:\n%s\nwant\n%s", out, want)
	}
}
