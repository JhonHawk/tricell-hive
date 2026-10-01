package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runProjectSet(t *testing.T, input string, interactive bool, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	err := projectSet(args, strings.NewReader(input), &out, interactive)
	return out.String(), err
}

func setArgs(root string, extra ...string) []string {
	return append([]string{"--project", root,
		"--set", "Project: demo", "--set", "Base branch: main",
		"--set", "Tracker: GitHub Issues · o/demo", "--set", "Specs: _support/openspec"}, extra...)
}

func TestProjectCommandWithoutTerminalOnlyAcceptsDryRun(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	h := hashOf(t, filepath.Join(root, "AGENTS.md"))
	_, err := runProjectSet(t, "y\n", false, setArgs(root)...)
	if err == nil || !strings.Contains(err.Error(), "an interactive terminal is required to confirm; use --dry-run to preview") {
		t.Fatalf("error %v", err)
	}
	if hashOf(t, filepath.Join(root, "AGENTS.md")) != h {
		t.Fatal("the file changed")
	}
	// Without a file either: nothing is created.
	fresh := newProjectRepo(t)
	os.MkdirAll(filepath.Join(fresh, "_support", "openspec"), 0o755)
	if _, err := runProjectSet(t, "", false, setArgs(fresh)...); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(filepath.Join(fresh, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md was created: %v", err)
	}
}

func TestProjectCommandDryRunShowsTheSummaryAndWritesNothing(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n\n## Hive\n\n- Project: old\n- Base branch: ghost\n- Tracker: t\n- Specs: nowhere\n")
	gitIn(t, root, "add", "AGENTS.md")
	gitIn(t, root, "commit", "-q", "-m", "agents")
	os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# T\n\n## Hive\n\n- Project: old\n- Base branch: ghost\n- Tracker: t\n- Specs: nowhere\n\nedited\n"), 0o644)
	os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("own\n"), 0o644)
	h := hashOf(t, filepath.Join(root, "AGENTS.md"))
	out, err := runProjectSet(t, "", false, "--project", root, "--set", "Project: new", "--dry-run")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{
		filepath.Join(root, "AGENTS.md"),
		"Before:", "- Project: old", "After:", "- Project: new",
		"Base branch: ghost is not a local branch",
		"Specs: nowhere is not an existing directory",
		"CLAUDE.md does not import @AGENTS.md",
		"AGENTS.md has uncommitted changes",
		"Preview: nothing was changed.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if hashOf(t, filepath.Join(root, "AGENTS.md")) != h {
		t.Fatal("dry run wrote")
	}
}

func TestProjectCommandConfirmationApplyAndCancel(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	path := filepath.Join(root, "AGENTS.md")
	h := hashOf(t, path)
	out, err := runProjectSet(t, "n\n", true, setArgs(root)...)
	if err != nil || !strings.Contains(out, "Cancelled. No changes applied.") {
		t.Fatalf("cancel: %v\n%s", err, out)
	}
	if hashOf(t, path) != h {
		t.Fatal("cancel wrote")
	}
	out, err = runProjectSet(t, "y\n", true, setArgs(root)...)
	if err != nil {
		t.Fatalf("apply: %v\n%s", err, out)
	}
	got, _ := os.ReadFile(path)
	want := "# T\n\n## Hive\n\n- Project: demo\n- Base branch: main\n- Tracker: GitHub Issues · o/demo\n- Specs: _support/openspec\n"
	if string(got) != want {
		t.Fatalf("file:\n%q\nwant:\n%q", got, want)
	}
	if sec := projectText(t, root); !strings.Contains(sec, "Valid") {
		t.Fatalf("doctor:\n%s", sec)
	}
	// Running it again changes nothing and says so without asking.
	out, err = runProjectSet(t, "", false, setArgs(root)...)
	if err != nil || !strings.Contains(out, "Nothing to change") {
		t.Fatalf("again: %v\n%s", err, out)
	}
}

func TestProjectCommandUsageErrors(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"nothing to do", []string{"--project", root}, "at least one --set or --unset"},
		{"set without a colon", []string{"--project", root, "--set", "Project demo"}, "--set expects 'Key: value'"},
		{"set twice", []string{"--project", root, "--set", "Project: a", "--set", "Project: b"}, "set more than once"},
		{"set and unset", []string{"--project", root, "--set", "Review: a", "--unset", "Review"}, "both set and unset"},
		{"positional", []string{"--project", root, "--set", "Project: a", "extra"}, "unexpected positional"},
		{"unknown key", []string{"--project", root, "--set", "Mystery: a"}, "Mystery: unknown key"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := runProjectSet(t, "", false, append(c.args, "--dry-run")...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %v, want %q", err, c.want)
			}
		})
	}
}

func TestProjectCommandUnsetOptionalKey(t *testing.T) {
	root := projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Specs: _support/openspec\n", "- Specs: _support/openspec\n- Review: /code-review\n", 1))
	out, err := runProjectSet(t, "y\n", true, "--project", root, "--unset", "Review")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	got, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if strings.Contains(string(got), "Review") {
		t.Fatalf("Review still there:\n%s", got)
	}
}

// D4: the writer refuses a Base branch value Git rejects, judged on the whole
// value (the doctor check reads only its first word and is left as it is).
func TestProjectCommandRefusesABaseBranchGitRejects(t *testing.T) {
	root := projectRepoWithSpecs(t, "# T\n")
	path := filepath.Join(root, "AGENTS.md")
	h := hashOf(t, path)
	for _, bad := range []string{"bad branch", "main extra", "a..b", "-x", "a~b"} {
		args := []string{"--project", root, "--set", "Project: demo", "--set", "Base branch: " + bad,
			"--set", "Tracker: GitHub Issues · o/demo", "--set", "Specs: _support/openspec"}
		out, err := runProjectSet(t, "y\n", true, args...)
		if err == nil || !strings.Contains(err.Error(), "Base branch") || !strings.Contains(err.Error(), "not a valid branch name") {
			t.Errorf("%q: error %v\n%s", bad, err, out)
		}
		if hashOf(t, path) != h {
			t.Fatalf("%q: the file changed", bad)
		}
	}
	for _, good := range []string{"main", "feature/x", "release-1.2"} {
		if err := validateHiveRequest([]hiveItem{{"Base branch", good}}, nil); err != nil {
			t.Errorf("%q refused: %v", good, err)
		}
	}
}

// D5: creating the file says "Created", changing it says "Updated".
func TestProjectCommandReportsCreatedForANewFile(t *testing.T) {
	root := formRepo(t) // no AGENTS.md yet
	out, err := runProjectSet(t, "y\n", true, setArgs(root)...)
	if err != nil || !strings.Contains(out, "Created ") || strings.Contains(out, "Updated ") {
		t.Fatalf("new file: %v\n%s", err, out)
	}
	out, err = runProjectSet(t, "y\n", true, setArgs(root, "--set", "Delivery: direct-base")...)
	if err != nil || !strings.Contains(out, "Updated ") || strings.Contains(out, "Created ") {
		t.Fatalf("existing file: %v\n%s", err, out)
	}
}
