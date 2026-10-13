package opencode

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestContextLoaderProtocol(t *testing.T) {
	runContextLoaderScript(t, "context-loader.test.mjs")
}

func TestContextLoaderAncestorReplacementRace(t *testing.T) {
	runContextLoaderScript(t, "context-loader.race.test.mjs")
}

func runContextLoaderScript(t *testing.T, script string) {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal("Node is required for context loader verification")
	}
	root := t.TempDir()
	// The temporary native module captures the definition without starting OpenCode.
	for name, data := range map[string]string{
		"package.json": `{"type":"module"}`,
		"node_modules/@opencode/plugin/package.json": `{"type":"module","exports":"./index.js"}`,
		"node_modules/@opencode/plugin/index.js":     `export const Plugin = {define: value => value}`,
	} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	source, err := os.ReadFile("context-loader.js")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"content/skills/flow-plan/SKILL.md": []byte("# Plan\n")}
	expected := map[string]string{}
	for _, mapping := range contextReferences {
		text := "\ufeffFull canonical text: " + mapping.resource + "\r\nQuotes \"'` ${notExecutable} </script> backslash \\ Unicode 雪😀\u2028\u2029\n__HIVE_CONTEXT_SNAPSHOT_DATA__\x00\n"
		files["content/skills/"+mapping.resource] = []byte(text)
		expected[mapping.role] = text
		path := filepath.Join(root, "skills", mapping.resource)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	scriptData, metadata, err := RenderContextLoader(source, files)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"context-loader.js": scriptData, "context-loader.json": metadata, "fixture.json": fixture,
		"missing.js": bytes.Replace(source, []byte(contextSnapshotMarker), []byte(`{"schema":2,"references":[]}`), 1)} {
		if err := os.WriteFile(filepath.Join(root, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	files["content/skills/flow-plan/references/plan-review.md"] = []byte("Regenerated reference\n")
	reloaded, _, err := RenderContextLoader(source, files)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "reloaded.js"), reloaded, 0600); err != nil {
		t.Fatal(err)
	}
	test, err := filepath.Abs(script)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, test, root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("callback protocol: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}
