package providers

import (
	"encoding/json"
	"testing"
)

func TestValidateStepAcceptsOnlyManualStepForKnownProvider(t *testing.T) {
	valid := Step{ID: "context7-manual", Provider: Context7, Status: Manual, ManualReason: "fixture reason"}
	payload, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateStep(payload); err != nil {
		t.Fatalf("valid manual step rejected: %v", err)
	}

	unknownProvider := valid
	unknownProvider.Provider = "not-a-provider"
	payload, _ = json.Marshal(unknownProvider)
	if ValidateStep(payload) == nil {
		t.Fatal("unknown provider accepted")
	}

	notManual := valid
	notManual.Status = "verified"
	payload, _ = json.Marshal(notManual)
	if ValidateStep(payload) == nil {
		t.Fatal("non-manual status accepted although this build has no recipe")
	}

	noReason := valid
	noReason.ManualReason = ""
	payload, _ = json.Marshal(noReason)
	if ValidateStep(payload) == nil {
		t.Fatal("manual step accepted without a reason")
	}

	if ValidateStep(json.RawMessage(`{"id":"x","provider":"context7","status":"manual","manual_reason":"r","extra":true}`)) == nil {
		t.Fatal("payload with unknown field accepted")
	}
}

func stepPayload(t *testing.T, s Step) json.RawMessage {
	t.Helper()
	payload, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestValidateStepAgentBrowserRecipes(t *testing.T) {
	accepted := []Step{
		{ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
		{ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "npm", Argv: []string{"npm", "install", "-g", "agent-browser"}},
		{ID: "agent-browser-chrome", Provider: AgentBrowser, Status: Planned, Method: "agent-browser", Argv: []string{"agent-browser", "install"}},
		{ID: "agent-browser-manual", Provider: AgentBrowser, Status: Manual, ManualReason: "brew was not found"},
	}
	for _, s := range accepted {
		if err := ValidateStep(stepPayload(t, s)); err != nil {
			t.Errorf("%s %v rejected: %v", s.ID, s.Argv, err)
		}
	}
	rejected := map[string]Step{
		"argv on the chrome ID":      {ID: "agent-browser-chrome", Provider: AgentBrowser, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
		"chrome argv on install ID":  {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "agent-browser", Argv: []string{"agent-browser", "install"}},
		"shell":                      {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "sh", Argv: []string{"sh", "-c", "brew install agent-browser"}},
		"another formula":            {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "otra"}},
		"extra argument":             {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "npm", Argv: []string{"npm", "install", "-g", "agent-browser", "--foo"}},
		"method does not match argv": {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "npm", Argv: []string{"brew", "install", "agent-browser"}},
		"manual step with argv":      {ID: "agent-browser-manual", Provider: AgentBrowser, Status: Manual, ManualReason: "r", Argv: []string{"brew", "install", "agent-browser"}},
		"unknown agent-browser ID":   {ID: "agent-browser-other", Provider: AgentBrowser, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
		"planned without argv":       {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned},
		"argv on another provider":   {ID: "context7-manual", Provider: Context7, Status: Manual, ManualReason: "r", Argv: []string{"npm", "install", "-g", "agent-browser"}},
		"planned other provider":     {ID: "engram-install", Provider: Engram, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
		"install with a reason":      {ID: "agent-browser-install", Provider: AgentBrowser, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}, ManualReason: "r"},
	}
	for name, s := range rejected {
		if ValidateStep(stepPayload(t, s)) == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if ValidateStep(json.RawMessage(`{"id":"agent-browser-chrome","provider":"agent-browser","status":"planned","method":"agent-browser","argv":["agent-browser","install"],"env":"x"}`)) == nil {
		t.Error("unknown field accepted")
	}
	// A manual payload written before the executable steps existed still validates.
	if err := ValidateStep(json.RawMessage(`{"id":"pi-subagents-manual","provider":"pi-subagents","status":"manual","manual_reason":"r"}`)); err != nil {
		t.Errorf("old manual payload rejected: %v", err)
	}
}
