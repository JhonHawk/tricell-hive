package management

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// oldLeaf builds a rule-leaf folder name the way releases before hive-core
// named them, without writing the retired name as a literal.
func oldLeaf(name string) string { return "hive-rule-" + name }

const leafSkillText = "---\nname: %s\ndescription: A leaf.\ndisable-model-invocation: true\n---\nLeaf body.\n"

func leafFrontmatter(name string) string { return strings.Replace(leafSkillText, "%s", name, 1) }

// priorLeafInstall installs a release that holds one folder per name, each with
// a hidden SKILL.md, a nested reference and a script, and returns the paths the
// state records for each folder.
func priorLeafInstall(t *testing.T, o Options, names ...string) map[string][]string {
	t.Helper()
	for _, n := range names {
		base := filepath.Join(o.Source, "content", "skills", oldLeaf(n))
		put(t, filepath.Join(base, "SKILL.md"), leafFrontmatter(oldLeaf(n)))
		put(t, filepath.Join(base, "references", "extra.md"), "Extra.\n")
		put(t, filepath.Join(base, "scripts", "run.py"), "print(1)\n")
	}
	apply(t, plan(t, "install", o))
	recorded := map[string][]string{}
	for path := range stateFor(t, o).Records {
		for _, n := range names {
			if strings.Contains(path, string(filepath.Separator)+oldLeaf(n)+string(filepath.Separator)) || strings.HasSuffix(path, string(filepath.Separator)+oldLeaf(n)) {
				recorded[n] = append(recorded[n], path)
			}
		}
	}
	for _, n := range names {
		var hasPolicy, hasLink bool
		for _, p := range recorded[n] {
			hasPolicy = hasPolicy || strings.HasSuffix(p, "openai.yaml")
			hasLink = hasLink || !strings.Contains(p, filepath.Join(".agents", "skills"))
		}
		if len(recorded[n]) < 4 || !hasPolicy || !hasLink {
			t.Fatalf("prior state of %s lacks files, policy or catalogue link: %v", n, recorded[n])
		}
	}
	return recorded
}

func installedLeafDir(o Options, name string) string {
	return filepath.Join(o.Home, ".agents", "skills", oldLeaf(name))
}

// replaceWithCore swaps the source for a release that holds hive-core only.
func replaceWithCore(t *testing.T, o Options, global string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(o.Source, "content", "skills")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nPreserve evidence.\n")
	core := filepath.Join(o.Source, "content", "skills", "hive-core")
	put(t, filepath.Join(core, "SKILL.md"), "---\nname: hive-core\ndescription: Rule leaves.\ndisable-model-invocation: true\n---\nIndex.\n")
	put(t, filepath.Join(core, "references", "aa.md"), "# AA\n\nRead this.\n")
	put(t, filepath.Join(core, "references", "aa", "extra.md"), "Extra.\n")
	put(t, filepath.Join(core, "scripts", "checks", "run_check.py"), "print(1)\n")
	put(t, filepath.Join(o.Source, GlobalSource), global)
}

func plantFile(t *testing.T, path string) {
	t.Helper()
	put(t, path, "cache")
}

func TestUpdateRetiresLeafFoldersAndInstallsHiveCore(t *testing.T) {
	o := setup(t)
	recorded := priorLeafInstall(t, o, "aa", "bb")
	plantFile(t, filepath.Join(installedLeafDir(o, "aa"), "scripts", "__pycache__", "run.cpython-312.pyc"))

	replaceWithCore(t, o, "# Rules\nSee [aa](skill:hive-core/references/aa.md) and [extra](skill:hive-core/references/aa/extra.md).\n")
	out := apply(t, plan(t, "install", o))

	for n, paths := range recorded {
		for _, p := range paths {
			absent(t, p)
		}
		absent(t, installedLeafDir(o, n))
	}
	state := stateFor(t, o)
	policies := 0
	for path, r := range state.Records {
		if strings.Contains(path, "hive-rule-") {
			t.Errorf("state still records %s", path)
		}
		if strings.HasSuffix(path, "agents/openai.yaml") {
			policies++
			if want := filepath.Join(o.Home, ".agents", "skills", "hive-core", "agents", "openai.yaml"); path != want || r.Target.Source != "content/skills/hive-core/agents/openai.yaml" {
				t.Errorf("policy %s from %s, want only hive-core's", path, r.Target.Source)
			}
		}
	}
	if policies != 1 {
		t.Fatalf("%d openai.yaml records, want exactly one", policies)
	}
	core := filepath.Join(o.Home, ".agents", "skills", "hive-core")
	for _, f := range []string{"SKILL.md", "references/aa.md", "references/aa/extra.md", "scripts/checks/run_check.py", "agents/openai.yaml"} {
		if _, err := os.Lstat(filepath.Join(core, f)); err != nil {
			t.Errorf("hive-core lacks %s: %v", f, err)
		}
	}
	if strings.Contains(out, "kept") {
		t.Errorf("a clean retirement reported a kept directory: %s", out)
	}
}

func TestRetirementRemovesPycacheThatHoldsOnlyPyc(t *testing.T) {
	o := setup(t)
	priorLeafInstall(t, o, "aa")
	cache := filepath.Join(installedLeafDir(o, "aa"), "scripts", "__pycache__")
	plantFile(t, filepath.Join(cache, "run.cpython-312.pyc"))
	plantFile(t, filepath.Join(cache, "run.cpython-313.pyc"))

	replaceWithCore(t, o, "# Rules\n")
	apply(t, plan(t, "install", o))
	absent(t, installedLeafDir(o, "aa"))
}

func TestRetirementKeepsAndReportsPycacheWithForeignFile(t *testing.T) {
	o := setup(t)
	priorLeafInstall(t, o, "aa")
	cache := filepath.Join(installedLeafDir(o, "aa"), "scripts", "__pycache__")
	plantFile(t, filepath.Join(cache, "run.cpython-312.pyc"))
	plantFile(t, filepath.Join(cache, "notes.txt"))

	replaceWithCore(t, o, "# Rules\n")
	out := apply(t, plan(t, "install", o))
	if got := get(t, filepath.Join(cache, "notes.txt")); got != "cache" {
		t.Fatalf("foreign file changed: %q", got)
	}
	absent(t, filepath.Join(cache, "run.cpython-312.pyc"))
	if !strings.Contains(out, "kept") || !strings.Contains(out, cache) {
		t.Fatalf("apply output does not report the kept directory %s:\n%s", cache, out)
	}
}

func TestRetirementRejectsSymlinkedPycache(t *testing.T) {
	o := setup(t)
	priorLeafInstall(t, o, "aa")
	outside := filepath.Join(filepath.Dir(o.Home), "outside-cache")
	plantFile(t, filepath.Join(outside, "run.cpython-312.pyc"))
	cache := filepath.Join(installedLeafDir(o, "aa"), "scripts", "__pycache__")
	if err := os.Symlink(outside, cache); err != nil {
		t.Fatal(err)
	}

	replaceWithCore(t, o, "# Rules\n")
	out := apply(t, plan(t, "install", o))
	if got := get(t, filepath.Join(outside, "run.cpython-312.pyc")); got != "cache" {
		t.Fatalf("the link target was touched: %q", got)
	}
	if info, err := os.Lstat(cache); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the link was removed or replaced: %v", err)
	}
	if !strings.Contains(out, "link") || !strings.Contains(out, cache) {
		t.Fatalf("apply output does not report the rejected link %s:\n%s", cache, out)
	}
}

func TestPycacheOfASkillThatStaysIsLeftAlone(t *testing.T) {
	o := setup(t)
	priorLeafInstall(t, o, "aa")
	keep := filepath.Join(o.Home, ".agents", "skills", "workspace-conventions", "scripts", "__pycache__", "x.pyc")
	plantFile(t, keep)
	userDir := filepath.Join(o.Home, ".agents", "skills", "user-skill", "__pycache__", "y.pyc")
	plantFile(t, userDir)

	replaceWithCore(t, o, "# Rules\n")
	apply(t, plan(t, "install", o))
	for _, p := range []string{keep, userDir} {
		if _, err := os.Lstat(p); err != nil {
			t.Errorf("%s was removed: %v", p, err)
		}
	}
}

var skillLink = regexp.MustCompile(`skill:[A-Za-z0-9._/-]+`)

func TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink(t *testing.T) {
	o := setup(t)
	recorded := priorLeafInstall(t, o, "aa", "bb")
	plantFile(t, filepath.Join(installedLeafDir(o, "aa"), "scripts", "__pycache__", "run.cpython-312.pyc"))

	real, err := os.ReadFile("../../content/guidance/global.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(o.Source, "content", "skills")); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(filepath.Join(o.Source, "content", "skills"), os.DirFS("../../content/skills")); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, GlobalSource), string(real))
	p := plan(t, "install", o)
	apply(t, p)

	for _, paths := range recorded {
		for _, path := range paths {
			absent(t, path)
		}
	}
	policies, coreFiles := 0, 0
	core := filepath.Join(o.Home, ".agents", "skills", "hive-core") + string(filepath.Separator)
	for path := range stateFor(t, o).Records {
		if strings.Contains(path, "hive-rule-") {
			t.Errorf("state still records %s", path)
		}
		if strings.HasSuffix(path, "agents/openai.yaml") && strings.Contains(path, "hive-rule") {
			t.Errorf("retired policy %s", path)
		}
		if strings.HasPrefix(path, core) {
			coreFiles++
		}
		if strings.HasSuffix(path, "agents/openai.yaml") && strings.HasPrefix(path, core) {
			policies++
		}
	}
	if policies != 1 || coreFiles < 60 {
		t.Fatalf("hive-core policies = %d, files = %d", policies, coreFiles)
	}
	for _, f := range []string{"messages/questions.md", "messages/questions/unanswered.md", "principles/index.md", "principles/boundary-discipline.md", "assets/domains.json", "scripts/checks/run_check.py", "scripts/trackers/jira_rest.py", "scripts/trackers/work_time.py"} {
		if _, err := os.Lstat(filepath.Join(core, f)); err != nil {
			t.Errorf("hive-core lacks %s: %v", f, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(core, "references")); !os.IsNotExist(err) {
		t.Errorf("hive-core still has references/: %v", err)
	}

	links := skillLink.FindAllString(string(real), -1)
	if len(links) == 0 {
		t.Fatal("the global block links no skill")
	}
	for _, block := range []string{filepath.Join(o.Home, ".claude", "CLAUDE.md"), filepath.Join(o.Home, ".codex", "AGENTS.md")} {
		text := get(t, block)
		if strings.Contains(text, "](skill:") {
			t.Errorf("%s keeps an unrewritten skill: link", block)
		}
	}
	for _, l := range links {
		rel := strings.TrimPrefix(l, "skill:")
		if _, err := os.Stat(filepath.Join(o.Home, ".agents", "skills", filepath.FromSlash(rel))); err != nil {
			t.Errorf("global link %s does not resolve: %v", l, err)
		}
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("second update is not idempotent")
	}
}
