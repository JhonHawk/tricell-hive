package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// Tests of the CLIs view (T8). Each one drives the full application through
// the driver in tui_test.go and, where the plan asks for it, compares the
// resulting home and state.json with a twin built by the equivalent command.

// hostsTestDeps is the default install dependencies with a DiscoverHosts that
// reports every supported CLI as detected: a synthetic home never detects an
// executable, and the view lists only detected or registered CLIs.
func hostsTestDeps(factory onboardingAdapterFactory) installDependencies {
	deps := defaultInstallDependencies(factory)
	deps.DiscoverHosts = func(o management.Options) ([]hostCandidate, error) {
		candidates, err := detectInstallerHosts(o)
		for i := range candidates {
			candidates[i].Detected = true
		}
		return candidates, err
	}
	return deps
}

func hostsAppConfig(t *testing.T, home, stateDir, source string, deps installDependencies) appConfig {
	t.Helper()
	mo := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source}
	_, normalized, err := management.NormalizeOptions(mo)
	if err != nil {
		t.Fatal(err)
	}
	mo.StateDir = normalized
	return appConfig{Options: mo, ExplicitStateDir: true, Deps: deps, Dark: true}
}

// openHostsApp opens the application over the home and enters the CLIs view.
func openHostsApp(t *testing.T, home, stateDir, source string, deps installDependencies) (*appModel, *appDriver) {
	t.Helper()
	m, d := newTestApp(t, hostsAppConfig(t, home, stateDir, source, deps), 80, 24)
	d.key("enter")
	d.mustShow("CLIs")
	return m, d
}

type parsedRow struct {
	cursor, checked                      bool
	name, state, release, version, drift string
}

var hostRowPattern = regexp.MustCompile(`^(> |  )\[( |x)\] (.*)$`)
var columnGap = regexp.MustCompile(`\s{2,}`)

// hostRows parses the CLI rows on the screen.
func hostRows(d *appDriver) []parsedRow {
	var rows []parsedRow
	for _, line := range d.lines() {
		m := hostRowPattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		cols := columnGap.Split(strings.TrimSpace(m[3]), -1)
		row := parsedRow{cursor: m[1] == "> ", checked: m[2] == "x"}
		for i, dst := range []*string{&row.name, &row.state, &row.release, &row.version, &row.drift} {
			if i < len(cols) {
				*dst = cols[i]
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// cursorTo moves the cursor to the named CLI's row.
func cursorTo(t *testing.T, d *appDriver, name string) {
	t.Helper()
	rows := hostRows(d)
	from, to := -1, -1
	for i, r := range rows {
		if r.cursor {
			from = i
		}
		if r.name == name {
			to = i
		}
	}
	if from < 0 || to < 0 {
		t.Fatalf("cannot move to %q (cursor row %d, target row %d):\n%s", name, from, to, d.screen())
	}
	for ; from < to; from++ {
		d.key("down")
	}
	for ; from > to; from-- {
		d.key("up")
	}
}

func toggle(t *testing.T, d *appDriver, names ...string) {
	t.Helper()
	for _, name := range names {
		cursorTo(t, d, name)
		d.key("space")
	}
}

func rowFor(t *testing.T, d *appDriver, name string) parsedRow {
	t.Helper()
	for _, r := range hostRows(d) {
		if r.name == name {
			return r
		}
	}
	t.Fatalf("no row for %q:\n%s", name, d.screen())
	return parsedRow{}
}

func assertTwin(t *testing.T, home, stateDir, twinHome, twinState string) {
	t.Helper()
	assertHomesMatch(t, home, twinHome)
	assertStateJSONMatches(t, stateDir, home, twinState, twinHome)
}

// removeViaCommand is the equivalent of `hive plan remove` followed by `hive
// apply`.
func removeViaCommand(t *testing.T, home, stateDir string, hosts ...string) {
	t.Helper()
	plan, err := management.BuildPlan("remove", management.Options{Scope: "user", Home: home, StateDir: stateDir, Hosts: hosts})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(plan); err != nil {
		t.Fatal(err)
	}
}

// seedLegacyCodex writes the legacy Hive files a pre-rebuild installation left
// for Codex, so DetectLegacyHosts reports it.
func seedLegacyCodex(t *testing.T, home string) {
	t.Helper()
	b, err := os.ReadFile("../legacy/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Files []struct {
			Root, Path string
			Data       []byte
		}
	}
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	seeded := false
	for _, f := range c.Files {
		if f.Root == "codex" && f.Path == "AGENTS.md" {
			path := filepath.Join(home, ".codex", f.Path)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, f.Data, 0600); err != nil {
				t.Fatal(err)
			}
			seeded = true
		}
	}
	if !seeded {
		t.Fatal("legacy catalog has no codex AGENTS.md")
	}
}

// statusFields returns, per host, what `hive status` reports: the short
// release, the product version and the number of resources in drift.
func statusFields(t *testing.T, home, stateDir, hosts string) map[string][3]string {
	t.Helper()
	out := captureStdout(t, func() {
		args := []string{"status", "--scope", "user", "--home", home, "--state-dir", stateDir, "--hosts", hosts}
		if err := run(args); err != nil {
			t.Fatalf("hive status: %v", err)
		}
	})
	var entries []management.StatusEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("hive status output %q: %v", out, err)
	}
	fields := map[string][3]string{}
	drift := map[string]int{}
	for _, e := range entries {
		if e.Kind == "voice" {
			continue
		}
		f, ok := fields[e.Host]
		if !ok {
			f = [3]string{"-", "-", ""}
		}
		if e.Release != "" {
			f[0] = shortHash(e.Release)
		}
		if e.ProductVersion != "" {
			f[1] = e.ProductVersion
		}
		if e.Status == "drift" {
			drift[e.Host]++
		}
		fields[e.Host] = f
	}
	for host, f := range fields {
		f[2] = fmt.Sprint(drift[host])
		fields[host] = f
	}
	return fields
}

// TestHostsViewRowsMatchStatus covers AC3: one row per detected or registered
// CLI with a checkbox, its state, and the release, version and drift `hive
// status` reports for the same home.
func TestHostsViewRowsMatchStatus(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude,codex", "y\n", deps)
	// Make claude drift: change its managed guidance file.
	guidance := filepath.Join(home, ".claude", "CLAUDE.md")
	data, err := os.ReadFile(guidance)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(guidance, bytes.Replace(data, []byte("Minimal test guidance."), []byte("Edited by hand."), 1), 0600); err != nil {
		t.Fatal(err)
	}
	want := statusFields(t, home, stateDir, "claude,codex")
	if want["claude"][2] == "0" {
		t.Fatalf("test setup did not produce drift: %v", want)
	}

	_, d := openHostsApp(t, home, stateDir, source, deps)
	rows := hostRows(d)
	names := make([]string, len(rows))
	for i, r := range rows {
		names[i] = r.name
	}
	if strings.Join(names, ",") != "claude,codex,cursor,grok,opencode,pi" {
		t.Fatalf("rows = %v\n%s", names, d.screen())
	}
	for _, name := range []string{"claude", "codex"} {
		r := rowFor(t, d, name)
		got := [3]string{r.release, r.version, r.drift}
		if !r.checked || r.state != "registered" || got != want[name] {
			t.Errorf("%s row = %+v, want checked, registered and %v", name, r, want[name])
		}
	}
	for _, name := range []string{"cursor", "grok", "opencode", "pi"} {
		r := rowFor(t, d, name)
		if r.checked || r.state != "detected" || r.release != "-" || r.version != "-" || r.drift != "-" {
			t.Errorf("%s row = %+v, want unchecked, detected and dashes", name, r)
		}
	}
	if !hostRows(d)[0].cursor {
		t.Errorf("the cursor does not start on the first row:\n%s", d.screen())
	}
}

// TestHostsViewLabelsLegacyInstall covers AC3: an unregistered CLI with a
// legacy installation is labeled so.
func TestHostsViewLabelsLegacyInstall(t *testing.T) {
	source := minimalTestSource(t)
	home, stateDir := newHostsTestHome(t)
	seedLegacyCodex(t, home)
	deps := defaultInstallDependencies(coreOnlyAdapterFactory) // real detection: only legacy shows
	_, d := openHostsApp(t, home, stateDir, source, deps)
	rows := hostRows(d)
	if len(rows) != 1 || rows[0].name != "codex" || rows[0].state != "legacy install" || rows[0].checked {
		t.Fatalf("rows = %+v\n%s", rows, d.screen())
	}
}

// TestHostsViewEmptyStateNamesTheProblem covers the limit state: nothing
// detected or registered.
func TestHostsViewEmptyStateNamesTheProblem(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, minimalTestSource(t), defaultInstallDependencies(coreOnlyAdapterFactory))
	d.mustShow("No CLI hosts were detected or registered")
	d.key("a")
	d.mustShow("Nothing to apply")
	d.key("u")
	d.mustShow("No CLI hosts are registered")
}

// TestHostsViewOpenAndLeaveKeepsState covers AC3: opening the view and leaving
// it leaves state.json and the home identical.
func TestHostsViewOpenAndLeaveKeepsState(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude,codex", "y\n", deps)
	stateBefore, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	filesBefore := collectFiles(t, home)

	for _, leave := range []string{"esc", "backspace"} {
		_, d := openHostsApp(t, home, stateDir, source, deps)
		toggle(t, d, "claude", "pi") // moving and toggling is not applying
		d.key(leave)
		d.mustShow("> CLIs")
		d.mustNotShow("[x]")
	}
	stateAfter, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stateBefore, stateAfter) {
		t.Fatal("state.json changed after opening and leaving the view")
	}
	filesAfter := collectFiles(t, home)
	if len(filesAfter) != len(filesBefore) {
		t.Fatalf("the home changed: %d files before, %d after", len(filesBefore), len(filesAfter))
	}
	for rel, data := range filesBefore {
		if !bytes.Equal(filesAfter[rel], data) {
			t.Fatalf("%s changed", rel)
		}
	}
}

// TestHostsViewNothingToApply covers the no-change limit state.
func TestHostsViewNothingToApply(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", deps)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	d.key("a")
	d.mustShow("Nothing to apply")
	if len(hostRows(d)) != 6 {
		t.Fatalf("rows vanished:\n%s", d.screen())
	}
}

// TestHostsViewInstallsCodexAndClaude covers AC4 (additions only): checking
// two CLIs, reading the summary and confirming leaves the same files and
// state as `hive install --hosts claude,codex`, and the view shows the result
// and the refreshed rows.
func TestHostsViewInstallsCodexAndClaude(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	twinHome, twinState := newHostsTestHome(t)
	installViaText(t, twinHome, twinState, source, "claude,codex", "y\n", deps)

	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "claude", "codex")
	d.key("a")
	d.mustShow("Install claude, codex", "Hive files to install or update", "[Apply]")
	d.mustNotShow("Step 1 of 2", "Use --dry-run") // the view has no --dry-run flag to suggest
	d.key("enter")
	d.mustShow("Hive installed and verified", "Open new CLI sessions.")
	if len(menuRows(d)) != 0 {
		t.Fatal("the menu is drawn over the view")
	}
	if r := rowFor(t, d, "claude"); !r.checked || r.state != "registered" || r.release == "-" {
		t.Fatalf("rows were not refreshed from the state: %+v\n%s", r, d.screen())
	}
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewDecliningKeepsHomeUntouched covers AC8: rejecting the summary
// returns to the view with the message inside it and changes nothing.
func TestHostsViewDecliningKeepsHomeUntouched(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	for _, reject := range [][]string{{"n"}, {"esc"}, {"backspace"}, {"right", "enter"}} {
		t.Run(strings.Join(reject, "+"), func(t *testing.T) {
			home, stateDir := newHostsTestHome(t)
			_, d := openHostsApp(t, home, stateDir, source, deps)
			toggle(t, d, "claude")
			d.key("a")
			d.mustShow("[Apply]")
			d.key(reject...)
			d.mustShow("Cancelled. No changes applied.", "CLIs")
			d.mustNotShow("[Apply]")
			if rowFor(t, d, "claude").checked {
				t.Fatal("the checkboxes must reload from the state after the flow ends")
			}
			entries, err := os.ReadDir(home)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("a declined install wrote to the home: %v", entries)
			}
		})
	}
}

// TestHostsViewInstallsWithVersionedCapability covers AC4: an optional
// capability that asks for a version, chosen in the view, gives the same files,
// state and provider request as `hive install` with the same answers.
func TestHostsViewInstallsWithVersionedCapability(t *testing.T) {
	source := minimalTestSource(t)
	offers := []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture", Effects: []string{"installs the CLI"}}}
	newAdapter := func() *testOnboardingAdapter {
		return &testOnboardingAdapter{offers: offers, runner: testExternalRunner{}}
	}
	depsFor := func(a *testOnboardingAdapter) installDependencies {
		deps := hostsTestDeps(coreOnlyAdapterFactory)
		deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return a, nil }
		return deps
	}

	twinAdapter := newAdapter()
	twinHome, twinState := newHostsTestHome(t)
	installViaText(t, twinHome, twinState, source, "codex", "1\n0.5.11\ny\n", depsFor(twinAdapter))

	adapter := newAdapter()
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, depsFor(adapter))
	toggle(t, d, "codex")
	d.key("a")
	d.mustShow("Optional capabilities", "[ ] Context7")
	d.mustNotShow("[Apply]")
	d.key("space") // check it: the version field opens with the focus
	d.mustShow("[x] Context7", "Version")
	for _, k := range strings.Split("0.5.12", "") {
		d.key(k)
	}
	d.key("backspace") // inside a field Backspace edits, it does not go back
	d.mustShow("0.5.1")
	d.mustNotShow("0.5.12")
	d.key("1") // 0.5.11
	d.key("enter")
	d.mustShow("[x] Context7", "0.5.11")
	d.key("enter") // list focused: on to the summary
	d.mustShow("Optional capability: context7 0.5.11 from official fixture", "[Apply]")
	d.key("enter")
	d.mustShow("Hive installed and verified")

	if len(adapter.requests) != 1 || adapter.requests[0] != (providerRequest{ID: "context7", Version: "0.5.11"}) {
		t.Fatalf("provider request = %#v", adapter.requests)
	}
	if fmt.Sprint(adapter.requests) != fmt.Sprint(twinAdapter.requests) {
		t.Fatalf("requests differ from hive install: %v vs %v", adapter.requests, twinAdapter.requests)
	}
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewCapabilityNeedsVersion covers the version rule: a checked
// capability with a blank version does not proceed, and Esc from the list
// cancels the install.
func TestHostsViewCapabilityNeedsVersion(t *testing.T) {
	source := minimalTestSource(t)
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture"}},
		runner: testExternalRunner{},
	}
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex")
	d.key("a")
	d.key("space", "enter") // check, close the empty field
	d.key("enter")          // try to proceed
	d.mustShow("exact version required for Context7")
	d.mustNotShow("[Apply]")
	d.key("esc") // leaves the field
	d.key("esc") // cancels the install
	d.mustShow("Cancelled. No changes applied.")
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Fatalf("a cancelled install wrote to the home: %v", entries)
	}
}

// TestHostsViewInstallsGrokWithExpansionToClaude covers AC4: a shared resource
// that requires another CLI shows the notice with Accept and Cancel, and the
// result matches `hive install --hosts grok` accepting the same expansion.
func TestHostsViewInstallsGrokWithExpansionToClaude(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.RequiredHosts = func(management.Options) ([]string, error) { return []string{"claude", "grok"}, nil }

	twinHome, twinState := newHostsTestHome(t)
	installViaText(t, twinHome, twinState, source, "grok", "y\ny\n", deps)

	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "grok")
	d.key("a")
	d.mustShow("Shared resources require selecting: claude", "Accepting will rewrite those resources too", "[Accept]")
	d.key("enter")
	d.mustShow("Install claude, grok", "[Apply]")
	d.key("enter")
	d.mustShow("Hive installed and verified")
	assertTwin(t, home, stateDir, twinHome, twinState)

	// Cancelling the notice changes nothing.
	home2, state2 := newHostsTestHome(t)
	_, d = openHostsApp(t, home2, state2, source, deps)
	toggle(t, d, "grok")
	d.key("a")
	d.key("n")
	d.mustShow("Cancelled. No changes applied.")
	if entries, _ := os.ReadDir(home2); len(entries) != 0 {
		t.Fatalf("a cancelled expansion wrote to the home: %v", entries)
	}
}

// TestHostsViewMigratesLegacyInstall covers AC4: checking a CLI with a legacy
// installation means migrating and installing; the summary says so and the
// result matches `hive install`.
func TestHostsViewMigratesLegacyInstall(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)

	twinHome, twinState := newHostsTestHome(t)
	seedLegacyCodex(t, twinHome)
	installViaText(t, twinHome, twinState, source, "codex", "y\n", deps)

	home, stateDir := newHostsTestHome(t)
	seedLegacyCodex(t, home)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	if r := rowFor(t, d, "codex"); r.state != "legacy install" {
		t.Fatalf("codex row = %+v", r)
	}
	toggle(t, d, "codex")
	d.key("a")
	d.mustShow("Migrate legacy Hive and install the rebuild")
	d.key("enter")
	d.mustShow("Hive installed and verified")
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewTamperedPackageFailsLikeInstall covers AC4: a package that
// fails its verification fails with the message `hive install` gives, inside
// the view, and changes nothing.
func TestHostsViewTamperedPackageFailsLikeInstall(t *testing.T) {
	good := minimalTestSource(t)
	source := t.TempDir()
	if err := os.CopyFS(source, os.DirFS(good)); err != nil {
		t.Fatal(err)
	}
	// A package sentinel without its manifest: distribution.VerifyIfPackaged
	// refuses it.
	if err := os.MkdirAll(filepath.Join(source, "bin"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "bin", "hive"), []byte("x"), 0700); err != nil {
		t.Fatal(err)
	}
	deps := hostsTestDeps(coreOnlyAdapterFactory)

	twinHome, twinState := newHostsTestHome(t)
	var out bytes.Buffer
	args := []string{"--home", twinHome, "--state-dir", twinState, "--source", source, "--hosts", "claude"}
	commandErr := installWithDependencies(args, strings.NewReader("y\n"), &out, true, deps)
	if commandErr == nil {
		t.Fatal("hive install accepted the altered package")
	}

	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "claude")
	d.key("a")
	d.mustShow(commandErr.Error())
	d.mustNotShow("[Apply]")
	if entries, _ := os.ReadDir(home); len(entries) != 0 {
		t.Fatalf("the failed install wrote to the home: %v", entries)
	}
}

// TestHostsViewSourceWithoutCatalog covers the plan-error case: a source with
// no catalog shows its message inside the view.
func TestHostsViewSourceWithoutCatalog(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, t.TempDir(), hostsTestDeps(coreOnlyAdapterFactory))
	toggle(t, d, "claude")
	d.key("a")
	d.mustShow("Run hive from a Hive checkout or package, or pass --source")
	d.mustNotShow("[Apply]")
}

// TestHostsViewRemovesClaudeWithActiveVoice covers AC4 (removals only): the
// summary lists the voice block that goes away, and the result matches `plan
// remove` followed by `apply`.
func TestHostsViewRemovesClaudeWithActiveVoice(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	setUp := func(t *testing.T) (home, stateDir string) {
		t.Helper()
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude", "y\n", deps)
		vp, err := management.BuildVoicePlan("set", management.Options{Source: source, Home: home, StateDir: stateDir}, management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := (management.Engine{}).Apply(vp); err != nil {
			t.Fatal(err)
		}
		return home, stateDir
	}
	twinHome, twinState := setUp(t)
	removeViaCommand(t, twinHome, twinState, "claude")

	home, stateDir := setUp(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "claude") // uncheck
	d.key("a")
	d.mustShow("Remove claude", "Files to remove:", "Voice blocks to remove: 1", "[Apply]")
	d.mustNotShow("Step 1 of 2")
	d.key("enter")
	d.mustShow("Hive removed", "Open new CLI sessions.")
	if r := rowFor(t, d, "claude"); r.checked || r.state != "detected" {
		t.Fatalf("claude row after removal = %+v", r)
	}
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewRemovesThenInstallsInTwoSteps covers AC4 (D3-A): removing Codex
// and installing Pi at once runs two labeled steps, and the result matches
// `plan remove`+`apply` followed by `install --hosts pi`.
func TestHostsViewRemovesThenInstallsInTwoSteps(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	setUp := func(t *testing.T) (home, stateDir string) {
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "codex", "y\n", deps)
		return home, stateDir
	}
	twinHome, twinState := setUp(t)
	removeViaCommand(t, twinHome, twinState, "codex")
	installViaText(t, twinHome, twinState, source, "pi", "y\n", deps)

	home, stateDir := setUp(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex", "pi")
	d.key("a")
	d.mustShow("Step 1 of 2: remove codex", "Remove codex", "[Apply]")
	d.key("enter")
	d.mustShow("Step 1 of 2", "Removed: codex.")
	d.key("enter") // Continue
	d.mustShow("Step 2 of 2: install pi", "[Apply]")
	d.key("enter")
	d.mustShow("Removed: codex.", "Hive installed and verified")
	if r := rowFor(t, d, "pi"); !r.checked || r.state != "registered" {
		t.Fatalf("pi row = %+v", r)
	}
	if r := rowFor(t, d, "codex"); r.checked {
		t.Fatalf("codex row = %+v", r)
	}
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewDecliningSecondStepKeepsTheRemoval covers AC4: rejecting step 2
// leaves step 1 applied and says so.
func TestHostsViewDecliningSecondStepKeepsTheRemoval(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	setUp := func(t *testing.T) (home, stateDir string) {
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "codex", "y\n", deps)
		return home, stateDir
	}
	twinHome, twinState := setUp(t)
	removeViaCommand(t, twinHome, twinState, "codex")

	home, stateDir := setUp(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex", "pi")
	d.key("a", "enter", "enter") // step 1 applied, on to step 2
	d.mustShow("Step 2 of 2: install")
	d.key("n")
	d.mustShow("Removed: codex. Install cancelled; no further changes.")
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewRejectingFirstStepChangesNothing covers AC4: with additions and
// removals, rejecting step 1 applies neither.
func TestHostsViewRejectingFirstStepChangesNothing(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "codex", "y\n", deps)
	before := collectFiles(t, home)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex", "pi")
	d.key("a", "n")
	d.mustShow("Cancelled. No changes applied.")
	d.mustNotShow("Removed:")
	after := collectFiles(t, home)
	if len(before) != len(after) {
		t.Fatalf("files changed: %d before, %d after", len(before), len(after))
	}
}

// TestHostsViewSecondStepFailureShowsRemovedAndError covers AC4: when step 2
// fails, the view says what was removed and shows the error.
func TestHostsViewSecondStepFailureShowsRemovedAndError(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "codex", "y\n", deps)
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return nil, fmt.Errorf("adapter unavailable") }
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex", "pi")
	d.key("a", "enter", "enter")
	d.mustShow("Removed: codex.", "adapter unavailable")
	d.mustNotShow("[Apply]")
	if hosts, err := management.RegisteredHosts(management.Options{Scope: "user", Home: home, StateDir: stateDir}); err != nil || len(hosts) != 0 {
		t.Fatalf("registered hosts = %v (%v), want none: step 1 stays applied", hosts, err)
	}
}

// TestHostsViewRemovalLeavingPendingSkipsInstallAndOffersRecovery covers AC4:
// a removal that fails and leaves an operation pending skips the install and
// opens the recovery view.
func TestHostsViewRemovalLeavingPendingSkipsInstallAndOffersRecovery(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "codex", "y\n", deps)
	pendingFile := filepath.Join(stateDir, "pending.json")

	old := removeApply
	t.Cleanup(func() { removeApply = old })
	removeApply = func(management.Plan) (string, error) {
		if err := os.WriteFile(pendingFile, []byte(`{"id":"core"}`), 0600); err != nil {
			return "", err
		}
		return "", fmt.Errorf("transaction interrupted")
	}
	planned := false
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) {
		planned = true
		return coreOnlyAdapter{}, nil
	}
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex", "pi")
	d.key("a", "enter")
	d.mustShow("Recovery needed", "An operation was interrupted", "[Recover]")
	if planned {
		t.Fatal("the install step ran after a removal that left an operation pending")
	}
	d.mustNotShow("Step 2 of 2")
	d.key("l") // leave it
	d.mustShow("transaction interrupted", "operation is still pending")
	d.mustNotShow("Recovery needed")
}

// TestHostsViewUninstallAll covers AC5: Uninstall all starts on Cancel and
// ignores y; confirming removes every registered CLI and the voice, like `plan
// remove` for all of them.
func TestHostsViewUninstallAll(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	setUp := func(t *testing.T) (home, stateDir string) {
		t.Helper()
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "claude,codex,pi", "y\n", deps)
		vp, err := management.BuildVoicePlan("set", management.Options{Source: source, Home: home, StateDir: stateDir}, management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := (management.Engine{}).Apply(vp); err != nil {
			t.Fatal(err)
		}
		return home, stateDir
	}
	twinHome, twinState := setUp(t)
	removeViaCommand(t, twinHome, twinState, "claude", "codex", "pi")

	home, stateDir := setUp(t)
	before := collectFiles(t, home)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	d.key("u")
	d.mustShow("Uninstall all", "Remove claude, codex, pi", "[Cancel]")
	d.mustNotShow("[Apply]")
	d.key("y") // ignored
	d.mustShow("[Cancel]")
	if len(collectFiles(t, home)) != len(before) {
		t.Fatal("y applied the removal")
	}
	d.key("enter") // Cancel
	d.mustShow("Cancelled. No changes applied.")
	if len(collectFiles(t, home)) != len(before) {
		t.Fatal("cancelling changed the home")
	}

	d.key("u", "left", "enter")
	d.mustShow("Hive removed")
	if hosts, err := management.RegisteredHosts(management.Options{Scope: "user", Home: home, StateDir: stateDir}); err != nil || len(hosts) != 0 {
		t.Fatalf("registered hosts = %v (%v), want none", hosts, err)
	}
	if v, err := management.CurrentVoice(management.Options{Scope: "user", Home: home, StateDir: stateDir}); err != nil || v != nil {
		t.Fatalf("voice = %v (%v), want off", v, err)
	}
	for _, r := range hostRows(d) {
		if r.checked {
			t.Fatalf("a row is still checked: %+v", r)
		}
	}
	assertTwin(t, home, stateDir, twinHome, twinState)
}

// TestHostsViewWriteBlocksKeysAndShowsProgress covers AC2 and AC8: while the
// apply runs, the view shows a progress indicator and Esc, Backspace, Enter and
// Ctrl-C do nothing; when it finishes, the view shows the result.
func TestHostsViewWriteBlocksKeysAndShowsProgress(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	m, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "claude")
	d.key("a")
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the model does not mark the apply as a write")
	}
	d.mustShow("Applying")
	before := d.screen()
	d.key("esc", "backspace", "enter", "ctrl+c", "n", "left", "a", "u")
	if d.quit {
		t.Fatal("Ctrl-C quit during a write")
	}
	if after := d.screen(); after != before {
		t.Fatalf("keys changed the screen during a write:\n%s\n--- was ---\n%s", after, before)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude")); err == nil {
		t.Fatal("the apply ran before it was released")
	}
	d.hold = false
	d.release()
	if m.isWriting() {
		t.Fatal("still writing after the result")
	}
	d.mustShow("Hive installed and verified")
	d.mustNotShow("Applying")
	d.key("esc")
	d.mustShow("> CLIs")
}

// summaryAdapter offers no capability but reports many wrapped effect lines,
// so the install summary is longer than the screen.
type summaryAdapter struct{ effects int }

func (a summaryAdapter) Detect(management.Options) ([]providerOffer, error) {
	return []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture", ManualOnly: true}}, nil
}

func (a summaryAdapter) Plan(_ management.Plan, requests []providerRequest) (onboardingPreview, error) {
	if len(requests) == 0 {
		return onboardingPreview{}, nil
	}
	var effects []string
	for i := 1; i <= a.effects; i++ {
		effects = append(effects, fmt.Sprintf("effect %02d changes something the operator should read", i))
	}
	return onboardingPreview{
		Steps:   []management.ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}},
		Details: []providerDetail{{ID: requests[0].ID, Source: "official fixture", Effects: effects}},
	}, nil
}
func (summaryAdapter) Runner() management.ExternalRunner { return testExternalRunner{} }

// TestHostsViewLongSummaryScrolls covers AC8: a summary longer than the screen
// scrolls with the arrow and page keys inside the view.
func TestHostsViewLongSummaryScrolls(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return summaryAdapter{effects: 40}, nil }
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex")
	d.key("a")
	d.mustShow("Optional capabilities")
	d.key("space", "enter") // manual capability: no version field
	d.mustShow("Install codex", "lines 1-")
	d.mustNotShow("effect 40")
	d.key("down")
	d.mustShow("lines 2-")
	d.key("pgdown", "pgdown", "pgdown")
	d.mustShow("effect 40")
	d.mustNotShow("Install / update")
	assertFits(t, d, 80, 24)
}

// TestHostsViewFitsLongRows covers AC9 for this view: six CLIs with long names,
// releases and versions fit 80x24 and 120x40, long fields end with an
// ellipsis, and the cursor reaches the last row.
func TestHostsViewFitsLongRows(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			home, stateDir := newHostsTestHome(t)
			m, d := newTestApp(t, hostsAppConfig(t, home, stateDir, minimalTestSource(t), hostsTestDeps(coreOnlyAdapterFactory)), width, height)
			d.key("enter")
			v, ok := m.top().(*hostsView)
			if !ok {
				t.Fatalf("top view is %T", m.top())
			}
			var rows []hostRow
			for i := 1; i <= 6; i++ {
				rows = append(rows, hostRow{
					Name:       fmt.Sprintf("cli-with-a-very-long-name-number-%d", i),
					State:      "legacy install",
					Registered: i%2 == 0,
					Release:    "0123456789abcdef0123456789abcdef",
					Version:    "1.2.3-rc.1+build.abcdefghijklmnopqrstuvwxyz0123456789.abcdefghijklmnopqrstuvwxyz",
					Drift:      "12",
				})
			}
			d.send(hostsLoadedMsg{seq: v.seq, rows: rows})
			assertFits(t, d, width, height)
			if !strings.Contains(d.screen(), "…") {
				t.Fatalf("long fields were cut without an ellipsis:\n%s", d.screen())
			}
			parsed := hostRows(d)
			if len(parsed) != 6 {
				t.Fatalf("%d rows shown, want 6:\n%s", len(parsed), d.screen())
			}
			for range 5 {
				d.key("down")
			}
			assertFits(t, d, width, height)
			if last := hostRows(d)[5]; !last.cursor {
				t.Fatalf("the cursor did not reach the last row:\n%s", d.screen())
			}
			for _, line := range d.lines() {
				if hostRowPattern.MatchString(line) && !strings.HasSuffix(strings.TrimRight(line, " "), "12") {
					t.Fatalf("the drift column was cut off: %q", line)
				}
			}
		})
	}
}

// TestHostsViewRealProgramRemovesHostAndIgnoresInterrupt covers AC2 and AC8
// with the real program: a removal applied through the view finishes although
// an interrupt and Ctrl-C arrive during the write; with -race it covers the
// real Cmd goroutine.
func TestHostsViewRealProgramRemovesHostAndIgnoresInterrupt(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude", "y\n", deps)

	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)
	old := removeApply
	t.Cleanup(func() { removeApply = old })
	removeApply = func(p management.Plan) (string, error) {
		<-gate
		return (management.Engine{}).Apply(p)
	}

	write, out, p, exited := startProgram(t, hostsAppConfig(t, home, stateDir, source, deps))
	seen := func(s string) func() bool {
		return func() bool { return strings.Contains(stripANSI(out.String()), s) }
	}
	waitFor(t, "the menu", seen("Main menu"))
	write("\r") // CLIs
	waitFor(t, "the CLI rows", seen("[x] claude"))
	write(" ") // uncheck claude
	write("a")
	waitFor(t, "the removal summary", seen("Files to remove"))
	write("\r")
	waitFor(t, "the apply to start", seen("Applying"))
	p.Send(tea.InterruptMsg{})
	write("\x03")
	select {
	case err := <-exited:
		t.Fatalf("the program exited during a write: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	release()
	waitFor(t, "the result", seen("Hive removed"))
	write("\x03")
	if err := waitExit(t, exited); err != nil {
		t.Fatalf("Ctrl-C after the write ended the program with %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); !os.IsNotExist(err) {
		t.Fatalf("the removal did not run: %v", err)
	}
}

// TestHostsViewBackOnIntermediateNoticeStopsAfterStepOne covers the "Step 1 of
// 2 done" notice: Enter goes on to step 2, while Esc and Backspace stop after
// step 1, leaving the removal applied and saying so.
func TestHostsViewBackOnIntermediateNoticeStopsAfterStepOne(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	setUp := func(t *testing.T) (home, stateDir string) {
		home, stateDir = newHostsTestHome(t)
		installViaText(t, home, stateDir, source, "codex", "y\n", deps)
		return home, stateDir
	}
	twinHome, twinState := setUp(t)
	removeViaCommand(t, twinHome, twinState, "codex")

	for _, back := range []string{"esc", "backspace"} {
		t.Run(back, func(t *testing.T) {
			home, stateDir := setUp(t)
			_, d := openHostsApp(t, home, stateDir, source, deps)
			toggle(t, d, "codex", "pi")
			d.key("a", "enter")
			d.mustShow("Step 1 of 2 done", "[Continue]")
			d.key(back)
			d.mustShow("Removed: codex. Install cancelled; no further changes.")
			d.mustNotShow("Step 2 of 2", "Step 1 of 2 done", "[Apply]", "Planning")
			assertTwin(t, home, stateDir, twinHome, twinState)
		})
	}
}

// TestHostsViewRequiredHostsNoticeBackCancelsWithoutChanges covers the
// required-CLIs notice: Esc and Backspace cancel the install and change
// nothing.
func TestHostsViewRequiredHostsNoticeBackCancelsWithoutChanges(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.RequiredHosts = func(management.Options) ([]string, error) { return []string{"claude", "grok"}, nil }
	for _, back := range []string{"esc", "backspace"} {
		t.Run(back, func(t *testing.T) {
			home, stateDir := newHostsTestHome(t)
			_, d := openHostsApp(t, home, stateDir, source, deps)
			toggle(t, d, "grok")
			d.key("a")
			d.mustShow("Shared resources require selecting: claude", "[Accept]")
			d.key(back)
			d.mustShow("Cancelled. No changes applied.")
			d.mustNotShow("[Accept]", "[Apply]")
			if entries, _ := os.ReadDir(home); len(entries) != 0 {
				t.Fatalf("a cancelled expansion wrote to the home: %v", entries)
			}
		})
	}
}

// TestHostsViewPendingNoteNeverCutsTheMessage covers the CLIs view's layout: a
// long result message and the pending-operation note together fit the screen,
// anything cut ends in an ellipsis, and the note stays visible.
func TestHostsViewPendingNoteNeverCutsTheMessage(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	m, d := openHostsApp(t, home, stateDir, minimalTestSource(t), hostsTestDeps(coreOnlyAdapterFactory))
	v, ok := m.top().(*hostsView)
	if !ok {
		t.Fatalf("top view is %T", m.top())
	}
	var long []string
	for i := 1; i <= 30; i++ {
		long = append(long, fmt.Sprintf("line %02d of a long result message", i))
	}
	// A long state directory makes the note wrap onto several lines.
	v.cfg.Options.StateDir = "/state/" + strings.Repeat("deeply/nested/", 6) + "dir"
	note := "An interrupted operation is still pending. Run hive recover --state-dir " + v.cfg.Options.StateDir + "."
	v.message = strings.Join(long, "\n")
	v.stillPending = true
	assertFits(t, d, 80, 24)
	mustShowFlat(d, note)
	d.mustShow("…")
	// A short message is shown whole, with the note below it.
	v.message = "Cancelled. No changes applied."
	d.mustShow("Cancelled. No changes applied.")
	mustShowFlat(d, note)
	d.mustNotShow("…")
	assertFits(t, d, 80, 24)
}

// TestHostsViewLongResultIsReachable covers a result too long for the view: a
// partial installation ends with the line naming `hive recover`, which must
// not be lost to an ellipsis. The view says how to read it in full, and the
// full text scrolls in its own view.
func TestHostsViewLongResultIsReachable(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	m, d := openHostsApp(t, home, stateDir, minimalTestSource(t), hostsTestDeps(coreOnlyAdapterFactory))
	v := m.top().(*hostsView)
	recoverLine := "Run hive recover --state-dir /synthetic/state to check it."
	lines := []string{"Partial installation (0123456789abcdef).", "The core was installed; optional capabilities pending:"}
	for i := 1; i <= 8; i++ {
		lines = append(lines,
			fmt.Sprintf("  provider-%d: manual", i),
			"    Reason: no native recipe is available in this build, so the installation is manual only.",
			"    Next action: Complete the installation following the official instructions, then run hive recover to confirm it.")
	}
	lines = append(lines, "optional capabilities incomplete (0123456789abcdef)", recoverLine)
	v.message, v.messageErr = strings.Join(lines, "\n"), true
	assertFits(t, d, 80, 24)
	d.mustShow("press m to read all")
	help := d.lines()[len(d.lines())-1]
	if !strings.Contains(help, "m more") || !strings.Contains(help, "u uninstall all") || strings.Contains(help, "…") {
		t.Fatalf("the help bar of a cut message must fit 80 columns with m more and u uninstall all: %q", help)
	}
	d.key("m")
	d.mustShow("Error", "Partial installation") // a failure outcome, until it says otherwise
	for i := 0; i < 12 && !strings.Contains(strings.Join(strings.Fields(d.screen()), ""), strings.Join(strings.Fields(recoverLine), "")); i++ {
		d.key("pgdown")
	}
	mustShowFlat(d, recoverLine)
	assertFits(t, d, 80, 24)
	d.key("esc")
	d.mustShow("CLIs", "press m to read all")

	// A message that fits shows in full, with no hint and no `m more`.
	v.message = "Hive removed (x)."
	d.mustNotShow("press m")
	mustShowFlat(d, "Hive removed (x).")
	help = d.lines()[len(d.lines())-1]
	if strings.Contains(help, "m more") || !strings.Contains(help, "u uninstall all") || strings.Contains(help, "…") {
		t.Fatalf("the help bar of a whole message: %q", help)
	}
}

// TestHostsViewShowsPendingChangesAndEnterReviewsThem covers the pending
// changes: each changed row says what will happen, the view counts them, and
// Enter on the list reviews them like `a` does (`a` stays the documented key).
func TestHostsViewShowsPendingChangesAndEnterReviewsThem(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude,codex", "y\n", deps)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	d.mustNotShow("→", "pending change")

	toggle(t, d, "claude", "pi")
	rowLine := func(name string) string {
		for _, l := range d.lines() {
			if m := hostRowPattern.FindStringSubmatch(l); m != nil && strings.HasPrefix(m[3], name) {
				return l
			}
		}
		t.Fatalf("no row for %s:\n%s", name, d.screen())
		return ""
	}
	if l := rowLine("claude"); !strings.Contains(l, "→ remove") {
		t.Errorf("claude row does not say it will be removed: %q", l)
	}
	if l := rowLine("pi"); !strings.Contains(l, "→ install") {
		t.Errorf("pi row does not say it will be installed: %q", l)
	}
	if l := rowLine("codex"); strings.Contains(l, "→") {
		t.Errorf("an unchanged row shows a change: %q", l)
	}
	d.mustShow("2 pending changes")
	assertFits(t, d, 80, 24)

	toggle(t, d, "pi")
	d.mustShow("1 pending change")
	d.mustNotShow("2 pending changes", "→ install")

	d.key("enter") // reviews the pending change like `a`
	d.mustShow("Remove claude", "[Apply]")
	d.key("esc")
	d.mustShow("Cancelled. No changes applied.")
	d.mustNotShow("pending change") // the boxes reloaded from the state

	d.key("enter") // nothing pending: the same answer as `a`
	d.mustShow("Nothing to apply")
}

// TestHostsViewEmptyStateListsOnlyUsefulKeys covers the help bar of the CLIs
// view with nothing to act on.
func TestHostsViewEmptyStateListsOnlyUsefulKeys(t *testing.T) {
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, minimalTestSource(t), defaultInstallDependencies(coreOnlyAdapterFactory))
	d.mustShow("No CLI hosts were detected or registered")
	help := d.lines()[len(d.lines())-1]
	for _, useless := range []string{"space", "apply", "uninstall", "move"} {
		if strings.Contains(help, useless) {
			t.Fatalf("the help bar lists %q with no rows: %q", useless, help)
		}
	}
	if !strings.Contains(help, "esc") {
		t.Fatalf("the help bar lost esc: %q", help)
	}
}

// TestHostsViewPartialInstallResultIsTitledResult covers the read-in-full view
// of a partial installation, where the core installed and an optional
// capability is pending: it is a Result, not an Error.
func TestHostsViewPartialInstallResultIsTitledResult(t *testing.T) {
	source := minimalTestSource(t)
	adapter := manualStepAdapter{}
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	deps.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	home, stateDir := newHostsTestHome(t)
	_, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "codex")
	d.key("a", "space", "enter") // choose the manual capability, on to the summary
	d.key("enter")
	d.mustShow("Partial installation")
	d.key("m")
	d.mustShow("Result", "Partial installation")
	d.mustNotShow("Error")
	d.key("esc")

	// A real failure keeps its Error title.
	_, d = openHostsApp(t, home, stateDir, t.TempDir(), hostsTestDeps(coreOnlyAdapterFactory))
	toggle(t, d, "claude")
	d.key("a")
	d.mustShow("Run hive from a Hive checkout")
	d.key("m")
	d.mustShow("Error")
}

// manualStepAdapter offers one manual-only capability whose step stays manual,
// so an install with it ends partial: the core installs, the capability is
// left for the operator.
type manualStepAdapter struct{}

func (manualStepAdapter) Detect(management.Options) ([]providerOffer, error) {
	return []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture", ManualOnly: true}}, nil
}

func (manualStepAdapter) Plan(_ management.Plan, requests []providerRequest) (onboardingPreview, error) {
	if len(requests) == 0 {
		return onboardingPreview{}, nil
	}
	payload, err := json.Marshal(providers.Step{ID: "context7-manual", Provider: providers.Context7, Status: providers.Manual, ManualReason: "fixture"})
	if err != nil {
		return onboardingPreview{}, err
	}
	return onboardingPreview{
		Steps:   []management.ExternalStep{{ID: "context7-manual", Payload: payload}},
		Details: []providerDetail{{ID: requests[0].ID, Source: "official fixture"}},
	}, nil
}

func (manualStepAdapter) Runner() management.ExternalRunner { return manualRunner{} }

// manualRunner runs every step and reports it as needing a person, which makes
// the onboarding end partial.
type manualRunner struct{}

func (manualRunner) Validate(management.ExternalStep) error { return nil }
func (manualRunner) Execute(management.ExternalStep) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (manualRunner) Reconcile(management.ExternalStep, json.RawMessage) (string, error) {
	return management.StepManual, nil
}

// TestHostsViewKeepsPendingMarksAfterAReadOnlyView covers the checkboxes: they
// go back to the registered state only after a write or a recovery, not when a
// read-only view (the full message) closes.
func TestHostsViewKeepsPendingMarksAfterAReadOnlyView(t *testing.T) {
	source := minimalTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, "claude,codex", "y\n", deps)
	m, d := openHostsApp(t, home, stateDir, source, deps)
	toggle(t, d, "claude", "pi")
	d.mustShow("2 pending changes")
	m.top().(*hostsView).message = "Hive removed (x)."
	d.key("m") // read the message
	d.mustShow("Result")
	d.key("esc")
	d.mustShow("2 pending changes", "→ remove", "→ install")
	if r := rowFor(t, d, "claude"); r.checked {
		t.Fatalf("claude's mark was dropped: %+v", r)
	}
	if r := rowFor(t, d, "pi"); !r.checked {
		t.Fatalf("pi's mark was dropped: %+v", r)
	}
	// After a write the marks do reset, as before (covered by the apply tests).
}
