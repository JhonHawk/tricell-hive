// hive voice manages the optional voice layer's managed block (T4). Its
// contract is design.md "Operaciones": list reads the catalogue read-only;
// set and off build a plan through management.BuildVoicePlan and apply it
// with the same summary/confirm/--dry-run/--out pattern hive update uses
// (update.go).
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
	"tricell-hive/tooling/management"
)

func runVoice(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("voice requires a subcommand: list, set or off")
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "list":
		return voiceList(rest, os.Stdout)
	case "set":
		return voiceSet(rest, os.Stdin, os.Stdout, terminalInput(os.Stdin))
	case "off":
		return voiceOff(rest, os.Stdin, os.Stdout, terminalInput(os.Stdin))
	default:
		return fmt.Errorf("unknown voice subcommand %q; expected list, set or off", sub)
	}
}

func voiceList(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("voice list", flag.ContinueOnError)
	fs.SetOutput(out)
	source := fs.String("source", ".", "this distribution content")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("voice list accepts no positional arguments")
	}
	voices, err := management.ListVoices(*source)
	if err != nil {
		return err
	}
	for _, v := range voices {
		fmt.Fprintf(out, "%s: %s\n", v.ID, voiceDisplayDescription(v.Description))
	}
	return nil
}

// voiceDisplayDescription drops a leading "Name: " duplicate a voice file's
// first line commonly carries (see content/voices/jarvis.md), so the list
// does not read "jarvis: Jarvis: warm, formal...". Only a short, plain
// leading label is treated as a name; a description that instead starts
// with an ordinary sentence containing an early colon is left untouched.
func voiceDisplayDescription(description string) string {
	name, rest, ok := strings.Cut(description, ": ")
	if !ok || len(name) == 0 || len(name) > 20 || strings.ContainsAny(name, ".,;:!?") {
		return description
	}
	return rest
}

// voiceFlags holds the flags shared by "voice set" and "voice off" (design.md
// "Operaciones"); Address, Name and Intensity are meaningless for off and
// left at their zero value there.
type voiceFlags struct {
	Address, Name, Intensity, Source, Home, StateDir, Out string
	DryRun                                                bool
}

// parseVoiceFlags parses args as pure flags (no positional arguments left
// over): "voice set" extracts its <id> positional itself, args[0], before
// calling this, the same way runPlan extracts "install"/"remove" before its
// own flag.FlagSet — Go's flag package stops parsing at the first
// non-flag argument, so a positional occurring before later flags (as
// design.md's "hive voice set <id> [--address ...]" syntax requires) cannot
// share one FlagSet with them.
func parseVoiceFlags(name string, args []string, withChoice bool) (f voiceFlags, err error) {
	f.Source = "."
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	if withChoice {
		fs.StringVar(&f.Address, "address", "", "sir, name or none (default none)")
		fs.StringVar(&f.Name, "name", "", "given name; required when --address name")
		fs.StringVar(&f.Intensity, "intensity", "", "subtle or marked (default subtle)")
	}
	fs.StringVar(&f.Source, "source", f.Source, "this distribution content")
	fs.StringVar(&f.Home, "home", "", "explicit synthetic home; ignores host environment paths")
	fs.StringVar(&f.StateDir, "state-dir", "", "state directory (default: user Application Support/tricell-hive)")
	fs.BoolVar(&f.DryRun, "dry-run", false, "preview only; do not change anything")
	fs.StringVar(&f.Out, "out", "", "save the plan to FILE instead of applying")
	if err = fs.Parse(args); err != nil {
		return f, err
	}
	if fs.NArg() != 0 {
		return f, fmt.Errorf("unexpected positional arguments")
	}
	return f, nil
}

func voiceSet(args []string, in io.Reader, out io.Writer, interactive bool) error {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("voice set requires a voice ID")
	}
	id := args[0]
	f, err := parseVoiceFlags("voice set", args[1:], true)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	v := management.VoiceSetting{ID: id, Address: f.Address, Name: f.Name, Intensity: f.Intensity}
	o := management.Options{Source: f.Source, Home: f.Home, StateDir: f.StateDir}
	p, err := management.BuildVoicePlan("set", o, v)
	if err != nil {
		return err
	}
	return runVoicePlan(p, "Voice set", in, out, interactive, f.DryRun, f.Out)
}

func voiceOff(args []string, in io.Reader, out io.Writer, interactive bool) error {
	f, err := parseVoiceFlags("voice off", args, false)
	if err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	o := management.Options{Source: f.Source, Home: f.Home, StateDir: f.StateDir}
	p, err := management.BuildVoicePlan("off", o, management.VoiceSetting{})
	if err != nil {
		return err
	}
	return runVoicePlan(p, "Voice turned off", in, out, interactive, f.DryRun, f.Out)
}

// runVoicePlan shows the summary, then applies the same
// unchanged/dry-run/out/interactive-confirm sequence hive update uses
// (update.go's update function): unchanged is reported and the command
// exits 0 without writing anything, since a voice plan carries no source
// commit to record the way an install/update plan does.
func runVoicePlan(p management.Plan, doneVerb string, in io.Reader, out io.Writer, interactive, dry bool, outFile string) error {
	unchanged, err := management.PlanUnchanged(p)
	if err != nil {
		return err
	}
	showVoiceSummary(out, p, unchanged)
	if dry {
		fmt.Fprintln(out, "Preview: nothing was changed.")
		return nil
	}
	if outFile != "" {
		if err := management.SavePlan(outFile, p); err != nil {
			return err
		}
		fmt.Fprintf(out, "Plan saved to %s; run hive apply --plan %s to apply it.\n", outFile, outFile)
		return nil
	}
	if unchanged {
		fmt.Fprintln(out, "Nothing to change: the voice already matches.")
		return nil
	}
	if !interactive {
		return fmt.Errorf("an interactive terminal is required to confirm; use --dry-run to preview or --out FILE to save a plan for hive apply")
	}
	terminal := installTerminal{reader: bufio.NewReader(in), out: out, interactive: interactive}
	decision, err := confirmInstall(terminal, "Apply these changes?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	result, err := (management.Engine{}).Apply(p)
	if err != nil {
		return err
	}
	reportApplyResult(out, doneVerb, result)
	return nil
}

// showVoiceSummary reads the choice from p.VoiceSetting, already normalized
// by management.BuildVoicePlan (nil for "voice off"), rather than
// re-deriving or re-defaulting it here.
func showVoiceSummary(out io.Writer, p management.Plan, unchanged bool) {
	if p.VoiceSetting != nil {
		fmt.Fprintf(out, "Voice: %s (address %s, intensity %s)\n", p.VoiceSetting.ID, p.VoiceSetting.Address, p.VoiceSetting.Intensity)
	} else {
		fmt.Fprintln(out, "Voice: off")
	}
	if unchanged {
		fmt.Fprintln(out, "No files need to change.")
		return
	}
	// A voice plan holds one entry per file, including files already at the
	// chosen voice; only the entries whose span changes are reported.
	paths := make([]string, 0, len(p.Voice))
	for _, vc := range p.Voice {
		if !sameVoiceSpan(vc.Before, vc.After) {
			paths = append(paths, vc.Path)
		}
	}
	fmt.Fprintf(out, "Voice files to change: %d\n", len(paths))
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Fprintf(out, "  %s\n", path)
	}
}

func sameVoiceSpan(a, b *management.VoiceSpan) bool {
	if a == nil || b == nil {
		return a == b
	}
	return bytes.Equal(a.Managed, b.Managed) && a.SourceHash == b.SourceHash && slices.Equal(a.Consumers, b.Consumers)
}
