package management

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"tricell-hive/integrations/hostevent"
)

const hookTestExe = "/opt/hive/bin/hive"

// A Claude Code settings.json as other tools leave it: herdr, rex, codegraph
// and Engram hooks under the same events, plus settings and a secret in env.
// These are synthetic copies; no real configuration is read.
const thirdPartySettings = `{
  "env": {
    "ANTHROPIC_API_KEY": "sk-synthetic-not-a-real-key"
  },
  "permissions": {
    "allow": ["Bash(ls:*)"]
  },
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup|resume",
        "hooks": [
          { "type": "command", "command": "engram session-start", "timeout": 10 },
          { "type": "command", "command": "codegraph status --quiet" }
        ]
      },
      {
        "hooks": [
          { "type": "command", "command": "herdr agent-state working" }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "hooks": [
          { "type": "command", "command": "rex notify prompt" }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          { "type": "command", "command": "herdr agent-state idle" }
        ]
      }
    ]
  },
  "model": "opus"
}
`

func claudeSpec() hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word: "claude",
		Path: "/home/u/.claude/settings.json",
		Events: []string{
			"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop",
			"TaskCreated", "TaskCompleted", "SessionEnd",
		},
	}
}

func cursorSpec() hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word: "cursor", Path: "/home/u/.cursor/hooks.json", Flat: true, RootVersion: 1,
		Events: []string{"sessionStart", "beforeSubmitPrompt", "stop", "sessionEnd"},
	}
}

func mustEdit(t *testing.T, cur string, exists bool, f hostevent.JSONHookFile, prior *HostHooksRecord) hookEditResult {
	t.Helper()
	res, err := editJSONHooks([]byte(cur), exists, f, hookTestExe, prior)
	if err != nil {
		t.Fatalf("editJSONHooks: %v", err)
	}
	if res.Changed && !json.Valid(res.Data) {
		t.Fatalf("edit left invalid JSON:\n%s", res.Data)
	}
	return res
}

// rec turns an edit's delta into the record the state would hold.
func rec(f hostevent.JSONHookFile, delta HostHooksRecord) HostHooksRecord {
	return mergeHostHooks(nil, delta, f.Word)
}

func countHive(t *testing.T, data []byte, f hostevent.JSONHookFile) int {
	t.Helper()
	var doc struct {
		Hooks map[string][]map[string]any `json:"hooks"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	n := 0
	for event, entries := range doc.Hooks {
		for _, e := range entries {
			for _, c := range f.Commands(e) {
				if hostevent.IsHiveCommand(c, f.Word, event, nil) {
					n++
				}
			}
		}
	}
	return n
}

func TestInstallAddsEveryEventAndKeepsForeignHooksIntact(t *testing.T) {
	f := claudeSpec()
	res := mustEdit(t, thirdPartySettings, true, f, nil)
	if !res.Changed {
		t.Fatal("expected a change")
	}
	if got := countHive(t, res.Data, f); got != len(f.Events) {
		t.Fatalf("hive entries = %d, want %d\n%s", got, len(f.Events), res.Data)
	}
	// Every foreign byte survives: removing Hive's text restores the file exactly.
	un := undoJSONHooks(res.Data, f, rec(f, res.Delta))
	if un.Skipped != "" || string(un.Data) != thirdPartySettings {
		t.Fatalf("round trip differs (skipped %q):\n%s", un.Skipped, un.Data)
	}
	// And the secret and the settings are still there byte for byte.
	for _, keep := range []string{`"ANTHROPIC_API_KEY": "sk-synthetic-not-a-real-key"`, `"command": "herdr agent-state idle"`, `"model": "opus"`} {
		if !strings.Contains(string(res.Data), keep) {
			t.Fatalf("lost %q", keep)
		}
	}
	var doc map[string]any
	_ = json.Unmarshal(res.Data, &doc)
	hooks := doc["hooks"].(map[string]any)
	if len(hooks["SessionStart"].([]any)) != 3 || len(hooks["Stop"].([]any)) != 1 {
		t.Fatalf("event arrays: %v", hooks)
	}
}

func TestEntriesCarryAnExplicitTimeoutAndTheQuotedCommand(t *testing.T) {
	f := claudeSpec()
	res := mustEdit(t, thirdPartySettings, true, f, nil)
	var doc struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type    string `json:"type"`
				Command string `json:"command"`
				Timeout int    `json:"timeout"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(res.Data, &doc); err != nil {
		t.Fatal(err)
	}
	groups := doc.Hooks["SessionEnd"]
	if len(groups) != 1 || len(groups[0].Hooks) != 1 {
		t.Fatalf("SessionEnd: %+v", groups)
	}
	h := groups[0].Hooks[0]
	if h.Type != "command" || h.Command != "'/opt/hive/bin/hive' hook claude SessionEnd" || h.Timeout <= 2 {
		t.Fatalf("SessionEnd handler = %+v, want a quoted command and a timeout above the 1.5 s budget", h)
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	f := claudeSpec()
	first := mustEdit(t, thirdPartySettings, true, f, nil)
	second := mustEdit(t, string(first.Data), true, f, nil)
	if second.Changed {
		t.Fatalf("second install changed the file:\n%s", second.Data)
	}
	if got := countHive(t, second.Data, f); got != len(f.Events) {
		t.Fatalf("hive entries = %d", got)
	}
}

func TestMovedBinaryIsReplacedInPlaceNotDuplicated(t *testing.T) {
	f := claudeSpec()
	first := mustEdit(t, thirdPartySettings, true, f, nil)
	prior := rec(f, first.Delta)
	moved, err := editJSONHooks(first.Data, true, f, "/opt/hive-new/hive", &prior)
	if err != nil || !moved.Changed {
		t.Fatalf("moved binary: changed=%v err=%v", moved.Changed, err)
	}
	if strings.Contains(string(moved.Data), "/opt/hive/bin/hive") || !strings.Contains(string(moved.Data), "'/opt/hive-new/hive' hook claude SessionStart") {
		t.Fatalf("old command left or new command missing:\n%s", moved.Data)
	}
	if got := countHive(t, moved.Data, f); got != len(f.Events) {
		t.Fatalf("hive entries = %d, want %d", got, len(f.Events))
	}
	un := undoJSONHooks(moved.Data, f, mergeHostHooks(&prior, moved.Delta, "claude"))
	if string(un.Data) != thirdPartySettings {
		t.Fatalf("uninstall after a move differs:\n%s", un.Data)
	}
}

func TestRoundTripKeepsEveryFormattingStyle(t *testing.T) {
	cases := map[string]string{
		"tabs":         "{\n\t\"hooks\": {\n\t\t\"Stop\": [\n\t\t\t{ \"hooks\": [ { \"type\": \"command\", \"command\": \"x\" } ] }\n\t\t]\n\t}\n}\n",
		"four spaces":  "{\n    \"a\": 1,\n    \"hooks\": {\n        \"SessionStart\": []\n    }\n}\n",
		"compact":      `{"a":1,"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"x"}]}]}}`,
		"crlf":         "{\r\n  \"a\": 1,\r\n  \"hooks\": {\r\n    \"Stop\": []\r\n  }\r\n}\r\n",
		"no hooks":     "{\n  \"model\": \"opus\"\n}\n",
		"empty object": "{}",
		"no newline":   "{\n  \"model\": \"opus\"\n}",
		"empty hooks":  "{\n  \"hooks\": {}\n}\n",
		// Whitespace inside a container that is empty before Hive's entries
		// and empty again after they go is the user's, and stays.
		"empty array with a space": "{\n  \"hooks\": {\n    \"SessionStart\": [ ]\n  }\n}\n",
		"empty array over lines":   "{\n  \"hooks\": {\n    \"SessionStart\": [\n    ]\n  }\n}\n",
		"empty hooks with a space": "{\n  \"hooks\": { }\n}\n",
		"empty hooks over lines":   "{\n  \"hooks\": {\n  }\n}\n",
		"empty root with a space":  "{ }",
		"empty root over lines":    "{\n\n}\n",
		"trailing text":            "{\n  \"hooks\": {\n    \"SessionStart\": [\n      {\n        \"hooks\": []\n      }\n    ]\n  }\n}\n\n",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			f := claudeSpec()
			res := mustEdit(t, in, true, f, nil)
			if !res.Changed {
				t.Fatal("expected a change")
			}
			if got := countHive(t, res.Data, f); got != len(f.Events) {
				t.Fatalf("hive entries = %d\n%s", got, res.Data)
			}
			un := undoJSONHooks(res.Data, f, rec(f, res.Delta))
			if un.Skipped != "" || un.DeleteFile || string(un.Data) != in {
				t.Fatalf("round trip differs (skipped %q delete %v):\nwant %q\ngot  %q", un.Skipped, un.DeleteFile, in, un.Data)
			}
		})
	}
}

func TestMissingFileIsCreatedAndRemovedWhole(t *testing.T) {
	for _, f := range []hostevent.JSONHookFile{claudeSpec(), cursorSpec()} {
		res := mustEdit(t, "", false, f, nil)
		if !res.Changed || !res.Delta.CreatedFile {
			t.Fatalf("%s: changed=%v created=%v", f.Word, res.Changed, res.Delta.CreatedFile)
		}
		if f.RootVersion > 0 {
			var doc map[string]any
			_ = json.Unmarshal(res.Data, &doc)
			if doc["version"] != float64(1) {
				t.Fatalf("version = %v", doc["version"])
			}
			if !bytes.HasPrefix(res.Data, []byte("{\n  \"version\": 1,")) {
				t.Fatalf("version is not the first member:\n%s", res.Data)
			}
		}
		if !bytes.HasSuffix(res.Data, []byte("}\n")) {
			t.Fatalf("created file should end with a newline: %q", res.Data)
		}
		un := undoJSONHooks(res.Data, f, rec(f, res.Delta))
		if !un.DeleteFile {
			t.Fatalf("%s: a file Hive created must be deleted when its entries go:\n%s", f.Word, un.Data)
		}
	}
}

func TestCreatedFileIsKeptWhenSomethingElseJoinedIt(t *testing.T) {
	f := claudeSpec()
	res := mustEdit(t, "", false, f, nil)
	joined := strings.Replace(string(res.Data), "{\n", "{\n  \"model\": \"opus\",\n", 1)
	un := undoJSONHooks([]byte(joined), f, rec(f, res.Delta))
	if un.DeleteFile || un.Skipped != "" {
		t.Fatalf("a file with another setting must stay (delete %v, skipped %q)", un.DeleteFile, un.Skipped)
	}
	if strings.Contains(string(un.Data), "hook claude") || !strings.Contains(string(un.Data), `"model": "opus"`) {
		t.Fatalf("expected only Hive's entries gone:\n%s", un.Data)
	}
	if !json.Valid(un.Data) {
		t.Fatalf("invalid JSON:\n%s", un.Data)
	}
}

func TestCursorFlatEntriesAndForeignHooksSurvive(t *testing.T) {
	foreign := `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "herdr agent-state idle" }
    ],
    "afterFileEdit": [
      { "command": "./format.sh" }
    ]
  }
}
`
	f := cursorSpec()
	res := mustEdit(t, foreign, true, f, nil)
	var doc struct {
		Version int                         `json:"version"`
		Hooks   map[string][]map[string]any `json:"hooks"`
	}
	if err := json.Unmarshal(res.Data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Version != 1 || len(doc.Hooks["stop"]) != 2 || len(doc.Hooks["afterFileEdit"]) != 1 {
		t.Fatalf("doc = %+v", doc)
	}
	entry := doc.Hooks["sessionStart"][0]
	if entry["command"] != "'/opt/hive/bin/hive' hook cursor sessionStart" || entry["timeout"] != float64(hostevent.HookTimeoutSeconds) {
		t.Fatalf("entry = %v", entry)
	}
	if _, grouped := entry["hooks"]; grouped {
		t.Fatalf("Cursor entries are flat, got %v", entry)
	}
	un := undoJSONHooks(res.Data, f, rec(f, res.Delta))
	if string(un.Data) != foreign {
		t.Fatalf("round trip differs:\n%s", un.Data)
	}
}

func TestForeignLookalikesAreNeverTouched(t *testing.T) {
	f := claudeSpec()
	in := `{
  "hooks": {
    "SessionStart": [
      { "hooks": [ { "type": "command", "command": "'/usr/local/bin/other' hook claude SessionStart" } ] },
      { "hooks": [ { "type": "command", "command": "hive-wrapper hook claude SessionStart" } ] },
      { "hooks": [ { "type": "command", "command": "'/opt/hivemind/bin/hivectl' hook claude SessionStart" } ] },
      { "hooks": [ { "type": "command", "command": "'/opt/hive/bin/hive-tool' hook claude SessionStart" } ] },
      { "hooks": [ { "type": "command", "command": "'/opt/hive/bin/hive' hook claude SessionStart" }, { "type": "command", "command": "echo shared group" } ] }
    ]
  }
}
`
	res := mustEdit(t, in, true, f, nil)
	un := undoJSONHooks(res.Data, f, rec(f, res.Delta))
	if string(un.Data) != in {
		t.Fatalf("lookalikes were touched:\n%s", un.Data)
	}
}

func TestRefusedFilesAreNotRewritten(t *testing.T) {
	f := claudeSpec()
	cases := map[string]string{
		"invalid JSON":           `{"hooks": `,
		"comments":               "{\n // note\n \"hooks\": {}\n}\n",
		"array at the top":       `[]`,
		"hooks is a string":      `{"hooks": "x"}`,
		"hooks is an array":      `{"hooks": []}`,
		"event is an object":     `{"hooks": {"SessionStart": {}}}`,
		"duplicate hooks":        `{"hooks": {}, "hooks": {}}`,
		"duplicate event":        `{"hooks": {"SessionStart": [], "SessionStart": []}}`,
		"top-level scalar":       `"x"`,
		"unterminated top-level": `{`,
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			res, err := editJSONHooks([]byte(in), true, f, hookTestExe, nil)
			var refusal *hookRefusal
			if err == nil || !asRefusal(err, &refusal) {
				t.Fatalf("want a refusal, got err=%v changed=%v", err, res.Changed)
			}
			if string(res.Data) != in || res.Changed {
				t.Fatalf("a refused file must come back untouched")
			}
		})
	}
}

func asRefusal(err error, target **hookRefusal) bool {
	r, ok := err.(*hookRefusal)
	if ok {
		*target = r
	}
	return ok
}

func TestUndoSkipsAFileItCannotReadWithoutError(t *testing.T) {
	f := claudeSpec()
	res := undoJSONHooks([]byte(`{"hooks": `), f, HostHooksRecord{Commands: []string{"'/opt/hive/bin/hive' hook claude Stop"}})
	if res.Skipped == "" || string(res.Data) != `{"hooks": ` || res.DeleteFile {
		t.Fatalf("res = %+v", res)
	}
}

func TestUndoRemovesDuplicateAndStaleHiveEntries(t *testing.T) {
	f := claudeSpec()
	res := mustEdit(t, thirdPartySettings, true, f, nil)
	// A second Hive entry for one event, as a hand edit or an interrupted run could leave.
	doubled := strings.Replace(string(res.Data), `"SessionEnd": [`, `"SessionEnd": [
      { "hooks": [ { "type": "command", "command": "'/old/place/hive' hook claude SessionEnd", "timeout": 5 } ] },`, 1)
	un := undoJSONHooks([]byte(doubled), f, rec(f, res.Delta))
	if got := countHive(t, un.Data, f); got != 0 {
		t.Fatalf("hive entries left = %d\n%s", got, un.Data)
	}
	if !json.Valid(un.Data) {
		t.Fatalf("invalid JSON:\n%s", un.Data)
	}
}
