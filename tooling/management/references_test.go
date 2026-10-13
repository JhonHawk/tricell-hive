package management

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseRejectsBrokenInstructionDependencies(t *testing.T) {
	for _, link := range []string{"references/missing.md", "../missing/SKILL.md", "../../../README.md", "/Users/alice/private.md", "file:///tmp/private.md", "skill:missing/references/x.md", "../%2e%2e/README.md", "skill:other/../workspace-conventions/SKILL.md"} {
		t.Run(link, func(t *testing.T) {
			o := setup(t)
			put(t, filepath.Join(o.Source, SkillSource), "# Skill\nRead [required]("+link+").\n")
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatalf("accepted absent or nonportable dependency %s", link)
			}
		})
	}
	for _, tc := range []struct {
		name, body, reason string
	}{
		{"missing resource", "[required](skill:workspace-conventions/references/missing.md)", "target is not included in this release"},
		{"missing owner", "[required](skill:missing/references/guide.md)", "missing owning skill"},
		{"invalid locator", "[required](skill:workspace-conventions/../other/SKILL.md)", "invalid skill locator"},
		{"personal path", "[required](/Users/alice/private.md)", "nonportable personal path"},
		{"unsupported scheme", "[required](file:///tmp/private.md)", "unsupported local destination"},
		{"reference definition", "[required][guide]\n\n[guide]: skill:workspace-conventions/references/missing.md", "target is not included in this release"},
	} {
		t.Run("global/"+tc.name, func(t *testing.T) {
			o := setup(t)
			put(t, filepath.Join(o.Source, GlobalSource), "# Rules\n"+tc.body+"\n")
			if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), tc.reason) || !strings.Contains(err.Error(), GlobalSource) {
				t.Fatalf("expected global dependency rejection (%s), got %v", tc.reason, err)
			}
			absent(t, filepath.Join(o.Home, ".codex", "AGENTS.md"))
			absent(t, filepath.Join(o.Home, ".claude", "CLAUDE.md"))
		})
	}
}

func TestReleaseLinksResolveWithinPayload(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, GlobalSource), "# Rules\nRead [logical](skill:other/references/guide.md), [relative](../skills/other/SKILL.md), [directory](skill:other/references/), [web](https://example.com/docs), [http](http://example.com/docs), [mail](mailto:team@example.com) and [anchor](#rules).\nRead [guide][details].\n\n[details]: <skill:other/references/guide.md>\n```md\n[example](skill:missing/references/absent.md)\n/Users/alice/example.md\n```\n~~~md\n[example](file:///tmp/not-real.md)\n~~~\n")
	put(t, filepath.Join(o.Source, SkillSource), "# Skill\nRead [local](references/guide.md#section), [other](../other/SKILL.md), [external](https://example.com/docs) and [logical](skill:other/references/guide.md) and [directory](skill:other/references/).\n```md\n[example](references/not-real.md)\n```\n")
	put(t, filepath.Join(o.Source, "content/skills/workspace-conventions/references/guide.md"), "# Section\nRead [entry](../SKILL.md).\n")
	put(t, filepath.Join(o.Source, "content/skills/other/SKILL.md"), "# Other\nRead [guide][details].\n\n[details]: references/guide.md\n")
	put(t, filepath.Join(o.Source, "content/skills/other/references/guide.md"), "# Guide\n")
	p := plan(t, "install", o)
	apply(t, p)
	if get(t, filepath.Join(o.Home, ".agents/skills/other/references/guide.md")) != "# Guide\n" {
		t.Fatal("dependency not installed")
	}
	// Remove the referenced skill while leaving the consumer in source.
	if err := os.RemoveAll(filepath.Join(o.Source, "content/skills/other")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("accepted retirement of required skill")
	}
}

func TestApplyRejectsRehashedReleaseWithMissingDependency(t *testing.T) {
	for _, source := range []string{SkillSource, GlobalSource} {
		t.Run(source, func(t *testing.T) {
			o := setup(t)
			p := plan(t, "install", o)
			for i := range p.Release.Files {
				if p.Release.Files[i].Path == source {
					p.Release.Files[i].Data = []byte("# Rules\nRead [required](skill:workspace-conventions/references/absent.md).\n")
				}
			}
			p.Release.ID = releaseID(*p.Release)
			p.ID = planID(p)
			data, err := json.Marshal(p)
			if err != nil {
				t.Fatal(err)
			}
			savedDir, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			saved := filepath.Join(savedDir, "plan.json")
			put(t, saved, string(data))
			var loaded Plan
			if err := decodeFile(saved, &loaded); err != nil {
				t.Fatal(err)
			}
			if _, err := (Engine{}).Apply(loaded); err == nil || !strings.Contains(err.Error(), "instruction reference "+source) || !strings.Contains(err.Error(), "target is not included in this release") {
				t.Fatalf("expected dependency rejection, got %v", err)
			}
			for _, ch := range loaded.Changes {
				absent(t, ch.Target.Path)
			}
		})
	}
}

func TestRepositoryCatalogueInstructionReferences(t *testing.T) {
	o := setup(t)
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	o.Source = source
	// Plans are cheap: check the references on the default setup hosts. An
	// apply is expensive, so it runs with the single claude host, which
	// installs guidance, skills and agents.
	defaultPlan := plan(t, "install", o)
	if len(defaultPlan.Release.Files) < 3 {
		t.Fatal("empty repository catalogue")
	}
	o.Hosts = []string{"claude"}
	p := plan(t, "install", o)
	apply(t, p)
	if unchanged, err := PlanUnchanged(plan(t, "install", o)); err != nil || !unchanged {
		t.Fatalf("second plan after apply: unchanged=%v err=%v", unchanged, err)
	}
	claude := filepath.Join(o.Home, ".claude")
	if !strings.Contains(get(t, filepath.Join(claude, "CLAUDE.md")), Begin) {
		t.Fatal("global guidance block not installed")
	}
	// Skills live in the shared .agents root; claude renders its agents itself.
	for _, dir := range []string{filepath.Join(o.Home, ".agents", "skills"), filepath.Join(claude, "agents")} {
		found := false
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") {
				found = true
			}
			return err
		})
		if err != nil || !found {
			t.Fatalf("no installed markdown under %s: %v", dir, err)
		}
	}
}
