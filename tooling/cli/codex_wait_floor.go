package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"tricell-hive/integrations/codex"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// The codex-wait-floor capability sets a minimum wait for a parent agent in
// Codex's config.toml. Hive writes the two keys itself, through management's
// plan (Options.CodexWaitFloor), so the adapter emits no external step for it:
// Detect decides whether to offer it and Plan describes the effect. A
// config.toml Hive will not edit becomes a manual step that names the cause.

// Codex 0.164.0 is the first release whose [features.multi_agent_v2] table
// takes the wait keys; an older Codex may refuse to load a configuration that
// holds them.
var codexVersionPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// codexSupportsWaitFloor reports whether the stdout of `codex --version` names
// Codex 0.164.0 or later. Output without a version is not support.
func codexSupportsWaitFloor(stdout string) bool {
	m := codexVersionPattern.FindStringSubmatch(stdout)
	if m == nil {
		return false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major > 0 || minor >= 164
}

// waitFloorChosen reports whether requests select the wait floor as an
// editable offer. A manual offer (a file Hive will not edit) writes nothing, so
// choosing it changes no plan.
func waitFloorChosen(offers []providerOffer, requests []providerRequest) bool {
	for _, request := range requests {
		if request.ID != string(providers.CodexWaitFloor) {
			continue
		}
		for _, offer := range offers {
			if offer.ID == request.ID && !offer.ManualOnly {
				return true
			}
		}
	}
	return false
}

// planWithWaitFloor returns the plan the operator's choice leads to: p itself
// when the wait floor was not chosen, otherwise the install plan rebuilt with
// Options.CodexWaitFloor, with unchanged recomputed because the plan now writes
// Codex's config.
func planWithWaitFloor(o management.Options, p management.Plan, unchanged bool, offers []providerOffer, requests []providerRequest) (management.Plan, bool, error) {
	if !waitFloorChosen(offers, requests) {
		return p, unchanged, nil
	}
	o.CodexWaitFloor = true
	rebuilt, err := management.BuildPlan("install", o)
	if err != nil {
		return p, unchanged, err
	}
	same, err := management.PlanUnchanged(rebuilt)
	return rebuilt, same, err
}

// waitFloorDecision is Detect's cached verdict, so Plan describes exactly what
// Detect offered. cause is set when Hive will not edit the file.
type waitFloorDecision struct {
	cause string
}

// waitFloorCause reads Codex's config.toml and reports whether the wait floor
// is already applied and, when Hive cannot edit the file, a fixed phrase that
// names why. It never returns text from the file, which may hold secrets.
func waitFloorCause(codexHome string) (applied bool, cause string) {
	fail := func(err error) (bool, string) {
		var refusal *codex.WaitFloorRefusal
		if errors.As(err, &refusal) {
			return false, refusal.Cause
		}
		return false, "config.toml could not be inspected"
	}
	exists, err := codex.CheckWaitFloorTarget(codexHome)
	if err != nil {
		return fail(err)
	}
	var content []byte
	if exists {
		if content, err = os.ReadFile(filepath.Join(codexHome, "config.toml")); err != nil {
			return false, "config.toml could not be read"
		}
	}
	state, _, err := codex.WaitFloor(content, exists)
	if err != nil {
		return fail(err)
	}
	return state == codex.WaitFloorApplied, ""
}

// waitFloorOffer is Detect's decision for the capability. It reads the file
// first: an applied table is never offered and costs no program run, also under
// --dry-run. Otherwise it offers the capability only when the installed Codex
// is 0.164 or later, which it can verify only where programs run; with an older
// Codex, no Codex or no way to verify, it offers nothing, not even a manual step.
func (a *nativeProviderAdapter) waitFloorOffer(o management.Options) (providerOffer, bool) {
	a.waitFloor = nil
	if o.Scope == "project" || a.codexHome == "" {
		return providerOffer{}, false
	}
	applied, cause := waitFloorCause(a.codexHome)
	if applied || !a.runsPrograms() {
		return providerOffer{}, false
	}
	if _, err := a.system(o.Home).lookPath(codexBinary); err != nil {
		return providerOffer{}, false
	}
	res, stdout := codexVersionProbe()
	if res.ExitCode != 0 || res.TimedOut || res.Interrupted || !codexSupportsWaitFloor(stdout) {
		return providerOffer{}, false
	}
	a.waitFloor = &waitFloorDecision{cause: cause}
	return providerOffer{
		ID: string(providers.CodexWaitFloor), Name: string(providers.CodexWaitFloor), Source: providerSource(providers.CodexWaitFloor),
		ManualOnly: cause != "", DefaultOn: cause == "",
		Effect: "keeps Codex's parent agent waiting for its subagents in longer stretches",
	}, true
}

// planCodexWaitFloorSteps describes the capability for the plan Detect made. An
// editable file produces no step: the plan itself carries the write. A refusal
// produces the one manual step.
func (a *nativeProviderAdapter) planCodexWaitFloorSteps(preview *onboardingPreview) error {
	if a.waitFloor == nil {
		return fmt.Errorf("%s was not offered", providers.CodexWaitFloor)
	}
	config := filepath.Join(a.codexHome, "config.toml")
	keys := fmt.Sprintf("min_wait_timeout_ms = %d and default_wait_timeout_ms = %d under [features.multi_agent_v2]", codex.WaitFloorMinMS, codex.WaitFloorDefaultMS)
	var effects []string
	if a.waitFloor.cause != "" {
		reason := "Hive did not edit Codex's config.toml: " + a.waitFloor.cause + "; to set the wait floor, add " + keys + " by hand"
		step := providers.Step{ID: providers.CodexWaitFloorManualStep, Provider: providers.CodexWaitFloor, Status: providers.Manual, ManualReason: reason}
		payload, err := json.Marshal(step)
		if err != nil {
			return err
		}
		preview.Steps = append(preview.Steps, management.ExternalStep{ID: step.ID, Payload: payload})
		effects = []string{string(step.Status) + " — " + reason}
	} else {
		effects = []string{
			"writes " + keys + " in " + config + ", inserting only the keys that are missing; no other line changes",
			"before the first write, copies the file to config.toml.hive-backup (mode 0600) next to it, unless that copy already exists",
			"undo with `hive plan remove` for the codex host: it removes only these lines while they keep their values, then deletes the backup",
		}
	}
	preview.Details = append(preview.Details, providerDetail{
		ID:      string(providers.CodexWaitFloor),
		Source:  providerSource(providers.CodexWaitFloor),
		Effects: effects,
	})
	return nil
}
