package main

import (
	"encoding/json"
	"os/exec"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// nativeProviderAdapter offers Hive's optional capabilities. Engram, Context7
// and pi-subagents are manual instructions only. agent-browser is the one
// capability with an executable recipe: it is offered without a version, its
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
}

func nativeProviderAdapterFactory(in onboardingInput) (onboardingAdapter, error) {
	return newNativeProviderAdapter(in.Options.Home != "", in.DryRun, in.Interactive), nil
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

func (a *nativeProviderAdapter) Runner() management.ExternalRunner {
	return providerRunner{sys: a.system(""), synthetic: a.synthetic}
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
