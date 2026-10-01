package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// Tests of the Project view.

// openProjectView opens the app and pushes the Project view over dir, the way
// the menu will. git replaces the runner when it is not nil.
func openProjectView(t *testing.T, dir string, git gitRunner, width, height int) (*appModel, *appDriver, *projectView) {
	t.Helper()
	cfg := testAppConfig(t)
	m, d := newTestApp(t, cfg, width, height)
	v := newProjectViewWith(cfg, dir, doctorDeps{})
	if git != nil {
		v.git = git
	}
	d.send(pushViewMsg{v: v})
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	return m, d, v
}

func TestProjectViewShowsVerdictPathAndValues(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := projectRepoWithSpecs(t, validHiveSection)
			_, d, _ := openProjectView(t, root, nil, size[0], size[1])
			d.mustShow("Project", "Valid", "Values read:", "Base branch: main", "Specs: _support/openspec")
			mustShowWhole(t, d, realPath(t, filepath.Join(root, "AGENTS.md")))
			d.mustNotShow("Checking", "Loading")
			assertFits(t, d, size[0], size[1])
		})
	}
}

func TestProjectViewShowsFindings(t *testing.T) {
	root := projectRepoWithSpecs(t, strings.Replace(validHiveSection, "Base branch: main", "Base branch: nope", 1))
	_, d, v := openProjectView(t, root, nil, 80, 24)
	d.mustShow("Base branch: nope is not a local branch or on origin")
	d.mustNotShow("Valid", "r to retry")
	if v.failed {
		t.Fatal("a finding is not a failed check")
	}
}

func TestProjectViewWithoutHiveSectionListsTheRequiredKeys(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		for name, agents := range map[string]string{"no file": "", "no section": "# Guidance\n"} {
			t.Run(fmt.Sprintf("%s %dx%d", name, size[0], size[1]), func(t *testing.T) {
				root := newProjectRepo(t)
				if agents != "" {
					writeAgents(t, root, agents)
				}
				_, d, _ := openProjectView(t, root, nil, size[0], size[1])
				d.mustShow("Add a ## Hive section with: Project, Base branch, Tracker, Specs")
				assertFits(t, d, size[0], size[1])
			})
		}
	}
}

func TestProjectViewOutsideGitShowsTwoLinesAndTheDirectory(t *testing.T) {
	dir := t.TempDir()
	_, d, _ := openProjectView(t, dir, nil, 80, 24)
	d.mustShow(outsideGitLine, outsideGitLineTwo, "Directory:")
	assertFits(t, d, 80, 24)
}

func TestProjectViewShowsSpinnerWhileLoading(t *testing.T) {
	v := newProjectViewWith(testAppConfig(t), t.TempDir(), doctorDeps{})
	if !v.NeedsSpinner() {
		t.Fatal("a loading view must ask for the spinner")
	}
	th := newAppTheme(true, false)
	out := stripANSI(v.View(viewCtx{Width: 80, Height: 21, Theme: &th, Spinner: "*"}))
	mustContain(t, out, "Project", "* Checking")
	if lines := strings.Split(out, "\n"); len(lines) > 21 {
		t.Fatalf("%d lines", len(lines))
	}
}

func TestProjectViewReloadsWithRAndIgnoresStaleResults(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	_, d, v := openProjectView(t, root, nil, 80, 24)
	d.mustShow("Valid")

	writeAgents(t, root, strings.Replace(validHiveSection, "- Tracker: GitHub Issues · org/demo\n", "", 1))
	d.key("r")
	d.mustShow("Tracker: required value is missing")
	d.mustNotShow("Valid")

	stale := projectLoadedMsg{owned: owned{v}, seq: v.seq - 1}
	stale.check.Section = doctorSection{Title: "Project", Lines: []string{"STALE"}}
	d.send(stale)
	d.mustNotShow("STALE")
	d.mustShow("Tracker: required value is missing")

	// A double r while loading starts one load only.
	cmd, _ := v.Update(keyMsg(t, "r"))
	if cmd == nil {
		t.Fatal("r did not reload")
	}
	seq := v.seq
	if cmd, _ := v.Update(keyMsg(t, "r")); cmd != nil || v.seq != seq {
		t.Fatal("r while loading started another load")
	}
}

func TestProjectViewLateResultAfterLeavingIsDropped(t *testing.T) {
	m, d, v := openProjectView(t, t.TempDir(), nil, 80, 24)
	d.key("esc")
	if _, ok := m.top().(*menuView); !ok {
		t.Fatalf("top view is %T", m.top())
	}
	late := projectLoadedMsg{owned: owned{v}, seq: v.seq}
	late.check.Section = doctorSection{Title: "Project", Lines: []string{"LATE"}}
	d.send(late)
	d.mustNotShow("LATE")
}

func TestProjectViewEscAndBackspaceReturnToMenu(t *testing.T) {
	for _, k := range []string{"esc", "backspace"} {
		t.Run(k, func(t *testing.T) {
			m, d, _ := openProjectView(t, t.TempDir(), nil, 80, 24)
			d.key(k)
			if _, ok := m.top().(*menuView); !ok {
				t.Fatalf("top view is %T after %s", m.top(), k)
			}
		})
	}
}

func TestProjectViewLoadErrorIsShownInsideWithRetry(t *testing.T) {
	broken := func(string, ...string) (string, error) { return "", errors.New("git is broken") }
	_, d, v := openProjectView(t, t.TempDir(), broken, 80, 24)
	d.mustShow("Could not check everything", "git is broken", "r to retry")
	assertFits(t, d, 80, 24)
	if !v.failed {
		t.Fatal("the view must be marked failed")
	}
}

func TestProjectViewScrollsToTheLastFindingAndBack(t *testing.T) {
	var b strings.Builder
	b.WriteString("## Hive\n")
	for i := 0; i < 60; i++ {
		fmt.Fprintf(&b, "- Extra%02d: %s\n", i, strings.Repeat("value ", 20))
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := projectRepoWithSpecs(t, b.String())
			_, d, v := openProjectView(t, root, nil, size[0], size[1])
			assertFits(t, d, size[0], size[1])
			d.mustShow("Project: required value is missing")
			if !v.box.scrollable() {
				t.Fatal("the fixture must not fit the screen")
			}
			for range 400 {
				d.key("down")
			}
			d.mustShow("Extra59: value")
			d.mustNotShow("Project: required value is missing")
			assertFits(t, d, size[0], size[1])
			for range 400 {
				d.key("up")
			}
			d.mustShow("Project: required value is missing", "lines 1-")
			for range 60 {
				d.key("pgdown")
			}
			d.mustShow("Extra59: value")
			for range 60 {
				d.key("pgup")
			}
			d.mustShow("Project: required value is missing")
		})
	}
}

// realPath resolves symbolic links, as the view shows the resolved path.
func realPath(t *testing.T, p string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// mustShowWhole checks that text is on the screen in full, even when the view
// wrapped it over several lines: it compares with all whitespace removed.
func mustShowWhole(t *testing.T, d *appDriver, text string) {
	t.Helper()
	squash := func(s string) string { return strings.Join(strings.Fields(s), "") }
	if !strings.Contains(squash(d.screen()), squash(text)) {
		t.Fatalf("%q is not shown in full:\n%s", text, d.screen())
	}
}

// TestProjectViewWrapsALongPathInsteadOfCuttingIt covers N3: the location is the
// first line of the scrolling text, wrapped with a hanging indent, so a path
// longer than the screen is fully readable.
func TestProjectViewWrapsALongPathInsteadOfCuttingIt(t *testing.T) {
	// A repository whose root path is longer than the screen.
	longRoot := filepath.Join(t.TempDir(), strings.Repeat("a-long-directory-name", 6))
	if err := os.MkdirAll(longRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, longRoot, "init", "-q", "-b", "main")
	writeAgents(t, longRoot, validHiveSection)
	file := realPath(t, filepath.Join(longRoot, "AGENTS.md"))
	if len(file) <= 80 {
		t.Fatalf("the fixture path is only %d columns", len(file))
	}
	_, d, v := openProjectView(t, longRoot, nil, 80, 24)
	mustShowWhole(t, d, file)
	d.mustNotShow("…")
	d.mustShow("Values read:")
	assertFits(t, d, 80, 24)
	if got := v.box.vp.Height(); got != 21-2 {
		t.Fatalf("scrolling text has %d rows, want 19 (only the heading and the position are fixed)", got)
	}
	// The first line of the scrolling text starts the path, and the rest hangs.
	lines := d.lines()
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "Project") {
			start = i
			break
		}
	}
	if start < 0 || !strings.HasPrefix(lines[start+1], "/") || !strings.HasPrefix(lines[start+2], "  ") {
		t.Fatalf("the path must open the scrolling text and continue with an indent:\n%s", d.screen())
	}
}

// TestProjectViewShowsTheSameLocationLineAsTheCommand covers N3's second half:
// the view and hive doctor share the location line, for a repository and for a
// directory outside one.
func TestProjectViewShowsTheSameLocationLineAsTheCommand(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	for _, dir := range []string{root, t.TempDir()} {
		sec := collectProject(dir, doctorDeps{})
		_, d, _ := openProjectView(t, dir, nil, 120, 40)
		mustShowWhole(t, d, sec.Lines[0])
		mustContain(t, projectBoxText(checkProject(dir, newGitRunner())), sec.Lines[0]+"\n")
	}
}

func TestProjectViewLeavesTheRepositoryUnchanged(t *testing.T) {
	root := projectRepoWithSpecs(t, validHiveSection)
	before, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	filesBefore := fmt.Sprint(collectFiles(t, root))
	_, d, _ := openProjectView(t, root, nil, 80, 24)
	d.key("r", "down", "up", "pgdown", "pgup")
	after, _ := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if !bytes.Equal(before, after) || filesBefore != fmt.Sprint(collectFiles(t, root)) {
		t.Fatal("the repository changed")
	}
}

func TestProjectViewIgnoresOtherMessages(t *testing.T) {
	v := newProjectViewWith(testAppConfig(t), t.TempDir(), doctorDeps{})
	if cmd, act := v.Update(tea.WindowSizeMsg{}); cmd != nil || act.nav != navNone {
		t.Fatalf("cmd=%v act=%+v", cmd, act)
	}
}

func TestProjectViewCurrentDirectoryConstructor(t *testing.T) {
	v := newProjectView(testAppConfig(t))
	if v.dir != "" || !v.loading {
		t.Fatalf("dir %q loading %v", v.dir, v.loading)
	}
}

// ---------------------------------------------------------------------------
// The form (T5).
// ---------------------------------------------------------------------------

// formRepo is the repository of the user's walkthrough: a GitHub remote,
// origin/HEAD, a specs directory and no AGENTS.md.
func formRepo(t *testing.T) string {
	t.Helper()
	root := newProjectRepo(t)
	gitIn(t, root, "remote", "add", "origin", "git@github.com:o/r.git")
	gitIn(t, root, "update-ref", "refs/remotes/origin/main", "HEAD")
	gitIn(t, root, "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	if err := os.MkdirAll(filepath.Join(root, "_support", "openspec"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func openForm(t *testing.T, root string, w, h int) (*appModel, *appDriver, *projectView) {
	t.Helper()
	m, d, v := openProjectView(t, root, nil, w, h)
	d.mustNotShow("Checking")
	d.key("e")
	// Render once before any typing: until the first View the text input still
	// blinks and each typed key would run a ~530 ms blink command (#59).
	d.screen()
	return m, d, v
}

// clearField empties the focused text field.
func clearField(d *appDriver) {
	d.t.Helper()
	d.send(tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
}

// formRows are the field labels in the order the form shows them.
var formRows = []string{"Project", "Base branch", "Tracker", "Specs", "Environments", "Review", "Delivery", "Hive guidance"}

func TestProjectViewFormOpensWithEightFieldsAndSuggestions(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			root := formRepo(t)
			_, d, _ := openForm(t, root, size[0], size[1])
			screen := d.screen()
			last := -1
			for _, label := range formRows {
				at := strings.Index(screen, label)
				if at < 0 || at < last {
					t.Fatalf("row %q missing or out of order:\n%s", label, screen)
				}
				last = at
			}
			for _, want := range []string{"> Project", "r", "main", "GitHub Issues · o/r", "_support/openspec", "‹ none ›"} {
				d.mustShow(want)
			}
			if n := strings.Count(screen, "suggested"); n != 4 {
				t.Fatalf("%d suggested labels, want 4:\n%s", n, screen)
			}
			assertFits(t, d, size[0], size[1])
		})
	}
}

func TestProjectViewFormChoiceFieldsCycleAndRequiredFieldsAreChecked(t *testing.T) {
	_, d, _ := openForm(t, formRepo(t), 80, 24)
	d.key("down", "down", "down", "down", "down", "down") // Delivery
	d.mustShow("> Delivery      ‹ none ›")
	d.key("right")
	d.mustShow("‹ direct-base ›")
	d.key("right")
	d.mustShow("Delivery      ‹ none ›")
	d.key("left")
	d.mustShow("‹ direct-base ›")

	d.key("up", "up", "up", "up", "up", "up") // Project
	clearField(d)
	d.key("enter")
	d.mustShow("Project is required")
	d.mustNotShow("Write AGENTS.md", "Apply")
	d.mustShow("> Project") // the focus went to the empty field
	if n := len(d.lines()); n > 24 {
		t.Fatalf("%d lines", n)
	}
}

func TestProjectViewFormLoadsTheExistingSectionWithoutSuggestions(t *testing.T) {
	root := formRepo(t)
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs: _support/openspec\n", "- Specs: _support/openspec\n- Delivery: direct-base\n", 1))
	_, d, _ := openForm(t, root, 80, 24)
	d.mustShow("> Project", "demo", "main", "GitHub Issues · org/demo", "_support/openspec", "‹ direct-base ›")
	d.mustNotShow("suggested", "o/r")
}

func TestProjectViewFormConfirmationShowsTheChangeAndTheWarnings(t *testing.T) {
	root := formRepo(t)
	writeAgents(t, root, "# T\n\n## Hive\n\n- Project: old\n- Base branch: ghost\n- Tracker: t\n- Specs: nowhere\n")
	gitIn(t, root, "add", "AGENTS.md")
	gitIn(t, root, "commit", "-q", "-m", "agents")
	writeAgents(t, root, "# T\n\n## Hive\n\n- Project: old\n- Base branch: ghost\n- Tracker: t\n- Specs: nowhere\n\nedited\n")
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("own\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, d, _ := openForm(t, root, 120, 40)
	clearField(d)
	typeText(d, "new")
	d.key("enter")
	d.mustShow("Write AGENTS.md", "Before:", "- Project: old", "After:", "- Project: new",
		"Base branch: ghost is not a local branch", "Specs: nowhere is not an existing directory",
		"CLAUDE.md does not import @AGENTS.md", "AGENTS.md has uncommitted changes")
	assertFits(t, d, 120, 40)
}

func TestProjectViewFormApplyWritesWhatTheCommandWrites(t *testing.T) {
	root := formRepo(t)
	_, d, v := openForm(t, root, 80, 24)
	d.key("down", "down", "down", "down", "down", "down", "right") // Delivery: direct-base
	d.key("enter")
	d.mustShow("Write AGENTS.md", "- Delivery: direct-base")
	d.key("y")
	d.mustShow("Valid")
	d.mustNotShow("> Project", "Write AGENTS.md")
	if m := d.model.(*appModel); m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	got, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The same values through the command, in another repository.
	other := formRepo(t)
	var out bytes.Buffer
	args := []string{"--project", other, "--set", "Project: r", "--set", "Base branch: main",
		"--set", "Tracker: GitHub Issues · o/r", "--set", "Specs: _support/openspec", "--set", "Delivery: direct-base"}
	if err := projectSet(args, strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	want, _ := os.ReadFile(filepath.Join(other, "AGENTS.md"))
	if !bytes.Equal(got, want) {
		t.Fatalf("view wrote:\n%q\ncommand wrote:\n%q", got, want)
	}
}

func TestProjectViewFormCancelKeepsWhatWasTyped(t *testing.T) {
	root := formRepo(t)
	_, d, _ := openForm(t, root, 80, 24)
	typeText(d, "x y")
	d.key("enter")
	d.mustShow("Write AGENTS.md")
	d.key("n")
	d.mustShow("> Project", "rx y", "Cancelled. No changes applied.")
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md exists after cancel: %v", err)
	}
	// An existing file keeps its hash.
	writeAgents(t, root, "# T\n")
	h := hashOf(t, filepath.Join(root, "AGENTS.md"))
	d.key("enter")
	d.key("esc") // Esc in the confirmation cancels it
	d.mustShow("Cancelled. No changes applied.", "rx y")
	if hashOf(t, filepath.Join(root, "AGENTS.md")) != h {
		t.Fatal("cancel wrote")
	}
}

func TestProjectViewFormSaysNothingToChange(t *testing.T) {
	root := formRepo(t)
	writeAgents(t, root, strings.Replace(validHiveSection, "- Specs: _support/openspec\n", "- Specs: _support/openspec\n", 1))
	_, d, _ := openForm(t, root, 80, 24)
	d.key("enter")
	d.mustShow("Nothing to change")
	d.mustNotShow("Write AGENTS.md")
}

func TestProjectViewFormShowsAWriteErrorAndKeepsTheValues(t *testing.T) {
	root := formRepo(t)
	writeAgents(t, root, "# T\n")
	_, d, _ := openForm(t, root, 80, 24)
	typeText(d, "zz")
	d.key("enter")
	d.mustShow("Write AGENTS.md")
	writeAgents(t, root, "# T\nsomeone edited\n")
	h := hashOf(t, filepath.Join(root, "AGENTS.md"))
	d.key("y")
	d.mustShow("AGENTS.md changed since the preview", "> Project", "rzz")
	if hashOf(t, filepath.Join(root, "AGENTS.md")) != h {
		t.Fatal("the other edit was overwritten")
	}
}

func TestProjectViewFormKeys(t *testing.T) {
	root := formRepo(t)
	m, d, v := openForm(t, root, 80, 24)
	// Esc closes only the form.
	d.key("esc")
	d.mustNotShow("> Project")
	d.mustShow("Add a ## Hive section")
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	// Backspace edits the text; r and e are typed.
	d.key("e")
	d.mustShow("> Project")
	d.key("backspace") // deletes the "r" of the suggestion
	d.mustNotShow("> Project       r")
	typeText(d, "re")
	d.mustShow("> Project       re")
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
	// Backspace on a choice field leaves neither the form nor the view.
	d.key("down", "down", "down", "down", "down", "down")
	d.key("backspace")
	d.mustShow("> Delivery")
	if m.top() != view(v) {
		t.Fatalf("top view is %T", m.top())
	}
}

func TestProjectViewEOutsideARepositoryDoesNothing(t *testing.T) {
	m, d, v := openProjectView(t, t.TempDir(), nil, 80, 24)
	before := d.screen()
	d.key("e")
	if d.screen() != before || m.top() != view(v) {
		t.Fatalf("e changed the screen:\n%s", d.screen())
	}
	d.mustNotShow("e edit")
}

func TestProjectViewFormFitsWithALongValueAndHelpBarShowsTheKeys(t *testing.T) {
	long := strings.Repeat("0123456789", 20)
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			_, d, v := openForm(t, formRepo(t), size[0], size[1])
			d.key("down", "down") // Tracker
			// Typing 200 keys one by one would wait for each cursor blink.
			v.form.inputs[2].SetValue(long)
			v.form.inputs[2].CursorEnd()
			d.key("end")
			assertFits(t, d, size[0], size[1])
			d.mustShow("> Tracker")
			d.key("enter") // a 200-column value in the confirmation
			assertFits(t, d, size[0], size[1])
			d.key("n")
			help := d.lines()[len(d.lines())-1]
			for _, want := range []string{"esc close", "ctrl+c quit", "enter review"} {
				if !strings.Contains(help, want) {
					t.Fatalf("help bar %q lacks %q", help, want)
				}
			}
		})
	}
}

func TestProjectViewHelpBarOfTheValidationListsEdit(t *testing.T) {
	_, d, _ := openProjectView(t, formRepo(t), nil, 80, 24)
	help := d.lines()[len(d.lines())-1]
	for _, want := range []string{"e edit", "esc back", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help bar %q lacks %q", help, want)
		}
	}
}

func TestProjectViewTextFocusedFollowsTheFocusedField(t *testing.T) {
	_, d, v := openProjectView(t, formRepo(t), nil, 80, 24)
	if v.TextFocused() {
		t.Fatal("TextFocused is true with the form closed")
	}
	d.key("e")
	d.screen()
	if !v.TextFocused() {
		t.Fatal("TextFocused is false on a text field")
	}
	d.key("down", "down", "down", "down", "down", "down") // Delivery
	if v.TextFocused() {
		t.Fatal("TextFocused is true on Delivery")
	}
	d.key("down") // Hive guidance
	if v.TextFocused() {
		t.Fatal("TextFocused is true on Hive guidance")
	}
	d.key("up", "up", "up") // Specs
	if !v.TextFocused() {
		t.Fatal("TextFocused is false on Specs")
	}
	d.key("esc")
	if v.TextFocused() {
		t.Fatal("TextFocused is true after the form closed")
	}
}

func TestProjectViewFormLongErrorRowFits(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			_, d, v := openForm(t, formRepo(t), size[0], size[1])
			v.form.message = "cannot write AGENTS.md: " + strings.Repeat("e", 200)
			v.form.messageOK = false
			d.mustShow("cannot write AGENTS.md", "…")
			assertFits(t, d, size[0], size[1])
		})
	}
}

func TestProjectViewSaysCreatedOrUpdatedAfterTheWrite(t *testing.T) {
	root := formRepo(t) // no AGENTS.md yet
	_, d, _ := openForm(t, root, 80, 24)
	d.key("enter")
	d.key("y")
	d.mustShow("Created AGENTS.md")
	d.mustNotShow("Updated AGENTS.md")
	d.key("e")
	d.key("down", "down", "down", "down", "down", "down", "right") // Delivery
	d.key("enter")
	d.key("y")
	d.mustShow("Updated AGENTS.md")
	d.mustNotShow("Created AGENTS.md")
}

func TestProjectViewKeepsTheClaudeMDWarningAfterTheWrite(t *testing.T) {
	root := formRepo(t)
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("own\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, d, _ := openForm(t, root, 80, 24)
	d.key("enter")
	d.key("y")
	d.mustShow("Valid", "CLAUDE.md does not import @AGENTS.md")
	assertFits(t, d, 80, 24)
	// Once CLAUDE.md imports AGENTS.md the warning goes away.
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("@AGENTS.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	d.key("r")
	d.mustShow("Valid")
	d.mustNotShow("does not import")
}

func TestProjectDoctorSectionDoesNotCarryTheClaudeMDWarning(t *testing.T) {
	root := formRepo(t)
	writeAgents(t, root, "# T\n\n## Hive\n\n- Project: p\n- Base branch: main\n- Tracker: t\n- Specs: _support/openspec\n")
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.md"), []byte("own\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(checkProject(root, newGitRunner()).Section.Lines, "\n"); strings.Contains(got, "CLAUDE.md") {
		t.Fatalf("the shared check changed hive doctor output:\n%s", got)
	}
}
