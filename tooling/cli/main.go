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
func run(args []string) error {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Println("hive setup [--home DIR]  (read-only optional Context7 guidance)")
		fmt.Println("hive plan install|remove --hosts codex,claude,grok,pi,opencode,cursor --scope user [--out FILE]\nhive plan install|remove --hosts codex,claude --scope project --root DIR [--out FILE]\nhive apply --plan FILE\nhive status --hosts codex,claude,grok,pi,opencode,cursor --scope user\nhive recover [--state-dir DIR]")
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("usage: hive setup | plan install|remove | apply --plan FILE | status | recover --state-dir DIR")
	}
	if args[0] == "setup" {
		return setup(args[1:], os.Stdout)
	}
	cmd := args[0]
	args = args[1:]
	action := ""
	if cmd == "plan" {
		if len(args) == 0 {
			return fmt.Errorf("plan requires install or remove")
		}
		action = args[0]
		args = args[1:]
	}
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	var o management.Options
	var hosts, out, planFile string
	if cmd == "plan" || cmd == "status" {
		fs.StringVar(&o.Scope, "scope", "", "user or project (required for plan/status)")
		fs.StringVar(&o.Home, "home", "", "explicit synthetic home; ignores host environment paths")
		fs.StringVar(&o.Root, "root", "", "project root")
		fs.StringVar(&o.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
		fs.StringVar(&hosts, "hosts", "", "comma-separated codex,claude,grok,pi,opencode,cursor (project scope: codex,claude)")
	}
	if cmd == "plan" {
		fs.StringVar(&o.Source, "source", ".", "source checkout")
		fs.StringVar(&o.ReleaseID, "release", "", "previously installed release hash")
		fs.StringVar(&out, "out", "", "new plan output file")
	}
	if cmd == "apply" {
		fs.StringVar(&planFile, "plan", "", "saved plan")
	}
	if cmd == "recover" {
		fs.StringVar(&o.StateDir, "state-dir", "", "state directory")
	}
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
	switch cmd {
	case "plan":
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
	case "apply":
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
	case "status":
		s, err := management.Status(o)
		if err != nil {
			return err
		}
		return output(s)
	case "recover":
		if o.StateDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			o.StateDir = filepath.Join(home, "Library", "Application Support", "tricell-hive")
		}
		s, err := (management.Engine{}).Recover(o.StateDir)
		if err != nil {
			return err
		}
		return output(map[string]string{"result": s})
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
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
