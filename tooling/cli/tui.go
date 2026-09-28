// tui.go is the terminal interface's entry (D14-A): bare `hive` and the
// explicit `hive tui` subcommand both open the full-screen application in
// tui_app.go when stdin and stdout are terminals. Without a terminal, bare
// `hive` keeps today's usage error and `hive tui` names the text commands.
//
// The sequential interface (openInterface, runMenu and the huh screens) is no
// longer reachable from `hive` or `hive tui`; T10 deletes it, so it stays here
// only until then.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"tricell-hive/tooling/management"
)

// options is the interface's own explicit home/state-dir/source, exactly
// like install's own flags, letting an operator or a test open it over a
// synthetic home instead of the real one (design.md "Punto de entrada").
type options struct {
	Home, StateDir, Source string
}

// interfaceStdio is the entry's own real-environment dependency: the real
// terminal-detection check, the real stdin/stdout, (since bare `hive` takes no
// --home/--state-dir of its own) the real, default home and state directory,
// and start, which opens the application. A test overwrites these fields (and
// restores them) so run(nil) and run(["tui"]) are testable without a terminal
// and without a program (T2 fix round item 4; T3 fix-round leftover (b)).
var interfaceStdio = struct {
	isTTY    func() bool
	in       io.Reader
	out      io.Writer
	home     string
	stateDir string
	start    func(o options, in io.Reader, out io.Writer) error
}{
	isTTY: func() bool { return terminalInput(os.Stdin) && terminalInput(os.Stdout) },
	in:    os.Stdin,
	out:   os.Stdout,
	start: runApp,
}

// noTerminalMessage is what `hive tui` fails with when stdin or stdout is not
// a terminal: it names the text commands that work without one, including
// plan/apply for removing hosts and recover.
const noTerminalMessage = "hive tui needs a terminal; use the text commands: hive status, install, update, releases, voice, plan/apply to remove hosts, recover"

// runApp opens the full-screen application over a normalized copy of o. The
// terminal background is detected before the program starts, as the earlier
// interface did; NO_COLOR skips the detection and removes every color.
func runApp(o options, in io.Reader, out io.Writer) error {
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir, Source: o.Source}
	_, stateDir, err := management.NormalizeOptions(mo)
	if err != nil {
		return err
	}
	mo.StateDir = stateDir
	noColor := os.Getenv("NO_COLOR") != ""
	dark := true
	if inFile, ok := in.(*os.File); ok && !noColor {
		if outFile, ok := out.(*os.File); ok {
			dark = lipgloss.HasDarkBackground(inFile, outFile)
		}
	}
	return runAppWith(appConfig{
		Options:          mo,
		ExplicitStateDir: o.StateDir != "",
		Deps:             defaultInstallDependencies(nativeProviderAdapterFactory),
		Dark:             dark,
		NoColor:          noColor,
	}, in, out)
}

// runBareInterface implements bare `hive` with no arguments (D14-A): exactly
// `hive tui` with no options when stdin and stdout are terminals; otherwise
// today's usage error, unchanged. main.go's run keeps this as a single call so
// run itself stays within its own line budget.
func runBareInterface() error {
	if !interfaceStdio.isTTY() {
		return errors.New(usageMessage)
	}
	o := options{Home: interfaceStdio.home, StateDir: interfaceStdio.stateDir, Source: "."}
	return interfaceStdio.start(o, interfaceStdio.in, interfaceStdio.out)
}

// openInterface is the retired sequential interface's entry point. `hive` and
// `hive tui` no longer reach it (they open the full-screen application);
// only tests of the sequential screens still call it, until T10 deletes it
// together with the huh prompter.
func openInterface(isTTY, accessible bool, in io.Reader, out io.Writer, o options) error {
	if !isTTY && !accessible {
		return errors.New(usageMessage)
	}
	// explicitStateDir is o's own, pre-normalization StateDir: once known,
	// NormalizeOptions always resolves mo.StateDir to some concrete path,
	// so this is the only point that can still tell whether the operator
	// (or a test) asked for a specific one, which checkPendingOnOpen's own
	// recovery phrase must then name explicitly (T2 fix round item 5).
	explicitStateDir := o.StateDir != ""
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir, Source: o.Source}
	_, stateDir, err := management.NormalizeOptions(mo)
	if err != nil {
		return err
	}
	mo.StateDir = stateDir
	p := newHuhPrompter(accessible, in, out)
	dependencies := defaultInstallDependencies(nativeProviderAdapterFactory)
	if _, err := checkPendingOnOpen(mo, explicitStateDir, out, p, dependencies); err != nil {
		return err
	}
	return runMenu(mo, out, p)
}

// checkPendingOnOpen is the retired sequential interface's on-open check: it
// reuses handlePendingInstallOperation so its precedence and prompts match
// hive install's. dependencies is injectable so a test can exercise the
// offer/decline/recover paths without a real transaction journal, the same
// seam installWithDependencies already uses (install.go). The application
// (tui_app.go) does not call it; T10 deletes it.
func checkPendingOnOpen(o management.Options, explicitStateDir bool, out io.Writer, p prompter, dependencies installDependencies) (handled bool, err error) {
	return handlePendingInstallOperation(o, false, p, out, recoveryTextOnOpen(explicitStateDir), dependencies)
}

// recoveryTextOnOpen is the recovery phrase the interface uses (T2 fix round
// item 5): it names hive recover — with --state-dir when the interface was
// opened against an explicit one, so the operator recovers the same, possibly
// synthetic, state they are looking at — and only when the attempted recovery
// still leaves something pending, never after one that already succeeded.
func recoveryTextOnOpen(explicitStateDir bool) func(stateDir string, stillNeeded bool) string {
	return func(stateDir string, stillNeeded bool) string {
		if !stillNeeded {
			return ""
		}
		if explicitStateDir {
			return fmt.Sprintf("Run hive recover --state-dir %s", stateDir)
		}
		return "Run hive recover"
	}
}

// recoverPendingCmd recovers the pending operation or onboarding as a Cmd, so
// the application can run it as a write. It uses the same detection and
// recovery functions as handlePendingInstallOperation and `hive recover`
// (dependencies.AdapterFactory, RecoverOnboarding and RecoverCore), without
// that function's text question. o is the Cmd's own copy.
func recoverPendingCmd(kind management.PendingKind, o management.Options, explicitStateDir bool, dependencies installDependencies) tea.Cmd {
	return func() tea.Msg {
		text, err := recoverPending(kind, o, explicitStateDir, dependencies)
		return recoverDoneMsg{text: text, err: err}
	}
}

// recoverPending returns the same result line handlePendingInstallOperation
// prints for the interface.
func recoverPending(kind management.PendingKind, o management.Options, explicitStateDir bool, dependencies installDependencies) (string, error) {
	stateDir := o.StateDir
	recoveryText := recoveryTextOnOpen(explicitStateDir)
	switch kind {
	case management.PendingOnboarding:
		adapter, err := dependencies.AdapterFactory(newOnboardingInput(o, false))
		if err != nil {
			return "", err
		}
		result, err := dependencies.RecoverOnboarding(stateDir, adapter)
		if err != nil {
			return "", err
		}
		return withRecoverySentence(fmt.Sprintf("Onboarding recovery: %s (%s).", result.ID, result.Phase), recoveryText(stateDir, result.Phase != "completed")), nil
	case management.PendingCore:
		result, err := dependencies.RecoverCore(stateDir)
		if err != nil {
			return "", err
		}
		return withRecoverySentence(fmt.Sprintf("Recovery: %s.", result), recoveryText(stateDir, false)), nil
	}
	return "", nil
}

// menuEntry is one of the interface's seven fixed entries, in the order
// design.md "La interfaz" fixes them: the accessible mode's own numbering
// must never change between runs.
type menuEntry int

const (
	menuStatus menuEntry = iota
	menuInstall
	menuRemove
	menuUpdate
	menuReleases
	menuVoice
	menuQuit
)

// menuLabels names each entry in menuEntry's own order; selectMenuEntry
// (tui_prompter.go) turns them into a huh.Select whose numbering matches
// this order exactly.
var menuLabels = [...]string{
	menuStatus:   "Status",
	menuInstall:  "Install CLIs",
	menuRemove:   "Remove CLIs",
	menuUpdate:   "Update",
	menuReleases: "Releases",
	menuVoice:    "Voice",
	menuQuit:     "Quit",
}

// runMenu is the interface's own loop (design.md "La interfaz"): compute the
// status line, present the menu, and either open the chosen screen or stop.
// Quit, Ctrl-C and the end of input all exit cleanly (0). Every one of the
// six non-Quit entries (Status, Install CLIs, Remove CLIs, Update, Releases,
// Voice) is wired to its own real flow (T3 for Install/Remove, T4 for the
// rest); runMenuEntry's own default case stays only as a defensive
// fallback, never reached by selectMenuEntry's fixed set of choices.
func runMenu(o management.Options, out io.Writer, p *huhPrompter) error {
	for {
		line, err := interfaceStatusLine(o)
		if err != nil {
			return err
		}
		choice, cancelled, err := p.selectMenuEntry(line)
		if err != nil {
			return err
		}
		if cancelled || choice == menuQuit {
			return nil
		}
		if err := runMenuEntry(choice, o, out, p); err != nil {
			fmt.Fprintln(out, err.Error())
		}
	}
}

// runMenuEntry dispatches to the chosen screen (design.md "La interfaz").
// Every screen already handles its own cancellation and empty/limit states
// internally, printing them and returning nil (design.md "Cancelar y
// errores"); a non-nil error here is a genuine flow error, printed the same
// way for every screen instead of duplicating that print at each call site
// — including a Releases rollback the current manager cannot validate
// (tui_screens.go's rollbackFlow already wraps that one with its own
// documentation hint before returning it here).
func runMenuEntry(choice menuEntry, o management.Options, out io.Writer, p *huhPrompter) error {
	switch choice {
	case menuStatus:
		return statusScreen(o, out, p)
	case menuInstall:
		return installScreen(o, out, p)
	case menuRemove:
		return removeScreen(o, out, p)
	case menuUpdate:
		return updateScreen(o, out, p)
	case menuReleases:
		return releasesScreen(o, out, p)
	case menuVoice:
		return voiceScreen(o, out, p)
	default:
		fmt.Fprintln(out, "Not available yet.")
		return nil
	}
}

// interfaceStatusLine renders the menu's own one-line summary (design.md
// "La interfaz"): the registered host count, the release currently
// installed for them, and the active voice, each shown only when known. It
// only reads state, the same read-only contract hive status itself has.
func interfaceStatusLine(o management.Options) (string, error) {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return "", err
	}
	if len(hosts) == 0 {
		return "No CLI hosts are registered.", nil
	}
	unit := "hosts"
	if len(hosts) == 1 {
		unit = "host"
	}
	parts := []string{fmt.Sprintf("%d CLI %s", len(hosts), unit)}
	so := o
	so.Hosts = hosts
	entries, err := management.Status(so)
	if err != nil {
		return "", err
	}
	if release, ok := latestRelease(entries); ok {
		parts = append(parts, "release "+shortHash(release))
	}
	if voice, ok := activeVoiceID(entries); ok {
		parts = append(parts, "voice "+voice)
	}
	return strings.Join(parts, " · "), nil
}

// latestRelease returns the first non-empty per-file release ID from a
// Status call, which is enough to name "the" release for the menu's own
// one-line summary: every registered host installs from the same core
// release once hive install or update applies a plan.
func latestRelease(entries []management.StatusEntry) (string, bool) {
	for _, e := range entries {
		if e.Release != "" {
			return e.Release, true
		}
	}
	return "", false
}

// activeVoiceID extracts just the voice ID from Status's own "id (address,
// intensity)" voice row (management.formatVoiceStatus), which is all the
// menu's one-line summary shows.
func activeVoiceID(entries []management.StatusEntry) (string, bool) {
	for _, e := range entries {
		if e.Kind == "voice" && e.Voice != "" {
			id, _, _ := strings.Cut(e.Voice, " (")
			return id, true
		}
	}
	return "", false
}

// runInterfaceCommand implements the explicit `hive tui` subcommand
// (design.md "Punto de entrada"): its own --home/--state-dir/--source mirror
// install's. Without a terminal it fails with noTerminalMessage, which names
// the text commands: running this subcommand already expresses the intent to
// open the interface, unlike bare `hive`, whose failure stays today's plain
// usage error.
func runInterfaceCommand(args []string) error {
	var o options
	o.Source = "."
	fs := flag.NewFlagSet("tui", flag.ContinueOnError)
	fs.StringVar(&o.Home, "home", "", "synthetic home; ignores user paths and executables")
	fs.StringVar(&o.StateDir, "state-dir", "", "private state directory")
	fs.StringVar(&o.Source, "source", o.Source, "this distribution content")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("tui accepts no positional arguments")
	}
	if !interfaceStdio.isTTY() {
		return errors.New(noTerminalMessage)
	}
	return interfaceStdio.start(o, interfaceStdio.in, interfaceStdio.out)
}
