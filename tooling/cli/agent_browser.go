package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"tricell-hive/tooling/providers"
)

// agentBrowserSystem holds every probe the agent-browser capability makes of
// the machine, so tests substitute them and never run brew, npm, node or
// agent-browser. lookPath, getenv, pathExists, isFile and dirHasEntries only
// read; nodeVersion and npmPrefix run a program, so they are called only in a
// real interactive install (see agentBrowserDetector).
type agentBrowserSystem struct {
	goos, goarch      string
	lookPath          func(name string) (string, error)
	getenv            func(key string) string
	userHome          func() (string, error)
	pathExists        func(path string) bool
	isFile            func(path string) bool
	dirHasEntries     func(path string) bool
	browserCandidates func(goos string) []browserCandidate
	nodeVersion       func() (string, error)
	npmPrefix         func() (string, error)
	prefixWritable    func(prefix string) bool
}

// browserCandidate is one place the browser check looks: an absolute Path, or
// a Command looked up in PATH. The lists mirror agent-browser's own
// findSystemChrome (its postinstall script, version 0.39.0).
type browserCandidate struct{ Path, Command string }

// agentBrowserSys is the production wiring; tests replace it.
var agentBrowserSys = realAgentBrowserSystem()

// agentBrowserExec runs one validated argv. Tests replace it with a recorder.
var agentBrowserExec = runAgentBrowserCommand

const (
	agentBrowserBinary      = "agent-browser"
	agentBrowserNodeMajor   = 24
	agentBrowserPreflight   = 5 * time.Second
	agentBrowserPreflightKB = 4 << 10
)

func realAgentBrowserSystem() agentBrowserSystem {
	return agentBrowserSystem{
		goos:              runtime.GOOS,
		goarch:            runtime.GOARCH,
		lookPath:          exec.LookPath,
		getenv:            os.Getenv,
		userHome:          os.UserHomeDir,
		pathExists:        func(p string) bool { _, err := os.Stat(p); return err == nil },
		isFile:            func(p string) bool { info, err := os.Stat(p); return err == nil && info.Mode().IsRegular() },
		dirHasEntries:     dirHasEntries,
		browserCandidates: realBrowserCandidates,
		nodeVersion:       func() (string, error) { return preflightOutput("node", "--version") },
		npmPrefix:         func() (string, error) { return preflightOutput("npm", "config", "get", "prefix") },
		prefixWritable:    npmPrefixWritable,
	}
}

func realBrowserCandidates(goos string) []browserCandidate {
	switch goos {
	case "darwin":
		return []browserCandidate{
			{Path: "/Applications/Google Chrome.app"},
			{Path: "/Applications/Google Chrome Canary.app"},
			{Path: "/Applications/Chromium.app"},
		}
	case "linux":
		return []browserCandidate{
			{Command: "google-chrome"}, {Command: "google-chrome-stable"},
			{Command: "chromium-browser"}, {Command: "chromium"},
		}
	}
	return nil
}

func dirHasEntries(dir string) bool {
	entries, err := os.ReadDir(dir)
	return err == nil && len(entries) > 0
}

// preflightOutput runs a read-only version query with the bounded runner and
// returns its trimmed standard output; a warning on standard error is not part
// of the answer and is kept only for the failure message.
func preflightOutput(argv ...string) (string, error) {
	res, stdout := runBoundedCommandContext(context.Background(), argv, agentBrowserPreflight, time.Second, agentBrowserPreflightKB)
	if res.TimedOut || res.ExitCode != 0 {
		if res.Message == "" {
			res.Message = "failed"
		}
		return "", fmt.Errorf("%s: %s", argv[0], res.Message)
	}
	return strings.TrimSpace(stdout), nil
}

// npmPrefixWritable reports whether npm's global install directories under
// prefix can be written without elevated rights. A directory that does not
// exist yet is created under prefix, so the prefix itself must be writable.
func npmPrefixWritable(prefix string) bool {
	for _, dir := range []string{filepath.Join(prefix, "lib", "node_modules"), filepath.Join(prefix, "bin")} {
		target := dir
		if _, err := os.Stat(dir); err != nil {
			target = prefix
		}
		if syscall.Access(target, 2) != nil { // W_OK
			return false
		}
	}
	return true
}

// browserFound reports whether any browser agent-browser would use exists, and
// what it found. It reads PATH and files only (D-5): AGENT_BROWSER_EXECUTABLE_PATH
// pointing at a file, a system browser in the places findSystemChrome checks,
// or a browser downloaded under $HOME/.agent-browser/browsers.
func (s agentBrowserSystem) browserFound() (bool, string) {
	if path := s.getenv("AGENT_BROWSER_EXECUTABLE_PATH"); path != "" && s.isFile(path) {
		return true, path
	}
	for _, c := range s.browserCandidates(s.goos) {
		switch {
		case c.Path != "" && s.pathExists(c.Path):
			return true, c.Path
		case c.Command != "":
			if p, err := s.lookPath(c.Command); err == nil {
				return true, p
			}
		}
	}
	if home, err := s.userHome(); err == nil && home != "" {
		dir := filepath.Join(home, ".agent-browser", "browsers")
		if s.dirHasEntries(dir) {
			return true, dir
		}
	}
	return false, ""
}

// browserCheckedPlaces names what browserFound looks at, for diagnostics.
func (s agentBrowserSystem) browserCheckedPlaces() string {
	places := []string{"AGENT_BROWSER_EXECUTABLE_PATH"}
	for _, c := range s.browserCandidates(s.goos) {
		if c.Path != "" {
			places = append(places, c.Path)
		} else {
			places = append(places, c.Command+" in PATH")
		}
	}
	return strings.Join(append(places, "~/.agent-browser/browsers"), ", ")
}

// agentBrowserPlan is what Detect decided: which steps to run, or why the
// capability is manual.
type agentBrowserPlan struct {
	Method       string // providers.MethodBrew or MethodNpm; empty when no install step
	Chrome       bool
	ManualReason string // set when the capability must be finished by hand
	Unchecked    bool   // Node and the npm prefix were not checked (no program may run)
	Nothing      bool   // already installed with a browser
}

func (p agentBrowserPlan) manual() bool { return p.ManualReason != "" }

const agentBrowserManualCommands = "brew install agent-browser (macOS arm64) or npm install -g agent-browser (Node.js 24 or newer), then agent-browser install to download a browser"

func agentBrowserManualReason(cause string) string {
	return cause + " Install it by hand: " + agentBrowserManualCommands + ". See github.com/vercel-labs/agent-browser."
}

// planAgentBrowser decides the offer from the probes. runPrograms is true only
// in a real interactive install: otherwise no program runs, so Node and the
// npm prefix are not checked and the plan says so (D-9).
func planAgentBrowser(s agentBrowserSystem, runPrograms bool) agentBrowserPlan {
	_, present := s.lookPath(agentBrowserBinary)
	hasProgram := present == nil
	browser, _ := s.browserFound()
	// Only the tested platforms get executable steps, the Chrome step included.
	supported := (s.goos == "darwin" && s.goarch == "arm64") || (s.goos == "linux" && s.goarch == "amd64")
	if !supported && !(hasProgram && browser) {
		return agentBrowserPlan{ManualReason: agentBrowserManualReason("Automatic installation is tested only on macOS arm64 and Linux amd64, not " + s.goos + "/" + s.goarch + ".")}
	}
	if hasProgram {
		if browser {
			return agentBrowserPlan{Nothing: true}
		}
		return agentBrowserPlan{Chrome: true}
	}
	plan := agentBrowserPlan{Chrome: !browser}
	switch {
	case s.goos == "darwin" && s.goarch == "arm64":
		if _, err := s.lookPath("brew"); err != nil {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("Homebrew was not found in PATH.")}
		}
		plan.Method = providers.MethodBrew
	case s.goos == "linux" && s.goarch == "amd64":
		if _, err := s.lookPath("npm"); err != nil {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("Node.js with npm was not found in PATH.")}
		}
		plan.Method = providers.MethodNpm
		if !runPrograms {
			plan.Unchecked = true
			break
		}
		version, err := s.nodeVersion()
		if err != nil {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("Node.js could not be queried: " + sanitizeLine(err.Error()) + ".")}
		}
		if major, ok := nodeMajor(version); !ok || major < agentBrowserNodeMajor {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("Node.js " + sanitizeLine(version) + " is older than " + strconv.Itoa(agentBrowserNodeMajor) + ", which agent-browser requires.")}
		}
		prefix, err := s.npmPrefix()
		if err != nil || prefix == "" {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("npm's global prefix could not be determined.")}
		}
		if !s.prefixWritable(prefix) {
			return agentBrowserPlan{ManualReason: agentBrowserManualReason("npm's global prefix " + sanitizeLine(prefix) + " is not writable, and Hive never uses sudo.")}
		}
	default:
		return agentBrowserPlan{ManualReason: agentBrowserManualReason("Automatic installation is tested only on macOS arm64 and Linux amd64, not " + s.goos + "/" + s.goarch + ".")}
	}
	return plan
}

// nodeMajor parses the major version of "v24.1.0".
func nodeMajor(version string) (int, bool) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	major, _, _ := strings.Cut(version, ".")
	n, err := strconv.Atoi(major)
	return n, err == nil
}

// effects are the consent lines the summary shows before the operator applies.
func (p agentBrowserPlan) effects() []string {
	if p.manual() {
		return []string{string(providers.Manual) + " — " + p.ManualReason}
	}
	var lines []string
	switch p.Method {
	case providers.MethodBrew:
		lines = append(lines, "Runs `brew install agent-browser`, which installs the latest version. Homebrew also installs Node.js and other dependencies agent-browser needs.")
	case providers.MethodNpm:
		lines = append(lines, "Runs `npm install -g agent-browser`, which installs the latest version. Its postinstall script downloads a native binary from GitHub without a checksum.")
	}
	if p.Chrome {
		lines = append(lines, "Runs `agent-browser install` to download Chrome for Testing into ~/.agent-browser/browsers, because no browser was found.")
	}
	lines = append(lines, "No shell and no sudo; each command is stopped after 15 minutes.")
	if p.Unchecked {
		lines = append(lines, "Node.js 24 or newer and a writable npm prefix are checked only in a real interactive install; if they fail then, the capability stays manual.")
	}
	return lines
}
