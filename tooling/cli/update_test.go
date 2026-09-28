package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/integrations/agents"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

// requireGit skips the test when Git is not available, since these tests
// build real Git repositories and run real `git archive` output through the
// adapter and distribution.Extract.
func requireGit(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not found in PATH")
	}
	return path
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// gitInitLocal initializes dir as a Git repository with a commit identity
// set locally in that repository only (never global), per the worktree
// brief.
func gitInitLocal(t *testing.T, dir string) {
	t.Helper()
	gitRun(t, dir, "init", "-q")
	gitRun(t, dir, "config", "user.name", "hive-update-test")
	gitRun(t, dir, "config", "user.email", "hive-update-test@example.invalid")
}

func gitCommitAll(t *testing.T, dir, message string) string {
	t.Helper()
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-q", "-m", message)
	return strings.TrimSpace(gitRun(t, dir, "rev-parse", "HEAD"))
}

func skillBody(text string) string {
	return "---\nname: workspace-conventions\ndescription: Organize records.\n---\n" + text
}

// writeUpdateCatalog writes the minimal catalog tasks.md's T3 test setup
// calls for: the global instruction, a skill, an agent role and the real
// agent profiles this repository ships (management's own helpers, mirrored
// here per tasks.md, are unexported and not reusable across packages).
func writeUpdateCatalog(t *testing.T, dir, skill string) {
	t.Helper()
	putCharacterization(t, filepath.Join(dir, management.GlobalSource), "# Rules\nKeep user content.\n")
	putCharacterization(t, filepath.Join(dir, management.SkillSource), skill)
	profiles, err := os.ReadFile(filepath.Join("..", "..", agents.ProfilesSource))
	if err != nil {
		t.Fatal(err)
	}
	putCharacterization(t, filepath.Join(dir, agents.ProfilesSource), string(profiles))
	putCharacterization(t, filepath.Join(dir, "content/agents/design/test-agent.md"), "---\nname: test-agent\ndescription: Test role\nmodel_profile: execution\naccess_profile: observe\n---\nUse evidence.\n")
}

type updateEnv struct {
	home, stateDir, repo string
}

func newUpdateEnv(t *testing.T) updateEnv {
	t.Helper()
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := updateEnv{home: filepath.Join(base, "home"), stateDir: filepath.Join(base, "state"), repo: filepath.Join(base, "repo")}
	if err := os.MkdirAll(e.home, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(e.repo, 0700); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e updateEnv) sharedSkillPath() string {
	return filepath.Join(e.home, ".agents", "skills", "workspace-conventions", "SKILL.md")
}

func (e updateEnv) args(extra ...string) []string {
	return append([]string{"--home", e.home, "--state-dir", e.stateDir, "--source", e.repo}, extra...)
}

func installDirect(t *testing.T, home, stateDir, source string, hosts []string) management.Plan {
	t.Helper()
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: hosts}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
	return p
}

// installFromCommit installs commit's content directly (not through update,
// so no host needs to be registered yet), but through the same git
// archive/extract pipeline update.go uses, so file modes come out
// Git-normalized (644/755) exactly as a later `hive update` would produce
// them. Installing straight from putCharacterization's fixture files
// instead (mode 0640) would make an otherwise content-identical later
// commit look changed on file mode alone, which is a fixture artifact, not
// a real difference between two commits with identical catalogue content.
func installFromCommit(t *testing.T, home, stateDir, repo, commit string, hosts []string) {
	t.Helper()
	gitPath := requireGit(t)
	raw, err := archiveGitCommit(gitPath, filteredGitEnv(), repo, commit, "hive-baseline")
	if err != nil {
		t.Fatal(err)
	}
	gz, err := adaptGitArchive(raw)
	if err != nil {
		t.Fatal(err)
	}
	dest, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root, err := distribution.Extract(gz, dest)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(root))
	installDirect(t, home, stateDir, root, hosts)
}

func stateBytes(t *testing.T, stateDir string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// newUpdateFixtureWithPendingCommit builds the shared setup tasks.md
// describes for AC1-AC3: a Git repository whose first commit is already
// installed for two hosts, a second commit that changes the skill (the
// commit `hive update` must resolve to by default, since --rev defaults to
// HEAD), and an uncommitted change on top of it that `update` must never
// pick up.
func newUpdateFixtureWithPendingCommit(t *testing.T) (env updateEnv, commit2 string) {
	t.Helper()
	requireGit(t)
	env = newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, env.repo)
	gitCommitAll(t, env.repo, "init")
	installDirect(t, env.home, env.stateDir, env.repo, []string{"codex", "claude"})

	putCharacterization(t, filepath.Join(env.repo, management.SkillSource), skillBody("Preserve evidence, updated.\n"))
	commit2 = gitCommitAll(t, env.repo, "update skill")

	// Left uncommitted on purpose: update must resolve HEAD (commit2), not
	// this dirty working tree content.
	putCharacterization(t, filepath.Join(env.repo, management.SkillSource), skillBody("Preserve evidence, DIRTY UNCOMMITTED.\n"))

	return env, commit2
}

// TestUpdateInteractiveResolvesCommitAndApplies covers AC1: in an
// interactive terminal, `hive update` resolves the requested revision,
// shows the summary for the registered hosts, applies on "y", and installs
// the resolved commit's content rather than the dirty working tree.
func TestUpdateInteractiveResolvesCommitAndApplies(t *testing.T) {
	env, commit2 := newUpdateFixtureWithPendingCommit(t)
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("update: %v\noutput:\n%s", err, out.String())
	}
	body, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	if want := skillBody("Preserve evidence, updated.\n"); string(body) != want {
		t.Fatalf("installed content mismatch:\ngot:  %q\nwant: %q", body, want)
	}
	got := out.String()
	for _, want := range []string{"codex", "claude", commit2[:12], "HEAD"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected update output to mention %q, got:\n%s", want, got)
		}
	}
}

// TestUpdateDryRunChangesNothing covers AC2: --dry-run shows the same
// summary and changes neither the manager's state nor the installed
// destinations, with and without an interactive terminal.
func TestUpdateDryRunChangesNothing(t *testing.T) {
	for _, interactive := range []bool{true, false} {
		t.Run(fmt.Sprintf("interactive=%v", interactive), func(t *testing.T) {
			env, commit2 := newUpdateFixtureWithPendingCommit(t)
			before := stateBytes(t, env.stateDir)
			var out bytes.Buffer
			if err := update(env.args("--dry-run"), strings.NewReader(""), &out, interactive); err != nil {
				t.Fatalf("update --dry-run: %v", err)
			}
			if after := stateBytes(t, env.stateDir); string(before) != string(after) {
				t.Fatal("--dry-run changed manager state")
			}
			body, err := os.ReadFile(env.sharedSkillPath())
			if err != nil {
				t.Fatal(err)
			}
			if want := skillBody("Preserve evidence.\n"); string(body) != want {
				t.Fatal("--dry-run changed the installed destination")
			}
			if !strings.Contains(out.String(), "Preview: installation was not changed.") {
				t.Fatalf("expected the dry-run notice, got:\n%s", out.String())
			}
			if !strings.Contains(out.String(), commit2[:12]) {
				t.Fatalf("expected the summary to name the resolved commit, got:\n%s", out.String())
			}
		})
	}
}

// TestUpdateNonInteractiveRequiresOutOrDryRun covers the first half of AC3:
// without a terminal, without --dry-run and without --out, update fails
// before changing anything and names --out.
func TestUpdateNonInteractiveRequiresOutOrDryRun(t *testing.T) {
	env, _ := newUpdateFixtureWithPendingCommit(t)
	before := stateBytes(t, env.stateDir)
	err := update(env.args(), strings.NewReader(""), io.Discard, false)
	if err == nil || !strings.Contains(err.Error(), "--out") {
		t.Fatalf("expected an error naming --out, got %v", err)
	}
	if after := stateBytes(t, env.stateDir); string(before) != string(after) {
		t.Fatal("the failed non-interactive update changed manager state")
	}
}

// TestUpdateNonInteractiveWithOutSavesApplicablePlan covers the second half
// of AC3: --out saves a plan carrying the full source commit, which `hive
// apply --plan FILE` applies successfully.
func TestUpdateNonInteractiveWithOutSavesApplicablePlan(t *testing.T) {
	env, commit2 := newUpdateFixtureWithPendingCommit(t)
	planFile := filepath.Join(filepath.Dir(env.stateDir), "update.plan.json")
	var out bytes.Buffer
	if err := update(env.args("--out", planFile), strings.NewReader(""), &out, false); err != nil {
		t.Fatalf("update --out: %v", err)
	}
	if !strings.Contains(out.String(), "hive apply --plan "+planFile) {
		t.Fatalf("expected the output to name hive apply --plan %s, got:\n%s", planFile, out.String())
	}
	saved, err := management.LoadPlan(planFile)
	if err != nil {
		t.Fatalf("saved plan does not load: %v", err)
	}
	if saved.SourceCommit != commit2 {
		t.Fatalf("expected the saved plan to carry the full source commit %s, got %q", commit2, saved.SourceCommit)
	}
	if err := run([]string{"apply", "--plan", planFile}); err != nil {
		t.Fatalf("hive apply --plan %s: %v", planFile, err)
	}
	body, err := os.ReadFile(env.sharedSkillPath())
	if err != nil {
		t.Fatal(err)
	}
	if want := skillBody("Preserve evidence, updated.\n"); string(body) != want {
		t.Fatal("applying the saved plan did not install the resolved commit's content")
	}
}

// TestUpdateRejectsBadRevSourceOrMissingGit covers AC4: a leading-dash rev,
// an unknown commit, a --source that is not a Git checkout, and a PATH
// without git each fail with a clear error before anything is written.
func TestUpdateRejectsBadRevSourceOrMissingGit(t *testing.T) {
	env, _ := newUpdateFixtureWithPendingCommit(t)
	before := stateBytes(t, env.stateDir)

	t.Run("rev starts with dash", func(t *testing.T) {
		if err := update(env.args("--rev", "-x"), strings.NewReader(""), io.Discard, false); err == nil {
			t.Fatal("expected an error for a leading-dash --rev")
		}
	})
	t.Run("unknown commit", func(t *testing.T) {
		if err := update(env.args("--rev", strings.Repeat("f", 40)), strings.NewReader(""), io.Discard, false); err == nil {
			t.Fatal("expected an error for an unknown commit")
		}
	})
	t.Run("source is not a Git checkout", func(t *testing.T) {
		notGit := t.TempDir()
		if err := update([]string{"--home", env.home, "--state-dir", env.stateDir, "--source", notGit}, strings.NewReader(""), io.Discard, false); err == nil {
			t.Fatal("expected an error for a non-Git --source")
		}
	})
	t.Run("git missing from PATH", func(t *testing.T) {
		t.Setenv("PATH", "")
		err := update(env.args(), strings.NewReader(""), io.Discard, false)
		if err == nil || !strings.Contains(err.Error(), "install.sh") {
			t.Fatalf("expected a Git-missing error naming install.sh, got %v", err)
		}
	})

	if after := stateBytes(t, env.stateDir); string(before) != string(after) {
		t.Fatal("a rejected update changed manager state")
	}
}

func extractDirs(t *testing.T, tmp string) []string {
	t.Helper()
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".hive-extract-") {
			names = append(names, e.Name())
		}
	}
	return names
}

// TestUpdateCleansUpExtractionDirOnSuccessAndFailure covers AC5: the
// `.hive-extract-*` directory Extract creates under TMPDIR is gone after
// update returns, whether it succeeds or fails after extraction.
func TestUpdateCleansUpExtractionDirOnSuccessAndFailure(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env, _ := newUpdateFixtureWithPendingCommit(t)
		tmp := t.TempDir()
		t.Setenv("TMPDIR", tmp)
		var out bytes.Buffer
		if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
			t.Fatalf("update: %v\noutput:\n%s", err, out.String())
		}
		if leftover := extractDirs(t, tmp); len(leftover) != 0 {
			t.Fatalf("leftover extraction directories after success: %v", leftover)
		}
	})
	t.Run("failure after extraction", func(t *testing.T) {
		env, _ := newUpdateFixtureWithPendingCommit(t)
		// Two agent files sharing one role name: validateRelease rejects
		// this catalogue only after BuildPlan reads it from the extracted
		// tree, i.e. after Extract has already run. No hook is used.
		putCharacterization(t, filepath.Join(env.repo, "content/agents/other/test-agent.md"), "---\nname: test-agent\ndescription: Duplicate role name\nmodel_profile: execution\naccess_profile: observe\n---\nDuplicate.\n")
		gitCommitAll(t, env.repo, "duplicate role")
		tmp := t.TempDir()
		t.Setenv("TMPDIR", tmp)
		if err := update(env.args(), strings.NewReader(""), io.Discard, false); err == nil {
			t.Fatal("expected an error for a duplicate agent role name")
		}
		if leftover := extractDirs(t, tmp); len(leftover) != 0 {
			t.Fatalf("leftover extraction directories after a post-extraction failure: %v", leftover)
		}
	})
}

// TestUpdateRecordsCommitWhenContentUnchanged covers the "mismo contenido"
// case: a second update whose commit only changes a file outside content/
// produces an identical release ("unchanged") and still records its own
// commit, alongside the first one.
func TestUpdateRecordsCommitWhenContentUnchanged(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, env.repo)
	commit1 := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit1, []string{"codex", "claude"})

	var out1 bytes.Buffer
	if err := update(env.args(), strings.NewReader(""), &out1, false); err != nil {
		t.Fatalf("first update (same content as installed): %v\noutput:\n%s", err, out1.String())
	}
	if !strings.Contains(out1.String(), "already up to date") {
		t.Fatalf("expected the first update to report no change, got:\n%s", out1.String())
	}

	putCharacterization(t, filepath.Join(env.repo, "README.md"), "notes outside the catalogue\n")
	commit2 := gitCommitAll(t, env.repo, "unrelated change")

	var out2 bytes.Buffer
	if err := update(env.args(), strings.NewReader(""), &out2, false); err != nil {
		t.Fatalf("second update (unrelated file change): %v\noutput:\n%s", err, out2.String())
	}
	if !strings.Contains(out2.String(), "already up to date") {
		t.Fatalf("expected the second update to report no change, got:\n%s", out2.String())
	}

	entries, err := management.Releases(management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one release (identical content), got %d", len(entries))
	}
	if len(entries[0].Commits) != 2 || entries[0].Commits[0] != commit1 || entries[0].Commits[1] != commit2 {
		t.Fatalf("expected both commits recorded once each in order, got %+v", entries[0].Commits)
	}
}

// TestAdaptGitArchivePassesThroughDistributionExtract is the adapter test
// tasks.md asks for: real `git archive --format=tar --prefix=…/` output,
// with its pax global header and trailing-slash directory names, extracted
// successfully by distribution.Extract once normalized.
func TestAdaptGitArchivePassesThroughDistributionExtract(t *testing.T) {
	gitPath := requireGit(t)
	dir := t.TempDir()
	writeUpdateCatalog(t, dir, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, dir)
	commit := gitCommitAll(t, dir, "init")

	raw, err := archiveGitCommit(gitPath, filteredGitEnv(), dir, commit, "hive-test")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(raw, []byte("pax_global_header")) {
		t.Fatal("test setup: expected git archive to emit a pax global header entry")
	}
	if !bytes.Contains(raw, []byte("hive-test/")) {
		t.Fatal("test setup: expected git archive to emit the prefix directory with a trailing slash")
	}

	gz, err := adaptGitArchive(raw)
	if err != nil {
		t.Fatalf("adaptGitArchive: %v", err)
	}
	root, err := distribution.Extract(gz, t.TempDir())
	if err != nil {
		t.Fatalf("distribution.Extract rejected the adapted archive: %v", err)
	}
	if filepath.Base(root) != "hive-test" {
		t.Fatalf("unexpected extraction root: %s", root)
	}
	if _, err := os.Stat(filepath.Join(root, management.GlobalSource)); err != nil {
		t.Fatalf("expected the global instruction file in the extracted tree: %v", err)
	}
}

// TestArchiveGitCommitFailsPastSizeLimit covers an archive larger than the
// limit: the reader stops at the limit, so Git must be stopped rather than
// left blocked on a full pipe while the command waits for it.
func TestArchiveGitCommitFailsPastSizeLimit(t *testing.T) {
	fake := filepath.Join(t.TempDir(), "git")
	if err := os.WriteFile(fake, []byte("#!/bin/sh\nexec yes\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	old := archiveSizeLimit
	archiveSizeLimit = 1 << 10
	t.Cleanup(func() { archiveSizeLimit = old })

	done := make(chan error, 1)
	go func() {
		_, err := archiveGitCommit(fake, os.Environ(), t.TempDir(), strings.Repeat("a", 40), "hive-x")
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "exceeds the package size limit") {
			t.Fatalf("expected a size-limit error, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("archiveGitCommit did not return after exceeding the size limit")
	}
}

func TestFilteredGitEnvDropsRepositoryOverrides(t *testing.T) {
	t.Setenv("GIT_DIR", "/elsewhere/.git")
	t.Setenv("GIT_WORK_TREE", "/elsewhere")
	t.Setenv("GIT_COMMON_DIR", "/elsewhere/.git")
	t.Setenv("GIT_AUTHOR_NAME", "kept")
	var kept bool
	for _, kv := range filteredGitEnv() {
		key, _, _ := strings.Cut(kv, "=")
		switch key {
		case "GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR":
			t.Fatalf("%s reached the Git environment", key)
		case "GIT_AUTHOR_NAME":
			kept = true
		}
	}
	if !kept {
		t.Fatal("unrelated Git variables must be preserved")
	}
}
