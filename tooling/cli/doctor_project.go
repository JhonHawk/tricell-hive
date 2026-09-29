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
	return checkProject(project, newGitRunner()).Section
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
		res.Section.Lines = []string{"AGENTS.md: not found at the repository root"}
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

// parseHiveSection finds the exact `## Hive` headings and returns the items of
// the first one: the `- Key: value` lines until the next heading. Lines inside
// fenced code blocks are neither headings nor items, so an example of the
// section in documentation is not mistaken for the section.
func parseHiveSection(content string) (items []hiveItem, headings int) {
	fenced, in := false, false
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimRight(raw, " \t\r")
		if trimmed := strings.TrimLeft(line, " "); strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
			continue
		}
		if fenced {
			continue
		}
		if strings.HasPrefix(line, "#") {
			in = line == hiveHeading
			if in {
				headings++
			}
			continue
		}
		if !in || headings != 1 {
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
		items = append(items, hiveItem{Key: key, Value: strings.TrimSpace(value)})
	}
	return items, headings
}

// validateHiveSection checks the section in content, the text of the AGENTS.md
// at root, and returns the lines to show: one finding per line, each naming
// its key, or "Valid"; then the values read.
func validateHiveSection(root, content string, git gitRunner) []string {
	items, headings := parseHiveSection(content)
	switch {
	case headings == 0:
		return []string{hiveHeading + ": section not found in AGENTS.md"}
	case headings > 1:
		return []string{fmt.Sprintf("%s: section appears %d times; keep one", hiveHeading, headings)}
	}
	values := map[string]string{}
	for _, it := range items {
		if _, dup := values[it.Key]; !dup {
			values[it.Key] = it.Value
		}
	}
	var findings []string
	add := func(key, format string, a ...any) {
		findings = append(findings, key+": "+fmt.Sprintf(format, a...))
	}
	for _, k := range hiveSettingKeys {
		v, present := values[k.Name]
		switch {
		case k.Required && !present:
			add(k.Name, "required value is missing")
			continue
		case k.Required && v == "":
			add(k.Name, "value is empty")
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
				add(k.Name, "%s is not an existing directory", shown(path))
			}
		case "Base branch":
			branch := strings.Fields(v)[0]
			if problem := checkBaseBranch(root, branch, git); problem != "" {
				add(k.Name, "%s %s", shown(branch), problem)
			}
		case "Delivery":
			if present && v != hiveDeliveryWant {
				add(k.Name, "expected %q, found %q", hiveDeliveryWant, shown(v))
			}
		case "Hive guidance":
			if present && v != hiveGuidanceWant {
				add(k.Name, "expected %q, found %q", hiveGuidanceWant, shown(v))
			}
		}
	}
	reported := map[string]bool{}
	for _, it := range items {
		if !knownHiveKey(it.Key) && !reported[it.Key] {
			reported[it.Key] = true
			add(truncateRunes(sanitizeLine(it.Key), 60), "unknown key")
		}
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
	return "is not a local branch or on origin"
}
