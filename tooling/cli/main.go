// Command hive manages versioned Hive content from this checkout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "hive:", err)
		os.Exit(1)
	}
}
func output(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// usageMessage is hive's own top-level usage error. It is unchanged by the
// interface (D14-A): bare `hive` outside a terminal still prints exactly this
// (runBareInterface, tui.go).
const usageMessage = "usage: hive --version | setup | install | plan install|remove | apply --plan FILE | status | recover --state-dir DIR | update | releases | voice list|set|off"

// run resolves --version and --help, then delegates every subcommand to its
// own function. Each subcommand function defines its own flag.FlagSet, so
// its flags and usage text stay local to it.
func run(args []string) error {
	if len(args) == 1 && args[0] == "--version" {
		fmt.Println(version.Current)
		return nil
	}
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printHelp()
		return nil
	}
	if len(args) == 0 {
		return runBareInterface()
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "setup":
		return runSetup(rest)
	case "install":
		return runInstall(rest)
	case "bootstrap":
		return runBootstrap(rest)
	case "plan":
		return runPlan(rest)
	case "apply":
		return runApply(rest)
	case "status":
		return runStatus(rest)
	case "recover":
		return runRecover(rest)
	case "update":
		return runUpdate(rest)
	case "releases":
		return runReleases(rest)
	case "voice":
		return runVoice(rest)
	case "doctor":
		return runDoctor(rest, os.Stdout, realDoctorDeps())
	case "models":
		return runModels(rest, os.Stdout)
	case "project":
		return runProject(rest)
	case "tui":
		return runInterfaceCommand(rest)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printHelp() {
	fmt.Println("hive tui [--home DIR] [--state-dir DIR] [--source DIR]  (full-screen interface; bare `hive` in a terminal opens it too; without a terminal use the text commands below)")
	fmt.Println("hive bootstrap --origin URL --version X.Y.Z --manager PATH --manager-sha256 HASH  (invoked by bootstrap.sh only)")
	fmt.Println("hive install [--hosts codex,claude,grok,pi,opencode,cursor] [--dry-run]  (interactive installer)")
	fmt.Println("hive setup [--home DIR]  (read-only optional Context7 guidance)")
	fmt.Println("hive plan install|remove --hosts codex,claude,grok,pi,opencode,cursor --scope user [--out FILE]\nhive plan install|remove --hosts codex,claude --scope project --root DIR [--out FILE]\nhive apply --plan FILE\nhive status --hosts codex,claude,grok,pi,opencode,cursor --scope user\nhive recover [--state-dir DIR]")
	fmt.Println("hive update [--rev REV] [--source DIR] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]  (update from a Git commit; needs Git and a checkout, unlike install.sh)\nhive releases [--home DIR] [--state-dir DIR]  (list retained release snapshots)")
	fmt.Println("hive voice list [--source DIR]  (print each voice's ID and description)\nhive voice set ID [--address sir|name|none] [--name NAME] [--intensity subtle|marked] [--source DIR] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]\nhive voice off [--source DIR] [--home DIR] [--state-dir DIR] [--dry-run] [--out FILE]")
	fmt.Println("hive doctor [--home DIR] [--state-dir DIR] [--project DIR]  (read-only diagnostics: CLIs, installation, sessions, integrations, project)\nhive models [--home DIR] [--state-dir DIR]  (read-only effective model and effort per role)")
	fmt.Println("hive project set [--project DIR] --set 'Key: value'... [--unset Key...] [--dry-run]  (edit the ## Hive section of the repository's AGENTS.md after confirmation)")
}

// readOnlyOptions parses the --home and --state-dir flags that the read-only
// commands share, and fails when the state directory cannot be one: findings
// never fail these commands, only their own errors do.
func readOnlyOptions(fs *flag.FlagSet, args []string) (management.Options, bool, error) {
	var o management.Options
	fs.StringVar(&o.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&o.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return o, false, nil
		}
		return o, false, err
	}
	if fs.NArg() != 0 {
		return o, false, fmt.Errorf("unexpected positional arguments")
	}
	o.Scope = "user"
	_, stateDir, err := management.NormalizeOptions(o)
	if err != nil {
		return o, false, err
	}
	if info, err := os.Stat(stateDir); err == nil && !info.IsDir() {
		return o, false, fmt.Errorf("state directory %s is not a directory", stateDir)
	}
	return o, true, nil
}

// runDoctor prints the read-only diagnostics with the same text as the views.
func runDoctor(args []string, w io.Writer, deps doctorDeps) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	var project string
	fs.StringVar(&project, "project", "", "project directory whose ## Hive section is checked (default: the current directory)")
	o, ok, err := readOnlyOptions(fs, args)
	if !ok {
		return err
	}
	if project == "" {
		if project, err = os.Getwd(); err != nil {
			return err
		}
	}
	renderDoctorText(collectDoctor(o, project, deps), w)
	return nil
}

// runModels prints the effective model and effort of every installed role.
func runModels(args []string, w io.Writer) error {
	o, ok, err := readOnlyOptions(flag.NewFlagSet("models", flag.ContinueOnError), args)
	if !ok {
		return err
	}
	hosts, rows, err := collectModels(o)
	if err != nil {
		return err
	}
	renderModelsText(hosts, rows, w)
	return nil
}

func runUpdate(args []string) error {
	return update(args, os.Stdin, os.Stdout, terminalInput(os.Stdin))
}

func runReleases(args []string) error {
	var o management.Options
	fs := flag.NewFlagSet("releases", flag.ContinueOnError)
	fs.StringVar(&o.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&o.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	o.Scope = "user"
	entries, err := management.Releases(o)
	if err != nil {
		return err
	}
	if entries == nil {
		entries = []management.ReleaseEntry{}
	}
	return output(entries)
}

func runSetup(args []string) error {
	return setup(args, os.Stdout)
}

func runInstall(args []string) error {
	return install(args, os.Stdin, os.Stdout, terminalInput(os.Stdin))
}

func runBootstrap(args []string) error {
	return bootstrap(args, os.Stdin, os.Stdout, terminalInput(os.Stdin))
}

// planFlagSet registers the flags shared by plan and status (scope, home,
// root, state-dir, hosts) on a FlagSet named for the given subcommand.
func planFlagSet(name string, o *management.Options, hosts *string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.StringVar(&o.Scope, "scope", "", "user or project (required for plan/status)")
	fs.StringVar(&o.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&o.Root, "root", "", "project root")
	fs.StringVar(&o.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	fs.StringVar(hosts, "hosts", "", "comma-separated codex,claude,grok,pi,opencode,cursor (project scope: codex,claude)")
	return fs
}

func runPlan(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("plan requires install or remove")
	}
	action := args[0]
	args = args[1:]
	var o management.Options
	var hosts, out string
	fs := planFlagSet("plan", &o, &hosts)
	fs.StringVar(&o.Source, "source", ".", "source checkout")
	fs.StringVar(&o.ReleaseID, "release", "", "previously installed release hash")
	fs.StringVar(&out, "out", "", "new plan output file")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if hosts != "" {
		o.Hosts = strings.Split(hosts, ",")
	}
	p, err := management.BuildPlan(action, o)
	if err != nil {
		return err
	}
	if action == "install" && o.Scope == "user" {
		if _, err := os.Stat(filepath.Join(p.StateDir, "state.json")); os.IsNotExist(err) {
			fmt.Println("First setup: Context7 is recommended but optional. Run hive setup for local discovery and official install/update guidance; this plan can proceed without it.")
		}
	}
	if out != "" {
		if err = management.SavePlan(out, p); err != nil {
			return err
		}
	}
	// Only managed content is printed: never personal text outside the block.
	for _, ch := range p.Changes {
		fmt.Printf("%s %s [%s]\n", action, ch.Target.Path, ch.Target.Kind)
		if ch.Before != nil {
			printRecord("--- managed before", ch.Before)
		}
		if ch.After != nil {
			printRecord("+++ managed after", ch.After)
		}
	}
	for _, vc := range p.Voice {
		fmt.Printf("%s voice %s\n", action, vc.Path)
		if vc.Before != nil {
			printVoiceSpan("--- managed before", vc.Before)
		}
		if vc.After != nil {
			printVoiceSpan("+++ managed after", vc.After)
		}
	}
	if p.VoiceWarning != "" {
		fmt.Println(p.VoiceWarning)
	}
	fmt.Printf("Plan %s\n", p.ID)
	if out == "" {
		fmt.Println("Preview only; use --out FILE to save an applicable plan.")
	}
	return nil
}

func runApply(args []string) error {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	var planFile string
	fs.StringVar(&planFile, "plan", "", "saved plan")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if planFile == "" {
		return fmt.Errorf("apply requires --plan")
	}
	p, err := management.LoadPlan(planFile)
	if err != nil {
		return err
	}
	result, err := (management.Engine{}).Apply(p)
	if err != nil {
		return err
	}
	return output(map[string]string{"result": result})
}

func runStatus(args []string) error {
	var o management.Options
	var hosts string
	fs := planFlagSet("status", &o, &hosts)
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if hosts != "" {
		o.Hosts = strings.Split(hosts, ",")
	}
	s, err := management.Status(o)
	if err != nil {
		return err
	}
	return output(s)
}

func runRecover(args []string) error {
	fs := flag.NewFlagSet("recover", flag.ContinueOnError)
	var o management.Options
	fs.StringVar(&o.StateDir, "state-dir", "", "state directory")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if o.StateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		o.StateDir, err = management.DefaultStateDir(home, false)
		if err != nil {
			return err
		}
	}
	// recover has no --home/--scope flags of its own: it always targets
	// the real user's own state, mirroring the user-scope resolution
	// just used above to default o.StateDir. Without an explicit scope,
	// NormalizeOptions (inside the adapter factory) refuses every
	// recovery with "explicit scope must be user or project", so no
	// pending onboarding could ever be reconciled from a second
	// terminal.
	o.Scope = "user"
	// RecoverOnboarding (reached through recoverWithAdapterFactory) is
	// the one recovery path (M5): when no onboarding journal is pending
	// it falls back to the core Engine.Recover itself, so this command
	// never branches on management.OnboardingPending on its own.
	result, err := recoverWithAdapterFactory(o, nativeProviderAdapterFactory)
	if err != nil {
		return err
	}
	return output(result)
}

func printRecord(label string, r *management.Record) {
	fmt.Println(label)
	if r.Target.LinkTarget != "" {
		fmt.Printf("symlink -> %s\n", r.Target.LinkTarget)
	} else {
		fmt.Printf("%s\n", r.Managed)
	}
	for _, c := range r.Consumers {
		fmt.Printf("consumer: %s (%s, %s)\n", c.Host, c.Scope, c.Context)
	}
}

func printVoiceSpan(label string, s *management.VoiceSpan) {
	fmt.Println(label)
	fmt.Printf("%s\n", s.Managed)
	for _, c := range s.Consumers {
		fmt.Printf("consumer: %s (%s, %s)\n", c.Host, c.Scope, c.Context)
	}
}
