// Package providers describes optional third-party capabilities without
// owning their configuration or lifecycle. None of them has passed its own
// native install/host-integration validation gate yet (design.md,
// "Dependencias y recetas"), so this build never executes an installer for
// them: every capability the catalog lists resolves to a single Manual step
// naming the official source and the reason. management persists consent and
// receipts for that step; this package only types it and validates it.
package providers

type ID string

const (
	Engram      ID = "engram"
	Context7    ID = "context7"
	PiSubagents ID = "pi-subagents"
)

// Status is kept distinct from management's own step-status strings so a
// provider payload is never confused with the parent onboarding journal's
// vocabulary, even though both currently serialize "manual" the same way.
type Status string

// Manual is the only status this build ever produces: no recipe here has
// passed its own native validation gate.
const Manual Status = "manual"

// Provider is the stable inventory entry the CLI renders in its capability
// list and summary.
type Provider struct {
	ID ID
	// Hosts are the CLIs this capability is relevant to; not proof that a
	// listed host has received native integration testing.
	Hosts []string
	// Source is the official documentation the operator should follow.
	Source string
	// Reason is the short, human-facing explanation for why Hive only offers
	// manual instructions for this capability in this build.
	Reason string
}

// Step is the provider-side journal unit persisted inside the parent
// onboarding journal's payload (management.ExternalStep.Payload).
type Step struct {
	ID           string `json:"id"`
	Provider     ID     `json:"provider"`
	Status       Status `json:"status"`
	ManualReason string `json:"manual_reason,omitempty"`
}
