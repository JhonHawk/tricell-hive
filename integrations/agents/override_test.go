package agents

import (
	"errors"
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
		{"effort on antigravity", "antigravity", ModelOverride{Model: "pro", Effort: "high"}, "antigravity does not support effort"},
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
		{"antigravity tier", "antigravity", "execution", "", ModelOverride{Model: "flash_lite"}, Model{"flash_lite", ""}},
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

func TestResolveOverrideRejectsInvalidValues(t *testing.T) {
	profiles := syntheticProfiles(t)
	for _, c := range []struct {
		host string
		o    ModelOverride
	}{
		{"grok", ModelOverride{Effort: "high"}},
		{"claude", ModelOverride{Model: "a b"}},
		{"claude", ModelOverride{Effort: "bogus"}},
		{"claude", ModelOverride{}},
	} {
		_, _, err := Resolve(resolveSource, resolveRole("execution", ""), profiles, c.host, &c.o)
		var oe *OverrideError
		if !errors.As(err, &oe) {
			t.Errorf("Resolve(%s, %+v) = %v, want an *OverrideError", c.host, c.o, err)
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

	// Cursor renders a model override in place of its inherited model.
	base, err = Render(resolveSource, data, profiles, "cursor", "/skills", nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err = Render(resolveSource, data, profiles, "cursor", "/skills", &ModelOverride{Model: "c"})
	if err != nil {
		t.Fatal(err)
	}
	want = strings.Replace(string(base), `model: "inherit"`, `model: "c"`, 1)
	if want == string(base) || string(out) != want {
		t.Fatalf("cursor override:\n%s\nwant\n%s", out, want)
	}
}
