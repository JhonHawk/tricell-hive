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

func TestCatalogEndsWithAgentBrowserForAllHosts(t *testing.T) {
	catalog := Catalog()
	last := catalog[len(catalog)-1]
	if last.ID != AgentBrowser {
		t.Fatalf("last catalog entry = %q, want agent-browser", last.ID)
	}
	want := []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}
	got := append([]string(nil), last.Hosts...)
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("agent-browser hosts = %v, want %v", got, want)
	}
	if last.Source != "github.com/vercel-labs/agent-browser" {
		t.Fatalf("agent-browser source = %q", last.Source)
	}
	// The existing indexes stay valid: the three manual providers come first.
	for i, id := range []ID{Engram, Context7, PiSubagents} {
		if catalog[i].ID != id {
			t.Fatalf("catalog[%d] = %q, want %q", i, catalog[i].ID, id)
		}
	}
}
