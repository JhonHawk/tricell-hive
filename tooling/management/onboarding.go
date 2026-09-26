package management

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
)

type ExternalStep struct {
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

// ExternalRunner drives one optional onboarding capability. Nothing calls
// Revert; a runner may still implement it during its owner's own migration.
type ExternalRunner interface {
	Validate(ExternalStep) error
	Execute(ExternalStep) (json.RawMessage, error)
	Reconcile(ExternalStep, json.RawMessage) (string, error)
}

// Step statuses, used by StepReceipt.Status and validated by validStepStatus.
const (
	StepPending     = "pending"      // queued, not yet validated or run.
	StepRunning     = "running"      // Execute is in flight; ends verified/failed/unknown.
	StepUnknown     = "unknown"      // an interrupted or unreconciled outcome; blocks new operations.
	StepVerified    = "verified"     // Reconcile confirmed the capability is installed.
	StepFailed      = "failed"       // Validate or Reconcile reported a known, final failure.
	StepSkipped     = "skipped"      // never started, or recovery gave up on it without running it.
	StepAuthPending = "auth_pending" // waiting on the user to complete a provider's own auth flow.
	StepManual      = "manual"       // needs a person, not this runner, to finish or reconcile it.
)

var stepStatuses = map[string]bool{
	StepPending: true, StepRunning: true, StepUnknown: true, StepVerified: true,
	StepFailed: true, StepSkipped: true, StepAuthPending: true, StepManual: true,
}

// terminalStepStatuses are the outcomes Reconcile (or its recovery replay)
// may settle a step on: every status except the transient pending/running
// and the unresolved unknown, which must never be "reconciled into".
var terminalStepStatuses = map[string]bool{
	StepVerified: true, StepFailed: true, StepSkipped: true, StepAuthPending: true, StepManual: true,
}

func validStepStatus(s string) bool    { return stepStatuses[s] }
func terminalStepStatus(s string) bool { return terminalStepStatuses[s] }

type StepReceipt struct {
	Step   ExternalStep    `json:"step"`
	Status string          `json:"status"`
	Result json.RawMessage `json:"result,omitempty"`
}

// PendingKind reports which of the two independent journals blocks a new
// core or onboarding operation, if either does.
type PendingKind int

const (
	PendingNone PendingKind = iota
	PendingCore
	PendingOnboarding
)

// Pending reports whether stateDir has a pending core transaction or a
// pending onboarding journal, so callers never need to read pending.json
// by name to answer that question. It is read-only and does not create
// the state directory.
func Pending(stateDir string) (PendingKind, error) {
	onboarding, err := OnboardingPending(stateDir)
	if err != nil {
		return PendingNone, err
	}
	if onboarding {
		return PendingOnboarding, nil
	}
	s, err := read(filepath.Join(stateDir, "pending.json"))
	if err != nil {
		return PendingNone, err
	}
	if s.Exists {
		return PendingCore, nil
	}
	return PendingNone, nil
}

type OnboardingResult struct {
	ID    string        `json:"id"`
	Phase string        `json:"phase"`
	Steps []StepReceipt `json:"steps,omitempty"`
}
type onboardingJournal struct {
	Version   int                             `json:"version"`
	Installer *distribution.RetainedInstaller `json:"installer,omitempty"`
	OnboardingResult
	StateDir  string `json:"state_dir"`
	CoreID    string `json:"core_id"`
	Integrity string `json:"integrity"`
}

func onboardingPath(dir string) string          { return filepath.Join(dir, "onboarding-pending.json") }
func onboardingHash(j onboardingJournal) string { j.Integrity = ""; return hash(encode(j)) }
func saveOnboarding(j onboardingJournal) error {
	j.Integrity = onboardingHash(j)
	return writeJSON(onboardingPath(j.StateDir), j)
}
func checkOnboarding(dir string) error {
	s, err := read(onboardingPath(dir))
	if err != nil {
		return err
	}
	if s.Exists {
		return fmt.Errorf("unfinished onboarding: recover optional steps first")
	}
	return nil
}

// OnboardingPending is read-only and does not create the state directory.
func OnboardingPending(dir string) (bool, error) {
	s, err := read(onboardingPath(dir))
	return s.Exists, err
}
func loadOnboarding(dir string) (onboardingJournal, error) {
	var j onboardingJournal
	if err := decodeFile(onboardingPath(dir), &j); err != nil {
		return j, err
	}
	if j.Version != 1 || j.StateDir != dir || !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(j.ID) || j.CoreID != j.ID || j.Integrity != onboardingHash(j) {
		return j, fmt.Errorf("invalid onboarding journal")
	}
	switch j.Phase {
	case "prepared", "core_pending", "core_committed", "providers_pending", "completed", "partial":
	default:
		return j, fmt.Errorf("invalid onboarding phase")
	}
	seen := map[string]bool{}
	for _, s := range j.Steps {
		if seen[s.Step.ID] || !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`).MatchString(s.Step.ID) || !json.Valid(s.Step.Payload) {
			return j, fmt.Errorf("invalid onboarding step")
		}
		seen[s.Step.ID] = true
		if !validStepStatus(s.Status) {
			return j, fmt.Errorf("invalid provider status")
		}
	}
	return j, nil
}
func finishOnboarding(j onboardingJournal) error {
	if err := saveOnboarding(j); err != nil {
		return err
	}
	dir := filepath.Join(j.StateDir, "onboarding")
	if err := target.Safe(dir); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	j.Integrity = onboardingHash(j)
	if err := writeJSON(filepath.Join(dir, j.ID+".json"), j); err != nil {
		return err
	}
	return os.Remove(onboardingPath(j.StateDir))
}

// discardUnstartedOnboarding removes the parent journal when the core failed
// before writing its own journal (for example a stale target), so nothing needs
// recovery and a fresh plan can proceed. Any core trace keeps the parent.
func discardUnstartedOnboarding(j onboardingJournal, cause error) error {
	child, err := read(filepath.Join(j.StateDir, "transactions", j.CoreID+".json"))
	if err != nil {
		return cause
	}
	pending, err := read(filepath.Join(j.StateDir, "pending.json"))
	if err != nil || child.Exists || pending.Exists {
		return cause
	}
	if err := os.Remove(onboardingPath(j.StateDir)); err != nil {
		return fmt.Errorf("%w; onboarding journal retained: %v", cause, err)
	}
	return cause
}
func validateExternal(steps []ExternalStep, r ExternalRunner) error {
	if len(steps) > 0 && r == nil {
		return fmt.Errorf("optional steps require a runner")
	}
	seen := map[string]bool{}
	for _, s := range steps {
		if seen[s.ID] || !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`).MatchString(s.ID) || len(s.Payload) > 1024*1024 || !json.Valid(s.Payload) {
			return fmt.Errorf("invalid optional step")
		}
		seen[s.ID] = true
		if err := r.Validate(s); err != nil {
			return err
		}
	}
	return nil
}

// beginOnboarding writes the parent onboarding journal (prepared, then
// core_pending) before the core transaction starts, so a crash before the
// core commits still leaves a journal RecoverOnboarding can inspect.
func beginOnboarding(p Plan, steps []ExternalStep) (onboardingJournal, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return onboardingJournal{}, err
	}
	j := onboardingJournal{Version: 1, Installer: p.Installer, StateDir: p.StateDir, OnboardingResult: OnboardingResult{ID: hex.EncodeToString(id), Phase: "prepared"}}
	j.CoreID = j.ID
	for _, s := range steps {
		j.Steps = append(j.Steps, StepReceipt{Step: s, Status: StepPending})
	}
	if err := saveOnboarding(j); err != nil {
		return j, err
	}
	j.Phase = "core_pending"
	return j, saveOnboarding(j)
}

// runCore applies p's own bytes as a core transaction nested under
// Onboard's lock. state is what Onboard's own preflight already validated
// under that same lock, so the nested Apply trusts it instead of
// re-validating (see Engine.nested). It then advances the parent journal
// to core_committed once the core has durably committed.
func (e Engine) runCore(p Plan, j onboardingJournal, state State) (onboardingJournal, error) {
	core := e
	core.nested = true
	core.presetState = state
	core.transactionID = j.CoreID
	if _, err := core.Apply(p); err != nil {
		return j, discardUnstartedOnboarding(j, err)
	}
	// A crash here leaves a committed child under a core_pending parent;
	// RecoverOnboarding must advance the parent without reverting the core.
	if err := e.fail("core_applied"); err != nil {
		return j, err
	}
	// Even an unchanged core gets an explicit receipt before optional writes.
	j.Phase = "core_committed"
	if err := saveOnboarding(j); err != nil {
		return j, err
	}
	return j, e.fail("core_committed")
}

// runProviders validates, executes, and reconciles each optional step in
// order, keeping the core installed however far it gets. It reports
// whether any step ended short of verified/skipped.
func (e Engine) runProviders(j onboardingJournal, runner ExternalRunner) (onboardingJournal, bool, error) {
	j.Phase = "providers_pending"
	if err := saveOnboarding(j); err != nil {
		return j, false, err
	}
	partial := false
	for i := range j.Steps {
		if err := runner.Validate(j.Steps[i].Step); err != nil {
			j.Steps[i].Status = StepFailed
			partial = true
			if err := saveOnboarding(j); err != nil {
				return j, false, err
			}
			continue
		}
		j.Steps[i].Status = StepRunning
		if err := saveOnboarding(j); err != nil {
			return j, false, err
		}
		result, runErr := runner.Execute(j.Steps[i].Step)
		if err := e.fail("provider_result"); err != nil {
			return j, false, err
		}
		if len(result) > 1024*1024 || (len(result) > 0 && !json.Valid(result)) {
			return j, false, fmt.Errorf("invalid optional step result; reconcile before retry")
		}
		j.Steps[i].Result = result
		if runErr != nil {
			j.Steps[i].Status = StepUnknown
			if err := saveOnboarding(j); err != nil {
				return j, false, err
			}
			return j, false, fmt.Errorf("optional step %s failed; core retained; reconcile before retry", j.Steps[i].Step.ID)
		}
		status, inspectErr := runner.Reconcile(j.Steps[i].Step, result)
		if inspectErr != nil || status == StepUnknown {
			j.Steps[i].Status = StepUnknown
			if err := saveOnboarding(j); err != nil {
				return j, false, err
			}
			return j, false, fmt.Errorf("optional step outcome unknown; recover first")
		}
		if !terminalStepStatus(status) {
			return j, false, fmt.Errorf("invalid optional step outcome")
		}
		j.Steps[i].Status = status
		if status != StepVerified && status != StepSkipped {
			partial = true
		}
		if err := saveOnboarding(j); err != nil {
			return j, false, err
		}
	}
	return j, partial, nil
}

func (e Engine) Onboard(p Plan, steps []ExternalStep, runner ExternalRunner) (OnboardingResult, error) {
	if len(steps) == 0 {
		id, err := e.Apply(p)
		return OnboardingResult{ID: id, Phase: "completed"}, err
	}
	if err := validateExternal(steps, runner); err != nil {
		return OnboardingResult{}, err
	}
	state, unlock, err := preflight(p)
	if err != nil {
		return OnboardingResult{}, err
	}
	defer unlock()
	j, err := beginOnboarding(p, steps)
	if err != nil {
		return j.OnboardingResult, err
	}
	if j, err = e.runCore(p, j, state); err != nil {
		return j.OnboardingResult, err
	}
	partial := false
	if j, partial, err = e.runProviders(j, runner); err != nil {
		return j.OnboardingResult, err
	}
	j.Phase = "completed"
	if partial {
		j.Phase = "partial"
	}
	if err = finishOnboarding(j); err != nil {
		return j.OnboardingResult, err
	}
	if partial {
		return j.OnboardingResult, fmt.Errorf("core installed; optional capabilities incomplete")
	}
	return j.OnboardingResult, nil
}

// RecoverOnboarding never launches installers or authentication. It inspects
// uncertain steps and leaves unreconciled outcomes pending for manual action.
func (e Engine) RecoverOnboarding(stateDir string, runner ExternalRunner) (OnboardingResult, error) {
	dir, err := target.Canonical(stateDir)
	if err != nil {
		return OnboardingResult{}, err
	}
	if _, err = os.Stat(onboardingPath(dir)); os.IsNotExist(err) {
		id, err := e.Recover(dir)
		return OnboardingResult{ID: id, Phase: "completed"}, err
	}
	unlock, err := lock(dir)
	if err != nil {
		return OnboardingResult{}, err
	}
	defer unlock()
	j, err := loadOnboarding(dir)
	if err != nil {
		return OnboardingResult{}, err
	}
	if j.Phase == "prepared" || j.Phase == "core_pending" {
		var child journal
		err = decodeFile(filepath.Join(dir, "transactions", j.CoreID+".json"), &child)
		if err != nil && !os.IsNotExist(err) {
			return j.OnboardingResult, err
		}
		if err == nil && (child.ID != j.CoreID || child.Plan.StateDir != dir || child.Integrity != journalHash(child)) {
			return j.OnboardingResult, fmt.Errorf("invalid child transaction")
		}
		committed := err == nil && child.Integrity == journalHash(child) && child.Phase == "committed"
		core := e
		core.nested = true
		if _, err = core.Recover(dir); err != nil {
			return j.OnboardingResult, err
		}
		if !committed {
			for i := range j.Steps {
				j.Steps[i].Status = StepSkipped
			}
			j.Phase = "partial"
			return j.OnboardingResult, finishOnboarding(j)
		}
		j.Phase = "core_committed"
		if err = saveOnboarding(j); err != nil {
			return j.OnboardingResult, err
		}
	}
	for i := range j.Steps {
		step := &j.Steps[i]
		if step.Status == StepPending {
			step.Status = StepSkipped
			continue
		}
		if step.Status != StepRunning && step.Status != StepUnknown {
			continue
		}
		if runner == nil {
			return j.OnboardingResult, fmt.Errorf("provider reconciliation unavailable")
		}
		status, err := runner.Reconcile(step.Step, step.Result)
		if err != nil || status == StepUnknown {
			step.Status = StepUnknown
			if err := saveOnboarding(j); err != nil {
				return j.OnboardingResult, err
			}
			return j.OnboardingResult, fmt.Errorf("manual reconciliation required for %s", step.Step.ID)
		}
		if !terminalStepStatus(status) {
			return j.OnboardingResult, fmt.Errorf("invalid recovery outcome")
		}
		step.Status = status
	}
	j.Phase = "completed"
	for _, s := range j.Steps {
		if s.Status != StepVerified {
			j.Phase = "partial"
		}
	}
	return j.OnboardingResult, finishOnboarding(j)
}
