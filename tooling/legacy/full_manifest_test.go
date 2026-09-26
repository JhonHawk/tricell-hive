package legacy

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

// TestFullCatalogRecognizesHistoricManifests exercises the exact inventory
// captured from Hive 16e7d335.  It intentionally constructs all three
// historical manifest forms instead of relying on a partial test fixture.
func TestFullCatalogRecognizesHistoricManifests(t *testing.T) {
	c := fixture(t)
	c.PiHome = filepath.Join(c.Home, "Pi user root")
	cat := loadCatalog()
	paths := map[string]bool{}
	for _, f := range cat.Files {
		root := fullRoot(c, f.Root)
		path := filepath.Join(root, filepath.FromSlash(f.Path))
		put(t, path, render(f, root))
		paths[path] = true
	}

	putFullClaudeManifest(t, c, cat)
	putFullJSONManifests(t, c, cat)
	putFullUserSettings(t, c, cat)

	r, err := Scan(c, append([]string(nil), allHosts...))
	if err != nil {
		t.Fatal(err)
	}
	if !r.Detected || strings.Join(r.Hosts, ",") != strings.Join(allHosts, ",") {
		t.Fatalf("unexpected detection: %+v", r)
	}
	edits := map[string]Edit{}
	for _, edit := range r.Edits {
		edits[edit.Path] = edit
	}
	for path := range paths {
		edit, ok := edits[path]
		if !ok || !edit.Delete {
			t.Fatalf("catalogue file was not retired: %s (%+v)", path, edit)
		}
	}
	for _, path := range []string{
		filepath.Join(c.ClaudeHome, ".deploy-manifest"),
		filepath.Join(c.Home, ".agents", ".hive-deploy-manifest.json"),
		filepath.Join(c.PiHome, ".hive-deploy-manifest.json"),
	} {
		if edit, ok := edits[path]; !ok || !edit.Delete {
			t.Fatalf("manifest was not retired: %s (%+v)", path, edit)
		}
	}

	assertJSONHookPreserved(t, edits[filepath.Join(c.ClaudeHome, "settings.json")].After, "claude-user")
	assertJSONHookPreserved(t, edits[filepath.Join(c.CodexHome, "hooks.json")].After, "codex-user")
	piSettings := edits[filepath.Join(c.PiHome, "extensions", "subagent", "config.json")].After
	if bytes.Contains(piSettings, []byte("forceTopLevelAsync")) || !bytes.Contains(piSettings, []byte("dark")) {
		t.Fatalf("Pi preferences were not preserved: %s", piSettings)
	}
}

func fullRoot(c target.Config, root string) string {
	switch root {
	case "claude":
		return c.ClaudeHome
	case "codex":
		return c.CodexHome
	case "grok":
		return c.GrokHome
	case "pi":
		return c.PiHome
	case "opencode":
		return c.OpenCodeHome
	case "shared":
		return filepath.Join(c.Home, ".agents")
	default:
		panic("unknown fixture root: " + root)
	}
}

func putFullClaudeManifest(t *testing.T, c target.Config, cat catalog) {
	t.Helper()
	entries := map[string]bool{}
	for _, f := range cat.Files {
		entries[f.Manifest] = true
	}
	var ordered []string
	for entry := range entries {
		ordered = append(ordered, entry)
	}
	sort.Strings(ordered)
	put(t, filepath.Join(c.ClaudeHome, ".deploy-manifest"), []byte("# source_commit: "+cat.Revision+"\n"+strings.Join(ordered, "\n")+"\n"))
}

func putFullJSONManifests(t *testing.T, c target.Config, cat catalog) {
	t.Helper()
	for _, root := range []string{"shared", "pi"} {
		managed := map[string]map[string]string{}
		fixtureRoot := fullRoot(c, root)
		for _, f := range cat.Files {
			if f.Root != root {
				continue
			}
			managed[f.Path] = map[string]string{
				"sha256": digest(render(f, fixtureRoot)),
				"source": f.Source,
				"kind":   "file",
			}
		}
		manifest := map[string]any{
			"schemaVersion": 1,
			"scope":         root,
			"managedFiles":  managed,
			"managedConfig": map[string]any{},
		}
		if root == "shared" {
			manifest["scope"] = "shared-skills"
		} else {
			manifest["managedConfig"] = map[string]any{
				"extensions/subagent/config.json": map[string]any{"forceTopLevelAsync": true},
			}
		}
		body, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		put(t, filepath.Join(fixtureRoot, ".hive-deploy-manifest.json"), body)
	}
}

func putFullUserSettings(t *testing.T, c target.Config, cat catalog) {
	t.Helper()
	known := map[string]map[string][]map[string]string{"claude": {}, "codex": {}}
	for _, hook := range cat.Hooks {
		known[hook.Host][hook.Event] = append(known[hook.Host][hook.Event], map[string]string{"type": "command", "command": hook.Command})
	}
	for host, name := range map[string]string{"claude": "settings.json", "codex": "hooks.json"} {
		events := map[string][]map[string]any{}
		for event, hooks := range known[host] {
			entries := make([]map[string]any, 0, len(hooks))
			for _, hook := range hooks {
				entries = append(entries, map[string]any{"hooks": []map[string]string{hook}})
			}
			events[event] = entries
		}
		user := host + "-user"
		events["UserPromptSubmit"] = append(events["UserPromptSubmit"], map[string]any{"hooks": []map[string]string{{"type": "command", "command": user}}})
		body, err := json.Marshal(map[string]any{"preference": user, "hooks": events})
		if err != nil {
			t.Fatal(err)
		}
		root := c.ClaudeHome
		if host == "codex" {
			root = c.CodexHome
		}
		put(t, filepath.Join(root, name), body)
	}
	put(t, filepath.Join(c.PiHome, "extensions", "subagent", "config.json"), []byte(`{"forceTopLevelAsync":true,"theme":"dark"}`))
}

func assertJSONHookPreserved(t *testing.T, data []byte, want string) {
	t.Helper()
	if !bytes.Contains(data, []byte(want)) {
		t.Fatalf("user hook was removed: %s", data)
	}
	for _, hook := range loadCatalog().Hooks {
		if bytes.Contains(data, []byte(hook.Command)) {
			t.Fatalf("legacy hook remained: %s", hook.Command)
		}
	}
}
