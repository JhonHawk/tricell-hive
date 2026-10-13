package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Regression: a persisted native projection must retain the completed child's
// findings and actual operation provenance, but not unrelated native fields.
func TestFlashClassProjectionRejectsIncompleteProvenance(t *testing.T) {
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	c := flashChild{SessionID: "child", FinalText: []string{"Unfinished claim"}}
	if saveFlashProjection(root, &c) == nil {
		t.Fatal("unfinished child result persisted as final evidence")
	}
}

// Regression: accepting another model, unsupported arm/host/case or an enlarged
// timeout would spend outside D7-B rather than mark coverage blocked.
func TestFlashClassFlags(t *testing.T) {
	for _, arm := range []string{"A", "B", "C"} {
		if err := validateFlashClassFlags("opencode", "plan-reference", arm, "/source", flashModels[0], "deployed-global", 3*time.Minute); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][6]string{
		{"claude", "plan-reference", "A", "/source", flashModels[0], "deployed-global"},
		{"opencode", "other", "A", "/source", flashModels[0], "deployed-global"},
		{"opencode", "plan-reference", "D", "/source", flashModels[0], "deployed-global"},
		{"opencode", "plan-reference", "C", "", flashModels[0], "deployed-global"},
		{"opencode", "plan-reference", "C", "/source", "openai/other", "deployed-global"},
		{"opencode", "plan-reference", "C", "/source", flashModels[0], "project"},
	} {
		if err := validateFlashClassFlags(args[0], args[1], args[2], args[3], args[4], args[5], 3*time.Minute); err == nil {
			t.Errorf("invalid pilot flags accepted: %q", args[:4])
		}
	}
	for _, limit := range []time.Duration{0, 4 * time.Minute} {
		if validateFlashClassFlags("opencode", "plan-reference", "B", "/source", flashModels[0], "deployed-global", limit) == nil {
			t.Error("invalid time limit accepted")
		}
	}
}

// Regression: skipping the arm rotation, adding retries or dropping a case
// changes the authorized population even though individual runs can succeed.
func TestFlashClassSchedule(t *testing.T) {
	schedule := flashSchedule()
	if len(schedule) != 48 {
		t.Fatal("schedule must contain exactly 48 roots")
	}
	seen := map[string]bool{}
	for _, r := range schedule {
		key := fmt.Sprintf("%s/%s/%s/%d", r.Model, r.Arm, r.Case, r.Repetition)
		if seen[key] {
			t.Fatal("duplicate root execution")
		}
		seen[key] = true
		if validateFlashClassFlags("opencode", r.Case, r.Arm, "/source", r.Model, "deployed-global", 3*time.Minute) != nil {
			t.Fatal("invalid scheduled execution")
		}
	}
	if schedule[0].Arm == schedule[6].Arm {
		t.Fatal("repetitions did not rotate arm order")
	}
}

// Regression: counting only the legacy CLI stream can miss V2 native tool
// dispatches; duplicate updates to one tool ID must not become extra children.
func TestFlashClassNativeDispatchCount(t *testing.T) {
	for raw, want := range map[string]int{
		`{"messages":[{"content":[{"type":"tool","id":"one","name":"subagent"}]}]}`:                                              1,
		`{"messages":[{"content":[{"type":"tool","id":"one","name":"subagent"},{"type":"tool","id":"two","name":"subagent"}]}]}`: 2,
		`{"messages":[{"content":[{"type":"tool","id":"one","name":"subagent"},{"type":"tool","id":"one","name":"subagent"}]}]}`: 1,
		`{"messages":[{"content":[{"type":"tool","id":"read","name":"read"}]}]}`:                                                 0,
	} {
		got, err := flashNativeDispatches([]byte(raw))
		if err != nil || got != want {
			t.Fatalf("dispatch count %d, want %d: %v", got, want, err)
		}
	}
}

// Regression: accepting the first assistant or ignoring idle/model/parent
// metadata lets a tool continuation or mismatched child masquerade as final.
func TestFlashClassNativeCompletionMetadata(t *testing.T) {
	base := map[string]any{"info": map[string]any{"id": "child", "parentID": "root", "agent": "native-role-id", "outcome": "succeeded", "model": map[string]any{"providerID": "opencode-go", "id": "claude-haiku-5-5", "variant": "high"}, "time": map[string]any{"idle": 123}}, "messages": []any{map[string]any{"id": "final", "type": "assistant", "agent": "native-role-id", "finish": "stop", "model": map[string]any{"providerID": "opencode-go", "id": "claude-haiku-5-5", "variant": "high"}, "time": map[string]any{"completed": 122}}}}
	data, _ := json.Marshal(base)
	if _, err := flashChildFromExport(data, "root", "hive-review-plan", flashModels[0]); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(map[string]any){
		func(d map[string]any) { object(d["info"])["parentID"] = "other" },
		func(d map[string]any) { object(object(d["info"])["model"])["variant"] = "medium" },
		func(d map[string]any) { object(object(d["info"])["time"])["idle"] = nil },
		func(d map[string]any) { object(list(d["messages"])[0])["finish"] = "tool-calls" },
		func(d map[string]any) { d["messages"] = append(list(d["messages"]), map[string]any{"type": "tool"}) },
	} {
		var d map[string]any
		_ = json.Unmarshal(data, &d)
		change(d)
		b, _ := json.Marshal(d)
		if _, err := flashChildFromExport(b, "root", "hive-review-plan", flashModels[0]); err == nil {
			t.Error("incomplete native child metadata accepted")
		}
	}
}

// Regression: trusting a loader-declared hash, duplicate copies, an unsupported
// transport or an uncorrelated request would overclaim pre-dispatch coverage.
func TestFlashClassOutgoingCoverage(t *testing.T) {
	root := flashObservation{Phase: "http", SessionID: "root", Model: flashModels[0], Correlated: true, Protocol: "chat-completions", LoaderActive: true}
	child := flashObservation{Phase: "http", SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], Correlated: true, Protocol: "chat-completions", Canonical: true, Copies: 1, InjectedCopies: 1, Hash: "actual", LoaderActive: true}
	if got := flashCoverage("C", "root", flashModels[0], child.Agent, "actual", []flashObservation{root, child}); got != "observed" {
		t.Fatal(got)
	}
	for _, change := range []func(*flashObservation){func(r *flashObservation) { r.Hash = "declared" }, func(r *flashObservation) { r.InjectedCopies = 2 }, func(r *flashObservation) { r.Correlated = false }, func(r *flashObservation) { r.Protocol = "unsupported" }, func(r *flashObservation) { r.Phase = "websocket" }} {
		r := child
		change(&r)
		if got := flashCoverage("C", "root", flashModels[0], child.Agent, "actual", []flashObservation{root, r}); got == "observed" {
			t.Fatal("invalid outgoing observation accepted")
		}
	}
	root.LoaderActive = false
	if got := flashCoverage("B", "root", flashModels[0], child.Agent, "actual", []flashObservation{root}); got != "unverified" {
		t.Fatal("absence of reading counted as success")
	}
}

// Regression: normal parent exit without process-group cleanup leaves a real
// background writer able to recreate credentials after the runner removes them.
func TestFlashClassProcessGroupCleanupAfterParentExit(t *testing.T) {
	pidfile := filepath.Join(t.TempDir(), "child.pid")
	cmd := exec.Command("sh", "-c", `sleep 30 & echo $! > "$1"`, "sh", pidfile)
	code, message := runProcess(cmd, 2*time.Second)
	if code != 0 || message != "" {
		t.Fatal(code, message)
	}
	b, err := os.ReadFile(pidfile)
	if err != nil {
		t.Fatal(err)
	}
	var pid int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(b)), "%d", &pid); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(pid, syscall.SIGKILL) })
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		state, _ := exec.Command("ps", "-o", "stat=", "-p", fmt.Sprint(pid)).Output()
		if len(strings.TrimSpace(string(state))) == 0 || strings.HasPrefix(strings.TrimSpace(string(state)), "Z") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("background writer survived parent exit")
}

// Regression: filepath.Join alone permits a cwd or intermediate symlink to
// escape the freshly created fixture before any process starts.
func TestFlashClassFixtureDirectoryConfined(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{"../outside", outside, "escape"} {
		if _, err := fixtureWorkingDirectory(root, relative); err == nil {
			t.Errorf("escaping cwd accepted: %q", relative)
		}
	}
	if got, err := fixtureWorkingDirectory(root, "."); err != nil || got == "" {
		t.Fatal("valid fixture cwd rejected", err)
	}
}

// Regression: parent exit alone, an unobserved child, extra children, a wrong
// parent, wrong role or effective model must never close a root execution.
func TestFlashClassChildCompletion(t *testing.T) {
	valid := flashChild{SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], FinalAssistantID: "final", Completed: true}
	if err := flashCompletion("root", flashModels[0], "hive-review-plan", []flashChild{valid}); err != nil {
		t.Fatal(err)
	}
	for _, children := range [][]flashChild{nil, {valid, valid}, {{SessionID: "child", ParentID: "other", AgentName: valid.AgentName, Model: valid.Model, Completed: true}}, {{SessionID: "child", ParentID: "root", AgentName: "other", Model: valid.Model, Completed: true}}, {{SessionID: "child", ParentID: "root", AgentName: valid.AgentName, Model: "not_observed", Completed: true}}, {{SessionID: "child", ParentID: "root", AgentName: valid.AgentName, Model: valid.Model, Completed: false}}} {
		if flashCompletion("root", valid.Model, valid.Agent, children) == nil {
			t.Error("incomplete or mismatched native child accepted")
		}
	}
}

// Regression: autoactivating B, loading the observer before C, inheriting a
// role's default model or losing source hashes invalidates the comparison.
func TestFlashClassShadowConfigurationArms(t *testing.T) {
	for _, arm := range []string{"B", "C"} {
		t.Run(arm, func(t *testing.T) {
			source, out, shadow := t.TempDir(), t.TempDir(), t.TempDir()
			configDir := filepath.Join(shadow, ".config", "opencode")
			skills := filepath.Join(configDir, "skills")
			role, resource := flashRole("plan-reference")
			files := map[string]string{
				filepath.Join(source, "tests/fixtures", flashSuite, "observer.js"): "// Fake observer source\n",
				filepath.Join(source, "content/guidance/global.md"):                "# Fake source guidance\n",
				filepath.Join(source, "content/skills", resource):                  "# Fake canonical source\n",
				filepath.Join(skills, resource):                                    "# Fake canonical source\n",
				filepath.Join(configDir, "opencode.json"):                          `{"mcp":{"servers":{"engram":{"type":"local","command":["engram","mcp"],"environment":{"ENGRAM_DATA_DIR":"fixture","ENGRAM_CLOUD_AUTOSYNC":"0"}}}}}`,
				filepath.Join(configDir, "agents", role+".md"):                     "---\nmodel: \"github-copilot/gpt-6.1-sol#medium\"\n---\n# Keep instructions\n",
				filepath.Join(configDir, "hive/context-loader/index.js"):           "// Fake inactive loader\n",
			}
			_, harnessResource := flashRole("harness-reference")
			files[filepath.Join(source, "content/skills", harnessResource)] = "# Fake harness source\n"
			files[filepath.Join(skills, harnessResource)] = "# Fake harness source\n"
			metadata, _ := json.Marshal(map[string]any{"schema": 2, "references": []map[string]string{
				{"role": "hive-review-harness", "resource": harnessResource, "sha256": digest([]byte("# Fake harness source\n"))},
				{"role": role, "resource": resource, "sha256": digest([]byte("# Fake canonical source\n"))},
			}})
			files[filepath.Join(configDir, "hive/context-loader/context-loader.json")] = string(metadata)
			for name, body := range files {
				if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(name, []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			g := &guidanceVariant{shadowHome: shadow, host: "opencode", report: guidanceVariantReport{Source: source, FlowBuildSkillPath: filepath.Join(skills, "flow-build/SKILL.md")}}
			r, err := setupFlashClass(source, "plan-reference", arm, out, flashModels[0], g, "fixture")
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(filepath.Join(configDir, "opencode.json"))
			if err != nil {
				t.Fatal(err)
			}
			var cfg map[string]any
			if json.Unmarshal(b, &cfg) != nil {
				t.Fatal("configuration invalid")
			}
			if object(object(object(cfg["mcp"])["servers"])["engram"])["codemode"] != false {
				t.Error("Engram code mode still enabled")
			}
			plugins := list(cfg["plugins"])
			if arm == "C" {
				if len(plugins) != 2 || str(plugins[0]) != filepath.Join(configDir, "hive/context-loader") {
					t.Fatal("C did not register loader before observer")
				}
			} else if len(plugins) != 1 {
				t.Fatal("B activated product loader")
			}
			if object(object(cfg["agents"])[role])["model"] != flashModels[0] {
				t.Fatal("child model override absent")
			}
			if !strings.HasSuffix(str(object(plugins[len(plugins)-1])["package"]), "pilot-observer") {
				t.Fatal("independent observer missing")
			}
			roleBytes, err := os.ReadFile(filepath.Join(configDir, "agents", role+".md"))
			if err != nil {
				t.Fatal(err)
			}
			if r.RoleFileHashBefore != digest([]byte(files[filepath.Join(configDir, "agents", role+".md")])) || r.RoleFileHashAfter != digest(roleBytes) || r.OriginalRoleModel != "github-copilot/gpt-6.1-sol#medium" {
				t.Error("role hash/model provenance absent from flash report")
			}
			if !strings.Contains(string(roleBytes), "model: \""+flashModels[0]+"\"") {
				t.Error("setup did not override Markdown role model")
			}
			if r.ObserverHash != hashOrAbsent(filepath.Join(configDir, "hive/pilot-observer/index.js")) || r.LoaderHash != hashOrAbsent(filepath.Join(configDir, "hive/context-loader/index.js")) {
				t.Error("plugin index hashes missing")
			}
			if r.CanonicalHash == "" || r.ObserverHash == "" || r.SourceGlobalHash == "absent" || r.PluginActive != (arm == "C") {
				t.Fatal("arm evidence absent")
			}
			if _, err := setupFlashClass(source, "plan-reference", "A", out, flashModels[0], g, "fixture"); err == nil {
				t.Fatal("unidentified baseline accepted")
			}
			if err := os.Remove(filepath.Join(skills, resource)); err != nil {
				t.Fatal(err)
			}
			if _, err := setupFlashClass(source, "plan-reference", arm, out, flashModels[0], g, "fixture"); err == nil {
				t.Fatal("arm without canonical reference accepted")
			}
		})
	}
}
