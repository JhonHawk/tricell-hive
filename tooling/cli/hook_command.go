package main

// hook_command.go is `hive hook <host> <event>`, the silent subcommand every
// host's hook (or plugin, or extension) runs. It reads the host's payload on
// stdin, tells which host really ran it, records the session map, and reports the
// event to the viewer or queues it. It never prints and never fails: hosts read
// a non-zero exit, and some read stdout, as a verdict or as model context, so
// every error and panic ends in empty output and exit 0.
//
// The host named on the command line is only a hint: Grok and Cursor CLI also
// run the hooks registered in Claude Code's settings, with their own payloads.

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tricell-hive/integrations/claude"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/cursor"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/opencode"
	"tricell-hive/integrations/pi"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

const (
	hookInputLimit = 1 << 20
	// hookBudget bounds the whole run. Hosts kill a slow hook, and a Claude Code
	// SessionEnd hook may have 1.5 s: whatever is unfinished stays in the queues.
	hookBudgetDefault = 2 * time.Second
)

var (
	hookBudget = hookBudgetDefault
	hookNow    = time.Now
	// hookClock and hookAfter are the budget's clock and timer; tests replace
	// them so the budget is driven by observable events, not by sleeping.
	hookClock = time.Now
	hookAfter = time.After
)

type hookOptions struct{ home, stateDir string }

// runHook never returns an error and writes nothing to stdout or stderr.
func runHook(args []string, stdin io.Reader) {
	defer func() { _ = recover() }()
	var opts hookOptions
	var positional []string
	for i := 0; i < len(args); i++ {
		name, value, has := strings.Cut(args[i], "=")
		switch name {
		case "--home", "--state-dir":
			if !has {
				if i+1 >= len(args) {
					return
				}
				i++
				value = args[i]
			}
			if name == "--home" {
				opts.home = value
			} else {
				opts.stateDir = value
			}
		default:
			positional = append(positional, args[i])
		}
	}
	if len(positional) != 2 {
		return
	}
	deps := viewIdentityDepsFn()
	done := make(chan struct{})
	run := newHookRun(hookClock().Add(hookBudget))
	go func() {
		defer close(done)
		defer func() { _ = recover() }()
		processHook(opts, positional[0], positional[1], stdin, deps, run)
	}()
	select {
	case <-done:
	case <-hookAfter(hookBudget):
		run.finish() // the budget ended: what is not delivered stays queued
	}
}

// canonicalHookHost maps the host word on the command line to the viewer's host.
func canonicalHookHost(hint string) string {
	switch strings.ToLower(hint) {
	case "claude", "claude-code":
		return claude.HookHost
	case "codex":
		return codex.HookHost
	case "grok":
		return grok.HookHost
	case "opencode":
		return opencode.HookHost
	case "pi":
		return pi.HookHost
	case "cursor", "cursor-agent":
		return cursor.HookHost
	}
	return ""
}

// detectHookHost tells which host ran the hook. The OpenCode plugin and the Pi
// extension are Hive's own files and name themselves; for the others the payload
// and the environment win over the word on the command line, because Grok and
// Cursor CLI run hooks registered for Claude Code.
func detectHookHost(hint string, payload map[string]any, getenv func(string) string) string {
	host := canonicalHookHost(hint)
	switch host {
	case "", opencode.HookHost, pi.HookHost:
		return host
	}
	switch {
	case grok.IsHookPayload(payload, getenv):
		return grok.HookHost
	case cursor.IsHookPayload(payload):
		return cursor.HookHost
	}
	return host
}

func translateHook(host string, in hostevent.Input) []hostevent.Event {
	switch host {
	case claude.HookHost:
		return claude.HookEvents(in)
	case codex.HookHost:
		return codex.HookEvents(in)
	case grok.HookHost:
		return grok.HookEvents(in)
	case cursor.HookHost:
		return cursor.HookEvents(in)
	case opencode.HookHost:
		return opencode.HookEvents(in)
	case pi.HookHost:
		return pi.HookEvents(in)
	}
	return nil
}

func processHook(opts hookOptions, hint, event string, stdin io.Reader, deps viewIdentityDeps, run *hookRun) {
	raw, err := io.ReadAll(io.LimitReader(stdin, hookInputLimit+1))
	if err != nil || len(raw) > hookInputLimit {
		return
	}
	payload := hostevent.Decode(raw)
	if payload == nil {
		return
	}
	host := detectHookHost(hint, payload, deps.getenv)
	events := translateHook(host, hostevent.Input{Event: event, Payload: payload, Getenv: deps.getenv})
	if len(events) == 0 {
		return
	}
	options := management.Options{Scope: "user", Home: opts.home, StateDir: opts.stateDir}
	_, state, err := management.NormalizeOptions(options)
	if err != nil {
		return
	}
	if abs, err := filepath.Abs(state); err == nil {
		state = abs
	}
	var built []viewer.TimelineHostEvent
	for n, ev := range events {
		if validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(ev.Host), NativeSessionID: ev.SessionID}) != nil {
			continue
		}
		var hostStart time.Time
		if hookUsesMap(ev.Host) && n == 0 {
			// The map is written whether or not the viewer runs.
			hostStart, _ = recordSessionMap(state, deps, ev, hookNow())
		} else if hookUsesMap(ev.Host) && ev.OccurredAt.IsZero() {
			if pid, ok := findHostProcess(deps, ev.Host); ok && deps.processStart != nil {
				hostStart, _ = deps.processStart(pid)
			}
		}
		built = append(built, buildHostEvent(ev, hostStart))
	}
	sendHookEvents(options, state, built, run)
}

// hookOccurredAt is the time the event is reported with. The viewer derives the
// operation ID from the event's content, and stamps last activity with this
// time. A task event is reported with a time that is stable across runs and
// never the hook's clock: the host's own timestamp, the modification time of the
// host's transcript, the start time of the host process, then the Unix epoch. An
// activity-only event (a prompt, a stop, a subagent, a session start or end) has
// no such need, since applying activity twice only moves it forward: it takes
// the host's timestamp, then the transcript's, then the hook's clock, so the
// activity counts as it happens whatever the host reports.
func hookOccurredAt(ev hostevent.Event, hostStart time.Time) time.Time {
	if !ev.OccurredAt.IsZero() {
		return ev.OccurredAt.UTC()
	}
	if ev.TranscriptPath != "" {
		if st, err := os.Stat(ev.TranscriptPath); err == nil && st.Mode().IsRegular() {
			return st.ModTime().UTC()
		}
	}
	if !strings.HasPrefix(ev.Type, "task_") {
		return hookNow().UTC()
	}
	if !hostStart.IsZero() {
		return hostStart.UTC()
	}
	return time.Unix(0, 0).UTC()
}
