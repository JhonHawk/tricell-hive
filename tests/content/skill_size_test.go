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
// Host limit: after compaction Claude Code keeps only the first 5,000 tokens
// of each invoked skill (code.claude.com/docs/en/skills, «Skill content
// lifecycle», read 2026-11-22 on 2.1.330), so a longer entrypoint loses its
// final steps. It changes only when that cut changes.
const skillEntrypointLimit = 16000

// wholeReadLimit is the largest size, in bytes, of any other file a model
// reads in one pass: a skill reference, a rule or principle leaf, a playbook.
// It is the stage packet limit (tooling/management/stage_packets.go), and it
// leaves a model that caps a long read (issue #204) its later rules. One limit
// for every such file, with no exception list (user decision D5 A, D6 A,
// 2026-11-22); a leaf a packet holds also fits the packet's own limit.
const wholeReadLimit = 24000

// wholeReadProblem returns the failure for a whole-read file of size bytes,
// or "" when it fits.
func wholeReadProblem(label string, size int64) string {
	if size <= wholeReadLimit {
		return ""
	}
	return fmt.Sprintf("%s: whole-read file is %d bytes, over the %d-byte limit: split it by responsibility into files each linked with its read condition; do not raise the limit or trim a rule's text to fit", label, size, wholeReadLimit)
}

// languageLeafBudget is the largest size, in bytes, of the language leaf's
// SKILL.md. A loading budget, not a host limit: its `paths` load it in every
// session that edits a TypeScript file, so it changes only by the user's
// decision.
const languageLeafBudget = 4000

// languageLeafProblem returns the failure for a language leaf of size bytes,
// or "" when it fits.
func languageLeafProblem(label string, size int64) string {
	if size <= languageLeafBudget {
		return ""
	}
	return fmt.Sprintf("%s: %d bytes, over the %d-byte budget of a leaf every TypeScript session loads: move what only some edits need to a reference the leaf links with its read condition; raising the budget is the user's decision", label, size, languageLeafBudget)
}

// entrypointProblem returns the failure for a SKILL.md of size bytes, or ""
// when it fits.
func entrypointProblem(label string, size int64) string {
	if size <= skillEntrypointLimit {
		return ""
	}
	return fmt.Sprintf("%s: %d bytes, over the %d-byte skill entrypoint limit, past which Claude Code drops the end of an invoked skill after compaction: move a situational section to a reference linked with its read condition; do not trim a rule's text to fit", label, size, skillEntrypointLimit)
}

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
// of skillsDir that is over wholeReadLimit, and one when none is found. A file
// directly in a hive-core domain folder with the .md extension is a leaf,
// which leafProblems sizes, so it is skipped here and an oversize leaf is
// reported once; the principle index is not a leaf and is sized here.
func referenceSizeProblems(skillsDir string) []string {
	roots, problems := referenceRoots(skillsDir)
	coreDir := filepath.Join(skillsDir, coreSkillName)
	found := 0
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			isLeaf := filepath.Dir(path) == root && filepath.Dir(root) == coreDir &&
				filepath.Ext(path) == ".md" &&
				!(filepath.Base(root) == principlesDomain && entry.Name() == principleIndexFile)
			if isLeaf {
				return nil
			}
			found++
			info, err := entry.Info()
			if err != nil {
				return err
			}
			rel, _ := filepath.Rel(skillsDir, path)
			if problem := wholeReadProblem("content/skills/"+filepath.ToSlash(rel), info.Size()); problem != "" {
				problems = append(problems, problem)
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
	big := strings.Repeat("x", wholeReadLimit+1)
	writeFile(t, filepath.Join(skills, coreSkillName, "assets", "domains.json"), `{"domains": ["build", "principles"]}`)
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "checks.md"), "# Checks\n")
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "huge.md"), big)
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "checks", "own.md"), big)
	writeFile(t, filepath.Join(skills, coreSkillName, "build", "data.json"), big)
	writeFile(t, filepath.Join(skills, coreSkillName, "principles", "index.md"), big)
	writeFile(t, filepath.Join(skills, coreSkillName, "unlisted", "x.md"), big)
	writeFile(t, filepath.Join(skills, "flow-x", "references", "r.md"), big)
	got := strings.Join(referenceSizeProblems(skills), "\n")
	for _, want := range []string{
		"content/skills/hive-core/build/checks/own.md: ",
		"content/skills/hive-core/build/data.json: ",
		"content/skills/hive-core/principles/index.md: ",
		"content/skills/flow-x/references/r.md: ",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "unlisted") {
		t.Errorf("a folder outside domains.json is the leaf test's to report, got:\n%s", got)
	}
	if strings.Contains(got, "huge.md") {
		t.Errorf("a leaf is the leaf test's to size, got:\n%s", got)
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
		if problem := entrypointProblem(filepath.ToSlash(path[len("../../"):]), info.Size()); problem != "" {
			t.Error(problem)
		}
	}
}

func TestSizeLimitBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name    string
		problem func(string, int64) string
		limit   int64
		hint    string
	}{
		{"whole read", wholeReadProblem, wholeReadLimit, "split it by responsibility"},
		{"entrypoint", entrypointProblem, skillEntrypointLimit, "move a situational section to a reference"},
		{"language leaf", languageLeafProblem, languageLeafBudget, "raising the budget is the user's decision"},
	} {
		if got := tc.problem("f.md", tc.limit); got != "" {
			t.Errorf("%s: %d bytes must fit, got %q", tc.name, tc.limit, got)
		}
		got := tc.problem("f.md", tc.limit+1)
		if !strings.Contains(got, "f.md: ") || !strings.Contains(got, tc.hint) {
			t.Errorf("%s: %d bytes must fail naming the file and %q, got %q", tc.name, tc.limit+1, tc.hint, got)
		}
	}
}
