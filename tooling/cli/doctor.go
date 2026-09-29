// doctor.go collects the read-only diagnostics behind the Diagnostics view and
// the future `hive doctor` command: which CLIs are installed and at what
// version, whether Hive's own files are healthy, and which open sessions
// started before the installed release. It never writes, and it only executes
// `<cli> --version` for a detected CLI (design.md "Reglas por sección").
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/charmbracelet/x/ansi"

	"tricell-hive/tooling/management"
)

// doctorDeps is everything the diagnostics take from the outside world, so a
// test replaces all of it and never reads the developer's configuration.
type doctorDeps struct {
	lookPath     func(name string) (string, error)
	runVersion   func(bin string) (string, error)
	processAlive func(pid int) bool
	now          func() time.Time
	userHome     func() (string, error)
	getenv       func(key string) string
}

// versionTimeout bounds `<cli> --version`; versionWaitDelay bounds how long a
// grandchild that keeps the output open can extend it. Tests do not change them.
const (
	versionTimeout   = 3 * time.Second
	versionWaitDelay = time.Second
	versionMaxOutput = 64 << 10
	versionMaxRunes  = 80
)

// realDoctorDeps is the production wiring.
func realDoctorDeps() doctorDeps {
	return doctorDeps{
		lookPath:     exec.LookPath,
		runVersion:   runVersionCommand,
		processAlive: processAlive,
		now:          time.Now,
		userHome:     os.UserHomeDir,
		getenv:       os.Getenv,
	}
}

// forOptions adapts the dependencies to the options: under a synthetic --home
// nothing is detected or executed and no environment variable is read, as in
// detectInstallerHosts and setup, and the home is the synthetic one.
func (d doctorDeps) forOptions(o management.Options) doctorDeps {
	if o.Home == "" {
		return d
	}
	home := o.Home
	if abs, err := filepath.Abs(home); err == nil {
		home = abs
	}
	d.lookPath = func(name string) (string, error) {
		return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
	}
	d.runVersion = func(string) (string, error) { return "", errors.New("not executed under a synthetic home") }
	d.getenv = func(string) string { return "" }
	d.userHome = func() (string, error) { return home, nil }
	return d
}

// runVersionCommand runs `bin --version` without a shell, with a time limit,
// and returns at most versionMaxOutput bytes of its combined output.
func runVersionCommand(bin string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--version")
	cmd.WaitDelay = versionWaitDelay
	out := &cappedBuffer{limit: versionMaxOutput}
	cmd.Stdout, cmd.Stderr = out, out
	err := cmd.Run()
	if ctx.Err() != nil {
		return "", fmt.Errorf("timed out after %s", versionTimeout)
	}
	if err != nil && !errors.Is(err, exec.ErrWaitDelay) {
		return "", err
	}
	return out.buf.String(), nil
}

// cappedBuffer keeps the first limit bytes and drops the rest, so a program
// that prints without end cannot grow memory.
type cappedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	if room := c.limit - c.buf.Len(); room > 0 {
		c.buf.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}

// processAlive reports whether pid names a running process: signal 0 succeeds,
// or fails with EPERM when the process belongs to another user. A pid of zero
// or less is never alive here, because kill(0) and kill(-1) address groups.
func processAlive(pid int) bool {
	if pid <= 0 || pid > 1<<31-1 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// noHostsText is the empty state of every view and command that needs a
// registered CLI: it says what happened and the first action.
const noHostsText = "No CLI hosts are registered. Open CLIs from the menu, or run hive install, to install Hive."

// unreadableStateText explains a state that could not be read: first in plain
// words with the way out, then the technical error after a line break, so a
// renderer can put it on its own line. stateDir is "" when it is not known.
func unreadableStateText(stateDir string, err error) string {
	where := "Hive's state"
	if stateDir != "" {
		where += " in " + sanitizeLine(stateDir)
	}
	return where + " could not be read. Repair or restore its files; hive status reports the same problem.\nDetail: " + sanitizeLine(err.Error())
}

// stateDirOf resolves the state directory of the options for a message, or ""
// when it cannot be resolved.
func stateDirOf(o management.Options) string {
	if o.Scope == "" {
		o.Scope = "user"
	}
	if _, dir, err := management.NormalizeOptions(o); err == nil {
		return dir
	}
	return ""
}

// errLines splits a section's Err into its lines: the first says what failed,
// the rest are detail.
func errLines(err string) (first string, rest []string) {
	lines := strings.Split(err, "\n")
	return lines[0], lines[1:]
}

// doctorSection is one titled block of the report. Lines are already
// sanitized. Err is set when the section could not be fully checked, and the
// view then offers to retry. Err may hold several lines: the first says what
// failed in plain words, the others are technical detail.
type doctorSection struct {
	Title string
	Lines []string
	Err   string
}

// doctorReport holds the five sections in display order. Integrations and
// Project are filled by doctor_integrations.go and doctor_project.go.
type doctorReport struct {
	CLIs, Installation, Sessions, Integrations, Project doctorSection
}

// sanitizeLine removes terminal escape sequences and control characters from
// text that came from outside Hive (versions, working directories, paths), so
// a foreign file cannot alter the terminal. Tabs and line breaks become one
// space; invalid UTF-8 and bidirectional overrides are dropped.
func sanitizeLine(s string) string {
	s = strings.ToValidUTF8(ansi.Strip(s), "")
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\t' || r == '\n' || r == '\r':
			b.WriteRune(' ')
		case unicode.IsControl(r):
		case r == '‎' || r == '‏' || (r >= '‪' && r <= '‮') || (r >= '⁦' && r <= '⁩'):
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collectDoctor builds the five sections. An error in one section stays in
// that section; the others are still produced.
func collectDoctor(o management.Options, project string, deps doctorDeps) doctorReport {
	deps = deps.forOptions(o)
	var r doctorReport
	r.CLIs, r.Installation, r.Sessions = collectHostSections(o, deps)
	r.Integrations = collectIntegrations(o, deps)
	r.Project = collectProject(project, deps)
	return r
}

// collectHostSections builds CLIs, Installation and Sessions, the three
// sections that depend on the installation state. The Diagnostics view calls
// it directly, so opening it does not also run the Integrations and Project
// checks.
func collectHostSections(o management.Options, deps doctorDeps) (clis, installation, sessions doctorSection) {
	deps = deps.forOptions(o)
	if o.Scope == "" {
		o.Scope = "user"
	}
	st := loadDoctorState(o)
	return cliSection(deps, st), installationSection(o, st), sessionsSection(deps, st)
}

// doctorState is what the three sections share: the registered CLIs, their
// Status rows, the pending operation, and each retained release's write time.
type doctorState struct {
	stateDir   string
	registered []string
	entries    []management.StatusEntry
	pending    management.PendingKind
	released   map[string]time.Time
	err        error // the state could not be read at all
	warn       []string
}

func loadDoctorState(o management.Options) doctorState {
	st := doctorState{released: map[string]time.Time{}}
	_, dir, err := management.NormalizeOptions(o)
	if err != nil {
		st.err = err
		return st
	}
	o.StateDir = dir
	st.stateDir = dir
	if st.registered, err = management.RegisteredHosts(o); err != nil {
		st.err = err
		return st
	}
	// An interrupted first install leaves an operation pending and no
	// registered CLI, so the pending check does not wait for the CLIs.
	if st.pending, err = management.Pending(dir); err != nil {
		st.warn = append(st.warn, "Cannot check for an unfinished operation: "+sanitizeLine(err.Error()))
	}
	if len(st.registered) == 0 {
		return st
	}
	so := o
	so.Hosts = st.registered
	if st.entries, err = management.Status(so); err != nil {
		st.err = err
		return st
	}
	if releases, err := management.Releases(o); err != nil {
		st.warn = append(st.warn, "Cannot read the retained releases: "+sanitizeLine(err.Error()))
	} else {
		for _, rel := range releases {
			if t, err := time.Parse(time.RFC3339, rel.LastWrittenAt); err == nil {
				st.released[rel.ID] = t
			}
		}
	}
	return st
}

func (st doctorState) isRegistered(host string) bool {
	for _, h := range st.registered {
		if h == host {
			return true
		}
	}
	return false
}

// hostInstall is what the Status rows say about one registered CLI.
type hostInstall struct {
	Release string // full release ID, "" when unknown
	State   string // verified, partial, drift or legacy; "" when unknown
}

// hostInstallation reads a CLI's installed release and state from its rows. A
// row the CLI owns (installed, drift or recovery_required) wins over a row it
// only shares; among several releases the newest write time wins.
func (st doctorState) hostInstallation(host string) hostInstall {
	var own, shared hostInstall
	for _, e := range st.entries {
		if e.Host != host || e.Kind == "voice" || e.Release == "" {
			continue
		}
		cand := hostInstall{Release: e.Release, State: e.VersionStatus}
		switch e.Status {
		case "installed", "drift", "recovery_required":
			if own.Release == "" || st.released[cand.Release].After(st.released[own.Release]) {
				own = cand
			}
		default:
			if shared.Release == "" {
				shared = cand
			}
		}
	}
	if own.Release != "" {
		return own
	}
	return shared
}

// ---------------------------------------------------------------------------
// CLIs.
// ---------------------------------------------------------------------------

// versionBinary is the executable that reports a CLI's version.
func versionBinary(host string) string {
	if host == "cursor" {
		return "cursor-agent"
	}
	return host
}

func cliSection(deps doctorDeps, st doctorState) doctorSection {
	sec := doctorSection{Title: "CLIs"}
	versions := make([]string, len(installerHosts))
	detected := make([]bool, len(installerHosts))
	var wg sync.WaitGroup
	for i, host := range installerHosts {
		if _, err := deps.lookPath(host); err != nil {
			continue
		}
		detected[i] = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			versions[i] = cliVersion(deps, host)
		}()
	}
	wg.Wait()
	for i, host := range installerHosts {
		parts := []string{fmt.Sprintf("%-8s", host)}
		if detected[i] {
			parts = append(parts, "detected")
		} else {
			parts = append(parts, "not detected")
		}
		switch {
		case st.err != nil:
			parts = append(parts, "release unknown")
		case !st.isRegistered(host):
			parts = append(parts, "not installed by Hive")
		default:
			inst := st.hostInstallation(host)
			switch {
			case inst.Release == "":
				parts = append(parts, "release -")
			case inst.State == "":
				parts = append(parts, "release "+shortHash(inst.Release))
			default:
				parts = append(parts, fmt.Sprintf("release %s (%s)", shortHash(inst.Release), inst.State))
			}
		}
		if detected[i] {
			parts = append(parts, versions[i])
		}
		sec.Lines = append(sec.Lines, strings.Join(parts, "  "))
	}
	if st.err != nil {
		sec.Err = unreadableStateText(st.stateDir, st.err)
	}
	return sec
}

// cliVersion returns the text after the CLI's name in its row: its version, or
// why that is unavailable.
func cliVersion(deps doctorDeps, host string) string {
	bin := versionBinary(host)
	path, err := deps.lookPath(bin)
	if err != nil {
		return "version unavailable: " + bin + " not found"
	}
	out, err := deps.runVersion(path)
	if err != nil {
		return "version unavailable: " + truncateRunes(sanitizeLine(err.Error()), versionMaxRunes)
	}
	line, _, _ := strings.Cut(strings.TrimSpace(out), "\n")
	line = strings.TrimSpace(sanitizeLine(line))
	if line == "" {
		return "version unavailable: no output"
	}
	return "version " + truncateRunes(line, versionMaxRunes)
}

// ---------------------------------------------------------------------------
// Installation.
// ---------------------------------------------------------------------------

// installationPhrases is the fixed English sentence for each Status state a
// person needs to act on. installed, retained_shared and not_installed are
// healthy and never listed; an unknown state shows its code.
var installationPhrases = map[string]string{
	"drift":                  "The installed file was changed or cannot be read",
	"shadowed":               "Hive recorded this file, but the current layout no longer points to it",
	"unowned":                "A file exists that Hive did not write",
	"unowned_or_conflicting": "Hive markers are missing, duplicated or broken",
	"migration_required":     "A legacy layout was found and must be migrated by installing again",
	"recovery_required":      "An unfinished operation must be recovered",
}

func installationSection(o management.Options, st doctorState) doctorSection {
	sec := doctorSection{Title: "Installation"}
	if st.err != nil {
		sec.Err = unreadableStateText(st.stateDir, st.err)
		return sec
	}
	if len(st.registered) == 0 {
		sec.Lines = []string{noHostsText}
	}
	for _, e := range st.entries {
		switch e.Status {
		case "installed", "retained_shared", "not_installed":
			continue
		case "recovery_required":
			if st.pending != management.PendingNone {
				continue
			}
		}
		phrase, ok := installationPhrases[e.Status]
		if !ok {
			phrase = "Unknown state " + sanitizeLine(e.Status)
		}
		host := e.Host
		if host == "" {
			host = "shared"
		}
		sec.Lines = append(sec.Lines,
			fmt.Sprintf("%s  %s  %s", sanitizeLine(e.Status), sanitizeLine(host), phrase),
			"  "+sanitizeLine(e.Path))
	}
	if st.pending != management.PendingNone {
		sec.Lines = append(sec.Lines, "An unfinished Hive operation is pending; "+recoverCommand(o, st)+" to finish it")
	}
	sec.Lines = append(sec.Lines, st.warn...)
	if len(sec.Lines) == 0 {
		sec.Lines = []string{"No problems found"}
	}
	return sec
}

// recoverCommand names hive recover, with --state-dir when the state
// directory is not the default one a bare `hive recover` uses.
func recoverCommand(o management.Options, st doctorState) string {
	home := o.Home
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		}
	}
	if def, err := management.DefaultStateDir(home, o.Home != ""); err == nil && filepath.Clean(def) == filepath.Clean(st.stateDir) {
		return "run hive recover"
	}
	return "run hive recover --state-dir " + shellQuote(st.stateDir)
}

// ---------------------------------------------------------------------------
// Text rendering, shared by the view and the future command.
// ---------------------------------------------------------------------------

// doctorSectionsText renders sections one after another. A section with no
// title is a placeholder that has nothing to show and is skipped.
func doctorSectionsText(sections ...doctorSection) string {
	var b strings.Builder
	for _, s := range sections {
		if s.Title == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(s.Title + "\n")
		if s.Err != "" {
			first, rest := errLines(s.Err)
			b.WriteString("  Could not check everything: " + first + "\n")
			for _, l := range rest {
				b.WriteString("    " + l + "\n")
			}
		}
		for _, l := range s.Lines {
			b.WriteString("  " + l + "\n")
		}
	}
	return b.String()
}

// renderDoctorText writes every section of the report.
func renderDoctorText(report doctorReport, w io.Writer) {
	fmt.Fprint(w, doctorSectionsText(report.CLIs, report.Installation, report.Sessions, report.Integrations, report.Project))
}
