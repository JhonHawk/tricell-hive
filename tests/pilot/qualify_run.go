package main

// Driver of the qualification batch: `qualify run` launches the runner once per
// attempt and `qualify score` prints the scorer's report (design.md «Driver»;
// AC58). Both are recognized in os.Args[1] by main() before any flag parsing.
//
// Exit codes: `qualify run` returns 2 for every refusal before the first
// launch, 1 when a launch or a write failed (the batch can be resumed with the
// same flags) and 0 otherwise. `qualify score` returns 1 when it refuses a
// verdict (AC57), 2 for a usage error and 0 when it prints a report.

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"tricell-hive/integrations/target"
)

const (
	qualifyFloorPath = "_support/docs/harness-engineering/model-floor.md"

	qualifyRunUsage = `usage: pilot qualify run --host <host> --models <a,b> --reference <model> --source <checkout> --out <dir> --configured-model <value> [flags]

  --host <host>               codex, claude, grok, pi or opencode (required)
  --models <a,b>              candidate models, comma separated (required)
  --reference <model>         reference model, run in the deployed arm only (required)
  --source <checkout>         checkout holding the suite and model-floor.md (default ".")
  --out <dir>                 batch directory, below <source>/_support/workspace (required)
  --configured-model <value>  configuration value observed before launch, passed to the runner (required)
  --guidance-source <dir>     checkout installed for arm A; required when A is scheduled (codex, opencode)
  --runs <n>                  repetitions per task and arm (default 9)
  --split <name>              heldback or tuning (default "heldback")
  --effort <model>=<value>    effort of a model; repeatable; needed when model-floor.md «Floor» lists several
                              values or no row for the model; on OpenCode it becomes the #variant
  --provider <name>           native provider, passed to the runner on Pi only
  --confirm-sessions <n>      must equal the sessions this call will launch (the repetitions without a
                              usable attempt); without it the call prints the count and launches nothing

exit codes: 0 done, 1 a launch failed (call again to relaunch it), 2 refused before launching anything
`
	qualifyScoreUsage = `usage: pilot qualify score --out <dir> --reference <model> [--source <checkout>]

  --out <dir>           batch directory written by qualify run (required)
  --reference <model>   reference model; must equal the one in batch.json
  --source <checkout>   checkout holding the suite and model-floor.md (default ".")

exit codes: 0 report printed, 1 no verdict (batch.json, reference, repetitions or cells refuse it), 2 usage
`
	qualifyUsage = `usage: pilot qualify run|score [flags]

  run    launch the qualification runs of a batch
  score  print the report and verdict of a batch
`
)

// qualifyRunnerCommand builds the runner invocation for one attempt: the
// driver's own executable. Tests replace it.
var qualifyRunnerCommand = func(args []string) *exec.Cmd {
	exe, err := os.Executable()
	if err != nil {
		exe = os.Args[0]
	}
	return exec.Command(exe, args...)
}

// qualifySuiteLoader loads the suite of a checkout; the subcommands call
// loadQualificationSuite, tests pass a smaller suite.
type qualifySuiteLoader func(source string) (qualifySuite, error)

// qualifyMain dispatches `qualify <run|score>`.
func qualifyMain(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, qualifyUsage)
		return 2
	}
	switch args[0] {
	case "run":
		return qualifyRunCmd(args[1:], loadQualificationSuite, stdout, stderr)
	case "score":
		return qualifyScoreCmd(args[1:], loadQualificationSuite, stdout, stderr)
	}
	fmt.Fprintf(stderr, "unknown qualify subcommand %q\n%s", args[0], qualifyUsage)
	return 2
}

// qualifyEffortFlags collects the repeatable --effort <model>=<value>.
type qualifyEffortFlags map[string]string

func (f qualifyEffortFlags) String() string { return "" }

func (f qualifyEffortFlags) Set(v string) error {
	model, value, ok := strings.Cut(v, "=")
	model, value = qualifyBaseModel(strings.TrimSpace(model)), strings.TrimSpace(value)
	if !ok || model == "" || value == "" {
		return fmt.Errorf("--effort %q must be <model>=<value>", v)
	}
	if _, dup := f[model]; dup {
		return fmt.Errorf("--effort given twice for %s", model)
	}
	f[model] = value
	return nil
}

// qualifyFloorRows reads the «Floor» table of model-floor.md: the efforts of
// each (host, model) row. A row with «—» is present with no value.
func qualifyFloorRows(doc string) (map[string][]string, error) {
	const head = "\n## Floor\n"
	i := strings.Index(doc, head)
	if i < 0 {
		return nil, errors.New("model-floor.md has no «## Floor» section")
	}
	rest := doc[i+len(head):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	rows := map[string][]string{}
	header := true
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if header {
			header = false
			continue
		}
		if strings.HasPrefix(line, "| ---") || strings.HasPrefix(line, "|---") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 3 {
			continue
		}
		model := qualifyBacktick.FindStringSubmatch(cells[1])
		if model == nil {
			return nil, fmt.Errorf("«Floor» row %q has no backticked model", line)
		}
		var values []string
		for _, m := range qualifyBacktick.FindAllStringSubmatch(cells[2], -1) {
			values = append(values, m[1])
		}
		rows[strings.TrimSpace(cells[0])+"\x00"+model[1]] = values
	}
	if len(rows) == 0 {
		return nil, errors.New("«## Floor» holds no rows")
	}
	return rows, nil
}

// resolveQualifyEffort returns the effort of one model and the model string the
// runner receives. A --effort override wins; otherwise the «Floor» row decides:
// one value is used, «—» passes none, several values or no row refuse. On
// OpenCode the value is the #variant of the model and a suffix that
// contradicts it refuses.
func resolveQualifyEffort(host, model string, overrides map[string]string, rows map[string][]string) (effort, runnerModel string, err error) {
	base := qualifyBaseModel(model)
	if v, ok := overrides[base]; ok {
		effort = v
	} else {
		values, ok := rows[host+"\x00"+base]
		switch {
		case !ok:
			return "", "", fmt.Errorf("model %s has no row for host %s in model-floor.md «Floor»: pass --effort %s=<value>", base, host, base)
		case len(values) > 1:
			return "", "", fmt.Errorf("model %s lists several efforts (%s) in model-floor.md «Floor»: pass --effort %s=<value>", base, strings.Join(values, ", "), base)
		case len(values) == 1:
			effort = values[0]
		}
	}
	if host != "opencode" {
		return effort, model, nil
	}
	if suffix := qualifyVariant(model); suffix != "" && suffix != effort {
		return "", "", fmt.Errorf("model %s asks variant %q but its effort is %q", model, suffix, effort)
	}
	if effort == "" {
		return effort, base, nil
	}
	return effort, base + "#" + effort, nil
}

// qualifyExpectedAHash is the guidance hash an A attempt records: the checkout
// installed by the runner's own install path into a temporary home, read with
// guidanceHash. Nothing outside the temporary home is touched.
func qualifyExpectedAHash(host, source string) (string, error) {
	tmp, err := os.MkdirTemp("", "hive-qualify-a-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)
	home, err := target.Canonical(tmp)
	if err != nil {
		return "", err
	}
	if _, err := installGuidanceVariant(source, home, host); err != nil {
		return "", fmt.Errorf("installing %s for the A hash: %w", source, err)
	}
	return guidanceHash(host, home)
}

// qualifyJob is one session of the calendar: one repetition of a cell.
type qualifyJob struct {
	Model, Arm string
	Task       qualifyTask
	Rep        int
	Attempt    int // number of the attempt this call launches
}

func (j qualifyJob) dir(out, host string) string {
	return qualifyAttemptDir(out, host, j.Model, j.Arm, j.Task.ID, strconv.Itoa(j.Rep), j.Attempt)
}

// qualifyCalendar lists the repetitions of the batch in launch order:
// repetition, task, model (the reference first), with the arms of each cell
// rotated by repetition as flashSchedule does. The reference and the arms A
// and none appear as AC58 says.
func qualifyCalendar(host, reference string, candidates []string, tasks []qualifyTask, runs int) []qualifyJob {
	hasNone := qualifyNoneArmHosts[host]
	models := append([]string{reference}, candidates...)
	var jobs []qualifyJob
	for rep := 1; rep <= runs; rep++ {
		for ti, t := range tasks {
			for mi, m := range models {
				arms := []string{qualifyArmDeployed}
				if hasNone && t.CompareNone && m != reference {
					arms = append(arms, "A", qualifyArmNone)
				}
				for k := range arms {
					jobs = append(jobs, qualifyJob{Model: m, Arm: arms[(rep-1+ti+mi+k)%len(arms)], Task: t, Rep: rep})
				}
			}
		}
	}
	return jobs
}

// qualifyPending keeps the jobs whose repetition has no usable attempt and
// numbers the attempt each one launches (one after the highest that exists).
func qualifyPending(out string, batch qualifyBatch, floor qualifyFloor, jobs []qualifyJob) ([]qualifyJob, error) {
	var pending []qualifyJob
	for _, j := range jobs {
		repDir := filepath.Dir(j.dir(out, batch.Host))
		nums, err := qualifyAttemptNumbers(repDir)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		usable := false
		for _, n := range nums {
			var run result
			data, err := os.ReadFile(filepath.Join(repDir, "attempt-"+strconv.Itoa(n), "run.json"))
			if err != nil || json.Unmarshal(data, &run) != nil {
				continue
			}
			if ok, _ := usableAttempt(run, batch, floor); ok {
				usable = true
				break
			}
		}
		if usable {
			continue
		}
		j.Attempt = 1
		if len(nums) > 0 {
			j.Attempt = nums[0] + 1
		}
		pending = append(pending, j)
	}
	return pending, nil
}

// qualifyBatchDiff names the first field of the stored batch that differs from
// the current one, with both values (AC58).
func qualifyBatchDiff(old, cur qualifyBatch) string {
	norm := func(m map[string]string) map[string]string {
		if m == nil {
			return map[string]string{}
		}
		return m
	}
	split := func(s string) string {
		if s == "" {
			return qualifySplitHeldback
		}
		return s
	}
	switch {
	case old.Host != cur.Host:
		return fmt.Sprintf("host: batch.json has %q, this call has %q", old.Host, cur.Host)
	case split(old.Split) != split(cur.Split):
		return fmt.Sprintf("split: batch.json has %q, this call has %q", split(old.Split), split(cur.Split))
	case !reflect.DeepEqual(old.Models, cur.Models):
		return fmt.Sprintf("models: batch.json has %v, this call has %v", old.Models, cur.Models)
	case old.Reference != cur.Reference:
		return fmt.Sprintf("reference: batch.json has %q, this call has %q", old.Reference, cur.Reference)
	case !reflect.DeepEqual(norm(old.Effort), norm(cur.Effort)):
		return fmt.Sprintf("effort: batch.json has %v, this call has %v", norm(old.Effort), norm(cur.Effort))
	case old.DeployedHash != cur.DeployedHash:
		return fmt.Sprintf("deployed guidance hash: batch.json has %s, now %s", old.DeployedHash, cur.DeployedHash)
	case old.AHash != cur.AHash:
		return fmt.Sprintf("A guidance hash: batch.json has %s, now %s", old.AHash, cur.AHash)
	}
	return ""
}

func qualifyRunCmd(args []string, load qualifySuiteLoader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("qualify run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, qualifyRunUsage) }
	host := fs.String("host", "", "")
	modelsFlag := fs.String("models", "", "")
	reference := fs.String("reference", "", "")
	source := fs.String("source", ".", "")
	outFlag := fs.String("out", "", "")
	configured := fs.String("configured-model", "", "")
	guidanceSource := fs.String("guidance-source", "", "")
	runs := fs.Int("runs", 9, "")
	split := fs.String("split", qualifySplitHeldback, "")
	provider := fs.String("provider", "", "")
	confirm := fs.Int("confirm-sessions", -1, "")
	overrides := qualifyEffortFlags{}
	fs.Var(overrides, "effort", "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	refuse := func(format string, a ...any) int {
		fmt.Fprintf(stderr, "qualify run: "+format+"\n", a...)
		return 2
	}
	if fs.NArg() > 0 {
		return refuse("unexpected argument %q", fs.Arg(0))
	}
	switch {
	case !validHost(*host):
		return refuse("--host must be codex, claude, grok, pi or opencode")
	case *modelsFlag == "":
		return refuse("--models is required")
	case *reference == "":
		return refuse("--reference is required")
	case *outFlag == "":
		return refuse("--out is required")
	case *configured == "":
		return refuse("--configured-model is required")
	case *runs < 1:
		return refuse("--runs must be at least 1")
	case *split != qualifySplitHeldback && *split != qualifySplitTuning:
		return refuse("--split must be %s or %s", qualifySplitHeldback, qualifySplitTuning)
	case *provider != "" && *host != "pi":
		return refuse("--provider is only for --host pi")
	}
	src, err := filepath.Abs(*source)
	if err != nil {
		return refuse("%v", err)
	}
	workspace, err := target.Canonical(filepath.Join(src, "_support/workspace"))
	if err != nil {
		return refuse("%v", err)
	}
	out, err := target.Canonical(*outFlag)
	if err != nil {
		return refuse("%v", err)
	}
	if !under(out, workspace) || out == workspace {
		return refuse("--out %s must be below %s", out, workspace)
	}
	suite, err := load(src)
	if err != nil {
		return refuse("suite: %v", err)
	}
	floorDoc, err := os.ReadFile(filepath.Join(src, qualifyFloorPath))
	if err != nil {
		return refuse("%v", err)
	}
	floor, err := parseQualifyFloor(string(floorDoc))
	if err != nil {
		return refuse("%v", err)
	}
	rows, err := qualifyFloorRows(string(floorDoc))
	if err != nil {
		return refuse("%v", err)
	}

	// Models and efforts.
	var requested []string
	seen := map[string]bool{}
	for _, m := range strings.Split(*modelsFlag, ",") {
		m = strings.TrimSpace(m)
		if m == "" || seen[qualifyBaseModel(m)] {
			return refuse("--models holds an empty or repeated model")
		}
		seen[qualifyBaseModel(m)] = true
		requested = append(requested, m)
	}
	batch := qualifyBatch{Host: *host, Split: *split, Effort: map[string]string{}}
	known := map[string]bool{}
	var candidates []string
	resolve := func(m string) (string, error) {
		effort, runnerModel, err := resolveQualifyEffort(*host, m, overrides, rows)
		if err != nil {
			return "", err
		}
		if effort != "" {
			batch.Effort[qualifyBaseModel(m)] = effort
		}
		known[qualifyBaseModel(m)] = true
		return runnerModel, nil
	}
	if batch.Reference, err = resolve(*reference); err != nil {
		return refuse("%v", err)
	}
	for _, m := range requested {
		rm, err := resolve(m)
		if err != nil {
			return refuse("%v", err)
		}
		batch.Models = append(batch.Models, rm)
		if qualifyBaseModel(m) != qualifyBaseModel(*reference) {
			candidates = append(candidates, rm)
		}
	}
	for m := range overrides {
		if !known[m] {
			return refuse("--effort names %s, which is neither in --models nor the --reference", m)
		}
	}

	// Calendar and hashes.
	var tasks []qualifyTask
	needA := false
	for _, t := range suite.Tasks {
		if t.Split == *split {
			tasks = append(tasks, t)
			needA = needA || t.CompareNone
		}
	}
	if len(tasks) == 0 {
		return refuse("the suite has no %s task", *split)
	}
	needA = needA && qualifyNoneArmHosts[*host] && len(candidates) > 0
	if batch.DeployedHash, err = guidanceHash(*host, ""); err != nil {
		return refuse("deployed guidance hash: %v", err)
	}
	if needA {
		if *guidanceSource == "" {
			return refuse("--guidance-source is required: arm A is scheduled")
		}
		gs, err := filepath.Abs(*guidanceSource)
		if err != nil {
			return refuse("%v", err)
		}
		*guidanceSource = gs
		if batch.AHash, err = qualifyExpectedAHash(*host, gs); err != nil {
			return refuse("%v", err)
		}
	}
	stored, readErr := readQualifyBatch(out)
	switch {
	case readErr == nil:
		if d := qualifyBatchDiff(stored, batch); d != "" {
			return refuse("%s; a different checkout, redeploy or configuration needs a new --out", d)
		}
	case !errors.Is(readErr, os.ErrNotExist):
		return refuse("%v", readErr)
	}
	jobs := qualifyCalendar(*host, batch.Reference, candidates, tasks, *runs)
	pending, err := qualifyPending(out, batch, floor, jobs)
	if err != nil {
		return refuse("%v", err)
	}
	n := len(pending)
	breakdown := map[string]int{}
	for _, j := range pending {
		breakdown[j.Model+" / "+j.Arm]++
	}
	keys := make([]string, 0, len(breakdown))
	for k := range breakdown {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Fprintf(stdout, "this call will launch %d sessions (%d of %d repetitions have no usable attempt)\n", n, n, len(jobs))
	for _, k := range keys {
		fmt.Fprintf(stdout, "  %-60s %d\n", k, breakdown[k])
	}
	if *confirm != n {
		if *confirm < 0 {
			return refuse("nothing launched: pass --confirm-sessions %d to launch these sessions", n)
		}
		return refuse("nothing launched: --confirm-sessions %d differs from the %d sessions this call would launch", *confirm, n)
	}

	// Launch.
	if err := os.MkdirAll(out, 0o700); err != nil {
		fmt.Fprintln(stderr, "qualify run:", err)
		return 1
	}
	if readErr != nil {
		if err := writeQualifyBatch(out, batch); err != nil {
			fmt.Fprintln(stderr, "qualify run:", err)
			return 1
		}
	}
	var interrupted atomic.Bool
	var current atomic.Pointer[os.Process]
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sig)
	go func() {
		for s := range sig {
			interrupted.Store(true)
			if p := current.Load(); p != nil {
				_ = p.Signal(s)
			}
		}
	}()
	failed := 0
	for i, j := range pending {
		if interrupted.Load() {
			fmt.Fprintf(stderr, "qualify run: interrupted before session %d of %d; call again with the same flags\n", i+1, n)
			return 1
		}
		dir := j.dir(out, *host)
		if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
			fmt.Fprintln(stderr, "qualify run:", err)
			return 1
		}
		runnerArgs := []string{"--suite", qualificationSuite, "--task", j.Task.ID, "--arm", j.Arm, "--host", *host, "--model", j.Model,
			"--configured-model", *configured, "--delivery", "deployed-global", "--source", src, "--out", dir}
		if j.Arm == "A" {
			runnerArgs = append(runnerArgs, "--guidance-source", *guidanceSource)
		}
		if *host == "pi" && *provider != "" {
			runnerArgs = append(runnerArgs, "--provider", *provider)
		}
		if e := batch.Effort[qualifyBaseModel(j.Model)]; e != "" && *host != "opencode" {
			runnerArgs = append(runnerArgs, "--effort", e)
		}
		cmd := qualifyRunnerCommand(runnerArgs)
		var combined bytes.Buffer
		cmd.Stdout, cmd.Stderr = &combined, &combined
		code := 0
		if err := cmd.Start(); err != nil {
			code = -1
			fmt.Fprintf(&combined, "%v\n", err)
		} else {
			current.Store(cmd.Process)
			if interrupted.Load() { // a signal arrived before the process was stored
				_ = cmd.Process.Signal(syscall.SIGTERM)
			}
			if err := cmd.Wait(); err != nil {
				var ee *exec.ExitError
				if errors.As(err, &ee) {
					code = ee.ExitCode()
				} else {
					code = -1
				}
			}
			current.Store(nil)
		}
		status := "ok"
		if code != 0 && code != 3 { // 3: files changed outside the fixture, run.json is complete
			status = "failed"
			failed++
		}
		fmt.Fprintf(stdout, "[%d/%d] %s / %s / %s rep %d attempt %d: runner exit %d (%s)\n", i+1, n, j.Model, j.Arm, j.Task.ID, j.Rep, j.Attempt, code, status)
		if status == "failed" {
			lines := strings.Split(strings.TrimSpace(combined.String()), "\n")
			fmt.Fprintf(stderr, "  %s\n", lines[len(lines)-1])
		}
	}
	fmt.Fprintf(stdout, "launched %d sessions; %d runner failures\n", n, failed)
	if failed > 0 {
		fmt.Fprintln(stderr, "qualify run: call again with the same flags to relaunch the repetitions without a usable attempt")
		return 1
	}
	return 0
}

func qualifyScoreCmd(args []string, load qualifySuiteLoader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("qualify score", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() { fmt.Fprint(stderr, qualifyScoreUsage) }
	source := fs.String("source", ".", "")
	outFlag := fs.String("out", "", "")
	reference := fs.String("reference", "", "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 0 || *outFlag == "" {
		fs.Usage()
		return 2
	}
	fail := func(err error) int {
		fmt.Fprintln(stderr, "qualify score:", err)
		return 1
	}
	suite, err := load(*source)
	if err != nil {
		return fail(err)
	}
	cases, err := loadFlowCaseIndex(*source)
	if err != nil {
		return fail(err)
	}
	floorDoc, err := os.ReadFile(filepath.Join(*source, qualifyFloorPath))
	if err != nil {
		return fail(err)
	}
	rep, err := scoreQualification(*outFlag, suite, cases, string(floorDoc), *reference)
	if err != nil {
		return fail(err)
	}
	printQualifyReport(stdout, rep)
	return 0
}
