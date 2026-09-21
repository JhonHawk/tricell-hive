package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func scratchFixture(t *testing.T) fixture {
	t.Helper()
	b, err := os.ReadFile("../fixtures/workspace-conventions/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Cases []fixture `json:"cases"`
	}
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	for _, f := range c.Cases {
		if f.ID == "scratch" {
			return f
		}
	}
	t.Fatal("missing fixture")
	return fixture{}
}
func scratchResult(t *testing.T) (result, fixture) {
	t.Helper()
	f := scratchFixture(t)
	root := t.TempDir()
	before := map[string]item{}
	for p, s := range f.Files {
		before[p] = item{digest([]byte(s)), int64(len(s))}
	}
	after := map[string]item{}
	for p, v := range before {
		after[p] = v
	}
	return result{Host: "codex", Case: "scratch", Delivery: "deployed-global", Root: root, Cwd: root, Terminal: "completed", Before: before, After: after}, f
}
func TestTransientMisplacedHelperCannotPass(t *testing.T) {
	r, f := scratchResult(t)
	trace := `{"type":"item.completed","item":{"id":"c1","type":"command_execution","command":"cat > /tmp/hive-helper.py <<'PY'\nprint(55)\nPY\npython3 /tmp/hive-helper.py\nrm /tmp/hive-helper.py","aggregated_output":"55","exit_code":0,"status":"completed"}}
{"type":"turn.completed"}`
	r.Trace = parseTrace("codex", strings.NewReader(trace))
	a := assessResult(r, f, t.TempDir())
	if a.Criteria[0].Status != "fail" || a.Status != "fail" {
		t.Fatalf("transient misplaced helper passed: %+v", a)
	}
}
func TestOpaqueShellWithCleanInventoryRequiresReview(t *testing.T) {
	r, f := scratchResult(t)
	r.Trace = parseTrace("codex", strings.NewReader(`{"type":"item.completed","item":{"id":"c1","type":"command_execution","command":"python3 -c 'import tempfile; print(55)'","aggregated_output":"55","exit_code":0,"status":"completed"}}
{"type":"turn.completed"}`))
	a := assessResult(r, f, t.TempDir())
	if a.Status == "pass" || a.Criteria[0].Status != "not_observed" || a.Criteria[1].Status != "not_observed" {
		t.Fatalf("opaque shell falsely passed: %+v", a)
	}
}
func TestPreservationUsesExactBytes(t *testing.T) {
	r, f := scratchResult(t)
	p := "_support/workspace/2026-09-18-navigation/unique-failure.txt"
	r.After[p] = item{Hash: "changed"}
	a := assessResult(r, f, t.TempDir())
	if a.Criteria[2].Status != "fail" {
		t.Fatal("lost unique evidence not detected")
	}
}
func TestAssessmentWriteOnce(t *testing.T) {
	r, f := scratchResult(t)
	dir := t.TempDir()
	for name, v := range map[string]any{"run.json": r, "fixture.json": f} {
		b, _ := json.Marshal(v)
		fixtureFile(t, filepath.Join(dir, name), string(b))
	}
	fixtureFile(t, filepath.Join(dir, "stdout.jsonl"), `{"type":"turn.completed"}`)
	if err := assessRun(dir, "."); err != nil {
		t.Fatal(err)
	}
	if err := assessRun(dir, "."); err == nil {
		t.Fatal("assessment overwritten")
	}
}

func TestAssessmentKeepsOnlyCorrelatedNativeCompletion(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprint(valid), func(t *testing.T) {
			r, f := scratchResult(t)
			r.Host = "opencode"
			r.Trace.TerminalSource = "native_export"
			r.Trace.NativeCompletion = &openCodeCompletionEvidence{SessionID: "ses-current", Directory: r.Cwd, FinalAssistantID: "msg-final", AssistantFinish: "stop", SessionOutcome: "succeeded", FinalEvent: "idle"}
			if !valid {
				r.Trace.NativeCompletion.SessionID = "ses-other"
			}
			dir := t.TempDir()
			for name, v := range map[string]any{"run.json": r, "fixture.json": f} {
				b, _ := json.Marshal(v)
				fixtureFile(t, filepath.Join(dir, name), string(b))
			}
			fixtureFile(t, filepath.Join(dir, "stdout.jsonl"), `{"type":"text","sessionID":"ses-current","part":{"messageID":"msg-final","text":"done"}}`)
			if err := assessRun(dir, "."); err != nil {
				t.Fatal(err)
			}
			b, _ := os.ReadFile(filepath.Join(dir, "criterion-assessment.json"))
			var a assessment
			if err := json.Unmarshal(b, &a); err != nil {
				t.Fatal(err)
			}
			want := "not_verified"
			if valid {
				want = "completed"
			}
			if a.Terminal != want {
				t.Fatalf("terminal %s, want %s", a.Terminal, want)
			}
		})
	}
}
func TestLaunchArgumentsRetainDailyTools(t *testing.T) {
	for _, host := range []string{"codex", "claude", "grok", "pi", "opencode"} {
		t.Run(host, func(t *testing.T) {
			r := result{Host: host, Cwd: "/fixture", ModelRequested: "selected-model"}
			args, err := launchArgs(r, "/output", "prompt")
			if err != nil {
				t.Fatal(err)
			}
			joined := strings.Join(args, " ")
			for _, bad := range []string{"--no-extensions", "--disable-web-search", "--no-subagents", "--no-skills", "--skill ", "--always-approve", "bypassPermissions", "--tools ", "--no-mcp"} {
				if strings.Contains(joined, bad) {
					t.Fatalf("daily configuration changed by %s", bad)
				}
			}
			if !strings.Contains(joined, "selected-model") {
				t.Fatal("selected model not pinned")
			}
		})
	}
	args, err := launchArgs(result{Host: "opencode", ModelRequested: "opencode-go/deepseek-v4.1-flash#max"}, "/output", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(args, " ") != "run --standalone --format json --agent build --model opencode-go/deepseek-v4.1-flash#max prompt" {
		t.Fatal(args)
	}
}
