// tui_hosts.go holds the pieces of the CLIs view that come from the removal and
// install commands: the source-catalog check, the remove plan and its summary.
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

// sourceHasCatalog reports whether source looks like a real Hive checkout
// or package (design.md "Entry point" (archived record heading, originally in Spanish):
// "Screens that need a source ... show an error ... if that source has no
// catalog").
// It only needs to check for content/guidance/global.md
// (management.GlobalSource), the one file every install plan writes
// unconditionally (management.loadRelease's own sources list starts with
// it) — a source without even that file cannot ever produce a valid plan.
func sourceHasCatalog(source string) bool {
	_, err := os.Stat(filepath.Join(source, management.GlobalSource))
	return err == nil
}

// removeApply applies a remove plan. It is a variable only so tests can
// interpose on the write (hold it, or make it fail); production never changes
// it.
var removeApply = func(p management.Plan) (string, error) { return (management.Engine{}).Apply(p) }

// buildRemovePlan builds the removal plan for hosts: the same
// BuildPlan("remove") `hive plan remove` runs for them.
func buildRemovePlan(o management.Options, hosts []string) (management.Plan, error) {
	ro := o
	ro.Hosts = hosts
	return management.BuildPlan("remove", ro)
}

// showRemoveSummary is the removal summary (design.md "The interface" (archived record heading, originally in Spanish), Remove
// CLIs row), never showInstallSummary's: which files are removed outright, which shared
// resources are kept because other, non-selected hosts still consume them,
// and which voice blocks are removed. It never reuses showInstallSummary:
// a remove plan's own Change.After == nil means the resource is deleted,
// the opposite of what "unchanged" means for install/update.
func showRemoveSummary(out io.Writer, p management.Plan) {
	fmt.Fprintf(out, "Remove %s\n", strings.Join(p.Hosts, ", "))
	fmt.Fprintf(out, "Private backups: %s\n", filepath.Join(p.StateDir, "transactions"))
	var removed, kept, released []string
	for _, ch := range p.Changes {
		if ch.Before == nil {
			continue
		}
		if ch.Gone && ch.After == nil {
			// Already deleted by the user: the record is released, nothing is written.
			released = append(released, ch.Target.Path)
		} else if ch.After == nil {
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
		if vc.Gone && vc.After == nil {
			released = append(released, vc.Path)
		} else if vc.After == nil {
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
	if len(released) > 0 {
		sort.Strings(released)
		fmt.Fprintf(out, "Already missing, released without writing: %d\n", len(released))
		for _, path := range released {
			fmt.Fprintf(out, "  %s\n", path)
		}
	}
	fmt.Fprintln(out, "Close these CLI sessions before continuing.")
}
