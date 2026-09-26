package management

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

func sharedPath(o Options) string {
	return filepath.Join(o.Home, ".agents", "skills", "workspace-conventions", "SKILL.md")
}
func aliasPath(o Options) string {
	return filepath.Join(o.Home, ".claude", "skills", "workspace-conventions")
}
func stateFor(t *testing.T, o Options) State {
	t.Helper()
	s, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestFiveHostSharedConsumersAndPartialRemoval(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "pi", "opencode"}
	p := plan(t, "install", o)
	if len(p.Changes) != 6 {
		t.Fatalf("expected four blocks, shared skill and alias; got %d", len(p.Changes))
	}
	apply(t, p)
	s := stateFor(t, o)
	if s.Version != 5 || len(s.Records[sharedPath(o)].Consumers) != 5 {
		t.Fatal("missing shared consumers")
	}
	link, err := os.Readlink(aliasPath(o))
	if err != nil || link != "../../.agents/skills/workspace-conventions" {
		t.Fatalf("alias: %q %v", link, err)
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("reinstall was not idempotent")
	}
	partial := o
	partial.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", partial))
	absent(t, aliasPath(o))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 4 {
		t.Fatal("removed another consumer")
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".claude", "CLAUDE.md")), Begin) {
		t.Fatal("removed Grok's shared block")
	}
	entries, err := Status(partial)
	if err != nil {
		t.Fatal(err)
	}
	retained := 0
	for _, en := range entries {
		if en.Status == "retained_shared" {
			retained++
		}
	}
	if retained != 2 {
		t.Fatalf("retained resource visibility: %+v", entries)
	}
	for _, host := range []string{"codex", "grok", "pi", "opencode"} {
		partial.Hosts = []string{host}
		apply(t, plan(t, "remove", partial))
	}
	absent(t, sharedPath(o))
	absent(t, filepath.Join(o.Home, ".claude", "CLAUDE.md"))
	if len(stateFor(t, o).Records) != 0 {
		t.Fatal("ownership not released")
	}
}

func TestConsumerOnlyChangesAndSharedUpdateConflict(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	apply(t, plan(t, "install", o))
	before := get(t, sharedPath(o))
	firstRelease := stateFor(t, o).Records[sharedPath(o)].Release
	o.Hosts = []string{"opencode"}
	if apply(t, plan(t, "install", o)) == "unchanged" {
		t.Fatal("consumer-only update was skipped")
	}
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 2 || get(t, sharedPath(o)) != before {
		t.Fatal("consumer registration changed bytes")
	}
	o.Hosts = []string{"codex"}
	put(t, filepath.Join(o.Source, GlobalSource), "Changed global only.\n")
	apply(t, plan(t, "install", o))
	if stateFor(t, o).Records[sharedPath(o)].Release != firstRelease {
		t.Fatal("unselected consumer release was relabeled")
	}
	put(t, filepath.Join(o.Source, SkillSource), before+"New shared instruction.\n")
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "all consumers") {
		t.Fatalf("partial update accepted: %v", err)
	}
	if get(t, sharedPath(o)) != before {
		t.Fatal("conflicting plan wrote payload")
	}
	o.Hosts = []string{"codex", "opencode"}
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, sharedPath(o)), "New shared instruction") {
		t.Fatal("complete update failed")
	}
}

func TestManagedAliasRejectsUnknownLinksAndRetargeting(t *testing.T) {
	t.Run("unowned identical", func(t *testing.T) {
		o := setup(t)
		os.MkdirAll(filepath.Dir(aliasPath(o)), 0700)
		if err := os.Symlink("../../.agents/skills/workspace-conventions", aliasPath(o)); err != nil {
			t.Fatal(err)
		}
		if _, err := BuildPlan("install", o); err == nil {
			t.Fatal("adopted unknown link")
		}
		absent(t, sharedPath(o))
	})
	t.Run("owned retarget", func(t *testing.T) {
		o := setup(t)
		apply(t, plan(t, "install", o))
		before := get(t, sharedPath(o))
		victim := filepath.Join(o.Home, "user-directory")
		put(t, filepath.Join(victim, "SKILL.md"), "user data")
		os.Remove(aliasPath(o))
		os.Symlink(victim, aliasPath(o))
		for _, action := range []string{"install", "remove"} {
			if _, err := BuildPlan(action, o); err == nil {
				t.Fatal("retargeted link accepted")
			}
		}
		if get(t, filepath.Join(victim, "SKILL.md")) != "user data" || get(t, sharedPath(o)) != before {
			t.Fatal("followed retargeted link")
		}
	})
	t.Run("stale link", func(t *testing.T) {
		o := setup(t)
		p := plan(t, "install", o)
		os.MkdirAll(filepath.Dir(aliasPath(o)), 0700)
		os.Symlink("../../.agents/skills/workspace-conventions", aliasPath(o))
		if _, err := (Engine{}).Apply(p); err == nil {
			t.Fatal("late unknown link accepted")
		}
		absent(t, sharedPath(o))
	})
}

// This fixture uses the previous per-host layout and old serialization (all
// added v2 fields are omitted). Planning must not rewrite its state or files.
func legacyInstall(t *testing.T, o Options) []byte {
	t.Helper()
	r, err := loadRelease(o, o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	s := State{Version: 1, Records: map[string]Record{}}
	for _, host := range o.Hosts {
		base := filepath.Join(o.Home, ".codex")
		instruction := "AGENTS.md"
		skill := sharedPath(o)
		if host == "claude" {
			base = filepath.Join(o.Home, ".claude")
			instruction = "CLAUDE.md"
			skill = filepath.Join(aliasPath(o), "SKILL.md")
			s.CreatedDirs = append(s.CreatedDirs, aliasPath(o))
		}
		for _, kind := range []string{"block", "skill"} {
			path := filepath.Join(base, instruction)
			managed := managedBlock(r.Files[0].Data, nil)
			if kind == "skill" {
				path = skill
				managed = r.Files[1].Data
			}
			put(t, path, string(managed))
			s.Records[path] = Record{Target: target.Target{Path: path, Kind: kind, Host: host, Scope: "user", Context: o.Home}, Managed: managed, CreatedFile: true, Release: r.ID}
		}
	}
	data := encode(s)
	put(t, filepath.Join(o.StateDir, "state.json"), string(data))
	return data
}

func TestLegacyClaudeMigrationAndRemoval(t *testing.T) {
	t.Run("migrate", func(t *testing.T) {
		o := setup(t)
		old := legacyInstall(t, o)
		p := plan(t, "install", o)
		if get(t, filepath.Join(o.StateDir, "state.json")) != string(old) {
			t.Fatal("planning migrated state")
		}
		apply(t, p)
		if _, err := os.Readlink(aliasPath(o)); err != nil {
			t.Fatal(err)
		}
		s := stateFor(t, o)
		if s.Version != 5 || len(s.Records[sharedPath(o)].Consumers) != 2 {
			t.Fatal("migration ownership")
		}
		if _, exists := s.Records[filepath.Join(aliasPath(o), "SKILL.md")]; exists {
			t.Fatal("legacy ownership retained")
		}
		apply(t, plan(t, "remove", o))
		absent(t, aliasPath(o))
		absent(t, sharedPath(o))
	})
	t.Run("remove without adoption", func(t *testing.T) {
		o := setup(t)
		legacyInstall(t, o)
		o.Hosts = []string{"claude"}
		apply(t, plan(t, "remove", o))
		absent(t, aliasPath(o))
		if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
			t.Fatal("removed Codex ownership")
		}
	})
	for _, bad := range []string{"extra", "drift", "unowned-directory"} {
		t.Run(bad, func(t *testing.T) {
			o := setup(t)
			legacyInstall(t, o)
			switch bad {
			case "extra":
				put(t, filepath.Join(aliasPath(o), ".user"), "retain")
			case "drift":
				put(t, filepath.Join(aliasPath(o), "SKILL.md"), "retain")
			case "unowned-directory":
				s, _, err := readState(o.StateDir)
				if err != nil {
					t.Fatal(err)
				}
				s.Version = 2
				s.CreatedDirs = nil
				put(t, filepath.Join(o.StateDir, "state.json"), string(encode(s)))
			}
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("unsafe legacy migration accepted")
			}
			if info, err := os.Lstat(aliasPath(o)); err != nil || !info.IsDir() {
				t.Fatal("changed conflicting directory")
			}
		})
	}
}

func TestMigrationRecoveryAtIntermediateAndWriteBoundaries(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "migration:skill", "migration:directory", "write:2", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"claude"}
			beforeState := legacyInstall(t, o)
			beforeSkill := get(t, filepath.Join(aliasPath(o), "SKILL.md"))
			put(t, filepath.Join(o.Source, GlobalSource), "Updated rules during migration.\n")
			p := plan(t, "install", o)
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("stop")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatalf("failpoint not reached: %s", stage)
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, filepath.Join(o.StateDir, "state.json")) != string(beforeState) || get(t, filepath.Join(aliasPath(o), "SKILL.md")) != beforeSkill {
				t.Fatal("legacy bytes not restored")
			}
			if info, err := os.Lstat(aliasPath(o)); err != nil || !info.IsDir() {
				t.Fatal("legacy directory not restored")
			}
			absent(t, sharedPath(o))
			apply(t, plan(t, "install", o))
		})
	}
}

func TestMigrationRecoveryConflictAndRetry(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(fmt.Sprint(conflict), func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"claude"}
			before := legacyInstall(t, o)
			p := plan(t, "install", o)
			e := Engine{failpoint: func(stage string) error {
				if stage == "write:2" {
					return errors.New("stop")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("expected interrupted migration")
			}
			if conflict {
				os.Remove(aliasPath(o))
				os.Symlink("unrelated", aliasPath(o))
				if _, err := (Engine{}).Recover(o.StateDir); err == nil {
					t.Fatal("retargeted alias overwritten")
				}
				if link, _ := os.Readlink(aliasPath(o)); link != "unrelated" {
					t.Fatal("conflict not preserved")
				}
				if _, err := os.Stat(sharedPath(o)); err != nil {
					t.Fatal("preflight did not preserve other resources")
				}
				return
			}
			recovery := Engine{failpoint: func(stage string) error {
				if stage == "recover:2" {
					return errors.New("stop recovery")
				}
				return nil
			}}
			if _, err := recovery.Recover(o.StateDir); err == nil {
				t.Fatal("recovery interruption not reached")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, filepath.Join(o.StateDir, "state.json")) != string(before) {
				t.Fatal("state not restored after retry")
			}
		})
	}
}

func TestRehashedConsumerForgeryAndLegacyPlanRejected(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	for i := range p.Changes {
		if p.Changes[i].Target.Kind == "skill" {
			p.Changes[i].After.Consumers = nil
		}
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("forged consumer set accepted")
	}
	absent(t, sharedPath(o))
	p = plan(t, "install", o)
	p.Version = 1
	p.ID = planID(p)
	file := filepath.Join(o.Home, "v1-plan.json")
	put(t, file, string(encode(p)))
	if _, err := LoadPlan(file); err == nil {
		t.Fatal("legacy saved plan accepted")
	}
}

func TestSharedIdentityKeepsConsumerContextsSeparate(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex"}
	apply(t, plan(t, "install", o))
	project := o
	project.Scope = "project"
	project.Root = o.Home
	apply(t, plan(t, "install", project))
	s := stateFor(t, o)
	if len(s.Records[sharedPath(o)].Consumers) != 2 {
		t.Fatal("consumer scopes collapsed")
	}
	apply(t, plan(t, "remove", project))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
		t.Fatal("project removal deleted user consumer")
	}
	if _, err := os.Stat(sharedPath(o)); err != nil {
		t.Fatal(err)
	}
}

func TestSharedBlockDoesNotHideRelocatedConsumer(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "grok"}
	apply(t, plan(t, "install", o))
	c, _, err := normalize(o)
	if err != nil {
		t.Fatal(err)
	}
	c.ClaudeHome = filepath.Join(o.Home, "different-claude")
	if _, err := desiredResources(c, o.Hosts, stateFor(t, o), "remove"); err == nil {
		t.Fatal("Grok's retained physical block hid relocated Claude ownership")
	}
}

func TestLegacyPendingJournalUsesOriginalHashes(t *testing.T) {
	for _, version := range []int{1, 2, 3, 4} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"codex"}
			// Golden V1 types intentionally lack every newly added field.
			type oldConfig struct {
				Scope      string `json:"scope"`
				Home       string `json:"home"`
				Root       string `json:"root,omitempty"`
				CodexHome  string `json:"codex_home"`
				ClaudeHome string `json:"claude_home"`
			}
			type oldTarget struct{ Path, Kind, Host, Scope, Context string }
			type oldRecord struct {
				Target      oldTarget `json:"target"`
				Managed     []byte    `json:"managed"`
				Leading     string    `json:"leading,omitempty"`
				CreatedFile bool      `json:"created_file"`
				Release     string    `json:"release"`
			}
			type oldFingerprint struct {
				Exists bool   `json:"exists"`
				Hash   string `json:"hash"`
				Mode   uint32 `json:"mode"`
			}
			type oldChange struct {
				Target   oldTarget      `json:"target"`
				Expected oldFingerprint `json:"expected"`
				Before   *oldRecord     `json:"before,omitempty"`
				After    *oldRecord     `json:"after,omitempty"`
			}
			type oldPlan struct {
				Version   int         `json:"version"`
				Action    string      `json:"action"`
				Config    oldConfig   `json:"config"`
				Hosts     []string    `json:"hosts"`
				StateDir  string      `json:"state_dir"`
				StateHash string      `json:"state_hash"`
				Release   *Release    `json:"release,omitempty"`
				Changes   []oldChange `json:"changes"`
				ID        string      `json:"id"`
			}
			type oldSnapshot struct {
				Exists bool
				Data   []byte
				Mode   uint32
			}
			type oldEntry struct {
				Change        oldChange
				Before, After oldSnapshot
			}
			type oldJournal struct {
				Version                 int
				ID, Phase               string
				Plan                    oldPlan
				Entries                 []oldEntry
				BeforeState, AfterState oldSnapshot
				CreatedDirs             []string
				Integrity               string
			}
			path := filepath.Join(o.Home, ".codex", "AGENTS.md")
			targetV1 := oldTarget{path, "block", "codex", "user", o.Home}
			r, err := loadRelease(o, o.StateDir)
			if err != nil {
				t.Fatal(err)
			}
			managed := managedBlock(r.Files[0].Data, nil)
			ch := oldChange{Target: targetV1, Expected: oldFingerprint{false, hash(nil), 0}, After: &oldRecord{Target: targetV1, Managed: managed, CreatedFile: true, Release: r.ID}}
			op := oldPlan{Version: version, Action: "install", Config: oldConfig{Scope: "user", Home: o.Home, CodexHome: filepath.Join(o.Home, ".codex"), ClaudeHome: filepath.Join(o.Home, ".claude")}, Hosts: []string{"codex"}, StateDir: o.StateDir, StateHash: hash(nil), Release: &r, Changes: []oldChange{ch}}
			op.ID = hash(encode(op))
			j := oldJournal{Version: version, ID: strings.Repeat("a", 32), Phase: "prepared", Plan: op, Entries: []oldEntry{{Change: ch, After: oldSnapshot{true, managed, 0600}}}}
			j.Integrity = hash(encode(j))
			put(t, path, string(managed))
			os.Chmod(path, 0600)
			put(t, filepath.Join(o.StateDir, "transactions", j.ID+".json"), string(encode(j)))
			put(t, filepath.Join(o.StateDir, "pending.json"), string(encode(pending{j.ID})))
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, path)
			absent(t, filepath.Join(o.StateDir, "pending.json"))
			var recovered journal
			if err := decodeFile(filepath.Join(o.StateDir, "transactions", j.ID+".json"), &recovered); err != nil {
				t.Fatal(err)
			}
			if recovered.Version != version || recovered.Integrity != journalHash(recovered) || !bytes.Equal(recovered.Entries[0].After.Data, managed) {
				t.Fatal("legacy recovery changed serialization contract")
			}
		})
	}
}
