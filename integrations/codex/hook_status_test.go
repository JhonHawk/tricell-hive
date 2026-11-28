package codex

import (
	"strings"
	"testing"

	"tricell-hive/integrations/target"
)

func TestEveryRegisteredEventHasAnEnglishStatusMessageNamingHive(t *testing.T) {
	f := SessionEventsFile(target.Config{Scope: "user", Home: t.TempDir(), CodexHome: t.TempDir()})
	for _, e := range f.Events {
		if m := f.StatusMessages[e]; !strings.HasPrefix(m, "Hive: ") {
			t.Fatalf("event %s statusMessage = %q", e, m)
		}
	}
	if len(f.StatusMessages) != len(f.Events) {
		t.Fatalf("status messages = %v", f.StatusMessages)
	}
}

// The /hooks screen tells the user that SessionStart also gives the model
// instructions, since that is what the user approves.
func TestSessionStartStatusMessageNamesTheDeveloperContext(t *testing.T) {
	f := SessionEventsFile(target.Config{Scope: "user", Home: t.TempDir(), CodexHome: t.TempDir()})
	if m := f.StatusMessages["SessionStart"]; !strings.Contains(m, "developer context") {
		t.Fatalf("SessionStart statusMessage = %q", m)
	}
}
