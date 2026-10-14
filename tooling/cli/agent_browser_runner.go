package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

const (
	// agentBrowserStepTimeout bounds one step; agentBrowserWaitDelay bounds
	// how long a descendant that keeps the output pipe open can extend it.
	agentBrowserStepTimeout = 15 * time.Minute
	agentBrowserWaitDelay   = 2 * time.Second
	// agentBrowserMaxOutput is the most output a step stores. Output beyond it
	// is dropped without stopping the process.
	agentBrowserMaxOutput = 64 << 10
	// agentBrowserDetailLines is how many trailing output lines the result
	// detail shows.
	agentBrowserDetailLines = 5
)

// agentBrowserResult is the stored outcome of one step (ExternalStep result
// in the onboarding journal). A failure, timeout or cancellation is a result,
// never a Go error, so the journal keeps going and never records `unknown`.
type agentBrowserResult struct {
	ExitCode int  `json:"exit_code"`
	TimedOut bool `json:"timed_out,omitempty"`
	// Interrupted marks a step stopped by SIGINT or SIGTERM.
	Interrupted bool   `json:"interrupted,omitempty"`
	Skipped     bool   `json:"skipped,omitempty"`
	Message     string `json:"message,omitempty"`
	Output      string `json:"output,omitempty"`
}

func (r agentBrowserResult) failed() bool { return r.TimedOut || r.Interrupted || r.ExitCode != 0 }

// agentBrowserInterrupted becomes true once a step was stopped by SIGINT or
// SIGTERM; later steps then fail without running anything.
var agentBrowserInterrupted atomic.Bool

// agentBrowserContext returns the context one step runs under: it is cancelled
// by SIGINT or SIGTERM, so Ctrl-C on `hive install` kills the step's process
// group instead of orphaning brew or npm. Tests replace it.
var agentBrowserContext = func() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

// runAgentBrowserCommand is the production agentBrowserExec.
func runAgentBrowserCommand(argv []string) agentBrowserResult {
	ctx, stop := agentBrowserContext()
	defer stop()
	res, _ := runBoundedCommandContext(ctx, argv, agentBrowserStepTimeout, agentBrowserWaitDelay, agentBrowserMaxOutput)
	if res.Interrupted {
		agentBrowserInterrupted.Store(true)
	}
	return res
}

func runBoundedCommand(argv []string, timeout, waitDelay time.Duration, maxOutput int) agentBrowserResult {
	res, _ := runBoundedCommandContext(context.Background(), argv, timeout, waitDelay, maxOutput)
	return res
}

// commandOutcome classifies a finished command. A command that exited
// successfully is a success even when the deadline or a signal arrived in the
// same instant; only then are the deadline, a cancellation and an exit error
// told apart.
func commandOutcome(runErr, ctxErr error) (ok, timedOut, interrupted bool) {
	switch {
	case runErr == nil || errors.Is(runErr, exec.ErrWaitDelay):
		return true, false, false
	case errors.Is(ctxErr, context.DeadlineExceeded):
		return false, true, false
	case errors.Is(ctxErr, context.Canceled):
		return false, false, true
	}
	return false, false, false
}

// runBoundedCommandContext runs argv without a shell, from a private empty
// directory (a shared /tmp could hold a foreign .npmrc that changes what npm
// reports), with an empty stdin, Homebrew's auto-update disabled, in its own
// session. The whole process group is killed at the deadline, on cancellation
// and after every run, so a descendant never outlives its step. stdout and
// stderr are merged into a buffer capped at maxOutput (the last bytes are
// kept; the process is not cancelled); stdout alone is also returned, for
// callers that parse it.
func runBoundedCommandContext(parent context.Context, argv []string, timeout, waitDelay time.Duration, maxOutput int) (agentBrowserResult, string) {
	if len(argv) == 0 {
		return agentBrowserResult{ExitCode: -1, Message: "no command"}, ""
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return agentBrowserResult{ExitCode: -1, Message: argv[0] + " was not found in PATH"}, ""
	}
	dir, err := os.MkdirTemp("", "hive-agent-browser-")
	if err != nil {
		return agentBrowserResult{ExitCode: -1, Message: "cannot create a private working directory: " + err.Error()}, ""
	}
	defer os.RemoveAll(dir)
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
	cmd.Stdin = nil // /dev/null
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = waitDelay
	out := &tailBuffer{limit: maxOutput}
	stdout := &tailBuffer{limit: maxOutput}
	cmd.Stdout, cmd.Stderr = io.MultiWriter(out, stdout), out
	runErr := cmd.Run()
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) // ESRCH when nothing is left
	}
	res := agentBrowserResult{Output: out.String()}
	ok, timedOut, interrupted := commandOutcome(runErr, ctx.Err())
	switch {
	case ok:
	case timedOut:
		res.ExitCode, res.TimedOut = -1, true
		res.Message = "timed out after " + timeout.String()
	case interrupted:
		res.ExitCode, res.Interrupted = -1, true
		res.Message = "interrupted"
	default:
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
		}
		res.Message = fmt.Sprintf("%s failed: %s", argv[0], runErr)
	}
	return res, stdout.String()
}

// tailBuffer keeps the last limit bytes and drops what came before, so the
// error that usually ends a package manager's output is what gets stored. It
// accepts every write, so a long output never cancels the process.
type tailBuffer struct {
	mu    sync.Mutex
	buf   []byte
	limit int
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if len(t.buf) > 2*t.limit {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.limit:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return lastBytes(string(t.buf), t.limit)
}

func lastBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

// providerRunner is the ExternalRunner for every capability: manual steps are
// confirmed back unchanged, and agent-browser's planned steps run through
// agentBrowserExec. A synthetic home never runs a program.
type providerRunner struct {
	manual    manualProviderRunner
	sys       agentBrowserSystem
	synthetic bool
}

func (r providerRunner) Validate(s management.ExternalStep) error {
	return providers.ValidateStep(s.Payload)
}

func decodeStep(s management.ExternalStep) (providers.Step, error) {
	var step providers.Step
	if err := providers.ValidateStep(s.Payload); err != nil {
		return step, err
	}
	return step, json.Unmarshal(s.Payload, &step)
}

func (r providerRunner) Execute(s management.ExternalStep) (json.RawMessage, error) {
	step, err := decodeStep(s)
	if err != nil {
		return nil, err
	}
	if step.Status == providers.Manual {
		return r.manual.Execute(s)
	}
	res := r.run(step)
	res.Output = lastBytes(res.Output, agentBrowserMaxOutput)
	return json.Marshal(res)
}

func (r providerRunner) run(step providers.Step) agentBrowserResult {
	if r.synthetic {
		return agentBrowserResult{Skipped: true, Message: "a synthetic home never runs a program"}
	}
	if agentBrowserInterrupted.Load() {
		return agentBrowserResult{ExitCode: -1, Interrupted: true, Message: "interrupted"}
	}
	if step.ID == providers.AgentBrowserChromeStep {
		if _, err := r.sys.lookPath(agentBrowserBinary); err != nil {
			return agentBrowserResult{Skipped: true, Message: "agent-browser is not in PATH, so its browser cannot be downloaded; the install step did not complete"}
		}
	}
	res := agentBrowserExec(step.Argv)
	switch {
	case res.failed():
		switch {
		case res.Message != "":
		case res.TimedOut:
			res.Message = "timed out after " + agentBrowserStepTimeout.String()
		default:
			res.Message = fmt.Sprintf("%s exited with code %d", step.Argv[0], res.ExitCode)
		}
		if strings.Contains(res.Output, "EACCES") && step.Method == providers.MethodNpm {
			res.Message = "npm cannot write to its global prefix (EACCES). Fix the prefix permissions or install agent-browser by hand; Hive never uses sudo."
		}
	case step.ID == providers.AgentBrowserInstallStep:
		if _, err := r.sys.lookPath(agentBrowserBinary); err != nil {
			res.Message = "agent-browser was installed but is not on PATH; add the directory of its executable (npm's global bin or Homebrew's bin) to PATH"
		}
	case step.ID == providers.AgentBrowserChromeStep:
		if found, _ := r.sys.browserFound(); !found {
			res.Message = "agent-browser install finished but no browser was found"
		}
	}
	return res
}

// Reconcile settles a step from its stored result and the machine's state
// (D-7). A step with no stored result was interrupted: it is failed, never
// unknown, and is never run again here.
func (r providerRunner) Reconcile(s management.ExternalStep, result json.RawMessage) (string, error) {
	step, err := decodeStep(s)
	if err != nil {
		return "", err
	}
	if step.Status == providers.Manual {
		return r.manual.Reconcile(s, result)
	}
	var res agentBrowserResult
	if len(result) == 0 || string(result) == "{}" || json.Unmarshal(result, &res) != nil {
		return management.StepFailed, nil
	}
	switch {
	case res.Skipped:
		return management.StepSkipped, nil
	case res.failed():
		return management.StepFailed, nil
	}
	switch step.ID {
	case providers.AgentBrowserInstallStep:
		if _, err := r.sys.lookPath(agentBrowserBinary); err == nil {
			return management.StepVerified, nil
		}
	case providers.AgentBrowserChromeStep:
		if found, _ := r.sys.browserFound(); found {
			return management.StepVerified, nil
		}
	}
	return management.StepFailed, nil
}

// stepResultDetail returns the failure message and the last output lines of a
// stored step result for the partial-install detail. Every line is sanitized:
// a package manager's output is untrusted terminal text.
func stepResultDetail(step management.StepReceipt) (message string, tail []string) {
	status, result := step.Status, step.Result
	var res agentBrowserResult
	if len(result) == 0 {
		// RecoverOnboarding settles an interrupted step with no saved result.
		if status == management.StepFailed && strings.HasPrefix(step.Step.ID, string(providers.AgentBrowser)+"-") {
			return "interrupted before its result was recorded", nil
		}
		return "", nil
	}
	if json.Unmarshal(result, &res) != nil {
		return "", nil
	}
	message = sanitizeLine(res.Message)
	var lines []string
	for _, l := range strings.Split(res.Output, "\n") {
		if l = strings.TrimSpace(sanitizeLine(l)); l != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) > agentBrowserDetailLines {
		lines = lines[len(lines)-agentBrowserDetailLines:]
	}
	return message, lines
}
