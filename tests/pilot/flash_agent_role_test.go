package main

import (
	"encoding/json"
	"testing"
)

func TestFlashAgentRoleID(t *testing.T) {
	role := "hive-review-plan"
	cases := []struct {
		name, agent, agentName, model string
		valid                         bool
	}{
		{"empty matching ID", role, "", flashModels[0], true},
		{"empty different ID", "other", "", flashModels[0], false},
		{"different display label", role, "Plan Reviewer", flashModels[0], true},
		{"model mismatch", role, "", "other/model", false},
		{"matching label wrong ID", "other", role, flashModels[0], false},
		{"matching display name", role, role, flashModels[0], true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := flashChild{SessionID: "child", ParentID: "root", Agent: tc.agent, AgentName: tc.agentName, Model: tc.model, Completed: true, FinalAssistantID: "final"}
			if got := flashCompletion("root", flashModels[0], role, []flashChild{c}) == nil; got != tc.valid {
				t.Error("incorrect child role/model acceptance")
			}
			root := flashObservation{Phase: "http", SessionID: "root", Model: flashModels[0], Correlated: true, Protocol: "chat-completions", LoaderActive: true}
			child := flashObservation{Phase: "http", SessionID: "child", ParentID: "root", Agent: tc.agent, AgentName: tc.agentName, Model: tc.model, Correlated: true, Protocol: "chat-completions", Canonical: true, Copies: 1, InjectedCopies: 1, Hash: "canonical", LoaderActive: true}
			if got := flashCoverage("C", "root", flashModels[0], role, "canonical", []flashObservation{root, child}) == "observed"; got != tc.valid {
				t.Error("incorrect observer role/model acceptance")
			}
		})
	}
}

func TestFlashNativeChildWithoutAgentName(t *testing.T) {
	role := "hive-review-plan"
	model := map[string]any{"providerID": "opencode-go", "id": "claude-haiku-5-5", "variant": "high"}
	doc := map[string]any{"info": map[string]any{"id": "child", "parentID": "root", "agent": role, "model": model, "outcome": "succeeded", "time": map[string]any{"idle": 1}}, "messages": []any{map[string]any{"id": "final", "type": "assistant", "agent": role, "model": model, "finish": "stop", "time": map[string]any{"completed": 1}}}}
	raw, _ := json.Marshal(doc)
	child, err := flashChildFromExport(raw, "root", "", flashModels[0])
	if err != nil {
		t.Fatal(err)
	}
	if child.AgentName != "" {
		t.Fatal("missing name fabricated")
	}
	if flashCompletion("root", flashModels[0], role, []flashChild{child}) != nil {
		t.Error("native Markdown agent rejected")
	}
}
