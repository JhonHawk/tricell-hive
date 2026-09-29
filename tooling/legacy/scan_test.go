package legacy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

func fixture(t *testing.T) target.Config {
	t.Helper()
	h, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return target.Config{Scope: "user", Home: h, ClaudeHome: filepath.Join(h, ".claude"), CodexHome: filepath.Join(h, ".codex"), PiHome: filepath.Join(h, ".pi", "agent"), OpenCodeHome: filepath.Join(h, ".config", "opencode"), GrokHome: filepath.Join(h, ".grok")}
}
func put(t *testing.T, p string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
}
func TestEmptyIsReadOnly(t *testing.T) {
	c := fixture(t)
	r, e := Scan(c, []string{"codex"})
	if e != nil || r.Detected || len(r.Edits) != 0 {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestExactAndEditedLegacy(t *testing.T) {
	c := fixture(t)
	cat := loadCatalog()
	var b []byte
	for _, f := range cat.Files {
		if f.Root == "codex" && f.Path == "AGENTS.md" {
			b = f.Data
			break
		}
	}
	if len(b) == 0 {
		t.Fatal("catalog core absent")
	}
	p := filepath.Join(c.CodexHome, "AGENTS.md")
	put(t, p, b)
	r, e := Scan(c, []string{"codex"})
	if e != nil || len(r.Edits) != 1 || !r.Edits[0].Delete {
		t.Fatalf("%+v %v", r, e)
	}
	put(t, p, append(b, []byte("\nuser edit")...))
	if _, e = Scan(c, []string{"codex"}); e == nil {
		t.Fatal("edited legacy accepted")
	}
}

// TestModifiedFileErrorCarriesThePath: a file that differs from what Hive
// expects at a path it once used fails the scan with a *ModifiedFileError, so
// a caller finds the path with errors.As instead of parsing the message. The
// message does not call the file legacy, since it may be the user's own.
func TestModifiedFileErrorCarriesThePath(t *testing.T) {
	c := fixture(t)
	var b []byte
	for _, f := range loadCatalog().Files {
		if f.Root == "codex" && f.Path == "AGENTS.md" {
			b = f.Data
			break
		}
	}
	p := filepath.Join(c.CodexHome, "AGENTS.md")
	put(t, p, append(b, []byte("\nuser edit")...))
	_, e := Scan(c, []string{"codex"})
	var modified *ModifiedFileError
	if !errors.As(e, &modified) || modified.Path != p {
		t.Fatalf("errors.As found %+v in %v, want path %s", modified, e, p)
	}
	wrapped := fmt.Errorf("planning: %w", e)
	if !errors.As(wrapped, &modified) || modified.Path != p {
		t.Fatalf("errors.As does not see through a wrapped error: %v", wrapped)
	}
	want := p + " differs from what Hive expects there; undo the change, restore it from a backup, or move your own file elsewhere"
	if e.Error() != want {
		t.Errorf("message = %q, want %q", e.Error(), want)
	}
	if strings.Contains(e.Error(), "legacy") {
		t.Errorf("message calls the file legacy: %q", e.Error())
	}
}
func TestManifestTraversalRejected(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.ClaudeHome, ".deploy-manifest"), []byte("# source_commit: 16e7d335\n../secret\n"))
	if _, e := Scan(c, []string{"claude", "grok"}); e == nil {
		t.Fatal("unsafe manifest accepted")
	}
}
func TestUserHooksPreserved(t *testing.T) {
	c := fixture(t)
	p := filepath.Join(c.CodexHome, "hooks.json")
	put(t, p, []byte(`{"user":7,"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"\"$HOME/.codex/hooks/flow-session-context.sh\""},{"type":"command","command":"user-hook"}]}]}}`))
	r, e := Scan(c, []string{"codex"})
	if e != nil || len(r.Edits) != 1 {
		t.Fatalf("%+v %v", r, e)
	}
	if string(r.Edits[0].After) == "" {
		t.Fatal("empty result")
	}
}
func TestPiRenderedTemplate(t *testing.T) {
	c := fixture(t)
	c.PiHome = filepath.Join(c.Home, "custom pi")
	var f fileRecord
	for _, x := range loadCatalog().Files {
		if x.Root == "pi" && bytes.Contains(x.Data, []byte("__HIVE_PI_ROOT__")) {
			f = x
			break
		}
	}
	if f.Path == "" {
		t.Fatal("template missing")
	}
	put(t, filepath.Join(c.PiHome, f.Path), render(f, c.PiHome))
	r, e := Scan(c, []string{"pi"})
	if e != nil || len(r.Edits) != 1 {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestPatchedPackageRestoredAndPristinePreserved(t *testing.T) {
	c := fixture(t)
	for rel, hashes := range patches().Files {
		data, e := os.ReadFile(filepath.Join("testdata", filepath.Base(rel)+".txt"))
		if e != nil {
			t.Fatal(e)
		}
		if digest(data) != hashes.SHA256After {
			t.Fatal("bad fixture")
		}
		put(t, filepath.Join(c.PiHome, patchPrefix, rel), data)
	}
	r, e := Scan(c, []string{"pi"})
	if e != nil || len(r.Edits) != 3 {
		t.Fatalf("%+v %v", r, e)
	}
	for _, x := range r.Edits {
		if x.Delete {
			t.Fatal("package deleted")
		}
		put(t, x.Path, x.After)
	}
	r, e = Scan(c, []string{"pi"})
	if e != nil || r.Detected {
		t.Fatalf("pristine changed: %+v %v", r, e)
	}
}
func TestSharedConsumersAndUnknownChild(t *testing.T) {
	c := fixture(t)
	var f fileRecord
	for _, x := range loadCatalog().Files {
		if x.Root == "shared" && strings.HasSuffix(x.Path, "/SKILL.md") {
			f = x
			break
		}
	}
	p := filepath.Join(c.Home, ".agents", f.Path)
	put(t, p, f.Data)
	if _, e := Scan(c, []string{"codex"}); e == nil {
		t.Fatal("shared scope silently broadened")
	}
	r, e := Scan(c, allHosts)
	if e != nil || len(r.Hosts) != 5 {
		t.Fatalf("%+v %v", r, e)
	}
	put(t, filepath.Join(filepath.Dir(p), "personal.txt"), []byte("mine"))
	if _, e = Scan(c, allHosts); e == nil {
		t.Fatal("unknown child allowed")
	}
}
func TestTrustedAliasExcluded(t *testing.T) {
	c := fixture(t)
	p := filepath.Join(c.ClaudeHome, "skills", "flow-build")
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.Symlink(filepath.Join(c.Home, ".agents", "skills", "flow-build"), p); e != nil {
		t.Fatal(e)
	}
	if _, e := Scan(c, allHosts); e == nil {
		t.Fatal("unknown alias trusted")
	}
	if _, e := ScanExcluding(c, allHosts, []string{p}); e != nil {
		t.Fatal(e)
	}
}
func TestDuplicateConfigRejected(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.CodexHome, "hooks.json"), []byte(`{"x":1,"x":2}`))
	if _, e := Scan(c, []string{"codex"}); e == nil {
		t.Fatal("duplicate keys silently lost")
	}
}
func TestPersonalInstructionsPreserved(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.CodexHome, "AGENTS.md"), []byte("my own instructions\n"))
	r, e := Scan(c, []string{"codex"})
	if e != nil || r.Detected {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestThreeManifestFormats(t *testing.T) {
	for _, kind := range []string{"claude", "shared", "pi"} {
		t.Run(kind, func(t *testing.T) {
			c := fixture(t)
			var f fileRecord
			for _, x := range loadCatalog().Files {
				if x.Root == kind && (x.Path == "CLAUDE.md" || x.Path == "AGENTS.md" || x.Path == "skills/README.md") {
					f = x
					break
				}
			}
			root := roots(c)[kind][0]
			data := render(f, root)
			put(t, filepath.Join(root, f.Path), data)
			manifest := ".hive-deploy-manifest.json"
			var b []byte
			if kind == "claude" {
				manifest = ".deploy-manifest"
				b = []byte("# source_commit: 16e7d335\n" + f.Manifest + "\n")
			} else {
				scope := kind
				if scope == "shared" {
					scope = "shared-skills"
				}
				b = []byte(fmt.Sprintf(`{"schemaVersion":1,"scope":%q,"managedFiles":{%q:{"sha256":%q,"source":%q}},"managedConfig":{}}`, scope, f.Path, digest(data), f.Source))
			}
			put(t, filepath.Join(root, manifest), b)
			r, e := Scan(c, allHosts)
			if e != nil || len(r.Edits) != 2 {
				t.Fatalf("%+v %v", r, e)
			}
			if kind != "claude" {
				b = bytes.ReplaceAll(b, []byte(digest(data)), []byte(strings.Repeat("0", 64)))
				put(t, filepath.Join(root, manifest), b)
				if _, e = Scan(c, allHosts); e == nil {
					t.Fatal("forged hash accepted")
				}
			}
		})
	}
}
func TestHistoricalDefaultAndEffectiveRoot(t *testing.T) {
	c := fixture(t)
	old := c.CodexHome
	c.CodexHome = filepath.Join(c.Home, "relocated codex")
	for _, f := range loadCatalog().Files {
		if f.Root == "codex" && f.Path == "AGENTS.md" {
			put(t, filepath.Join(old, f.Path), f.Data)
			put(t, filepath.Join(c.CodexHome, f.Path), f.Data)
			break
		}
	}
	r, e := Scan(c, []string{"codex"})
	if e != nil || len(r.Edits) != 2 {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestUnknownHiveHookRejected(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.CodexHome, "hooks.json"), []byte(`{"hooks":{"SessionStart":[{"hooks":[{"type":"command","command":"\"$HOME/.codex/hooks/flow-session-context.sh\" --unexpected"}]}]}}`))
	if _, e := Scan(c, []string{"codex"}); e == nil {
		t.Fatal("unknown active Hive hook accepted")
	}
}
func TestHooksOutsideBytesUnchanged(t *testing.T) {
	c := fixture(t)
	prefix := []byte("{\n \"personal\": { \"nested\" : [1,  2] }, \"hooks\" : ")
	suffix := []byte(",\n  \"after\":   true\n}\n")
	hooks := []byte(`{"SessionStart":[{"hooks":[{"type":"command","command":"\"$HOME/.codex/hooks/flow-session-context.sh\""},{"type":"command","command":"user-hook"}]}]}`)
	put(t, filepath.Join(c.CodexHome, "hooks.json"), append(append(append([]byte{}, prefix...), hooks...), suffix...))
	r, e := Scan(c, []string{"codex"})
	if e != nil || len(r.Edits) != 1 {
		t.Fatalf("%+v %v", r, e)
	}
	after := r.Edits[0].After
	if !bytes.HasPrefix(after, prefix) || !bytes.HasSuffix(after, suffix) || !bytes.Contains(after, []byte("user-hook")) {
		t.Fatal("unrelated content changed")
	}
}
func TestPiConfigOwnAsyncRemovedOnly(t *testing.T) {
	c := fixture(t)
	p := filepath.Join(c.PiHome, "extensions/subagent/config.json")
	for _, b := range []string{`{"forceTopLevelAsync":true,"personal": 7}`, `{"personal": 7,"forceTopLevelAsync":true}`, `{"forceTopLevelAsync":true}`} {
		put(t, p, []byte(b))
		edits, e := cleanPiConfig(p, "extensions/subagent/config.json", []byte(`{"forceTopLevelAsync":true}`), c.PiHome)
		if e != nil || len(edits) != 1 {
			t.Fatalf("%+v %v", edits, e)
		}
		var obj map[string]any
		if e = json.Unmarshal(edits[0].After, &obj); e != nil {
			t.Fatal(e)
		}
		if _, ok := obj["forceTopLevelAsync"]; ok {
			t.Fatal("owned key retained")
		}
		if strings.Contains(b, "personal") && !bytes.Contains(edits[0].After, []byte(`"personal": 7`)) {
			t.Fatal("personal bytes changed")
		}
	}
}
func TestEditedPartialPiPatchRejectedWithoutManifest(t *testing.T) {
	c := fixture(t)
	for rel := range patches().Files {
		b, e := os.ReadFile(filepath.Join("testdata", filepath.Base(rel)+".txt"))
		if e != nil {
			t.Fatal(e)
		}
		if strings.HasSuffix(rel, "child-tool-plan.ts") {
			b = append(b, []byte("\n// user change\n")...)
		}
		put(t, filepath.Join(c.PiHome, patchPrefix, rel), b)
	}
	if _, e := Scan(c, []string{"pi"}); e == nil {
		t.Fatal("mixed patched package silently accepted")
	}
}
func TestPiRuntimeReferenceWithoutManifestRejected(t *testing.T) {
	c := fixture(t)
	put(t, filepath.Join(c.PiHome, "settings.json"), []byte(`{"extensions":["./extensions/hive-hooks.ts"]}`))
	if _, e := Scan(c, []string{"pi"}); e == nil {
		t.Fatal("runtime reference silently retained")
	}
}
