package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/opencode"
	"tricell-hive/integrations/pi"
)

// Tests of the six hook registrations on a temporary home that already holds
// other tools' hooks (synthetic copies of the shapes herdr, rex, codegraph and
// Engram leave). The real configuration files are never read or written.

var allSixHosts = []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}

const (
	foreignCodexHooks = `{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          { "type": "command", "command": "engram codex session-start", "timeout": 10 }
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
  }
}
`
	foreignCursorHooks = `{
  "version": 1,
  "hooks": {
    "stop": [
      { "command": "herdr agent-state idle" }
    ],
    "afterFileEdit": [
      { "command": "codegraph sync" }
    ]
  }
}
`
	foreignGrokHooks = `{
  "hooks": {
    "SessionStart": [
      { "hooks": [ { "type": "command", "command": "engram grok session-start" } ] }
    ]
  }
}
`
	foreignOpenCodePlugin = "// installed by herdr\nexport default { id: \"herdr\", async setup() {} }\n"
	foreignPiExtension    = "// installed by herdr\nexport default function () {}\n"
)

type hookPaths struct {
	claude, codex, cursor, grokEvents, grokForeign, opencode, opencodeForeign, pi, piForeign string
}

func pathsFor(o Options) hookPaths {
	h := o.Home
	return hookPaths{
		claude:          filepath.Join(h, ".claude", "settings.json"),
		codex:           filepath.Join(h, ".codex", "hooks.json"),
		cursor:          filepath.Join(h, ".cursor", "hooks.json"),
		grokEvents:      filepath.Join(h, ".grok", "hooks", "hive-session-events.json"),
		grokForeign:     filepath.Join(h, ".grok", "hooks", "engram.json"),
		opencode:        filepath.Join(h, ".config", "opencode", "plugins", "hive-events.js"),
		opencodeForeign: filepath.Join(h, ".config", "opencode", "plugins", "gk-hooks.js"),
		pi:              filepath.Join(h, ".pi", "agent", "extensions", "hive-events.ts"),
		piForeign:       filepath.Join(h, ".pi", "agent", "extensions", "herdr-agent-state.ts"),
	}
}

// hooksSetup builds a home that already has third-party hooks for every host.
func hooksSetup(t *testing.T) (Options, hookPaths, map[string]string) {
	t.Helper()
	o := setup(t)
	o.Hosts = allSixHosts
	hp := pathsFor(o)
	originals := map[string]string{
		hp.claude:          thirdPartySettings,
		hp.codex:           foreignCodexHooks,
		hp.cursor:          foreignCursorHooks,
		hp.grokForeign:     foreignGrokHooks,
		hp.opencodeForeign: foreignOpenCodePlugin,
		hp.piForeign:       foreignPiExtension,
	}
	for path, body := range originals {
		put(t, path, body)
	}
	return o, hp, originals
}

func assertForeignIntact(t *testing.T, originals map[string]string) {
	t.Helper()
	for path, want := range originals {
		if got := get(t, path); got != want {
			t.Fatalf("%s changed:\n%s", path, got)
		}
	}
}

func hiveCommands(t *testing.T, path string, wantWord string) map[string]string {
	t.Helper()
	var doc struct {
		Hooks map[string][]map[string]any `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(get(t, path)), &doc); err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	out := map[string]string{}
	for event, entries := range doc.Hooks {
		for _, e := range entries {
			cmd, _ := e["command"].(string)
			if cmd == "" {
				if hs, ok := e["hooks"].([]any); ok && len(hs) == 1 {
					cmd, _ = hs[0].(map[string]any)["command"].(string)
				}
			}
			if strings.Contains(cmd, " hook "+wantWord+" ") {
				out[event] = cmd
			}
		}
	}
	return out
}

func TestInstallRegistersAllSixHostsAndUninstallLeavesForeignHooksByteIdentical(t *testing.T) {
	o, hp, originals := hooksSetup(t)
	exe := suiteExecutable(t)
	p := plan(t, "install", o)
	if p.Executable != exe || len(p.HostHooks) != 3 {
		t.Fatalf("plan: executable %q, %d host hook steps", p.Executable, len(p.HostHooks))
	}
	notes, err := HostHookNotes(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := apply(t, p); strings.Contains(got, ";") || len(got) != 32 {
		t.Fatalf("Apply's result must stay a transaction ID, got %q", got)
	}
	if paths := HostHookPaths(p); len(paths) != 3 || paths[0] != hp.claude || paths[1] != hp.codex || paths[2] != hp.cursor {
		t.Fatalf("host hook paths = %v", paths)
	}

	// Claude Code: seven events, explicit timeouts, the quoted command.
	claude := hiveCommands(t, hp.claude, "claude")
	if len(claude) != 7 {
		t.Fatalf("claude events = %v", claude)
	}
	if want := "'" + exe + "' hook claude SessionEnd"; claude["SessionEnd"] != want {
		t.Fatalf("SessionEnd command = %q, want %q", claude["SessionEnd"], want)
	}
	if !strings.Contains(get(t, hp.claude), `"timeout": 5`) {
		t.Fatal("entries need an explicit timeout")
	}
	// Codex: exactly four events, never Stop or SubagentStop, no trust entry.
	codex := hiveCommands(t, hp.codex, "codex")
	if len(codex) != 4 || codex["Stop"] != "" || codex["SubagentStop"] != "" {
		t.Fatalf("codex events = %v", codex)
	}
	for _, path := range []string{hp.codex, filepath.Join(o.Home, ".codex", "config.toml")} {
		if b, err := os.ReadFile(path); err == nil && strings.Contains(string(b), "trusted_hash") {
			t.Fatalf("%s holds a trust entry Hive must never write", path)
		}
	}
	// Cursor: flat entries and its four events.
	cursor := hiveCommands(t, hp.cursor, "cursor")
	if len(cursor) != 4 || cursor["beforeSubmitPrompt"] == "" || cursor["sessionEnd"] == "" {
		t.Fatalf("cursor events = %v", cursor)
	}
	// Grok: a second native file, program and arguments, no quoted path.
	var gk struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Type, Command string
				Args          []string
				Timeout       int
			}
		}
	}
	if err := json.Unmarshal([]byte(get(t, hp.grokEvents)), &gk); err != nil {
		t.Fatal(err)
	}
	h := gk.Hooks["SessionEnd"][0].Hooks[0]
	if h.Command != exe || strings.Join(h.Args, " ") != "hook grok SessionEnd" || h.Timeout == 0 || strings.Contains(h.Command, "'") {
		t.Fatalf("grok handler = %+v", h)
	}
	// OpenCode and Pi: own files that name the binary.
	for _, path := range []string{hp.opencode, hp.pi} {
		if !strings.Contains(get(t, path), `"`+exe+`"`) || !strings.Contains(get(t, path), "hive-managed") {
			t.Fatalf("%s does not carry the binary or the marker", path)
		}
	}
	if !strings.Contains(get(t, hp.opencode), `id: "hive.session-events"`) || !strings.Contains(get(t, hp.opencode), "async setup(ctx)") {
		t.Fatal("the OpenCode plugin must use the v2 {id, setup(ctx)} shape")
	}
	for _, name := range []string{"session_start", "agent_settled", "session_shutdown"} {
		if !strings.Contains(get(t, hp.pi), `"`+name+`"`) {
			t.Fatalf("the Pi extension does not listen to %s", name)
		}
	}
	// The OpenCode context loader stays inactive: no opencode.json is written.
	absent(t, filepath.Join(o.Home, ".config", "opencode", "opencode.json"))
	absent(t, filepath.Join(o.Home, ".config", "opencode", "opencode.jsonc"))

	// The report tells the user to approve Codex's hook and flags Cursor; an
	// update that finds the entries current says nothing.
	if joined := strings.Join(notes, "\n"); len(notes) != 2 || !strings.Contains(joined, "/hooks") || !strings.Contains(joined, "Cursor CLI") {
		t.Fatalf("notes = %q", notes)
	}
	if again, err := HostHookNotes(plan(t, "install", o)); err != nil || len(again) != 0 {
		t.Fatalf("notes after the entries are current = %q, %v", again, err)
	}

	// Backups of the JSON files, outside the other tools' bytes.
	for path, body := range map[string]string{hp.claude: thirdPartySettings, hp.codex: foreignCodexHooks, hp.cursor: foreignCursorHooks} {
		if get(t, path+".hive-backup") != body {
			t.Fatalf("backup of %s differs from the original", path)
		}
	}
	// The foreign plugin, extension and Grok file are not touched.
	for _, path := range []string{hp.grokForeign, hp.opencodeForeign, hp.piForeign} {
		if get(t, path) != originals[path] {
			t.Fatalf("%s changed", path)
		}
	}

	// Install again: nothing to do.
	if got := apply(t, plan(t, "install", o)); got != "unchanged" {
		t.Fatalf("second install = %q", got)
	}

	// Uninstall removes only Hive's entries and files.
	apply(t, plan(t, "remove", o))
	assertForeignIntact(t, originals)
	for _, path := range []string{hp.grokEvents, hp.opencode, hp.pi} {
		absent(t, path)
	}
	// The backups outlive the uninstall (still private); a later install
	// overwrites them with the file as it is then.
	for path, body := range map[string]string{hp.claude: thirdPartySettings, hp.codex: foreignCodexHooks, hp.cursor: foreignCursorHooks} {
		if get(t, path+".hive-backup") != body || wfMode(t, path+".hive-backup") != 0600 {
			t.Fatalf("the backup of %s was not kept after the uninstall", path)
		}
	}
	put(t, hp.claude, strings.Replace(get(t, hp.claude), `"model": "opus"`, `"model": "sonnet"`, 1))
	apply(t, plan(t, "install", o))
	if want := strings.Replace(thirdPartySettings, `"model": "opus"`, `"model": "sonnet"`, 1); get(t, hp.claude+".hive-backup") != want {
		t.Fatal("a later install must overwrite the kept backup")
	}
	apply(t, plan(t, "remove", o))
	if s := stateFor(t, o); len(s.HostHooks) != 0 || len(s.Records) != 0 {
		t.Fatalf("state after uninstall: %d host hook records, %d records", len(s.HostHooks), len(s.Records))
	}
}

func TestFilesHiveCreatesAreRemovedAndTheirDirectoriesToo(t *testing.T) {
	o := setup(t)
	o.Hosts = allSixHosts
	hp := pathsFor(o)
	apply(t, plan(t, "install", o))
	for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s was not written: %v", path, err)
		}
	}
	if !strings.HasPrefix(get(t, hp.cursor), "{\n  \"version\": 1,") {
		t.Fatalf("a Cursor file Hive creates needs its version: %s", get(t, hp.cursor))
	}
	if m := wfMode(t, hp.claude); m != 0600 {
		t.Fatalf("a created settings.json should be private, got %v", m)
	}
	for _, path := range []string{hp.claude, hp.codex, hp.cursor} {
		absent(t, path+".hive-backup") // nothing existed to back up
	}
	apply(t, plan(t, "remove", o))
	for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
		absent(t, path)
	}
}

func TestUpdateRewritesEveryEntryWhenTheBinaryMoves(t *testing.T) {
	o, hp, originals := hooksSetup(t)
	apply(t, plan(t, "install", o))
	moved := copyExecutable(t, "hive-test-executable-moved")
	withHiveExecutable(t, moved)
	p := plan(t, "install", o)
	if p.Executable != moved {
		t.Fatalf("plan executable = %q", p.Executable)
	}
	apply(t, p)
	for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
		body := get(t, path)
		if !strings.Contains(body, moved) || strings.Contains(body, "hive-test-executable\"") || strings.Contains(body, "hive-test-executable' ") {
			t.Fatalf("%s still names the old binary:\n%s", path, body)
		}
	}
	if n := len(hiveCommands(t, hp.claude, "claude")); n != 7 {
		t.Fatalf("claude entries after the move = %d (duplicated?)", n)
	}
	apply(t, plan(t, "remove", o))
	assertForeignIntact(t, originals)
}

func TestRemoveOneHostLeavesTheOthersRegistered(t *testing.T) {
	o, hp, originals := hooksSetup(t)
	apply(t, plan(t, "install", o))
	only := o
	only.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", only))
	if get(t, hp.claude) != originals[hp.claude] {
		t.Fatalf("claude settings not restored:\n%s", get(t, hp.claude))
	}
	if len(hiveCommands(t, hp.codex, "codex")) != 4 || len(hiveCommands(t, hp.cursor, "cursor")) != 4 {
		t.Fatal("removing Claude Code touched the Codex or Cursor entries")
	}
	if _, err := os.Stat(hp.pi); err != nil {
		t.Fatal("removing Claude Code removed Pi's extension")
	}
	if _, ok := stateFor(t, o).HostHooks[hp.claude]; ok {
		t.Fatal("the Claude Code record survived")
	}
}

func TestCollisionsFailClosedAndWriteNothing(t *testing.T) {
	cases := map[string]func(t *testing.T, hp hookPaths){
		"unowned grok file":      func(t *testing.T, hp hookPaths) { put(t, hp.grokEvents, "{}\n") },
		"unowned opencode file":  func(t *testing.T, hp hookPaths) { put(t, hp.opencode, "// mine\n") },
		"unowned pi file":        func(t *testing.T, hp hookPaths) { put(t, hp.pi, "// mine\n") },
		"codex hooks not object": func(t *testing.T, hp hookPaths) { put(t, hp.codex, `{"hooks": []}`) },
		// A shared JSON hook file Hive's editor cannot read (Cursor's here) is
		// skipped, not a collision: see host_hooks_skip_test.go.
	}
	for name, prepare := range cases {
		t.Run(name, func(t *testing.T) {
			o, hp, _ := hooksSetup(t)
			prepare(t, hp)
			before := map[string]string{}
			for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
				if b, err := os.ReadFile(path); err == nil {
					before[path] = string(b)
				}
			}
			p, err := BuildPlan("install", o)
			if err == nil {
				_, err = (Engine{}).Apply(p)
			}
			if err == nil {
				t.Fatal("install over a colliding file succeeded")
			}
			for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
				b, readErr := os.ReadFile(path)
				want, existed := before[path]
				if (readErr == nil) != existed || (existed && string(b) != want) {
					t.Fatalf("%s changed by a failed install", path)
				}
			}
			absent(t, filepath.Join(o.StateDir, "pending.json"))
		})
	}
}

func TestEveryNewTargetUsesTheExecutableCheck(t *testing.T) {
	suite := suiteExecutable(t)
	for _, host := range []string{"claude", "codex", "cursor", "opencode", "pi"} {
		t.Run(host, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{host}
			temp := filepath.Join(t.TempDir(), "hive")
			if err := os.WriteFile(temp, []byte("#!/bin/sh\n"), 0o755); err != nil {
				t.Fatal(err)
			}
			withHiveExecutable(t, temp)
			if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "go build") {
				t.Fatalf("temporary binary = %v, want the go build message", err)
			}
			// A plan edited by hand cannot point the hooks at an unusable path.
			withHiveExecutable(t, suite)
			p := plan(t, "install", o)
			p.Executable = "relative/hive"
			p.ID = planID(p)
			if _, err := (Engine{}).Apply(p); err == nil {
				t.Fatal("apply accepted a relative executable")
			}
		})
	}
}

func TestPlanValidationRejectsTamperedHostHooksSteps(t *testing.T) {
	o, _, _ := hooksSetup(t)
	good := plan(t, "install", o)
	mutations := map[string]func(*Plan){
		"path":    func(p *Plan) { p.HostHooks[0].Path = filepath.Join(o.Home, "elsewhere.json") },
		"host":    func(p *Plan) { p.HostHooks[0].Host = "pi" },
		"dropped": func(p *Plan) { p.HostHooks = p.HostHooks[1:] },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			p := good
			p.HostHooks = append([]HostHooksStep(nil), good.HostHooks...)
			mutate(&p)
			p.ID = planID(p)
			if _, err := (Engine{}).Apply(p); err == nil {
				t.Fatal("a tampered plan was applied")
			}
		})
	}
	rp := plan(t, "remove", o)
	rp.HostHooks = good.HostHooks
	rp.ID = planID(rp)
	if _, err := (Engine{}).Apply(rp); err == nil {
		t.Fatal("a remove plan with host hook steps was applied")
	}
}

func TestPlanUnchangedAfterInstall(t *testing.T) {
	o, _, _ := hooksSetup(t)
	apply(t, plan(t, "install", o))
	unchanged, err := PlanUnchanged(plan(t, "install", o))
	if err != nil || !unchanged {
		t.Fatalf("unchanged = %v, err = %v", unchanged, err)
	}
	// A hand edit that drops one entry makes the update write again.
	hp := pathsFor(o)
	edited := strings.Replace(get(t, hp.codex), "SubagentStart", "SubagentStartX", 1)
	put(t, hp.codex, edited)
	unchanged, err = PlanUnchanged(plan(t, "install", o))
	if err != nil || unchanged {
		t.Fatalf("after a hand edit unchanged = %v, err = %v", unchanged, err)
	}
}

func TestNothingPersistedNextToTheStateHoldsTheJSONFilesBytes(t *testing.T) {
	o, _, _ := hooksSetup(t)
	apply(t, plan(t, "install", o))
	apply(t, plan(t, "remove", o))
	filepath.WalkDir(o.StateDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if b, _ := os.ReadFile(path); strings.Contains(string(b), "sk-synthetic-not-a-real-key") {
				t.Errorf("%s holds settings.json bytes", path)
			}
		}
		return nil
	})
}

func hookFail(stage string) Engine {
	return Engine{failpoint: func(s string) error {
		if s == stage {
			return errors.New("injected")
		}
		return nil
	}}
}

func TestRecoverUndoesAnInterruptedInstallByEntries(t *testing.T) {
	for _, stage := range []string{"host-hooks-backup", "host-hooks-write", "host-hooks-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o, hp, originals := hooksSetup(t)
			if _, err := hookFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if stage != "host-hooks-backup" {
				// Claude Code changes another setting before recovery.
				put(t, hp.claude, strings.Replace(get(t, hp.claude), `"model": "opus"`, `"model": "sonnet"`, 1))
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			want := originals[hp.claude]
			if stage != "host-hooks-backup" {
				want = strings.Replace(want, `"model": "opus"`, `"model": "sonnet"`, 1)
			}
			if got := get(t, hp.claude); got != want {
				t.Fatalf("settings.json after recovery:\n%s", got)
			}
			for _, path := range []string{hp.codex, hp.cursor} {
				if got := get(t, path); got != originals[path] {
					t.Fatalf("%s after recovery:\n%s", path, got)
				}
			}
			for _, path := range []string{hp.claude, hp.codex, hp.cursor} {
				absent(t, path+".hive-backup")
			}
			absent(t, filepath.Join(o.StateDir, "pending.json"))
			if len(stateFor(t, o).HostHooks) != 0 {
				t.Fatal("a host hooks record survived a recovery")
			}
		})
	}
}

func TestRecoverOfAnInterruptedInstallDeletesFilesHiveCreated(t *testing.T) {
	o := setup(t)
	o.Hosts = allSixHosts
	hp := pathsFor(o)
	if _, err := hookFail("host-hooks-after").Apply(plan(t, "install", o)); err == nil {
		t.Fatal("expected the failpoint")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{hp.claude, hp.codex, hp.cursor, hp.grokEvents, hp.opencode, hp.pi} {
		absent(t, path)
	}
}

func TestInterruptedRemoveIsRecoverableAndRerunnable(t *testing.T) {
	o, hp, originals := hooksSetup(t)
	apply(t, plan(t, "install", o))
	if _, err := hookFail("host-hooks-write").Apply(plan(t, "remove", o)); err == nil {
		t.Fatal("expected the failpoint")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "remove", o))
	assertForeignIntact(t, originals)
	if get(t, hp.claude+".hive-backup") != thirdPartySettings {
		t.Fatal("the backup must outlive the uninstall")
	}
	if len(stateFor(t, o).HostHooks) != 0 {
		t.Fatal("record left after the second remove")
	}
}

func TestRemoveLeavesAFileItCannotReadAndSaysSo(t *testing.T) {
	o, hp, _ := hooksSetup(t)
	apply(t, plan(t, "install", o))
	put(t, hp.claude, `{"hooks": `)
	result := apply(t, plan(t, "remove", o))
	if get(t, hp.claude) != `{"hooks": ` {
		t.Fatalf("an unreadable settings.json was rewritten: %q", get(t, hp.claude))
	}
	if !strings.Contains(result, "left "+hp.claude) {
		t.Fatalf("report does not name the file left: %q", result)
	}
	if get(t, hp.claude+".hive-backup") != thirdPartySettings {
		t.Fatal("the backup must be kept when the file could not be cleaned")
	}
}

func TestStateRejectsAHandEditedHostHooksRecord(t *testing.T) {
	o, hp, _ := hooksSetup(t)
	apply(t, plan(t, "install", o))
	for name, mutate := range map[string]func(*HostHooksRecord){
		"host": func(r *HostHooksRecord) { r.Host = "grok" },
		"path": func(r *HostHooksRecord) { r.Path = "relative.json" },
	} {
		t.Run(name, func(t *testing.T) {
			st := stateFor(t, o)
			r := st.HostHooks[hp.claude]
			mutate(&r)
			st.HostHooks[hp.claude] = r
			if _, _, err := readStateFrom(t, st); err == nil {
				t.Fatal("a hand-edited record was accepted")
			}
		})
	}
}

// readStateFrom writes st to a scratch directory and reads it back.
func readStateFrom(t *testing.T, st State) (State, string, error) {
	t.Helper()
	dir := canonicalTemp(t)
	if err := writeJSON(filepath.Join(dir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	return readState(dir)
}

func TestGeneratedHooksValidateTheirRecordedShape(t *testing.T) {
	for source, name := range map[string]string{
		grok.SessionEventsSource:   grok.SessionEventsName,
		opencode.EventPluginSource: opencode.EventPluginName,
		pi.EventExtensionSource:    pi.EventExtensionName,
	} {
		if !validSource(source) || !generatedSource(source) {
			t.Fatalf("%s is not a generated source", source)
		}
		if validRecordedHook(source, filepath.Join("/x", "wrongdir", name)) {
			t.Fatalf("%s accepted in the wrong folder", source)
		}
	}
	// A release cannot carry a generated payload.
	rel := Release{Files: []Payload{{Path: GlobalSource, Data: []byte("x")}, {Path: pi.EventExtensionSource, Data: []byte("x")}}}
	rel.ID = releaseID(rel)
	if err := validateRelease(rel); err == nil {
		t.Fatal("a release carrying a generated hook payload was accepted")
	}
}
