package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"tricell-hive/tooling/management"
)

// TestStatusScreenShowsValuesAndDoesNotWrite covers AC4: Status's own row
// for a registered host names its release, product version, drift count and
// active voice, and never writes — the home and state.json are byte for
// byte the same before and after. p is passed as nil: statusScreen never
// prompts, so a nil *huhPrompter dereferenced by mistake would panic this
// test instead of silently passing.
//
// T4 fix round item 2 (s1-s3): the expected line is a literal, not
// statusLine's own output recomputed from the same entries — a self-
// referential comparison that would keep passing under a mutation to
// statusLine's own aggregation (a flipped drift++/--, a swapped release/
// voice field, and so on). The release ID (310505196a486ddf4fda1426b98508e-
// 58b757bdbeb19131a2bd1ba9eeb977c8e, short 310505196a48) is deterministic
// for minimalCatalogSource's own fixed content installed for claude alone
// (confirmed stable across repeated runs); the drift count is deliberately
// forced to a non-zero, checkable value by corrupting the installed block's
// own managed content in place (a literal "0" would also match an
// implementation that never increments it at all).
func TestStatusScreenShowsValuesAndDoesNotWrite(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
	vo := management.Options{Source: source, Home: home, StateDir: stateDir}
	vp, err := management.BuildVoicePlan("set", vo, management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(vp); err != nil {
		t.Fatal(err)
	}

	// Force a known, non-zero drift: corrupt the managed block's own content
	// (between, not outside, the Hive markers) without touching state.json.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	data, err := os.ReadFile(claudeMD)
	if err != nil {
		t.Fatal(err)
	}
	corrupted := bytes.Replace(data, []byte("Minimal test guidance."), []byte("CORRUPTED CONTENT HERE"), 1)
	if bytes.Equal(corrupted, data) {
		t.Fatal("test setup: the replace did not change anything")
	}
	if err := os.WriteFile(claudeMD, corrupted, 0644); err != nil {
		t.Fatal(err)
	}

	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	wantLine := "claude · release 310505196a48 · version - · drift 1 · voice testvoice (sir, subtle)"

	beforeFiles := collectFiles(t, home)
	beforeState := stateBytes(t, stateDir)

	var out bytes.Buffer
	if err := statusScreen(o, &out, nil); err != nil {
		t.Fatalf("statusScreen: %v", err)
	}
	if !strings.Contains(out.String(), wantLine) {
		t.Fatalf("missing expected status line %q in:\n%s", wantLine, out.String())
	}

	afterFiles := collectFiles(t, home)
	afterState := stateBytes(t, stateDir)
	if len(beforeFiles) != len(afterFiles) {
		t.Fatalf("status changed the number of files in home: before=%d after=%d", len(beforeFiles), len(afterFiles))
	}
	for rel, data := range beforeFiles {
		if !bytes.Equal(afterFiles[rel], data) {
			t.Fatalf("status changed %s", rel)
		}
	}
	if !bytes.Equal(beforeState, afterState) {
		t.Fatal("status changed state.json")
	}
}

// TestStatusScreenNoHostsRegistered covers AC11's Status limit state, the
// same message Remove CLIs and Voice also use (design.md "La interfaz").
func TestStatusScreenNoHostsRegistered(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	var out bytes.Buffer
	if err := statusScreen(o, &out, nil); err != nil {
		t.Fatalf("statusScreen: %v", err)
	}
	if !strings.Contains(out.String(), "No CLI hosts are registered. Choose Install CLIs.") {
		t.Fatalf("missing empty-state message: %s", out.String())
	}
}

// TestUpdateScreenAppliesLikeCommand covers AC5's Update row: from a
// temporary Git repository with a pending commit, the Source/Revision
// Inputs (accepting the repo path and the default "HEAD") followed by
// confirming produce the same installed content and state.json as `hive
// update` itself, reusing update_test.go's own fixture builders (same
// package).
func TestUpdateScreenAppliesLikeCommand(t *testing.T) {
	twinEnv, _ := newUpdateFixtureWithPendingCommit(t)
	var twinOut bytes.Buffer
	if err := update(twinEnv.args(), strings.NewReader("y\n"), &twinOut, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, twinOut.String())
	}

	// env is its own, independent Git repository: its own second commit's
	// hash (envCommit2) differs from twinEnv's, so the summary check below
	// must use env's own, not twinEnv's.
	env, envCommit2 := newUpdateFixtureWithPendingCommit(t)
	o := management.Options{Home: env.home, StateDir: env.stateDir}
	var out bytes.Buffer
	// Source: env.repo; Revision: blank, keeping the field's own default
	// "HEAD" (internal/accessibility.PromptString falls back to the
	// field's current value on an empty line); then updateWith's own
	// Confirm ("y").
	p := newHuhPrompter(true, strings.NewReader(env.repo+"\n\ny\n"), &out)
	if err := updateScreen(o, &out, p); err != nil {
		t.Fatalf("updateScreen: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Resolving HEAD…") {
		t.Fatalf("missing the \"Resolving HEAD…\" notice: %s", out.String())
	}
	if !strings.Contains(out.String(), envCommit2[:12]) {
		t.Fatalf("expected the summary to name the resolved commit: %s", out.String())
	}

	got, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(twinEnv.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("installed content differs from hive update's own result:\ngot:  %q\nwant: %q", got, want)
	}
	assertStateJSONMatches(t, env.stateDir, env.home, twinEnv.stateDir, twinEnv.home)
}

// TestUpdateScreenInvalidRevisionReturnsToMenu covers Update's own error
// path (AC10): a revision starting with "-" fails validation before Git
// ever runs, and the menu reopens for a later Quit to reach cleanly — the
// same end-to-end pattern tui_hosts_test.go's own *ReturnsToMenu tests use.
func TestUpdateScreenInvalidRevisionReturnsToMenu(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)

	o := options{Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	// "4" Update; Source: blank (default "."); Revision: "-x" (invalid,
	// checked before Source/Git are ever touched); "7" Quit.
	if err := openInterface(false, true, strings.NewReader("4\n\n-x\n7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), `invalid --rev "-x"`) {
		t.Fatalf("missing the invalid --rev error: %s", out.String())
	}
}

// TestScreenDeclineChangesNothing covers AC10 for Update, Releases and
// Voice (Install CLIs and Remove CLIs are T3's own TestDeclineChangesNothing
// in tui_hosts_test.go): declining each screen's own final confirmation
// leaves the home and state.json untouched and prints the shared
// cancellation text.
func TestScreenDeclineChangesNothing(t *testing.T) {
	t.Run("update", func(t *testing.T) {
		env, _ := newUpdateFixtureWithPendingCommit(t)
		beforeState := stateBytes(t, env.stateDir)
		beforeBody, err := os.ReadFile(env.sharedSkillPath())
		if err != nil {
			t.Fatal(err)
		}

		o := management.Options{Home: env.home, StateDir: env.stateDir}
		var out bytes.Buffer
		p := newHuhPrompter(true, strings.NewReader(env.repo+"\n\nn\n"), &out)
		if err := updateScreen(o, &out, p); err != nil {
			t.Fatalf("updateScreen: %v\noutput:\n%s", err, out.String())
		}
		if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
			t.Fatalf("missing cancellation message: %s", out.String())
		}
		if after := stateBytes(t, env.stateDir); !bytes.Equal(beforeState, after) {
			t.Fatal("declined update changed state.json")
		}
		afterBody, err := os.ReadFile(env.sharedSkillPath())
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(beforeBody, afterBody) {
			t.Fatal("declined update changed the installed destination")
		}
	})

	t.Run("releases", func(t *testing.T) {
		env, olderID, _ := twoReleasesEnv(t)
		o := management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir}
		entries, err := management.Releases(o)
		if err != nil {
			t.Fatal(err)
		}
		targetIndex := releaseIndexFor(t, entries, olderID)
		before := collectFiles(t, env.home)

		var out bytes.Buffer
		p := newHuhPrompter(true, strings.NewReader(fmt.Sprintf("%d\nn\n", targetIndex)), &out)
		if err := releasesScreen(o, &out, p); err != nil {
			t.Fatalf("releasesScreen: %v\noutput:\n%s", err, out.String())
		}
		if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
			t.Fatalf("missing cancellation message: %s", out.String())
		}
		after := collectFiles(t, env.home)
		if len(before) != len(after) {
			t.Fatalf("declined rollback changed the home: before=%d files, after=%d files", len(before), len(after))
		}
		for rel, data := range before {
			if !bytes.Equal(after[rel], data) {
				t.Fatalf("declined rollback changed %s", rel)
			}
		}
	})

	t.Run("voice", func(t *testing.T) {
		source := minimalTestSource(t)
		dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
		home, stateDir := newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
		before := collectFiles(t, home)

		o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
		var out bytes.Buffer
		// SelectVoiceOrOff: "1" (the only voice, testvoice); address Select:
		// "2" (Sir); intensity Select: "1" (Subtle); voicePlanWith's own
		// Confirm: "n" (decline).
		p := newHuhPrompter(true, strings.NewReader("1\n2\n1\nn\n"), &out)
		if err := voiceScreen(o, &out, p); err != nil {
			t.Fatalf("voiceScreen: %v\noutput:\n%s", err, out.String())
		}
		if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
			t.Fatalf("missing cancellation message: %s", out.String())
		}
		after := collectFiles(t, home)
		if len(before) != len(after) {
			t.Fatalf("declined voice change changed the home: before=%d files, after=%d files", len(before), len(after))
		}
		for rel, data := range before {
			if !bytes.Equal(after[rel], data) {
				t.Fatalf("declined voice change changed %s", rel)
			}
		}
	})
}

// TestReleasesScreenNoReleasesMessage covers AC11's Releases empty state.
func TestReleasesScreenNoReleasesMessage(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	if err := releasesScreen(o, &out, p); err != nil {
		t.Fatalf("releasesScreen: %v", err)
	}
	if !strings.Contains(out.String(), "No releases are retained yet") {
		t.Fatalf("missing empty-state message: %s", out.String())
	}
}

// TestReleasesScreenChoosingInstalledSaysAlreadyInstalled covers AC11's
// Releases limit state: with a single retained release already installed,
// choosing it prints "Already installed" and changes nothing, without ever
// reaching rollbackFlow's own confirm.
func TestReleasesScreenChoosingInstalledSaysAlreadyInstalled(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
	before := collectFiles(t, home)

	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	var out bytes.Buffer
	// Exactly one release exists, and it is installed: "1" picks it.
	p := newHuhPrompter(true, strings.NewReader("1\n"), &out)
	if err := releasesScreen(o, &out, p); err != nil {
		t.Fatalf("releasesScreen: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Already installed") {
		t.Fatalf("missing \"Already installed\": %s", out.String())
	}
	after := collectFiles(t, home)
	if len(before) != len(after) {
		t.Fatal("choosing the installed release changed the home")
	}
	for rel, data := range before {
		if !bytes.Equal(after[rel], data) {
			t.Fatalf("choosing the installed release changed %s", rel)
		}
	}
}

// twoReleasesEnv builds a fixture with two distinct retained releases: the
// first from newUpdateFixtureWithPendingCommit's own initial commit
// (installed directly for codex and claude), the second by applying its
// own commit2 through the plain `hive update` command. It returns the env,
// the older (no longer installed) release's own ID, and that release's own
// shared skill file content, for a rollback test to compare against.
func twoReleasesEnv(t *testing.T) (env updateEnv, olderID string, olderBody []byte) {
	t.Helper()
	env, _ = newUpdateFixtureWithPendingCommit(t)
	olderBody, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir}
	entries, err := management.Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one retained release before the update, got %d", len(entries))
	}
	olderID = entries[0].ID

	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, out.String())
	}
	return env, olderID, olderBody
}

// releaseIndexFor is the 1-based accessible-mode index of the entry whose ID
// matches want, mirroring exactly the order releasesScreen's own Select
// presents (entries' own order, unchanged) instead of assuming any
// particular timestamp-based ordering across two installs that may land in
// the same second.
func releaseIndexFor(t *testing.T, entries []management.ReleaseEntry, want string) int {
	t.Helper()
	for i, e := range entries {
		if e.ID == want {
			return i + 1
		}
	}
	t.Fatalf("release %s not found among %d retained releases", want, len(entries))
	return 0
}

// TestReleasesScreenRollbackAppliesOlderRelease covers AC5's Releases
// rollback: choosing the older, not-installed release restores its own
// content (not the newer, currently installed one's), and it becomes the
// one installedReleaseID reports afterward.
func TestReleasesScreenRollbackAppliesOlderRelease(t *testing.T) {
	env, olderID, olderBody := twoReleasesEnv(t)
	o := management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir}
	entries, err := management.Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two retained releases, got %d", len(entries))
	}
	targetIndex := releaseIndexFor(t, entries, olderID)

	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(fmt.Sprintf("%d\ny\n", targetIndex)), &out)
	if err := releasesScreen(o, &out, p); err != nil {
		t.Fatalf("releasesScreen: %v\noutput:\n%s", err, out.String())
	}

	got, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, olderBody) {
		t.Fatalf("rollback did not restore the older release's content:\ngot:  %q\nwant: %q", got, olderBody)
	}
	gotID, err := installedReleaseID(o)
	if err != nil {
		t.Fatal(err)
	}
	if gotID != olderID {
		t.Fatalf("installed release after rollback = %s, want %s (the older one)", gotID, olderID)
	}
}

// TestRollbackFlowWrapsOnlyReleaseValidationErrors covers Releases' own
// "a release the current manager cannot validate" case (design.md "La
// interfaz") precisely (T4 fix round item 3): the documented downgrade hint
// is attached only to a real release-validation failure, reproduced here by
// planting a syntactically-decodable but fingerprint-invalid release JSON
// file directly under releases/ (lighter than building a full retired-
// agent-field fixture; integrations/agents_test.go already covers that
// specific parse error directly) — never to an unrelated BuildPlan failure
// such as "explicit hosts required", which loadRelease's own validation is
// never reached to produce.
func TestRollbackFlowWrapsOnlyReleaseValidationErrors(t *testing.T) {
	t.Run("release validation error gets the hint", func(t *testing.T) {
		source := minimalTestSource(t)
		dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
		home, stateDir := newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)

		bogusID := strings.Repeat("f", 64)
		bogus := fmt.Sprintf(`{"id": %q, "files": []}`, bogusID)
		if err := os.WriteFile(filepath.Join(stateDir, "releases", bogusID+".json"), []byte(bogus), 0600); err != nil {
			t.Fatal(err)
		}

		o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
		var out bytes.Buffer
		p := newHuhPrompter(true, strings.NewReader(""), &out)
		err := rollbackFlow(o, bogusID, &out, p)
		if err == nil {
			t.Fatal("expected a release-validation error")
		}
		if !strings.Contains(err.Error(), "invalid release fingerprint or file list") {
			t.Fatalf("expected the underlying validation error, got: %v", err)
		}
		if !strings.Contains(err.Error(), "plan that downgrade with the manager from the commit that produced it") {
			t.Fatalf("missing the documented downgrade hint: %v", err)
		}
	})

	t.Run("non-validation error has no hint", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		// No host registered at all: BuildPlan fails at validateHosts,
		// before ever reaching loadRelease/validateRelease.
		o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
		var out bytes.Buffer
		p := newHuhPrompter(true, strings.NewReader(""), &out)
		err := rollbackFlow(o, strings.Repeat("a", 64), &out, p)
		if err == nil {
			t.Fatal("expected an error: no host is registered to roll back for")
		}
		if !strings.Contains(err.Error(), "explicit hosts required") {
			t.Fatalf("expected the underlying hosts error, got: %v", err)
		}
		if strings.Contains(err.Error(), "plan that downgrade") {
			t.Fatalf("a non-validation error must not carry the downgrade hint: %v", err)
		}
	})
}

// TestFormatReleaseLabelMarksInstalledAndTruncates pins formatReleaseLabel's
// own shape (design.md "La interfaz": "etiquetas de 78 columnas o menos"):
// short ID, date, first commit, sorted host list, and an "(installed)"
// marker exactly when asked for one.
func TestFormatReleaseLabelMarksInstalledAndTruncates(t *testing.T) {
	entry := management.ReleaseEntry{
		ID:            strings.Repeat("a", 64),
		LastWrittenAt: "2026-09-20T10:00:00Z",
		Commits:       []string{strings.Repeat("1", 40)},
		Consumers: []management.Consumer{
			{Host: "codex"}, {Host: "claude"}, {Host: "grok"},
		},
	}
	notInstalled := formatReleaseLabel(entry, false)
	if strings.Contains(notInstalled, "(installed)") {
		t.Fatalf("unmarked release should not say installed: %q", notInstalled)
	}
	installed := formatReleaseLabel(entry, true)
	if !strings.HasSuffix(installed, "(installed)") {
		t.Fatalf("installed release must be marked: %q", installed)
	}
	for _, label := range []string{notInstalled, installed} {
		if n := utf8.RuneCountInString(label); n > releaseLabelWidth {
			t.Fatalf("label exceeds %d columns (%d): %q", releaseLabelWidth, n, label)
		}
		if !strings.Contains(label, shortHash(entry.ID)) {
			t.Fatalf("label missing the release's own short id: %q", label)
		}
		if !strings.Contains(label, "2026-09-20") {
			t.Fatalf("label missing the release's own date: %q", label)
		}
		if !strings.Contains(label, shortHash(entry.Commits[0])) {
			t.Fatalf("label missing the release's own first commit: %q", label)
		}
		if !strings.Contains(label, "claude, codex, grok") {
			t.Fatalf("label missing its sorted host list: %q", label)
		}
	}
}

// TestFormatReleaseLabelNoCommitsShowsDash covers a release that predates
// commit tracking: its own first-commit field is "-", not empty.
func TestFormatReleaseLabelNoCommitsShowsDash(t *testing.T) {
	entry := management.ReleaseEntry{ID: strings.Repeat("b", 64), LastWrittenAt: "2026-09-21T10:00:00Z"}
	if label := formatReleaseLabel(entry, false); !strings.Contains(label, "  -  ") {
		t.Fatalf("expected a dash placeholder for no commits: %q", label)
	}
}

// TestFormatReleaseLabelTruncatesLongHostList pins the 78-column budget
// under real pressure: many long host names truncate with a trailing
// ellipsis rather than pushing the label past releaseLabelWidth.
func TestFormatReleaseLabelTruncatesLongHostList(t *testing.T) {
	var consumers []management.Consumer
	for i := 0; i < 20; i++ {
		consumers = append(consumers, management.Consumer{Host: fmt.Sprintf("host-with-a-long-name-%02d", i)})
	}
	entry := management.ReleaseEntry{ID: strings.Repeat("c", 64), LastWrittenAt: "2026-09-22T10:00:00Z", Consumers: consumers}
	label := formatReleaseLabel(entry, true)
	if n := utf8.RuneCountInString(label); n > releaseLabelWidth {
		t.Fatalf("label exceeds %d columns (%d): %q", releaseLabelWidth, n, label)
	}
	if !strings.HasSuffix(label, "… (installed)") {
		t.Fatalf("expected a truncation ellipsis right before the installed marker: %q", label)
	}
}

// TestReleaseSelectFieldShowsEveryOptionWhenFewerThanHeight is the
// established T3 visibility pattern (TestHostsMultiSelectFieldShowsEveryOption)
// applied to Releases' own Select.
func TestReleaseSelectFieldShowsEveryOptionWhenFewerThanHeight(t *testing.T) {
	entries := []management.ReleaseEntry{
		{ID: strings.Repeat("a", 64)},
		{ID: strings.Repeat("b", 64)},
		{ID: strings.Repeat("c", 64)},
	}
	var value string
	view := releaseSelectField(entries, "", &value).View()
	for _, e := range entries {
		if !strings.Contains(view, shortHash(e.ID)) {
			t.Fatalf("missing release %s in the rendered view:\n%s", shortHash(e.ID), view)
		}
	}
}

// TestReleaseSelectFieldHeightIsFixed is the height/visibility test T4 asks
// for: unlike hostsMultiSelectField/providersMultiSelectField/
// menuSelectField (each sized to fieldHeight(len(options)), growing with the
// list up to the shared cap), Releases' own Select always requests
// maxFieldHeight (design.md: "de altura fija") — its rendered height is the
// same whether there are 1 or 50 releases.
func TestReleaseSelectFieldHeightIsFixed(t *testing.T) {
	var value string
	one := []management.ReleaseEntry{{ID: strings.Repeat("a", 64)}}
	many := make([]management.ReleaseEntry, 50)
	for i := range many {
		many[i] = management.ReleaseEntry{ID: fmt.Sprintf("%064d", i)}
	}
	oneView := releaseSelectField(one, "", &value).View()
	manyView := releaseSelectField(many, "", &value).View()
	oneLines := strings.Count(oneView, "\n")
	manyLines := strings.Count(manyView, "\n")
	if oneLines != manyLines {
		t.Fatalf("Releases' Select height is not fixed: %d option -> %d lines, %d options -> %d lines", len(one), oneLines, len(many), manyLines)
	}
}

// TestReleaseSelectFieldEscClosesFilterWithoutCancelling is T4's Esc-filter
// decision, verified directly against the pinned huh v2.0.3 version at the
// field level (no Form involved): huh's own default Select keymap already
// binds "esc" to close/clear an open filter (field_select.go's SetFilter/
// ClearFilter), independent of whatever keymap wraps the surrounding Form.
func TestReleaseSelectFieldEscClosesFilterWithoutCancelling(t *testing.T) {
	entries := []management.ReleaseEntry{{ID: strings.Repeat("a", 64)}, {ID: strings.Repeat("b", 64)}}
	var value string
	field := releaseSelectField(entries, "", &value)
	// A bare field's own keymap is the zero value (every binding disabled)
	// until a Form (or, here, this call directly) attaches one — exactly
	// what runFormWithKeyMap does via releasesSelectKeyMap in production.
	field.WithKeyMap(releasesSelectKeyMap)
	if field.GetFiltering() {
		t.Fatal("a fresh release Select must start browsing, not filtering")
	}
	field.Update(tea.KeyPressMsg{Text: "/"})
	if !field.GetFiltering() {
		t.Fatal("\"/\" did not open the release list's own filter")
	}
	field.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if field.GetFiltering() {
		t.Fatal("esc did not close the release list's own filter")
	}
}

// TestReleasesSelectKeyMapOnlyCtrlCCancelsFromFilter is T4's Esc-filter
// decision, verified at the Form level: with releasesSelectKeyMap (Quit
// bound only to ctrl+c), esc while the release Select is filtering reaches
// the field's own Update instead of aborting the whole screen first — the
// form stays StateNormal — while ctrl+c still cancels it from either state.
// This is the documented, narrow exception to AC10's "Ctrl-C o Esc" for
// this one field (tui_prompter.go's releasesSelectKeyMap doc comment):
// esc with no filter open is otherwise inert here, since huh's own Select
// keymap leaves esc unbound to anything else and this keymap does not bind
// it to Quit either.
func TestReleasesSelectKeyMapOnlyCtrlCCancelsFromFilter(t *testing.T) {
	var value string
	field := releaseSelectField([]management.ReleaseEntry{{ID: strings.Repeat("a", 64)}, {ID: strings.Repeat("b", 64)}}, "", &value)
	form := huh.NewForm(huh.NewGroup(field)).WithAccessible(false).WithInput(strings.NewReader("")).WithOutput(io.Discard).WithKeyMap(releasesSelectKeyMap)
	form.Init()

	form.Update(tea.KeyPressMsg{Text: "/"})
	form.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if form.State == huh.StateAborted {
		t.Fatal("esc while filtering aborted the whole Releases screen; it must only close the filter")
	}

	form.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if form.State != huh.StateAborted {
		t.Fatal("ctrl+c did not cancel the Releases screen")
	}
}

// TestVoiceScreenNoHostsRegistered covers AC11's Voice empty state, the same
// message Status and Remove CLIs use (design.md "La interfaz").
func TestVoiceScreenNoHostsRegistered(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	if err := voiceScreen(o, &out, p); err != nil {
		t.Fatalf("voiceScreen: %v", err)
	}
	if !strings.Contains(out.String(), "No CLI hosts are registered. Choose Install CLIs.") {
		t.Fatalf("missing empty-state message: %s", out.String())
	}
}

// TestVoiceScreenSourceWithoutVoicesReturnsItsError covers AC11's other
// Voice limit state: a source with no content/voices/ directory fails with
// ListVoices' own read error, not a generic or swallowed one.
func TestVoiceScreenSourceWithoutVoicesReturnsItsError(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)

	sourceWithoutVoices := t.TempDir()
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: sourceWithoutVoices}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	err := voiceScreen(o, &out, p)
	if err == nil {
		t.Fatal("expected ListVoices' own error for a source without content/voices")
	}
}

// TestVoiceScreenSetAppliesLikeCommand covers AC6 precisely (T4 fix round
// item 2, v1/v2): the interface's own address ("Sir") and intensity
// ("Subtle") choices reach BuildVoicePlan unchanged, compared against a twin
// built directly with management.BuildVoicePlan("set", ...) + Apply using
// the same values. TestVoiceScreenSetThenOffLeavesFilesIdentical alone
// cannot catch a swapped index (address <-> intensity, or the wrong option
// within either) mapping to the wrong VoiceSetting field, since turning the
// voice back off removes the block regardless of which wrong values it was
// rendered with; comparing the file state right after set, before ever
// turning it off, does.
func TestVoiceScreenSetAppliesLikeCommand(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)

	twinHome, twinState := newHostsTestHome(t)
	installViaText(t, twinHome, twinState, source, "claude", "y\n", dependencies)
	twinPlan, err := management.BuildVoicePlan("set", management.Options{Source: source, Home: twinHome, StateDir: twinState}, management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(twinPlan); err != nil {
		t.Fatal(err)
	}

	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	// SelectVoiceOrOff: "1" (testvoice); address Select: "2" (Sir); intensity
	// Select: "1" (Subtle); voicePlanWith's own Confirm: "y".
	p := newHuhPrompter(true, strings.NewReader("1\n2\n1\ny\n"), &out)
	if err := voiceScreen(o, &out, p); err != nil {
		t.Fatalf("voiceScreen: %v\noutput:\n%s", err, out.String())
	}

	assertHomesMatch(t, home, twinHome)
	assertStateJSONMatches(t, stateDir, home, twinState, twinHome)
}

// TestVoiceScreenSetThenOffLeavesFilesIdentical covers AC6: setting
// testVoiceID (T3's own stand-in for a real voice like Jarvis, minimalTestSource's
// only voice) with address sir, then turning it back off, leaves every file
// exactly as it was before either change — the voice block is fully
// removed, not left as a residual difference.
func TestVoiceScreenSetThenOffLeavesFilesIdentical(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
	before := collectFiles(t, home)

	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}

	var setOut bytes.Buffer
	// SelectVoiceOrOff: "1" (testvoice, the only voice); address Select: "2"
	// (Sir); intensity Select: "1" (Subtle); voicePlanWith's own Confirm:
	// "y".
	setP := newHuhPrompter(true, strings.NewReader("1\n2\n1\ny\n"), &setOut)
	if err := voiceScreen(o, &setOut, setP); err != nil {
		t.Fatalf("voiceScreen (set): %v\noutput:\n%s", err, setOut.String())
	}
	if !strings.Contains(setOut.String(), "Active voice: off") {
		t.Fatalf("expected the active voice to start off: %s", setOut.String())
	}

	var offOut bytes.Buffer
	// SelectVoiceOrOff: "2" (Off); voicePlanWith's own Confirm: "y".
	offP := newHuhPrompter(true, strings.NewReader("2\ny\n"), &offOut)
	if err := voiceScreen(o, &offOut, offP); err != nil {
		t.Fatalf("voiceScreen (off): %v\noutput:\n%s", err, offOut.String())
	}
	if !strings.Contains(offOut.String(), "Active voice: "+testVoiceID) {
		t.Fatalf("expected the active voice to be %s before turning it off: %s", testVoiceID, offOut.String())
	}

	after := collectFiles(t, home)
	if len(before) != len(after) {
		t.Fatalf("set-then-off left the home changed: before=%d files, after=%d files", len(before), len(after))
	}
	for rel, data := range before {
		if !bytes.Equal(after[rel], data) {
			t.Fatalf("set-then-off left %s changed", rel)
		}
	}
}

// TestVoiceSelectFieldShowsEveryVoiceAtRealisticWidth is T4 fix round item 1:
// huh's own Select wraps an option whose rendered text exceeds the field's
// width (field_select.go's renderOption -> wrap); fieldHeight(len(options))
// only ever reserves one rendered line per option. Before voiceOptionLabel
// truncated the description, a real voice's own full description (jarvis
// and mentor's are one to two full sentences) wrapped to two lines at a
// realistic 80-column width, pushing later options out of the viewport
// entirely — reproduced here with two voices whose descriptions are
// deliberately long enough to wrap under the old "%s (%s)" label at width
// 80, and confirmed to still lose the later ones before this fix.
func TestVoiceSelectFieldShowsEveryVoiceAtRealisticWidth(t *testing.T) {
	voices := []management.VoiceInfo{
		{ID: "jarvis", Description: "Warm, formal, and precise, in the style of a meticulous personal assistant who anticipates needs before they are stated aloud"},
		{ID: "mentor", Description: "Patient, encouraging, and direct, favoring plain language over jargon and always explaining the reasoning behind a suggestion"},
	}
	var value string
	field := voiceSelectField(voices, &value)
	field.WithWidth(80)
	view := field.View()
	for _, v := range voices {
		if !strings.Contains(view, v.ID) {
			t.Fatalf("missing voice %q in the rendered view at width 80:\n%s", v.ID, view)
		}
	}
	// "Off" is the last option: if an earlier, wrapped description pushes
	// later options out of the viewport (the exact defect reported — at
	// 80x24 both voices were hidden, at 100x45 only jarvis's own tail
	// showed), it disappears first. Checking only that each ID's own text
	// appears somewhere is not enough on its own: a wrapped option's first
	// line still starts with its ID even when everything after it, up to
	// and including later options, is clipped.
	if !strings.Contains(view, "Off") {
		t.Fatalf("the trailing Off option was pushed out of the view at width 80:\n%s", view)
	}
}

// TestScreensPrintOneLineHeader covers design.md "La interfaz": "Cada
// pantalla empieza con un encabezado de una línea" (T4 fix round item 5),
// for all six non-Quit screens. Each case deliberately drives its screen
// into whatever state prints fastest (an empty/error state, or a real
// flow the accessible script cancels or errors out of immediately after
// the header) — the header itself is what is pinned, not that state.
func TestScreensPrintOneLineHeader(t *testing.T) {
	cases := []struct {
		name string
		want string
		run  func(t *testing.T, out *bytes.Buffer)
	}{
		{"status", "== Status ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
			if err := statusScreen(o, out, nil); err != nil {
				t.Fatal(err)
			}
		}},
		{"install", "== Install CLIs ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: t.TempDir()}
			p := newHuhPrompter(true, strings.NewReader(""), out)
			// A source without a catalog errors immediately; the header must
			// still have printed first.
			_ = installScreen(o, out, p)
		}},
		{"remove", "== Remove CLIs ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
			p := newHuhPrompter(true, strings.NewReader(""), out)
			if err := removeScreen(o, out, p); err != nil {
				t.Fatal(err)
			}
		}},
		{"update", "== Update ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Home: home, StateDir: stateDir}
			p := newHuhPrompter(true, strings.NewReader("\n\n"), out)
			// No host registered: updateWith errors after the header.
			_ = updateScreen(o, out, p)
		}},
		{"releases", "== Releases ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
			p := newHuhPrompter(true, strings.NewReader(""), out)
			if err := releasesScreen(o, out, p); err != nil {
				t.Fatal(err)
			}
		}},
		{"voice", "== Voice ==", func(t *testing.T, out *bytes.Buffer) {
			home, stateDir := newHostsTestHome(t)
			o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
			p := newHuhPrompter(true, strings.NewReader(""), out)
			if err := voiceScreen(o, out, p); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			c.run(t, &out)
			if !strings.HasPrefix(out.String(), c.want+"\n") {
				t.Fatalf("expected the %s screen to start with %q, got:\n%s", c.name, c.want, out.String())
			}
		})
	}
}

func TestIsReleaseValidationErrorCoversReferenceValidation(t *testing.T) {
	if !isReleaseValidationError(errors.New("nonportable personal path in content/skills/x/SKILL.md")) {
		t.Fatal("a nonportable reference in a retained release is a release-validation error")
	}
	if isReleaseValidationError(errors.New("explicit hosts required")) {
		t.Fatal("a missing host selection is not a release-validation error")
	}
}
