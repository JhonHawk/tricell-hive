package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

var installerHosts = []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}

func install(args []string, in io.Reader, out io.Writer, interactive bool) error {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(out)
	o := management.Options{Scope: "user"}
	var hosts string
	var dry bool
	fs.StringVar(&hosts, "hosts", "", "selected CLIs, comma-separated; detected by default")
	fs.StringVar(&o.Home, "home", "", "synthetic home; ignores user paths and executables")
	fs.StringVar(&o.StateDir, "state-dir", "", "private state directory")
	fs.StringVar(&o.Source, "source", ".", "this distribution content")
	fs.BoolVar(&dry, "dry-run", false, "preview changes without applying them")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("install accepts no positional arguments")
	}
	if err := distribution.VerifyIfPackaged(o.Source); err != nil {
		return err
	}
	_, stateDir, err := management.NormalizeOptions(o)
	if err != nil {
		return err
	}
	o.StateDir = stateDir
	if _, err := os.Lstat(filepath.Join(stateDir, "pending.json")); err == nil {
		fmt.Fprintf(out, "An operation is pending. Recovery directory: %s\n", stateDir)
		if dry {
			return nil
		}
		yes, err := confirmInstall(in, out, interactive, "Recover the pending operation?")
		if err != nil || !yes {
			return err
		}
		result, err := (management.Engine{}).Recover(stateDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "Recovery: %s. Run ./install.sh again to install.\n", result)
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if hosts != "" {
		for _, h := range strings.Split(hosts, ",") {
			o.Hosts = append(o.Hosts, strings.TrimSpace(h))
		}
	} else {
		seen := map[string]bool{}
		registered, err := management.RegisteredHosts(o)
		if err != nil {
			return err
		}
		for _, h := range registered {
			seen[h] = true
		}
		oldHosts, err := management.DetectLegacyHosts(o)
		if err != nil {
			return err
		}
		for _, h := range oldHosts {
			seen[h] = true
		}
		if o.Home == "" {
			for _, h := range installerHosts {
				if _, err := exec.LookPath(h); err == nil {
					seen[h] = true
				}
			}
		}
		for h := range seen {
			o.Hosts = append(o.Hosts, h)
		}
		sort.Strings(o.Hosts)
	}
	if len(o.Hosts) == 0 {
		return fmt.Errorf("no CLIs detected; use --hosts codex,claude,grok,pi,opencode,cursor to select them")
	}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		return err
	}
	unchanged, err := management.PlanUnchanged(p)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Hive · %s\n", strings.Join(p.Hosts, ", "))
	if unchanged {
		fmt.Fprintln(out, "Hive is already up to date. No changes.")
		return nil
	}
	action := "Install / update"
	if len(p.Legacy) > 0 {
		action = "Migrate legacy Hive and install the rebuild"
	}
	fmt.Fprintln(out, action)
	fmt.Fprintf(out, "Private backups: %s\n", filepath.Join(p.StateDir, "transactions"))
	fmt.Fprintf(out, "Rebuild resources: %d; legacy changes: %d\n", len(p.Changes), len(p.Legacy))
	destinations := map[string]bool{}
	for _, ch := range p.Changes {
		destinations[ch.Target.Path] = true
	}
	for _, ch := range p.Legacy {
		destinations[ch.Path] = true
	}
	var paths []string
	for path := range destinations {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	shown := paths
	if !dry {
		shown = installRoots(paths, p.Config)
	}
	for _, path := range shown {
		fmt.Fprintf(out, "  %s\n", path)
	}
	if !dry {
		fmt.Fprintln(out, "Use --dry-run to see the full file list before applying.")
	}
	fmt.Fprintln(out, "Close these CLI sessions before continuing.")
	if dry {
		fmt.Fprintln(out, "Preview: installation was not changed.")
		return nil
	}
	yes, err := confirmInstall(in, out, interactive, "Apply these changes?")
	if err != nil || !yes {
		return err
	}
	result, err := (management.Engine{}).Apply(p)
	if err != nil {
		return fmt.Errorf("installation did not complete: %w; run ./install.sh again to check recovery", err)
	}
	fmt.Fprintf(out, "Hive installed and verified (%s). Open new CLI sessions.\n", result)
	return nil
}

func installRoots(paths []string, c target.Config) []string {
	roots := []string{c.CodexHome, c.ClaudeHome, c.GrokHome, c.PiHome, c.OpenCodeHome, c.CursorHome, filepath.Join(c.Home, ".agents"), filepath.Join(c.Home, ".codex"), filepath.Join(c.Home, ".claude"), filepath.Join(c.Home, ".grok"), filepath.Join(c.Home, ".pi", "agent"), filepath.Join(c.Home, ".config", "opencode")}
	sort.Slice(roots, func(i, j int) bool { return len(roots[i]) > len(roots[j]) })
	seen := map[string]bool{}
	for _, path := range paths {
		displayed := path
		for _, root := range roots {
			if root != "" && (path == root || strings.HasPrefix(path, root+string(filepath.Separator))) {
				displayed = root
				break
			}
		}
		seen[displayed] = true
	}
	result := make([]string, 0, len(seen))
	for path := range seen {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func confirmInstall(in io.Reader, out io.Writer, interactive bool, prompt string) (bool, error) {
	if !interactive {
		return false, fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to inspect")
	}
	fmt.Fprintf(out, "%s [y/N] ", prompt)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	// A truncated answer (EOF) is cancellation, never implicit consent.
	if err == io.EOF || (answer != "y" && answer != "yes") {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return false, nil
	}
	return true, nil
}
