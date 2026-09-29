package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// The tests here cover the load-error states of the CLIs and Voice views
// (issue #46): a managed file changed by hand must not empty the CLIs view,
// and a state that cannot be read must read in plain words first.

// driftedEnv installs hosts from the update fixture's catalog, whose skill is
// a shared file every host consumes, then edits that shared skill by hand, the
// way a user's change leaves a managed file in drift.
func driftedEnv(t *testing.T, hosts []string) updateEnv {
	t.Helper()
	env := newUpdateEnv(t)
	writeUpdateCatalog(t, env.repo, skillBody("Preserve evidence.\n"))
	installDirect(t, env.home, env.stateDir, env.repo, hosts)
	path := env.sharedSkillPath()
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(original), "Preserve evidence.", "Edited by hand.", 1)
	if edited == string(original) {
		t.Fatal("setup: nothing was edited")
	}
	if err := os.WriteFile(path, []byte(edited), 0o600); err != nil {
		t.Fatal(err)
	}
	return env
}

func openDriftedCLIs(t *testing.T, env updateEnv, width, height int) *appDriver {
	t.Helper()
	_, d := newTestApp(t, hostsAppConfig(t, env.home, env.stateDir, env.repo, hostsTestDeps(coreOnlyAdapterFactory)), width, height)
	d.key("enter")
	d.mustShow("CLIs")
	return d
}

// TestHostsViewListsHostsWhenAManagedFileWasChanged (J1): a hand-edited shared
// skill makes the legacy scan fail; the view still lists every host with its
// state and drift, says in plain words what happened and where to go, and
// keeps the raw error on its own Detail line.
func TestHostsViewListsHostsWhenAManagedFileWasChanged(t *testing.T) {
	env := driftedEnv(t, installerHosts)
	skill := env.sharedSkillPath()
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		d := openDriftedCLIs(t, env, size[0], size[1])
		d.mustNotShow("Cannot read the CLIs")
		for _, want := range []string{
			"A file Hive installed was changed (" + skill + "). Open Diagnostics to see how to restore it before installing or removing.",
			"Detail: modified legacy file requires manual resolution: " + skill,
		} {
			mustShowFlat(d, want)
		}
		// The first mention of "legacy" is the raw detail: the plain words never call the file legacy.
		plain, _, _ := strings.Cut(d.screen(), "Detail:")
		if strings.Contains(plain, "legacy") {
			t.Errorf("the plain words mention a legacy file:\n%s", d.screen())
		}
		rows := hostRows(d)
		if len(rows) != len(installerHosts) {
			t.Fatalf("%dx%d: %d rows, want %d:\n%s", size[0], size[1], len(rows), len(installerHosts), d.screen())
		}
		for _, r := range rows {
			if r.state != "registered" || !r.checked || r.drift == "0" || r.drift == "-" {
				t.Errorf("%s row = %+v, want registered, checked and drift counted", r.name, r)
			}
		}
		assertFits(t, d, size[0], size[1])
	}
}

// TestHostsViewRefusesChangesWhenAManagedFileWasChanged (J1): listing the hosts
// does not let a change through; both a removal and an installation are
// refused as before, and nothing is written.
func TestHostsViewRefusesChangesWhenAManagedFileWasChanged(t *testing.T) {
	t.Run("removal", func(t *testing.T) {
		env := driftedEnv(t, installerHosts)
		d := openDriftedCLIs(t, env, 80, 24)
		homeBefore, stateBefore := collectFiles(t, env.home), stateBytes(t, env.stateDir)
		toggle(t, d, "claude")
		d.key("a")
		mustShowFlat(d, "modified managed skill")
		assertFits(t, d, 80, 24)
		assertUntouched(t, env, homeBefore, stateBefore)
	})
	t.Run("installation", func(t *testing.T) {
		env := driftedEnv(t, installerHosts[:5])
		d := openDriftedCLIs(t, env, 80, 24)
		homeBefore, stateBefore := collectFiles(t, env.home), stateBytes(t, env.stateDir)
		toggle(t, d, "pi")
		d.key("a")
		mustShowFlat(d, "modified legacy file requires manual resolution")
		assertFits(t, d, 80, 24)
		assertUntouched(t, env, homeBefore, stateBefore)
	})
}

func assertUntouched(t *testing.T, env updateEnv, homeBefore map[string][]byte, stateBefore []byte) {
	t.Helper()
	after := collectFiles(t, env.home)
	if len(after) != len(homeBefore) {
		t.Errorf("the home has %d files, had %d", len(after), len(homeBefore))
	}
	for rel, data := range homeBefore {
		if !bytes.Equal(after[rel], data) {
			t.Errorf("%s changed", rel)
		}
	}
	if !bytes.Equal(stateBytes(t, env.stateDir), stateBefore) {
		t.Error("state.json changed")
	}
}

// TestHostsViewNamesAnUnknownLegacyScanFailureWithoutBlamingAFile (J1): a scan
// failure other than a changed file keeps the rows and gets the general note.
func TestHostsViewNamesAnUnknownLegacyScanFailureWithoutBlamingAFile(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.DiscoverHosts = func(o management.Options) ([]hostCandidate, error) {
		candidates, _ := detectInstallerHosts(o)
		for i := range candidates {
			candidates[i].Detected = true
		}
		return candidates, &legacyScanError{err: errors.New("unsupported legacy manifest: /x/manifest.json")}
	}
	_, d := openHostsApp(t, home, stateDir, minimalTestSource(t), deps)
	mustShowFlat(d, "Hive could not check for a legacy installation, so installing or removing may be refused. Open Diagnostics for details.")
	mustShowFlat(d, "Detail: unsupported legacy manifest: /x/manifest.json")
	d.mustNotShow("A file Hive installed was changed")
	if n := len(hostRows(d)); n != len(installerHosts) {
		t.Fatalf("%d rows, want %d:\n%s", n, len(installerHosts), d.screen())
	}
	assertFits(t, d, 80, 24)
}
