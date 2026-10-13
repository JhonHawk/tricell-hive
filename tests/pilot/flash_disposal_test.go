package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Exercise cleanup, the runner's result writer and the persisted assessment.
func TestFlashRunnerDeferredDisposalConsumers(t *testing.T) {
	for _, mode := range []string{"container", "unverified stop", "default", "import failure"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			fixtureFile(t, filepath.Join(dir, "observer.jsonl"), "")
			fixtureFile(t, filepath.Join(dir, "stdout.jsonl"), `{"type":"text","sessionID":"root","part":{"messageID":"final","text":"done"}}`)
			g := &guidanceVariant{containerNative: mode != "default", credentialImported: true}
			if mode == "unverified stop" {
				bin := t.TempDir()
				path := filepath.Join(bin, "ps")
				if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
					t.Fatal(err)
				}
				t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
				// No process is launched: the fake enumeration cannot verify this group.
				g.importGroups = []*credentialProcess{{cmd: &exec.Cmd{Process: &os.Process{Pid: 2147483647}}}}
			}
			g.cleanup()
			if mode == "default" {
				g.report.ShadowCredentialStoreRemoved = true
			}
			r := result{Suite: flashSuite, Host: "opencode", Case: "plan-reference", Arm: "C", Cwd: dir, Terminal: "completed", ModelRequested: flashModels[0], ModelObserved: flashModels[0], GuidanceVariant: &g.report}
			if mode != "default" {
				r.MemoryStoreMode = disposableContainer
			}
			if mode == "import failure" {
				r.Error = "credential import failed"
				r.ExitCode = 1
				r.Terminal = "error"
			}
			r.Trace.TerminalSource = "native_export"
			r.Trace.NativeCompletion = &openCodeCompletionEvidence{SessionID: "root", Directory: dir, FinalAssistantID: "final", AssistantFinish: "stop", SessionOutcome: "succeeded", FinalEvent: "idle"}
			r.FlashClass = &flashClassReport{Role: "hive-review-plan", Memory: deferredVolumeDestruction, MemoryRouting: "observed", Completion: "observed", Coverage: "observed", EvidenceHash: hashOrAbsent(filepath.Join(dir, "observer.jsonl")), PrimaryCompletion: &flashChild{SessionID: "root", Agent: "build", Model: flashModels[0], Completed: true}, Children: []flashChild{{SessionID: "child", ParentID: "root", Agent: "hive-review-plan", AgentName: "Plan Reviewer", Model: flashModels[0], Completed: true, FinalAssistantID: "child-final"}}}
			r.FlashClass.Observations = []flashObservation{{SessionID: "root", Phase: "http", Agent: "build", Model: flashModels[0], Correlated: true, Protocol: "chat-completions"}}
			fixtureFile(t, filepath.Join(dir, "fixture.json"), `{"id":"plan-reference"}`)
			completeGuidanceCleanup(&r, g)
			writeRunResult(filepath.Join(dir, "run.json"), r)
			raw, err := os.ReadFile(filepath.Join(dir, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			var written result
			if err := json.Unmarshal(raw, &written); err != nil {
				t.Fatal(err)
			}
			if mode == "container" {
				if written.Terminal != "completed" || written.GuidanceVariant.ShadowCredentialWarning != "" || written.GuidanceVariant.ShadowCredentialDisposal != "deferred_to_container_destruction" || written.GuidanceVariant.ShadowCredentialStoreRemoved {
					t.Errorf("expected deferred credential disposition: %+v", written)
				}
			} else if mode != "default" && written.Terminal == "completed" {
				t.Error("failed cleanup or import accepted by result writer")
			}
			if err := assessRun(dir, "."); err != nil {
				t.Fatal(err)
			}
			raw, err = os.ReadFile(filepath.Join(dir, "criterion-assessment.json"))
			if err != nil {
				t.Fatal(err)
			}
			var a assessment
			if err := json.Unmarshal(raw, &a); err != nil {
				t.Fatal(err)
			}
			if mode == "container" && (a.Terminal != "completed" || a.Status != "pass") {
				t.Errorf("container assessment terminal %s", a.Terminal)
			}
			if mode == "default" && a.Terminal != "protocol_unverified" {
				t.Errorf("default mode waived memory protocol: %s", a.Terminal)
			}
			if mode == "unverified stop" && (a.Terminal != "credential_cleanup_unverified" || written.GuidanceVariant.ShadowCredentialWarning == "" || written.GuidanceVariant.ShadowCredentialDisposal != "") {
				t.Error("process-stop failure hidden by deferral")
			}
			if mode != "container" && (a.Terminal == "completed" || a.Status == "pass") {
				t.Error("failed or unauthorized deferral accepted by assessment")
			}
		})
	}
}
