package content

import (
	"os"
	"path/filepath"
	"testing"
)

// skillEntrypointLimit is the largest size, in bytes, of a skill's SKILL.md.
// A host may keep only the start of an invoked skill after compaction
// (Claude Code keeps its first 5,000 tokens), so a longer entrypoint can lose
// its final steps; harness-audit's HA-SK-02 states the same limit.
const skillEntrypointLimit = 16000

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
