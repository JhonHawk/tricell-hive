package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
)

// Tests of the Project section: validation of the `## Hive` section of the
// repository's AGENTS.md. Repositories are real `git init` results in
// t.TempDir(); nothing here reads the developer's repositories.

// gitIn runs git in dir with the inherited repository variables removed, so a
// test started from a Git hook still builds its own repository.
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is not installed")
	}
	env, err := filteredGitEnv(gitPath)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(gitPath, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(env,
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com",
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, errOut.String())
	}
	return strings.TrimSpace(out.String())
}

// newProjectRepo creates a repository whose only branch is main, with one
// commit, and returns its (symlink-resolved) root.
func newProjectRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "init", "-q", "-b", "main")
	gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", "init")
	return dir
}

func writeAgents(t *testing.T, root, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// validHiveSection is a section with no findings for a repository made by
// newProjectRepo that also holds a specs directory.
const validHiveSection = `# Guidance

## Hive

- Project: demo
- Base branch: main
- Tracker: GitHub Issues · org/demo
- Specs: _support/openspec

## Other

- Not: part of the section
`

func projectRepoWithSpecs(t *testing.T, agents string) string {
	t.Helper()
	root := newProjectRepo(t)
	if err := os.MkdirAll(filepath.Join(root, "_support", "openspec"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeAgents(t, root, agents)
	return root
}

func projectText(t *testing.T, dir string) string {
	t.Helper()
	sec := collectProject(dir, doctorDeps{})
	if sec.Title != "Project" {
		t.Fatalf("title %q", sec.Title)
	}
	if sec.Err != "" {
		t.Fatalf("unexpected error: %s", sec.Err)
	}
	// The first line names what was checked, as the view's second row does.
	if len(sec.Lines) == 0 || !(strings.HasSuffix(sec.Lines[0], "AGENTS.md") || strings.HasPrefix(sec.Lines[0], "Directory: ")) {
		t.Fatalf("the first line does not name what was checked: %q", sec.Lines)
	}
	return strings.Join(sec.Lines[1:], "\n")
}

func TestHiveSectionOutsideGitLines(t *testing.T) {
	sec := collectProject(t.TempDir(), doctorDeps{})
	if sec.Title != "Project" || sec.Err != "" {
		t.Fatalf("%+v", sec)
	}
	if len(sec.Lines) != 3 || !strings.HasPrefix(sec.Lines[0], "Directory: ") || sec.Lines[1] != outsideGitLine || sec.Lines[2] != outsideGitLineTwo {
		t.Fatalf("lines: %q", sec.Lines)
	}
}

func TestHiveSectionMissingAgentsFile(t *testing.T) {
	root := newProjectRepo(t)
	text := projectText(t, root)
	mustContain(t, text, "AGENTS.md: not found")
	mustContain(t, text, "Add a ## Hive section with: Project, Base branch, Tracker, Specs")
	mustNotContain(t, text, "Valid")
}

func TestHiveSectionAgentsFileIsADirectory(t *testing.T) {
	root := newProjectRepo(t)
	if err := os.Mkdir(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustContain(t, projectText(t, root), "AGENTS.md: not a regular file")
}

func TestHiveSectionMissingAndDuplicatedSection(t *testing.T) {
	root := newProjectRepo(t)
	writeAgents(t, root, "# Guidance\n\n## Hive settings\n\n- Project: x\n")
	text := projectText(t, root)
	mustContain(t, text, "## Hive: section not found")
	mustContain(t, text, "Add a ## Hive section with: Project, Base branch, Tracker, Specs")
	mustNotContain(t, text, "Valid")

	writeAgents(t, root, "## Hive\n\n- Project: a\n\n## Hive\n\n- Project: b\n")
	text = projectText(t, root)
	mustContain(t, text, "## Hive: section appears 2 times")
	mustNotContain(t, text, "Valid")
}

func TestHiveSectionIgnoresHeadingsInsideCodeFences(t *testing.T) {
	root := projectRepoWithSpecs(t, "# Guidance\n\nExample:\n\n```markdown\n## Hive\n- Project: fake\n```\n\n"+validHiveSection)
	mustContain(t, projectText(t, root), "Valid")
}

func TestHiveSectionMissingAndEmptyRequiredValues(t *testing.T) {
	root := projectRepoWithSpecs(t, "## Hive\n\n- Project:\n- Base branch: main\n- Specs: _support/openspec\n")
	lines := strings.Split(projectText(t, root), "\n")
	var found []string
	for _, l := range lines {
		if strings.HasPrefix(l, "Tracker") || strings.HasPrefix(l, "Project") {
			found = append(found, l)
		}
	}
	if len(found) != 2 {
		t.Fatalf("want one line for Tracker and one for Project, got %q in %q", found, lines)
	}
	mustContain(t, strings.Join(found, "\n"), "Tracker: required value is missing", "Project: value is empty")
}

func TestHiveSectionSpecsMustBeAnExistingDirectory(t *testing.T) {
	root := projectRepoWithSpecs(t, strings.Replace(validHiveSection, "_support/openspec", "nowhere/openspec", 1))
	mustContain(t, projectText(t, root), "Specs: nowhere/openspec is not an existing directory")

	// A file is not a directory.
	writeAgents(t, root, strings.Replace(validHiveSection, "_support/openspec", "AGENTS.md", 1))
	mustContain(t, projectText(t, root), "Specs: AGENTS.md is not an existing directory")

	// The git URL after " · " is not part of the path, and a path outside the
	// repository resolves from the repository root.
	sibling := filepath.Join(filepath.Dir(root), filepath.Base(root)+"-specs", "openspec")
	if err := os.MkdirAll(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(filepath.Dir(sibling)) })
	writeAgents(t, root, strings.Replace(validHiveSection, "_support/openspec", "../"+filepath.Base(root)+"-specs/openspec · https://example.com/x.git", 1))
	mustContain(t, projectText(t, root), "Valid")
}

func TestHiveSectionBaseBranch(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	gitIn(t, root, "update-ref", "refs/remotes/origin/only-remote", "HEAD")
	gitIn(t, root, "branch", "feature/x")
	with := func(branch string) string {
		writeAgents(t, root, strings.Replace(validHiveSection, "Base branch: main", "Base branch: "+branch, 1))
		return projectText(t, root)
	}

	mustContain(t, with("main"), "Valid")
	mustContain(t, with("feature/x"), "Valid")
	mustContain(t, with("only-remote"), "Valid")
	// Only the first word is the branch.
	mustContain(t, with("main (until the rebuild lands)"), "Valid")

	text := with("missing")
	mustContain(t, text, "Base branch: missing is not a local branch or on origin")
	mustNotContain(t, text, "Valid")
	for _, bad := range []string{"a..b", "-x", "@{-1}", "a/../c", "x@{y}"} {
		got := with(bad)
		first := strings.Fields(bad)[0]
		mustContain(t, got, "Base branch: "+first+" is not a valid branch name")
		mustNotContain(t, got, "Valid")
	}
}

func TestHiveSectionBaseBranchThatCannotBeCheckedOnlyWarns(t *testing.T) {
	timeout := func(string, ...string) (string, error) { return "", errors.New("git timed out") }
	v := validateHiveSection(t.TempDir(), "## Hive\n- Project: p\n- Base branch: main\n- Tracker: t\n- Specs: .\n", timeout)
	mustContain(t, strings.Join(v.Lines, "\n"), "Base branch: main could not be checked")
	if v.Blocking {
		t.Errorf("a base branch that could not be checked blocks:\n%s", strings.Join(v.Lines, "\n"))
	}
}

// recordingGit is a gitRunner that records its calls and answers by script.
type recordingGit struct {
	mu    sync.Mutex
	calls [][]string
}

func (r *recordingGit) run(dir string, args ...string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, append([]string{dir}, args...))
	return "", nil
}

func TestHiveSectionRejectsDashAndReflogBranchesWithoutCallingGit(t *testing.T) {
	root := t.TempDir()
	for _, bad := range []string{"-x", "--help", "@{-1}", "main@{upstream}"} {
		rec := &recordingGit{}
		content := "## Hive\n- Project: p\n- Base branch: " + bad + "\n- Tracker: t\n- Specs: .\n"
		lines := validateHiveSection(root, content, rec.run).Lines
		mustContain(t, strings.Join(lines, "\n"), "Base branch: "+bad+" is not a valid branch name")
		if len(rec.calls) != 0 {
			t.Fatalf("%q called git: %q", bad, rec.calls)
		}
	}
	// A plausible name does reach git, with a fixed argument vector.
	rec := &recordingGit{}
	lines := validateHiveSection(root, "## Hive\n- Project: p\n- Base branch: main\n- Tracker: t\n- Specs: .\n", rec.run).Lines
	mustContain(t, strings.Join(lines, "\n"), "Valid")
	want := [][]string{
		{root, "check-ref-format", "--branch", "main"},
		{root, "show-ref", "--verify", "--quiet", "refs/heads/main"},
	}
	if len(rec.calls) != len(want) {
		t.Fatalf("calls: %q", rec.calls)
	}
	for i := range want {
		if strings.Join(rec.calls[i], "\x00") != strings.Join(want[i], "\x00") {
			t.Fatalf("call %d: %q, want %q", i, rec.calls[i], want[i])
		}
	}
}

func TestHiveSectionUnknownKeyAndOptionalValues(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Colour: blue\n- Specs:", 1))
	text := projectText(t, root)
	mustContain(t, text, "Colour: unknown key")
	mustNotContain(t, text, "Valid")
	// The known keys are named once, however many keys are unknown.
	known := "Known keys: Project, Base branch, Tracker, Other trackers, Specs, Environments, Delivery, Completion, Hive guidance"
	mustContain(t, text, known)
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Colour: blue\n- Shape: round\n- Specs:", 1))
	text = projectText(t, root)
	mustContain(t, text, "Colour: unknown key", "Shape: unknown key")
	if strings.Count(text, "Known keys:") != 1 {
		t.Fatalf("the known keys must be named once:\n%s", text)
	}
	mustNotContain(t, projectText(t, projectRepoWithSpecs(t, validHiveSection)), "Known keys:")
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Colour: blue\n- Specs:", 1))

	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Delivery: pr\n- Hive guidance: optional\n- Specs:", 1))
	text = projectText(t, root)
	mustContain(t, text, `Delivery: expected "direct-base", found "pr"`, `Hive guidance: expected "required", found "optional"`)
	mustNotContain(t, text, "Valid")

	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Delivery: direct-base\n- Hive guidance: required\n- Environments: a → b\n- Completion: To Deploy on origin/a · Done on origin/b\n- Specs:", 1))
	mustContain(t, projectText(t, root), "Valid", "Delivery: direct-base", "Hive guidance: required", "Completion: To Deploy on origin/a · Done on origin/b")
}

func TestDoctorProjectDeliveryAcceptsOnlyDirectBase(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Delivery: direct-base\n- Specs:", 1))
	text := projectText(t, root)
	mustContain(t, text, "Valid", "Delivery: direct-base")
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Delivery: pr\n- Specs:", 1))
	text = projectText(t, root)
	mustContain(t, text, `Delivery: expected "direct-base", found "pr"`)
	mustNotContain(t, text, "Valid", "retired")
}

// A retired key or Delivery value warns, names how to remove it, and is not an
// unknown key. The value of a retired key is never judged.
func TestDoctorProjectWarnsOnEachRetiredKeyWithoutBlocking(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	cases := []struct{ line, key, subject string }{
		{"Review: /code-review", "Review", "Review:"},
		{"UI verification: full", "UI verification", "UI verification:"},
		{"UI verification: partial", "UI verification", "UI verification:"},
		{"Delivery: automatic", "Delivery", "Delivery: automatic is"},
		{"Delivery: interactive", "Delivery", "Delivery: interactive is"},
	}
	for _, c := range cases {
		t.Run(c.line, func(t *testing.T) {
			writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- "+c.line+"\n- Specs:", 1))
			sec := collectProject(root, doctorDeps{})
			text := projectText(t, root)
			mustContain(t, text, c.subject+" retired; Hive asks it each session, and the old value only informs the recommended option — remove it with hive project set --unset "+c.key+" or run harness-audit")
			mustNotContain(t, text, "unknown key", "Known keys:", "Valid")
			if sec.Blocking {
				t.Fatalf("a retired setting must not block:\n%s", text)
			}
		})
	}
}

func TestDoctorProjectChecksTheShapeOfCompletion(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	for value, ok := range map[string]bool{
		"To Deploy on origin/development · Done on origin/qa": true,
		"Merged on main · Released on origin/master":          true,
		"Done":                          false,
		"To Deploy on main, Done on qa": false,
		"":                              false,
	} {
		writeAgents(t, root, strings.Replace(validHiveSection, "- Specs:", "- Completion: "+value+"\n- Specs:", 1))
		sec := collectProject(root, doctorDeps{})
		text := projectText(t, root)
		mustNotContain(t, text, "unknown key")
		if sec.Blocking {
			t.Fatalf("Completion %q must never block:\n%s", value, text)
		}
		if ok {
			mustContain(t, text, "Valid")
		} else {
			mustContain(t, text, "Completion: expected \"<state> on <ref> · <state> on <ref>\"")
			mustNotContain(t, text, "Valid")
		}
	}
}

func TestHiveSectionInheritedGitDirDoesNotChangeTheResult(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	other := newProjectRepo(t)
	gitIn(t, other, "branch", "-m", "main", "elsewhere")
	before := projectText(t, root)
	mustContain(t, before, "Valid")

	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	if after := projectText(t, root); after != before {
		t.Fatalf("result changed under GIT_DIR:\n%s\nvs\n%s", after, before)
	}
	// Outside a repository the inherited variables do not make one appear.
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	sec := collectProject(t.TempDir(), doctorDeps{})
	if len(sec.Lines) != 3 || sec.Lines[1] != outsideGitLine {
		t.Fatalf("lines: %q", sec.Lines)
	}
}

func TestHiveSectionValidListsValuesAndWritesNothing(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	path := filepath.Join(root, "AGENTS.md")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	filesBefore := collectFiles(t, root)
	text := projectText(t, root)
	if !strings.HasPrefix(text, "Valid\n") {
		t.Fatalf("the verdict must come first:\n%s", text)
	}
	mustContain(t, text, "Project: demo", "Base branch: main", "Tracker: GitHub Issues · org/demo", "Specs: _support/openspec")
	mustNotContain(t, text, "Not: part of the section")
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("AGENTS.md changed")
	}
	if fmt.Sprint(filesBefore) != fmt.Sprint(collectFiles(t, root)) {
		t.Fatal("the repository files changed")
	}
}

func TestHiveSectionFindingsAlsoListTheValuesRead(t *testing.T) {
	root := projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Tracker: GitHub Issues · org/demo\n", "", 1))
	text := projectText(t, root)
	mustContain(t, text, "Tracker: required value is missing", "Values read:", "Project: demo")
}

func TestHiveSectionSanitizesFileText(t *testing.T) {
	agents := "## Hive\n- Project: \x1b[31mred\x1b[0m\x07 demo\n- Base branch: main\n- Tracker: t\n- Specs: _support/\x1b]0;title\x07openspec\n- \x1b[2Jevil: 1\n"
	root := projectRepoWithSpecs(t, agents)
	text := projectText(t, root)
	if strings.ContainsRune(text, 0x1b) || strings.ContainsRune(text, 0x07) {
		t.Fatalf("control characters reached the output: %q", text)
	}
	mustContain(t, text, "Project: red demo")
}

func TestHiveSectionReadsAtMostOneMiB(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	big := validHiveSection + strings.Repeat("x", (1<<20)+10) + "\n"
	writeAgents(t, root, big)
	text := projectText(t, root)
	mustContain(t, text, "AGENTS.md: larger than 1 MiB; only the first 1 MiB was read")

	// A section that sits past the limit is not found.
	writeAgents(t, root, strings.Repeat("x\n", 1<<19)+validHiveSection)
	mustContain(t, projectText(t, root), "## Hive: section not found")
}

func TestHiveSectionProjectFlagUsesThatRepository(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	// The current directory of the test is the package directory, another
	// repository; the argument decides.
	mustContain(t, projectText(t, root), "Valid")
	sub := filepath.Join(root, "pkg", "deep")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	mustContain(t, projectText(t, sub), "Valid")
}

func TestHiveSectionGitNotFoundIsAnErrorNotAFinding(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	sec := collectProject(t.TempDir(), doctorDeps{})
	if sec.Title != "Project" || sec.Err == "" {
		t.Fatalf("%+v", sec)
	}
}

// ---------------------------------------------------------------------------
// AC10: the keys the validator knows and the guidance stay in step.
// ---------------------------------------------------------------------------

// projectSettingsSection returns the canonical settings declaration, split into
// prose lines and lines inside fenced code blocks. It lives in the global
// guidance, which every session loads.
func projectSettingsSection(t *testing.T) (prose, code []string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "content", "guidance", "global.md"))
	if err != nil {
		t.Fatal(err)
	}
	in, fenced := false, false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case !fenced && strings.HasPrefix(line, "## "):
			in = strings.TrimSpace(strings.TrimPrefix(line, "## ")) == "Project settings"
			continue
		case !in:
			continue
		case strings.HasPrefix(trimmed, "```"):
			fenced = !fenced
		case fenced:
			code = append(code, trimmed)
		default:
			prose = append(prose, line)
		}
	}
	if len(prose) == 0 || len(code) == 0 {
		t.Fatalf("the canonical settings declaration was not found (prose %d lines, code %d lines)", len(prose), len(code))
	}
	return prose, code
}

func TestHiveSettingKeysAppearInTheGuidance(t *testing.T) {
	prose, _ := projectSettingsSection(t)
	var quoted []string
	for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(strings.Join(prose, "\n"), -1) {
		quoted = append(quoted, m[1])
	}
	for _, k := range hiveSettingKeys {
		ok := false
		for _, q := range quoted {
			if strings.HasPrefix(q, k.Name) {
				ok = true
				break
			}
		}
		if !ok {
			t.Errorf("key %q is not the start of any backtick-quoted text in the canonical «Project settings» declaration of global.md", k.Name)
		}
	}
}

func TestHiveSettingKeysCoverTheGuidanceExample(t *testing.T) {
	_, code := projectSettingsSection(t)
	seen := 0
	for _, line := range code {
		rest, ok := strings.CutPrefix(line, "- ")
		if !ok {
			continue
		}
		name, _, ok := strings.Cut(rest, ":")
		if !ok {
			continue
		}
		seen++
		if !knownHiveKey(strings.TrimSpace(name)) {
			t.Errorf("the example in global.md «Project settings» uses key %q, which hiveSettingKeys does not know", strings.TrimSpace(name))
		}
	}
	if seen == 0 {
		t.Fatal("no `- Key: value` line found in the example code block")
	}
}

func TestHiveSettingKeysRequiredSet(t *testing.T) {
	var required []string
	for _, k := range hiveSettingKeys {
		if k.Required {
			required = append(required, k.Name)
		}
	}
	if got := strings.Join(required, ","); got != "Project,Base branch,Tracker,Specs" {
		t.Fatalf("required keys: %s", got)
	}
	for _, name := range []string{"Other trackers", "Environments", "Delivery", "Completion", "Hive guidance"} {
		if !knownHiveKey(name) {
			t.Errorf("optional key %q is missing", name)
		}
	}
	for _, name := range []string{"Review", "UI verification"} {
		if knownHiveKey(name) {
			t.Errorf("retired key %q must not be a known key", name)
		}
	}
}

// The retired keys and Delivery values the manager warns about are the ones
// the global guidance's «Project settings» names on its «Retired keys:» line.
func TestRetiredHiveKeysMatchTheGuidance(t *testing.T) {
	prose, _ := projectSettingsSection(t)
	var line string
	for _, l := range prose {
		if strings.HasPrefix(l, "Retired keys:") {
			line = l
		}
	}
	if line == "" {
		t.Fatal("global.md «Project settings» has no line starting \"Retired keys:\"")
	}
	var sentence string
	sentence, _, _ = strings.Cut(line, ".")
	var got []string
	for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(sentence, -1) {
		got = append(got, m[1])
	}
	want := append(append(append([]string{}, retiredHiveKeys...), "Delivery"), retiredDeliveryValues...)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("guidance names %q as retired, the manager %q", got, want)
	}
}

func TestCheckProjectReportsAMissingDirectoryWithoutCallingGit(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-dir")
	rec := &recordingGit{}
	sec := checkProject(missing, rec.run).sectionWithLocation()
	if len(rec.calls) != 0 {
		t.Fatalf("git was called for a missing directory: %q", rec.calls)
	}
	if sec.Err != "" || len(sec.Lines) != 1 || sec.Lines[0] != "Directory not found: "+missing {
		t.Fatalf("section = %+v", sec)
	}
	mustNotContain(t, strings.Join(sec.Lines, "\n"), outsideGitLine)
}

func TestCheckProjectReportsAFileAsNotADirectoryWithoutCallingGit(t *testing.T) {
	file := filepath.Join(t.TempDir(), "AGENTS.md")
	if err := os.WriteFile(file, []byte("## Hive\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	rec := &recordingGit{}
	sec := checkProject(file, rec.run).sectionWithLocation()
	if len(rec.calls) != 0 {
		t.Fatalf("git was called for a file: %q", rec.calls)
	}
	if sec.Err != "" || len(sec.Lines) != 1 || sec.Lines[0] != "Not a directory: "+file {
		t.Fatalf("section = %+v", sec)
	}
	mustNotContain(t, strings.Join(sec.Lines, "\n"), outsideGitLine)
}

// writeChangeFile writes content under <root>/_support/openspec/changes/<rel>.
func writeChangeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, "_support", "openspec", "changes", filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorProjectChangeRecordChecksWarnWithoutBlocking(t *testing.T) {
	proposal := "# P\n\n- AC1 — one\n- AC2 — two\n- AC3 — three\n"
	cases := []struct {
		name    string
		files   map[string]string
		want    []string
		notWant []string
	}{
		{"covered set", map[string]string{"good-change/proposal.md": proposal, "good-change/tasks.md": "**Closes:** AC1, AC2\n**Closes:** AC3\n"}, nil, []string{"Closes", "kebab"}},
		{"missing AC", map[string]string{"good-change/proposal.md": proposal, "good-change/tasks.md": "**Closes:** AC1, AC3\n"},
			[]string{`Specs: change "good-change": AC2 in proposal.md not named by any **Closes:** line in tasks.md`}, nil},
		{"closes with extra words and ranges", map[string]string{"good-change/proposal.md": proposal, "good-change/tasks.md": "**Closes:** AC1 (forma), AC2–AC3 (UI)\n"}, nil, []string{"Closes"}},
		{"slash and short range", map[string]string{"good-change/proposal.md": proposal, "good-change/tasks.md": "**Closes:** AC1/AC2-3\n"}, nil, []string{"Closes"}},
		{"folder without tasks", map[string]string{"good-change/proposal.md": proposal}, nil, []string{"Closes"}},
		{"non-kebab folder", map[string]string{"Bad_Change/notes.md": "x"},
			[]string{`Specs: change folder "Bad_Change" is not English kebab-case`}, nil},
		{"bad archive name", map[string]string{"archive/2026-09-25-ok-name/x.md": "x", "archive/old-thing/x.md": "x", "archive/2026-9-1-late/x.md": "x"},
			[]string{`Specs: archived change "old-thing" does not match YYYY-MM-DD-<kebab-case>`, `Specs: archived change "2026-9-1-late" does not match`},
			[]string{`"2026-09-25-ok-name"`, "kebab-case (lowercase"}},
		{"archive itself is not a change", map[string]string{"archive/2026-09-25-ok/x.md": "x"}, nil, []string{"kebab", "archived"}},
		{"no changes folder", nil, nil, []string{"kebab", "Closes", "archived"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := projectRepoWithSpecs(t, validHiveSection)
			for rel, content := range c.files {
				writeChangeFile(t, root, rel, content)
			}
			sec := collectProject(root, doctorDeps{})
			text := projectText(t, root)
			if sec.Blocking {
				t.Fatalf("change record warnings must not block:\n%s", text)
			}
			mustContain(t, text, c.want...)
			mustNotContain(t, text, c.notWant...)
			if len(c.want) == 0 {
				mustContain(t, text, "Valid")
			} else {
				mustNotContain(t, text, "Valid")
			}
		})
	}
}

func TestDoctorProjectChangeChecksIgnoreAMissingSpecsDirectory(t *testing.T) {
	root := newProjectRepo(t)
	writeAgents(t, root, validHiveSection) // Specs points at a directory that does not exist
	text := projectText(t, root)
	mustContain(t, text, "Specs: _support/openspec is not an existing directory")
	mustNotContain(t, text, "kebab", "Closes", "archived")
}

// withLines puts extra `- Key: value` lines into the valid section.
func withLines(lines ...string) string {
	return strings.Replace(validHiveSection, "- Specs:", strings.Join(lines, "\n")+"\n- Specs:", 1)
}

func TestDoctorProjectAcceptsOtherTrackersAndPrefixedCompletion(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	for name, lines := range map[string][]string{
		"one tracker":        {"- Other trackers: Jira · example.atlassian.net · keys XYZ", "- Completion: XYZ: QA on qa · Done on production"},
		"two trackers":       {"- Other trackers: Jira · example.atlassian.net · keys XYZ, ABC1; Linear · team Example · key QRS", "- Completion: XYZ, ABC1: QA on qa · Done on production; QRS: Merged on main · Done on production"},
		"primary and other":  {"- Other trackers: Jira · example.atlassian.net · keys XYZ", "- Completion: To Deploy on origin/a · Done on origin/b; XYZ: QA on qa · Done on production"},
		"trackers only":      {"- Other trackers: Jira · example.atlassian.net · keys XYZ"},
		"underscore prefix":  {"- Other trackers: Jira · example.atlassian.net · keys AB_C, X_1", "- Completion: AB_C, X_1: QA on qa · Done on production"},
		"tight separator":    {"- Other trackers: Jira · example.atlassian.net · keys XYZ;Linear · team Example · key QRS", "- Completion: XYZ: QA on qa · Done on production;QRS: A on a · B on b"},
		"trailing separator": {"- Other trackers: Jira · example.atlassian.net · keys XYZ; ", "- Completion: To Deploy on origin/a · Done on origin/b; "},
		"doubled separator":  {"- Other trackers: Jira · example.atlassian.net · keys XYZ;; ;Linear · team Example · key QRS", "- Completion: XYZ: QA on qa · Done on production;;"},
	} {
		t.Run(name, func(t *testing.T) {
			writeAgents(t, root, withLines(lines...))
			sec := collectProject(root, doctorDeps{})
			text := projectText(t, root)
			mustContain(t, text, "Valid")
			mustNotContain(t, text, "unknown key", "Known keys:", "expected")
			if sec.Blocking {
				t.Fatalf("must not block:\n%s", text)
			}
		})
	}
}

func TestDoctorProjectWarnsOnMalformedOtherTrackersAndCompletionValues(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	cases := []struct {
		name  string
		lines []string
		want  string
	}{
		{"tracker without keys", []string{"- Other trackers: Jira · example.atlassian.net"}, `Other trackers: expected "<tool> · <location> · keys <PREFIX>[, <PREFIX>]", found "Jira · example.atlassian.net"`},
		{"keys with a lowercase prefix", []string{"- Other trackers: Jira · example.atlassian.net · keys xyz"}, `Other trackers: expected`},
		{"key with two prefixes", []string{"- Other trackers: Jira · example.atlassian.net · key XYZ, ABC"}, `Other trackers: expected`},
		{"repeated prefix", []string{"- Other trackers: Jira · example.atlassian.net · keys XYZ; Linear · team Example · keys XYZ"}, `Other trackers: prefix XYZ is declared more than once`},
		{"undeclared completion prefix", []string{"- Other trackers: Jira · example.atlassian.net · keys XYZ", "- Completion: ABC: QA on qa · Done on production"}, `Completion: prefix ABC is not declared in Other trackers`},
		{"repeated completion prefix", []string{"- Other trackers: Jira · example.atlassian.net · keys XYZ", "- Completion: XYZ: QA on qa · Done on production; XYZ: QA on a · Done on b"}, `Completion: prefix XYZ appears more than once`},
		{"malformed value after the separator", []string{"- Other trackers: Jira · example.atlassian.net · keys XYZ", "- Completion: Done on qa · Done on production; XYZ: Done"}, `Completion: expected "<state> on <ref> · <state> on <ref>", found "XYZ: Done"`},
		{"tight separator is still two entries", []string{"- Other trackers: Jira · example.atlassian.net · keys XYZ;Linear · team Example · keys XYZ"}, `Other trackers: prefix XYZ is declared more than once`},
		{"tight separator completion", []string{"- Completion: A on a · B on b;C on c · D on d"}, `Completion: at most one value may have no prefix`},
		{"two values without a prefix", []string{"- Completion: A on a · B on b; C on c · D on d"}, `Completion: at most one value may have no prefix`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			writeAgents(t, root, withLines(c.lines...))
			sec := collectProject(root, doctorDeps{})
			text := projectText(t, root)
			mustContain(t, text, c.want)
			mustNotContain(t, text, "Valid", "unknown key")
			if sec.Blocking {
				t.Fatalf("a finding on these keys must never block:\n%s", text)
			}
		})
	}
}
