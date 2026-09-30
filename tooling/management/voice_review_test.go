package management

// Fix batch from the dedicated code review (/code-review high over
// 6fd741c..HEAD). All homes are synthetic; nothing here touches a real user
// home or the repository's real content/voices/.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- item 1: Recover must try the exact same(cur, en.After) restore before
// falling back to the block-by-block voice/Hive reconstruction -----------------

// Scenario (a): removing a file that carries Hive, voice and trailing user
// text, interrupted after the write lands, must restore the exact original
// byte layout (Hive block, then voice block, then the user's text) — not
// reconstruct it by appending the Hive block after whatever remains.
func TestVoiceRecoveryOfRemoveInterruptedAfterWriteRestoresExactLayout(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	withUser := get(t, codexPath(o)) + "user text\n"
	put(t, codexPath(o), withUser)
	original := get(t, codexPath(o))

	p := plan(t, "remove", o)
	e := Engine{failpoint: func(s string) error {
		if s == "state" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("expected the injected failure to stop the apply")
	}
	if got := get(t, codexPath(o)); got != "user text\n" {
		t.Fatalf("test setup: expected the write to have already landed (just the user text left), got %q", got)
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if got := get(t, codexPath(o)); got != original {
		t.Fatalf("recovery did not restore the exact original layout:\ngot:  %q\nwant: %q", got, original)
	}
}

// Scenario (c): recovering a removal that deleted a Hive-created file (both
// blocks removed, nothing left) must restore the file's original mode, not
// recompute a hardcoded default — exercised with a non-default mode so the
// two cannot coincide by chance.
func TestVoiceRecoveryOfDeletedFilePreservesOriginalMode(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	if err := os.Chmod(codexPath(o), 0640); err != nil {
		t.Fatal(err)
	}
	original := get(t, codexPath(o))
	info, err := os.Stat(codexPath(o))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Fatalf("test setup: expected mode 0640, got %o", info.Mode().Perm())
	}

	p := plan(t, "remove", o)
	e := Engine{failpoint: func(s string) error {
		if s == "state" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("expected the injected failure to stop the apply")
	}
	absent(t, codexPath(o))
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if get(t, codexPath(o)) != original {
		t.Fatal("recovery did not restore the exact original content")
	}
	info2, err := os.Stat(codexPath(o))
	if err != nil {
		t.Fatal(err)
	}
	if info2.Mode().Perm() != 0640 {
		t.Fatalf("recovery did not restore the original mode: got %o, want 0640", info2.Mode().Perm())
	}
}

// --- item 2: a voice that can no longer be rendered must not block install ---

// TestVoiceRenderFailureDuringInstallSkipsRegenerationWithWarning covers a
// voice file renamed or removed out from under an active choice: install
// must still succeed, leaving the stored spans untouched and surfacing a
// warning naming the voice.
func TestVoiceRenderFailureDuringInstallSkipsRegenerationWithWarning(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	spansBefore := stateFor(t, o).VoiceSpans

	if err := os.Remove(filepath.Join(o.Source, "content/voices/jarvis.md")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, GlobalSource), "# Changed\nUnrelated Hive text.\n")

	p, err := BuildPlan("install", o)
	if err != nil {
		t.Fatalf("install must not fail when the active voice can no longer be rendered: %v", err)
	}
	if len(p.Voice) != 0 {
		t.Fatalf("expected no voice changes when the voice cannot be rendered, got %d", len(p.Voice))
	}
	if p.VoiceWarning == "" || !strings.Contains(p.VoiceWarning, "jarvis") {
		t.Fatalf("expected a voice warning naming jarvis, got %q", p.VoiceWarning)
	}
	apply(t, p)
	if !reflect.DeepEqual(stateFor(t, o).VoiceSpans, spansBefore) {
		t.Fatal("voice spans changed despite the render failure")
	}
}

// TestVoiceRenderFailureFromMarkerInPreambleSkipsRegenerationWithWarning
// covers the other RenderVoice failure mode: an injected marker.
func TestVoiceRenderFailureFromMarkerInPreambleSkipsRegenerationWithWarning(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	spansBefore := stateFor(t, o).VoiceSpans

	put(t, filepath.Join(o.Source, "content/voices/preamble.md"), "Priority.\n"+Begin+"\ninjected\n"+End+"\n")
	put(t, filepath.Join(o.Source, GlobalSource), "# Changed\nUnrelated Hive text.\n")

	p, err := BuildPlan("install", o)
	if err != nil {
		t.Fatalf("install must not fail when the voice source is invalid: %v", err)
	}
	if len(p.Voice) != 0 {
		t.Fatalf("expected no voice changes, got %d", len(p.Voice))
	}
	if p.VoiceWarning == "" {
		t.Fatal("expected a voice warning")
	}
	apply(t, p)
	if !reflect.DeepEqual(stateFor(t, o).VoiceSpans, spansBefore) {
		t.Fatal("voice spans changed despite the render failure")
	}
}

// --- item 3: BuildVoicePlan must check the Hive block exists at plan time ----

// TestVoiceSetFailsAtPlanTimeWhenHiveBlockMissing covers a file whose
// registered Hive block was manually deleted (state still thinks it is
// registered — a drift the file's own block-status already reports): "voice
// set" must fail naming that file before any confirmation, not at apply
// time via a lower-level "no Hive block to attach to" error.
func TestVoiceSetFailsAtPlanTimeWhenHiveBlockMissing(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	// Delete the Hive block by hand, leaving only user text.
	put(t, codexPath(o), "user text only, no Hive block\n")

	_, err := BuildVoicePlan("set", o, jarvisSirSubtle)
	if err == nil {
		t.Fatal("expected voice set to fail when the target file has no Hive block")
	}
	if !strings.Contains(err.Error(), codexPath(o)) {
		t.Fatalf("expected the error to name the file %s, got %v", codexPath(o), err)
	}
}

func TestVoiceOffFailsAtPlanTimeWhenHiveBlockMissing(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
	// Delete both blocks by hand.
	put(t, codexPath(o), "user text only, no Hive block\n")

	_, err := BuildVoicePlan("off", o, VoiceSetting{})
	if err == nil {
		t.Fatal("expected voice off to fail when the target file has no Hive block")
	}
	if !strings.Contains(err.Error(), codexPath(o)) {
		t.Fatalf("expected the error to name the file %s, got %v", codexPath(o), err)
	}
}

// --- item 4: voice Record/error paths must name the file ---------------------

// TestVoiceRecordFromSpanNamesTheFileInErrors covers voiceRecordFromSpan's
// empty Target.Path: owned's "modified or missing managed block" error (and
// every other owned/transform error keyed on Target.Path) must name the
// actual file path exactly once, not a bare trailing colon.
func TestVoiceRecordFromSpanNamesTheFileInErrors(t *testing.T) {
	const path = "/some/synthetic/path/CLAUDE.md"
	span := &VoiceSpan{Managed: []byte(VoiceBegin + "\nbody\n" + VoiceEnd + "\n")}
	rec := voiceRecordFromSpan(path, span)
	if rec.Target.Path != path {
		t.Fatalf("expected voiceRecordFromSpan to set Target.Path, got %q", rec.Target.Path)
	}
	tampered := snapshot{Exists: true, Data: []byte(VoiceBegin + "\nTAMPERED\n" + VoiceEnd + "\n")}
	err := owned(tampered, *rec, voiceMarkers)
	if err == nil {
		t.Fatal("expected a mismatch error")
	}
	if n := strings.Count(err.Error(), path); n != 1 {
		t.Fatalf("expected the error to name the path %s exactly once, got %d times in %q", path, n, err.Error())
	}
}

// --- item 7: an unreadable voice file must not abort the whole status --------

// TestVoiceStatusUnreadableFileBecomesDriftNotAbort covers a voice-span path
// that Status's own block-resolution no longer reaches (shadowed by a
// Codex AGENTS.override.md, so status's resolve() step, which scans every
// currently resolved block file for shared imports, never touches the
// original file at all) but whose voice span is still registered and read
// directly: if that original file becomes unreadable, its row must report
// "drift", not abort the rest of status.
func TestVoiceStatusUnreadableFileBecomesDriftNotAbort(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "claude"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))

	put(t, filepath.Join(o.Home, ".codex", "AGENTS.override.md"), "override content\n")
	if err := os.Remove(codexPath(o)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(codexPath(o), 0700); err != nil {
		t.Fatal(err)
	}

	entries, err := Status(o)
	if err != nil {
		t.Fatalf("Status must not abort when a shadowed voice file is unreadable: %v", err)
	}
	row := voiceStatusFor(entries, codexPath(o))
	if row == nil || row.Status != "drift" {
		t.Fatalf("expected a drift voice row for the unreadable file, got %+v", row)
	}
	if voiceStatusFor(entries, claudePath(o)) == nil {
		t.Fatal("the other voice row must still be reported")
	}
}

// --- item 6: State.Voice clears only when no Hive block remains --------------

// TestVoiceSettingSurvivesWhenAHiveBlockRemainsWithoutASpan constructs a host
// with a Hive block but no voice span (installed while the voice could not
// be rendered, see item 2), then removes the only host that DOES have a
// span: State.Voice must survive, since a Hive block is still registered and
// should regain a span once rendering works again — clearing must depend on
// whether any Hive block remains, not merely on VoiceSpans becoming empty.
func TestVoiceSettingSurvivesWhenAHiveBlockRemainsWithoutASpan(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	apply(t, voicePlan(t, "set", o, jarvisSirSubtle))

	if err := os.Remove(filepath.Join(o.Source, "content/voices/jarvis.md")); err != nil {
		t.Fatal(err)
	}
	claudeOnly := o
	claudeOnly.Hosts = []string{"claude"}
	apply(t, plan(t, "install", claudeOnly))
	if _, ok := stateFor(t, o).VoiceSpans[claudePath(o)]; ok {
		t.Fatal("test setup: Claude must not have gotten a span while the voice could not be rendered")
	}
	if _, ok := stateFor(t, o).Records[claudePath(o)]; !ok {
		t.Fatal("test setup: Claude's Hive block must be registered")
	}

	codexOnly := o
	codexOnly.Hosts = []string{"codex"}
	apply(t, plan(t, "remove", codexOnly))
	if len(stateFor(t, o).VoiceSpans) != 0 {
		t.Fatal("test setup: expected no voice spans left after removing Codex")
	}
	if stateFor(t, o).Voice == nil {
		t.Fatal("State.Voice was cleared even though Claude's Hive block is still registered")
	}
}

// --- item 8: validatePlan must validate VoiceSetting and restrict vc.Path ----

// TestValidatePlanRejectsInvalidVoiceSetting covers a hand-crafted (e.g.
// tampered --out) plan whose VoiceSetting has an invalid address, an
// invalid intensity, or an ID that fails the same pattern RenderVoice
// enforces: validatePlan must reject each before any write.
func TestValidatePlanRejectsInvalidVoiceSetting(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))

	for _, bad := range []VoiceSetting{
		{ID: "Not Valid!", Address: "none", Intensity: "subtle"},
		{ID: "jarvis", Address: "loud", Intensity: "subtle"},
		{ID: "jarvis", Address: "none", Intensity: "screaming"},
		{ID: "jarvis", Address: "name", Name: "", Intensity: "subtle"},
	} {
		vp := voicePlan(t, "set", o, jarvisSirSubtle)
		setting := bad
		vp.VoiceSetting = &setting
		vp.ID = planID(vp)
		if err := validatePlan(vp, stateFor(t, o)); err == nil {
			t.Fatalf("expected validatePlan to reject VoiceSetting %+v", bad)
		}
	}
}

// TestValidatePlanRestrictsVoicePathsToRegisteredHiveBlocks covers a
// hand-crafted plan whose VoiceChange.Path does not name a registered
// user-scope Hive block at Config.Home: validatePlan must reject it, so an
// arbitrary path can never receive a voice span through a tampered plan.
func TestValidatePlanRestrictsVoicePathsToRegisteredHiveBlocks(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)

	outside := filepath.Join(o.Home, "not-a-hive-file.md")
	vp.Voice[0].Path = outside
	vp.ID = planID(vp)
	if err := validatePlan(vp, stateFor(t, o)); err == nil {
		t.Fatal("expected validatePlan to reject a voice change on a path with no registered Hive block")
	}
}

// TestVoiceOnlyRecoveryStillUsesExactRestoreShortcut confirms the reordering
// does not regress the plain voice-only case: it must still restore exactly.
func TestVoiceOnlyRecoveryStillUsesExactRestoreShortcut(t *testing.T) {
	o := setup(t)
	voiceSource(t, o)
	apply(t, plan(t, "install", o))
	codexBefore := get(t, codexPath(o))
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	e := Engine{failpoint: func(s string) error {
		if s == "state" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(vp); err == nil {
		t.Fatal("expected the injected failure to stop the apply")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if get(t, codexPath(o)) != codexBefore {
		t.Fatal("voice-only recovery did not restore the exact original bytes")
	}
}
