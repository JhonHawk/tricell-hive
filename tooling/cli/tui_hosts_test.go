package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// testVoiceID names the voice content/voices/testvoice.md carries in
// minimalCatalogSource, distinct from any real voice (jarvis, ...) so a
// test never depends on that content's own wording.
const testVoiceID = "testvoice"

// minimalCatalogSource builds, once per test binary (sync.OnceValues), a
// synthetic Hive source tree with just enough content to exercise
// Install/Remove's own equivalence: content/guidance/global.md (every
// install plan writes it unconditionally — management.GlobalSource) and one
// voice (content/voices/preamble.md and testvoice.md), which the
// active-voice removal test needs. Real Install/Remove behavior does not
// depend on which skills or agents a source happens to carry, so this
// file's twin-comparison tests use this instead of the real content/ tree
// (T3 fix round item 5): a full real-tree install touches ~80 files and
// dominated this package's own -race run time (549s -> 857s, over go
// test's own default 10-minute timeout).
var minimalCatalogSource = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "hive-tui-test-catalog-")
	if err != nil {
		return "", err
	}
	// A macOS temp dir is itself commonly a symlink (/tmp -> /private/tmp);
	// BuildPlan's own target.Safe rejects a symlinked source ancestor.
	dir, err = filepath.EvalSymlinks(dir)
	if err != nil {
		return "", err
	}
	guidanceDir := filepath.Join(dir, "content", "guidance")
	if err := os.MkdirAll(guidanceDir, 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(guidanceDir, "global.md"), []byte("# Global\n\nMinimal test guidance.\n"), 0600); err != nil {
		return "", err
	}
	voicesDir := filepath.Join(dir, "content", "voices")
	if err := os.MkdirAll(voicesDir, 0700); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(voicesDir, "preamble.md"), []byte("Voice preamble.\n"), 0600); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(voicesDir, testVoiceID+".md"), []byte("Test voice: a minimal fixture voice.\n"), 0600); err != nil {
		return "", err
	}
	return dir, nil
})

// minimalTestSource returns minimalCatalogSource's own directory, failing
// the test if it could not be built.
func minimalTestSource(t *testing.T) string {
	t.Helper()
	dir, err := minimalCatalogSource()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestMain removes minimalCatalogSource's own directory once every test in
// this binary has run: it is built with os.MkdirTemp (not t.TempDir()) so
// it survives across tests, but nothing needs it once the binary exits.
func TestMain(m *testing.M) {
	code := m.Run()
	if dir, err := minimalCatalogSource(); err == nil {
		os.RemoveAll(dir)
	}
	os.Exit(code)
}

// newHostsTestHome builds a fresh synthetic home and state directory, never
// touching the real user's own state. stateDir is a sibling of home, not
// nested under it, so assertHomesMatch (which walks home) never has to
// reason about the state directory's own journal/transaction bookkeeping;
// assertStateJSONMatches covers state.json on its own terms.
func newHostsTestHome(t *testing.T) (home, stateDir string) {
	t.Helper()
	parent := t.TempDir()
	home = filepath.Join(parent, "home")
	stateDir = filepath.Join(parent, "state")
	if err := os.MkdirAll(home, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	return home, stateDir
}

// installViaText builds a home through the plain text installer with an
// explicit --hosts (install.go's own installWithDependencies) — the twin
// AC2 asks the interface's own result to match. dependencies lets a test
// share the exact same fake (never real) provider/RequiredHosts behavior
// with the huh-driven side under test, so both sides differ only in which
// prompter answers the same questions.
func installViaText(t *testing.T, home, stateDir, source, hosts, answers string, dependencies installDependencies) {
	t.Helper()
	var out bytes.Buffer
	args := []string{"--home", home, "--state-dir", stateDir, "--source", source, "--hosts", hosts}
	if err := installWithDependencies(args, strings.NewReader(answers), &out, true, dependencies); err != nil {
		t.Fatalf("twin text install: %v\noutput:\n%s", err, out.String())
	}
}

// collectFiles reads every regular file under root into a map keyed by its
// root-relative path, for comparing two homes byte for byte. A symlink
// (Hive installs skills as symlinks into a shared skills directory) is
// recorded as its own target, prefixed to distinguish it from an identical-
// looking regular file, since os.ReadFile on a symlink-to-directory fails
// with "is a directory".
func collectFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			// Normalize an absolute in-root target (Hive symlinks skills
			// into a shared directory under the same home) to be
			// root-relative, so two different homes' own absolute prefixes
			// never make an otherwise-identical symlink look different.
			if relTarget, err := filepath.Rel(root, target); err == nil && !strings.HasPrefix(relTarget, "..") {
				target = relTarget
			}
			files[rel] = []byte("symlink:" + target)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[rel] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// assertHomesMatch fails with every path difference between two homes
// instead of stopping at the first one, so a real mismatch is never hidden
// behind an incidental first difference (AC2/AC3: "el resultado en archivos
// ... es el mismo").
func assertHomesMatch(t *testing.T, gotRoot, wantRoot string) {
	t.Helper()
	got := collectFiles(t, gotRoot)
	want := collectFiles(t, wantRoot)
	var mismatches []string
	for rel, wantData := range want {
		gotData, ok := got[rel]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("missing in %s: %s", gotRoot, rel))
			continue
		}
		if !bytes.Equal(gotData, wantData) {
			mismatches = append(mismatches, fmt.Sprintf("content differs: %s", rel))
		}
		delete(got, rel)
	}
	for rel := range got {
		mismatches = append(mismatches, fmt.Sprintf("unexpected in %s: %s", gotRoot, rel))
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		t.Fatalf("home mismatch (%s vs %s):\n%s", gotRoot, wantRoot, strings.Join(mismatches, "\n"))
	}
}

// assertStateJSONMatches compares state.json (AC2/AC3: "... y el mismo
// estado") between two independent synthetic homes, rather than byte for
// byte: state.json embeds each home's own absolute path throughout
// (Config.Home and every Records key), which necessarily differs between
// two different t.TempDir()s even for an otherwise identical result, so
// each home's own absolute prefix is normalized to a shared placeholder
// first (which also normalizes every path nested inside a migrations[]
// entry's own "config", the same way). Only migrations[*].transaction is
// dropped (T3 fix round item 3: the previous version dropped the whole
// migrations array): BuildPlan/Apply's own legacy-detection scan
// (management.scanMigration) always logs a receipt — with a fresh random
// transaction ID (apply.go's own crypto/rand.Read) — even when it finds
// nothing to migrate, so only that one field never matches between two
// separate runs regardless of correctness; hosts/result/detector/release
// stay comparable.
func assertStateJSONMatches(t *testing.T, gotStateDir, gotHome, wantStateDir, wantHome string) {
	t.Helper()
	got := normalizedStateJSON(t, gotStateDir, gotHome)
	want := normalizedStateJSON(t, wantStateDir, wantHome)
	if got != want {
		t.Fatalf("state.json differs (home paths normalized, migrations[*].transaction excluded):\ngot:  %s\nwant: %s", got, want)
	}
}

func normalizedStateJSON(t *testing.T, stateDir, home string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("parsing %s: %v", filepath.Join(stateDir, "state.json"), err)
	}
	if migrations, ok := raw["migrations"].([]any); ok {
		for _, m := range migrations {
			if entry, ok := m.(map[string]any); ok {
				delete(entry, "transaction")
			}
		}
	}
	normalized, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(normalized), home, "HOME")
}

// TestInstallScreenInstallsCodexAndClaude covers AC2's plain case: choosing
// Codex and Claude (installerHosts' own order is claude, codex, cursor,
// grok, opencode, pi, so "1" and "2" toggle them) produces the same files
// and state as `hive install --hosts claude,codex`.
func TestInstallScreenInstallsCodexAndClaude(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)

	twinHome, twinState := newHostsTestHome(t)
	installViaText(t, twinHome, twinState, source, "claude,codex", "y\n", dependencies)

	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	// MultiSelect: toggle claude(1), codex(2), confirm(0); no shared-host
	// expansion needed; the final 3-option Select (Apply/Back/Cancel, since
	// hosts were chosen interactively) picks Apply(1).
	p := newHuhPrompter(true, strings.NewReader("1\n2\n0\n1\n"), &out)
	if err := runInstallFlowWith(o, false, &out, p, dependencies, true); err != nil {
		t.Fatalf("runInstallFlowWith: %v\noutput:\n%s", err, out.String())
	}

	assertHomesMatch(t, home, twinHome)
	assertStateJSONMatches(t, stateDir, home, twinState, twinHome)
}

// TestInstallScreenExpandsToRequiredHost covers AC2's shared-resource
// expansion case: a required-hosts dependency (the same injection seam
// install_test.go's own TestInstallSharedHostClosureRequiresConsent uses,
// since no pair of real hosts in this checkout happens to require another
// on a fresh install) forces selecting an additional host, consented to
// through its own plain confirm, before the final apply.
func TestInstallScreenExpandsToRequiredHost(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.RequiredHosts = func(management.Options) ([]string, error) {
		return []string{"claude", "grok"}, nil
	}

	twinHome, twinState := newHostsTestHome(t)
	// --hosts grok alone; expandToRequiredHosts's own plain confirm ("y"),
	// then the final plain confirm (explicit --hosts means allowBack=false
	// there too).
	installViaText(t, twinHome, twinState, source, "grok", "y\ny\n", dependencies)

	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	// MultiSelect: toggle grok(4), confirm(0); expandToRequiredHosts's own
	// plain Confirm ("y"); final 3-option Select picks Apply(1).
	p := newHuhPrompter(true, strings.NewReader("4\n0\ny\n1\n"), &out)
	if err := runInstallFlowWith(o, false, &out, p, dependencies, true); err != nil {
		t.Fatalf("runInstallFlowWith: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Shared resources require selecting: claude") {
		t.Fatalf("missing the expansion notice: %s", out.String())
	}

	assertHomesMatch(t, home, twinHome)
	assertStateJSONMatches(t, stateDir, home, twinState, twinHome)
}

// TestDeclineChangesNothing covers AC10 for both Install CLIs and Remove
// CLIs in one test (T3 fix round item 5: the two were separate tests):
// rejecting the final confirmation leaves the home untouched and prints the
// same cancellation text every screen uses.
func TestDeclineChangesNothing(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)

	t.Run("install", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
		var out bytes.Buffer
		// MultiSelect: toggle claude(1), confirm(0); final 3-option Select
		// picks Cancel(3).
		p := newHuhPrompter(true, strings.NewReader("1\n0\n3\n"), &out)
		if err := runInstallFlowWith(o, false, &out, p, dependencies, true); err != nil {
			t.Fatalf("runInstallFlowWith: %v\noutput:\n%s", err, out.String())
		}
		if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
			t.Fatalf("missing cancellation message: %s", out.String())
		}
		entries, err := os.ReadDir(home)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("declined install wrote to home: %v", entries)
		}
	})

	t.Run("remove", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
		before := collectFiles(t, home)

		o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
		var out bytes.Buffer
		p := newHuhPrompter(true, strings.NewReader("1\n0\nn\n"), &out)
		if err := removeScreen(o, &out, p); err != nil {
			t.Fatalf("removeScreen: %v\noutput:\n%s", err, out.String())
		}
		if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
			t.Fatalf("missing cancellation message: %s", out.String())
		}
		after := collectFiles(t, home)
		if len(before) != len(after) {
			t.Fatalf("declined remove changed the home: before=%d files, after=%d files", len(before), len(after))
		}
		for rel, data := range before {
			if !bytes.Equal(after[rel], data) {
				t.Fatalf("declined remove changed %s", rel)
			}
		}
	})
}

// TestInstallScreenSourceWithoutCatalog covers AC11's Install CLIs limit
// state (design.md "Punto de entrada"): a --source with no
// content/guidance/global.md fails with the documented message instead of
// running the wizard at all.
func TestInstallScreenSourceWithoutCatalog(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: t.TempDir()}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	err := installScreen(o, &out, p)
	if err == nil {
		t.Fatal("expected an error for a source without a catalog")
	}
	want := "Run hive from a Hive checkout or package, or pass --source"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

// TestInstallScreenSourceWithoutCatalogReturnsToMenu is the end-to-end
// counterpart, through the real menu dispatch: the error prints and the
// menu still reaches Quit cleanly (AC10's "un error del flujo imprime el
// mensaje del comando y vuelve al menú").
func TestInstallScreenSourceWithoutCatalogReturnsToMenu(t *testing.T) {
	o := interfaceTestOptions(t)
	o.Source = t.TempDir()
	var out bytes.Buffer
	// "2" is Install CLIs; the catalog check fails before ever building a
	// prompter form, so no further scripted input is consumed by it; "7"
	// then quits.
	if err := openInterface(false, true, strings.NewReader("2\n7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Run hive from a Hive checkout or package, or pass --source") {
		t.Fatalf("missing catalog error: %s", out.String())
	}
}

// TestRemoveScreenRemovesClaudeWithActiveVoice covers AC3's voice-removal
// case: removing the only host consuming an active voice block removes
// that block too, matching a twin built by BuildPlan("remove") + Apply
// directly (the equivalent of `plan remove` followed by `apply`).
func TestRemoveScreenRemovesClaudeWithActiveVoice(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)

	setUpClaudeWithVoice := func(t *testing.T) (home, stateDir string) {
		t.Helper()
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", dependencies)
		vo := management.Options{Source: source, Home: home, StateDir: stateDir}
		vp, err := management.BuildVoicePlan("set", vo, management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := (management.Engine{}).Apply(vp); err != nil {
			t.Fatal(err)
		}
		return home, stateDir
	}

	twinHome, twinState := setUpClaudeWithVoice(t)
	ro := management.Options{Scope: "user", Home: twinHome, StateDir: twinState, Hosts: []string{"claude"}}
	twinPlan, err := management.BuildPlan("remove", ro)
	if err != nil {
		t.Fatal(err)
	}
	if len(twinPlan.Voice) == 0 {
		t.Fatal("test setup did not exercise the active-voice removal path")
	}
	if _, err := (management.Engine{}).Apply(twinPlan); err != nil {
		t.Fatal(err)
	}

	home, stateDir := setUpClaudeWithVoice(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	// MultiSelect (one candidate, claude): toggle it (1), confirm (0);
	// removeFlow's own plain Confirm ("y").
	p := newHuhPrompter(true, strings.NewReader("1\n0\ny\n"), &out)
	if err := removeScreen(o, &out, p); err != nil {
		t.Fatalf("removeScreen: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Voice blocks to remove: 1") {
		t.Fatalf("missing voice-block removal in the summary: %s", out.String())
	}

	assertHomesMatch(t, home, twinHome)
	assertStateJSONMatches(t, stateDir, home, twinState, twinHome)
}

// TestRemoveScreenNoHostsRegistered covers AC11's Remove CLIs limit state,
// the same message Status uses (design.md "La interfaz").
func TestRemoveScreenNoHostsRegistered(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	if err := removeScreen(o, &out, p); err != nil {
		t.Fatalf("removeScreen: %v", err)
	}
	want := "No CLI hosts are registered. Choose Install CLIs."
	if !strings.Contains(out.String(), want) {
		t.Fatalf("missing empty-state message: %s", out.String())
	}
}

// TestRemoveScreenNoHostsReturnsToMenu is the end-to-end counterpart,
// through the real menu dispatch.
func TestRemoveScreenNoHostsReturnsToMenu(t *testing.T) {
	o := interfaceTestOptions(t)
	var out bytes.Buffer
	// "3" is Remove CLIs; "7" then quits.
	if err := openInterface(false, true, strings.NewReader("3\n7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "No CLI hosts are registered. Choose Install CLIs.") {
		t.Fatalf("missing empty-state message: %s", out.String())
	}
}

// TestHostCandidateLabelMatchesStatus is T3 fix round item 2 (mutation m2):
// pins hostCandidateLabel's exact per-candidate status text and its
// priority order (Detected > Registered > Legacy > "not detected"), so a
// mutation flipping or dropping a branch there is caught.
func TestHostCandidateLabelMatchesStatus(t *testing.T) {
	cases := []struct {
		name string
		c    hostCandidate
		want string
	}{
		{"not detected", hostCandidate{Name: "codex"}, "codex (not detected)"},
		{"detected", hostCandidate{Name: "codex", Detected: true}, "codex (executable detected)"},
		{"registered", hostCandidate{Name: "codex", Registered: true}, "codex (registered by Hive)"},
		{"legacy", hostCandidate{Name: "codex", Legacy: true}, "codex (legacy installation detected)"},
		{"detected beats registered", hostCandidate{Name: "codex", Detected: true, Registered: true}, "codex (executable detected)"},
		{"detected beats legacy", hostCandidate{Name: "codex", Detected: true, Legacy: true}, "codex (executable detected)"},
		{"registered beats legacy", hostCandidate{Name: "codex", Registered: true, Legacy: true}, "codex (registered by Hive)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hostCandidateLabel(c.c); got != c.want {
				t.Fatalf("hostCandidateLabel(%+v) = %q, want %q", c.c, got, c.want)
			}
		})
	}
}

// TestShowRemoveSummaryListsFilesToRemoveAndKept is T3 fix round item 2
// (mutation m5): pins that showRemoveSummary's own "Files to remove" and
// "Shared resources kept" sections list the *right* paths, not merely the
// right counts — a mutation swapping the removed/kept branches, or one that
// drops the loop body, would still satisfy a count-only assertion.
func TestShowRemoveSummaryListsFilesToRemoveAndKept(t *testing.T) {
	home := "/synthetic-home"
	cfg := target.Config{
		Home:       home,
		ClaudeHome: filepath.Join(home, ".claude"),
		CodexHome:  filepath.Join(home, ".codex"),
	}
	plan := management.Plan{
		Hosts:    []string{"claude"},
		StateDir: "/synthetic-state",
		Config:   cfg,
		Changes: []management.Change{
			{
				Target: target.Target{Path: filepath.Join(home, ".claude", "CLAUDE.md")},
				Before: &management.Record{},
				After:  nil, // the only consumer is gone: the file is removed
			},
			{
				Target: target.Target{Path: filepath.Join(home, ".agents", "skills", "shared", "SKILL.md")},
				Before: &management.Record{},
				After:  &management.Record{}, // another consumer remains: kept
			},
		},
	}
	var out bytes.Buffer
	showRemoveSummary(&out, plan)
	got := out.String()

	wantRemoved := fmt.Sprintf("Files to remove: 1\n  %s\n", filepath.Join(home, ".claude"))
	if !strings.Contains(got, wantRemoved) {
		t.Fatalf("missing the removed-file listing: want it to contain %q, got:\n%s", wantRemoved, got)
	}
	wantKept := fmt.Sprintf("Shared resources kept for other consumers: 1\n  %s\n", filepath.Join(home, ".agents"))
	if !strings.Contains(got, wantKept) {
		t.Fatalf("missing the kept-resource listing: want it to contain %q, got:\n%s", wantKept, got)
	}
}

// TestInstallScreenOmitsDryRunHint is T3 fix round item 4's first half: the
// interface has no --dry-run flag of its own, so showInstallSummary's own
// closing hint about it does not apply inside the TUI path.
func TestInstallScreenOmitsDryRunHint(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("1\n0\n1\n"), &out)
	if err := runInstallFlowWith(o, false, &out, p, dependencies, true); err != nil {
		t.Fatalf("runInstallFlowWith: %v\noutput:\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "Use --dry-run") {
		t.Fatalf("the interface must not mention --dry-run, a flag it has no way to pass: %s", out.String())
	}
}

// TestInstallCommandKeepsDryRunHint is TestInstallScreenOmitsDryRunHint's
// characterization counterpart: hive install itself has a real --dry-run
// flag, so its own output must keep mentioning it, unchanged.
func TestInstallCommandKeepsDryRunHint(t *testing.T) {
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	var out bytes.Buffer
	args := []string{"--home", home, "--state-dir", stateDir, "--source", source, "--hosts", "claude"}
	if err := installWithDependencies(args, strings.NewReader("y\n"), &out, true, dependencies); err != nil {
		t.Fatalf("installWithDependencies: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Use --dry-run to see the full file list before applying.") {
		t.Fatalf("hive install lost its --dry-run hint: %s", out.String())
	}
}

// TestRunInstallFlowWithPendingOnboardingRecoveryNamesHiveRecoverForInterface
// is T3 fix round item 4's second half: runInstallFlowWith's own
// pending-operation check (handlePendingInstallOperation, reached before the
// wizard even starts) must also name hive recover when fromInterface is
// true, not only tui.go's separate checkPendingOnOpen (used only at
// menu-open time). Every synthetic-home fixture in this file sets an
// explicit --state-dir (newHostsTestHome never uses the real default), so
// this end-to-end path always exercises the --state-dir-naming branch too
// (T4 leftover from T3 verification); recoveryPhraseFor's own unit test
// (TestRecoveryPhraseForInterfaceAlwaysNamesHiveRecover) covers the
// non-explicit branch directly, since nothing here may touch real user
// state to exercise it end to end.
func TestRunInstallFlowWithPendingOnboardingRecoveryNamesHiveRecoverForInterface(t *testing.T) {
	source := minimalTestSource(t)
	home, stateDir := newHostsTestHome(t)
	// NormalizeOptions resolves symlinks in an explicit --state-dir (e.g.
	// macOS's /var -> /private/var); resolve it here too so the expected
	// recovery text compares against the same, fully resolved path.
	resolvedStateDir, err := filepath.EvalSymlinks(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
	dependencies.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "partial"}, nil
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("y\n"), &out)
	if err := runInstallFlowWith(o, false, &out, p, dependencies, true); err != nil {
		t.Fatalf("runInstallFlowWith: %v\noutput:\n%s", err, out.String())
	}
	want := fmt.Sprintf("Run hive recover --state-dir %s.", resolvedStateDir)
	if !strings.Contains(out.String(), want) {
		t.Fatalf("interface-originated pending-onboarding recovery must name %q, got: %s", want, out.String())
	}
	if strings.Contains(out.String(), "install.sh") {
		t.Fatalf("interface-originated recovery must not name install.sh: %s", out.String())
	}
}

// TestRecoveryPhraseForInterfaceAlwaysNamesHiveRecover pins
// recoveryPhraseFor's own contract directly: fromInterface always wins,
// regardless of online, and names --state-dir too when explicitStateDir is
// true (T4 leftover from T3 verification).
func TestRecoveryPhraseForInterfaceAlwaysNamesHiveRecover(t *testing.T) {
	for _, online := range []bool{false, true} {
		got := recoveryPhraseFor(true, false, online, "/synthetic-state")
		if got != "run hive recover" {
			t.Fatalf("recoveryPhraseFor(fromInterface=true, explicitStateDir=false, online=%v, ...) = %q, want %q", online, got, "run hive recover")
		}
		got = recoveryPhraseFor(true, true, online, "/synthetic-state")
		want := "run hive recover --state-dir /synthetic-state"
		if got != want {
			t.Fatalf("recoveryPhraseFor(fromInterface=true, explicitStateDir=true, online=%v, ...) = %q, want %q", online, got, want)
		}
	}
}

// TestRecoveryPhraseForNonInterfaceUnchanged pins that fromInterface=false
// defers to recoveryPhrase exactly, for install's and bootstrap's own
// callers, regardless of explicitStateDir.
func TestRecoveryPhraseForNonInterfaceUnchanged(t *testing.T) {
	stateDir := t.TempDir()
	for _, online := range []bool{false, true} {
		for _, explicitStateDir := range []bool{false, true} {
			got := recoveryPhraseFor(false, explicitStateDir, online, stateDir)
			want := recoveryPhrase(online, stateDir)
			if got != want {
				t.Fatalf("recoveryPhraseFor(fromInterface=false, explicitStateDir=%v, online=%v, ...) = %q, want %q (recoveryPhrase's own result)", explicitStateDir, online, got, want)
			}
		}
	}
}
