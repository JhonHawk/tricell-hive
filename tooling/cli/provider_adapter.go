package main

import (
	"encoding/json"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// nativeProviderAdapter offers Hive's optional capabilities as manual
// instructions only: no capability in providers.Catalog has passed its own
// native install/host-integration validation gate yet, so every selection
// produces a single Manual step naming the official source and the reason,
// never a process execution. It carries no state and switches on nothing
// injected by its caller: Detect/Plan behave identically in a dry run, a
// real install, and every test.
type nativeProviderAdapter struct{}

func nativeProviderAdapterFactory(onboardingInput) (onboardingAdapter, error) {
	return nativeProviderAdapter{}, nil
}

func (nativeProviderAdapter) Detect(o management.Options) ([]providerOffer, error) {
	var result []providerOffer
	for _, p := range providers.Catalog() {
		if !hostsIntersect(o.Hosts, p.Hosts) {
			continue
		}
		result = append(result, providerOffer{ID: string(p.ID), Name: string(p.ID), Source: p.Source, ManualOnly: true})
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

func (nativeProviderAdapter) Plan(_ management.Plan, requests []providerRequest) (onboardingPreview, error) {
	var preview onboardingPreview
	for _, request := range requests {
		id := providers.ID(request.ID)
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

func (nativeProviderAdapter) Runner() management.ExternalRunner { return manualProviderRunner{} }

// manualProviderRunner never executes a process: every step this adapter
// plans already carries a terminal Manual status. Validate is the only check
// that can fail (an invalid or tampered payload); Execute and Reconcile only
// confirm that invariant back to the parent onboarding journal.
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
