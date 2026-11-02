package content

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// skillEntrypointLimit is the largest size, in bytes, of a skill's SKILL.md.
// A host may keep only the start of an invoked skill after compaction
// (Claude Code keeps its first 5,000 tokens), so a longer entrypoint can lose
// its final steps; harness-audit's HA-SK-02 states the same limit.
const skillEntrypointLimit = 16000

// skillReferenceLimit is the largest size, in bytes, of a file in a skill's
// references directory. A model that caps or cuts a long read loses the rules
// after the cut (issue #204), so a reference past this size is split by
// responsibility; harness-audit's HA-SK-02 states the same limit.
const skillReferenceLimit = 16000

func TestSkillReferenceSize(t *testing.T) {
	roots, err := filepath.Glob("../../content/skills/*/references")
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			found++
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Size() > skillReferenceLimit {
				t.Errorf("%s: %d bytes, over the %d-byte skill reference limit", filepath.ToSlash(path[len("../../"):]), info.Size(), skillReferenceLimit)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if found == 0 {
		t.Fatal("no skill references found under content/skills")
	}
}

func TestSkillEntrypointSize(t *testing.T) {
	paths, err := filepath.Glob("../../content/skills/*/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no skill entrypoints found under content/skills")
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() > skillEntrypointLimit {
			t.Errorf("%s: %d bytes, over the %d-byte skill entrypoint limit", filepath.ToSlash(path[len("../../"):]), info.Size(), skillEntrypointLimit)
		}
	}
}
