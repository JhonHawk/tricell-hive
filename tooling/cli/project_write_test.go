package main

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// Tests of the AGENTS.md write path. Repositories come from `git init` in
// t.TempDir(); nothing here touches a real repository or configuration.

func hashOf(t *testing.T, path string) [32]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

func mustEdit(t *testing.T, in string, set []hiveItem, unset []string) string {
	t.Helper()
	out, err := editHiveSection([]byte(in), set, unset)
	if err != nil {
		t.Fatalf("editHiveSection: %v", err)
	}
	return string(out)
}

func TestHiveSectionEditTable(t *testing.T) {
	const base = "# Doc\n\nProse stays.\n\n## Hive\n\n- Project: demo\n- Base branch: main\n- Mystery: kept\n\n## Other\n\n- Not: part\n"
	cases := []struct {
		name  string
		in    string
		set   []hiveItem
		unset []string
		want  string
	}{
		{"replace", base, []hiveItem{{"Base branch", "dev"}}, nil,
			"# Doc\n\nProse stays.\n\n## Hive\n\n- Project: demo\n- Base branch: dev\n- Mystery: kept\n\n## Other\n\n- Not: part\n"},
		{"add after last item in key order", base, []hiveItem{{"Specs", "s"}, {"Tracker", "t"}}, nil,
			"# Doc\n\nProse stays.\n\n## Hive\n\n- Project: demo\n- Base branch: main\n- Mystery: kept\n- Tracker: t\n- Specs: s\n\n## Other\n\n- Not: part\n"},
		{"unset", base, nil, []string{"Base branch"},
			"# Doc\n\nProse stays.\n\n## Hive\n\n- Project: demo\n- Mystery: kept\n\n## Other\n\n- Not: part\n"},
		{"unset absent key is a no-op", base, nil, []string{"Review"}, base},
		{"code fence with a heading inside is ignored",
			"```\n## Hive\n- Project: fake\n```\n\n## Hive\n\n- Project: a\n",
			[]hiveItem{{"Project", "b"}}, nil,
			"```\n## Hive\n- Project: fake\n```\n\n## Hive\n\n- Project: b\n"},
		{"crlf kept", "# T\r\n\r\n## Hive\r\n\r\n- Project: a\r\n",
			[]hiveItem{{"Project", "b"}, {"Specs", "s"}}, nil,
			"# T\r\n\r\n## Hive\r\n\r\n- Project: b\r\n- Specs: s\r\n"},
		{"no section appends", "# T\n\nbody\n", []hiveItem{{"Specs", "s"}, {"Project", "p"}}, nil,
			"# T\n\nbody\n\n## Hive\n\n- Project: p\n- Specs: s\n"},
		{"no section, no final newline", "# T\nbody", []hiveItem{{"Project", "p"}}, nil,
			"# T\nbody\n\n## Hive\n\n- Project: p\n"},
		{"no section crlf", "# T\r\nbody\r\n", []hiveItem{{"Project", "p"}}, nil,
			"# T\r\nbody\r\n\r\n## Hive\r\n\r\n- Project: p\r\n"},
		{"empty file", "", []hiveItem{{"Project", "p"}}, nil, "## Hive\n\n- Project: p\n"},
		{"empty section gets items after the blank line", "## Hive\n\n## Next\n", []hiveItem{{"Project", "p"}}, nil,
			"## Hive\n\n- Project: p\n\n## Next\n"},
		{"prose and an html comment inside the section are kept",
			"## Hive\n\nWhy these values:\n<!-- keep: reviewed 2026-10 -->\n- Project: a\nMore prose.\n- Base branch: main\n<!-- tail -->\n\n## Next\n",
			[]hiveItem{{"Project", "b"}, {"Specs", "s"}}, nil,
			"## Hive\n\nWhy these values:\n<!-- keep: reviewed 2026-10 -->\n- Project: b\nMore prose.\n- Base branch: main\n- Specs: s\n<!-- tail -->\n\n## Next\n"},
		{"empty section as the last line keeps the final newline", "# T\n\n## Hive\n", []hiveItem{{"Project", "p"}}, nil,
			"# T\n\n## Hive\n\n- Project: p\n"},
		{"empty section as the last line, crlf", "# T\r\n\r\n## Hive\r\n", []hiveItem{{"Project", "p"}, {"Specs", "s"}}, nil,
			"# T\r\n\r\n## Hive\r\n\r\n- Project: p\r\n- Specs: s\r\n"},
		{"heading without a newline gets its blank line and no newline is added", "# T\n\n## Hive", []hiveItem{{"Project", "p"}}, nil,
			"# T\n\n## Hive\n\n- Project: p"},
		{"heading without a newline, crlf: no bare CR at the end", "# T\r\n\r\n## Hive", []hiveItem{{"Project", "p"}}, nil,
			"# T\r\n\r\n## Hive\r\n\r\n- Project: p"},
		{"heading and a blank last line", "## Hive\n\n", []hiveItem{{"Project", "p"}}, nil,
			"## Hive\n\n- Project: p\n"},
		{"section at end without newline", "## Hive\n- Project: a", []hiveItem{{"Specs", "s"}}, nil,
			"## Hive\n- Project: a\n- Specs: s"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mustEdit(t, c.in, c.set, c.unset); got != c.want {
				t.Fatalf("got:\n%q\nwant:\n%q", got, c.want)
			}
		})
	}
}

func TestHiveSectionEditRejections(t *testing.T) {
	cases := []struct {
		name, in, want string
		set            []hiveItem
	}{
		{"duplicate section", "## Hive\n- Project: a\n\n## Hive\n- Project: b\n", "appears 2 times", []hiveItem{{"Project", "c"}}},
		{"repeated edited key", "## Hive\n- Project: a\n- Project: b\n", "Project appears more than once", []hiveItem{{"Project", "c"}}},
		{"newline in value", "## Hive\n- Project: a\n", "line break", []hiveItem{{"Project", "a\n- Delivery: x"}}},
		{"colon in key", "## Hive\n- Project: a\n", "invalid key", []hiveItem{{"A: b", "x"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := editHiveSection([]byte(c.in), c.set, nil)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %v, want it to contain %q", err, c.want)
			}
		})
	}
	// A repeated key that is not edited is left alone.
	if _, err := editHiveSection([]byte("## Hive\n- Review: a\n- Review: b\n- Project: p\n"), []hiveItem{{"Project", "q"}}, nil); err != nil {
		t.Fatalf("repeated key not edited: %v", err)
	}
}

func TestHiveTextSeparatesBlockingFromWarnings(t *testing.T) {
	root := projectRepoWithSpecs(t, "")
	git := newGitRunner()
	items := []hiveItem{{"Project", "p"}, {"Base branch", "nope"}, {"Tracker", "t"}, {"Specs", "missing/dir"}, {"Extra", "x"}}
	blocking, warnings := checkHiveText(items, root, git)
	if len(blocking) != 0 {
		t.Fatalf("blocking %q", blocking)
	}
	joined := strings.Join(warnings, "\n")
	for _, want := range []string{"Base branch: nope is not a local branch", "Specs: missing/dir is not an existing directory", "Extra: unknown key"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings %q lack %q", warnings, want)
		}
	}
	blocking, _ = checkHiveText([]hiveItem{{"Project", "p"}, {"Base branch", "-x"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}, {"Delivery", "pr"}, {"Hive guidance", "optional"}}, root, git)
	joined = strings.Join(blocking, "\n")
	for _, want := range []string{"Base branch: -x is not a valid branch name", "Delivery: expected", "Hive guidance: expected"} {
		if !strings.Contains(joined, want) {
			t.Errorf("blocking %q lacks %q", blocking, want)
		}
	}
	blocking, _ = checkHiveText(nil, root, git)
	if len(blocking) != 4 {
		t.Errorf("empty items: blocking %q, want the four required keys", blocking)
	}
}

// prepare runs prepareProjectEdit for the repository root.
func prepare(t *testing.T, root string, set []hiveItem, unset []string) (projectEdit, error) {
	t.Helper()
	return prepareProjectEdit(root, set, unset, newGitRunner())
}

var fullSet = []hiveItem{{"Project", "demo"}, {"Base branch", "main"}, {"Tracker", "GitHub Issues · o/demo"}, {"Specs", "_support/openspec"}}

func TestProjectWriteRejections(t *testing.T) {
	long := strings.Repeat("a", hiveFileLimit+1)
	cases := []struct {
		name  string
		setup func(t *testing.T) (project string, file string) // file "" when there is none to hash
		set   []hiveItem
		unset []string
		want  string
	}{
		{"outside a repository", func(t *testing.T) (string, string) { return t.TempDir(), "" }, fullSet, nil, "Not inside a Git repository"},
		{"AGENTS.md is a directory", func(t *testing.T) (string, string) {
			r := projectRepoWithSpecs(t, "")
			os.Remove(filepath.Join(r, "AGENTS.md"))
			os.Mkdir(filepath.Join(r, "AGENTS.md"), 0o755)
			return r, ""
		}, fullSet, nil, "not a regular file"},
		{"AGENTS.md is a symlink", func(t *testing.T) (string, string) {
			r := projectRepoWithSpecs(t, "")
			os.Remove(filepath.Join(r, "AGENTS.md"))
			target := filepath.Join(r, "real.md")
			os.WriteFile(target, []byte("# x\n"), 0o644)
			os.Symlink(target, filepath.Join(r, "AGENTS.md"))
			return r, target
		}, fullSet, nil, "symbolic link"},
		{"larger than 1 MiB", func(t *testing.T) (string, string) {
			r := projectRepoWithSpecs(t, long)
			return r, filepath.Join(r, "AGENTS.md")
		}, fullSet, nil, "larger than 1 MiB"},
		{"required missing", nil, []hiveItem{{"Project", "p"}}, nil, "Base branch: required value is missing"},
		{"required empty", nil, []hiveItem{{"Project", ""}, {"Base branch", "main"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}}, nil, "Project: value is empty"},
		{"value with newline", nil, []hiveItem{{"Project", "a\nb"}, {"Base branch", "main"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}}, nil, "Project: value has a line break or control character"},
		{"value with escape sequence", nil, []hiveItem{{"Project", "a\x1b[31m"}, {"Base branch", "main"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}}, nil, "Project: value has a line break or control character"},
		{"value with right-to-left override", nil, []hiveItem{{"Project", "a‮b"}, {"Base branch", "main"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}}, nil, "Project: value has a line break or control character"},
		{"Delivery pr", nil, append(append([]hiveItem{}, fullSet...), hiveItem{"Delivery", "pr"}), nil, "Delivery: expected"},
		{"Hive guidance optional", nil, append(append([]hiveItem{}, fullSet...), hiveItem{"Hive guidance", "optional"}), nil, "Hive guidance: expected"},
		{"Base branch -x", nil, []hiveItem{{"Project", "p"}, {"Base branch", "-x"}, {"Tracker", "t"}, {"Specs", "_support/openspec"}}, nil, "is not a valid branch name"},
		{"unknown key in set", nil, append(append([]hiveItem{}, fullSet...), hiveItem{"Mystery", "x"}), nil, "Mystery: unknown key; known keys are"},
		{"unset a required key", func(t *testing.T) (string, string) {
			r := projectRepoWithSpecs(t, validHiveSection)
			return r, filepath.Join(r, "AGENTS.md")
		}, nil, []string{"Project"}, "Project: required key cannot be unset"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var project, file string
			if c.setup != nil {
				project, file = c.setup(t)
			} else {
				project = projectRepoWithSpecs(t, "# Existing\n")
				file = filepath.Join(project, "AGENTS.md")
			}
			var before [32]byte
			if file != "" {
				before = hashOf(t, file)
			}
			_, err := prepare(t, project, c.set, c.unset)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %v, want it to contain %q", err, c.want)
			}
			if file != "" && hashOf(t, file) != before {
				t.Fatal("the file changed")
			}
		})
	}
}

func TestProjectWriteRefusesAnExistingValueThatSanitizingWouldChange(t *testing.T) {
	cases := map[string]string{
		"escape sequence":    "- Tracker: t\x1b[31mred\n",
		"right-to-left mark": "- Tracker: t\u202ered\n",
		"control in the key": "- Tracker: t\n- Mystery\x1b[1m: x\n",
	}
	for name, line := range cases {
		t.Run(name, func(t *testing.T) {
			root := projectRepoWithSpecs(t, "# T\n\n## Hive\n\n- Project: p\n- Base branch: main\n"+line+"- Specs: _support/openspec\n")
			path := filepath.Join(root, "AGENTS.md")
			h := hashOf(t, path)
			_, err := prepare(t, root, []hiveItem{{"Project", "q"}}, nil)
			if err == nil || !strings.Contains(err.Error(), "existing value has a control or hidden character") {
				t.Fatalf("error %v", err)
			}
			if strings.Contains(err.Error(), "\x1b") || strings.Contains(err.Error(), "\u202e") {
				t.Fatalf("the error leaks the raw characters: %q", err)
			}
			if hashOf(t, path) != h {
				t.Fatal("the file changed")
			}
		})
	}
	// The doctor still only reads and shows it cleaned.
	root := projectRepoWithSpecs(t, "## Hive\n- Project: p\n- Tracker: t\x1b[31mred\n")
	if out := strings.Join(validateHiveSection(root, "## Hive\n- Tracker: t\x1b[31mred\n", newGitRunner()), "\n"); strings.Contains(out, "hidden character") {
		t.Fatalf("doctor output changed:\n%s", out)
	}
}

func TestProjectWriteWarningsDoNotBlock(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n\n## Hive\n\n- Project: p\n- Base branch: ghost\n- Tracker: t\n- Specs: nowhere\n- Mystery: stays\n")
	e, err := prepare(t, root, []hiveItem{{"Project", "q"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(e.Warnings, "\n")
	for _, want := range []string{"Base branch: ghost is not a local branch", "Specs: nowhere is not an existing directory", "Mystery: unknown key"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings %q lack %q", e.Warnings, want)
		}
	}
	if err := writeProjectFile(e.File, e.Before, e.Existed, e.After); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(e.File)
	if !strings.Contains(string(got), "- Project: q\n") || !strings.Contains(string(got), "- Mystery: stays\n") {
		t.Fatalf("file:\n%s", got)
	}
}

func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func noTemps(t *testing.T, dir string) {
	t.Helper()
	for _, n := range listDir(t, dir) {
		if strings.Contains(n, "hive-tmp") {
			t.Errorf("leftover temporary file %s", n)
		}
	}
}

func TestProjectWriteKeepsMode0600(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	path := filepath.Join(root, "AGENTS.md")
	if err := os.Chmod(path, 0o600); err != nil {
		t.Fatal(err)
	}
	e, err := prepare(t, root, fullSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeProjectFile(e.File, e.Before, e.Existed, e.After); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v, want 0600", info.Mode().Perm())
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(e.After) {
		t.Fatalf("content %q", got)
	}
	noTemps(t, root)
}

func TestProjectWriteCreatesWith0644UnderUmask077(t *testing.T) {
	root := newProjectRepo(t)
	os.MkdirAll(filepath.Join(root, "_support", "openspec"), 0o755)
	e, err := prepare(t, root, fullSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	if e.Existed {
		t.Fatal("the file should not exist yet")
	}
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)
	if err := writeProjectFile(e.File, e.Before, e.Existed, e.After); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(e.File)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("mode %v, want 0644", info.Mode().Perm())
	}
	noTemps(t, root)
}

func TestProjectWriteRejectsAFileChangedSinceThePreview(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	e, err := prepare(t, root, fullSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(e.File, []byte("# T\nsomeone edited\n"), 0o644)
	h := hashOf(t, e.File)
	err = writeProjectFile(e.File, e.Before, e.Existed, e.After)
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md changed since the preview") {
		t.Fatalf("error %v", err)
	}
	if hashOf(t, e.File) != h {
		t.Fatal("the file was overwritten")
	}
	noTemps(t, root)
}

func TestProjectWriteConcurrentCreateFailsTheLinkWithoutOverwriting(t *testing.T) {
	root := newProjectRepo(t)
	os.MkdirAll(filepath.Join(root, "_support", "openspec"), 0o755)
	e, err := prepare(t, root, fullSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	projectWriteHook = func() { os.WriteFile(e.File, []byte("theirs\n"), 0o644) }
	defer func() { projectWriteHook = nil }()
	err = writeProjectFile(e.File, e.Before, e.Existed, e.After)
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md changed since the preview") {
		t.Fatalf("error %v", err)
	}
	if got, _ := os.ReadFile(e.File); string(got) != "theirs\n" {
		t.Fatalf("the other writer's file was replaced: %q", got)
	}
	noTemps(t, root)
}

func TestProjectWriteCreatedByOthersBeforeTheSecondRead(t *testing.T) {
	root := newProjectRepo(t)
	path := filepath.Join(root, "AGENTS.md")
	os.WriteFile(path, []byte("theirs\n"), 0o644)
	err := writeProjectFile(path, nil, false, []byte("mine\n"))
	if err == nil || !strings.Contains(err.Error(), "AGENTS.md changed since the preview") {
		t.Fatalf("error %v", err)
	}
	noTemps(t, root)
}

func TestProjectWriteRemovesTheTemporaryOnError(t *testing.T) {
	root := newProjectRepo(t)
	path := filepath.Join(root, "AGENTS.md")
	os.WriteFile(path, []byte("a\n"), 0o644)
	projectWriteHook = func() { os.Remove(path); os.Mkdir(path, 0o755) } // makes the rename fail
	defer func() { projectWriteHook = nil }()
	if err := writeProjectFile(path, []byte("a\n"), true, []byte("b\n")); err == nil {
		t.Fatal("expected an error")
	}
	noTemps(t, root)
}

func TestProjectWriteLeavesClaudeMDAlone(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	cm := filepath.Join(root, "CLAUDE.md")
	os.WriteFile(cm, []byte("Own rules.\n"), 0o644)
	h := hashOf(t, cm)
	e, err := prepare(t, root, fullSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(e.Warnings, "\n"), "CLAUDE.md does not import @AGENTS.md") {
		t.Fatalf("warnings %q", e.Warnings)
	}
	if err := writeProjectFile(e.File, e.Before, e.Existed, e.After); err != nil {
		t.Fatal(err)
	}
	if hashOf(t, cm) != h {
		t.Fatal("CLAUDE.md changed")
	}
}

func TestClaudeMDWarning(t *testing.T) {
	root := t.TempDir()
	if w := claudeMDWarning(root); w != "" {
		t.Fatalf("no CLAUDE.md: %q", w)
	}
	cm := filepath.Join(root, "CLAUDE.md")
	os.WriteFile(cm, []byte("# x\n  @AGENTS.md  \n"), 0o644)
	if w := claudeMDWarning(root); w != "" {
		t.Fatalf("with import: %q", w)
	}
	os.WriteFile(cm, []byte("# x\n@AGENTS.md.bak\n"), 0o644)
	if w := claudeMDWarning(root); !strings.Contains(w, "does not import @AGENTS.md") {
		t.Fatalf("without import: %q", w)
	}
	// A CLAUDE.md that is a link to AGENTS.md is AGENTS.md itself.
	os.Remove(cm)
	os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# a\n"), 0o644)
	os.Symlink("AGENTS.md", cm)
	if w := claudeMDWarning(root); w != "" {
		t.Fatalf("symlink to AGENTS.md: %q", w)
	}
}

func TestSuggestHiveValues(t *testing.T) {
	git := newGitRunner()
	cases := []struct {
		name, remote string
		wantTracker  string // "" means no suggestion
		wantProject  string
		notContains  string
	}{
		{"scp form", "git@github.com:o/r.git", "GitHub Issues · o/r", "r", ""},
		{"https with credentials", "https://x-access-token:SECRET@github.com/o/r.git", "GitHub Issues · o/r", "r", "SECRET"},
		{"lookalike host", "https://github.com.evil.com/o/r", "", "", ""},
		{"lookalike userinfo host", "https://github.com@evil.com/o/r", "", "", ""},
		{"ssh scheme", "ssh://git@github.com/o/r.git", "GitHub Issues · o/r", "r", ""},
		{"extra path", "https://github.com/o/r/extra", "", "", ""},
		{"bad characters", "git@github.com:o/r$(id).git", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := newProjectRepo(t)
			gitIn(t, root, "remote", "add", "origin", c.remote)
			got := suggestHiveValues(root, git)
			if got["Tracker"] != c.wantTracker {
				t.Errorf("Tracker %q, want %q", got["Tracker"], c.wantTracker)
			}
			wantProject := c.wantProject
			if wantProject == "" {
				wantProject = filepath.Base(root)
			}
			if got["Project"] != wantProject {
				t.Errorf("Project %q, want %q", got["Project"], wantProject)
			}
			if c.notContains != "" {
				for k, v := range got {
					if strings.Contains(v, c.notContains) {
						t.Errorf("%s leaks the credential: %q", k, v)
					}
				}
			}
		})
	}
}

func TestSuggestHiveValuesBranchAndSpecs(t *testing.T) {
	git := newGitRunner()
	root := newProjectRepo(t)
	if got := suggestHiveValues(root, git); got["Base branch"] != "" || got["Specs"] != "" {
		t.Fatalf("bare repository: %v", got)
	}
	gitIn(t, root, "remote", "add", "origin", "git@github.com:o/r.git")
	gitIn(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	os.MkdirAll(filepath.Join(root, "_support", "openspec"), 0o755)
	got := suggestHiveValues(root, git)
	if got["Base branch"] != "main" || got["Specs"] != "_support/openspec" {
		t.Fatalf("suggestions %v", got)
	}
	if _, ok := got["Delivery"]; ok {
		t.Error("optional keys must not be suggested")
	}
	// An origin/HEAD that is not a usable branch name is not suggested.
	gitIn(t, root, "update-ref", "refs/remotes/origin/-x", "HEAD")
	gitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/-x")
	if got := suggestHiveValues(root, git); got["Base branch"] != "" {
		t.Fatalf("unusable branch suggested: %v", got)
	}
}

// F1: a Base branch value with any kind of space is refused, not only a space or a tab.
func TestHiveRequestRefusesEveryKindOfSpaceInABaseBranch(t *testing.T) {
	for name, v := range map[string]string{"no-break space": "main x", "em space": "a b", "ideographic space": "a　b", "line separator": "a b"} {
		err := validateHiveRequest([]hiveItem{{"Base branch", v}}, nil)
		if err == nil || !strings.Contains(err.Error(), "Base branch") {
			t.Errorf("%s: %q was accepted (%v)", name, v, err)
		}
	}
}

func TestClaudeMDWarningAcceptsAnImportAnywhereOutsideCode(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		warns      bool
	}{
		{"inline", "# x\nRead @AGENTS.md first.\n", false},
		{"dot slash", "@./AGENTS.md\n", false},
		{"in parentheses", "See (@AGENTS.md) for rules\n", true},
		{"double quoted", "\"@AGENTS.md\"\n", true},
		{"single quoted", "'@AGENTS.md'\n", true},
		{"trailing period", "Read @AGENTS.md.\n", true},
		{"tilde block holding a backtick fence", "~~~md\n```\n@AGENTS.md\n```\n~~~\n", true},
		{"longer fence is not closed by a shorter one", "````\n```\n@AGENTS.md\n````\n", true},
		{"import after a nested block", "~~~\n```\n~~~\n@AGENTS.md\n", false},
		{"fenced example", "```\n@AGENTS.md\n```\n", true},
		{"tilde fence", "~~~md\nRead @AGENTS.md\n~~~\n", true},
		{"inline code span", "Write `@AGENTS.md` to import it.\n", true},
		{"other file", "Read @AGENTS.md.bak and @docs/AGENTS.md\n", true},
		{"after a fence", "```\nx\n```\n@AGENTS.md\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte(tc.text), 0o644)
			if got := claudeMDWarning(root) != ""; got != tc.warns {
				t.Fatalf("warns = %v, want %v for %q", got, tc.warns, tc.text)
			}
		})
	}
}
