// doctor_project.go builds the Project section: it validates the `## Hive`
// section of the AGENTS.md at the root of the Git repository that contains a
// directory (design.md "Reglas por sección"). It only reads; the only
// processes it starts are fixed `git` argument vectors without a shell.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// hiveKey is one setting `## Hive` may hold.
type hiveKey struct {
	Name     string
	Required bool
}

// hiveSettingKeys is the table of known settings, in display order. The
// canonical definition is «Project settings» in content/guidance/global.md, and
// a test keeps the two in step.
var hiveSettingKeys = []hiveKey{
	{"Project", true},
	{"Base branch", true},
	{"Tracker", true},
	{"Specs", true},
	{"Environments", false},
	{"Review", false},
	{"Delivery", false},
	{"Hive guidance", false},
}

const (
	hiveFileLimit     = 1 << 20 // AGENTS.md is read up to this many bytes
	hiveValueRunes    = 200     // longest value shown
	gitTimeout        = 5 * time.Second
	gitWaitDelay      = time.Second
	gitMaxOutput      = 64 << 10
	outsideGitLine    = "Not inside a Git repository; run hive from a repository."
	outsideGitLineTwo = "Workspace-level ## Hive sections are not checked."
	hiveHeading       = "## Hive"
	hiveDeliveryWant  = "direct-base"
	hiveGuidanceWant  = "required"
)

// gitRunner runs `git -C dir args...` and returns its standard output. A
// non-zero exit is an *exec.ExitError; any other error means git could not be
// run at all.
type gitRunner func(dir string, args ...string) (string, error)

// newGitRunner returns the production runner. It finds git in PATH and
// computes the environment of filteredGitEnv once, on first use, so an
// inherited GIT_DIR cannot override -C. Each call has a time limit and a
// capped output; the arguments are never interpreted by a shell.
func newGitRunner() gitRunner {
	var once sync.Once
	var gitPath string
	var env []string
	var setupErr error
	return func(dir string, args ...string) (string, error) {
		once.Do(func() {
			if gitPath, setupErr = exec.LookPath("git"); setupErr != nil {
				return
			}
			env, setupErr = filteredGitEnv(gitPath)
		})
		if setupErr != nil {
			return "", setupErr
		}
		ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, gitPath, append([]string{"-C", dir}, args...)...)
		cmd.Env = append(append([]string(nil), env...), "GIT_OPTIONAL_LOCKS=0")
		cmd.WaitDelay = gitWaitDelay
		out := &cappedBuffer{limit: gitMaxOutput}
		cmd.Stdout = out
		err := cmd.Run()
		if ctx.Err() != nil {
			return "", fmt.Errorf("git timed out after %s", gitTimeout)
		}
		if err != nil && !errors.Is(err, exec.ErrWaitDelay) {
			return "", err
		}
		return out.buf.String(), nil
	}
}

// gitAnswered reports whether err is git's own negative answer (a non-zero
// exit), as opposed to a failure to run git.
func gitAnswered(err error) bool {
	var exit *exec.ExitError
	return errors.As(err, &exit)
}

// projectCheck is the outcome of checking one directory: the section plus what
// the Project view shows above it.
type projectCheck struct {
	Dir     string // the directory checked, absolute
	Root    string // the repository root; "" outside a repository
	File    string // path of AGENTS.md; "" outside a repository
	Section doctorSection
}

// collectProject builds the Project section for the directory project ("" means
// the current directory). deps arrives already adapted to the options; git is
// not one of its members because the validation runs `git` only against the
// chosen repository, whatever a synthetic --home says about CLIs.
func collectProject(project string, deps doctorDeps) doctorSection {
	return checkProject(project, newGitRunner()).sectionWithLocation()
}

// sectionWithLocation is the section with the location line first. The Project
// view shows this same line at the top of its scrolling text.
func (c projectCheck) sectionWithLocation() doctorSection {
	sec := c.Section
	if label, path := c.location(); path != "" {
		sec.Lines = append([]string{label + path}, sec.Lines...)
	}
	return sec
}

// location names what was checked: the AGENTS.md path, or the directory when
// it is not inside a repository.
func (c projectCheck) location() (label, path string) {
	switch {
	case c.File != "":
		return "", sanitizeLine(c.File)
	case c.Dir != "":
		return "Directory: ", sanitizeLine(c.Dir)
	}
	return "", ""
}

func checkProject(project string, git gitRunner) projectCheck {
	res := projectCheck{Section: doctorSection{Title: "Project"}}
	fail := func(format string, a ...any) projectCheck {
		res.Section.Err = sanitizeLine(fmt.Sprintf(format, a...))
		return res
	}
	dir := project
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fail("cannot read the current directory: %v", err)
		}
		dir = wd
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return fail("cannot resolve %s: %v", dir, err)
	}

	// git -C on a missing directory exits like a directory outside a
	// repository, so look at the directory first and say what is wrong.
	switch info, err := os.Stat(dir); {
	case errors.Is(err, os.ErrNotExist):
		res.Section.Lines = []string{"Directory not found: " + sanitizeLine(dir)}
		return res
	case err != nil:
		return fail("cannot read %s: %v", dir, err)
	case !info.IsDir():
		res.Section.Lines = []string{"Not a directory: " + sanitizeLine(dir)}
		return res
	}
	res.Dir = dir

	out, err := git(dir, "rev-parse", "--show-toplevel")
	root := strings.TrimRight(out, "\r\n")
	switch {
	case err != nil && !gitAnswered(err):
		return fail("cannot run git: %v", err)
	case err != nil || root == "":
		res.Section.Lines = []string{outsideGitLine, outsideGitLineTwo}
		return res
	}
	res.Root = root
	res.File = filepath.Join(root, "AGENTS.md")

	info, err := os.Stat(res.File)
	switch {
	case errors.Is(err, os.ErrNotExist):
		res.Section.Lines = []string{"AGENTS.md: not found at the repository root", addHiveSectionLine()}
		return res
	case err != nil:
		return fail("cannot read AGENTS.md: %v", err)
	case !info.Mode().IsRegular():
		res.Section.Lines = []string{"AGENTS.md: not a regular file"}
		return res
	}
	f, err := os.Open(res.File)
	if err != nil {
		return fail("cannot read AGENTS.md: %v", err)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, hiveFileLimit+1))
	if err != nil {
		return fail("cannot read AGENTS.md: %v", err)
	}
	truncated := len(data) > hiveFileLimit
	if truncated {
		data = data[:hiveFileLimit]
	}
	lines := validateHiveSection(root, string(data), git)
	if truncated {
		lines = append([]string{"AGENTS.md: larger than 1 MiB; only the first 1 MiB was read"}, dropValid(lines)...)
	}
	res.Section.Lines = lines
	return res
}

// dropValid removes the "Valid" verdict, which a truncated read cannot claim.
func dropValid(lines []string) []string {
	if len(lines) > 0 && lines[0] == "Valid" {
		return lines[1:]
	}
	return lines
}

// hiveItem is one `- Key: value` line of the section.
type hiveItem struct{ Key, Value string }

// hiveWalk is the result of walking a document for its `## Hive` section.
// Positions are indexes into strings.Split(content, "\n"), so the validator
// and the editor read the same lines.
type hiveWalk struct {
	Items    []hiveItem // the `- Key: value` lines of the first section
	Lines    []int      // Lines[i] is the line index of Items[i]
	Headings int        // how many exact `## Hive` headings the document has
	Heading  int        // line index of the first heading; -1 without one
	End      int        // line index of the heading that ends the first section, or the line count
}

// walkHiveSection finds the exact `## Hive` headings and returns the items of
// the first one: the `- Key: value` lines until the next heading. Lines inside
// fenced code blocks are neither headings nor items, so an example of the
// section in documentation is not mistaken for the section.
func walkHiveSection(content string) hiveWalk {
	lines := strings.Split(content, "\n")
	w := hiveWalk{Heading: -1, End: len(lines)}
	fenced, in := false, false
	for i, raw := range lines {
		line := strings.TrimRight(raw, " \t\r")
		if trimmed := strings.TrimLeft(line, " "); strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if in && w.Headings == 1 {
				w.End = i
			}
			in = line == hiveHeading
			if in {
				w.Headings++
				if w.Headings == 1 {
					w.Heading = i
				}
			}
			continue
		}
		if !in || w.Headings != 1 {
			continue
		}
		rest, ok := strings.CutPrefix(line, "- ")
		if !ok {
			continue
		}
		key, value, ok := strings.Cut(rest, ":")
		if key = strings.TrimSpace(key); !ok || key == "" {
			continue
		}
		w.Items = append(w.Items, hiveItem{Key: key, Value: strings.TrimSpace(value)})
		w.Lines = append(w.Lines, i)
	}
	return w
}

// parseHiveSection returns the items of the first `## Hive` section and how
// many such headings the content holds.
func parseHiveSection(content string) (items []hiveItem, headings int) {
	w := walkHiveSection(content)
	return w.Items, w.Headings
}

// hiveFinding is one validation finding, in display order.
type hiveFinding struct {
	Text     string
	Blocking bool // a write must not proceed; otherwise it only warns
}

// checkHiveText validates a section's items without reading the file: it
// returns the findings that must stop a write apart from those that only
// warn. A missing or invalid required setting, an invalid Delivery or Hive
// guidance, and a branch name git rejects block; a branch or Specs path that
// does not exist, and an unknown key, warn.
func checkHiveText(items []hiveItem, root string, git gitRunner) (blocking, warnings []string) {
	for _, f := range hiveFindings(items, root, git) {
		if f.Blocking {
			blocking = append(blocking, f.Text)
		} else {
			warnings = append(warnings, f.Text)
		}
	}
	return blocking, warnings
}

// hiveFindings is checkHiveText in the order the doctor shows it.
func hiveFindings(items []hiveItem, root string, git gitRunner) []hiveFinding {
	values := map[string]string{}
	for _, it := range items {
		if _, dup := values[it.Key]; !dup {
			values[it.Key] = it.Value
		}
	}
	var findings []hiveFinding
	add := func(blocking bool, key, format string, a ...any) {
		findings = append(findings, hiveFinding{Text: key + ": " + fmt.Sprintf(format, a...), Blocking: blocking})
	}
	for _, k := range hiveSettingKeys {
		v, present := values[k.Name]
		switch {
		case k.Required && !present:
			add(true, k.Name, "required value is missing")
			continue
		case k.Required && v == "":
			add(true, k.Name, "value is empty")
			continue
		}
		switch k.Name {
		case "Specs":
			path, _, _ := strings.Cut(v, " · ")
			path = strings.TrimSpace(path)
			full := path
			if !filepath.IsAbs(full) {
				full = filepath.Join(root, full)
			}
			if info, err := os.Stat(full); err != nil || !info.IsDir() {
				add(false, k.Name, "%s is not an existing directory", shown(path))
			}
		case "Base branch":
			branch := strings.Fields(v)[0]
			if problem := checkBaseBranch(root, branch, git); problem != "" {
				add(problem != branchAbsent, k.Name, "%s %s", shown(branch), problem)
			}
		case "Delivery":
			if present && v != hiveDeliveryWant {
				add(true, k.Name, "expected %q, found %q", hiveDeliveryWant, shown(v))
			}
		case "Hive guidance":
			if present && v != hiveGuidanceWant {
				add(true, k.Name, "expected %q, found %q", hiveGuidanceWant, shown(v))
			}
		}
	}
	reported := map[string]bool{}
	for _, it := range items {
		if !knownHiveKey(it.Key) && !reported[it.Key] {
			reported[it.Key] = true
			add(false, truncateRunes(sanitizeLine(it.Key), 60), "unknown key")
		}
	}
	return findings
}

// validateHiveSection checks the section in content, the text of the AGENTS.md
// at root, and returns the lines to show: one finding per line, each naming
// its key, or "Valid"; then the values read.
func validateHiveSection(root, content string, git gitRunner) []string {
	items, headings := parseHiveSection(content)
	switch {
	case headings == 0:
		return []string{hiveHeading + ": section not found in AGENTS.md", addHiveSectionLine()}
	case headings > 1:
		return []string{fmt.Sprintf("%s: section appears %d times; keep one", hiveHeading, headings)}
	}
	var findings []string
	unknown := false
	for _, f := range hiveFindings(items, root, git) {
		findings = append(findings, f.Text)
	}
	for _, it := range items {
		unknown = unknown || !knownHiveKey(it.Key)
	}
	if unknown {
		findings = append(findings, "Known keys: "+strings.Join(hiveKeyNames(false), ", "))
	}
	lines := findings
	if len(lines) == 0 {
		lines = []string{"Valid"}
	}
	if len(items) > 0 {
		lines = append(lines, "Values read:")
		for _, it := range items {
			lines = append(lines, "  "+truncateRunes(sanitizeLine(it.Key), 60)+": "+shown(it.Value))
		}
	}
	return lines
}

// hiveKeyNames lists the names of hiveSettingKeys in display order: only the
// required ones when requiredOnly, every known key otherwise.
func hiveKeyNames(requiredOnly bool) []string {
	var names []string
	for _, k := range hiveSettingKeys {
		if k.Required || !requiredOnly {
			names = append(names, k.Name)
		}
	}
	return names
}

// addHiveSectionLine says what to add when AGENTS.md has no usable section.
func addHiveSectionLine() string {
	return "Add a " + hiveHeading + " section with: " + strings.Join(hiveKeyNames(true), ", ")
}

func knownHiveKey(name string) bool {
	for _, k := range hiveSettingKeys {
		if k.Name == name {
			return true
		}
	}
	return false
}

// shown makes text from AGENTS.md safe and short enough to display.
func shown(s string) string { return truncateRunes(sanitizeLine(s), hiveValueRunes) }

// branchAbsent is the checkBaseBranch problem for a well-formed name that is
// neither a local branch nor on origin; a write only warns about it.
const branchAbsent = "is not a local branch or on origin"

// checkBaseBranch returns why branch is not usable as the base branch, or "".
// A name that starts with "-" or holds "@{" is rejected without running git:
// the first would be read as an option and the second expands through the
// reflog. Otherwise git validates the name, then looks for it as a local
// branch and as a branch of origin.
func checkBaseBranch(root, branch string, git gitRunner) string {
	const invalid = "is not a valid branch name"
	if strings.HasPrefix(branch, "-") || strings.Contains(branch, "@{") || sanitizeLine(branch) != branch {
		return invalid
	}
	if _, err := git(root, "check-ref-format", "--branch", branch); err != nil {
		if gitAnswered(err) {
			return invalid
		}
		return "could not be checked: " + shown(err.Error())
	}
	for _, ref := range []string{"refs/heads/" + branch, "refs/remotes/origin/" + branch} {
		_, err := git(root, "show-ref", "--verify", "--quiet", ref)
		switch {
		case err == nil:
			return ""
		case !gitAnswered(err):
			return "could not be checked: " + shown(err.Error())
		}
	}
	return branchAbsent
}
