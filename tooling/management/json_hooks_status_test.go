package management

import (
	"encoding/json"
	"strings"
	"testing"

	"tricell-hive/integrations/hostevent"
)

func namedCodexSpec() hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word: "codex", Path: "/home/u/.codex/hooks.json",
		Events: []string{"SessionStart", "SessionEnd"},
		StatusMessages: map[string]string{
			"SessionStart": "Hive: report session start to the viewer",
			"SessionEnd":   "Hive: report session end to the viewer",
		},
	}
}

func ptrRec(r HostHooksRecord) *HostHooksRecord { return &r }

func TestStatusMessageIsWrittenOnInstallAndAddedInPlaceOnUpdate(t *testing.T) {
	named := namedCodexSpec()
	bare := named
	bare.StatusMessages = nil
	foreignEntry := `{ "type": "command", "command": "engram codex session-start", "timeout": 10, "statusMessage": "Loading engram memory..." }`
	foreign := "{\n  \"hooks\": {\n    \"SessionStart\": [\n      { \"hooks\": [ " + foreignEntry + " ] }\n    ]\n  }\n}\n"

	fresh := mustEdit(t, "", false, named, nil)
	if got := strings.Count(string(fresh.Data), `"statusMessage"`); got != 2 {
		t.Fatalf("fresh install statusMessage count = %d\n%s", got, fresh.Data)
	}

	old := mustEdit(t, foreign, true, bare, nil)
	if strings.Contains(string(old.Data), "Hive: report") {
		t.Fatalf("legacy entries must not carry the field:\n%s", old.Data)
	}
	upd := mustEdit(t, string(old.Data), true, named, ptrRec(rec(bare, old.Delta)))
	if !upd.Changed {
		t.Fatal("an entry without the field must be rewritten")
	}
	if got := countHive(t, upd.Data, named); got != 2 {
		t.Fatalf("hive entries = %d, want 2 (no duplicate)\n%s", got, upd.Data)
	}
	if got := strings.Count(string(upd.Data), "Hive: report"); got != 2 {
		t.Fatalf("hive status messages = %d\n%s", got, upd.Data)
	}
	if !strings.Contains(string(upd.Data), foreignEntry) {
		t.Fatalf("foreign entry changed:\n%s", upd.Data)
	}
	again := mustEdit(t, string(upd.Data), true, named, ptrRec(rec(named, upd.Delta)))
	if again.Changed {
		t.Fatalf("a second update must be a no-op:\n%s", again.Data)
	}

	un := undoJSONHooks(upd.Data, named, rec(named, upd.Delta))
	if un.Skipped != "" || strings.Contains(string(un.Data), "hook codex") || strings.Contains(string(un.Data), "Hive: report") {
		t.Fatalf("uninstall left Hive entries (skipped %q):\n%s", un.Skipped, un.Data)
	}
	if !strings.Contains(string(un.Data), foreignEntry) || !json.Valid(un.Data) {
		t.Fatalf("uninstall damaged foreign entries:\n%s", un.Data)
	}
}
