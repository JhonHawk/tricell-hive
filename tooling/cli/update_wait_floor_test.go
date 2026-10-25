package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// useFloorAdapterForUpdate makes hive update offer the Codex wait floor through
// a, as the native adapter does on a machine with Codex 0.164 or later.
func useFloorAdapterForUpdate(t *testing.T, a *floorAdapter) *int {
	t.Helper()
	detects := new(int)
	saved := updateAdapterFactory
	t.Cleanup(func() { updateAdapterFactory = saved })
	updateAdapterFactory = func(updateFlags, bool, string) onboardingAdapter {
		return &countingAdapter{floorAdapter: a, detects: detects}
	}
	return detects
}

type countingAdapter struct {
	*floorAdapter
	detects *int
}

func (c *countingAdapter) Detect(o management.Options) ([]providerOffer, error) {
	*c.detects++
	return c.floorAdapter.Detect(o)
}

// upToDateUpdateEnv installs the repository's HEAD for codex and claude, so a
// later update of the same commit changes nothing in Hive's own files.
func upToDateUpdateEnv(t *testing.T) updateEnv {
	t.Helper()
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, env.repo)
	commit := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit, []string{"codex", "claude"})
	return env
}

func (e updateEnv) configPath() string { return filepath.Join(e.home, ".codex", "config.toml") }

func TestUpdateOffersTheWaitFloorAndAsksConfirmationEvenWhenHiveIsUpToDate(t *testing.T) {
	env := upToDateUpdateEnv(t)
	a := newFloorAdapter(false)
	useFloorAdapterForUpdate(t, a)
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("\ny\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Apply these changes?") {
		t.Fatalf("no confirmation although the plan writes Codex's config:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "writes the wait floor") {
		t.Fatalf("the summary does not describe the capability:\n%s", out.String())
	}
	got, err := os.ReadFile(env.configPath())
	if err != nil || !strings.Contains(string(got), "default_wait_timeout_ms = 600000") {
		t.Fatalf("config.toml = %q err=%v", got, err)
	}
	if a.plan.CodexWaitFloor == nil {
		t.Fatal("the adapter was given a plan without the wait floor")
	}
}

func TestUpdateDeclinedWaitFloorKeepsTheDirectApply(t *testing.T) {
	env := upToDateUpdateEnv(t)
	useFloorAdapterForUpdate(t, newFloorAdapter(false))
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("none\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "Apply these changes?") || !strings.Contains(out.String(), "already up to date") {
		t.Fatalf("a declined capability must leave the up-to-date path:\n%s", out.String())
	}
	mustNotExist(t, env.configPath())
}

func TestUpdateWaitFloorDeclinedAtConfirmationWritesNothing(t *testing.T) {
	env := upToDateUpdateEnv(t)
	useFloorAdapterForUpdate(t, newFloorAdapter(false))
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("\nn\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	mustNotExist(t, env.configPath())
}

func TestUpdateDoesNotOfferTheWaitFloorWithoutATerminal(t *testing.T) {
	env := upToDateUpdateEnv(t)
	detects := useFloorAdapterForUpdate(t, newFloorAdapter(false))
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader(""), &out, false); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	if *detects != 0 {
		t.Fatalf("Detect ran %d times without a terminal", *detects)
	}
	mustNotExist(t, env.configPath())
}

func TestUpdateManualWaitFloorOfferChangesNothing(t *testing.T) {
	env := upToDateUpdateEnv(t)
	useFloorAdapterForUpdate(t, newFloorAdapter(true))
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("1\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	mustNotExist(t, env.configPath())
	if !strings.Contains(out.String(), "already up to date") {
		t.Fatalf("a manual offer must not turn the update into a write:\n%s", out.String())
	}
}

// The Update view offers the wait floor like the command: the choice rebuilds
// the plan and the view asks for confirmation although Hive's files are current.
func TestUpdateViewOffersTheWaitFloorAndConfirmsEvenWhenHiveIsUpToDate(t *testing.T) {
	env := newUpdateTwinEnv(t, true)
	a := newFloorAdapter(false)
	useFloorAdapterForUpdate(t, a)
	_, d := env.openApp(t, 100, 30)
	setSource(d, env.repo)
	d.key("enter")
	d.mustShow("Optional capabilities", "[x] codex-wait-floor")
	d.key("enter")
	d.mustShow("[Apply]", "writes the wait floor")
	d.key("enter")
	d.mustShow("Hive updated")
	got, err := os.ReadFile(filepath.Join(env.aHome, ".codex", "config.toml"))
	if err != nil || !strings.Contains(string(got), "default_wait_timeout_ms = 600000") {
		t.Fatalf("config.toml = %q err=%v", got, err)
	}
	if a.plan.CodexWaitFloor == nil {
		t.Fatal("the adapter was given a plan without the wait floor")
	}
}

func TestUpdateViewWaitFloorSkippedLeavesTheDirectApply(t *testing.T) {
	env := newUpdateTwinEnv(t, true)
	useFloorAdapterForUpdate(t, newFloorAdapter(false))
	_, d := env.openApp(t, 100, 30)
	setSource(d, env.repo)
	d.key("enter")
	d.mustShow("[x] codex-wait-floor")
	d.key("space")
	d.key("enter")
	d.mustShow("already up to date")
	mustNotExist(t, filepath.Join(env.aHome, ".codex", "config.toml"))
}
