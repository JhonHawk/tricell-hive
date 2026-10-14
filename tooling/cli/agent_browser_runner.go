package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out,omitempty"`
	Skipped  bool   `json:"skipped,omitempty"`
	Message  string `json:"message,omitempty"`
	Output   string `json:"output,omitempty"`
}

func (r agentBrowserResult) failed() bool { return r.TimedOut || r.ExitCode != 0 }

// runAgentBrowserCommand is the production agentBrowserExec.
func runAgentBrowserCommand(argv []string) agentBrowserResult {
	return runBoundedCommand(argv, agentBrowserStepTimeout, agentBrowserWaitDelay, agentBrowserMaxOutput)
}

// runBoundedCommand runs argv without a shell, from the temporary directory,
// with an empty stdin, Homebrew's auto-update disabled, in its own session so
// the whole process group is killed at the deadline, and with stdout and
// stderr merged into a buffer capped at maxOutput (the excess is dropped, the
// process is not cancelled).
func runBoundedCommand(argv []string, timeout, waitDelay time.Duration, maxOutput int) agentBrowserResult {
	if len(argv) == 0 {
		return agentBrowserResult{ExitCode: -1, Message: "no command"}
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return agentBrowserResult{ExitCode: -1, Message: argv[0] + " was not found in PATH"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, argv[1:]...)
	cmd.Dir = os.TempDir()
	cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
	cmd.Stdin = nil // /dev/null
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = waitDelay
	out := &tailBuffer{limit: maxOutput}
	cmd.Stdout, cmd.Stderr = out, out
	runErr := cmd.Run()
	res := agentBrowserResult{Output: out.String()}
	switch {
	case ctx.Err() == context.DeadlineExceeded:
		res.ExitCode, res.TimedOut = -1, true
		res.Message = "timed out after " + timeout.String()
	case runErr == nil || errors.Is(runErr, exec.ErrWaitDelay):
	default:
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = -1
		}
		res.Message = fmt.Sprintf("%s failed: %s", argv[0], runErr)
	}
	return res
}

// tailBuffer keeps the last limit bytes and drops what came before, so the
// error that usually ends a package manager's output is what gets stored. It
// accepts every write, so a long output never cancels the process.
type tailBuffer struct {
	buf   []byte
	limit int
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.buf = append(t.buf, p...)
	if len(t.buf) > 2*t.limit {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.limit:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) String() string { return lastBytes(string(t.buf), t.limit) }

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
