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
