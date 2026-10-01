package main

import (
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the #91 follow-ups to the Models view (H1).

func TestModelsViewCodexEffortResetSaysSoInThePanel(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	toHost(t, e.d, "codex")
	store(t, e.home, e.stateDir, "codex", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pick(e.d, "gpt-5.5")
	e.d.mustShow("Effort max dropped (not offered by gpt-5.5); using release default")
	// A choice that keeps the effort says nothing.
	e.d.key("right")
	pick(e.d, "gpt-6.1")
	e.d.mustNotShow("Effort max dropped")
}

func TestModelsViewXOnARoleWithoutOverrideSaysThereIsNothingToReset(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Model: "opus"}})
	e.d.key("r")
	selectRole(t, e.d, "hive-design-architecture")
	before := snap(t, e)
	e.d.key("x")
	e.d.mustShow("Nothing to reset: hive-design-architecture has no override")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("x changed files")
	}
	e.d.key("down")
	e.d.mustNotShow("Nothing to reset")
	for _, l := range e.d.lines() {
		if len([]rune(l)) > 80 {
			t.Fatalf("line wider than 80: %q", l)
		}
	}
}

func TestModelsViewXOnAGroupWithoutOverrideSaysThereIsNothingToReset(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	e.d.key("r")
	for range 60 {
		if it, ok := e.v.selectedItem(); ok && it.header {
			break
		}
		e.d.key("down")
	}
	if it, ok := e.v.selectedItem(); !ok || !it.header {
		t.Fatal("no group header reached")
	}
	e.d.key("x")
	e.d.mustShow("Nothing to reset: ")
	e.d.mustShow("group has no override")
}

func TestModelsViewPickerShowsThePositionAmongTheModels(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		e := pickerEnv(t, "claude,codex", standardFake(), size[0], size[1])
		toHost(t, e.d, "codex")
		selectRole(t, e.d, "plain-role")
		e.d.key("enter", "right")
		in := boxInterior(t, e.d)
		last := in[len(in)-1]
		if !strings.Contains(last, "1 of 3") {
			t.Fatalf("bottom row = %q, want 1 of 3", last)
		}
		e.d.key("down")
		if last := boxInterior(t, e.d); !strings.Contains(last[len(last)-1], "2 of 3") {
			t.Fatalf("bottom row = %q, want 2 of 3", last[len(last)-1])
		}
		e.d.key("end")
		if last := boxInterior(t, e.d); strings.Contains(last[len(last)-1], " of ") {
			t.Fatalf("on a bottom entry no position is shown: %q", last[len(last)-1])
		}
		for _, l := range e.d.lines() {
			if len([]rune(l)) > size[0]-2 && strings.HasPrefix(l, "│") {
				t.Fatalf("box line wider than %d: %q", size[0]-2, l)
			}
		}
	}
}

func TestModelsViewEffortResetNoticeKeepsTheEffortWhenTheIDIsLong(t *testing.T) {
	long := "gpt-5.5-with-an-extremely-long-model-identifier-that-cannot-fit-in-one-row-at-all"
	f := standardFake()
	f.out["codex"] = codexList(map[string][]string{"gpt-6.1-sol": allEfforts, long: {"low", "medium"}}, "gpt-6.1-sol", long)
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	store(t, e.home, e.stateDir, "codex", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pick(e.d, "extremely-long")
	e.d.mustShow("Effort max dropped")
	assertFits(t, e.d, 80, 24)
}
