package management

import (
	"encoding/json"
	"strings"
	"testing"

	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
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

func TestCodexSessionEndTimeoutOnInstallAndUpdate(t *testing.T) {
	spec := codex.SessionEventsFile(target.Config{CodexHome: t.TempDir()})
	foreign := `{ "type": "command", "command": "other-session-end", "timeout": 20 }`
	initial := `{"hooks":{"SessionEnd":[{"hooks":[` + foreign + `]}]}}`
	fresh := mustEdit(t, initial, true, spec, nil)
	check := func(data []byte) {
		t.Helper()
		var doc struct {
			Hooks map[string][]struct {
				Hooks []struct {
					Command string `json:"command"`
					Timeout int    `json:"timeout"`
				} `json:"hooks"`
			} `json:"hooks"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatal(err)
		}
		for _, event := range spec.Events {
			count := 0
			for _, group := range doc.Hooks[event] {
				for _, h := range group.Hooks {
					if h.Command != hostevent.Command(hookTestExe, "codex", event) {
						continue
					}
					count++
					want := 5
					if event == "SessionEnd" {
						want = 3
					}
					if h.Timeout != want {
						t.Errorf("%s timeout = %d, want %d", event, h.Timeout, want)
					}
				}
			}
			if count != 1 {
				t.Errorf("%s Hive entries = %d, want 1", event, count)
			}
		}
		if !strings.Contains(string(data), foreign) {
			t.Fatal("foreign hook changed")
		}
	}
	check(fresh.Data)
	legacy := strings.ReplaceAll(string(fresh.Data), `"timeout":3`, `"timeout":5`)
	legacy = strings.ReplaceAll(legacy, `"timeout": 3`, `"timeout": 5`)
	updated := mustEdit(t, legacy, true, spec, ptrRec(rec(spec, fresh.Delta)))
	if !updated.Changed {
		t.Error("legacy timeout must be updated")
	}
	check(updated.Data)
	again := mustEdit(t, string(updated.Data), true, spec, ptrRec(rec(spec, updated.Delta)))
	if again.Changed {
		t.Fatal("repeated update must be a no-op")
	}
}

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
