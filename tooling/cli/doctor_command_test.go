package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the `hive doctor` and `hive models` commands (#46, T5).

// installCodexFixture installs the codex host over a characterization fixture
// and returns it.
func installCodexFixture(t *testing.T) characterizationFixture {
	t.Helper()
	f := newCharacterizationFixture(t)
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir, Source: f.source, Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestDoctorCommandPrintsFindingsAndExitsZero(t *testing.T) {
	f := installCodexFixture(t)
	block := filepath.Join(f.home, ".codex", "AGENTS.md")
	data, err := os.ReadFile(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(block, []byte(strings.Replace(string(data), "Keep user content.", "Edited by hand.", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	var runErr error
	t.Chdir(t.TempDir())
	out := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir})
	})
	if runErr != nil {
		t.Fatalf("doctor failed on a finding: %v\n%s", runErr, out)
	}
	for _, want := range []string{"CLIs\n", "Installation\n", "Sessions\n", "drift  codex", block, "If the file was edited or its permissions changed"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// The command prints exactly what the shared renderer produces.
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir}
	var want strings.Builder
	project, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	renderDoctorText(collectDoctor(o, project, realDoctorDeps()), &want)
	got := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir})
	})
	if runErr != nil || got != want.String() {
		t.Errorf("doctor output differs from renderDoctorText (err %v):\n%s\n---\n%s", runErr, got, want.String())
	}
}

func TestDoctorCommandProjectFlagChecksThatRepository(t *testing.T) {
	f := newCharacterizationFixture(t)
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	var runErr error
	out := captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir, "--project", repo})
	})
	canonical, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatal(err)
	}
	// An explicit --project that names a repository without AGENTS.md fails
	// after printing the same report (D10-A).
	if runErr == nil || runErr.Error() != projectBlockingMessage || !strings.Contains(out, "Project\n  "+filepath.Join(canonical, "AGENTS.md")+"\n  AGENTS.md: not found at the repository root\n") {
		t.Errorf("doctor --project did not check %s (err %v):\n%s", repo, runErr, out)
	}
	outside := t.TempDir()
	out = captureStdout(t, func() {
		runErr = run([]string{"doctor", "--home", f.home, "--state-dir", f.stateDir, "--project", outside})
	})
	if runErr == nil || runErr.Error() != projectBlockingMessage || !strings.Contains(out, "Project\n  Directory: ") || !strings.Contains(out, outsideGitLine+"\n  "+outsideGitLineTwo+"\n") {
		t.Errorf("doctor outside a repository (err %v):\n%s", runErr, out)
	}
}

// Exit status of `hive doctor` for the Project section (AC5): blocking findings
// fail, warnings do not, and a missing AGENTS.md or section fails only for an
// explicit --project.
func TestDoctorCommandExitStatusFollowsTheProjectSection(t *testing.T) {
	f := newCharacterizationFixture(t)
	big := validHiveSection + strings.Repeat("x", hiveFileLimit)
	cases := []struct {
		name     string
		repo     func(t *testing.T) string
		explicit bool
		fails    bool
		report   string
	}{
		{"valid section", func(t *testing.T) string { return projectRepoWithSpecs(t, validHiveSection) }, true, false, "Valid"},
		{"missing required value", func(t *testing.T) string {
			return projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Tracker: GitHub Issues · org/demo\n", "", 1))
		}, false, true, "Tracker: required value is missing"},
		{"empty required value", func(t *testing.T) string {
			return projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Project: demo", "- Project:", 1))
		}, false, true, "Project: value is empty"},
		{"invalid Delivery", func(t *testing.T) string {
			return projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Specs:", "- Delivery: pr\n- Specs:", 1))
		}, false, true, "Delivery: expected"},
		{"invalid Hive guidance", func(t *testing.T) string {
			return projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Specs:", "- Hive guidance: optional\n- Specs:", 1))
		}, false, true, "Hive guidance: expected"},
		{"duplicated section", func(t *testing.T) string {
			return projectRepoWithSpecs(t, validHiveSection+"\n## Hive\n\n- Project: again\n")
		}, false, true, "section appears 2 times"},
		{"AGENTS.md over 1 MiB", func(t *testing.T) string { return projectRepoWithSpecs(t, big) }, false, true, "larger than 1 MiB"},
		{"AGENTS.md is not a regular file", func(t *testing.T) string {
			root := projectRepoWithSpecs(t, "")
			if err := os.Remove(filepath.Join(root, "AGENTS.md")); err != nil {
				t.Fatal(err)
			}
			if err := os.Mkdir(filepath.Join(root, "AGENTS.md"), 0o755); err != nil {
				t.Fatal(err)
			}
			return root
		}, false, true, "not a regular file"},
		{"only warnings", func(t *testing.T) string {
			return projectRepoWithSpecs(t, strings.Replace(validHiveSection, "- Specs:", "- Review: x\n- Delivery: automatic\n- Colour: blue\n- Completion: Done\n- Specs:", 1))
		}, true, false, "Review: no longer read"},
		{"no AGENTS.md without the flag", func(t *testing.T) string { return newProjectRepo(t) }, false, false, "not found at the repository root"},
		{"no AGENTS.md with the flag", func(t *testing.T) string { return newProjectRepo(t) }, true, true, "not found at the repository root"},
		{"no section without the flag", func(t *testing.T) string { return projectRepoWithSpecs(t, "# T\n") }, false, false, "section not found"},
		{"no section with the flag", func(t *testing.T) string { return projectRepoWithSpecs(t, "# T\n") }, true, true, "section not found"},
		{"outside a repository without the flag", func(t *testing.T) string { return t.TempDir() }, false, false, outsideGitLine},
		{"outside a repository with the flag", func(t *testing.T) string { return t.TempDir() }, true, true, outsideGitLine},
		{"missing directory with the flag", func(t *testing.T) string { return filepath.Join(t.TempDir(), "nope") }, true, true, "Directory not found"},
		{"file instead of a directory with the flag", func(t *testing.T) string {
			p := filepath.Join(t.TempDir(), "f")
			if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
				t.Fatal(err)
			}
			return p
		}, true, true, "Not a directory"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := c.repo(t)
			args := []string{"doctor", "--home", f.home, "--state-dir", f.stateDir}
			if c.explicit {
				args = append(args, "--project", repo)
			} else {
				t.Chdir(repo)
			}
			var runErr error
			out := captureStdout(t, func() { runErr = run(args) })
			if !strings.Contains(out, c.report) {
				t.Errorf("report lacks %q:\n%s", c.report, out)
			}
			switch {
			case c.fails && (runErr == nil || runErr.Error() != projectBlockingMessage):
				t.Errorf("want the error %q, got %v", projectBlockingMessage, runErr)
			case !c.fails && runErr != nil:
				t.Errorf("want exit 0, got %v", runErr)
			}
		})
	}
}

func TestDoctorCommandFailsOnAStateDirThatIsAFile(t *testing.T) {
	f := newCharacterizationFixture(t)
	file := filepath.Join(t.TempDir(), "state")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	var runErr error
	captureStdout(t, func() { runErr = run([]string{"doctor", "--home", f.home, "--state-dir", file}) })
	if runErr == nil {
		t.Fatal("doctor accepted a state directory that is a file")
	}
	captureStdout(t, func() { runErr = run([]string{"models", "--home", f.home, "--state-dir", file}) })
	if runErr == nil {
		t.Fatal("models accepted a state directory that is a file")
	}
}

func TestModelsCommandPrintsTheRowsAndExitsZero(t *testing.T) {
	f := installCodexFixture(t)
	var runErr error
	out := captureStdout(t, func() { runErr = run([]string{"models", "--home", f.home, "--state-dir", f.stateDir}) })
	if runErr != nil {
		t.Fatalf("models failed: %v", runErr)
	}
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir}
	hosts, rows, err := collectModels(o)
	if err != nil {
		t.Fatal(err)
	}
	var want strings.Builder
	renderModelsText(hosts, rows, &want)
	if out != want.String() || !strings.HasPrefix(out, "codex\n") {
		t.Errorf("models output:\n%s\nwant:\n%s", out, want.String())
	}

	empty := newCharacterizationFixture(t)
	out = captureStdout(t, func() { runErr = run([]string{"models", "--home", empty.home, "--state-dir", empty.stateDir}) })
	if runErr != nil || out != "No CLI hosts are registered. Open CLIs from the menu, or run hive install, to install Hive.\n" {
		t.Errorf("models without hosts = %q, %v", out, runErr)
	}
}

func TestDoctorAndModelsRejectPositionalArguments(t *testing.T) {
	for _, cmd := range []string{"doctor", "models"} {
		var runErr error
		captureStdout(t, func() { runErr = run([]string{cmd, "extra"}) })
		if runErr == nil {
			t.Errorf("%s accepted a positional argument", cmd)
		}
	}
}
