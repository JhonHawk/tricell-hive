package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlashPrimaryModelFromNativeExport(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprint(valid), func(t *testing.T) {
			r := result{ModelRequested: flashModels[0], ModelObserved: "not_observed", Trace: traceReport{SessionID: "root", FinalAssistantID: "final"}, FlashClass: &flashClassReport{}}
			primary := flashChild{SessionID: "root", Model: flashModels[0], FinalAssistantID: "final", Agent: "build", Completed: true}
			var err error
			if !valid {
				primary.Model = "other/model"
				primary.Completed = false
				err = fmt.Errorf("model mismatch")
			}
			accepted := recordFlashPrimary(&r, primary, err)
			if accepted != valid {
				t.Fatal("native proof acceptance changed")
			}
			observeFlashPrimary(&r, flashObservation{Agent: "build", AgentName: "build", Model: "http/model"})
			want := "not_observed"
			if valid {
				want = flashModels[0]
			}
			if r.ModelObserved != want {
				t.Errorf("model %s, want native proof %s", r.ModelObserved, want)
			}
			if valid {
				observeFlashPrimary(&r, flashObservation{Agent: "wrong", Model: flashModels[0]})
				if r.FlashClass.PrimaryCompletion.Completed {
					t.Error("HTTP identity mismatch accepted")
				}
			}
		})
	}
}

func TestFlashProtocolTerminalSurvivesOperationGate(t *testing.T) {
	for _, valid := range []bool{true, false} {
		t.Run(fmt.Sprint(valid), func(t *testing.T) {
			dir := t.TempDir()
			fixtureFile(t, filepath.Join(dir, "observer.jsonl"), "")
			fixtureFile(t, filepath.Join(dir, "stdout.jsonl"), `{"type":"text","sessionID":"root","part":{"messageID":"final","text":"done"}}`)
			r := result{Suite: flashSuite, Host: "opencode", Case: "plan-reference", Arm: "C", Cwd: dir, MemoryStoreMode: disposableContainer, ModelRequested: flashModels[0], ModelObserved: flashModels[0], Terminal: "completed"}
			r.Trace.SessionID = "root"
			r.Trace.TerminalSource = "native_export"
			r.Trace.NativeCompletion = &openCodeCompletionEvidence{SessionID: "root", Directory: dir, FinalAssistantID: "final", AssistantFinish: "stop", SessionOutcome: "succeeded", FinalEvent: "idle"}
			r.FlashClass = &flashClassReport{Role: "hive-review-plan", Completion: "observed", Coverage: "observed", Memory: deferredVolumeDestruction, MemoryRouting: "observed", EvidenceHash: hashOrAbsent(filepath.Join(dir, "observer.jsonl")), PrimaryCompletion: &flashChild{SessionID: "root", Agent: "build", Model: flashModels[0], Completed: true, Operations: []traceEvent{{Kind: "tool", Tool: "execute", Command: "console.log('read only')"}}}, Children: []flashChild{{SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], FinalAssistantID: "child-final", Completed: true}}}
			r.FlashClass.Observations = []flashObservation{{SessionID: "root", Phase: "http", Agent: "build", Model: flashModels[0], Correlated: true, Protocol: "chat-completions"}}
			if !valid {
				r.FlashClass.Completion = "unverified"
			}
			gateFlashOperations(&r)
			writeRunResult(filepath.Join(dir, "run.json"), r)
			b, _ := os.ReadFile(filepath.Join(dir, "run.json"))
			var written result
			if err := json.Unmarshal(b, &written); err != nil {
				t.Fatal(err)
			}
			if written.Terminal != "operation_inspection_required" {
				t.Error("automatic gate relaxed")
			}
			if valid && written.ProtocolTerminal != "completed" {
				t.Errorf("protocol lost: %s", written.ProtocolTerminal)
			}
			if !valid && (written.ProtocolTerminal == "" || written.ProtocolTerminal == "completed") {
				t.Error("failed protocol accepted")
			}
			fixtureFile(t, filepath.Join(dir, "fixture.json"), `{"id":"plan-reference"}`)
			if err := assessRun(dir, "."); err != nil {
				t.Fatal(err)
			}
			b, _ = os.ReadFile(filepath.Join(dir, "criterion-assessment.json"))
			var a map[string]any
			if err := json.Unmarshal(b, &a); err != nil {
				t.Fatal(err)
			}
			if valid && a["ProtocolTerminal"] != "completed" {
				t.Error("assessment lost protocol terminal")
			}
			if !valid && (a["ProtocolTerminal"] == nil || a["ProtocolTerminal"] == "completed") {
				t.Error("assessment waived protocol failure")
			}
		})
	}
}

func TestFlashNativeOperationCommandProjection(t *testing.T) {
	for _, role := range []string{"build", "hive-review-plan"} {
		for _, tool := range []string{"shell", "execute"} {
			t.Run(role+"/"+tool, func(t *testing.T) {
				root := "root"
				session := "child"
				parent := root
				if role == "build" {
					parent = ""
					session = root
				}
				command := "cat /fixture/plan.md"
				key := "command"
				if tool == "execute" {
					key = "code"
					command = "console.log(await readFile('/fixture/plan.md'))"
				}
				model := map[string]any{"providerID": "opencode-go", "id": "claude-haiku-5-5", "variant": "high"}
				doc := map[string]any{"info": map[string]any{"id": session, "parentID": parent, "agent": role, "model": model, "outcome": "succeeded", "time": map[string]any{"idle": 1}}, "messages": []any{map[string]any{"id": "final", "type": "assistant", "agent": role, "model": model, "finish": "stop", "time": map[string]any{"completed": 1}, "content": []any{map[string]any{"type": "tool", "id": "op", "name": tool, "state": map[string]any{"status": "completed", "input": map[string]any{key: command, "private": "FAKE_PRIVATE_INPUT"}, "output": "FAKE_PRIVATE_OUTPUT"}}}}}}
				raw, _ := json.Marshal(doc)
				c, err := flashChildFromExport(raw, parent, role, flashModels[0])
				if err != nil {
					t.Fatal(err)
				}
				if tool == "execute" {
					command = ""
				}
				if len(c.Operations) != 1 || c.Operations[0].Command != command || (tool == "execute" && !c.Operations[0].CommandWithheld) {
					t.Error("native command/code omitted")
				}
				evidence, _ := json.Marshal(c)
				if strings.Contains(string(evidence), "FAKE_PRIVATE") {
					t.Error("private unrelated input/output retained")
				}
			})
		}
	}
}

func TestFlashPrimaryRequiresMatchingNativeModel(t *testing.T) {
	for _, provider := range []string{"opencode-go", "other"} {
		t.Run(provider, func(t *testing.T) {
			nativeModel := map[string]any{"providerID": provider, "id": "claude-haiku-5-5", "variant": "high"}
			doc := map[string]any{"info": map[string]any{"id": "root", "parentID": "", "agent": "build", "model": nativeModel, "outcome": "succeeded", "time": map[string]any{"idle": 1}}, "messages": []any{map[string]any{"id": "final", "type": "assistant", "agent": "build", "model": nativeModel, "finish": "stop", "time": map[string]any{"completed": 1}}}}
			raw, _ := json.Marshal(doc)
			primary, err := flashChildFromExport(raw, "", "build", flashModels[0])
			r := result{ModelRequested: flashModels[0], ModelObserved: "not_observed", Trace: traceReport{SessionID: "root", FinalAssistantID: "final"}, FlashClass: &flashClassReport{}}
			accepted := recordFlashPrimary(&r, primary, err)
			if provider == "opencode-go" {
				if !accepted || r.ModelObserved != flashModels[0] {
					t.Fatal("matching native model not recorded")
				}
			} else if accepted || r.ModelObserved != "not_observed" {
				t.Fatal("mismatched native model accepted")
			}
		})
	}
}

func TestFlashOperationCommandWithholdsSensitiveText(t *testing.T) {
	for _, command := range []string{"curl -H 'Authorization: Bearer FAKE_PRIVATE_TOKEN' https://example.invalid", "API_KEY=FAKE_PRIVATE_VALUE command", "execute('password=FAKE_PRIVATE_VALUE')"} {
		if flashOperationCommand(command) != "" {
			t.Error("sensitive command persisted")
		}
	}
	if flashOperationCommand("") != "" {
		t.Error("missing command fabricated")
	}
}

func TestFlashProtocolRequiresPrimaryHTTPObserver(t *testing.T) {
	for _, mode := range []string{"observed", "missing", "wrong model", "wrong agent", "uncorrelated"} {
		t.Run(mode, func(t *testing.T) {
			r := result{Arm: "B", ModelRequested: flashModels[0], ModelObserved: flashModels[0], Trace: traceReport{SessionID: "root"}, FlashClass: &flashClassReport{Role: "hive-review-plan", Completion: "observed", Coverage: "unverified", Memory: "verified_clean", MemoryRouting: "observed", EvidenceHash: "hash", PrimaryCompletion: &flashChild{SessionID: "root", Agent: "build", Model: flashModels[0], Completed: true}, Children: []flashChild{{SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], Completed: true, FinalAssistantID: "final"}}}}
			o := flashObservation{SessionID: "root", Phase: "http", Agent: "build", Model: flashModels[0], Correlated: true, Protocol: "chat-completions"}
			switch mode {
			case "wrong model":
				o.Model = "other/model"
			case "wrong agent":
				o.Agent = "other"
			case "uncorrelated":
				o.Correlated = false
			}
			if mode != "missing" {
				r.FlashClass.Observations = []flashObservation{o}
			}
			got := flashProtocolTerminal(r, "hash")
			if mode == "observed" && got != "completed" {
				t.Errorf("valid HTTP proof failed: %s", got)
			}
			if mode != "observed" && got == "completed" {
				t.Error("native model substituted for missing or invalid HTTP proof")
			}
		})
	}
}
