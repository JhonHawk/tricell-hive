// Command hive manages versioned Hive content from this checkout.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
		return fmt.Errorf("usage: hive --version | setup | install | plan install|remove | apply --plan FILE | status | recover --state-dir DIR")
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
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printHelp() {
	fmt.Println("hive bootstrap --origin URL --version X.Y.Z --manager PATH --manager-sha256 HASH  (invoked by bootstrap.sh only)")
	fmt.Println("hive install [--hosts codex,claude,grok,pi,opencode,cursor] [--dry-run]  (interactive installer)")
	fmt.Println("hive setup [--home DIR]  (read-only optional Context7 guidance)")
	fmt.Println("hive plan install|remove --hosts codex,claude,grok,pi,opencode,cursor --scope user [--out FILE]\nhive plan install|remove --hosts codex,claude --scope project --root DIR [--out FILE]\nhive apply --plan FILE\nhive status --hosts codex,claude,grok,pi,opencode,cursor --scope user\nhive recover [--state-dir DIR]")
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
