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

// Update detects the wait floor alone: no question-card or agent-browser probe
// runs, only the Codex version check the wait floor needs.
func TestUpdateRunsOnlyTheWaitFloorProbe(t *testing.T) {
	env := upToDateUpdateEnv(t)
	f := &abFake{goos: "darwin", goarch: "arm64", codexVersion: "codex-cli 0.164.0", path: map[string]bool{"codex": true, "node": true, "npm": true, "agent-browser": true}, node: "v24.1.0"}
	useAgentBrowserFake(t, f)
	saved := updateAdapterFactory
	t.Cleanup(func() { updateAdapterFactory = saved })
	updateAdapterFactory = func(_ updateFlags, interactive bool, codexHome string) onboardingAdapter {
		a := newNativeProviderAdapter(false, false, interactive)
		a.codexHome = codexHome
		return a
	}
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("\ny\n"), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
	if f.codexProbes != 0 || f.nodeCalls != 0 || f.npmCalls != 0 || len(f.execs) != 0 {
		t.Fatalf("unrelated probes ran: features=%d node=%d npm=%d execs=%v", f.codexProbes, f.nodeCalls, f.npmCalls, f.execs)
	}
	if f.versionProbes != 1 {
		t.Fatalf("version probes = %d, want 1", f.versionProbes)
	}
	if _, err := os.Stat(env.configPath()); err != nil {
		t.Fatalf("the offer did not reach the plan: %v", err)
	}
}

func TestUpdateViewEscOnTheCapabilityDialogCancelsTheUpdateLikeTheCommand(t *testing.T) {
	env := newUpdateTwinEnv(t, true)
	useFloorAdapterForUpdate(t, newFloorAdapter(false))
	_, d := env.openApp(t, 100, 30)
	setSource(d, env.repo)
	d.key("enter")
	d.mustShow("[x] codex-wait-floor")
	d.key("esc")
	d.mustShow("Cancelled. No changes applied.")
	d.mustNotShow("already up to date")
	mustNotExist(t, filepath.Join(env.aHome, ".codex", "config.toml"))
	entries, err := management.Releases(management.Options{Scope: "user", Home: env.aHome, StateDir: env.aState})
	if err != nil || len(entries) == 0 || len(entries[0].Commits) != 0 {
		t.Fatalf("a cancelled update recorded a commit: %+v err=%v", entries, err)
	}
}

func TestUpdateViewUncheckingTheOfferAndConfirmingRecordsTheSourceCommit(t *testing.T) {
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
	entries, err := management.Releases(management.Options{Scope: "user", Home: env.aHome, StateDir: env.aState})
	if err != nil || len(entries) == 0 || len(entries[0].Commits) < 1 {
		t.Fatalf("commit not recorded: %+v err=%v", entries, err)
	}
}

// nativeFloorFactory makes update use the native adapter against codexHome, on
// a fake machine where Codex 0.164.0 is installed.
func nativeFloorFactory(t *testing.T, codexHome string) {
	t.Helper()
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", codexVersion: "codex-cli 0.164.0", path: map[string]bool{"codex": true}})
	saved := updateAdapterFactory
	t.Cleanup(func() { updateAdapterFactory = saved })
	updateAdapterFactory = func(_ updateFlags, interactive bool, _ string) onboardingAdapter {
		a := newNativeProviderAdapter(false, false, interactive)
		a.codexHome = codexHome
		return a
	}
}

// recordWaitFloorFor writes the wait floor under home and leaves Hive's record
// of it in stateDir.
func recordWaitFloorFor(t *testing.T, home, stateDir string, hosts []string) {
	t.Helper()
	p, err := management.BuildPlan("install", management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: minimalTestSource(t), Hosts: hosts, CodexWaitFloor: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateDoesNotOfferTheWaitFloorWhenTheRecordIsForAnotherCodexHome(t *testing.T) {
	env := upToDateUpdateEnv(t)
	recordWaitFloorFor(t, env.home, env.stateDir, []string{"claude", "codex"})
	other := t.TempDir()
	nativeFloorFactory(t, other)
	p, _, _, err := planUpdate(updateFlags{Rev: "HEAD", Source: env.repo, Home: env.home, StateDir: env.stateDir})
	if err != nil {
		t.Fatal(err)
	}
	_, offers, err := detectUpdateOffers(p, updateFlags{Home: env.home, StateDir: env.stateDir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(offers) != 1 || !offers[0].ManualOnly || offers[0].DefaultOn {
		t.Fatalf("offers = %+v, want one manual offer", offers)
	}
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader("1\ny\n"), &out, true); err != nil {
		t.Fatalf("update aborted: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "another Codex home") {
		t.Fatalf("the summary does not explain the record:\n%s", out.String())
	}
	mustNotExist(t, filepath.Join(other, "config.toml"))
}

func TestInstallDoesNotOfferTheWaitFloorWhenTheRecordIsForAnotherCodexHome(t *testing.T) {
	home := t.TempDir()
	installWithFloor(t, newFloorAdapter(false), home, "\ny\n")
	other := t.TempDir()
	nativeFloorFactory(t, other)
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	var offers []providerOffer
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) {
		a := newNativeProviderAdapter(false, false, true)
		a.codexHome = other
		return &recordingDetect{onboardingAdapter: a, offers: &offers}, nil
	}
	var out strings.Builder
	if err := installWithDependencies(installArgs(t, home), strings.NewReader("none\ny\n"), &out, true, dependencies); err != nil {
		t.Fatalf("install aborted: %v\n%s", err, out.String())
	}
	var floor *providerOffer
	for i := range offers {
		if offers[i].ID == "codex-wait-floor" {
			floor = &offers[i]
		}
	}
	if floor == nil || !floor.ManualOnly || floor.DefaultOn {
		t.Fatalf("offers = %+v", offers)
	}
	mustNotExist(t, filepath.Join(other, "config.toml"))
}

type recordingDetect struct {
	onboardingAdapter
	offers *[]providerOffer
}

func (r *recordingDetect) Detect(o management.Options) ([]providerOffer, error) {
	offers, err := r.onboardingAdapter.Detect(o)
	*r.offers = offers
	return offers, err
}

func TestUpdateDoesNotOfferTheWaitFloorWithoutCodexRegistered(t *testing.T) {
	requireGit(t)
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, env.repo)
	commit := gitCommitAll(t, env.repo, "init")
	installFromCommit(t, env.home, env.stateDir, env.repo, commit, []string{"claude", "pi"})
	nativeFloorFactory(t, filepath.Join(env.home, ".codex"))
	f := updateFlags{Rev: "HEAD", Source: env.repo, Home: env.home, StateDir: env.stateDir}
	p, _, _, err := planUpdate(f)
	if err != nil {
		t.Fatal(err)
	}
	if _, offers, err := detectUpdateOffers(p, f, true); err != nil || len(offers) != 0 {
		t.Fatalf("offers = %+v err=%v", offers, err)
	}
	var out bytes.Buffer
	if err := update(env.args(), strings.NewReader(""), &out, true); err != nil {
		t.Fatalf("update: %v\n%s", err, out.String())
	}
}
