package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"tricell-hive/tooling/management"
)

// interfaceTestOptions builds a fresh synthetic home and state directory for
// one test, mirroring the pattern install_test.go and update_test.go use.
func interfaceTestOptions(t *testing.T) options {
	t.Helper()
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return options{Home: home, StateDir: stateDir, Source: source}
}

// TestOpenInterfaceWithoutTerminalOrAccessible covers D14-A: with neither a
// terminal nor HIVE_ACCESSIBLE, openInterface returns today's usage error,
// unchanged.
func TestOpenInterfaceWithoutTerminalOrAccessible(t *testing.T) {
	err := openInterface(false, false, strings.NewReader(""), io.Discard, options{})
	if err == nil || err.Error() != usageMessage {
		t.Fatalf("got %v, want %q", err, usageMessage)
	}
}

// TestMenuQuitExitsZero covers AC1/AC10: explicitly choosing Quit (the
// menu's seventh and last entry) in accessible mode returns nil (exit 0).
func TestMenuQuitExitsZero(t *testing.T) {
	o := interfaceTestOptions(t)
	var out bytes.Buffer
	if err := openInterface(false, true, strings.NewReader("7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
}

// TestMenuEOFExitsZero covers AC1/AC10: the end of input at the menu selects
// Quit (its own default) and returns nil (exit 0), instead of huh indexing
// a nonexistent option and panicking.
func TestMenuEOFExitsZero(t *testing.T) {
	o := interfaceTestOptions(t)
	var out bytes.Buffer
	if err := openInterface(false, true, strings.NewReader(""), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
}

// TestMenuStubReturnsToMenu covers AC10: choosing a screen that is still a
// stub in T2 prints "Not available yet." and returns to the menu instead of
// exiting, so a later Quit still ends the session cleanly. The status line,
// printed once per menu render, appearing exactly twice pins that the menu
// actually rendered again after the stub and that the later "7" (Quit)
// choice was consumed by that second render, not some other exit path.
func TestMenuStubReturnsToMenu(t *testing.T) {
	o := interfaceTestOptions(t)
	var out bytes.Buffer
	// "2" is Install CLIs (menuLabels' own order: Status, Install CLIs,
	// Remove CLIs, Update, Releases, Voice, Quit).
	if err := openInterface(false, true, strings.NewReader("2\n7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Not available yet.") {
		t.Fatalf("stub did not print its placeholder: %s", out.String())
	}
	if renders := strings.Count(out.String(), "No CLI hosts are registered."); renders != 2 {
		t.Fatalf("expected the menu to render twice (once before and once after the stub), got %d renders in: %s", renders, out.String())
	}
}

// TestMenuLabelsOrder pins design.md "La interfaz"'s fixed menu order and
// text: Status, Install CLIs, Remove CLIs, Update, Releases, Voice, Quit,
// with Quit as the last entry (selectMenuEntry's own default, and
// runMenu's own exit check, both key off menuQuit's index matching it).
func TestMenuLabelsOrder(t *testing.T) {
	want := []string{"Status", "Install CLIs", "Remove CLIs", "Update", "Releases", "Voice", "Quit"}
	if len(menuLabels) != len(want) {
		t.Fatalf("got %d menu entries %v, want %d: %v", len(menuLabels), menuLabels, len(want), want)
	}
	for i, label := range want {
		if menuLabels[i] != label {
			t.Fatalf("menuLabels[%d] = %q, want %q (full: %v)", i, menuLabels[i], label, menuLabels)
		}
	}
	if menuQuit != menuEntry(len(want)-1) {
		t.Fatalf("menuQuit = %d, want %d (the last entry)", menuQuit, len(want)-1)
	}
}

// TestSelectMenuEntryTitleIsStatusLine pins that selectMenuEntry's Select
// field is titled with the status line the caller gives it, not a fixed or
// missing title.
func TestSelectMenuEntryTitleIsStatusLine(t *testing.T) {
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("7\n"), &out)
	status := "3 CLI hosts · release abc123 · voice jarvis"
	if _, _, err := p.selectMenuEntry(status); err != nil {
		t.Fatalf("selectMenuEntry: %v", err)
	}
	if !strings.Contains(out.String(), status) {
		t.Fatalf("menu title did not include the status line: got %q, want it to contain %q", out.String(), status)
	}
}

// TestSelectMenuEntryDefaultsToQuit pins Quit as the Select field's own
// default value, distinctly from EOF-driven cancellation (TestMenuEOFExits
// Zero): a blank Enter is a real, non-cancelled choice of Quit, not a
// cancellation that merely happens to also land on Quit.
func TestSelectMenuEntryDefaultsToQuit(t *testing.T) {
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("\n"), &out)
	choice, cancelled, err := p.selectMenuEntry("status")
	if err != nil {
		t.Fatalf("selectMenuEntry: %v\noutput:\n%s", err, out.String())
	}
	if cancelled {
		t.Fatal("a blank Enter is a real default choice, not a cancellation")
	}
	if choice != menuQuit {
		t.Fatalf("selectMenuEntry(blank Enter) = %v, want menuQuit as the field's own default", choice)
	}
}

// stubInterfaceStdio substitutes interfaceStdio for the duration of a test
// (T2 fix round item 4: "inject the TTY check so the test doesn't depend on
// the real stdin"), restoring it via t.Cleanup. The returned buffer
// collects everything the interface would otherwise have printed to
// os.Stdout.
func stubInterfaceStdio(t *testing.T, isTTY bool, input string) *bytes.Buffer {
	t.Helper()
	old := interfaceStdio
	t.Cleanup(func() { interfaceStdio = old })
	var out bytes.Buffer
	interfaceStdio.isTTY = func() bool { return isTTY }
	interfaceStdio.in = strings.NewReader(input)
	interfaceStdio.out = &out
	return &out
}

// TestRunBareArgsOpensMenuWithAccessible pins that the real top-level run
// dispatcher, called exactly as bare `hive` would (run(nil)), opens the
// menu when HIVE_ACCESSIBLE=1, without depending on the test process's own
// real stdin (T2 fix round item 4).
func TestRunBareArgsOpensMenuWithAccessible(t *testing.T) {
	t.Setenv("HIVE_ACCESSIBLE", "1")
	out := stubInterfaceStdio(t, false, "7\n")
	if err := run(nil); err != nil {
		t.Fatalf("run(nil): %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Status") {
		t.Fatalf("bare hive did not open the menu: %s", out.String())
	}
}

// TestRunTuiSubcommandReachesInterface pins that the real top-level run
// dispatcher reaches the interface for the explicit `hive tui` subcommand
// too, with its own --home/--state-dir/--source, and without depending on
// the test process's own real stdin.
func TestRunTuiSubcommandReachesInterface(t *testing.T) {
	t.Setenv("HIVE_ACCESSIBLE", "1")
	o := interfaceTestOptions(t)
	out := stubInterfaceStdio(t, false, "7\n")
	args := []string{"tui", "--home", o.Home, "--state-dir", o.StateDir, "--source", o.Source}
	if err := run(args); err != nil {
		t.Fatalf("run(tui): %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Status") {
		t.Fatalf("hive tui did not reach the interface: %s", out.String())
	}
}

// TestRunFormHonorsCancelledResult is an integration-level pin (as opposed
// to TestCancelledResult's own pure-function unit test) that runForm itself
// actually applies cancelledResult's verdict to a real form's outcome: an
// immediate end of input on a plain Confirm field must be reported as
// cancelled.
func TestRunFormHonorsCancelledResult(t *testing.T) {
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader(""), &out)
	apply := false
	field := huh.NewConfirm().Title("Apply?").Value(&apply)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		t.Fatalf("runForm: %v\noutput:\n%s", err, out.String())
	}
	if !cancelled {
		t.Fatal("an immediate end of input must be reported as cancelled")
	}
}

// TestMenuStatusLineNoHosts covers the empty-state line design.md "La
// interfaz" requires when nothing is registered.
func TestMenuStatusLineNoHosts(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	line, err := interfaceStatusLine(mo)
	if err != nil {
		t.Fatal(err)
	}
	if line != "No CLI hosts are registered." {
		t.Fatalf("got %q", line)
	}
}

// TestCancelledResult is the unit test for the prompter's own mapping from
// huh's outcome to cancellation (design.md "Cancelar"): ErrUserAborted and
// the accessible end of input both cancel; any other error does not.
func TestCancelledResult(t *testing.T) {
	cases := []struct {
		name string
		err  error
		eof  bool
		want bool
	}{
		{"user aborted", huh.ErrUserAborted, false, true},
		{"eof with no error", nil, true, true},
		{"eof wrapping aborted too", huh.ErrUserAborted, true, true},
		{"unrelated error", errors.New("boom"), false, false},
		{"no error no eof", nil, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := cancelledResult(c.err, c.eof); got != c.want {
				t.Fatalf("cancelledResult(%v, %v) = %v, want %v", c.err, c.eof, got, c.want)
			}
		})
	}
}

// TestOneByteReaderLatchesEOFAndReadsSingleBytes covers the reader design.md
// "El `prompter` de `huh`" requires: every Read returns at most one byte
// (even when asked for more), and eof latches true once the source is
// exhausted.
func TestOneByteReaderLatchesEOFAndReadsSingleBytes(t *testing.T) {
	r := &oneByteReader{r: strings.NewReader("ab")}
	buf := make([]byte, 4)
	n, err := r.Read(buf)
	if err != nil || n != 1 || buf[0] != 'a' {
		t.Fatalf("first read: n=%d err=%v buf=%q", n, err, buf[:n])
	}
	if r.eof {
		t.Fatal("eof latched too early")
	}
	n, err = r.Read(buf)
	if err != nil || n != 1 || buf[0] != 'b' {
		t.Fatalf("second read: n=%d err=%v buf=%q", n, err, buf[:n])
	}
	if r.eof {
		t.Fatal("eof latched too early")
	}
	if _, err := r.Read(buf); err == nil {
		t.Fatal("expected EOF on third read")
	}
	if !r.eof {
		t.Fatal("eof was not latched")
	}
}

// TestRunFormRecoversInvalidAnswerThenEOF is the T2 fix round item 1 fixture:
// huh v2.0.3's accessible PromptInt/PromptString return an out-of-range or
// non-numeric answer's stale text, unfiltered, once real end-of-input
// follows it, and Select/MultiSelect index that text into their options
// slice with no bounds check (field_select.go). Each of these scripted
// inputs (a too-large number with no trailing newline, zero, and letters)
// reaches genuine EOF right after an invalid answer, which must never
// reach the operator as a panic: runForm must recover it and report
// cancellation instead.
func TestRunFormRecoversInvalidAnswerThenEOF(t *testing.T) {
	inputs := []string{"8", "0\n", "abc\n"}

	t.Run("menu", func(t *testing.T) {
		for _, in := range inputs {
			t.Run(in, func(t *testing.T) {
				var out bytes.Buffer
				p := newHuhPrompter(true, strings.NewReader(in), &out)
				choice, cancelled, err := p.selectMenuEntry("status")
				if err != nil {
					t.Fatalf("selectMenuEntry(%q): %v\noutput:\n%s", in, err, out.String())
				}
				if !cancelled || choice != menuQuit {
					t.Fatalf("selectMenuEntry(%q) = (%v, %v, nil), want (menuQuit, true, nil)", in, choice, cancelled)
				}
			})
		}
	})

	t.Run("confirm with back", func(t *testing.T) {
		for _, in := range inputs {
			t.Run(in, func(t *testing.T) {
				var out bytes.Buffer
				p := newHuhPrompter(true, strings.NewReader(in), &out)
				decision, err := p.Confirm("Apply these changes?", true)
				if err != nil {
					t.Fatalf("Confirm(%q): %v\noutput:\n%s", in, err, out.String())
				}
				if decision != installCancelled {
					t.Fatalf("Confirm(%q) = %v, want installCancelled", in, decision)
				}
			})
		}
	})

	t.Run("select hosts", func(t *testing.T) {
		candidates := []hostCandidate{{Name: "codex"}, {Name: "claude"}, {Name: "grok"}}
		for _, in := range inputs {
			t.Run(in, func(t *testing.T) {
				var out bytes.Buffer
				p := newHuhPrompter(true, strings.NewReader(in), &out)
				hosts, ok, err := p.SelectHosts(candidates)
				if err != nil {
					t.Fatalf("SelectHosts(%q): %v\noutput:\n%s", in, err, out.String())
				}
				if ok || hosts != nil {
					t.Fatalf("SelectHosts(%q) = (%v, %v, nil), want (nil, false, nil)", in, hosts, ok)
				}
			})
		}
	})
}

// TestFormKeyMapBindsEscToQuit is the unit test on the keymap binding T2 fix
// round item 2 asks for: huh v2.0.3 binds Quit to ctrl+c only
// (NewDefaultKeyMap); formKeyMap must add esc without dropping ctrl+c, so
// Form.Update's key.Matches(msg, f.keymap.Quit) check (form.go) aborts the
// form on either key (AC10: "Ctrl-C o Esc").
func TestFormKeyMapBindsEscToQuit(t *testing.T) {
	keys := formKeyMap.Quit.Keys()
	want := map[string]bool{"ctrl+c": false, "esc": false}
	for _, k := range keys {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, seen := range want {
		if !seen {
			t.Fatalf("formKeyMap.Quit %v is missing %q", keys, k)
		}
	}
	if !formKeyMap.Quit.Enabled() {
		t.Fatal("formKeyMap.Quit is disabled")
	}
}

// TestAccessibleModeEmitsNoANSICodes is T2 fix round item 6: accessible
// mode's own titles must never carry ANSI escape codes, whether or not
// NO_COLOR is set, since huh's default theme (ThemeCharm) colors them even
// though RunAccessible never renders a full-screen view.
func TestAccessibleModeEmitsNoANSICodes(t *testing.T) {
	for _, noColor := range []string{"", "1"} {
		t.Run("NO_COLOR="+noColor, func(t *testing.T) {
			t.Setenv("NO_COLOR", noColor)
			o := interfaceTestOptions(t)
			var out bytes.Buffer
			if err := openInterface(false, true, strings.NewReader("7\n"), &out, o); err != nil {
				t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
			}
			if strings.ContainsRune(out.String(), '\x1b') {
				t.Fatalf("accessible output contains an ANSI escape code: %q", out.String())
			}
		})
	}
}

// TestCheckPendingOnOpenCoreSuccessOmitsRecoverPhrase covers T2 fix round
// item 5: a PendingCore recovery only ever reaches its own success message
// after RecoverCore has already returned without an error, so the
// interface must never then also tell the operator to "Run hive recover"
// — that already happened. A fake Pending/RecoverCore pair (the same
// injection seam install_test.go's own pending tests use) with the real huh
// prompter answering "y" in accessible mode.
func TestCheckPendingOnOpenCoreSuccessOmitsRecoverPhrase(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingCore, nil }
	recovered := false
	dependencies.RecoverCore = func(string) (string, error) {
		recovered = true
		return "recovered-id", nil
	}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("y\n"), &out)
	handled, err := checkPendingOnOpen(mo, false, &out, p, dependencies)
	if err != nil {
		t.Fatalf("checkPendingOnOpen: %v\noutput:\n%s", err, out.String())
	}
	if !handled {
		t.Fatal("expected the pending operation to be handled")
	}
	if !recovered {
		t.Fatal("RecoverCore was not called")
	}
	if !strings.Contains(out.String(), "Recovery: recovered-id.") {
		t.Fatalf("missing the recovery success message: %s", out.String())
	}
	if strings.Contains(out.String(), "hive recover") || strings.Contains(out.String(), "install.sh") {
		t.Fatalf("a recovery that already succeeded must not also tell the operator to recover: %s", out.String())
	}
}

// TestCheckPendingOnOpenOnboardingPartialNamesRecover covers T2 fix round
// item 5's other half: unlike PendingCore, an onboarding recovery can
// finish "partial" (some optional capability is still pending), and only
// then must the interface still name hive recover, properly capitalized as
// a new sentence.
func TestCheckPendingOnOpenOnboardingPartialNamesRecover(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
	dependencies.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "partial"}, nil
	}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("y\n"), &out)
	handled, err := checkPendingOnOpen(mo, false, &out, p, dependencies)
	if err != nil {
		t.Fatalf("checkPendingOnOpen: %v\noutput:\n%s", err, out.String())
	}
	if !handled {
		t.Fatal("expected the pending operation to be handled")
	}
	if !strings.Contains(out.String(), "Run hive recover.") {
		t.Fatalf("a still-partial onboarding recovery must name Run hive recover: %s", out.String())
	}
}

// TestCheckPendingOnOpenOnboardingCompletedOmitsRecoverPhrase is
// OnboardingPartialNamesRecover's counterpart: phase "completed" means
// nothing more is pending, so the phrase must not appear.
func TestCheckPendingOnOpenOnboardingCompletedOmitsRecoverPhrase(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
	dependencies.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "completed"}, nil
	}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("y\n"), &out)
	if _, err := checkPendingOnOpen(mo, false, &out, p, dependencies); err != nil {
		t.Fatalf("checkPendingOnOpen: %v\noutput:\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "hive recover") {
		t.Fatalf("a completed onboarding recovery must not also tell the operator to recover: %s", out.String())
	}
}

// TestCheckPendingOnOpenNamesExplicitStateDir covers T2 fix round item 5's
// --state-dir requirement: when the interface itself was opened against an
// explicit state directory (a synthetic or test one, never the real user's
// default), a recovery phrase that still names hive recover must also name
// that same --state-dir, or "hive recover" would silently target the real
// one instead.
func TestCheckPendingOnOpenNamesExplicitStateDir(t *testing.T) {
	o := interfaceTestOptions(t)
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
	dependencies.RecoverOnboarding = func(string, onboardingAdapter) (management.OnboardingResult, error) {
		return management.OnboardingResult{ID: "onboarding-id", Phase: "partial"}, nil
	}
	var out bytes.Buffer
	p := newHuhPrompter(true, strings.NewReader("y\n"), &out)
	if _, err := checkPendingOnOpen(mo, true, &out, p, dependencies); err != nil {
		t.Fatalf("checkPendingOnOpen: %v\noutput:\n%s", err, out.String())
	}
	want := "Run hive recover --state-dir " + o.StateDir + "."
	if !strings.Contains(out.String(), want) {
		t.Fatalf("missing explicit --state-dir in the recovery phrase: got %q, want it to contain %q", out.String(), want)
	}
}

// TestOpenInterfaceDeclinesPendingAndReturnsToMenu covers "a cancelled
// stub/confirm returns to menu" for the on-open pending check: with a real
// (hand-crafted, like install_test.go's own TestInstallRecoversOnboarding-
// BeforeCorePending) pending.json and no onboarding journal, declining the
// recovery confirmation changes nothing and still reaches the menu, where
// Quit then exits cleanly.
func TestOpenInterfaceDeclinesPendingAndReturnsToMenu(t *testing.T) {
	o := interfaceTestOptions(t)
	if err := os.WriteFile(filepath.Join(o.StateDir, "pending.json"), []byte(`{"id":"core"}`), 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := openInterface(false, true, strings.NewReader("n\n7\n"), &out, o); err != nil {
		t.Fatalf("openInterface: %v\noutput:\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "An operation is pending") {
		t.Fatalf("missing pending notice: %s", out.String())
	}
	if strings.Contains(out.String(), "Recovery:") {
		t.Fatalf("declined recovery still ran: %s", out.String())
	}
}

// TestRunInterfaceCommandNoTerminalNoAccessible covers the explicit `hive
// tui` subcommand's own distinct failure message, which must name
// HIVE_ACCESSIBLE unlike bare hive's unchanged usage error.
func TestRunInterfaceCommandNoTerminalNoAccessible(t *testing.T) {
	t.Setenv("HIVE_ACCESSIBLE", "")
	err := runInterfaceCommand(nil)
	if err == nil || !strings.Contains(err.Error(), "HIVE_ACCESSIBLE") {
		t.Fatalf("got %v, want an error naming HIVE_ACCESSIBLE", err)
	}
	if err != nil && err.Error() == usageMessage {
		t.Fatal("hive tui must not fall back to bare hive's usage message")
	}
}
