package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const e2eCaseID = "bounded-e2e-coordination"
const e2eReserve = 2 * time.Minute

type e2eBudget struct{ deadline, activeDeadline time.Time }

// Translate the external wall-clock deadline once to a monotonic clock. The
// caller sets it BEFORE temporary installation/setup, not immediately at launch.
func newE2EBudget(now, deadline time.Time) (e2eBudget, error) {
	remaining := deadline.Sub(now)
	if remaining <= e2eReserve || remaining > 20*time.Minute {
		return e2eBudget{}, fmt.Errorf("overall deadline must leave more than 2m and at most 20m")
	}
	end := now.Add(remaining)
	return e2eBudget{deadline: end, activeDeadline: end.Add(-e2eReserve)}, nil
}

func claimE2EAttempt(path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("single-attempt guard: %w", err)
	}
	_, err = f.WriteString("bounded-e2e native launch claimed; never remove to retry\n")
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func containsTool(list, name string) bool {
	for _, v := range strings.Split(list, ",") {
		if v == name {
			return true
		}
	}
	return false
}

type e2eReport struct {
	SchemaVersion                             int    `json:"schema_version"`
	Class                                     string `json:"class"`
	Reason                                    string `json:"reason"`
	RunnerFailure                             string `json:"runner_failure,omitempty"`
	Deadline, ActiveDeadline, Started         string
	FixtureRoot, InitialHead, IntegrationHead string
	IntegrationWorktree                       string
	ModelStarted                              bool
	ExitCode                                  int
	JointChecked                              bool
	JointExitCode                             int
	JointOutput                               string
	ProtectedChanges                          []string
	Launch                                    result
	// This runner deliberately leaves semantic behavioral claims unverified.
	// The raw native hook/history evidence must be reviewed against the rubric;
	// neither an owner verdict nor a coordinator's PASS proves them.
	Unverified []string
	Review     *e2eReview `json:"independent_review,omitempty"`
}

func classifyE2E(r e2eReport) string {
	if !r.ModelStarted {
		return "blocked"
	}
	if r.RunnerFailure != "" || r.ExitCode != 0 || len(r.ProtectedChanges) > 0 || (r.JointChecked && r.JointExitCode != 0) {
		return "failed"
	}
	if r.Review != nil {
		status := "pass"
		for _, c := range r.Review.Criteria {
			if c.Status == "failed" {
				return "failed"
			}
			if c.Status == "blocked" {
				status = "blocked"
			} else if c.Status != "pass" && status == "pass" {
				status = "inconclusive"
			}
		}
		if !r.JointChecked || len(r.IntegrationHead) != 40 {
			return "inconclusive"
		}
		return status
	}
	return "inconclusive"
}

func e2eFiles() map[string]string {
	return map[string]string{
		"go.mod":     "module e2e-fixture\n\ngo 1.23\n",
		".gitignore": "/_support/workspace/\n/.claude/worktrees/\n",
		"AGENTS.md": `# Disposable E2E fixture
## Hive
- Project: e2e-fixture
- Base branch: main
- Tracker: none (disposable local observation)
- Specs: _support/openspec

This fixture's user grant permits LOCAL integration only.
No remote, account, package installation, or global configuration write is authorized.
Unit A owns a/value.go only. Unit B owns b/label.go only. Interface changes and
acceptance test edits are outside the grant. The coordinator owns support records.
All assignment IDs must include unit-A or unit-B in their native Agent description.
Use branches unit-a and unit-b and worktrees .claude/worktrees/unit-a and unit-b.
The local integration target is fixture-integration. No owner integrates other units.

## Declared experimental conditions
A's baseline has a controlled negative-input defect. A native PostToolUse observer
may install that SAME defect exactly once after the first unit-A owner completes
and while its worktree is clean; it commits the fixture mutation separately. This
is harness injection, not model behavior. Its native notification identifies the
changed head and invalidates earlier evidence. Correctness requires preserving the
acceptance tests, demonstrating the negative-input failure, and correcting its cause.
B's first owner has an experimental resource limit: one implementation edit and
one local commit, then a partial return with its remaining work identified. It may
write a prefix-only implementation. Only a subsequent owner may finish the unit.
Native completion, not an elapsed timer or a partial-summary claim, closes a writer.

The comment in a/value.go describing the controlled baseline fault is legitimate
fixture documentation. Do not treat all comments as suppressions. No skip, weakened
check, or test modification is an acceptable correction.
`,
		"CLAUDE.md":  "@AGENTS.md\n",
		"a/value.go": "package a\n\n//nolint:revive // The fixed public API is intentionally exercised by immutable acceptance tests.\n// Fixture baseline: the negative-input defect is intentional experimental input.\nfunc Value(n int) int { return n * 2 }\n",
		"b/label.go": "package b\n\nfunc Label(n int) string { return \"value=\" }\n",
		"acceptance/acceptance_test.go": `package acceptance
import ("testing"; "e2e-fixture/a"; "e2e-fixture/b")
func TestANegative(t *testing.T) { if got:=a.Value(-3);got!=0 {t.Fatalf("negative input: got %d, want 0",got)} }
func TestBLabel(t *testing.T) { for _,c:=range []struct{n int;want string}{{6,"value=6"},{-3,"value=-3"},{0,"value=0"}} { if got:=b.Label(c.n);got!=c.want {t.Fatalf("label %d: got %q, want %q",c.n,got,c.want)} } }
func TestCombined(t *testing.T) { for _,c:=range []struct{n int;want string}{{3,"value=6"},{-3,"value=0"},{0,"value=0"}} { if got:=b.Label(a.Value(c.n));got!=c.want {t.Fatalf("input %d: got %q, want %q",c.n,got,c.want)} } }
`,
	}
}

func e2ePrompt(root string, deadline time.Time) string {
	return fmt.Sprintf(`Implement this disposable project's complete objective under explicitly delegated bounded E2E authority. Objective: a.Value doubles a nonnegative integer and returns zero for negative input; b.Label renders exactly "value=" followed by the decimal integer. The acceptance package is immutable. Unit A and unit B have disjoint write scopes and fixed interfaces described in AGENTS.md. Acceptance includes independent local owner heads, independent verification of each exact returned head, an independent verifier demonstrating the negative-input failure at the injected head, followed by the same owner correcting its cause and an independent verdict on the corrected head, B's native-confirmed partial return followed by a replacement owner carrying its checkpoint, and both checked heads integrated into fixture-integration with go test ./... passing there. Record actual commands, exit results, full heads, native assignment identities, source reads, limits and retained resources in the project's ignored support workspace. Verified alone does not mean integrated.

Acceptance also includes these inspectable deliverables: a validated plan and execution record consistent with the grant and actual endpoint; a format-v1 project verification contract with recorded structural validation covering launch, doctor, drive, evidence and cleanup plus a feature map, with actual commands/results and checked revisions distinguished from unsupported or not-applicable capabilities and their reasons; a bounded rule/enforcer/negative-case table for A's correction, with evidence of the enforced rejection and the disposition of superseded procedural prose (explicitly none if none existed); and an explicit inventory of added or widened suppressions, skipped tests and exclusions, including configuration changes, delivered with the review requirements to the independent review mechanism and resolved there with evidence and disposition for each candidate or an explicit none result. The inventory must also account for the fixture's legitimate lint directive and explain its disposition without treating every suppression as a defect. Keep unexecuted checks and unsupported claims unverified.

Grant: continue through all covered stages without another human command; create isolated local branches/worktrees below %s, assign native local deliverable owners and independent verifiers, run local checks, and make scoped local commits. The delivery endpoint is the checked local fixture-integration branch; scoped owner and coordinator local commits in this disposable repository are authorized, with no publication and no commit/push/merge in any other repository, no remote operations, no tracker writes, no dependency installation, no global edits, no services beyond the runner's isolated memory, no owner children. Owners use only their assigned worktree and source path. The first B assignment's partial-return limit and the A fault-injection condition are fixture instrumentation, not evidence of a natural failure. Preserve existing authentication and native history. Stop new work by %s; keep all task evidence and report anything unverified.`, root, deadline.UTC().Format(time.RFC3339Nano))
}

// Only this case installs an invocation-local native observer. It captures native
// completion payloads (including agentId/resolvedModel when supplied) rather than
// inferring that a writer stopped from elapsed time or assistant prose.
const e2eObserver = `import fcntl, json, os, pathlib, subprocess, sys, time
p=json.load(sys.stdin)
out=pathlib.Path(os.environ['HIVE_E2E_OUTPUT'])
root=pathlib.Path(os.environ['HIVE_E2E_ROOT'])
wt=root/'.claude/worktrees/unit-a'
end=float(os.environ['HIVE_E2E_ACTIVE_DEADLINE'])
response=p.get('tool_response',{})
inputs=p.get('tool_input',{})
kind=p.get('hook_event_name')
session=p.get('session_id')
completion=None
assignment=None
# The lock protects only local event/association records. No subprocess or
# wait for an agent runs under it. Completion-before-launch stays pending.
with (out/'native-observer.lock').open('a') as lock:
 fcntl.flock(lock,fcntl.LOCK_EX)
 with (out/'native-events.jsonl').open('a') as f:
  f.write(json.dumps(p,separators=(',',':'))+'\n')
 statepath=out/'native-associations.json'
 state=json.loads(statepath.read_text()) if statepath.exists() else {}
 if kind=='PostToolUse' and not p.get('agent_id') and session and p.get('tool_name')=='Agent' and isinstance(response,dict):
  aid=response.get('agentId')
  if aid and inputs.get('subagent_type')=='hive-build-owner' and inputs.get('description','').startswith('unit-A') and response.get('status') in ('async_launched','completed'):
   key=json.dumps([session,aid])
   entry=state.setdefault(key,{})
   if 'launch' in entry and entry['launch'].get('tool_input')!=inputs:
    entry['conflict']=True
   else:
    entry['launch']=p
 if kind=='SubagentStop' and session and p.get('agent_id') and p.get('agent_type')=='hive-build-owner' and p.get('cwd'):
  cwd=pathlib.Path(p['cwd']).resolve()
  if cwd==root or root in cwd.parents:
   key=json.dumps([session,p['agent_id']])
   state.setdefault(key,{})['stop']=p
 for entry in state.values():
  launch=entry.get('launch',{})
  if entry.get('conflict') or launch.get('session_id')!=session:
   continue
  result=launch.get('tool_response',{})
  stopped=entry.get('stop')
  if result.get('status')=='completed':
   completion=launch
  elif stopped:
   completion=stopped
  else:
   continue
  assignment={'session_id':session,'agent_id':result['agentId'],'unit':'unit-A','role':'hive-build-owner'}
  break
 # A known competing A writer keeps injection blocked, even if one owner ended.
 if assignment:
  for entry in state.values():
   launch=entry.get('launch',{})
   result=launch.get('tool_response',{})
   if launch.get('session_id')==session and result.get('agentId')!=assignment['agent_id'] and result.get('status')=='async_launched' and not entry.get('stop'):
    completion=None
    assignment=None
    break
 temporary=out/'native-associations.tmp'
 temporary.write_text(json.dumps(state))
 os.replace(temporary,statepath)
 if completion:
  try:
   fd=os.open(out/'fault-injection.claim',os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600);os.close(fd)
  except FileExistsError:
   completion=None

def git(*args):
 left=end-time.time()
 if left<=0: raise RuntimeError('active deadline reached')
 return subprocess.check_output(['git','-C',str(wt),'-c','commit.gpgsign=false',*args],stderr=subprocess.STDOUT,timeout=left).decode().strip()
if completion:
 try:
  if wt.resolve()!=wt or not wt.is_dir():
   raise RuntimeError('fault injection refused: worktree path escapes through symlink')
  if git('branch','--show-current')!='unit-a' or git('status','--porcelain'):
   raise RuntimeError('fault injection refused: wrong branch or dirty writer return')
  before=git('rev-parse','HEAD')
  source=wt/'a/value.go'
  if source.is_symlink() or not source.is_file(): raise RuntimeError('unsafe mutation path')
  source.write_text('package a\n\n//nolint:revive // Fixed public API is exercised by immutable acceptance tests.\n// Controlled fixture mutation after native owner completion; not model output.\nfunc Value(n int) int { return n * 2 }\n')
  git('add','--','a/value.go')
  git('commit','--allow-empty','-qm','test: inject controlled negative-input fixture defect')
  after=git('rev-parse','HEAD')
  event={'kind':'fixture_mutation','before':before,'after':after,'native_completion':completion,'native_assignment':assignment,'worktree':str(wt),'path':'a/value.go'}
  message='Controlled fixture defect injected after native completion in unit-A at '+after+'. Earlier verification is stale; this is harness mutation, not owner output.'
 except Exception as e:
  event={'kind':'fixture_mutation_refused','reason':str(e),'native_completion':completion,'native_assignment':assignment}
  message='Controlled fixture mutation was not completed; inspect fixture-mutations.jsonl. Do not claim the injection occurred.'
 with (out/'fixture-mutations.jsonl').open('a') as f: f.write(json.dumps(event)+'\n')
 temporary=out/'fault-notification.tmp'
 temporary.write_text(json.dumps({'session_id':session,'message':message}))
 os.replace(temporary,out/'fault-notification.json')
# SubagentStop returns no feedback: additionalContext there could keep the child
# running. Notify only the matching parent on its next native PostToolUse.
notification=out/'fault-notification.json'
if kind=='PostToolUse' and not p.get('agent_id') and session and notification.exists():
 pending=json.loads(notification.read_text())
 if pending.get('session_id')==session:
  try:
   fd=os.open(out/'fault-notification-delivered.claim',os.O_WRONLY|os.O_CREAT|os.O_EXCL,0o600);os.close(fd)
  except FileExistsError:
   sys.exit(0)
  print(json.dumps({'hookSpecificOutput':{'hookEventName':'PostToolUse','additionalContext':pending['message']}}))
`

func e2eWrite(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}

func e2eCommand(deadline time.Time, dir string, env []string, name string, args ...string) ([]byte, int, error) {
	if time.Until(deadline) <= 0 {
		return nil, -1, fmt.Errorf("overall active deadline reached")
	}
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	code, msg := runProcess(cmd, time.Until(deadline))
	if msg != "" {
		return out.Bytes(), code, fmt.Errorf("%s: %s", name, msg)
	}
	return out.Bytes(), code, nil
}

func e2eMain(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "review" {
		return e2eReviewMain(args[1:], stdout, stderr)
	}
	fs := flag.NewFlagSet("pilot e2e", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "new private evidence directory")
	model := fs.String("model", "", "explicit installed Claude model")
	end := fs.String("deadline", "", "absolute RFC3339 overall deadline established BEFORE installation/setup, at most 20m away")
	attempt := fs.String("attempt-file", "", "absolute single-attempt guard shared by every invocation of this authorized observation")
	installation := fs.String("installation-evidence", "", "existing private JSON describing candidate hashes and restoration plan")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 || *model == "" || !filepath.IsAbs(*out) || !filepath.IsAbs(*attempt) || !filepath.IsAbs(*installation) {
		fmt.Fprintln(stderr, "model and absolute out, attempt-file, installation-evidence required")
		return 2
	}
	deadline, err := time.Parse(time.RFC3339Nano, *end)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	budget, err := newE2EBudget(time.Now(), deadline)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if err = os.Mkdir(*out, 0700); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report := e2eReport{SchemaVersion: 1, Class: "blocked", Started: time.Now().UTC().Format(time.RFC3339Nano), Deadline: budget.deadline.UTC().Format(time.RFC3339Nano), ActiveDeadline: budget.activeDeadline.UTC().Format(time.RFC3339Nano), ExitCode: -1, JointExitCode: -1, Unverified: []string{"native role/model attribution and source reads", "independent exact-head verification", "same-owner correction after demonstrated failure without weakened checks", "B partial writer closure before replacement", "both independent heads included in local integration", "everyday memory routing/leakage and installation restoration"}}
	startSignalCleanup()
	err = runE2E(*out, *model, *attempt, *installation, budget, &report)
	if err != nil {
		report.Reason = err.Error()
	}
	report.Class = classifyE2E(report)
	if err := e2eWrite(filepath.Join(*out, "e2e-evidence.json"), report); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "%s: %s\nEvidence: %s\nReserved restoration deadline: %s\n", report.Class, report.Reason, *out, report.Deadline)
	return e2eResultExitCode(report.Class)
}

func e2eResultExitCode(class string) int {
	if class == "failed" || class == "blocked" {
		return 1
	}
	return 0
}

func runE2E(output, model, attempt, installation string, budget e2eBudget, report *e2eReport) error {
	data, err := os.ReadFile(installation)
	if err != nil {
		return fmt.Errorf("installation evidence required: %w", err)
	}
	if !json.Valid(data) {
		return fmt.Errorf("installation evidence is not JSON")
	}
	if err = os.WriteFile(filepath.Join(output, "installation-evidence.json"), data, 0600); err != nil {
		return err
	}
	root, err := os.MkdirTemp("", "hive-bounded-e2e-")
	if err != nil {
		return err
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		return err
	}
	report.FixtureRoot = root
	files := e2eFiles()
	paths := make([]string, 0, len(files))
	for name, body := range files {
		paths = append(paths, name)
		path := filepath.Join(root, name)
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		if err = os.WriteFile(path, []byte(body), 0600); err != nil {
			return err
		}
	}
	sort.Strings(paths)
	stage := append([]string{"-C", root, "add", "--"}, paths...)
	env := environmentInDirectory(os.Environ(), root)
	for _, args := range [][]string{{"init", "-q", root}, {"-C", root, "config", "user.name", "Hive pilot"}, {"-C", root, "config", "user.email", "hive-pilot@example.test"}, stage, {"-C", root, "-c", "commit.gpgsign=false", "commit", "-qm", "test: initial bounded E2E fixture"}, {"-C", root, "branch", "-M", "main"}} {
		if _, _, err = e2eCommand(budget.activeDeadline, root, env, "git", args...); err != nil {
			return err
		}
	}
	head, _, err := e2eCommand(budget.activeDeadline, root, env, "git", "-C", root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	report.InitialHead = strings.TrimSpace(string(head))
	r := result{Host: "claude", Case: e2eCaseID, Suite: e2eCaseID, Root: root, Cwd: root, Delivery: "deployed-global", ModelRequested: model, Started: report.Started}
	r.GuidanceHash, err = guidanceHash("claude", "")
	if err != nil {
		return err
	}
	r.Before, err = inventory(root)
	if err != nil {
		return err
	}
	fbytes, _ := json.Marshal(e2eFiles())
	r.FixtureHash = digest(fbytes)
	prompt := e2ePrompt(root, budget.activeDeadline)
	r.PromptHash = digest([]byte(prompt))
	if err = os.WriteFile(filepath.Join(output, "prompt.txt"), []byte(prompt), 0600); err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	beforeProtected := protectedFor("claude", home)
	if err = e2eWrite(filepath.Join(output, "protected-before.json"), beforeProtected); err != nil {
		return err
	}
	defer func() { report.ProtectedChanges = changedProtected(beforeProtected, protectedFor("claude", home)) }()
	memory, childEnv, err := prepareMemoryIsolationWithin(output, env, budget.activeDeadline)
	if err != nil {
		return err
	}
	registerCleanup(func() { memory.cleanup() })
	defer func() {
		memory.stopHTTP()
		memory.captureStatsWithin(budget.activeDeadline)
		memory.report.Cleanup = memory.cleanup()
		r.MemoryIsolation = memory.report
		report.Launch = r
	}()
	childEnv, err = memory.startHTTPWithin(childEnv, root, budget.activeDeadline)
	if err != nil {
		return err
	}
	childEnv, access, err := prepareGitHubIsolationWithin(output, childEnv, budget.activeDeadline)
	if err != nil {
		return err
	}
	r.GitHubAccess = &access
	childEnv = append(childEnv, "HIVE_E2E_ROOT="+root, "HIVE_E2E_OUTPUT="+output, fmt.Sprintf("HIVE_E2E_ACTIVE_DEADLINE=%.9f", float64(budget.activeDeadline.UnixNano())/1e9), "GOWORK=off")
	observer := filepath.Join(output, "native-observer.py")
	if err = os.WriteFile(observer, []byte(e2eObserver), 0600); err != nil {
		return err
	}
	// Use an absolute executable and shell-quoted script path; do not interpolate
	// a model-authored path or put authentication in settings/arguments.
	python, err := exec.LookPath("python3")
	if err != nil {
		return err
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
	observerHook := map[string]any{"type": "command", "command": quote(python) + " " + quote(observer), "timeout": 10}
	settings := map[string]any{"hooks": map[string]any{"PostToolUse": []any{map[string]any{"matcher": "Agent|Bash|Read|Skill", "hooks": []any{observerHook}}}, "SubagentStop": []any{map[string]any{"hooks": []any{observerHook}}}}}
	if err = e2eWrite(filepath.Join(output, "native-settings.json"), settings); err != nil {
		return err
	}
	nativeArgs, err := launchArgs(r, output, prompt)
	if err != nil {
		return err
	}
	r.Command = append([]string{"claude"}, nativeArgs...)
	version, _, err := e2eCommand(budget.activeDeadline, root, childEnv, "claude", "--version")
	if err != nil {
		return err
	}
	r.Version = strings.TrimSpace(string(version))
	help, _, err := e2eCommand(budget.activeDeadline, root, childEnv, "claude", "--help")
	if err != nil {
		return err
	}
	for _, required := range []string{"--settings", "--allowedTools", "--permission-mode", "--output-format"} {
		if !bytes.Contains(help, []byte(required)) {
			return fmt.Errorf("installed Claude lacks %s", required)
		}
	}
	if err = e2eWrite(filepath.Join(output, "launch-preflight.json"), map[string]any{"command": r.Command, "version": r.Version, "normal_authentication": true, "history_preserved": true, "removed_tool_not_used": "TaskOutput", "installation_hash": digest(data)}); err != nil {
		return err
	}
	if time.Until(budget.activeDeadline) <= 0 {
		return fmt.Errorf("setup consumed active budget; no model launched")
	}
	// Claim immediately before launching. Even a failed spawn consumes the grant;
	// another invocation cannot turn a failed attempt into an automatic retry.
	if err = claimE2EAttempt(attempt); err != nil {
		return err
	}
	raw, err := os.OpenFile(filepath.Join(output, "stdout.jsonl"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer raw.Close()
	errors, err := os.OpenFile(filepath.Join(output, "stderr.log"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer errors.Close()
	cmd := exec.Command("claude", nativeArgs...)
	cmd.Dir = root
	cmd.Env = childEnv
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stdout = raw
	cmd.Stderr = errors
	report.ModelStarted = true
	report.ExitCode, r.Error = runProcess(cmd, time.Until(budget.activeDeadline))
	r.ExitCode = report.ExitCode
	raw.Close()
	errors.Close()
	trace, err := os.Open(filepath.Join(output, "stdout.jsonl"))
	if err != nil {
		return err
	}
	r.Trace = parseTrace("claude", trace)
	trace.Close()
	r.ModelObserved = r.Trace.Model
	r.ModelAtInit = r.Trace.ModelAtInit
	r.Terminal = terminalState(r.Trace, r.ExitCode, r.Error)
	r.After, err = inventory(root)
	if err != nil {
		return err
	}
	memory.report.RuntimeMemoryToolReadSeen = observedMemoryRead(r.Trace.Events)
	for _, name := range []string{"hive-mode", "flow-plan", "flow-build", "flow-verify", "flow-close", "git-workflow"} {
		if err := e2eWrite(filepath.Join(output, name+"-source-observation.json"), observeSkill(r, name)); err != nil {
			return err
		}
	}
	if err = e2eWrite(filepath.Join(output, "trace.json"), r.Trace); err != nil {
		return err
	}
	// Resolve the declared branch's actual checkout without integrating or
	// changing any checkout on behalf of the coordinator.
	target, integrationHead, err := inspectE2EIntegration(root, childEnv, budget.activeDeadline)
	if err != nil {
		return err
	}
	report.IntegrationWorktree, report.IntegrationHead = target, integrationHead
	if err := verifyE2EImmutableFiles(target, report); err != nil {
		return err
	}
	targetEnv := environmentInDirectory(childEnv, target)
	outputBytes, code, jointErr := e2eCommand(budget.activeDeadline, target, targetEnv, "go", "-C", target, "test", "./...")
	report.JointChecked = true
	report.JointExitCode = code
	report.JointOutput = string(outputBytes)
	if jointErr != nil {
		return jointErr
	}
	report.Reason = "joint local check passed; native behavioral attribution requires evidence review, never coordinator self-report"
	return nil
}

func inspectE2EIntegration(root string, env []string, deadline time.Time) (string, string, error) {
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", "", err
	}
	run := func(dir string, args ...string) (string, error) {
		output, _, err := e2eCommand(deadline, dir, environmentInDirectory(env, dir), "git", append([]string{"-C", dir}, args...)...)
		return strings.TrimSpace(string(output)), err
	}
	listing, err := run(canonicalRoot, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return "", "", err
	}
	target, listedHead := "", ""
	for _, record := range strings.Split(listing, "\x00\x00") {
		path, branch, head := "", "", ""
		for _, field := range strings.Split(record, "\x00") {
			if v, ok := strings.CutPrefix(field, "worktree "); ok {
				path = v
			}
			if v, ok := strings.CutPrefix(field, "branch "); ok {
				branch = v
			}
			if v, ok := strings.CutPrefix(field, "HEAD "); ok {
				head = v
			}
		}
		if branch != "refs/heads/fixture-integration" {
			continue
		}
		if target != "" {
			return "", "", fmt.Errorf("integration branch has ambiguous checkouts")
		}
		target, listedHead = path, head
	}
	if target == "" {
		return "", "", fmt.Errorf("declared local integration target not checked out; verified is not integrated")
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return "", "", err
	}
	if !under(target, canonicalRoot) {
		return "", "", fmt.Errorf("integration checkout escapes fixture")
	}
	common, err := run(canonicalRoot, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", "", err
	}
	targetCommon, err := run(target, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", "", err
	}
	common, err = filepath.EvalSymlinks(common)
	if err != nil {
		return "", "", err
	}
	targetCommon, err = filepath.EvalSymlinks(targetCommon)
	if err != nil {
		return "", "", err
	}
	if common != targetCommon {
		return "", "", fmt.Errorf("integration checkout belongs to a different repository")
	}
	branch, err := run(target, "symbolic-ref", "--short", "HEAD")
	if err != nil || branch != "fixture-integration" {
		return "", "", fmt.Errorf("integration checkout branch changed or is detached")
	}
	head, err := run(target, "rev-parse", "--verify", "HEAD")
	if err != nil || len(head) != 40 || head != listedHead {
		return "", "", fmt.Errorf("integration full head changed or could not be corroborated")
	}
	status, err := run(target, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return "", "", err
	}
	if status != "" {
		return "", "", fmt.Errorf("integration checkout is dirty")
	}
	return target, head, nil
}

func verifyE2EImmutableFiles(root string, report *e2eReport) error {
	for name, body := range e2eFiles() {
		if name == "a/value.go" || name == "b/label.go" {
			continue
		}
		actual, err := os.ReadFile(filepath.Join(root, name))
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("immutable fixture file could not be inspected: %s: %w", name, err)
		}
		if err != nil || string(actual) != body {
			report.RunnerFailure = fmt.Sprintf("immutable fixture file changed or removed: %s", name)
			return fmt.Errorf("%s", report.RunnerFailure)
		}
	}
	return nil
}

// Native behavioral attribution is a rubric review, not a keyword scorer. An
// independent reviewer supplies concrete, hash-bound raw-evidence anchors. The
// driver validates their integrity; the reviewer remains responsible for whether
// the cited native event actually establishes the criterion. This is explicitly
// distinguished from automatic Git/joint-check evidence in the report.
var e2eReviewCriteria = []string{"source-reads", "independent-owner-heads", "independent-exact-head-verification", "a-defect-correction", "b-closed-partial-replacement", "joint-local-integration", "suppression-review", "isolated-memory-routing-cleanup", "installation-restored"}

type e2eEvidenceAnchor struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Line   int    `json:"line"`
}
type e2eReviewedCriterion struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	Explanation string              `json:"explanation"`
	Evidence    []e2eEvidenceAnchor `json:"evidence"`
}
type e2eReview struct {
	ReviewerSessionID string                 `json:"reviewer_session_id"`
	RunHash           string                 `json:"run_sha256"`
	Criteria          []e2eReviewedCriterion `json:"criteria"`
}

func validateE2EReview(output string, r e2eReport, runHash string, review e2eReview) error {
	if review.ReviewerSessionID == "" || review.ReviewerSessionID == r.Launch.Trace.SessionID {
		return fmt.Errorf("independent reviewer session required")
	}
	if review.RunHash != runHash {
		return fmt.Errorf("review does not bind the recorded run")
	}
	seen := map[string]bool{}
	for _, c := range review.Criteria {
		known := false
		for _, id := range e2eReviewCriteria {
			if id == c.ID {
				known = true
			}
		}
		if !known || seen[c.ID] || strings.TrimSpace(c.Explanation) == "" {
			return fmt.Errorf("unknown, duplicate or unexplained criterion %s", c.ID)
		}
		seen[c.ID] = true
		if c.Status != "pass" && c.Status != "failed" && c.Status != "blocked" && c.Status != "inconclusive" {
			return fmt.Errorf("invalid review class")
		}
		if len(c.Evidence) == 0 {
			return fmt.Errorf("raw evidence anchors required: %s", c.ID)
		}
		for _, a := range c.Evidence {
			if filepath.IsAbs(a.Path) || a.Path == "e2e-evidence.json" || a.Path == "reviewed-e2e-evidence.json" || a.Path == "prompt.txt" {
				return fmt.Errorf("raw evidence required, not a model/assessment summary")
			}
			path, err := filepath.EvalSymlinks(filepath.Join(output, a.Path))
			if err != nil {
				return err
			}
			base, err := filepath.EvalSymlinks(output)
			if err != nil {
				return err
			}
			if !under(path, base) || path == base {
				return fmt.Errorf("evidence anchor escapes run")
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if digest(data) != a.SHA256 || a.Line < 1 || a.Line > len(bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n"))) {
				return fmt.Errorf("stale or missing evidence anchor %s", a.Path)
			}
		}
	}
	if len(seen) != len(e2eReviewCriteria) {
		return fmt.Errorf("all behavioral and restoration criteria must be reviewed")
	}
	return nil
}

func e2eReviewMain(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("pilot e2e review", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "existing run directory")
	reviewPath := fs.String("review", "", "independent rubric review JSON with raw evidence anchors")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 || !filepath.IsAbs(*out) || !filepath.IsAbs(*reviewPath) {
		fmt.Fprintln(stderr, "absolute out and review required")
		return 2
	}
	raw, err := os.ReadFile(filepath.Join(*out, "e2e-evidence.json"))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var r e2eReport
	if err = json.Unmarshal(raw, &r); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	reviewBytes, err := os.ReadFile(*reviewPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	var review e2eReview
	dec := json.NewDecoder(bytes.NewReader(reviewBytes))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&review); err == nil {
		err = validateE2EReview(*out, r, digest(raw), review)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	r.Review = &review
	r.Class = classifyE2E(r)
	if r.Class == "pass" {
		r.Unverified = nil
		r.Reason = "independent rubric review with hash-bound raw evidence; local joint check corroborated by runner"
	}
	// Preserve the original observational record, including its unverified fields.
	resultPath := filepath.Join(*out, "reviewed-e2e-evidence.json")
	f, err := os.OpenFile(resultPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	err = json.NewEncoder(f).Encode(r)
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		fmt.Fprintln(stderr, "write reviewed evidence failed")
		return 1
	}
	fmt.Fprintf(stdout, "%s: %s\n", r.Class, resultPath)
	if r.Class != "pass" {
		return 1
	}
	return 0
}
