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

func TestValidateStepCodexQuestionCard(t *testing.T) {
	argv := []string{"codex", "features", "enable", "default_mode_request_user_input"}
	accepted := []Step{
		{ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: argv},
		{ID: "codex-question-card-manual", Provider: CodexQuestionCard, Status: Manual, ManualReason: "codex was not found"},
	}
	for _, s := range accepted {
		if err := ValidateStep(stepPayload(t, s)); err != nil {
			t.Errorf("%s rejected: %v", s.ID, err)
		}
	}
	rejected := map[string]Step{
		"another feature":                 {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: []string{"codex", "features", "enable", "other"}},
		"disable":                         {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: []string{"codex", "features", "disable", "default_mode_request_user_input"}},
		"extra argument":                  {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: append(append([]string(nil), argv...), "--foo")},
		"other method":                    {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "npm", Argv: argv},
		"method differs from argv":        {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: []string{"sh", "-c", "codex"}},
		"planned without argv":            {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex"},
		"enable with a reason":            {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: argv, ManualReason: "r"},
		"unknown step ID":                 {ID: "codex-question-card-other", Provider: CodexQuestionCard, Status: Planned, Method: "codex", Argv: argv},
		"manual with another ID":          {ID: "codex-question-card-other", Provider: CodexQuestionCard, Status: Manual, ManualReason: "r"},
		"manual with a command":           {ID: "codex-question-card-manual", Provider: CodexQuestionCard, Status: Manual, ManualReason: "r", Method: "codex", Argv: argv},
		"enable ID under agent-browser":   {ID: "codex-question-card-enable", Provider: AgentBrowser, Status: Planned, Method: "codex", Argv: argv},
		"enable ID under engram":          {ID: "codex-question-card-enable", Provider: Engram, Status: Planned, Method: "codex", Argv: argv},
		"agent-browser argv under card":   {ID: "codex-question-card-enable", Provider: CodexQuestionCard, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
		"card ID with agent-browser argv": {ID: "agent-browser-install", Provider: CodexQuestionCard, Status: Planned, Method: "brew", Argv: []string{"brew", "install", "agent-browser"}},
	}
	for name, s := range rejected {
		if ValidateStep(stepPayload(t, s)) == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestValidateStepCodeGraphIsManualOnly(t *testing.T) {
	manual := Step{ID: "codegraph-manual", Provider: CodeGraph, Status: Manual, ManualReason: "follow the official instructions"}
	if err := ValidateStep(stepPayload(t, manual)); err != nil {
		t.Errorf("codegraph-manual rejected: %v", err)
	}
	rejected := map[string]Step{
		"planned with argv":       {ID: "codegraph-install", Provider: CodeGraph, Status: Planned, Method: "codegraph", Argv: []string{"codegraph", "install"}},
		"planned without argv":    {ID: "codegraph-manual", Provider: CodeGraph, Status: Planned},
		"manual with a command":   {ID: "codegraph-manual", Provider: CodeGraph, Status: Manual, ManualReason: "r", Method: "codegraph", Argv: []string{"codegraph", "install"}},
		"manual without a reason": {ID: "codegraph-manual", Provider: CodeGraph, Status: Manual},
	}
	for name, s := range rejected {
		if ValidateStep(stepPayload(t, s)) == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
