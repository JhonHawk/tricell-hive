// tui.go is the terminal interface's entry (D14-A): bare `hive` and the
// explicit `hive tui` subcommand both open the full-screen application in
// tui_app.go when stdin and stdout are terminals. Without a terminal, bare
// `hive` keeps today's usage error and `hive tui` names the text commands.
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
// synthetic home instead of the real one (design.md "Entry point" (archived record heading, originally in Spanish)).
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
const noTerminalMessage = "hive tui needs a terminal; use the text commands: hive status, install, update, releases, voice, doctor, models, plan/apply to remove hosts, recover"

// newAppConfig builds the application's configuration from the entry options:
// the options normalized, and the default dependencies. Colors are left to the
// caller.
func newAppConfig(o options) (appConfig, error) {
	mo := management.Options{Scope: "user", Home: o.Home, StateDir: o.StateDir, Source: o.Source}
	_, stateDir, err := management.NormalizeOptions(mo)
	if err != nil {
		return appConfig{}, err
	}
	mo.StateDir = stateDir
	return appConfig{
		Options: mo,
		// A synthetic --home moves the state directory too, so a bare `hive
		// recover` would target the real user's state: name it then as well.
		ExplicitStateDir: o.StateDir != "" || o.Home != "",
		Deps:             defaultInstallDependencies(nativeProviderAdapterFactory),
	}, nil
}

// runApp opens the full-screen application over a normalized copy of o. The
// terminal background is detected before the program starts, as the earlier
// interface did; NO_COLOR skips the detection and removes every color.
func runApp(o options, in io.Reader, out io.Writer) error {
	cfg, err := newAppConfig(o)
	if err != nil {
		return err
	}
	cfg.NoColor = os.Getenv("NO_COLOR") != ""
	cfg.Dark = true
	if inFile, ok := in.(*os.File); ok && !cfg.NoColor {
		if outFile, ok := out.(*os.File); ok {
			cfg.Dark = lipgloss.HasDarkBackground(inFile, outFile)
		}
	}
	return runAppWith(cfg, in, out)
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

// shellQuote returns s as one shell word: unchanged when it holds only safe
// characters, otherwise in single quotes with any quote inside escaped.
func shellQuote(s string) string {
	safe := s != ""
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_@%+=:,./-", r)) {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// recoveryTextOnOpen is the recovery phrase the interface uses: it names hive
// recover — with --state-dir (quoted for the shell) when the interface was
// opened against a state directory other than the default, so the operator
// recovers the same, possibly synthetic, state they are looking at — and only
// when the attempted recovery still leaves something pending, never after one
// that already succeeded.
func recoveryTextOnOpen(explicitStateDir bool) func(stateDir string, stillNeeded bool) string {
	return func(stateDir string, stillNeeded bool) string {
		if !stillNeeded {
			return ""
		}
		if explicitStateDir {
			return "Run hive recover --state-dir " + shellQuote(stateDir)
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

// interfaceStatusLine renders the menu's own one-line summary (design.md
// "The interface", archived record heading originally in Spanish): the registered host count, the release currently
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
// (design.md "Entry point" (archived record heading, originally in Spanish)): its own --home/--state-dir/--source mirror
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
