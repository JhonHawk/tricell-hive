package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// nativeProviderAdapter offers Hive's optional capabilities. Engram, Context7
// and pi-subagents are manual instructions only. agent-browser and codex-question-card have
// executable recipes (the latter only runs Codex's own `features enable`): it is offered without a version, its
// steps run only after the operator's consent, and it runs no program at all
// in a non-interactive session, a dry run or under a synthetic home (D-9).
// A fresh adapter is created for each invocation and keeps only Detect's
// decision so Plan describes exactly what Detect offered.
type nativeProviderAdapter struct {
	sys         agentBrowserSystem
	synthetic   bool
	dryRun      bool
	interactive bool
	browser     *agentBrowserPlan
	// codexHome is the resolved Codex home the question-card step is checked
	// against; codexOnPath caches the one PATH lookup for `codex`.
	codexHome   string
	codexOnPath *bool
	// codexSupport caches the one `codex features list` check made by Detect.
	codexSupport *codexSupportDecision
}

// codexSupportDecision is Detect's cached verdict on the question-card offer.
type codexSupportDecision struct {
	// found is true when `codex` is on PATH; executable additionally means the
	// installed Codex was confirmed to support the command, or nothing ran to
	// check it (dry run and non-interactive, where no step is executed).
	found, executable bool
}

// codexFeaturesTimeout bounds the `codex features list` check.
const codexFeaturesTimeout = 30 * time.Second

// codexUnsupportedReason is the manual reason for a Codex that is on PATH but
// did not report the feature.
const codexUnsupportedReason = "update Codex to 0.163.0 or later (`codex features list` did not report `" + providers.CodexQuestionCardFeature + "`) and run `codex features enable " + providers.CodexQuestionCardFeature + "`"

// codexFeaturesProbe runs `codex features list` and returns its result and
// standard output. Tests replace it.
var codexFeaturesProbe = func() (agentBrowserResult, string) {
	return runBoundedCommandContext(context.Background(), []string{codexBinary, "features", "list"}, codexFeaturesTimeout, time.Second, agentBrowserMaxOutput)
}

func nativeProviderAdapterFactory(in onboardingInput) (onboardingAdapter, error) {
	a := newNativeProviderAdapter(in.Options.Home != "", in.DryRun, in.Interactive)
	// The same resolution the install itself uses, so the runner reads the
	// config.toml that `codex features enable` writes (the child inherits
	// CODEX_HOME through the environment).
	// When the options cannot be resolved the home stays empty and the step is
	// never confirmed from the configuration (the install itself reports the
	// invalid options).
	if cfg, _, err := management.NormalizeOptions(in.Options); err == nil {
		a.codexHome = cfg.CodexHome
	}
	return a, nil
}

// newNativeProviderAdapter snapshots agentBrowserSys. A synthetic home detects
// nothing of the real machine (no PATH lookup, no environment) and uses its
// own directory as home, as detectInstallerHosts does.
func newNativeProviderAdapter(synthetic, dryRun, interactive bool) *nativeProviderAdapter {
	return &nativeProviderAdapter{sys: agentBrowserSys, synthetic: synthetic, dryRun: dryRun, interactive: interactive}
}

// system returns the probes for home: the real ones, or hermetic ones under a
// synthetic home.
func (a *nativeProviderAdapter) system(home string) agentBrowserSystem {
	s := a.sys
	if a.synthetic {
		s.lookPath = func(name string) (string, error) {
			return "", &exec.Error{Name: name, Err: exec.ErrNotFound}
		}
		s.getenv = func(string) string { return "" }
		s.userHome = func() (string, error) { return home, nil }
		// Nothing of the real machine is read: no browser exists here.
		s.pathExists = func(string) bool { return false }
		s.isFile = func(string) bool { return false }
		s.dirHasEntries = func(string) bool { return false }
		s.browserCandidates = func(string) []browserCandidate { return nil }
	}
	return s
}

// runsPrograms is true only in a real interactive install: the sole mode where
// Detect may run `node --version` and `npm config get prefix`.
func (a *nativeProviderAdapter) runsPrograms() bool {
	return a.interactive && !a.dryRun && !a.synthetic
}

func (a *nativeProviderAdapter) agentBrowserPlan(home string) agentBrowserPlan {
	if a.browser == nil {
		plan := planAgentBrowser(a.system(home), a.runsPrograms())
		a.browser = &plan
	}
	return *a.browser
}

// codexFound looks `codex` up once. A synthetic home detects nothing of the
// real machine, so it never finds it.
func (a *nativeProviderAdapter) codexFound(home string) bool {
	return a.codexDecision(home).found
}

// codexDecision looks `codex` up and, in a real interactive install, checks
// once that it supports the command: exit 0 and the feature in the stdout of
// `codex features list`. A dry run and a non-interactive session run nothing,
// so the PATH lookup alone makes the offer executable there; a synthetic home
// finds nothing.
func (a *nativeProviderAdapter) codexDecision(home string) codexSupportDecision {
	if a.codexSupport == nil {
		_, err := a.system(home).lookPath(codexBinary)
		d := codexSupportDecision{found: err == nil}
		d.executable = d.found
		if d.found && a.runsPrograms() {
			res, stdout := codexFeaturesProbe()
			d.executable = res.ExitCode == 0 && !res.TimedOut && !res.Interrupted && strings.Contains(stdout, providers.CodexQuestionCardFeature)
		}
		a.codexSupport = &d
	}
	return *a.codexSupport
}

func (a *nativeProviderAdapter) Detect(o management.Options) ([]providerOffer, error) {
	var result []providerOffer
	for _, p := range providers.Catalog() {
		if !hostsIntersect(o.Hosts, p.Hosts) {
			continue
		}
		offer := providerOffer{ID: string(p.ID), Name: string(p.ID), Source: p.Source, ManualOnly: true}
		if p.ID == providers.AgentBrowser {
			plan := a.agentBrowserPlan(o.Home)
			if plan.Nothing {
				continue
			}
			offer.ManualOnly = plan.manual()
			offer.Unversioned = !plan.manual()
		}
		if p.ID == providers.CodexQuestionCard {
			offer.ManualOnly = !a.codexDecision(o.Home).executable
			offer.DefaultOn = !offer.ManualOnly
			if offer.DefaultOn {
				offer.Effect = "enables Codex's question card in Default mode"
			}
		}
		result = append(result, offer)
	}
	return result, nil
}

func hostsIntersect(selected, supported []string) bool {
	for _, host := range selected {
		for _, s := range supported {
			if host == s {
				return true
			}
		}
	}
	return false
}

// providerReason returns the catalog reason for id, matching Detect's own
// Catalog() traversal instead of trusting an offer/request round-trip.
func providerReason(id providers.ID) string {
	for _, p := range providers.Catalog() {
		if p.ID == id {
			return p.Reason
		}
	}
	return ""
}

func providerSource(id providers.ID) string {
	for _, p := range providers.Catalog() {
		if p.ID == id {
			return p.Source
		}
	}
	return ""
}

func (a *nativeProviderAdapter) Plan(_ management.Plan, requests []providerRequest) (onboardingPreview, error) {
	var preview onboardingPreview
	for _, request := range requests {
		id := providers.ID(request.ID)
		if id == providers.AgentBrowser {
			if err := a.planAgentBrowserSteps(&preview); err != nil {
				return preview, err
			}
			continue
		}
		if id == providers.CodexQuestionCard {
			if err := a.planCodexQuestionCardSteps(&preview); err != nil {
				return preview, err
			}
			continue
		}
		reason := providerReason(id)
		step := providers.Step{ID: request.ID + "-manual", Provider: id, Status: providers.Manual, ManualReason: reason}
		payload, err := json.Marshal(step)
		if err != nil {
			return preview, err
		}
		preview.Steps = append(preview.Steps, management.ExternalStep{ID: step.ID, Payload: payload})
		preview.Details = append(preview.Details, providerDetail{
			ID:      request.ID,
			Source:  providerSource(id),
			Effects: []string{string(step.Status) + " — " + reason},
		})
	}
	return preview, nil
}

// planAgentBrowserSteps adds agent-browser's steps for the plan Detect made:
// the install step when the program is missing, the Chrome step when no
// browser was found, or a single manual step naming what is missing.
func (a *nativeProviderAdapter) planAgentBrowserSteps(preview *onboardingPreview) error {
	plan := a.agentBrowserPlan("")
	var steps []providers.Step
	switch {
	case plan.manual():
		steps = append(steps, providers.Step{ID: providers.AgentBrowserManualStep, Provider: providers.AgentBrowser, Status: providers.Manual, ManualReason: plan.ManualReason})
	default:
		if plan.Method == providers.MethodBrew {
			steps = append(steps, providers.Step{ID: providers.AgentBrowserInstallStep, Provider: providers.AgentBrowser, Status: providers.Planned, Method: providers.MethodBrew, Argv: []string{"brew", "install", "agent-browser"}})
		}
		if plan.Method == providers.MethodNpm {
			steps = append(steps, providers.Step{ID: providers.AgentBrowserInstallStep, Provider: providers.AgentBrowser, Status: providers.Planned, Method: providers.MethodNpm, Argv: []string{"npm", "install", "-g", "agent-browser"}})
		}
		if plan.Chrome {
			steps = append(steps, providers.Step{ID: providers.AgentBrowserChromeStep, Provider: providers.AgentBrowser, Status: providers.Planned, Method: providers.MethodAgentBrowser, Argv: []string{"agent-browser", "install"}})
		}
	}
	if len(steps) == 0 {
		return nil
	}
	for _, step := range steps {
		payload, err := json.Marshal(step)
		if err != nil {
			return err
		}
		preview.Steps = append(preview.Steps, management.ExternalStep{ID: step.ID, Payload: payload})
	}
	preview.Details = append(preview.Details, providerDetail{
		ID:      string(providers.AgentBrowser),
		Source:  providerSource(providers.AgentBrowser),
		Effects: plan.effects(),
	})
	return nil
}

// planCodexQuestionCardSteps adds the single step for the plan Detect made:
// Codex's own `features enable` when `codex` is on PATH, otherwise a manual
// step that names the command.
func (a *nativeProviderAdapter) planCodexQuestionCardSteps(preview *onboardingPreview) error {
	command := "codex features enable " + providers.CodexQuestionCardFeature
	undo := "codex features disable " + providers.CodexQuestionCardFeature
	var step providers.Step
	var effects []string
	decision := a.codexDecision("")
	if decision.executable {
		step = providers.Step{ID: providers.CodexQuestionCardEnableStep, Provider: providers.CodexQuestionCard, Status: providers.Planned, Method: providers.MethodCodex,
			Argv: []string{codexBinary, "features", "enable", providers.CodexQuestionCardFeature}}
		effects = []string{
			"runs `" + command + "`, which writes " + providers.CodexQuestionCardFeature + " = true to config.toml in the Codex home (Codex writes the file, not Hive)",
			"undo with `" + undo + "`",
		}
	} else {
		reason := "install Codex (the `codex` executable was not found in PATH) and run `" + command + "`"
		if decision.found {
			reason = codexUnsupportedReason
		}
		step = providers.Step{ID: providers.CodexQuestionCardManualStep, Provider: providers.CodexQuestionCard, Status: providers.Manual, ManualReason: reason}
		effects = []string{string(step.Status) + " — " + reason}
	}
	payload, err := json.Marshal(step)
	if err != nil {
		return err
	}
	preview.Steps = append(preview.Steps, management.ExternalStep{ID: step.ID, Payload: payload})
	preview.Details = append(preview.Details, providerDetail{
		ID:      string(providers.CodexQuestionCard),
		Source:  providerSource(providers.CodexQuestionCard),
		Effects: effects,
	})
	return nil
}

func (a *nativeProviderAdapter) Runner() management.ExternalRunner {
	return providerRunner{sys: a.system(""), synthetic: a.synthetic, codexHome: a.codexHome}
}

// manualProviderRunner never executes a process: every manual step already
// carries its terminal Manual status. Validate is the only check that can fail
// (an invalid or tampered payload); Execute and Reconcile only confirm that
// invariant back to the parent onboarding journal. providerRunner delegates
// manual steps to it.
type manualProviderRunner struct{}

func (manualProviderRunner) Validate(s management.ExternalStep) error {
	return providers.ValidateStep(s.Payload)
}
func (manualProviderRunner) Execute(s management.ExternalStep) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (manualProviderRunner) Reconcile(management.ExternalStep, json.RawMessage) (string, error) {
	return string(providers.Manual), nil
}
