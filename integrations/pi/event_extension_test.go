package pi

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/hostevent"
)

func TestRenderEventExtensionNamesTheBinaryAsAStringLiteral(t *testing.T) {
	for _, exe := range []string{"/opt/hive/bin/hive", `/Users/o'brien/a "b"/hive`, "/path with space/hive"} {
		out, err := RenderEventExtension(exe)
		if err != nil {
			t.Fatal(err)
		}
		src := string(out)
		if strings.Contains(src, executablePlaceholder) {
			t.Fatalf("placeholder left in the extension for %q", exe)
		}
		if !strings.HasPrefix(src, "// hive-managed: session events\n") {
			t.Fatal("the extension lacks its marker")
		}
		for _, want := range []string{`pi.on("session_start"`, `pi.on("agent_settled"`, `pi.on("session_shutdown"`, "detached: true", "child.unref()", "getSessionId"} {
			if !strings.Contains(src, want) {
				t.Fatalf("extension lacks %q", want)
			}
		}
	}
	for _, bad := range []string{"", "relative/hive", "/a\x00b/hive"} {
		if _, err := RenderEventExtension(bad); err == nil {
			t.Fatalf("RenderEventExtension(%q) accepted a bad path", bad)
		}
	}
}

// The extension runs under Node here (which strips the types) with a stand-in
// for Pi's extension API and a stand-in for the hive binary.
func TestEventExtensionReportsTheSessionThroughTheHook(t *testing.T) {
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
	script := "#!/bin/sh\n{ printf 'ARGS:%s\\n' \"$*\"; cat; printf '\\n'; } > \"$HIVE_TEST_CALLS/$$\"\n"
	if err := os.WriteFile(fake, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	ext, err := RenderEventExtension(fake)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extension.ts"), ext, 0o644); err != nil {
		t.Fatal(err)
	}
	harness := `
import extension from "./extension.ts"
const handlers = {}
extension({ on: (name, fn) => { handlers[name] = fn } })
const ctx = { cwd: "/work/app", sessionManager: { getSessionId: () => "pi-session-1" } }
handlers.session_start({ type: "session_start", reason: "startup" }, ctx)
handlers.agent_settled({ type: "agent_settled" }, ctx)
handlers.session_shutdown({ type: "session_shutdown" }, ctx)
handlers.agent_settled({ type: "agent_settled" }, { cwd: "/x", sessionManager: { getSessionId: () => "" } })
handlers.agent_settled({ type: "agent_settled" }, undefined)
await new Promise((r) => setTimeout(r, 600))
console.log(Object.keys(handlers).sort().join(","))
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
	if strings.TrimSpace(string(out)) != "agent_settled,session_shutdown,session_start" {
		t.Fatalf("harness output = %q", out)
	}
	var blocks []string
	for i := 0; i < 50; i++ {
		blocks = nil
		files, _ := os.ReadDir(calls)
		for _, f := range files {
			b, _ := os.ReadFile(filepath.Join(calls, f.Name()))
			blocks = append(blocks, string(b))
		}
		if len(blocks) >= 3 {
			break
		}
		exec.Command("sleep", "0.1").Run()
	}
	if len(blocks) != 3 {
		t.Fatalf("want 3 hook runs (no session ID means no report), got %d: %q", len(blocks), blocks)
	}
	seen := map[string]bool{}
	for _, block := range blocks {
		line, body, _ := strings.Cut(strings.TrimPrefix(block, "ARGS:"), "\n")
		name := strings.TrimPrefix(line, "hook pi ")
		seen[name] = true
		payload := hostevent.Decode([]byte(strings.TrimSpace(body)))
		events := HookEvents(hostevent.Input{Event: name, Payload: payload})
		if len(events) != 1 || events[0].SessionID != "pi-session-1" || events[0].Cwd != "/work/app" || events[0].OccurredAt.IsZero() {
			t.Fatalf("%s: events %+v from %v", name, events, payload)
		}
		want := map[string]string{"session_start": hostevent.SessionStarted, "agent_settled": hostevent.Activity, "session_shutdown": hostevent.SessionEnded}[name]
		if events[0].Type != want {
			t.Fatalf("%s: type %s, want %s", name, events[0].Type, want)
		}
	}
	if len(seen) != 3 {
		t.Fatalf("events seen: %v", seen)
	}
}
