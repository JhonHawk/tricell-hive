package main

// TDD tests for `hive voice list|set|off` and the voice line in `hive
// update`'s summary (T4). All homes are synthetic; nothing here touches a
// real user home or the repository's real content/voices/ except the list
// test, which reads it read-only.

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// voiceCLIEnv mirrors update_test.go's updateEnv: a synthetic home, state
// directory and source checkout, with a voices/ tree added.
type voiceCLIEnv struct {
	home, stateDir, source string
}

func newVoiceCLIEnv(t *testing.T) voiceCLIEnv {
	t.Helper()
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := voiceCLIEnv{home: filepath.Join(base, "home"), stateDir: filepath.Join(base, "state"), source: filepath.Join(base, "source")}
	if err := os.MkdirAll(e.home, 0700); err != nil {
		t.Fatal(err)
	}
	putCharacterization(t, filepath.Join(e.source, management.GlobalSource), "# Rules\nKeep user content.\n")
	putCharacterization(t, filepath.Join(e.source, management.SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nPreserve evidence.\n")
	putCharacterization(t, filepath.Join(e.source, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	putCharacterization(t, filepath.Join(e.source, "content/voices/jarvis.md"), "Jarvis: warm, formal, a little dry.\n")
	return e
}

func (e voiceCLIEnv) args(extra ...string) []string {
	return append([]string{"--home", e.home, "--state-dir", e.stateDir, "--source", e.source}, extra...)
}

// setArgs builds "voice set" arguments with id first, per design.md's
// documented syntax ("hive voice set <id> [--address ...] ..."), followed by
// any extra flags and then the common --home/--state-dir/--source ones.
func (e voiceCLIEnv) setArgs(id string, extra ...string) []string {
	args := append([]string{id}, extra...)
	return append(args, "--home", e.home, "--state-dir", e.stateDir, "--source", e.source)
}
func (e voiceCLIEnv) codexPath() string  { return filepath.Join(e.home, ".codex", "AGENTS.md") }
func (e voiceCLIEnv) claudePath() string { return filepath.Join(e.home, ".claude", "CLAUDE.md") }

func (e voiceCLIEnv) install(t *testing.T, hosts []string) {
	t.Helper()
	installDirect(t, e.home, e.stateDir, e.source, hosts)
}

// --- AC1: list --------------------------------------------------------------

// TestVoiceListShowsRepositoryVoices covers AC1 against the repository's
// own real content/voices/ (read-only), which must list jarvis, mentor and
// senior-direct.
func TestVoiceListShowsRepositoryVoices(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	got := captureStdout(t, func() {
		if err := run([]string{"voice", "list", "--source", root}); err != nil {
			t.Fatalf("voice list: %v", err)
		}
	})
	for _, id := range []string{"jarvis", "mentor", "senior-direct"} {
		if !strings.Contains(got, id) {
			t.Fatalf("expected voice list to mention %q, got:\n%s", id, got)
		}
	}
	if strings.Contains(got, "preamble") {
		t.Fatalf("voice list must not mention preamble.md, got:\n%s", got)
	}
}

// TestVoiceListStripsDuplicateNamePrefix covers the "Name: " duplicate the
// brief calls out: a synthetic voice file starting with "Jarvis: ..." must
// not be printed as "jarvis: Jarvis: ...".
func TestVoiceListStripsDuplicateNamePrefix(t *testing.T) {
	e := newVoiceCLIEnv(t)
	got := captureStdout(t, func() {
		if err := run([]string{"voice", "list", "--source", e.source}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(got, "Jarvis: Jarvis:") {
		t.Fatalf("expected the duplicate name prefix to be stripped, got:\n%s", got)
	}
	if !strings.Contains(got, "jarvis:") || !strings.Contains(got, "warm, formal") {
		t.Fatalf("expected the ID and the rest of the description, got:\n%s", got)
	}
}

// --- AC2/AC3: set and off ---------------------------------------------------

// TestVoiceSetInteractiveWritesBlockAfterHive covers AC2: in an interactive
// terminal, `hive voice set` writes the voice block right after the Hive
// block on both Codex and Claude, and does not change the installed
// release ID.
func TestVoiceSetInteractiveWritesBlockAfterHive(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex", "claude"})
	o := management.Options{Scope: "user", Home: e.home, StateDir: e.stateDir, Hosts: []string{"codex", "claude"}}
	before, err := management.Status(o)
	if err != nil {
		t.Fatal(err)
	}
	releaseBefore := map[string]string{}
	outsideBefore := map[string][]byte{}
	for _, en := range before {
		if en.Kind == "block" {
			releaseBefore[en.Path] = en.Release
		}
	}
	for _, path := range []string{e.codexPath(), e.claudePath()} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		outsideBefore[path] = body
	}

	var out bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis", "--address", "sir", "--intensity", "subtle"), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, out.String())
	}

	after, err := management.Status(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range after {
		if en.Kind != "block" {
			continue
		}
		if en.Release != releaseBefore[en.Path] {
			t.Fatalf("installed release ID changed for %s: %q -> %q", en.Path, releaseBefore[en.Path], en.Release)
		}
	}

	for _, path := range []string{e.codexPath(), e.claudePath()} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if strings.Count(text, management.VoiceBegin) != 1 {
			t.Fatalf("%s: expected exactly one voice block, got:\n%s", path, text)
		}
		endIdx := strings.Index(text, management.End)
		beginIdx := strings.Index(text, management.VoiceBegin)
		if strings.TrimSpace(text[endIdx+len(management.End):beginIdx]) != "" {
			t.Fatalf("%s: voice block not immediately after the Hive block:\n%s", path, text)
		}
		if prefix := text[:beginIdx]; prefix != string(outsideBefore[path]) {
			t.Fatalf("%s: bytes before the voice block changed:\ngot:  %q\nwant: %q", path, prefix, outsideBefore[path])
		}
		if !strings.Contains(text, "Address:") {
			t.Fatalf("%s: expected an \"Address:\" line, got:\n%s", path, text)
		}
		if !strings.Contains(text, "Intensity:") {
			t.Fatalf("%s: expected an \"Intensity:\" line, got:\n%s", path, text)
		}
	}
	if !strings.Contains(out.String(), "jarvis") {
		t.Fatalf("expected the summary to name the voice, got:\n%s", out.String())
	}
}

// TestVoiceSetDeclinedAtConfirmationChangesNothing covers G2: answering "n"
// at the confirmation prompt leaves every file untouched.
func TestVoiceSetDeclinedAtConfirmationChangesNothing(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	before, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis"), strings.NewReader("n\n"), &out, true); err != nil {
		t.Fatalf("voice set (declined): %v", err)
	}
	after, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("declining the confirmation must not change the file")
	}
	if !strings.Contains(out.String(), "Cancelled") {
		t.Fatalf("expected a cancellation notice, got:\n%s", out.String())
	}
}

// TestVoiceSetRepeatSameChoiceIsUnchangedWithoutConfirmation covers G3: a
// second "voice set" with the exact same choice is reported unchanged and
// exits 0 without writing, in a NON-interactive session with neither
// --dry-run nor --out — if the unchanged short-circuit were removed, this
// call would instead fail requiring a terminal, so this specifically
// exercises that branch rather than any of the others.
func TestVoiceSetRepeatSameChoiceIsUnchangedWithoutConfirmation(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	var firstOut bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis", "--address", "sir", "--intensity", "subtle"), strings.NewReader("y\n"), &firstOut, true); err != nil {
		t.Fatalf("first voice set: %v", err)
	}
	before, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis", "--address", "sir", "--intensity", "subtle"), strings.NewReader(""), &out, false); err != nil {
		t.Fatalf("repeat voice set: %v\noutput:\n%s", err, out.String())
	}
	after, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("repeating the identical choice changed the file")
	}
	if !strings.Contains(out.String(), "Nothing to change") {
		t.Fatalf("expected the unchanged notice, got:\n%s", out.String())
	}
}

// TestVoiceOffRestoresBytes covers AC3: `hive voice off` restores every
// file to its exact pre-set bytes.
func TestVoiceOffRestoresBytes(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex", "claude"})
	codexBefore, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	claudeBefore, err := os.ReadFile(e.claudePath())
	if err != nil {
		t.Fatal(err)
	}
	var setOut bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v", err)
	}
	var offOut bytes.Buffer
	if err := voiceOff(e.args(), strings.NewReader("y\n"), &offOut, true); err != nil {
		t.Fatalf("voice off: %v\noutput:\n%s", err, offOut.String())
	}
	codexAfter, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	claudeAfter, err := os.ReadFile(e.claudePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(codexBefore, codexAfter) {
		t.Fatalf("Codex bytes not restored:\ngot:  %q\nwant: %q", codexAfter, codexBefore)
	}
	if !bytes.Equal(claudeBefore, claudeAfter) {
		t.Fatalf("Claude bytes not restored:\ngot:  %q\nwant: %q", claudeAfter, claudeBefore)
	}
}

// --- --dry-run and --out ----------------------------------------------------

// TestVoiceSetDryRunChangesNothing covers --dry-run, with and without an
// interactive terminal.
func TestVoiceSetDryRunChangesNothing(t *testing.T) {
	for _, interactive := range []bool{true, false} {
		t.Run(map[bool]string{true: "interactive", false: "non-interactive"}[interactive], func(t *testing.T) {
			e := newVoiceCLIEnv(t)
			e.install(t, []string{"codex", "claude"})
			before, err := os.ReadFile(e.codexPath())
			if err != nil {
				t.Fatal(err)
			}
			var out bytes.Buffer
			if err := voiceSet(e.setArgs("jarvis", "--dry-run"), strings.NewReader(""), &out, interactive); err != nil {
				t.Fatalf("voice set --dry-run: %v", err)
			}
			after, err := os.ReadFile(e.codexPath())
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, after) {
				t.Fatal("--dry-run changed the installed file")
			}
			if !strings.Contains(out.String(), "Preview") {
				t.Fatalf("expected a preview notice, got:\n%s", out.String())
			}
		})
	}
}

// TestVoiceSetOutThenApply covers --out: it saves an applicable plan and
// changes nothing itself.
func TestVoiceSetOutThenApply(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex", "claude"})
	planFile := filepath.Join(filepath.Dir(e.stateDir), "voice.plan.json")
	var out bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis", "--out", planFile), strings.NewReader(""), &out, false); err != nil {
		t.Fatalf("voice set --out: %v", err)
	}
	if !strings.Contains(out.String(), "hive apply --plan "+planFile) {
		t.Fatalf("expected the output to name hive apply --plan %s, got:\n%s", planFile, out.String())
	}
	if _, err := os.ReadFile(e.codexPath()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(before), management.VoiceBegin) {
		t.Fatal("--out must not change the installed file")
	}
	if err := run([]string{"apply", "--plan", planFile}); err != nil {
		t.Fatalf("hive apply --plan %s: %v", planFile, err)
	}
	after, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(after), management.VoiceBegin) {
		t.Fatal("applying the saved plan did not write the voice block")
	}
}

// TestVoiceSetNonInteractiveRequiresOutOrDryRun mirrors update's AC3 half.
func TestVoiceSetNonInteractiveRequiresOutOrDryRun(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex", "claude"})
	err := voiceSet(e.setArgs("jarvis"), strings.NewReader(""), io.Discard, false)
	if err == nil || !strings.Contains(err.Error(), "--out") || !strings.Contains(err.Error(), "--dry-run") {
		t.Fatalf("expected an error naming --dry-run and --out, got %v", err)
	}
	if body, rerr := os.ReadFile(e.codexPath()); rerr != nil || strings.Contains(string(body), management.VoiceBegin) {
		t.Fatalf("a rejected non-interactive set must not change the file: %v %q", rerr, body)
	}
}

// --- AC9: errors before writing ----------------------------------------------

func TestVoiceSetUnknownVoiceFailsBeforeWriting(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	before, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := voiceSet(e.setArgs("nonexistent"), strings.NewReader("y\n"), io.Discard, true); err == nil {
		t.Fatal("expected an error for an unknown voice")
	}
	after, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("a rejected unknown-voice set must not change the file")
	}
}

func TestVoiceSetAddressNameWithoutNameFailsBeforeWriting(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	before, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if err := voiceSet(e.setArgs("jarvis", "--address", "name"), strings.NewReader("y\n"), io.Discard, true); err == nil {
		t.Fatal("expected an error for --address name without --name")
	}
	after, err := os.ReadFile(e.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("a rejected set must not change the file")
	}
}

func TestVoiceSetWithoutHiveInstalledFails(t *testing.T) {
	e := newVoiceCLIEnv(t)
	if err := voiceSet(e.setArgs("jarvis"), strings.NewReader("y\n"), io.Discard, true); err == nil {
		t.Fatal("expected an error when Hive is not installed for any host")
	}
	if _, err := os.Stat(e.stateDir); !os.IsNotExist(err) {
		t.Fatalf("expected no state directory to be created by a rejected set, stat err=%v", err)
	}
}

// --- usage errors -------------------------------------------------------------

func TestVoiceUnknownSubcommandAndMissingID(t *testing.T) {
	if err := run([]string{"voice", "frobnicate"}); err == nil {
		t.Fatal("expected an error for an unknown voice subcommand")
	}
	if err := run([]string{"voice", "set"}); err == nil {
		t.Fatal("expected an error when voice set has no ID")
	}
	if err := run([]string{"voice", "off", "extra"}); err == nil {
		t.Fatal("expected an error for an unexpected positional argument to voice off")
	}
	if err := run([]string{"voice", "list", "extra"}); err == nil {
		t.Fatal("expected an error for an unexpected positional argument to voice list")
	}
}

// --- item 9: the summary must use the plan's own normalized VoiceSetting ----

// TestVoiceSetSummaryUsesPlanNormalizedDefaults covers leaving --address and
// --intensity unset: the summary must show "none"/"subtle" sourced from
// management.BuildVoicePlan's own normalization (p.VoiceSetting), not a
// second, separately maintained default in the CLI layer.
func TestVoiceSetSummaryUsesPlanNormalizedDefaults(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	var out bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis"), strings.NewReader(""), &out, false); err == nil {
		t.Fatal("expected the non-interactive, no --dry-run/--out call to fail (checking the summary text first)")
	}
	if !strings.Contains(out.String(), "Voice: jarvis (address none, intensity subtle)") {
		t.Fatalf("expected the exact normalized summary line, got:\n%s", out.String())
	}
}

// --- item 5: `hive plan install|remove` preview must show voice changes ------

// TestPlanInstallPreviewPrintsVoiceChanges covers `hive plan install`'s raw
// preview: it must show a voice change's path and its before/after managed
// text the same way it already shows a block Change, so `plan --out` +
// `apply` never applies a voice change the operator never saw.
func TestPlanInstallPreviewPrintsVoiceChanges(t *testing.T) {
	e := newVoiceCLIEnv(t)
	e.install(t, []string{"codex"})
	var setOut bytes.Buffer
	if err := voiceSet(e.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, setOut.String())
	}
	putCharacterization(t, filepath.Join(e.source, "content/voices/jarvis.md"), "Jarvis: a brand new tone.\n")

	out := captureStdout(t, func() {
		if err := run([]string{"plan", "install", "--scope", "user", "--home", e.home, "--source", e.source, "--hosts", "codex", "--state-dir", e.stateDir}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(out, "voice "+e.codexPath()) {
		t.Fatalf("expected the preview to name the voice change for %s, got:\n%s", e.codexPath(), out)
	}
	if !strings.Contains(out, "a brand new tone") {
		t.Fatalf("expected the preview to show the new voice text, got:\n%s", out)
	}
}

// --- update's summary line ----------------------------------------------------

// TestUpdateSummaryShowsVoiceLineWhenTextChanged covers AC4's update-summary
// half: with a voice active, if its source text changed since the last
// install, `hive update`'s summary mentions the voice change.
func TestUpdateSummaryShowsVoiceLineWhenTextChanged(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	putCharacterization(t, filepath.Join(env.repo, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: warm, formal, a little dry.\n")
	gitInitLocal(t, env.repo)
	commit1 := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit1, []string{"codex", "claude"})

	var setOut bytes.Buffer
	voiceEnv := voiceCLIEnv{home: env.home, stateDir: env.stateDir, source: env.repo}
	if err := voiceSet(voiceEnv.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, setOut.String())
	}

	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: a completely different tone.\n")
	gitCommitAll(t, env.repo, "change voice text")

	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, out.String())
	}
	// Exact line, not a loose "Voice" substring: both hosts (codex, claude)
	// must regenerate, since the voice text itself changed.
	if !strings.Contains(out.String(), "Voice files to regenerate: 2\n") {
		t.Fatalf("expected the exact voice-regeneration line for 2 files, got:\n%s", out.String())
	}
	body, err := os.ReadFile(voiceEnv.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "a completely different tone") {
		t.Fatal("update did not regenerate the voice text")
	}
}

// TestUpdateSummaryShowsVoiceWarningWhenVoiceCannotBeRendered covers a voice
// file removed out from under an active choice: update must still succeed
// (not fail the whole operation), the summary must surface a warning naming
// the voice, and the already-installed span must stay untouched.
func TestUpdateSummaryShowsVoiceWarningWhenVoiceCannotBeRendered(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	putCharacterization(t, filepath.Join(env.repo, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: warm, formal, a little dry.\n")
	gitInitLocal(t, env.repo)
	commit1 := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit1, []string{"codex", "claude"})

	voiceEnv := voiceCLIEnv{home: env.home, stateDir: env.stateDir, source: env.repo}
	var setOut bytes.Buffer
	if err := voiceSet(voiceEnv.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, setOut.String())
	}

	if err := os.Remove(filepath.Join(env.repo, "content/voices/jarvis.md")); err != nil {
		t.Fatal(err)
	}
	putCharacterization(t, filepath.Join(env.repo, management.SkillSource), skillBody("Preserve evidence, updated.\n"))
	gitCommitAll(t, env.repo, "remove voice file")

	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update must not fail when the voice cannot be rendered: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "jarvis") {
		t.Fatalf("expected a voice warning naming jarvis, got:\n%s", out.String())
	}
	body, err := os.ReadFile(voiceEnv.codexPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "warm, formal") {
		t.Fatal("the existing voice span must remain untouched when the voice cannot be rendered")
	}
}

// TestUpdateSummaryCountsOnlyChangedHiveFilesOnVoiceOnlyUpdate covers a
// voice-only update: every p.Changes entry is a no-op (nothing about Hive
// itself changed), so the Hive count must read 0 (and say the core is
// already up to date), while the voice line is still present.
func TestUpdateSummaryCountsOnlyChangedHiveFilesOnVoiceOnlyUpdate(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	putCharacterization(t, filepath.Join(env.repo, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: warm, formal, a little dry.\n")
	gitInitLocal(t, env.repo)
	commit1 := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit1, []string{"codex", "claude"})

	voiceEnv := voiceCLIEnv{home: env.home, stateDir: env.stateDir, source: env.repo}
	var setOut bytes.Buffer
	if err := voiceSet(voiceEnv.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, setOut.String())
	}

	// Only the voice text changes; nothing else in the catalogue does.
	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: a completely different tone.\n")
	gitCommitAll(t, env.repo, "change voice text only")

	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "Hive files to install or update:") {
		t.Fatalf("expected no inflated \"to install or update\" count for a voice-only update, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "; none change.\n") {
		t.Fatalf("expected the Hive file count to say none changed, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Hive's core is already up to date.") {
		t.Fatalf("expected the header to say the Hive core is already up to date, got:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Voice files to regenerate: 2\n") {
		t.Fatalf("expected the voice line to still be present, got:\n%s", out.String())
	}
}

// TestUpdateSummaryOmitsVoiceLineWhenVoiceTextUnchanged is the negative half
// of AC4: when a commit changes something else (here, the shared skill) but
// not the voice text, the update summary must not mention regenerating any
// voice file.
func TestUpdateSummaryOmitsVoiceLineWhenVoiceTextUnchanged(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	putCharacterization(t, filepath.Join(env.repo, "content/voices/preamble.md"), "Priority: after Hive's rules, never overriding them.\n")
	putCharacterization(t, filepath.Join(env.repo, "content/voices/jarvis.md"), "Jarvis: warm, formal, a little dry.\n")
	gitInitLocal(t, env.repo)
	commit1 := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit1, []string{"codex", "claude"})

	voiceEnv := voiceCLIEnv{home: env.home, stateDir: env.stateDir, source: env.repo}
	var setOut bytes.Buffer
	if err := voiceSet(voiceEnv.setArgs("jarvis"), strings.NewReader("y\n"), &setOut, true); err != nil {
		t.Fatalf("voice set: %v\noutput:\n%s", err, setOut.String())
	}

	// Change only the shared skill; the voice text (preamble and jarvis.md)
	// is untouched.
	putCharacterization(t, filepath.Join(env.repo, management.SkillSource), skillBody("Preserve evidence, updated.\n"))
	gitCommitAll(t, env.repo, "change skill only")

	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "Voice files to regenerate") {
		t.Fatalf("expected no voice-regeneration line when the voice text did not change, got:\n%s", out.String())
	}
}

// TestVoiceSummaryCountsOnlyFilesThatChange: a voice plan carries one entry
// per file, including files already at the chosen voice; the summary must
// count and list only the files whose voice block actually changes.
func TestVoiceSummaryCountsOnlyFilesThatChange(t *testing.T) {
	span := &management.VoiceSpan{Managed: []byte("voice\n"), SourceHash: "h"}
	p := management.Plan{
		VoiceSetting: &management.VoiceSetting{ID: "jarvis", Address: "none", Intensity: "subtle"},
		Voice: []management.VoiceChange{
			{Path: "/home/.codex/AGENTS.md", Before: span, After: span},
			{Path: "/home/.claude/CLAUDE.md", After: span},
		},
	}
	var out bytes.Buffer
	showVoiceSummary(&out, p, false)
	if !strings.Contains(out.String(), "Voice files to change: 1\n") {
		t.Fatalf("expected one file to change, got:\n%s", out.String())
	}
	if strings.Contains(out.String(), "/home/.codex/AGENTS.md") {
		t.Fatalf("an unchanged file was listed:\n%s", out.String())
	}
}
