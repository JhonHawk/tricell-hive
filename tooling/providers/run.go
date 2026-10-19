package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
)

// allowedInstallArgv are the only commands the agent-browser-install step may
// carry, keyed by Step.Method. The argv is bound to the step ID so a tampered
// or foreign payload can never name another program or argument.
var allowedInstallArgv = map[string][]string{
	MethodBrew: {"brew", "install", "agent-browser"},
	MethodNpm:  {"npm", "install", "-g", "agent-browser"},
}

var chromeArgv = []string{"agent-browser", "install"}

// plannedRecipe binds an executable step ID to the provider that owns it, the
// method it must carry and, when argv is set, the one argv it may carry. The
// agent-browser install step has two argv by method and is checked apart.
type plannedRecipe struct {
	provider ID
	method   string
	argv     []string
}

// allowedPlannedSteps lists the Planned steps ValidateStep admits besides the
// agent-browser install step, keyed by step ID.
var allowedPlannedSteps = map[string]plannedRecipe{
	AgentBrowserChromeStep: {AgentBrowser, MethodAgentBrowser, chromeArgv},
	CodexQuestionCardEnableStep: {CodexQuestionCard, MethodCodex,
		[]string{"codex", "features", "enable", CodexQuestionCardFeature}},
}

// manualStepIDs are the only manual step IDs of providers that also have an
// executable recipe; the manual-only providers accept any ID.
var manualStepIDs = map[ID]string{
	AgentBrowser:      AgentBrowserManualStep,
	CodexQuestionCard: CodexQuestionCardManualStep,
}

// ValidateStep admits a manual step for any catalog provider and, for
// agent-browser and codex-question-card only, the Planned steps whose argv
// matches the recipe bound to their ID. Everything else is rejected, including
// argv or method on a manual step and any unknown field.
func ValidateStep(payload json.RawMessage) error {
	var s Step
	d := json.NewDecoder(bytes.NewReader(payload))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return fmt.Errorf("invalid provider payload")
	}
	switch s.Provider {
	case Engram, Context7, PiSubagents, AgentBrowser, CodexQuestionCard:
	default:
		return fmt.Errorf("unknown provider")
	}
	switch s.Status {
	case Manual:
		if s.Method != "" || len(s.Argv) != 0 {
			return fmt.Errorf("manual provider step cannot carry a command")
		}
		if s.ManualReason == "" {
			return fmt.Errorf("manual provider step requires a reason")
		}
		if want, ok := manualStepIDs[s.Provider]; ok && s.ID != want {
			return fmt.Errorf("unknown %s step", s.Provider)
		}
		return nil
	case Planned:
		if s.Provider != AgentBrowser && s.Provider != CodexQuestionCard {
			return fmt.Errorf("provider %s has no executable recipe", s.Provider)
		}
		if s.ManualReason != "" {
			return fmt.Errorf("executable provider step cannot carry a manual reason")
		}
		if s.ID == AgentBrowserInstallStep {
			want, ok := allowedInstallArgv[s.Method]
			if s.Provider != AgentBrowser || !ok || !slices.Equal(s.Argv, want) {
				return fmt.Errorf("command not allowed for %s", s.ID)
			}
			return nil
		}
		recipe, ok := allowedPlannedSteps[s.ID]
		if !ok || recipe.provider != s.Provider {
			return fmt.Errorf("unknown %s step", s.Provider)
		}
		if s.Method != recipe.method || !slices.Equal(s.Argv, recipe.argv) {
			return fmt.Errorf("command not allowed for %s", s.ID)
		}
		return nil
	}
	return fmt.Errorf("unsupported provider step status")
}
