package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestE2EGrantDeadlineAndAttempt(t *testing.T) {
	now := time.Now()
	for _, d := range []time.Time{now.Add(-time.Second), now.Add(21 * time.Minute), now.Add(time.Minute)} {
		if _, err := newE2EBudget(now, d); err == nil {
			t.Fatalf("accepted invalid deadline %v", d)
		}
	}
	b, err := newE2EBudget(now, now.Add(20*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if got := b.activeDeadline.Sub(now); got != 18*time.Minute {
		t.Fatal(got)
	}
	p := filepath.Join(t.TempDir(), "attempt")
	if err := claimE2EAttempt(p); err != nil {
		t.Fatal(err)
	}
	if err := claimE2EAttempt(p); err == nil {
		t.Fatal("second attempt accepted")
	}
}

func TestE2EClassificationNeverTrustsSummary(t *testing.T) {
	r := e2eReport{ModelStarted: true, ExitCode: 0, JointExitCode: 0, JointChecked: true, IntegrationHead: "abc"}
	if c := classifyE2E(r); c != "inconclusive" {
		t.Fatal("unproven native behavior accepted", c)
	}
	r.JointExitCode = 1
	if c := classifyE2E(r); c != "failed" {
		t.Fatal(c)
	}
	r.ModelStarted = false
	if c := classifyE2E(r); c != "blocked" {
		t.Fatal(c)
	}
}

func TestE2ELaunchIsCaseScoped(t *testing.T) {
	for _, id := range []string{"smoke", e2eCaseID} {
		args, err := launchArgs(result{Host: "claude", Case: id, ModelRequested: "opus"}, "/out", "prompt")
		if err != nil {
			t.Fatal(err)
		}
		wants := id == e2eCaseID
		if slices.Contains(args, "--no-session-persistence") == wants {
			t.Fatal(args)
		}
		var tools string
		for i, a := range args {
			if a == "--allowedTools" {
				tools = args[i+1]
			}
		}
		if containsTool(tools, "Agent") != wants || containsTool(tools, "Skill") != wants {
			t.Fatal(args)
		}
		if containsTool(tools, "TaskOutput") {
			t.Fatal("removed tool allowed")
		}
	}
}

func TestE2EFixtureJointCheckExercisesBothUnits(t *testing.T) {
	root := t.TempDir()
	for name, body := range e2eFiles() {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "-C", root, "test", "./...")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if err := cmd.Run(); err == nil {
		t.Fatal("faulty/partial baseline passes")
	}
	for name, body := range map[string]string{"a/value.go": "package a\nfunc Value(n int) int { if n<0 {return 0};return n*2 }\n", "b/label.go": "package b\nimport \"strconv\"\nfunc Label(n int) string { return \"value=\"+strconv.Itoa(n) }\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	cmd = exec.Command("go", "-C", root, "test", "./...")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s: %v", out, err)
	}
}

func TestE2EImmutableViolationPersistsFailedOutcome(t *testing.T) {
	root := t.TempDir()
	for name, body := range e2eFiles() {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, mutation := range []string{"changed", "missing"} {
		t.Run(mutation, func(t *testing.T) {
			report := e2eReport{ModelStarted: true, ExitCode: 0}
			if err := os.WriteFile(filepath.Join(root, "acceptance/acceptance_test.go"), []byte(e2eFiles()["acceptance/acceptance_test.go"]), 0600); err != nil {
				t.Fatal(err)
			}
			if err := verifyE2EImmutableFiles(root, &report); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(root, "acceptance/acceptance_test.go")
			var err error
			if mutation == "changed" {
				err = os.WriteFile(path, []byte("package acceptance\n"), 0600)
			} else {
				err = os.Remove(path)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := verifyE2EImmutableFiles(root, &report); err == nil {
				t.Fatal("weakened acceptance fixture accepted")
			}
			evidence := filepath.Join(t.TempDir(), "evidence.json")
			if err := e2eWrite(evidence, report); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(evidence)
			if err != nil {
				t.Fatal(err)
			}
			var persisted e2eReport
			if err := json.Unmarshal(data, &persisted); err != nil {
				t.Fatal(err)
			}
			class := classifyE2E(persisted)
			if class != "failed" || persisted.RunnerFailure == "" || e2eResultExitCode(class) == 0 {
				t.Fatalf("proven fixture violation lost: class=%s report=%+v", class, persisted)
			}
		})
	}
	unknown := e2eReport{ModelStarted: true, ExitCode: 0, Reason: "integration evidence unavailable"}
	if class := classifyE2E(unknown); class != "inconclusive" {
		t.Fatalf("unknown evidence misclassified: %s", class)
	}
}

func TestE2ENativeObserverRequiresCompletionAndInjectsOnce(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	wt := filepath.Join(root, ".claude/worktrees/unit-a")
	if err := os.MkdirAll(filepath.Join(wt, "a"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, "a/value.go"), []byte("package a\nfunc Value(n int) int {if n<0{return 0};return n*2}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := initGit(wt); err != nil {
		t.Fatal(err)
	}
	if err := setupInitialCommit(wt); err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput(wt, "branch", "-M", "unit-a"); err != nil {
		t.Fatal(err)
	}
	hooks := filepath.Join(out, "hooks")
	if err := os.Mkdir(hooks, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hooks, "pre-commit"), []byte("#!/bin/sh\n: > \"$HIVE_E2E_OUTPUT/hook-observed\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := gitOutput(wt, "config", "core.hooksPath", hooks); err != nil {
		t.Fatal(err)
	}
	before, _ := gitOutput(wt, "rev-parse", "HEAD")
	script := filepath.Join(out, "observer.py")
	if err := os.WriteFile(script, []byte(e2eObserver), 0600); err != nil {
		t.Fatal(err)
	}
	invoke := func(status string) {
		t.Helper()
		cmd := exec.Command("python3", script)
		cmd.Env = append(os.Environ(), "HIVE_E2E_ROOT="+root, "HIVE_E2E_OUTPUT="+out, fmt.Sprintf("HIVE_E2E_ACTIVE_DEADLINE=%d", time.Now().Add(time.Minute).Unix()))
		cmd.Stdin = strings.NewReader(`{"tool_name":"Agent","tool_input":{"subagent_type":"hive-build-owner","description":"unit-A initial owner"},"tool_response":{"status":"` + status + `","agentId":"native-a","resolvedModel":"model"}}`)
		if data, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v", data, err)
		}
	}
	invoke("async_launched")
	current, _ := gitOutput(wt, "rev-parse", "HEAD")
	if current != before {
		t.Fatal("mutated a live writer")
	}
	invoke("completed")
	if _, err := os.Stat(filepath.Join(out, "hook-observed")); err != nil {
		t.Fatal("configured commit hook was bypassed", err)
	}
	after, _ := gitOutput(wt, "rev-parse", "HEAD")
	if after == before {
		t.Fatal("controlled mutation did not change head")
	}
	data, err := os.ReadFile(filepath.Join(out, "fixture-mutations.jsonl"))
	if err != nil || !strings.Contains(string(data), `"kind": "fixture_mutation"`) {
		t.Fatalf("missing labeled evidence: %s %v", data, err)
	}
	invoke("completed")
	current, _ = gitOutput(wt, "rev-parse", "HEAD")
	if current != after {
		t.Fatal("injected twice")
	}
}

func TestE2EBudgetKillsOnlyStartedProcessGroup(t *testing.T) {
	// A real local sleep, not simulated model output. runProcess already owns its
	// process-group cleanup; this checks the case's remaining-time call boundary.
	deadline := time.Now().Add(30 * time.Millisecond)
	_, code, err := e2eCommand(deadline, t.TempDir(), os.Environ(), "/bin/sh", "-c", "sleep 10 & wait")
	if err == nil || code != -1 {
		t.Fatalf("deadline did not stop group: %d %v", code, err)
	}
	_, code, err = e2eCommand(time.Now().Add(-time.Second), t.TempDir(), os.Environ(), "/bin/sh", "-c", "exit 0")
	if err == nil || code != -1 {
		t.Fatal("started after deadline")
	}
}

func TestE2EReviewNeedsIndependentAnchoredEvidence(t *testing.T) {
	out := t.TempDir()
	raw := []byte("native completed owner and verifier evidence\n")
	if err := os.WriteFile(filepath.Join(out, "native-events.jsonl"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	r := e2eReport{ModelStarted: true, ExitCode: 0, JointChecked: true, JointExitCode: 0, IntegrationHead: strings.Repeat("a", 40), Launch: result{Trace: traceReport{SessionID: "coordinator"}}}
	review := e2eReview{ReviewerSessionID: "coordinator", RunHash: "run"}
	for _, id := range e2eReviewCriteria {
		review.Criteria = append(review.Criteria, e2eReviewedCriterion{ID: id, Status: "pass", Explanation: "independent native and Git evidence inspected", Evidence: []e2eEvidenceAnchor{{Path: "native-events.jsonl", SHA256: digest(raw), Line: 1}}})
	}
	if err := validateE2EReview(out, r, "run", review); err == nil {
		t.Fatal("self-review accepted")
	}
	review.ReviewerSessionID = "independent-root-session"
	if err := validateE2EReview(out, r, "run", review); err != nil {
		t.Fatal(err)
	}
	r.Review = &review
	if got := classifyE2E(r); got != "pass" {
		t.Fatal(got)
	}
	review.Criteria[0].Evidence[0].SHA256 = "wrong"
	if err := validateE2EReview(out, r, "run", review); err == nil {
		t.Fatal("stale evidence accepted")
	}
}

func TestE2ESetupDeadlinesOwnDescendantsAndPipes(t *testing.T) {
	for _, stage := range []string{"github-status", "github-token", "memory-projects", "memory-identity"} {
		t.Run(stage, func(t *testing.T) {
			bin := t.TempDir()
			output := t.TempDir()
			pidfile := filepath.Join(output, "child.pid")
			parentfile := filepath.Join(output, "parent.pid")
			hang := "/bin/sleep 30 &\necho $! > \"$HIVE_SETUP_CHILD_PID\"\necho $$ > \"$HIVE_SETUP_PARENT_PID\"\nwait\n"
			script := "#!/bin/sh\n"
			switch stage {
			case "github-token":
				script += "[ \"$1 $2\" = 'auth status' ] && exit 1\n" + hang
			case "github-status":
				script += hang
			default:
				script += hang
			}
			name := "engram"
			if strings.HasPrefix(stage, "github") {
				name = "gh"
			}
			if err := os.WriteFile(filepath.Join(bin, name), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			env := append(os.Environ(), "HIVE_SETUP_CHILD_PID="+pidfile, "HIVE_SETUP_PARENT_PID="+parentfile)
			killOwned := func() {
				for _, path := range []string{pidfile, parentfile} {
					data, _ := os.ReadFile(path)
					pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
					if pid > 0 {
						_ = syscall.Kill(pid, syscall.SIGKILL)
					}
				}
			}
			// Bound the RED run as well: a descendant's inherited pipe otherwise holds
			// exec.Output indefinitely after CommandContext kills only its parent.
			watchdog := time.AfterFunc(800*time.Millisecond, killOwned)
			defer watchdog.Stop()
			defer killOwned()
			started := time.Now()
			deadline := started.Add(80 * time.Millisecond)
			var err error
			switch stage {
			case "github-status", "github-token":
				_, _, err = prepareGitHubIsolationWithin(output, env, deadline)
			case "memory-projects":
				_, _, err = prepareMemoryIsolationWithin(output, env, deadline)
			case "memory-identity":
				m := &memoryIsolation{}
				_, err = m.startHTTPWithin(env, output, deadline)
			}
			elapsed := time.Since(started)
			if err == nil {
				t.Fatal("unfinished setup accepted")
			}
			if elapsed > 500*time.Millisecond {
				t.Errorf("setup exceeded remaining deadline while descendant held stdout: %s", elapsed)
			}
			for _, path := range []string{pidfile, parentfile} {
				data, readErr := os.ReadFile(path)
				if readErr != nil {
					t.Fatal(readErr)
				}
				pid, parseErr := strconv.Atoi(strings.TrimSpace(string(data)))
				if parseErr != nil || pid <= 0 {
					t.Fatalf("invalid owned PID: %q", data)
				}
				state, _ := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "stat=").Output()
				stateText := strings.TrimSpace(string(state))
				t.Logf("owned setup PID %d state after return: %q", pid, stateText)
				if stateText != "" && !strings.HasPrefix(stateText, "Z") {
					t.Errorf("owned setup process %d survived return: %s", pid, stateText)
				}
			}
		})
	}
}
