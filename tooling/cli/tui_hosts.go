// tui_hosts.go implements the Install CLIs and Remove CLIs screens
// (design.md "La interfaz").
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tricell-hive/tooling/management"
)

// installScreen implements Install CLIs (design.md "La interfaz"): its own
// one-line header ("Cada pantalla empieza con un encabezado de una línea"),
// then the same runInstallFlowWith the plain `hive install` command uses,
// driven by this session's own huh prompter instead of an installTerminal,
// so it inherits package verification, options normalization, the
// pending-operation check, the host/provider wizard, and
// BindRetainedInstaller's own nil default (only the online bootstrap entry
// point ever sets it) for free — including every cancel/decline/EOF and
// error message that flow already prints.
func installScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Install CLIs ==")
	if !sourceHasCatalog(o.Source) {
		return fmt.Errorf("Run hive from a Hive checkout or package, or pass --source")
	}
	dependencies := defaultInstallDependencies(nativeProviderAdapterFactory)
	return runInstallFlowWith(o, false, out, p, dependencies, true)
}

// sourceHasCatalog reports whether source looks like a real Hive checkout
// or package (design.md "Punto de entrada": "Las pantallas que necesitan
// una fuente ... muestran un error ... si esa fuente no tiene catálogo").
// It only needs to check for content/guidance/global.md
// (management.GlobalSource), the one file every install plan writes
// unconditionally (management.loadRelease's own sources list starts with
// it) — a source without even that file cannot ever produce a valid plan.
func sourceHasCatalog(source string) bool {
	_, err := os.Stat(filepath.Join(source, management.GlobalSource))
	return err == nil
}

// removeScreen implements Remove CLIs (design.md "La interfaz"): a
// MultiSelect of the hosts currently registered in user scope, then
// removeFlow's own summary/confirm/apply. The empty state matches Status's
// own message, since there is nothing to remove without a registered host.
func removeScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Remove CLIs ==")
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		fmt.Fprintln(out, "No CLI hosts are registered. Choose Install CLIs.")
		return nil
	}
	candidates := make([]hostCandidate, len(hosts))
	for i, h := range hosts {
		candidates[i] = hostCandidate{Name: h, Registered: true}
	}
	selected, ok, err := p.SelectHosts(candidates)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	return removeFlow(o, selected, out, p)
}

// removeFlow builds a "remove" plan for the selected hosts and applies it
// once confirmed (design.md "La interfaz", Remove CLIs row): its own
// summary (showRemoveSummary), never showInstallSummary's, since a remove
// plan's own Before/After meaning differs (After == nil means the resource
// is deleted, not merely unchanged).
func removeFlow(o management.Options, hosts []string, out io.Writer, p prompter) error {
	ro := o
	ro.Hosts = hosts
	plan, err := management.BuildPlan("remove", ro)
	if err != nil {
		return err
	}
	showRemoveSummary(out, plan)
	decision, err := p.Confirm("Apply these changes?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	result, err := (management.Engine{}).Apply(plan)
	if err != nil {
		return err
	}
	reportApplyResult(out, "Hive removed", result)
	return nil
}

// showRemoveSummary is removeFlow's own summary (design.md "La interfaz",
// Remove CLIs row): which files are removed outright, which shared
// resources are kept because other, non-selected hosts still consume them,
// and which voice blocks are removed. It never reuses showInstallSummary:
// a remove plan's own Change.After == nil means the resource is deleted,
// the opposite of what "unchanged" means for install/update.
func showRemoveSummary(out io.Writer, p management.Plan) {
	fmt.Fprintf(out, "Remove %s\n", strings.Join(p.Hosts, ", "))
	fmt.Fprintf(out, "Private backups: %s\n", filepath.Join(p.StateDir, "transactions"))
	var removed, kept []string
	for _, ch := range p.Changes {
		if ch.Before == nil {
			continue
		}
		if ch.After == nil {
			removed = append(removed, ch.Target.Path)
		} else {
			kept = append(kept, ch.Target.Path)
		}
	}
	sort.Strings(removed)
	sort.Strings(kept)
	fmt.Fprintf(out, "Files to remove: %d\n", len(removed))
	for _, path := range installRoots(removed, p.Config) {
		fmt.Fprintf(out, "  %s\n", path)
	}
	if len(kept) > 0 {
		fmt.Fprintf(out, "Shared resources kept for other consumers: %d\n", len(kept))
		for _, path := range installRoots(kept, p.Config) {
			fmt.Fprintf(out, "  %s\n", path)
		}
	}
	var voiceRemoved []string
	for _, vc := range p.Voice {
		if vc.After == nil {
			voiceRemoved = append(voiceRemoved, vc.Path)
		}
	}
	if len(voiceRemoved) > 0 {
		sort.Strings(voiceRemoved)
		fmt.Fprintf(out, "Voice blocks to remove: %d\n", len(voiceRemoved))
		for _, path := range voiceRemoved {
			fmt.Fprintf(out, "  %s\n", path)
		}
	}
	fmt.Fprintln(out, "Close these CLI sessions before continuing.")
}
