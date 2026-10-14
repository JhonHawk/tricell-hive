package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// abFake substitutes every probe the agent-browser capability makes: PATH,
// platform, environment, file existence, Node, npm and the command runner. No
// test using it runs brew, npm, node or agent-browser.
type abFake struct {
	goos, goarch string
	path         map[string]bool
	env          map[string]string
	exists       map[string]bool
	home         string
	node         string
	prefix       string
	unwritable   bool

	nodeCalls, npmCalls int
	execs               [][]string
	onExec              func(f *abFake, argv []string) agentBrowserResult
}

const macChrome = "/Applications/Google Chrome.app"

func useAgentBrowserFake(t *testing.T, f *abFake) {
	t.Helper()
	if f.path == nil {
		f.path = map[string]bool{}
	}
	if f.exists == nil {
		f.exists = map[string]bool{}
	}
	if f.home == "" {
		f.home = "/fake/home"
	}
	savedSys, savedExec := agentBrowserSys, agentBrowserExec
	t.Cleanup(func() { agentBrowserSys, agentBrowserExec = savedSys, savedExec })
	agentBrowserSys = agentBrowserSystem{
		goos:   f.goos,
		goarch: f.goarch,
		lookPath: func(name string) (string, error) {
			if f.path[name] {
				return "/fake/bin/" + name, nil
			}
			return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
		},
		getenv:            func(k string) string { return f.env[k] },
		userHome:          func() (string, error) { return f.home, nil },
		pathExists:        func(p string) bool { return f.exists[p] },
		isFile:            func(p string) bool { return f.exists[p] },
		dirHasEntries:     func(p string) bool { return f.exists[p] },
		browserCandidates: realBrowserCandidates,
		nodeVersion:       func() (string, error) { f.nodeCalls++; return f.node, nil },
		npmPrefix:         func() (string, error) { f.npmCalls++; return f.prefix, nil },
		prefixWritable:    func(string) bool { return !f.unwritable },
	}
	agentBrowserExec = func(argv []string) agentBrowserResult {
		f.execs = append(f.execs, append([]string(nil), argv...))
		if f.onExec != nil {
			return f.onExec(f, argv)
		}
		return agentBrowserResult{}
	}
}

func macBrewFake() *abFake {
	return &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"brew": true}}
}

func linuxNodeFake() *abFake {
	return &abFake{goos: "linux", goarch: "amd64", path: map[string]bool{"npm": true, "node": true}, node: "v24.1.0", prefix: "/home/u/.npm-global"}
}

// installingExec simulates successful commands: brew/npm put agent-browser on
// PATH and agent-browser install leaves a browser.
func installingExec(f *abFake, argv []string) agentBrowserResult {
	switch argv[0] {
	case "brew", "npm":
		f.path["agent-browser"] = true
	case "agent-browser":
		f.exists[macChrome] = true
		f.path["chromium"] = true
	}
	return agentBrowserResult{}
}

func newTestAdapter(t *testing.T, synthetic, dry, interactive bool) *nativeProviderAdapter {
	t.Helper()
	return newNativeProviderAdapter(synthetic, dry, interactive)
}

func agentBrowserOffer(t *testing.T, offers []providerOffer) (providerOffer, bool) {
	t.Helper()
	for _, o := range offers {
		if o.ID == "agent-browser" {
			return o, true
		}
	}
	return providerOffer{}, false
}

func detectAB(t *testing.T, a *nativeProviderAdapter) (providerOffer, bool) {
	t.Helper()
	offers, err := a.Detect(management.Options{Hosts: []string{"claude"}})
	if err != nil {
		t.Fatal(err)
	}
	return agentBrowserOffer(t, offers)
}

func planAB(t *testing.T, a *nativeProviderAdapter) onboardingPreview {
	t.Helper()
	preview, err := a.Plan(management.Plan{}, []providerRequest{{ID: "agent-browser"}})
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func stepIDs(preview onboardingPreview) []string {
	var ids []string
	for _, s := range preview.Steps {
		ids = append(ids, s.ID)
	}
	return ids
}

func stepArgv(t *testing.T, s management.ExternalStep) []string {
	t.Helper()
	var step providers.Step
	if err := json.Unmarshal(s.Payload, &step); err != nil {
		t.Fatal(err)
	}
	return step.Argv
}

func wizardFactory(onboardingInput) (onboardingAdapter, error) {
	return newNativeProviderAdapter(false, false, true), nil
}

func runAgentBrowserWizard(t *testing.T, input string) (out string, err error, stateDir, home string) {
	t.Helper()
	home = t.TempDir()
	stateDir = filepath.Join(home, "state")
	args := []string{"--home", home, "--hosts", "codex", "--source", minimalTestSource(t), "--state-dir", stateDir}
	var buf bytes.Buffer
	err = installWithDependencies(args, strings.NewReader(input), &buf, true, defaultInstallDependencies(wizardFactory))
	return buf.String(), err, stateDir, home
}

// (a) macOS arm64 with brew and without agent-browser: an unversioned offer
// whose summary names the command, the latest version and the Node warning.
func TestAgentBrowserOfferOnMacWithBrewIsUnversioned(t *testing.T) {
	f := macBrewFake()
	f.onExec = installingExec
	useAgentBrowserFake(t, f)
	out, err, stateDir, home := runAgentBrowserWizard(t, "3\ny\n")
	if err != nil {
		t.Fatalf("install failed: %v\n%s", err, out)
	}
	for _, want := range []string{"brew install agent-browser", "latest version", "Node", "agent-browser install"} {
		if !strings.Contains(out, want) {
			t.Fatalf("summary lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Exact version for") {
		t.Fatalf("asked for a version although the offer is unversioned:\n%s", out)
	}
	if got := len(f.execs); got != 2 || strings.Join(f.execs[0], " ") != "brew install agent-browser" || strings.Join(f.execs[1], " ") != "agent-browser install" {
		t.Fatalf("executed %v", f.execs)
	}
	result := readOnboardingReceipt(t, stateDir)
	for _, id := range []string{"agent-browser-install", "agent-browser-chrome"} {
		if got := stepStatus(t, result, id); got != "verified" {
			t.Fatalf("%s = %s", id, got)
		}
	}
	if _, statErr := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); statErr != nil {
		t.Fatalf("core not installed: %v", statErr)
	}
}

// (b) Linux amd64 with Node 24 and a writable prefix: npm.
func TestAgentBrowserOfferOnLinuxWithNode24UsesNpm(t *testing.T) {
	f := linuxNodeFake()
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	offer, ok := detectAB(t, a)
	if !ok || offer.ManualOnly || !offer.Unversioned {
		t.Fatalf("offer = %+v, ok=%v", offer, ok)
	}
	preview := planAB(t, a)
	if got := strings.Join(stepIDs(preview), ","); got != "agent-browser-install,agent-browser-chrome" {
		t.Fatalf("steps = %s", got)
	}
	if got := strings.Join(stepArgv(t, preview.Steps[0]), " "); got != "npm install -g agent-browser" {
		t.Fatalf("argv = %s", got)
	}
	effects := strings.Join(preview.Details[0].Effects, "\n")
	if !strings.Contains(effects, "npm install -g agent-browser") || !strings.Contains(effects, "latest version") {
		t.Fatalf("effects = %s", effects)
	}
	if f.nodeCalls != 1 || f.npmCalls != 1 {
		t.Fatalf("node/npm probes = %d/%d", f.nodeCalls, f.npmCalls)
	}
}

// (c) A browser anywhere D-5 looks removes the Chrome step.
func TestAgentBrowserChromeStepOnlyWithoutABrowser(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(f *abFake)
		chrome  bool
	}{
		{"none", func(f *abFake) {}, true},
		{"mac app", func(f *abFake) { f.exists[macChrome] = true }, false},
		{"env path to a file", func(f *abFake) {
			f.env = map[string]string{"AGENT_BROWSER_EXECUTABLE_PATH": "/opt/chrome/chrome"}
			f.exists["/opt/chrome/chrome"] = true
		}, false},
		{"env path to nothing", func(f *abFake) {
			f.env = map[string]string{"AGENT_BROWSER_EXECUTABLE_PATH": "/opt/missing"}
		}, true},
		{"downloaded browsers", func(f *abFake) { f.exists["/fake/home/.agent-browser/browsers"] = true }, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := macBrewFake()
			f.exists = map[string]bool{}
			c.prepare(f)
			useAgentBrowserFake(t, f)
			a := newTestAdapter(t, false, false, true)
			if _, ok := detectAB(t, a); !ok {
				t.Fatal("no offer")
			}
			ids := strings.Join(stepIDs(planAB(t, a)), ",")
			want := "agent-browser-install"
			if c.chrome {
				want += ",agent-browser-chrome"
			}
			if ids != want {
				t.Fatalf("steps = %s, want %s", ids, want)
			}
		})
	}
	t.Run("linux chromium on PATH", func(t *testing.T) {
		f := linuxNodeFake()
		f.path["chromium"] = true
		useAgentBrowserFake(t, f)
		a := newTestAdapter(t, false, false, true)
		detectAB(t, a)
		if ids := strings.Join(stepIDs(planAB(t, a)), ","); ids != "agent-browser-install" {
			t.Fatalf("steps = %s", ids)
		}
	})
}

// (d) Every missing requirement leaves the capability manual with its reason.
func TestAgentBrowserManualWhenRequirementMissing(t *testing.T) {
	cases := []struct {
		name   string
		fake   func() *abFake
		reason string
		cmd    string
	}{
		{"no brew", func() *abFake { f := macBrewFake(); f.path = map[string]bool{}; return f }, "Homebrew", "brew install agent-browser"},
		{"node 22", func() *abFake { f := linuxNodeFake(); f.node = "v22.5.0"; return f }, "22", "npm install -g agent-browser"},
		{"no npm", func() *abFake { f := linuxNodeFake(); f.path = map[string]bool{}; return f }, "Node.js", "npm install -g agent-browser"},
		{"unwritable prefix", func() *abFake { f := linuxNodeFake(); f.unwritable = true; return f }, "/home/u/.npm-global", "npm install -g agent-browser"},
		{"linux arm64", func() *abFake { f := linuxNodeFake(); f.goarch = "arm64"; return f }, "linux/arm64", "npm install -g agent-browser"},
		{"mac amd64", func() *abFake { f := macBrewFake(); f.goarch = "amd64"; return f }, "darwin/amd64", "brew install agent-browser"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := c.fake()
			useAgentBrowserFake(t, f)
			a := newTestAdapter(t, false, false, true)
			offer, ok := detectAB(t, a)
			if !ok || !offer.ManualOnly {
				t.Fatalf("offer = %+v ok=%v, want a manual offer", offer, ok)
			}
			preview := planAB(t, a)
			if got := strings.Join(stepIDs(preview), ","); got != "agent-browser-manual" {
				t.Fatalf("steps = %s", got)
			}
			var step providers.Step
			if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
				t.Fatal(err)
			}
			if step.Status != providers.Manual || !strings.Contains(step.ManualReason, c.reason) || !strings.Contains(step.ManualReason, c.cmd) {
				t.Fatalf("reason = %q, want %q and %q", step.ManualReason, c.reason, c.cmd)
			}
			if err := providers.ValidateStep(preview.Steps[0].Payload); err != nil {
				t.Fatal(err)
			}
			if len(f.execs) != 0 {
				t.Fatalf("executed %v", f.execs)
			}
		})
	}
}

// agent-browser already installed but no browser, on an untested platform:
// manual with the reason and the official command, nothing executed.
func TestAgentBrowserPresentWithoutBrowserOnUntestedPlatformIsManual(t *testing.T) {
	for _, pf := range [][2]string{{"linux", "arm64"}, {"darwin", "amd64"}} {
		t.Run(pf[0]+"/"+pf[1], func(t *testing.T) {
			f := &abFake{goos: pf[0], goarch: pf[1], path: map[string]bool{"agent-browser": true}}
			useAgentBrowserFake(t, f)
			a := newTestAdapter(t, false, false, true)
			offer, ok := detectAB(t, a)
			if !ok || !offer.ManualOnly {
				t.Fatalf("offer = %+v ok=%v, want manual", offer, ok)
			}
			preview := planAB(t, a)
			if got := strings.Join(stepIDs(preview), ","); got != "agent-browser-manual" {
				t.Fatalf("steps = %s", got)
			}
			var step providers.Step
			if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(step.ManualReason, pf[0]+"/"+pf[1]) || !strings.Contains(step.ManualReason, "agent-browser install") {
				t.Fatalf("reason = %q", step.ManualReason)
			}
			if len(f.execs) != 0 {
				t.Fatalf("executed %v", f.execs)
			}
		})
	}
}

// (e) agent-browser present without a browser: only the Chrome step; with a
// browser: no offer at all.
func TestAgentBrowserPresentOffersOnlyTheMissingChromeStep(t *testing.T) {
	f := macBrewFake()
	f.path["agent-browser"] = true
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	offer, ok := detectAB(t, a)
	if !ok || offer.ManualOnly || !offer.Unversioned {
		t.Fatalf("offer = %+v ok=%v", offer, ok)
	}
	preview := planAB(t, a)
	if got := strings.Join(stepIDs(preview), ","); got != "agent-browser-chrome" {
		t.Fatalf("steps = %s", got)
	}
	f.exists[macChrome] = true
	if _, ok := detectAB(t, newTestAdapter(t, false, false, true)); ok {
		t.Fatal("offered although agent-browser and a browser are present")
	}
}

func partialOutput(t *testing.T, f *abFake) (string, management.OnboardingResult) {
	t.Helper()
	useAgentBrowserFake(t, f)
	out, err, stateDir, home := runAgentBrowserWizard(t, "3\ny\n")
	if err == nil {
		t.Fatalf("expected a partial outcome:\n%s", out)
	}
	if _, statErr := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); statErr != nil {
		t.Fatalf("core not retained: %v", statErr)
	}
	return out, readOnboardingReceipt(t, stateDir)
}

// (f) A failing first step: the Chrome step is skipped and not executed, and
// the failure message is visible in the result detail.
func TestAgentBrowserFailedInstallSkipsChromeAndShowsMessage(t *testing.T) {
	f := macBrewFake()
	f.onExec = func(f *abFake, argv []string) agentBrowserResult {
		return agentBrowserResult{ExitCode: 1, Output: "Error: formula exploded\nfatal: kaboom\n"}
	}
	out, result := partialOutput(t, f)
	if got := stepStatus(t, result, "agent-browser-install"); got != "failed" {
		t.Fatalf("install = %s", got)
	}
	if got := stepStatus(t, result, "agent-browser-chrome"); got != "skipped" {
		t.Fatalf("chrome = %s", got)
	}
	if len(f.execs) != 1 {
		t.Fatalf("executed %v; the Chrome step must not run", f.execs)
	}
	for _, want := range []string{"agent-browser-install: failed", "kaboom", "agent-browser-chrome: skipped"} {
		if !strings.Contains(out, want) {
			t.Fatalf("detail lacks %q:\n%s", want, out)
		}
	}
	for _, line := range strings.Split(out[strings.Index(out, "Partial installation"):], "\n") {
		if len(line) > 80 {
			t.Fatalf("line exceeds 80 columns: %q", line)
		}
	}
}

func TestAgentBrowserTimeoutIsAFailedStepWithMessage(t *testing.T) {
	f := macBrewFake()
	f.onExec = func(f *abFake, argv []string) agentBrowserResult {
		return agentBrowserResult{ExitCode: -1, TimedOut: true}
	}
	out, result := partialOutput(t, f)
	if got := stepStatus(t, result, "agent-browser-install"); got != "failed" {
		t.Fatalf("install = %s", got)
	}
	if !strings.Contains(out, "timed out") {
		t.Fatalf("detail lacks the timeout:\n%s", out)
	}
}

func TestAgentBrowserNpmPermissionFailureHasASpecificMessage(t *testing.T) {
	f := linuxNodeFake()
	useAgentBrowserFake(t, f)
	f.onExec = func(f *abFake, argv []string) agentBrowserResult {
		return agentBrowserResult{ExitCode: 243, Output: "npm error code EACCES\nnpm error syscall mkdir\n"}
	}
	out, err, _, _ := runAgentBrowserWizard(t, "3\ny\n")
	if err == nil || !strings.Contains(out, "EACCES") || !strings.Contains(out, "sudo") {
		t.Fatalf("no specific EACCES message (err=%v):\n%s", err, out)
	}
}

func TestAgentBrowserInstalledButNotOnPathIsFailedWithMessage(t *testing.T) {
	f := macBrewFake()
	f.onExec = func(f *abFake, argv []string) agentBrowserResult { return agentBrowserResult{} }
	out, result := partialOutput(t, f)
	if got := stepStatus(t, result, "agent-browser-install"); got != "failed" {
		t.Fatalf("install = %s", got)
	}
	if !strings.Contains(out, "not on PATH") {
		t.Fatalf("detail lacks the PATH message:\n%s", out)
	}
}

// (h) An interruption with no saved result recovers to failed, never unknown.
func TestAgentBrowserInterruptedStepRecoversAsFailed(t *testing.T) {
	f := macBrewFake()
	f.onExec = installingExec
	useAgentBrowserFake(t, f)
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: minimalTestSource(t), Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	a := newTestAdapter(t, false, false, true)
	detectAB(t, a)
	preview := planAB(t, a)
	crash := &executeThenCrashRunner{inner: a.Runner()}
	done := make(chan struct{})
	go func() {
		defer close(done)
		management.Engine{}.Onboard(p, preview.Steps, crash)
	}()
	<-done
	if !crash.executed {
		t.Fatal("never reached Execute")
	}
	// The interrupted process left nothing on PATH.
	delete(f.path, "agent-browser")
	f.execs = nil
	var out bytes.Buffer
	args := []string{"--home", home, "--hosts", "codex", "--source", minimalTestSource(t), "--state-dir", stateDir}
	if err := installWithDependencies(args, strings.NewReader("y\n"), &out, true, defaultInstallDependencies(wizardFactory)); err != nil {
		t.Fatalf("recovery: %v\n%s", err, out.String())
	}
	if len(f.execs) != 0 {
		t.Fatalf("recovery executed %v", f.execs)
	}
	result := readOnboardingReceipt(t, stateDir)
	for _, s := range result.Steps {
		if s.Status == management.StepUnknown {
			t.Fatalf("%s left unknown", s.Step.ID)
		}
	}
	if got := result.Steps[0].Status; got != management.StepFailed {
		t.Fatalf("interrupted step = %s, want failed", got)
	}
	if !strings.Contains(out.String(), "agent-browser-install: failed") || !strings.Contains(out.String(), "interrupted before its result was recorded") {
		t.Fatalf("recovery output does not show the interruption:\n%s", out.String())
	}
	if pending, _ := management.OnboardingPending(stateDir); pending {
		t.Fatal("journal still pending")
	}
}

// (i) No program runs in a non-interactive session, a dry run or under a
// synthetic home: not in Detect and not in the runner.
func TestAgentBrowserSafeModesRunNothing(t *testing.T) {
	modes := []struct {
		name                        string
		synthetic, dry, interactive bool
	}{
		{"non-interactive", false, false, false},
		{"dry run", false, true, true},
		{"synthetic home", true, false, true},
	}
	for _, m := range modes {
		t.Run(m.name, func(t *testing.T) {
			f := linuxNodeFake()
			useAgentBrowserFake(t, f)
			a := newTestAdapter(t, m.synthetic, m.dry, m.interactive)
			if _, err := a.Detect(management.Options{Hosts: []string{"claude"}}); err != nil {
				t.Fatal(err)
			}
			if f.nodeCalls != 0 || f.npmCalls != 0 || len(f.execs) != 0 {
				t.Fatalf("node=%d npm=%d execs=%v", f.nodeCalls, f.npmCalls, f.execs)
			}
			planAB(t, a)
			if m.synthetic {
				planned := management.ExternalStep{ID: "agent-browser-install", Payload: json.RawMessage(`{"id":"agent-browser-install","provider":"agent-browser","status":"planned","method":"npm","argv":["npm","install","-g","agent-browser"]}`)}
				res, err := a.Runner().Execute(planned)
				if err != nil {
					t.Fatal(err)
				}
				if status, _ := a.Runner().Reconcile(planned, res); status != management.StepSkipped {
					t.Fatalf("synthetic step reconciled to %s, want skipped", status)
				}
			}
			if len(f.execs) != 0 || f.nodeCalls != 0 || f.npmCalls != 0 {
				t.Fatalf("a program ran: node=%d npm=%d execs=%v", f.nodeCalls, f.npmCalls, f.execs)
			}
		})
	}
	t.Run("dry-run wizard", func(t *testing.T) {
		f := linuxNodeFake()
		useAgentBrowserFake(t, f)
		home := t.TempDir()
		args := []string{"--home", home, "--hosts", "codex", "--source", minimalTestSource(t), "--state-dir", filepath.Join(home, "state"), "--dry-run"}
		var out bytes.Buffer
		if err := installWithDependencies(args, strings.NewReader("3\n"), &out, true, defaultInstallDependencies(func(in onboardingInput) (onboardingAdapter, error) {
			return newNativeProviderAdapter(false, in.DryRun, true), nil
		})); err != nil {
			t.Fatal(err)
		}
		if f.nodeCalls != 0 || f.npmCalls != 0 || len(f.execs) != 0 {
			t.Fatalf("node=%d npm=%d execs=%v", f.nodeCalls, f.npmCalls, f.execs)
		}
	})
	t.Run("factory wires the mode", func(t *testing.T) {
		f := linuxNodeFake()
		useAgentBrowserFake(t, f)
		for _, in := range []onboardingInput{
			{Options: management.Options{Home: t.TempDir()}, Interactive: true},
			{DryRun: true, Interactive: true},
			{Interactive: false},
		} {
			adapter, err := nativeProviderAdapterFactory(in)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := adapter.Detect(management.Options{Hosts: []string{"claude"}, Home: in.Options.Home}); err != nil {
				t.Fatal(err)
			}
		}
		if f.nodeCalls != 0 || f.npmCalls != 0 {
			t.Fatalf("node=%d npm=%d", f.nodeCalls, f.npmCalls)
		}
	})
}

// (k) Output over the cap never cancels the process; the LAST 64 KiB is
// stored, so the error that ends the output is kept and shown.
func TestAgentBrowserLargeOutputIsTrimmedNotCancelled(t *testing.T) {
	buf := &tailBuffer{limit: agentBrowserMaxOutput}
	chunk := bytes.Repeat([]byte("x"), 4096)
	for i := 0; i < 50; i++ {
		n, err := buf.Write(chunk)
		if n != len(chunk) || err != nil {
			t.Fatalf("write %d = %d, %v; the writer must accept everything", i, n, err)
		}
	}
	buf.Write([]byte("\nERROR-TAIL\n"))
	if got := buf.String(); len(got) != agentBrowserMaxOutput || !strings.HasSuffix(got, "ERROR-TAIL\n") {
		t.Fatalf("kept %d bytes, tail %q", len(got), got[len(got)-20:])
	}

	// The real runner with a harmless script that prints 200 KiB then fails.
	script := filepath.Join(t.TempDir(), "noisy.sh")
	body := "#!/bin/sh\nhead -c 200000 /dev/zero | tr '\\0' y\necho\necho ERROR-TAIL\nexit 3\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	real := runBoundedCommand([]string{script}, 20*time.Second, time.Second, agentBrowserMaxOutput)
	if real.TimedOut || real.ExitCode != 3 || len(real.Output) > agentBrowserMaxOutput || !strings.HasSuffix(strings.TrimSpace(real.Output), "ERROR-TAIL") {
		t.Fatalf("real runner: exit=%d timedOut=%v len=%d", real.ExitCode, real.TimedOut, len(real.Output))
	}

	f := macBrewFake()
	f.onExec = func(f *abFake, argv []string) agentBrowserResult {
		return agentBrowserResult{ExitCode: 1, Output: strings.Repeat("y", 200<<10) + "\nERROR-TAIL\n"}
	}
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	detectAB(t, a)
	step := planAB(t, a).Steps[0]
	raw, err := a.Runner().Execute(step)
	if err != nil {
		t.Fatal(err)
	}
	var res agentBrowserResult
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Output) > agentBrowserMaxOutput || !strings.Contains(res.Output, "ERROR-TAIL") {
		t.Fatalf("stored %d bytes without the tail", len(res.Output))
	}
	_, tail := stepResultDetail(management.StepReceipt{Step: step, Status: management.StepFailed, Result: raw})
	if len(tail) == 0 || tail[len(tail)-1] != "ERROR-TAIL" {
		t.Fatalf("detail lines = %v", tail)
	}
}

// (l) The Chrome step without agent-browser on PATH is skipped, with a reason.
func TestAgentBrowserChromeStepSkippedWithoutTheProgram(t *testing.T) {
	f := macBrewFake()
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	detectAB(t, a)
	var chrome management.ExternalStep
	for _, s := range planAB(t, a).Steps {
		if s.ID == "agent-browser-chrome" {
			chrome = s
		}
	}
	raw, err := a.Runner().Execute(chrome)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.execs) != 0 {
		t.Fatalf("executed %v", f.execs)
	}
	var res agentBrowserResult
	if err := json.Unmarshal(raw, &res); err != nil || !res.Skipped || res.Message == "" {
		t.Fatalf("result = %+v, %v", res, err)
	}
	if status, _ := a.Runner().Reconcile(chrome, raw); status != management.StepSkipped {
		t.Fatalf("status = %s", status)
	}
}

func TestAgentBrowserRunnerRejectsTamperedArgv(t *testing.T) {
	f := macBrewFake()
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	tampered := management.ExternalStep{ID: "agent-browser-install", Payload: json.RawMessage(`{"id":"agent-browser-install","provider":"agent-browser","status":"planned","method":"sh","argv":["sh","-c","id"]}`)}
	if a.Runner().Validate(tampered) == nil {
		t.Fatal("tampered payload validated")
	}
	if _, err := a.Runner().Execute(tampered); err == nil {
		t.Fatal("tampered payload executed")
	}
	if len(f.execs) != 0 {
		t.Fatalf("executed %v", f.execs)
	}
}

// (j) The real runner against a program that sleeps and hands its pipes to a
// background child: the whole process group dies at the deadline, the call
// returns, stdin is empty and the Homebrew update guard is set.
func TestRunBoundedCommandKillsTheProcessGroupAtTimeout(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")
	envFile := filepath.Join(dir, "env.txt")
	stdinFile := filepath.Join(dir, "stdin.txt")
	script := filepath.Join(dir, "sleeper.sh")
	body := "#!/bin/sh\n" +
		"printf '%s' \"$HOMEBREW_NO_AUTO_UPDATE\" > " + envFile + "\n" +
		"cat > " + stdinFile + "\n" +
		"echo started\n" +
		"sleep 60 &\n" +
		"echo $! > " + pidFile + "\n" +
		"sleep 60\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	res := runBoundedCommand([]string{script}, 700*time.Millisecond, 300*time.Millisecond, agentBrowserMaxOutput)
	if elapsed := time.Since(start); elapsed > 10*time.Second {
		t.Fatalf("returned after %s", elapsed)
	}
	if !res.TimedOut || res.ExitCode == 0 {
		t.Fatalf("result = %+v", res)
	}
	if !strings.Contains(res.Output, "started") {
		t.Fatalf("output = %q", res.Output)
	}
	if env, _ := os.ReadFile(envFile); string(env) != "1" {
		t.Fatalf("HOMEBREW_NO_AUTO_UPDATE = %q", env)
	}
	if in, err := os.ReadFile(stdinFile); err != nil || len(in) != 0 {
		t.Fatalf("stdin was not empty: %q, %v", in, err)
	}
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		// A killed child is reaped by init or lingers as a zombie; signal 0 on
		// a zombie still succeeds, so also accept a dead process state.
		if err := syscall.Kill(pid, 0); err != nil || processIsZombie(pid) {
			return
		}
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("background child %d survived the timeout", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func processIsZombie(pid int) bool {
	out, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	return err != nil || strings.HasPrefix(strings.TrimSpace(string(out)), "Z")
}

func TestRunBoundedCommandReportsExitCodeAndMissingProgram(t *testing.T) {
	res := runBoundedCommand([]string{"/definitely/not/here"}, time.Second, time.Second, 1024)
	if res.ExitCode == 0 || res.Message == "" {
		t.Fatalf("result = %+v", res)
	}
}
