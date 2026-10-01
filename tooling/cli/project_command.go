// hive project edits the `## Hive` section of the AGENTS.md at the root of a
// Git repository (design.md "Escritura de ## Hive"). The file belongs to the
// project and Git is its recovery, so there is no plan, no journal and no
// --out: the command shows what will change, asks, and writes.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func runProject(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("project requires a subcommand: set")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "set":
		return projectSet(rest, os.Stdin, os.Stdout, terminalInput(os.Stdin))
	default:
		return fmt.Errorf("unknown project subcommand %q; expected set", sub)
	}
}

// repeatedFlag collects every occurrence of a flag. Keys hold spaces and
// values may hold commas, so nothing is split.
type repeatedFlag []string

func (r *repeatedFlag) String() string     { return strings.Join(*r, "; ") }
func (r *repeatedFlag) Set(v string) error { *r = append(*r, v); return nil }

func projectSet(args []string, in io.Reader, out io.Writer, interactive bool) error {
	fs := flag.NewFlagSet("project set", flag.ContinueOnError)
	fs.SetOutput(out)
	var sets, unsets repeatedFlag
	project := fs.String("project", "", "directory inside the repository (default: the current directory)")
	dry := fs.Bool("dry-run", false, "preview only; do not change anything")
	fs.Var(&sets, "set", "'Key: value' to set; repeat for several keys")
	fs.Var(&unsets, "unset", "key to remove; repeat for several keys")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	if len(sets) == 0 && len(unsets) == 0 {
		return fmt.Errorf("project set needs at least one --set or --unset")
	}
	var items []hiveItem
	for _, raw := range sets {
		key, value, ok := strings.Cut(raw, ":")
		if key = strings.TrimSpace(key); !ok || key == "" {
			return fmt.Errorf("--set expects 'Key: value', got %q", shown(raw))
		}
		items = append(items, hiveItem{Key: key, Value: strings.TrimSpace(value)})
	}
	keys := make([]string, len(unsets))
	for i, k := range unsets {
		keys[i] = strings.TrimSpace(k)
	}
	edit, err := prepareProjectEdit(*project, items, keys, newGitRunner())
	if err != nil {
		return err
	}
	unchanged := edit.Existed && string(edit.Before) == string(edit.After)
	showProjectSummary(out, edit, unchanged)
	if *dry {
		fmt.Fprintln(out, "Preview: nothing was changed.")
		return nil
	}
	if unchanged {
		fmt.Fprintln(out, "Nothing to change: the section already matches.")
		return nil
	}
	if !interactive {
		return fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to preview")
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	decision, err := terminal.Confirm("Write "+projectFileName+"?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	if err := writeProjectFile(edit.File, edit.Before, edit.Existed, edit.After); err != nil {
		return err
	}
	verb := "Updated"
	if !edit.Existed {
		verb = "Created"
	}
	fmt.Fprintf(out, "%s %s\n", verb, sanitizeLine(edit.File))
	return nil
}

// showProjectSummary prints the file, the section before and after, and the
// warnings. Everything read from the file is sanitized for display.
func showProjectSummary(out io.Writer, e projectEdit, unchanged bool) {
	verb := "update"
	if !e.Existed {
		verb = "create"
	}
	if unchanged {
		verb = "keep"
	}
	fmt.Fprintf(out, "File: %s (%s)\n", sanitizeLine(e.File), verb)
	fmt.Fprintln(out, "Before:")
	for _, l := range hiveSectionLines(e.Before) {
		fmt.Fprintf(out, "  %s\n", l)
	}
	fmt.Fprintln(out, "After:")
	for _, l := range hiveSectionLines(e.After) {
		fmt.Fprintf(out, "  %s\n", l)
	}
	warnings := e.Warnings
	if e.Dirty {
		warnings = append(append([]string(nil), warnings...), "AGENTS.md has uncommitted changes")
	}
	if len(warnings) > 0 {
		fmt.Fprintln(out, "Warnings:")
		for _, w := range warnings {
			fmt.Fprintf(out, "  %s\n", w)
		}
	}
}
