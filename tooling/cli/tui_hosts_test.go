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
// with the view under test, so both sides differ only in how they are driven.
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
