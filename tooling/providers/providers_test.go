package providers

import (
	"sort"
	"strings"
	"testing"
)

func TestCatalogEntriesCarrySourceAndReason(t *testing.T) {
	for _, provider := range Catalog() {
		if provider.ID == "" {
			t.Fatalf("provider with empty ID: %#v", provider)
		}
		if len(provider.Hosts) == 0 {
			t.Fatalf("%s advertises no hosts", provider.ID)
		}
		if provider.Source == "" {
			t.Fatalf("%s has no official source", provider.ID)
		}
		if provider.Reason == "" {
			t.Fatalf("%s has no reason it is manual-only", provider.ID)
		}
	}
}

func TestCatalogOrderIsStableAndAgentBrowserPrecedesCodexQuestionCard(t *testing.T) {
	catalog := Catalog()
	// The existing indexes stay valid: the three manual providers come first,
	// then agent-browser, then the entries added later (codex-question-card,
	// then codegraph, then codex-wait-floor).
	wantOrder := []ID{Engram, Context7, PiSubagents, AgentBrowser, CodexQuestionCard, CodeGraph, CodexWaitFloor}
	if len(catalog) != len(wantOrder) {
		t.Fatalf("catalog has %d entries, want %d", len(catalog), len(wantOrder))
	}
	for i, id := range wantOrder {
		if catalog[i].ID != id {
			t.Fatalf("catalog[%d] = %q, want %q", i, catalog[i].ID, id)
		}
	}

	browser := catalog[3]
	want := []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}
	got := append([]string(nil), browser.Hosts...)
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("agent-browser hosts = %v, want %v", got, want)
	}
	if browser.Source != "github.com/vercel-labs/agent-browser" {
		t.Fatalf("agent-browser source = %q", browser.Source)
	}

	card := catalog[4]
	if len(card.Hosts) != 1 || card.Hosts[0] != "codex" {
		t.Fatalf("codex-question-card hosts = %v, want [codex]", card.Hosts)
	}
	if card.Source != "github.com/openai/codex" {
		t.Fatalf("codex-question-card source = %q", card.Source)
	}
	if !strings.Contains(card.Reason, "codex features enable default_mode_request_user_input") {
		t.Fatalf("codex-question-card reason does not name the command: %q", card.Reason)
	}
	graph := catalog[5]
	wantGraphHosts := []string{"claude", "codex", "cursor", "opencode"}
	gotGraphHosts := append([]string(nil), graph.Hosts...)
	sort.Strings(gotGraphHosts)
	if strings.Join(gotGraphHosts, ",") != strings.Join(wantGraphHosts, ",") {
		t.Fatalf("codegraph hosts = %v, want %v", gotGraphHosts, wantGraphHosts)
	}
	if graph.Source != "github.com/colbymchenry/codegraph" {
		t.Fatalf("codegraph source = %q", graph.Source)
	}
	if !strings.Contains(graph.Reason, "codegraph install") || !strings.Contains(graph.Reason, "Hive does not run it") {
		t.Fatalf("codegraph reason does not say Hive runs nothing: %q", graph.Reason)
	}

	floor := catalog[6]
	if len(floor.Hosts) != 1 || floor.Hosts[0] != "codex" {
		t.Fatalf("codex-wait-floor hosts = %v, want [codex]", floor.Hosts)
	}
	if floor.Source != "github.com/openai/codex" {
		t.Fatalf("codex-wait-floor source = %q", floor.Source)
	}
	if !strings.Contains(floor.Reason, "multi_agent_v2") || !strings.Contains(floor.Reason, "config.toml") {
		t.Fatalf("codex-wait-floor reason does not name what it writes: %q", floor.Reason)
	}
}
