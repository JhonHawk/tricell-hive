package management

// T2 fix round: tests adapted from the independent verifier's probes
// (P1-P12) plus the additional fixes (H1-H6, D2) requested alongside them.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// H1 / P1: the voice block must sit immediately after the Hive block's END
// line, not appended at EOF, so text the user wrote after the Hive block
// survives untouched and "off" restores the file byte for byte.
func TestVoicePlacementImmediatelyAfterHiveBlock(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	withUser := get(t, codexPath(o)) + "user notes after hive\n"
	put(t, codexPath(o), withUser)

	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	got := get(t, codexPath(o))
	endIdx := strings.Index(got, End)
	vIdx := strings.Index(got, VoiceBegin)
	if endIdx < 0 || vIdx < 0 {
		t.Fatalf("missing markers in %q", got)
	}
	between := got[endIdx+len(End) : vIdx]
	if strings.TrimSpace(between) != "" {
		t.Fatalf("voice block is not right after the Hive block; between = %q", between)
	}
	if strings.Count(got, VoiceBegin) != 1 {
		t.Fatalf("expected exactly one voice block, got %d", strings.Count(got, VoiceBegin))
	}
	if !strings.HasSuffix(got, "user notes after hive\n") {
		t.Fatalf("user text after the Hive block was not preserved: %q", got)
	}

	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	if get(t, codexPath(o)) != withUser {
		t.Fatalf("off not byte-identical:\ngot:  %q\nwant: %q", get(t, codexPath(o)), withUser)
	}
}

// H1 / P2: user text after the Hive block with no trailing newline must not
// produce a malformed voice block, and "off" must still be buildable
// afterwards.
func TestVoicePlacementUserTextWithoutTrailingNewline(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	withUser := get(t, codexPath(o)) + "user notes no newline"
	put(t, codexPath(o), withUser)

	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	got := get(t, codexPath(o))
	if _, _, err := blockRange([]byte(got), voiceMarkers); err != nil {
		t.Fatalf("written file has a malformed voice block: %v", err)
	}
	if !strings.HasSuffix(got, "user notes no newline") {
		t.Fatalf("trailing user text without a newline was not preserved: %q", got)
	}
	if _, err := BuildVoicePlan("off", o, VoiceSetting{}); err != nil {
		t.Fatalf("voice off must still work afterwards: %v", err)
	}
	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	if get(t, codexPath(o)) != withUser {
		t.Fatalf("off not byte-identical:\ngot:  %q\nwant: %q", get(t, codexPath(o)), withUser)
	}
}

// H1 / P9: CRLF files keep one consistent line ending, and off round-trips
// byte for byte.
func TestVoicePlacementCRLF(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	put(t, codexPath(o), "user\r\n")
	apply(t, plan(t, "install", o))
	before := get(t, codexPath(o))

	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	got := get(t, codexPath(o))
	if strings.Contains(got, "\n") && strings.Count(got, "\n") != strings.Count(got, "\r\n") {
		t.Fatalf("mixed line endings after voice set: %q", got)
	}
	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	if get(t, codexPath(o)) != before {
		t.Fatalf("CRLF off not byte-identical:\ngot:  %q\nwant: %q", get(t, codexPath(o)), before)
	}
}

// H2 / P5: installing only Codex must not touch Claude's file's voice span.
func TestVoiceInstallRestrictedToSelectedHosts(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	claudeSpanBefore := stateFor(t, o).VoiceSpans[claudePath(o)]

	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "Changed tone.\n")
	c := o
	c.Hosts = []string{"codex"}
	p := plan(t, "install", c)
	for _, vc := range p.Voice {
		if vc.Path == claudePath(o) {
			t.Fatalf("install --hosts codex must not produce a voice change for Claude's file")
		}
	}
	apply(t, p)
	if !reflect.DeepEqual(stateFor(t, o).VoiceSpans[claudePath(o)], claudeSpanBefore) {
		t.Fatal("install --hosts codex changed Claude's voice span")
	}
}

// H2 / P5b: a project-scope install must never touch the user-scope voice.
func TestVoiceProjectScopeInstallDoesNotTouchUserVoice(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	spansBefore := stateFor(t, o).VoiceSpans

	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "Changed tone.\n")
	pr := o
	pr.Scope = "project"
	pr.Root = filepath.Join(filepath.Dir(o.Home), "proj")
	if err := os.MkdirAll(pr.Root, 0700); err != nil {
		t.Fatal(err)
	}
	p, err := BuildPlan("install", pr)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Voice) != 0 {
		t.Fatalf("a project-scope install must add no voice changes, got %d", len(p.Voice))
	}
	apply(t, p)
	if !reflect.DeepEqual(stateFor(t, o).VoiceSpans, spansBefore) {
		t.Fatal("project-scope install touched user-scope voice spans")
	}
}

// D2 / P4: a host gaining its Hive block for the first time in an install
// where a voice is already active must get the voice span in that same
// plan, including surviving a recovery at each write point.
func TestVoiceNewHostGetsVoiceInSameInstall(t *testing.T) {
	// write:1 is the shared skill entry, already installed and unchanged by
	// Claude joining, so it is skipped by commitTransaction and never fires;
	// write:0 is the merged Hive+voice block entry and write:2 is the fresh
	// skill-directory symlink alias.
	for _, stage := range []string{"", "prepared", "write:0", "write:2", "state"} {
		t.Run("stage="+stage, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"codex"}
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			apply(t, voicePlan(t, "set", o, jarvisSirSubtle))

			o2 := o
			o2.Hosts = []string{"claude"}
			p := plan(t, "install", o2)
			foundClaudeVoice := false
			for _, vc := range p.Voice {
				if vc.Path == claudePath(o) {
					foundClaudeVoice = true
					if vc.Before != nil {
						t.Fatalf("expected Before nil for a brand-new voice span, got %+v", vc.Before)
					}
				}
			}
			if !foundClaudeVoice {
				t.Fatal("expected a voice change for Claude's brand-new file in the same install")
			}
			for _, ch := range p.Changes {
				if ch.Target.Path == claudePath(o) {
					for _, vc := range p.Voice {
						if vc.Path == claudePath(o) && vc.Expected != ch.Expected {
							t.Fatalf("voice Expected must equal the Hive Change's Expected for a new host")
						}
					}
				}
			}

			if stage == "" {
				apply(t, p)
				if !strings.Contains(get(t, claudePath(o)), VoiceBegin) {
					t.Fatal("Claude's file does not have the voice block after a normal apply")
				}
				return
			}
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("expected the injected failure to stop the apply")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, claudePath(o))
			apply(t, plan(t, "install", o2))
			if !strings.Contains(get(t, claudePath(o)), VoiceBegin) {
				t.Fatal("install after recovery did not write Claude's voice block")
			}
		})
	}
}

// H3 / P6: an unregistered voice block (markers present, no registered
// span) must block install and remove on that file, preserving it, even
// when no voice is currently active at all.
func TestVoiceUnregisteredMarkersBlockInstallAndRemoveWithNoActiveVoice(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	injected := get(t, codexPath(o)) + VoiceBegin + "\nunregistered\n" + VoiceEnd + "\n"
	put(t, codexPath(o), injected)
	put(t, filepath.Join(o.Source, GlobalSource), "# Changed\n")

	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("install accepted an unregistered voice block")
	}
	if _, err := BuildPlan("remove", o); err == nil {
		t.Fatal("remove accepted an unregistered voice block")
	}
	if get(t, codexPath(o)) != injected {
		t.Fatal("a rejected conflict must not change the file")
	}
}

// H4: narrowing a shared file's consumers must narrow the voice span's
// registered Consumers too, so status for the retired host stops listing it.
func TestVoiceSpanConsumersNarrowOnPartialRemoval(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "grok"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if len(stateFor(t, o).VoiceSpans[claudePath(o)].Consumers) != 2 {
		t.Fatalf("expected 2 consumers on the shared span, got %+v", stateFor(t, o).VoiceSpans[claudePath(o)].Consumers)
	}

	claudeOnly := o
	claudeOnly.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", claudeOnly))
	remaining := stateFor(t, o).VoiceSpans[claudePath(o)].Consumers
	if len(remaining) != 1 || remaining[0].Host != "grok" {
		t.Fatalf("expected only grok left as a consumer, got %+v", remaining)
	}

	entries, err := Status(claudeOnly)
	if err != nil {
		t.Fatal(err)
	}
	if voiceStatusFor(entries, claudePath(o)) != nil {
		t.Fatal("status --hosts claude still lists the voice row after claude was removed")
	}
}

// H5: the Kind tie-break must be effective for two rows at the same path
// (the voice row sorts after its block row).
func TestVoiceStatusSortKindTieBreak(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	blockIdx, voiceIdx := -1, -1
	for i, e := range entries {
		if e.Path == codexPath(o) && e.Kind == "block" {
			blockIdx = i
		}
		if e.Path == codexPath(o) && e.Kind == "voice" {
			voiceIdx = i
		}
	}
	if blockIdx < 0 || voiceIdx < 0 {
		t.Fatalf("missing block or voice row for %s", codexPath(o))
	}
	if voiceIdx != blockIdx+1 {
		t.Fatalf("expected the voice row immediately after its block row; block at %d, voice at %d", blockIdx, voiceIdx)
	}
}

// H6: validatePlan must reject a plan with two voice changes for the same path.
func TestValidatePlanRejectsDuplicateVoicePath(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	vp.Voice = append(vp.Voice, vp.Voice[0])
	vp.ID = planID(vp)
	if _, err := (Engine{}).Apply(vp); err == nil {
		t.Fatal("expected a duplicate voice path to be rejected")
	}
}

// P7: a recovery attempt that is itself interrupted (recover:N failpoint)
// must be safely resumable by a later recovery call.
func TestVoiceInterruptedRecoveryResumes(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	codexBefore, claudeBefore := get(t, codexPath(o)), get(t, claudePath(o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	e := Engine{failpoint: func(s string) error {
		if s == "write:1" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(vp); err == nil {
		t.Fatal("expected the injected failure to stop the apply")
	}
	r := Engine{failpoint: func(s string) error {
		if s == "recover:1" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := r.Recover(o.StateDir); err == nil {
		t.Fatal("expected the injected recovery failure to stop it")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if get(t, codexPath(o)) != codexBefore || get(t, claudePath(o)) != claudeBefore {
		t.Fatal("not fully restored after a resumed recovery")
	}
}

// H7 / P13: Grok joins Claude's shared file, already voiced. The install
// plan must widen the span's Consumers even though the rendered bytes don't
// change, so status --hosts grok shows the voice row right away.
func TestVoiceSpanConsumersWidenWhenHostJoinsSharedFile(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if len(stateFor(t, o).VoiceSpans[claudePath(o)].Consumers) != 1 {
		t.Fatalf("test setup: expected one consumer before Grok joins")
	}

	g := o
	g.Hosts = []string{"grok"}
	p := plan(t, "install", g)
	found := false
	for _, vc := range p.Voice {
		if vc.Path == claudePath(o) {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a voice change widening the shared span's consumers when Grok joins")
	}
	apply(t, p)
	consumers := stateFor(t, o).VoiceSpans[claudePath(o)].Consumers
	if len(consumers) != 2 {
		t.Fatalf("expected 2 consumers after Grok joins, got %+v", consumers)
	}
	entries, err := Status(g)
	if err != nil {
		t.Fatal(err)
	}
	if voiceStatusFor(entries, claudePath(o)) == nil {
		t.Fatal("status --hosts grok does not show the voice row right after Grok joins")
	}
}

// F1: BuildVoicePlan must normalize an empty Address/Intensity to their
// defaults before storing the choice, so State.Voice, the status Voice
// field and repeat-detection all see "none"/"subtle" rather than "".
func TestVoiceSetNormalizesDefaultsAndRecognizesUnchanged(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, VoiceSetting{ID: "jarvis"}))

	got := stateFor(t, o).Voice
	if got == nil || got.Address != "none" || got.Intensity != "subtle" {
		t.Fatalf("expected normalized defaults stored, got %+v", got)
	}
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	row := voiceStatusFor(entries, codexPath(o))
	if row == nil || row.Voice != "jarvis (none, subtle)" {
		t.Fatalf("expected status Voice \"jarvis (none, subtle)\", got %+v", row)
	}

	p2, err := BuildVoicePlan("set", o, VoiceSetting{ID: "jarvis", Address: "none", Intensity: "subtle"})
	if err != nil {
		t.Fatal(err)
	}
	if unchanged, err := PlanUnchanged(p2); err != nil || !unchanged {
		t.Fatalf("expected the explicit default choice to be recognized as unchanged, got unchanged=%v err=%v", unchanged, err)
	}
}

// validatePlan must reject a "voice" plan with zero voice changes (design.md:
// "at least one voice change"). BuildVoicePlan must never itself produce
// such a plan for a real "set"/"off" invocation, even when every path is
// already exactly at the requested state: PlanUnchanged, not an empty
// p.Voice, is how a caller learns there is nothing to apply.
func TestValidatePlanRejectsVoiceActionWithZeroVoiceChanges(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))

	repeat, err := BuildVoicePlan("set", o, jarvisSirSubtle)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeat.Voice) == 0 {
		t.Fatal("expected BuildVoicePlan to include at least one voice change even when nothing would change")
	}
	if unchanged, err := PlanUnchanged(repeat); err != nil || !unchanged {
		t.Fatalf("expected the repeat plan to be reported unchanged, got unchanged=%v err=%v", unchanged, err)
	}

	repeat.Voice = nil
	repeat.ID = planID(repeat)
	if err := validatePlan(repeat, stateFor(t, o)); err == nil {
		t.Fatal("expected validatePlan to reject a voice-action plan with zero voice changes")
	}
}

// --- test gaps: mutants that survived ------------------------------------------

// P10: a voice plan built before an outside edit must be rejected as stale.
func TestProbeStaleVoicePlanExpectedFingerprint(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	edited := get(t, codexPath(o)) + "late edit\n"
	put(t, codexPath(o), edited)
	if _, err := (Engine{}).Apply(vp); err == nil {
		t.Fatal("stale voice plan applied")
	}
	if get(t, codexPath(o)) != edited {
		t.Fatal("a rejected stale plan must not change the file")
	}
}

// P10b: a voice plan whose Before no longer matches State.VoiceSpans (a
// newer voice was already applied) must be rejected.
func TestProbeStaleVoiceOwnershipAfterNewerSet(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp1 := voicePlan(t, "set", o, jarvisSirSubtle)
	apply(t, voicePlan(t, "set", o, VoiceSetting{ID: "jarvis", Address: "none", Intensity: "marked"}))
	if _, err := (Engine{}).Apply(vp1); err == nil {
		t.Fatal("stale voice plan applied over a newer voice")
	}
}

// P11: a tampered plan whose After smuggles a Hive marker must be rejected
// by validateVoicePayload.
func TestProbeTamperedVoicePayloadRejected(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	vp.Voice[0].After.Managed = []byte(VoiceBegin + "\n" + Begin + "\n" + VoiceEnd + "\n")
	vp.ID = planID(vp)
	if _, err := (Engine{}).Apply(vp); err == nil {
		t.Fatal("tampered voice payload applied")
	}
}

// P12: an install whose only change is the voice text must not be "unchanged".
func TestProbeUnchangedInstallVoiceOnlyIsNotUnchanged(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "Changed tone.\n")
	p := plan(t, "install", o)
	if u, err := PlanUnchanged(p); err != nil || u {
		t.Fatalf("voice-only install reported unchanged=%v err=%v", u, err)
	}
}

// TestVoiceOffRecoveryKeepsVoiceAfterHiveBlock: recovering an interrupted
// voice off must put the voice block back right after the Hive block, not at
// the end of the file, also when user text follows the Hive block.
func TestVoiceOffRecoveryKeepsVoiceAfterHiveBlock(t *testing.T) {
	for _, tail := range []string{"user tail\n", "user tail no newline"} {
		t.Run(strings.TrimSpace(tail), func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"codex"}
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			put(t, codexPath(o), get(t, codexPath(o))+tail)
			apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
			before := get(t, codexPath(o))
			e := Engine{failpoint: func(s string) error {
				if s == "write:0" {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(voicePlan(t, "off", o, VoiceSetting{})); err == nil {
				t.Fatal("expected failure")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, codexPath(o)) != before {
				t.Errorf("not byte-identical after recovery:\ngot  %q\nwant %q", get(t, codexPath(o)), before)
			}
		})
	}
}
