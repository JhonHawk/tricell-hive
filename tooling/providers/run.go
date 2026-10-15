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

// ValidateStep admits a manual step for any catalog provider and, for
// agent-browser only, the two Planned steps whose argv matches the recipe
// bound to their ID. Everything else is rejected, including argv or method on
// a manual step and any unknown field.
func ValidateStep(payload json.RawMessage) error {
	var s Step
	d := json.NewDecoder(bytes.NewReader(payload))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		return fmt.Errorf("invalid provider payload")
	}
	switch s.Provider {
	case Engram, Context7, PiSubagents, AgentBrowser:
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
		if s.Provider == AgentBrowser && s.ID != AgentBrowserManualStep {
			return fmt.Errorf("unknown agent-browser step")
		}
		return nil
	case Planned:
		if s.Provider != AgentBrowser {
			return fmt.Errorf("only agent-browser has an executable recipe")
		}
		if s.ManualReason != "" {
			return fmt.Errorf("executable provider step cannot carry a manual reason")
		}
		switch s.ID {
		case AgentBrowserInstallStep:
			want, ok := allowedInstallArgv[s.Method]
			if !ok || !slices.Equal(s.Argv, want) {
				return fmt.Errorf("command not allowed for %s", s.ID)
			}
		case AgentBrowserChromeStep:
			if s.Method != MethodAgentBrowser || !slices.Equal(s.Argv, chromeArgv) {
				return fmt.Errorf("command not allowed for %s", s.ID)
			}
		default:
			return fmt.Errorf("unknown agent-browser step")
		}
		return nil
	}
	return fmt.Errorf("unsupported provider step status")
}
