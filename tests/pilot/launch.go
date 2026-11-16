package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// Setting Cmd.Env explicitly prevents os/exec from updating PWD for Cmd.Dir.
// Native clients must see the same directory through both channels.
func environmentInDirectory(env []string, cwd string) []string {
	out := make([]string, 0, len(env)+1)
	for _, item := range env {
		if !strings.HasPrefix(item, "PWD=") && !strings.HasPrefix(item, "OLDPWD=") {
			out = append(out, item)
		}
	}
	return append(out, "PWD="+cwd)
}

// gh reads a token from these variables before its configuration directory.
var githubTokenVariables = []string{"GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"}

// A launched session must not reach GitHub as the user: gh finds no token in
// the environment and no stored session in configDir. PATH is left alone,
// since the hosts' shells rebuild it from their startup files.
func withoutGitHubCredentials(env []string, configDir string) []string {
	out := make([]string, 0, len(env)+1)
	for _, item := range env {
		name, _, _ := strings.Cut(item, "=")
		if name != "GH_CONFIG_DIR" && !slices.Contains(githubTokenVariables, name) {
			out = append(out, item)
		}
	}
	return append(out, "GH_CONFIG_DIR="+configDir)
}

// githubAccessReport holds variable names only, never their values.
type githubAccessReport struct {
	ConfigDirectory    string
	PreflightCommand   string
	CLIFound           bool
	AuthStatusExitCode int
	Variables          []string
}

// prepareGitHubIsolation trims env, then asks gh itself, under that same
// environment and before any model runs, whether it still has a session. A gh
// that answers as signed in stops the launch; a missing gh is recorded.
func prepareGitHubIsolation(output string, env []string) ([]string, githubAccessReport, error) {
	report := githubAccessReport{ConfigDirectory: "gh-config", PreflightCommand: "gh auth status"}
	configDir := filepath.Join(output, report.ConfigDirectory)
	if err := os.Mkdir(configDir, 0700); err != nil {
		return nil, report, fmt.Errorf("create empty gh configuration directory: %w", err)
	}
	env = withoutGitHubCredentials(env, configDir)
	for _, item := range env {
		name, _, _ := strings.Cut(item, "=")
		if strings.HasPrefix(name, "GH_") || strings.HasPrefix(name, "GITHUB_") {
			report.Variables = append(report.Variables, name)
		}
	}
	slices.Sort(report.Variables)
	bin, err := exec.LookPath("gh")
	if err != nil {
		return env, report, nil
	}
	report.CLIFound = true
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "auth", "status")
	cmd.Env = env
	err = cmd.Run()
	var exit *exec.ExitError
	switch {
	case err == nil:
		return nil, report, fmt.Errorf("gh still answers as signed in without credentials in the child environment")
	case errors.As(err, &exit) && exit.ExitCode() > 0:
		report.AuthStatusExitCode = exit.ExitCode()
		return env, report, nil
	}
	return nil, report, fmt.Errorf("gh auth status preflight did not finish: %w", err)
}

func validHost(host string) bool {
	switch host {
	case "codex", "claude", "grok", "pi", "opencode":
		return true
	}
	return false
}

// Only invocation-local options; no tool removal or MCP filtering.
// Codex permission bypass is opt-in and requires explicit user authorization.
func launchArgs(r result, output, prompt string) ([]string, error) {
	if r.ModelRequested == "" {
		return nil, fmt.Errorf("explicit model required")
	}
	if r.Provider != "" && r.Host != "pi" {
		return nil, fmt.Errorf("--provider is only supported for Pi; use the host's model ID otherwise")
	}
	if r.CodexBypassSandbox && r.Host != "codex" {
		return nil, fmt.Errorf("Codex sandbox bypass requires codex")
	}
	switch r.Host {
	case "codex":
		a := []string{"-a", "never", "exec", "--json", "--ephemeral", "-m", r.ModelRequested, "-s", "workspace-write", "-C", r.Cwd}
		if r.CodexBypassSandbox {
			a = []string{"exec", "--json", "--ephemeral", "-m", r.ModelRequested, "--dangerously-bypass-approvals-and-sandbox", "-C", r.Cwd}
		}
		// Codex 0.155.1 native collaboration cannot resolve an ephemeral parent
		// thread. Keep persistence only for this explicit delegation probe.
		if r.Case == "native-agent-dispatch" {
			a = slices.DeleteFunc(a, func(arg string) bool { return arg == "--ephemeral" })
		}
		if r.Effort != "" {
			a = append(a, "-c", "model_reasoning_effort="+fmt.Sprintf("%q", r.Effort))
		}
		return append(a, "-"), nil
	case "claude":
		a := []string{"-p", "--model", r.ModelRequested, "--output-format", "stream-json", "--verbose", "--no-session-persistence", "--permission-mode", "acceptEdits", "--allowedTools", "Read,Write,Edit,Glob,Grep,Bash"}
		if r.Effort != "" {
			a = append(a, "--effort", r.Effort)
		}
		return a, nil
	case "grok":
		a := []string{"--cwd", r.Cwd, "-p", prompt, "--model", r.ModelRequested, "--output-format", "streaming-messages-json"}
		if r.Effort != "" {
			a = append(a, "--reasoning-effort", r.Effort)
		}
		return a, nil
	case "pi":
		a := []string{"-p", "--mode", "json", "--model", r.ModelRequested, "--session-dir", filepath.Join(output, "native-sessions")}
		if r.Provider != "" {
			a = append(a, "--provider", r.Provider)
		}
		if r.Effort != "" {
			a = append(a, "--thinking", r.Effort)
		}
		return a, nil
	case "opencode":
		if r.Effort != "" {
			return nil, fmt.Errorf("OpenCode effort belongs in --model provider/model#variant")
		}
		return []string{"run", "--standalone", "--format", "json", "--agent", "build", "--model", r.ModelRequested, prompt}, nil
	}
	return nil, fmt.Errorf("unsupported host %q", r.Host)
}
