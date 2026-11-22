package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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

// referenceRoots returns the folders under skillsDir whose files count as skill
// references: each skill's references folder, and each domain folder
// hive-core/assets/domains.json lists, which hold a leaf's own references in
// <domain>/<leaf>/. A manifest that cannot be read is reported as a problem.
func referenceRoots(skillsDir string) (roots, problems []string) {
	found, err := filepath.Glob(filepath.Join(skillsDir, "*", "references"))
	if err != nil {
		return nil, []string{err.Error()}
	}
	roots = append(roots, found...)
	coreDir := filepath.Join(skillsDir, coreSkillName)
	domains, err := readDomains(coreDir)
	if err != nil {
		return roots, []string{fmt.Sprintf("%s/%s: %v", coreSkillName, domainsManifestPath, err)}
	}
	for _, domain := range domains {
		roots = append(roots, filepath.Join(coreDir, domain))
	}
	return roots, nil
}

// referenceSizeProblems returns one line per file under the reference folders
// of skillsDir that is over skillReferenceLimit, and one when none is found.
func referenceSizeProblems(skillsDir string) []string {
	roots, problems := referenceRoots(skillsDir)
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
				rel, _ := filepath.Rel(skillsDir, path)
				problems = append(problems, fmt.Sprintf("content/skills/%s: %d bytes, over the %d-byte skill reference limit", filepath.ToSlash(rel), info.Size(), skillReferenceLimit))
			}
			return nil
		})
		if err != nil {
			problems = append(problems, err.Error())
		}
	}
	if found == 0 {
		problems = append(problems, "no skill references found under content/skills")
	}
	return problems
}

func TestSkillReferenceSize(t *testing.T) {
	for _, problem := range referenceSizeProblems("../../content/skills") {
		t.Error(problem)
	}
}

func TestSkillReferenceSizeCatchesOversizeFiles(t *testing.T) {
	skills := filepath.Join(t.TempDir(), "content", "skills")
	if got := referenceSizeProblems(skills); len(got) == 0 || !strings.Contains(strings.Join(got, "\n"), "no skill references found") {
		t.Errorf("an empty tree must report that nothing was found, got %v", got)
	}
	big := strings.Repeat("x", skillReferenceLimit+1)
	writeFile(t, filepath.Join(skills, coreSkillName, "assets", "domains.json"), `{"domains": ["build", "principles"]}`)
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "checks.md"), "# Checks\n")
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "checks", "own.md"), big)
	writeFile(t, filepath.Join(skills, coreSkillName, "principles", "index.md"), "# P\n")
	writeFile(t, filepath.Join(skills, coreSkillName, "unlisted", "x.md"), big)
	writeFile(t, filepath.Join(skills, "flow-x", "references", "r.md"), big)
	got := strings.Join(referenceSizeProblems(skills), "\n")
	for _, want := range []string{"content/skills/hive-core/build/checks/own.md: ", "content/skills/flow-x/references/r.md: "} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "unlisted") {
		t.Errorf("a folder outside domains.json is the leaf test's to report, got:\n%s", got)
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
