package management

import (
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
}

func TestReleaseLinksResolveWithinPayload(t *testing.T) {
	o := setup(t)
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
	o := setup(t)
	p := plan(t, "install", o)
	for i := range p.Release.Files {
		if p.Release.Files[i].Path == SkillSource {
			p.Release.Files[i].Data = []byte("# Skill\nRead [required](references/absent.md).\n")
		}
	}
	p.Release.ID = releaseID(*p.Release)
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil || !strings.Contains(err.Error(), "instruction reference") {
		t.Fatalf("expected dependency rejection, got %v", err)
	}
	for _, ch := range p.Changes {
		absent(t, ch.Target.Path)
	}
}

func TestRepositoryCatalogueInstructionReferences(t *testing.T) {
	o := setup(t)
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	o.Source = source
	p := plan(t, "install", o)
	if len(p.Release.Files) < 3 {
		t.Fatal("empty repository catalogue")
	}
}
