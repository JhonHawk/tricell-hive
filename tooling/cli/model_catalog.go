// model_catalog.go lists the models each CLI offers, for the Models view's
// model picker (#46, T7). It only provides the listing and the bounded runner;
// the per-session cache and the view wiring belong to the view. Nothing here
// runs unless a caller asks for a list: `hive models`, `hive doctor` and the
// views that do not edit never do.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"tricell-hive/integrations/agents"
	"tricell-hive/tooling/management"
)

const (
	catalogTimeout   = 8 * time.Second
	catalogWaitDelay = time.Second
	catalogMaxOutput = 8 << 20
)

// catalogModel is one model a CLI offers. Efforts is empty when the host's
// general effort list applies.
type catalogModel struct {
	ID      string
	Efforts []string
}

// catalogRunner runs a fixed listing command and returns its standard output.
type catalogRunner func(ctx context.Context, bin string, args []string) ([]byte, error)

// defaultCatalogRunner is the runner the view uses; tests replace it.
var defaultCatalogRunner catalogRunner = newCatalogRunner(catalogTimeout, catalogMaxOutput)

// claudeAliases are the documented model aliases; Claude Code has no listing
// command, so nothing is executed for it.
var claudeAliases = []string{"fable", "opus", "sonnet", "haiku"}

// catalogCommand is the executable and fixed arguments that list a host's
// models.
func catalogCommand(host string) (bin string, args []string, ok bool) {
	switch host {
	case "codex":
		return "codex", []string{"debug", "models"}, true
	case "opencode":
		return "opencode", []string{"models"}, true
	case "pi":
		return "pi", []string{"--list-models"}, true
	case "grok":
		return "grok", []string{"models"}, true
	case "cursor":
		return "cursor-agent", []string{"models"}, true
	}
	return "", nil, false
}

// catalogQueryHook, when set, is told of every attempt to list a host's models,
// before any runner is consulted. Tests use it to prove that a path never
// queries a list, whatever runner it would have used, so a refusing runner
// under --home cannot hide a call.
var catalogQueryHook func(host string)

// listHostModels runs the host's listing command and parses its output. Only
// ids that an override accepts are returned.
func listHostModels(ctx context.Context, host string, run catalogRunner) ([]catalogModel, error) {
	if catalogQueryHook != nil {
		catalogQueryHook(host)
	}
	if host == "claude" {
		models := make([]catalogModel, 0, len(claudeAliases))
		for _, id := range claudeAliases {
			models = append(models, catalogModel{ID: id})
		}
		return models, nil
	}
	bin, args, ok := catalogCommand(host)
	if !ok {
		return nil, fmt.Errorf("no model list for host %q", host)
	}
	if run == nil {
		return nil, errors.New("no runner")
	}
	out, err := run(ctx, bin, args)
	if err != nil {
		return nil, err
	}
	// OpenCode sometimes prints nothing and exits 0; one retry covers it.
	if host == "opencode" && len(bytes.TrimSpace(out)) == 0 {
		if out, err = run(ctx, bin, args); err != nil {
			return nil, err
		}
	}
	if len(bytes.TrimSpace(out)) == 0 {
		return nil, errors.New("the command printed nothing")
	}
	var candidates []catalogModel
	switch host {
	case "codex":
		if candidates, err = parseCodexModels(out); err != nil {
			return nil, err
		}
	case "opencode":
		candidates = parseLines(out, func(line string) string { return line })
	case "pi":
		candidates = parseLines(out, piModelID)
	case "grok":
		candidates = parseLines(out, grokModelID)
	case "cursor":
		candidates = parseLines(out, cursorModelID)
	}
	return acceptedModels(host, candidates)
}

// acceptedModels keeps the candidates whose id an override accepts, once each.
func acceptedModels(host string, candidates []catalogModel) ([]catalogModel, error) {
	seen := map[string]bool{}
	models := make([]catalogModel, 0, len(candidates))
	discarded := 0
	for _, c := range candidates {
		if agents.ValidateOverride(host, agents.ModelOverride{Model: c.ID}) != nil {
			discarded++
			continue
		}
		if seen[c.ID] {
			continue
		}
		seen[c.ID] = true
		models = append(models, c)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no usable model in the output (%d ids discarded)", discarded)
	}
	return models, nil
}

// parseLines turns each non-empty line into a candidate through id; an empty
// id skips the line.
func parseLines(out []byte, id func(line string) string) []catalogModel {
	var models []catalogModel
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if v := id(line); v != "" {
			models = append(models, catalogModel{ID: v})
		}
	}
	return models
}

// piModelID reads `provider model ...` rows and skips the header.
func piModelID(line string) string {
	f := strings.Fields(line)
	if len(f) < 2 || (f[0] == "provider" && f[1] == "model") {
		return ""
	}
	return f[0] + "/" + f[1]
}

// grokModelID reads `- id` and `* id (default)` rows; every other line, such
// as the session line, is skipped.
func grokModelID(line string) string {
	if len(line) < 3 || (line[0] != '-' && line[0] != '*') || (line[1] != ' ' && line[1] != '\t') {
		return ""
	}
	f := strings.Fields(line[1:])
	if len(f) == 0 {
		return ""
	}
	return f[0]
}

// cursorModelID reads `id - Name` rows; the header and the tip have no ` - `
// after a single-token id.
func cursorModelID(line string) string {
	id, _, found := strings.Cut(line, " - ")
	id = strings.TrimSpace(id)
	if !found || id == "" || strings.ContainsAny(id, " \t") {
		return ""
	}
	return id
}

// parseCodexModels reads `codex debug models`: models[].slug with visibility
// "list", and each model's supported_reasoning_levels[].effort.
func parseCodexModels(out []byte) ([]catalogModel, error) {
	var doc struct {
		Models []struct {
			Slug       string `json:"slug"`
			Visibility string `json:"visibility"`
			Levels     []struct {
				Effort string `json:"effort"`
			} `json:"supported_reasoning_levels"`
		} `json:"models"`
	}
	if err := json.Unmarshal(out, &doc); err != nil {
		return nil, fmt.Errorf("unreadable codex output: %w", err)
	}
	var models []catalogModel
	for _, m := range doc.Models {
		if m.Visibility != "list" {
			continue
		}
		c := catalogModel{ID: m.Slug}
		for _, l := range m.Levels {
			if agents.ValidateOverride("codex", agents.ModelOverride{Effort: l.Effort}) == nil {
				c.Efforts = append(c.Efforts, l.Effort)
			}
		}
		models = append(models, c)
	}
	return models, nil
}

// catalogRunnerFor returns the runner for a command's options: under a
// synthetic --home no CLI is executed, as detectInstallerHosts does.
func catalogRunnerFor(o management.Options) catalogRunner {
	if o.Home != "" {
		return func(context.Context, string, []string) ([]byte, error) {
			return nil, errors.New("not executed under a synthetic home")
		}
	}
	return defaultCatalogRunner
}

// newCatalogRunner returns a runner that executes the binary found on PATH
// with the given arguments and no shell. It runs with the process
// environment, which the CLI needs to authenticate, but from the temporary
// directory, because some CLIs load project configuration from the current
// one; with no input, in its own session so it cannot reach the terminal; and
// with a time limit that kills the whole process group. Standard output is
// capped at maxOutput bytes and a larger output is an error, not a partial
// list. Standard error is discarded.
func newCatalogRunner(timeout time.Duration, maxOutput int) catalogRunner {
	return func(ctx context.Context, bin string, args []string) ([]byte, error) {
		path, err := exec.LookPath(bin)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		cmd := exec.CommandContext(ctx, path, args...)
		cmd.Dir = os.TempDir()
		cmd.Stdin = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
		cmd.WaitDelay = catalogWaitDelay
		out := &overflowBuffer{limit: maxOutput, stop: cancel}
		cmd.Stdout = out
		err = cmd.Run()
		if out.overflow {
			return nil, fmt.Errorf("output exceeds %d bytes", maxOutput)
		}
		if ctx.Err() != nil {
			return nil, fmt.Errorf("timed out after %s", timeout)
		}
		if err != nil && !errors.Is(err, exec.ErrWaitDelay) {
			return nil, err
		}
		return out.buf.Bytes(), nil
	}
}

// overflowBuffer keeps up to limit bytes. Writing past it marks the overflow
// and calls stop, so the process is killed instead of filling memory.
type overflowBuffer struct {
	buf      bytes.Buffer
	limit    int
	overflow bool
	stop     func()
}

func (b *overflowBuffer) Write(p []byte) (int, error) {
	if b.overflow {
		return len(p), nil
	}
	if b.buf.Len()+len(p) > b.limit {
		b.overflow = true
		b.stop()
		return len(p), nil
	}
	return b.buf.Write(p)
}
