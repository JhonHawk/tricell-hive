package management

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/opencode"
)

func TestOpenCodeContextLoaderDeterministicReadbackAndRegeneration(t *testing.T) {
	o := loaderOptions(t)
	p := plan(t, "install", o)
	pair := loaderPair(t, p)
	script, metadata, err := renderContextLoader(p.Release)
	if err != nil {
		t.Fatal(err)
	}
	for _, ch := range pair {
		want := script
		if strings.HasSuffix(ch.Target.Path, ".json") {
			want = metadata
		}
		if !bytes.Equal(ch.After.Managed, want) {
			t.Fatal("pair differs from release-derived bytes")
		}
	}
	second := plan(t, "install", o)
	for i, ch := range loaderPair(t, second) {
		if !bytes.Equal(ch.After.Managed, pair[i].After.Managed) {
			t.Fatal("nondeterministic snapshot generation")
		}
	}
	apply(t, p)
	for _, ch := range pair {
		if get(t, ch.Target.Path) != string(ch.After.Managed) {
			t.Fatal("installed readback differs")
		}
	}
	if unchanged, err := PlanUnchanged(plan(t, "install", o)); err != nil || !unchanged {
		t.Fatalf("stable snapshot: %t %v", unchanged, err)
	}
	put(t, filepath.Join(o.Source, "content/skills/flow-plan/SKILL.md"), "Regenerated canonical instruction\n")
	updated := plan(t, "install", o)
	for i, ch := range loaderPair(t, updated) {
		if bytes.Equal(ch.After.Managed, pair[i].After.Managed) {
			t.Fatal("source edit did not regenerate both derived resources")
		}
	}
	apply(t, updated)
	for _, ch := range loaderPair(t, updated) {
		if get(t, ch.Target.Path) != string(ch.After.Managed) {
			t.Fatal("regenerated readback differs")
		}
	}
}

func TestOpenCodeContextLoaderRejectsCoherentlyRehashedSavedPlan(t *testing.T) {
	for _, mutation := range []string{"text", "digest", "role", "resource", "order", "metadata", "stale snapshot", "missing source", "invalid UTF8", "invalid marker"} {
		t.Run(mutation, func(t *testing.T) {
			o := loaderOptions(t)
			p := plan(t, "install", o)
			// Derive a false, self-consistent pair from a different reference. Keep
			// the authoritative release intact: declared hashes cannot bless it.
			files := map[string][]byte{}
			for _, f := range p.Release.Files {
				files[f.Path] = append([]byte(nil), f.Data...)
			}
			files["content/skills/flow-plan/SKILL.md"] = []byte("False snapshot\n")
			fakeJS, fakeMetadata, err := opencode.RenderContextLoader(files[loaderSourceTest], files)
			if err != nil {
				t.Fatal(err)
			}
			if mutation == "text" {
				for i := range p.Changes {
					if p.Changes[i].Target.Kind == "plugin" {
						if strings.HasSuffix(p.Changes[i].Target.Path, ".js") {
							p.Changes[i].After.Managed = fakeJS
						} else {
							p.Changes[i].After.Managed = fakeMetadata
						}
					}
				}
			} else if mutation == "stale snapshot" {
				for i := range p.Release.Files {
					if p.Release.Files[i].Path == "content/skills/flow-plan/SKILL.md" {
						p.Release.Files[i].Data = []byte("Updated canonical reference\n")
					}
				}
				for i := range p.Changes {
					if p.Changes[i].Target.Source == "content/skills/flow-plan/SKILL.md" {
						p.Changes[i].After.Managed = []byte("Updated canonical reference\n")
					}
				}
			} else if mutation == "missing source" || mutation == "invalid UTF8" || mutation == "invalid marker" {
				for i, f := range p.Release.Files {
					if mutation == "invalid marker" && f.Path == loaderSourceTest {
						p.Release.Files[i].Data = []byte("export default {};")
					}
					if f.Path == "content/skills/flow-plan/SKILL.md" {
						if mutation == "missing source" {
							p.Release.Files = append(p.Release.Files[:i], p.Release.Files[i+1:]...)
							break
						}
						if mutation == "invalid UTF8" {
							p.Release.Files[i].Data = []byte{0xff}
						}
					}
				}
			} else {
				for i := range p.Changes {
					ch := &p.Changes[i]
					if ch.Target.Kind != "plugin" {
						continue
					}
					if mutation == "metadata" {
						if strings.HasSuffix(ch.Target.Path, ".json") {
							ch.After.Managed = fakeMetadata
						}
						continue
					}
					isJS := strings.HasSuffix(ch.Target.Path, ".js")
					data := ch.After.Managed
					if isJS {
						data = bytes.TrimSuffix(bytes.TrimPrefix(data, []byte("const snapshot = ")), []byte("; export default {};\n"))
					}
					var object struct {
						Schema     int `json:"schema"`
						References []struct {
							Role     string  `json:"role"`
							Resource string  `json:"resource"`
							Text     *string `json:"text,omitempty"`
							SHA256   string  `json:"sha256"`
						} `json:"references"`
					}
					if err := json.Unmarshal(data, &object); err != nil {
						t.Fatal(err)
					}
					switch mutation {
					case "digest":
						object.References[0].SHA256 = strings.Repeat("0", 64)
					case "role":
						object.References[0].Role = "unknown-role"
					case "resource":
						object.References[0].Resource = "other/SKILL.md"
					case "order":
						object.References[0], object.References[1] = object.References[1], object.References[0]
					}
					if isJS {
						data, err = json.Marshal(object)
					} else {
						data, err = json.MarshalIndent(object, "", "  ")
						data = append(data, '\n')
					}
					if err != nil {
						t.Fatal(err)
					}
					if isJS {
						data = append(append([]byte("const snapshot = "), data...), []byte("; export default {};\n")...)
					}
					ch.After.Managed = data
				}
			}
			p.Release.ID = releaseID(*p.Release)
			for i := range p.Changes {
				if p.Changes[i].After != nil {
					p.Changes[i].After.Release = p.Release.ID
				}
			}
			p.ID = planID(p)
			saved := filepath.Join(filepath.Dir(o.Source), "saved-plan.json")
			put(t, saved, string(encode(p)))
			loaded, err := LoadPlan(saved)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := (Engine{}).Apply(loaded); err == nil {
				t.Fatal("accepted coherently rehashed false snapshot")
			} else if mutation != "missing source" && mutation != "invalid UTF8" && mutation != "invalid marker" && !strings.Contains(err.Error(), "invalid ownership or release payload") {
				t.Fatalf("expected derived-pair mismatch, not an unrelated rejection: %v", err)
			}
			for _, ch := range loaded.Changes {
				absent(t, ch.Target.Path)
			}
		})
	}
}

func TestOpenCodeContextLoaderRecovery(t *testing.T) {
	for _, downgrade := range []bool{false, true} {
		for _, lateEdit := range []bool{false, true} {
			t.Run(fmt.Sprintf("downgrade=%t/lateEdit=%t", downgrade, lateEdit), func(t *testing.T) {
				o := setup(t)
				o.Hosts = []string{"opencode"}
				old := plan(t, "install", o)
				apply(t, old)
				putLoaderSources(t, o)
				install := plan(t, "install", o)
				pair := loaderPair(t, install)
				p := install
				if downgrade {
					apply(t, install)
					previous := o
					previous.ReleaseID = old.Release.ID
					p = plan(t, "install", previous)
				}
				stop := -1
				for i, ch := range p.Changes {
					if ch.Target.Kind == "plugin" {
						stop = i
						break
					}
				}
				engine := Engine{failpoint: func(stage string) error {
					if stage == fmt.Sprintf("write:%d", stop) {
						return errors.New("interrupted pair")
					}
					return nil
				}}
				if _, err := engine.Apply(p); err == nil {
					t.Fatal("did not interrupt between plugin writes")
				}
				config := filepath.Join(p.Config.OpenCodeHome, "opencode.jsonc")
				put(t, config, "new user configuration")
				extra := filepath.Join(p.Config.OpenCodeHome, "hive", "user.txt")
				put(t, extra, "later extra")
				var postPath string
				if lateEdit {
					postPath = p.Changes[stop].Target.Path
					put(t, postPath, "post-interruption user edit")
					if _, err := (Engine{}).Recover(o.StateDir); err == nil {
						t.Fatal("overwrote post-interruption plugin edit")
					}
					if get(t, postPath) != "post-interruption user edit" {
						t.Fatal("lost edit")
					}
					// Restore the exact journaled after-image to unblock ordinary recovery.
					if p.Changes[stop].After == nil {
						if err := os.Remove(postPath); err != nil {
							t.Fatal(err)
						}
					} else {
						put(t, postPath, string(p.Changes[stop].After.Managed))
					}
				}
				if _, err := (Engine{}).Recover(o.StateDir); err != nil {
					t.Fatal(err)
				}
				state, _, err := readState(o.StateDir)
				if err != nil {
					t.Fatal(err)
				}
				for _, ch := range pair {
					_, recorded := state.Records[ch.Target.Path]
					if recorded != downgrade {
						t.Fatal("pair ownership is inconsistent")
					}
					if downgrade {
						if get(t, ch.Target.Path) != string(ch.After.Managed) {
							t.Fatal("pair not restored")
						}
					} else {
						absent(t, ch.Target.Path)
					}
				}
				if get(t, config) != "new user configuration" || get(t, extra) != "later extra" {
					t.Fatal("recovery lost foreign files")
				}
			})
		}
	}
}

const loaderSourceTest = "integrations/opencode/context-loader.js"

const loaderTemplateTest = "const snapshot = __HIVE_CONTEXT_SNAPSHOT_DATA__; export default {};\n"

func putLoaderSources(t *testing.T, o Options) {
	t.Helper()
	put(t, filepath.Join(o.Source, loaderSourceTest), loaderTemplateTest)
	put(t, filepath.Join(o.Source, "content/skills/flow-plan/SKILL.md"), "# Plan\n")
	put(t, filepath.Join(o.Source, "content/skills/flow-plan/SKILL.md"), "\ufeffPlan \"quoted\" \\ Unicode 雪\r\n__HIVE_CONTEXT_SNAPSHOT_DATA__\n")
	put(t, filepath.Join(o.Source, "content/skills/harness-audit/SKILL.md"), "# Harness\n")
}

func TestOpenCodeContextLoaderRejectsInvalidSnapshotSources(t *testing.T) {
	for _, tc := range []struct {
		name, source, body string
		remove             bool
	}{
		{"no marker", loaderSourceTest, "export default {};", false},
		{"two markers", loaderSourceTest, loaderTemplateTest + "// __HIVE_CONTEXT_SNAPSHOT_DATA__", false},
		{"invalid template UTF8", loaderSourceTest, loaderTemplateTest + string([]byte{0xff}), false},
		{"missing reference", "content/skills/flow-plan/SKILL.md", "", true},
		{"invalid reference UTF8", "content/skills/flow-plan/SKILL.md", string([]byte{0xff}), false},
		{"missing owner", "content/skills/flow-plan/SKILL.md", "", true},
		{"missing harness", "content/skills/harness-audit/SKILL.md", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := loaderOptions(t)
			if tc.remove {
				if err := os.Remove(filepath.Join(o.Source, tc.source)); err != nil {
					t.Fatal(err)
				}
			} else {
				put(t, filepath.Join(o.Source, tc.source), tc.body)
			}
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("accepted invalid snapshot source")
			}
			absent(t, filepath.Join(o.Home, ".agents", "skills"))
		})
	}
}

func loaderOptions(t *testing.T) Options {
	o := setup(t)
	o.Hosts = []string{"opencode"}
	putLoaderSources(t, o)
	return o
}

func loaderPair(t *testing.T, p Plan) []Change {
	t.Helper()
	var pair []Change
	for _, ch := range p.Changes {
		if ch.Target.Kind == "plugin" {
			pair = append(pair, ch)
		}
	}
	if len(pair) != 2 {
		t.Fatalf("expected inactive script and metadata, got %d", len(pair))
	}
	return pair
}

func TestOpenCodeContextLoaderLifecycle(t *testing.T) {
	o := loaderOptions(t)
	c, _, err := normalize(o)
	if err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(c.OpenCodeHome, "opencode.jsonc")
	original := "// user-owned\n{\"plugins\": [\"user-plugin\"]}\n"
	put(t, config, original)
	p := plan(t, "install", o)
	pair := loaderPair(t, p)
	for _, ch := range pair {
		if filepath.Dir(ch.Target.Path) != filepath.Join(c.OpenCodeHome, "hive", "context-loader") {
			t.Fatal("autoload destination")
		}
		if ch.After == nil {
			t.Fatal("missing record")
		}
	}
	apply(t, p)
	if get(t, config) != original {
		t.Fatal("native user configuration changed")
	}
	var metadata struct {
		Schema     int                                       `json:"schema"`
		References []struct{ Role, Resource, SHA256 string } `json:"references"`
	}
	if err := json.Unmarshal([]byte(get(t, filepath.Join(c.OpenCodeHome, "hive", "context-loader", "context-loader.json"))), &metadata); err != nil {
		t.Fatal(err)
	}
	absent(t, filepath.Join(c.OpenCodeHome, "hive", "context-loader", "package.json"))
	info, err := os.Stat(filepath.Join(c.OpenCodeHome, "hive", "context-loader", "context-loader.json"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("metadata must be private: %v", err)
	}
	if metadata.Schema != 2 || len(metadata.References) != 2 || strings.Contains(get(t, filepath.Join(c.OpenCodeHome, "hive", "context-loader", "context-loader.json")), "skillsDirectory") {
		t.Fatal("metadata not derived from canonical snapshots")
	}
	for _, ch := range pair {
		if err := os.Remove(ch.Target.Path); err != nil {
			t.Fatal(err)
		}
		apply(t, plan(t, "install", o))
		put(t, ch.Target.Path, "edited")
		if _, err := BuildPlan("install", o); err == nil {
			t.Fatal("accepted plugin drift")
		}
		if _, err := BuildPlan("remove", o); err == nil {
			t.Fatal("removed plugin drift")
		}
		put(t, ch.Target.Path, string(ch.After.Managed))
	}
	apply(t, plan(t, "remove", o))
	for _, ch := range pair {
		absent(t, ch.Target.Path)
	}
	if get(t, config) != original {
		t.Fatal("remove changed user config")
	}
}

func TestOpenCodeContextLoaderTamperedPlanAndCollision(t *testing.T) {
	o := loaderOptions(t)
	p := plan(t, "install", o)
	pair := loaderPair(t, p)
	for i := range p.Changes {
		if strings.HasSuffix(p.Changes[i].Target.Path, "context-loader.json") {
			p.Changes[i].After.Managed = []byte(`{"schema":1,"skillsDirectory":"/private"}`)
		}
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("accepted rehashed derived metadata")
	}
	for _, ch := range pair {
		absent(t, ch.Target.Path)
	}
	for _, ch := range pair {
		put(t, ch.Target.Path, "user-owned")
		if _, err := BuildPlan("install", o); err == nil {
			t.Fatal("adopted unowned plugin")
		}
		if get(t, ch.Target.Path) != "user-owned" {
			t.Fatal("collision overwritten")
		}
		if err := os.Remove(ch.Target.Path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestOpenCodeContextLoaderHostScopeAndOldSnapshot(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"opencode"}
	old := plan(t, "install", o)
	apply(t, old)
	putLoaderSources(t, o)
	newPlan := plan(t, "install", o)
	pair := loaderPair(t, newPlan)
	apply(t, newPlan)
	rollback := o
	rollback.ReleaseID = old.Release.ID
	apply(t, plan(t, "install", rollback))
	for _, ch := range pair {
		absent(t, ch.Target.Path)
	}
	for _, host := range []string{"claude", "codex", "grok", "pi", "cursor"} {
		other := loaderOptions(t)
		other.Hosts = []string{host}
		for _, ch := range plan(t, "install", other).Changes {
			if ch.Target.Kind == "plugin" {
				t.Fatalf("plugin delivered to %s", host)
			}
		}
	}
	for _, host := range []string{"claude", "codex"} {
		other := loaderOptions(t)
		other.Hosts = []string{host}
		other.Scope = "project"
		other.Root = filepath.Join(other.Home, "project")
		for _, ch := range plan(t, "install", other).Changes {
			if ch.Target.Kind == "plugin" {
				t.Fatalf("project plugin delivered to %s", host)
			}
		}
	}
}

func TestOpenCodeContextLoaderDirectoryCleanup(t *testing.T) {
	for _, downgrade := range []bool{false, true} {
		for _, preservation := range []string{"created", "foreign file", "preexisting"} {
			t.Run(fmt.Sprintf("downgrade=%t/%s", downgrade, preservation), func(t *testing.T) {
				o := setup(t)
				o.Hosts = []string{"opencode"}
				old := plan(t, "install", o)
				apply(t, old)
				putLoaderSources(t, o)
				p := plan(t, "install", o)
				pair := loaderPair(t, p)
				dir := filepath.Dir(pair[0].Target.Path)
				if preservation == "preexisting" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
				}
				foreign := filepath.Join(p.Config.OpenCodeHome, "hive", "user.txt")
				put(t, foreign, "user-owned sibling")
				apply(t, p)
				nested := filepath.Join(dir, "user.txt")
				if preservation == "foreign file" {
					put(t, nested, "user-owned nested file")
				}
				if downgrade {
					rollback := o
					rollback.ReleaseID = old.Release.ID
					apply(t, plan(t, "install", rollback))
				} else {
					apply(t, plan(t, "remove", o))
				}
				for _, ch := range pair {
					absent(t, ch.Target.Path)
				}
				if preservation == "created" {
					absent(t, dir)
				} else {
					info, err := os.Stat(dir)
					if err != nil || !info.IsDir() {
						t.Fatalf("preserved directory: %v", err)
					}
				}
				if get(t, foreign) != "user-owned sibling" {
					t.Fatal("lost sibling file")
				}
				if preservation == "foreign file" && get(t, nested) != "user-owned nested file" {
					t.Fatal("lost nested file")
				}
			})
		}
	}
}

func TestOpenCodeContextLoaderRejectsUnpublishedLegacyNames(t *testing.T) {
	o := loaderOptions(t)
	p := plan(t, "install", o)
	for i := range p.Changes {
		ch := &p.Changes[i]
		if ch.Target.Kind != "plugin" {
			continue
		}
		name := filepath.Base(ch.Target.Path)
		if name == "index.js" {
			name = "context-loader.js"
		}
		ch.Target.Path = filepath.Join(p.Config.OpenCodeHome, "hive", name)
		ch.After.Target.Path = ch.Target.Path
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("accepted unpublished legacy plugin paths")
	}
	for _, ch := range p.Changes {
		absent(t, ch.Target.Path)
	}
}
