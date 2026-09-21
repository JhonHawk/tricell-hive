package main

import (
	"fmt"
	"path/filepath"
)

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
		a := []string{"--cwd", r.Cwd, "-p", prompt, "--model", r.ModelRequested, "--output-format", "streaming-messages-json", "--permission-mode", "dontAsk", "--allow", "Edit(./**)", "--allow", "Write(./**)"}
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
