package management

import (
	"strings"
	"testing"
)

// tamper edits a plan the way a hand-edited --out file would and recomputes its
// ID, so the only thing that can reject it is validation of its content.
func tamper(p Plan, edit func(*Plan)) Plan {
	copied := map[string]map[string]ModelOverride{}
	for host, roles := range p.ModelOverrides {
		copied[host] = map[string]ModelOverride{}
		for role, v := range roles {
			copied[host][role] = v
		}
	}
	p.ModelOverrides = copied
	edit(&p)
	p.ID = planID(p)
	return p
}

func setOverride(p *Plan, host, role string, v ModelOverride) {
	if p.ModelOverrides == nil {
		p.ModelOverrides = map[string]map[string]ModelOverride{}
	}
	if p.ModelOverrides[host] == nil {
		p.ModelOverrides[host] = map[string]ModelOverride{}
	}
	p.ModelOverrides[host][role] = v
}

func TestModelOverrideApplyRejectsHandEditedPlansWithTheSpecificReason(t *testing.T) {
	o, _ := installedFiveHosts(t)
	base := buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "high"}})
	removal := o
	removal.Hosts = []string{"codex"}
	rm := plan(t, "remove", removal)
	cases := []struct {
		name string
		p    Plan
		want string
	}{
		{"an invalid override", tamper(base, func(p *Plan) { setOverride(p, "claude", plainRole, ModelOverride{Model: "a b"}) }), "model may only contain"},
		{"an empty override", tamper(base, func(p *Plan) { setOverride(p, "claude", plainRole, ModelOverride{}) }), "needs a model or an effort"},
		{"a different override for a CLI outside the plan", tamper(base, func(p *Plan) { setOverride(p, "codex", plainRole, ModelOverride{Model: "x"}) }), "model overrides for codex differ from the state"},
		{"an unknown CLI", tamper(base, func(p *Plan) { setOverride(p, "vim", plainRole, ModelOverride{Model: "x"}) }), "unsupported host"},
		{"a role the release does not have", tamper(base, func(p *Plan) { setOverride(p, "claude", "ghost-role", ModelOverride{Model: "x"}) }), `role "ghost-role" is not in the release`},
		{"an empty role map", tamper(base, func(p *Plan) { p.ModelOverrides["codex"] = map[string]ModelOverride{} }), "empty model override set"},
		{"overrides on a remove plan", tamper(rm, func(p *Plan) { setOverride(p, "claude", plainRole, ModelOverride{Model: "x"}) }), "must not carry model overrides"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stateBefore := stateBytes(t, o)
			_, err := (Engine{}).Apply(c.p)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("Apply error = %v, want one containing %q", err, c.want)
			}
			if strings.Contains(err.Error(), "invalid or legacy plan") {
				t.Fatalf("Apply failed with the generic plan error: %v", err)
			}
			if stateBytes(t, o) != stateBefore {
				t.Fatal("a rejected plan changed the state")
			}
		})
	}
	// An untouched plan still applies.
	apply(t, base)
}

func TestModelOverrideVoicePlanWithOverridesIsRejected(t *testing.T) {
	o, _ := installedFiveHosts(t)
	voiceSource(t, o)
	vp := tamper(voicePlan(t, "set", o, jarvisSirSubtle), func(p *Plan) { setOverride(p, "claude", plainRole, ModelOverride{Model: "x"}) })
	if _, err := (Engine{}).Apply(vp); err == nil || !strings.Contains(err.Error(), "must not carry model overrides") {
		t.Fatalf("Apply error = %v", err)
	}
}

func TestModelOverrideStalePlanIsRejectedByStateHash(t *testing.T) {
	o, _ := installedFiveHosts(t)
	// Both plans start from the same clean state. The first one stores an
	// override equal to the release value, which changes the state and no file,
	// so the second plan's records still match and only the state hash is stale.
	stale := buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "high"}})
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: claudeEffort}}))
	_, err := (Engine{}).Apply(stale)
	if err == nil || !strings.Contains(err.Error(), "stale plan: state changed") {
		t.Fatalf("Apply error = %v, want the state hash rejection", err)
	}
	if got := stateFor(t, o).ModelOverrides["claude"][plainRole]; got != (ModelOverride{Effort: claudeEffort}) {
		t.Fatalf("the stale plan changed the stored override to %+v", got)
	}
}

func TestModelOverrideApplyRejectsAnEffortOnCLIsWithoutEffort(t *testing.T) {
	o := setup(t)
	o.Hosts = allAgentHosts
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	for _, host := range []string{"grok", "cursor"} {
		t.Run(host, func(t *testing.T) {
			// A valid plan for the CLI, edited by hand to carry an effort.
			base := buildModels(t, o, host, map[string]ModelOverride{plainRole: {Model: "some-model"}})
			edited := tamper(base, func(p *Plan) { setOverride(p, host, plainRole, ModelOverride{Model: "some-model", Effort: "high"}) })
			stateBefore := stateBytes(t, o)
			_, err := (Engine{}).Apply(edited)
			want := host + " does not support effort"
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("Apply error = %v, want one containing %q", err, want)
			}
			if stateBytes(t, o) != stateBefore {
				t.Fatal("a rejected plan changed the state")
			}
		})
	}
}
