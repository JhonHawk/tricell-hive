// Package providers describes optional third-party capabilities without
// owning their configuration or lifecycle. Engram, Context7 and pi-subagents
// resolve to a single Manual step naming the official source and the reason.
// agent-browser is the one capability with an executable recipe: two steps
// whose argv is fixed by this package and bound to the step ID. management
// persists consent and receipts for the steps; this package only types them
// and validates them, and never runs anything.
package providers

type ID string

const (
	Engram       ID = "engram"
	Context7     ID = "context7"
	PiSubagents  ID = "pi-subagents"
	AgentBrowser ID = "agent-browser"
)

// Status is kept distinct from management's own step-status strings so a
// provider payload is never confused with the parent onboarding journal's
// vocabulary, even though both currently serialize "manual" the same way.
type Status string

const (
	// Manual marks a step a person finishes by hand.
	Manual Status = "manual"
	// Planned marks an executable step whose argv ValidateStep has bound to
	// its ID; the runner executes it after the operator's consent.
	Planned Status = "planned"
)

// Methods name the program an executable step runs; each must equal argv[0].
const (
	MethodBrew         = "brew"
	MethodNpm          = "npm"
	MethodAgentBrowser = "agent-browser"
)

// Step IDs of the agent-browser capability.
const (
	AgentBrowserInstallStep = "agent-browser-install"
	AgentBrowserChromeStep  = "agent-browser-chrome"
	AgentBrowserManualStep  = "agent-browser-manual"
)

// Provider is the stable inventory entry the CLI renders in its capability
// list and summary.
type Provider struct {
	ID ID
	// Hosts are the CLIs this capability is relevant to; not proof that a
	// listed host has received native integration testing.
	Hosts []string
	// Source is the official documentation the operator should follow.
	Source string
	// Reason is the short, human-facing explanation of what Hive offers for
	// this capability, shown when the capability is manual.
	Reason string
}

// Step is the provider-side journal unit persisted inside the parent
// onboarding journal's payload (management.ExternalStep.Payload).
type Step struct {
	ID           string `json:"id"`
	Provider     ID     `json:"provider"`
	Status       Status `json:"status"`
	ManualReason string `json:"manual_reason,omitempty"`
	// Method and Argv exist only on Planned steps; omitted from manual steps
	// so payloads written before executable steps keep validating.
	Method string   `json:"method,omitempty"`
	Argv   []string `json:"argv,omitempty"`
}
