// tui.go implements D14-A: bare `hive` and the explicit `hive tui`
// subcommand both open the terminal interface. openInterface is the
// injectable entry point every caller and every test goes through
// (design.md "Punto de entrada"); runMenu is the interface's own loop once
// it is open (design.md "La interfaz").
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"tricell-hive/tooling/management"
)

// options is the interface's own explicit home/state-dir/source, exactly
// like install's own flags, letting an operator or a test open it over a
// synthetic home instead of the real one (design.md "Punto de entrada").
type options struct {
	Home, StateDir, Source string
}

// interfaceStdio is bare `hive`'s own real-environment dependency: the real
// terminal-detection check, the real stdin/stdout, and (since bare `hive`
// takes no --home/--state-dir of its own) the real, default home and state
// directory. A test overwrites these fields (and restores them) so
// run(nil) is testable end to end against a synthetic home instead of the
// real user's own state (T2 fix round item 4; T3 fix-round leftover (b)).
// hive tui's own --home/--state-dir flags make it testable without this
// seam, so runInterfaceCommand only reuses isTTY/in/out from it.
var interfaceStdio = struct {
	isTTY    func() bool
	in       io.Reader
	out      io.Writer
	home     string
	stateDir string
}{
	isTTY: func() bool { return terminalInput(os.Stdin) && terminalInput(os.Stdout) },
	in:    os.Stdin,
	out:   os.Stdout,
}

// runBareInterface implements bare `hive` with no arguments (D14-A): exactly
// `hive tui` with no options, opened in a terminal or accessible mode when
// either is available; with neither, openInterface returns usageMessage
// unchanged. main.go's run keeps this as a single call so run itself stays
// within its own line budget.
func runBareInterface() error {
	accessible := os.Getenv("HIVE_ACCESSIBLE") == "1"
	o := options{Home: interfaceStdio.home, StateDir: interfaceStdio.stateDir, Source: "."}
	return openInterface(interfaceStdio.isTTY(), accessible, interfaceStdio.in, interfaceStdio.out, o)
}

// openInterface is D14-A's injectable entry point: bare `hive` and the
// explicit `hive tui` subcommand both reach it, with isTTY and accessible
// resolved by their own caller so no test depends on the real terminal.
// Neither present returns today's usage error, unchanged: this is exactly
// what bare `hive` prints outside a terminal. The explicit `hive tui`
// subcommand's own caller (runInterfaceCommand) never reaches this branch:
// it checks the same condition first and fails with its own message naming
// HIVE_ACCESSIBLE instead.
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

// checkPendingOnOpen offers to recover a pending optional-onboarding or core
// operation when the interface opens (design.md "Al abrir"). It reuses
// handlePendingInstallOperation so its precedence and prompts match hive
// install's, with a recovery phrase of its own (T2 fix round item 5):
// naming hive recover — with --state-dir when the interface itself was
// opened against an explicit one, so the operator recovers the same,
// possibly synthetic, state they are looking at, rather than the real
// user's default — and only when the attempted recovery still leaves
// something pending, never after one that already succeeded.
// dependencies is injectable so a test can exercise the offer/decline/
// recover paths without a real transaction journal, the same seam
// installWithDependencies already uses (install.go).
func checkPendingOnOpen(o management.Options, explicitStateDir bool, out io.Writer, p prompter, dependencies installDependencies) (handled bool, err error) {
	recoveryText := func(stateDir string, stillNeeded bool) string {
		if !stillNeeded {
			return ""
		}
		if explicitStateDir {
			return fmt.Sprintf("Run hive recover --state-dir %s", stateDir)
		}
		return "Run hive recover"
	}
	return handlePendingInstallOperation(o, false, p, out, recoveryText, dependencies)
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
// Quit, Ctrl-C and the end of input all exit cleanly (0). Every screen other
// than Quit, Install CLIs and Remove CLIs is still a stub; T4 wires the
// rest to their real flows.
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
// way for every screen instead of duplicating that print at each call site.
func runMenuEntry(choice menuEntry, o management.Options, out io.Writer, p *huhPrompter) error {
	switch choice {
	case menuInstall:
		return installScreen(o, out, p)
	case menuRemove:
		return removeScreen(o, out, p)
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
// (design.md "Punto de entrada"): its own --home/--state-dir/--source
// mirror install's. Failing to find a terminal or HIVE_ACCESSIBLE names the
// environment variable directly, since running this subcommand at all
// already expresses the intent to open the interface — unlike bare `hive`,
// whose failure must stay today's plain usage error.
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
	isTTY := interfaceStdio.isTTY()
	accessible := os.Getenv("HIVE_ACCESSIBLE") == "1"
	if !isTTY && !accessible {
		return fmt.Errorf("hive tui needs an interactive terminal; set HIVE_ACCESSIBLE=1 to answer as plain-text questions instead")
	}
	return openInterface(isTTY, accessible, interfaceStdio.in, interfaceStdio.out, o)
}
