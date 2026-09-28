package management

// TDD tests for T2: the voice layer's plan, apply, status and recovery
// integration (BuildVoicePlan, addVoiceChanges, the voice-aware entry
// mechanics in prepareEntries/prepareTransaction/Recover, and Status's
// voice rows). All homes are synthetic (setup(t)); none touch a real user
// home or the repository's real content/voices/.

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// voiceSource writes a synthetic content/voices/ tree (preamble + one
// voice) under o.Source, distinct from the repository's real content/voices/.
func voiceSource(t *testing.T, o Options) {
	t.Helper()
	put(t, filepath.Join(o.Source, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "Warm, formal, a little dry.\n")
}

func codexPath(o Options) string  { return filepath.Join(o.Home, ".codex", "AGENTS.md") }
func claudePath(o Options) string { return filepath.Join(o.Home, ".claude", "CLAUDE.md") }

func voicePlan(t *testing.T, action string, o Options, v VoiceSetting) Plan {
	t.Helper()
	p, err := BuildVoicePlan(action, o, v)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func voiceStatusFor(entries []StatusEntry, path string) *StatusEntry {
	for i := range entries {
		if entries[i].Path == path && entries[i].Kind == "voice" {
			return &entries[i]
		}
	}
	return nil
}

var jarvisSirSubtle = VoiceSetting{ID: "jarvis", Address: "sir", Intensity: "subtle"}

// --- set/off byte-identical (AC2, AC3) ------------------------------------------

func TestVoiceSetThenOffIsByteIdentical(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	codexBefore, claudeBefore := get(t, codexPath(o)), get(t, claudePath(o))

	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	if len(vp.Voice) != 2 {
		t.Fatalf("expected a voice change for both Codex and Claude, got %d", len(vp.Voice))
	}
	apply(t, vp)
	if get(t, codexPath(o)) == codexBefore || !strings.Contains(get(t, codexPath(o)), VoiceBegin) {
		t.Fatal("voice block not written to Codex's file")
	}
	if !strings.Contains(get(t, claudePath(o)), VoiceBegin) {
		t.Fatal("voice block not written to Claude's file")
	}
	if s := stateFor(t, o); s.Voice == nil || s.Voice.ID != "jarvis" || len(s.VoiceSpans) != 2 {
		t.Fatalf("voice state not recorded: %+v", s.Voice)
	}

	apply(t, voicePlan(t, "off", o, VoiceSetting{}))
	if get(t, codexPath(o)) != codexBefore {
		t.Fatal("Codex bytes not restored byte for byte")
	}
	if get(t, claudePath(o)) != claudeBefore {
		t.Fatal("Claude bytes not restored byte for byte")
	}
	if s := stateFor(t, o); s.Voice != nil || len(s.VoiceSpans) != 0 {
		t.Fatalf("voice state not cleared: %+v", s)
	}
}

func TestVoiceSetWithoutHiveInstalledFails(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	if _, err := BuildVoicePlan("set", o, jarvisSirSubtle); err == nil || !strings.Contains(err.Error(), "Hive not installed") {
		t.Fatalf("expected a Hive-not-installed error, got %v", err)
	}
}

func TestVoiceSetUnknownVoiceFailsBeforeWriting(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	before := get(t, codexPath(o))
	if _, err := BuildVoicePlan("set", o, VoiceSetting{ID: "nonexistent"}); err == nil {
		t.Fatal("expected an error for an unknown voice")
	}
	if get(t, codexPath(o)) != before {
		t.Fatal("a failed voice plan must not have changed anything")
	}
}

// --- install regenerates on changed text, keeps unchanged (AC4) ----------------

func TestVoiceInstallRegeneratesOnlyWhenTextChanges(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	original := stateFor(t, o).VoiceSpans[codexPath(o)]

	// Unrelated Hive text change: the voice span must not be touched.
	put(t, filepath.Join(o.Source, GlobalSource), "# New rules\nSomething else.\n")
	p := plan(t, "install", o)
	if len(p.Voice) != 0 {
		t.Fatalf("expected no voice changes when the voice text did not change, got %d", len(p.Voice))
	}
	apply(t, p)
	if !bytes.Equal(stateFor(t, o).VoiceSpans[codexPath(o)].Managed, original.Managed) {
		t.Fatal("unrelated Hive update touched the voice span")
	}

	// Voice text changes: the span must regenerate, keeping the same choice.
	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "A completely different tone.\n")
	p2 := plan(t, "install", o)
	if len(p2.Voice) != 2 {
		t.Fatalf("expected a voice change for both files, got %d", len(p2.Voice))
	}
	apply(t, p2)
	regenerated := stateFor(t, o).VoiceSpans[codexPath(o)]
	if bytes.Equal(regenerated.Managed, original.Managed) {
		t.Fatal("voice span was not regenerated after its source text changed")
	}
	if !strings.Contains(string(regenerated.Managed), "A completely different tone.") {
		t.Fatal("regenerated span does not contain the new voice text")
	}
	if !strings.Contains(strings.ToLower(string(regenerated.Managed)), "formal") {
		t.Fatal("regenerated span lost the original address choice (sir)")
	}
}

func TestVoiceInstallWithReleaseIDLeavesSpansUntouched(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	p := plan(t, "install", o)
	apply(t, p)
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	original := stateFor(t, o).VoiceSpans[codexPath(o)]

	put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "Changed, but --release must ignore this.\n")
	old := o
	old.ReleaseID = p.Release.ID
	rp := plan(t, "install", old)
	if len(rp.Voice) != 0 {
		t.Fatalf("expected --release to add no voice changes, got %d", len(rp.Voice))
	}
	apply(t, rp)
	if !bytes.Equal(stateFor(t, o).VoiceSpans[codexPath(o)].Managed, original.Managed) {
		t.Fatal("--release install touched the voice span")
	}
}

// --- remove: single consumer and a Grok+Claude shared file (AC5) ---------------

func TestVoiceRemoveSingleConsumerAndSharedFile(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "claude", "grok"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if len(stateFor(t, o).VoiceSpans) != 2 { // Codex's file and the shared Claude/Grok file
		t.Fatalf("expected 2 voice spans, got %d", len(stateFor(t, o).VoiceSpans))
	}

	// Codex is the sole consumer of its file: removing it drops the file and its span.
	codexOnly := o
	codexOnly.Hosts = []string{"codex"}
	apply(t, plan(t, "remove", codexOnly))
	absent(t, codexPath(o))
	if _, ok := stateFor(t, o).VoiceSpans[codexPath(o)]; ok {
		t.Fatal("Codex's voice span survived removing its only consumer")
	}

	// Claude and Grok share one file: removing Claude alone must keep Grok's
	// Hive block and the voice span, since Grok still consumes that file.
	claudeOnly := o
	claudeOnly.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", claudeOnly))
	if !strings.Contains(get(t, claudePath(o)), Begin) {
		t.Fatal("Grok's shared Hive block was removed by Claude's own removal")
	}
	if !strings.Contains(get(t, claudePath(o)), VoiceBegin) {
		t.Fatal("the shared file's voice block was removed while Grok still consumes it")
	}
	if _, ok := stateFor(t, o).VoiceSpans[claudePath(o)]; !ok {
		t.Fatal("the shared file's voice span was dropped while Grok still consumes it")
	}

	// Removing Grok too retires the last consumer: both blocks and the span go.
	grokOnly := o
	grokOnly.Hosts = []string{"grok"}
	apply(t, plan(t, "remove", grokOnly))
	absent(t, claudePath(o))
	if len(stateFor(t, o).VoiceSpans) != 0 {
		t.Fatal("voice span survived its last consumer's removal")
	}
	if stateFor(t, o).Voice != nil {
		t.Fatal("State.Voice was not cleared once no span remained")
	}
}

// --- status: ok and drift (AC6) --------------------------------------------------

func TestVoiceStatusInstalledAndDrift(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))

	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	row := voiceStatusFor(entries, codexPath(o))
	if row == nil {
		t.Fatal("no voice status row for Codex's file")
	}
	if row.Status != "installed" {
		t.Fatalf("expected installed, got %q", row.Status)
	}
	if row.Voice != "jarvis (sir, subtle)" {
		t.Fatalf("unexpected Voice field: %q", row.Voice)
	}

	tampered := strings.Replace(get(t, codexPath(o)), "Warm, formal, a little dry.", "TAMPERED", 1)
	put(t, codexPath(o), tampered)
	entries2, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	row2 := voiceStatusFor(entries2, codexPath(o))
	if row2 == nil || row2.Status != "drift" {
		t.Fatalf("expected a drift voice row, got %+v", row2)
	}
}

// --- hand-edit conflicts: set, off, install, remove (AC7) -----------------------

func TestVoiceHandEditConflictsSetOffInstallRemove(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	tampered := strings.Replace(get(t, codexPath(o)), "Warm, formal, a little dry.", "TAMPERED", 1)
	put(t, codexPath(o), tampered)

	if _, err := BuildVoicePlan("set", o, jarvisSirSubtle); err == nil {
		t.Fatal("voice set accepted a hand-edited voice block")
	}
	if _, err := BuildVoicePlan("off", o, VoiceSetting{}); err == nil {
		t.Fatal("voice off accepted a hand-edited voice block")
	}
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("install accepted a hand-edited voice block")
	}
	codexOnly := o
	codexOnly.Hosts = []string{"codex"}
	if _, err := BuildPlan("remove", codexOnly); err == nil {
		t.Fatal("remove accepted a hand-edited voice block")
	}
	if get(t, codexPath(o)) != tampered {
		t.Fatal("a rejected conflict must not change the file")
	}
}

func TestVoiceUnregisteredMarkersConflict(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	// Inject a well-formed but unregistered voice block by hand.
	injected := get(t, codexPath(o)) + VoiceBegin + "\nunregistered\n" + VoiceEnd + "\n"
	put(t, codexPath(o), injected)

	if _, err := BuildVoicePlan("set", o, jarvisSirSubtle); err == nil {
		t.Fatal("voice set accepted an unregistered voice block")
	}
	if get(t, codexPath(o)) != injected {
		t.Fatal("a rejected conflict must not change the file")
	}
}

// --- state bookkeeping: receipts, Voice/VoiceSpans untouched --------------------

func TestVoicePlanKeepsInstallationReceipts(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.0\n")
	apply(t, plan(t, "install", o))
	before := stateFor(t, o).Installations
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	after := stateFor(t, o).Installations
	if len(before) == 0 || len(after) != len(before) {
		t.Fatalf("a voice plan must not touch installation receipts: before=%v after=%v", before, after)
	}
	for k, v := range before {
		if after[k].Product != v.Product {
			t.Fatalf("installation receipt for %s changed by a voice plan", k)
		}
	}
}

func TestInstallAndRemoveKeepVoiceWhenNotTouchingIt(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "claude"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	spansBefore := stateFor(t, o).VoiceSpans

	// An install that changes neither Hive text nor voice text must leave
	// Voice/VoiceSpans exactly as they are.
	apply(t, plan(t, "install", o))
	if got := stateFor(t, o); got.Voice == nil || len(got.VoiceSpans) != len(spansBefore) {
		t.Fatal("an unrelated reinstall touched Voice/VoiceSpans")
	}

	// Adding a brand-new host (Grok, sharing Claude's file) via remove/install
	// unrelated to Codex/Claude must not touch their existing spans.
	o2 := o
	o2.Hosts = []string{"grok"}
	apply(t, plan(t, "install", o2))
	if !bytes.Equal(stateFor(t, o).VoiceSpans[codexPath(o)].Managed, spansBefore[codexPath(o)].Managed) {
		t.Fatal("installing an unrelated host touched Codex's voice span")
	}
}

// --- removing both blocks from a Hive-created file deletes it ------------------

func TestRemovingHiveAndVoiceDeletesHiveCreatedFile(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if s := stateFor(t, o).Records[codexPath(o)]; !s.CreatedFile {
		t.Fatal("test setup expects Codex's file to be Hive-created")
	}

	apply(t, plan(t, "remove", o))
	absent(t, codexPath(o))
}

// --- round trip and no-voice compatibility --------------------------------------

func TestVoicePlanRoundTripsSaveLoadPlan(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	out := filepath.Join(o.Home, "voice-plan.json")
	if err := SavePlan(out, vp); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadPlan(out)
	if err != nil || loaded.ID != vp.ID {
		t.Fatalf("round trip failed: %v", err)
	}
	if len(loaded.Voice) != len(vp.Voice) {
		t.Fatal("voice changes lost across SavePlan/LoadPlan")
	}
	apply(t, loaded)
	if !strings.Contains(get(t, codexPath(o)), VoiceBegin) {
		t.Fatal("plan loaded from disk did not apply its voice change")
	}
}

func TestPlanUnchangedAccountsForVoiceChanges(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	if unchanged, err := PlanUnchanged(vp); err != nil || unchanged {
		t.Fatalf("expected a real voice plan to be reported as changed, got unchanged=%v err=%v", unchanged, err)
	}
	apply(t, vp)
	reapplied := voicePlan(t, "set", o, jarvisSirSubtle)
	if unchanged, err := PlanUnchanged(reapplied); err != nil || !unchanged {
		t.Fatalf("expected reapplying the same voice to be unchanged, got unchanged=%v err=%v", unchanged, err)
	}
}

func TestPlanAndStateWithoutVoiceKeepBytesAndIDs(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	if bytes.Contains(encode(p), []byte("voice")) {
		t.Fatal("a plan without voice must not mention voice in its encoding")
	}
	apply(t, p)
	stateBytes := get(t, filepath.Join(o.StateDir, "state.json"))
	if strings.Contains(stateBytes, "voice") {
		t.Fatal("a state without voice must not mention voice in its encoding")
	}
	// Rebuilding the identical plan from the identical source must still
	// reproduce the identical ID and Change payloads (id stability check).
	p2 := plan(t, "install", o)
	if apply(t, p2) != "unchanged" {
		t.Fatal("expected the reinstall to be unchanged")
	}
}

// --- recovery: voice-only entries, merged entries, outside-text preservation ---

func TestVoiceOnlyRecoveryAtEachWritePoint(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			codexBefore, claudeBefore := get(t, codexPath(o)), get(t, claudePath(o))
			vp := voicePlan(t, "set", o, jarvisSirSubtle)
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(vp); err == nil {
				t.Fatal("expected the injected failure to stop the apply")
			}
			if _, err := BuildVoicePlan("set", o, jarvisSirSubtle); err == nil {
				t.Fatal("pending operation ignored")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, codexPath(o)) != codexBefore {
				t.Fatal("Codex's file not restored byte for byte")
			}
			if get(t, claudePath(o)) != claudeBefore {
				t.Fatal("Claude's file not restored byte for byte")
			}
			if s := stateFor(t, o); s.Voice != nil || len(s.VoiceSpans) != 0 {
				t.Fatal("voice state not rolled back")
			}
			apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
			if !strings.Contains(get(t, codexPath(o)), VoiceBegin) {
				t.Fatal("voice set after recovery did not write the block")
			}
		})
	}
}

func TestVoiceRecoveryOfMergedHiveAndVoiceChange(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
			codexBefore, claudeBefore := get(t, codexPath(o)), get(t, claudePath(o))

			// Change both the Hive text and the voice text, so the same file
			// carries a Change and a VoiceChange in the same install entry.
			put(t, filepath.Join(o.Source, GlobalSource), "# New rules\nBoth blocks change together.\n")
			put(t, filepath.Join(o.Source, "content/voices/jarvis.md"), "A new tone entirely.\n")
			p := plan(t, "install", o)
			foundMerged := false
			for _, vc := range p.Voice {
				for _, ch := range p.Changes {
					if ch.Target.Path == vc.Path {
						foundMerged = true
					}
				}
			}
			if !foundMerged {
				t.Fatal("test setup expects a path with both a Change and a VoiceChange")
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
			if get(t, codexPath(o)) != codexBefore {
				t.Fatal("Codex's file not restored byte for byte")
			}
			if get(t, claudePath(o)) != claudeBefore {
				t.Fatal("Claude's file not restored byte for byte")
			}
			apply(t, plan(t, "install", o))
			if !strings.Contains(get(t, codexPath(o)), "Both blocks change together.") {
				t.Fatal("install after recovery did not apply the new Hive text")
			}
			if !strings.Contains(get(t, codexPath(o)), "A new tone entirely.") {
				t.Fatal("install after recovery did not apply the new voice text")
			}
		})
	}
}

func TestVoiceRecoveryPreservesOutsideEditMadeAfterInterruption(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	claudeBefore := get(t, claudePath(o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	// Entry order is deterministic by sorted path; Claude's ".claude/CLAUDE.md"
	// sorts after Codex's ".codex/AGENTS.md", so write:0 lands on Codex first.
	e := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(vp); err == nil {
		t.Fatal("expected the injected failure to stop the apply")
	}
	// Claude's entry was never written (still the plain Hive-only file);
	// simulate a user edit made outside every block after the interruption.
	edited := claudeBefore + "a note the user added after the crash\n"
	put(t, claudePath(o), edited)
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(get(t, claudePath(o)), "a note the user added after the crash\n") {
		t.Fatal("outside edit made after the interruption was lost")
	}
	if strings.Contains(get(t, claudePath(o)), VoiceBegin) {
		t.Fatal("the never-written voice block should not appear after rollback")
	}
}

func TestRecoveryOfDeletedMultiBlockFileFromRemove(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"codex"}
			voiceSource(t, o)
			apply(t, plan(t, "install", o))
			apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
			original := get(t, codexPath(o))

			p := plan(t, "remove", o)
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
			if get(t, codexPath(o)) != original {
				t.Fatalf("file not fully reconstructed after recovering an interrupted removal:\ngot:  %q\nwant: %q", get(t, codexPath(o)), original)
			}
			apply(t, plan(t, "remove", o))
			absent(t, codexPath(o))
		})
	}
}
