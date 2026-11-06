package opencode

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/hostevent"
)

func TestRenderEventPluginNamesTheBinaryAsAStringLiteral(t *testing.T) {
	for _, exe := range []string{"/opt/hive/bin/hive", `/Users/o'brien/a "b"/hive`, "/path with space/hive", "/odd name/hive"} {
		out, err := RenderEventPlugin(exe)
		if err != nil {
			t.Fatal(err)
		}
		src := string(out)
		if strings.Contains(src, executablePlaceholder) || strings.Contains(src, " ") {
			t.Fatalf("placeholder or raw line separator left in the plugin for %q", exe)
		}
		if !strings.HasPrefix(src, "// hive-managed: session events\n") {
			t.Fatalf("the plugin lacks its marker")
		}
		for _, want := range []string{`id: "hive.session-events"`, "async setup(ctx)", `"execute.after"`, "ctx.event.subscribe", "detached: true", "child.unref()"} {
			if !strings.Contains(src, want) {
				t.Fatalf("plugin lacks %q", want)
			}
		}
	}
	for _, bad := range []string{"", "relative/hive", "/a\x00b/hive"} {
		if _, err := RenderEventPlugin(bad); err == nil {
			t.Fatalf("RenderEventPlugin(%q) accepted a bad path", bad)
		}
	}
}

// The plugin runs under Node here with a stand-in for OpenCode's context and a
// stand-in for the hive binary that records what it is given.
func TestEventPluginReportsSessionEventsAndTodosThroughTheHook(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	if err := os.Mkdir(calls, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := filepath.Join(dir, "hive")
	// One file per call: the hooks run detached and in parallel.
	script := "#!/bin/sh\n{ printf 'ARGS:%s\\n' \"$*\"; cat; printf '\\n'; } > \"$HIVE_TEST_CALLS/$$\"\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	plugin, err := RenderEventPlugin(fake)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.mjs"), plugin, 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
import plugin from "./plugin.mjs"
const hooks = {}
const events = [
  { type: "session.created", data: { sessionID: "ses_1" }, location: { directory: "/work/app" } },
  { type: "session.updated", data: { sessionID: "ses_1" } },
  { type: "session.execution.started", data: { sessionID: "ses_1" } },
  { type: "message.part.updated", data: { sessionID: "ses_1" } },
  { type: "session.deleted", data: { sessionID: "ses_2" } },
]
const ctx = {
  location: { directory: "/work/base" },
  tool: { hook: async (name, cb) => { hooks[name] = cb } },
  event: { subscribe: async function* () { for (const e of events) yield e } },
}
const dispose = await plugin.setup(ctx)
hooks["execute.after"]({ tool: "read", sessionID: "ses_1", input: {} })
hooks["execute.after"]({ tool: "todowrite", sessionID: "ses_1", input: { todos: [{ content: "Write tests", status: "in_progress", priority: "high" }, { content: "Ship", status: "pending" }] } })
await new Promise((r) => setTimeout(r, 600))
if (typeof dispose !== "function") throw new Error("setup must return a disposer")
dispose()
console.log(plugin.id)
`
	if err := os.WriteFile(filepath.Join(dir, "harness.mjs"), []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, "harness.mjs")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HIVE_TEST_CALLS="+calls)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	if strings.TrimSpace(string(out)) != "hive.session-events" {
		t.Fatalf("harness output = %q", out)
	}
	// The detached children may still be finishing.
	var blocks []string
	for i := 0; i < 50; i++ {
		blocks = nil
		files, _ := os.ReadDir(calls)
		for _, f := range files {
			b, _ := os.ReadFile(filepath.Join(calls, f.Name()))
			blocks = append(blocks, string(b))
		}
		if len(blocks) >= 4 {
			break
		}
		exec.Command("sleep", "0.1").Run()
	}
	joined := strings.Join(blocks, "")
	for _, want := range []string{
		"ARGS:hook opencode session.created", "ARGS:hook opencode session.execution.started",
		"ARGS:hook opencode session.deleted", "ARGS:hook opencode todowrite",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "session.updated") || strings.Contains(joined, "message.part") || len(blocks) != 4 {
		t.Fatalf("only session.created, session.execution.*, session.deleted and todowrite are reported:\n%s", joined)
	}

	// The envelope is the one T6's translator reads.
	var todo string
	for _, block := range blocks {
		line, body, _ := strings.Cut(strings.TrimPrefix(block, "ARGS:"), "\n")
		name := strings.TrimPrefix(line, "hook opencode ")
		payload := hostevent.Decode([]byte(strings.TrimSpace(body)))
		if payload == nil {
			t.Fatalf("%s: stdin is not a JSON object: %q", name, body)
		}
		events := HookEvents(hostevent.Input{Event: name, Payload: payload})
		if len(events) == 0 {
			t.Fatalf("%s: the translator reads no event from %v", name, payload)
		}
		if name == "todowrite" {
			todo = events[0].Title
			if len(events) != 2 || events[0].State != hostevent.StateInProgress || events[1].State != hostevent.StateNotStarted {
				t.Fatalf("todowrite events: %+v", events)
			}
		}
		if name == "session.created" && (events[0].Type != hostevent.SessionStarted || events[0].SessionID != "ses_1" || events[0].Cwd != "/work/app") {
			t.Fatalf("session.created event: %+v", events[0])
		}
		if name == "session.execution.started" && events[0].Cwd != "/work/app" {
			t.Fatalf("the directory remembered from session.created was lost: %+v", events[0])
		}
		if name == "session.deleted" && (events[0].Type != hostevent.SessionEnded || events[0].Cwd != "/work/base") {
			t.Fatalf("session.deleted event: %+v", events[0])
		}
		if payload["occurred_at"] == nil {
			t.Fatalf("%s: no occurred_at", name)
		}
	}
	if todo != "Write tests" {
		t.Fatalf("todo title = %q", todo)
	}
}

// A missing binary or a throwing host must never surface in OpenCode.
func TestEventPluginSwallowsAMissingBinary(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed")
	}
	dir := t.TempDir()
	plugin, err := RenderEventPlugin(filepath.Join(dir, "no-such-hive"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "plugin.mjs"), plugin, 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
import plugin from "./plugin.mjs"
const hooks = {}
const ctx = { tool: { hook: async (n, cb) => { hooks[n] = cb } }, event: { subscribe: async function* () { yield { type: "session.created", data: { sessionID: "s" } } } } }
const dispose = await plugin.setup(ctx)
hooks["execute.after"]({ tool: "todowrite", sessionID: "s", input: { todos: "not a list" } })
hooks["execute.after"](undefined)
await new Promise((r) => setTimeout(r, 300))
dispose()
console.log("ok")
`
	if err := os.WriteFile(filepath.Join(dir, "harness.mjs"), []byte(harness), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, "harness.mjs")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "ok" {
		t.Fatalf("node: %v\n%s", err, out)
	}
}
