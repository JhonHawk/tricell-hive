package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// Tests of the Update view (T9). They drive the full application and compare
// with a twin home updated by `hive update` from the same repository.

// openMenuEntry opens the named main-menu entry from the menu.
func openMenuEntry(t *testing.T, d *appDriver, name string) {
	t.Helper()
	for i, item := range mainMenuItems {
		if item.label == name {
			for range i {
				d.key("down")
			}
			d.key("enter")
			// Render once before any typing: until a view's first View call
			// its text input still blinks, and each typed key would then run
			// a ~530 ms blink command synchronously in this driver (#59).
			d.screen()
			return
		}
	}
	t.Fatalf("no menu entry %q", name)
}

// typeText sends each rune of s as a key press.
func typeText(d *appDriver, s string) {
	d.t.Helper()
	for _, r := range s {
		if r == ' ' {
			d.key("space")
			continue
		}
		d.key(string(r))
	}
}

// updateTwinEnv is one Git repository and two homes installed from its first
// commit: home A for the application, home B for the command. The repository's
// HEAD is a second commit, changed unless unchanged is set.
type updateTwinEnv struct {
	repo                         string
	aHome, aState, bHome, bState string
}

func newUpdateTwinEnv(t *testing.T, unchanged bool) updateTwinEnv {
	t.Helper()
	requireGit(t)
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := updateTwinEnv{
		repo:   filepath.Join(base, "repo"),
		aHome:  filepath.Join(base, "a-home"),
		aState: filepath.Join(base, "a-state"),
		bHome:  filepath.Join(base, "b-home"),
		bState: filepath.Join(base, "b-state"),
	}
	for _, dir := range []string{e.repo, e.aHome, e.bHome} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	writeUpdateCatalog(t, e.repo, skillBody("Preserve evidence.\n"))
	gitInitLocal(t, e.repo)
	first := gitCommitAll(t, e.repo, "init")
	installFromCommit(t, e.aHome, e.aState, e.repo, first, []string{"codex", "claude"})
	installFromCommit(t, e.bHome, e.bState, e.repo, first, []string{"codex", "claude"})
	if unchanged {
		gitRun(t, e.repo, "commit", "-q", "--allow-empty", "-m", "same content")
	} else {
		putCharacterization(t, filepath.Join(e.repo, management.SkillSource), skillBody("Preserve evidence, updated.\n"))
		gitCommitAll(t, e.repo, "update skill")
	}
	return e
}

// commandUpdate is the twin: `hive update --source repo` on home B.
func (e updateTwinEnv) commandUpdate(t *testing.T, answers string) string {
	t.Helper()
	var out bytes.Buffer
	args := []string{"--home", e.bHome, "--state-dir", e.bState, "--source", e.repo}
	if err := update(args, strings.NewReader(answers), &out, true); err != nil {
		t.Fatalf("hive update: %v\n%s", err, out.String())
	}
	return out.String()
}

func (e updateTwinEnv) openApp(t *testing.T, width, height int) (*appModel, *appDriver) {
	t.Helper()
	cfg := hostsAppConfig(t, e.aHome, e.aState, e.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
	m, d := newTestApp(t, cfg, width, height)
	openMenuEntry(t, d, "Update")
	d.mustShow("Source", "Revision")
	return m, d
}

// setSource replaces the Source field's default "." with path.
func setSource(d *appDriver, path string) {
	d.t.Helper()
	d.key("backspace")
	typeText(d, path)
}

// TestUpdateViewAppliesLikeCommand covers AC7: Source and Revision in one view,
// Enter resolves the commit and shows the command's summary, and confirming
// leaves the same files and state as `hive update`.
func TestUpdateViewAppliesLikeCommand(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	commandOut := env.commandUpdate(t, "y\n")
	_, d := env.openApp(t, 80, 24)
	d.mustShow(".", "HEAD")
	setSource(d, env.repo)
	d.key("enter")
	d.mustShow("Source commit", "(requested HEAD)", "[Apply]")
	if len(menuRows(d)) != 0 {
		t.Fatal("the menu is drawn over the view")
	}
	// The summary is the command's: its "Source commit" line matches.
	line := ""
	for _, l := range strings.Split(commandOut, "\n") {
		if strings.HasPrefix(l, "Source commit ") {
			line = l
		}
	}
	if line == "" {
		t.Fatalf("hive update printed no Source commit line:\n%s", commandOut)
	}
	d.mustShow(line)
	d.key("enter")
	d.mustShow("Hive updated")
	assertTwin(t, env.aHome, env.aState, env.bHome, env.bState)
}

// TestUpdateViewRevisionField covers AC7: a revision other than HEAD is
// resolved by Git.
func TestUpdateViewRevisionField(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	_, d := env.openApp(t, 80, 24)
	setSource(d, env.repo)
	d.key("tab")
	for range len("HEAD") {
		d.key("backspace")
	}
	typeText(d, "HEAD~1")
	d.key("enter")
	// HEAD~1 is the commit the homes were installed from: nothing changes, so
	// the view applies directly. Resolving HEAD instead would ask to confirm.
	d.mustShow("Hive is already up to date")
	d.mustNotShow("[Apply]")
}

// TestUpdateViewUnchangedAppliesDirectly covers AC7: with nothing to change,
// the view applies directly, like the command, and shows its result.
func TestUpdateViewUnchangedAppliesDirectly(t *testing.T) {
	env := newUpdateTwinEnv(t, true)
	env.commandUpdate(t, "")
	_, d := env.openApp(t, 80, 24)
	setSource(d, env.repo)
	d.key("enter")
	d.mustShow("Hive is already up to date")
	d.mustNotShow("[Apply]")
	assertTwin(t, env.aHome, env.aState, env.bHome, env.bState)
}

// TestUpdateViewShowsResolvingWhileGitRuns covers AC8: the view shows a
// progress indicator while the commit resolves.
func TestUpdateViewShowsResolvingWhileGitRuns(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	m, d := env.openApp(t, 80, 24)
	setSource(d, env.repo)
	_, cmd := m.Update(keyMsg(t, "enter")) // the resolving Cmd is not run yet
	d.mustShow("Resolving HEAD")
	d.run(cmd)
	d.mustNotShow("Resolving HEAD")
	d.mustShow("[Apply]")
}

// mustShowFlat checks that the screen shows text, ignoring where a long
// message was wrapped.
func mustShowFlat(d *appDriver, text string) {
	d.t.Helper()
	if !strings.Contains(squash(d.screen()), squash(text)) {
		d.t.Fatalf("screen does not show %q:\n%s", text, d.screen())
	}
}

// TestUpdateViewErrorsShowInsideTheView covers AC7: an invalid revision, a
// source that is not a checkout and a home without CLIs show the command's
// message inside the view and change nothing.
func TestUpdateViewErrorsShowInsideTheView(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	before := collectFiles(t, env.aHome)

	t.Run("invalid revision", func(t *testing.T) {
		_, d := env.openApp(t, 80, 24)
		setSource(d, env.repo)
		d.key("tab")
		for range len("HEAD") {
			d.key("backspace")
		}
		typeText(d, "-x")
		d.key("enter")
		d.mustShow(`invalid --rev "-x"`, "Source", "Revision")
		d.mustNotShow("[Apply]", "Resolving")
	})

	t.Run("not a checkout", func(t *testing.T) {
		notRepo := t.TempDir()
		var out bytes.Buffer
		commandErr := update([]string{"--home", env.bHome, "--state-dir", env.bState, "--source", notRepo}, strings.NewReader("y\n"), &out, true)
		if commandErr == nil {
			t.Fatal("hive update accepted a source that is not a checkout")
		}
		_, d := env.openApp(t, 80, 24)
		setSource(d, notRepo)
		d.key("enter")
		mustShowFlat(d, commandErr.Error())
		d.mustNotShow("[Apply]", "Resolving")
	})

	t.Run("no CLI hosts", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		cfg := hostsAppConfig(t, home, stateDir, env.repo, defaultInstallDependencies(coreOnlyAdapterFactory))
		_, d := newTestApp(t, cfg, 80, 24)
		openMenuEntry(t, d, "Update")
		setSource(d, env.repo)
		d.key("enter")
		mustShowFlat(d, "no CLI hosts are registered with Hive for this home; run hive install first")
	})

	if got := collectFiles(t, env.aHome); len(got) != len(before) {
		t.Fatalf("a failed update changed the home: %d files before, %d after", len(before), len(got))
	}
}

// TestUpdateViewBackspaceEditsAndNeverGoesBack covers AC2 for text fields:
// Backspace deletes a character in either field and the view stays; Esc goes
// back.
func TestUpdateViewBackspaceEditsAndNeverGoesBack(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	_, d := env.openApp(t, 80, 24)
	typeText(d, "abc")
	d.mustShow(".abc")
	d.key("backspace")
	d.mustShow(".ab")
	d.mustNotShow(".abc")
	d.key("down") // to Revision
	d.key("backspace")
	d.mustShow("HEA")
	d.key("up")
	d.key("backspace", "backspace", "backspace")
	d.mustShow("Source", "Revision")
	if len(menuRows(d)) != 0 {
		t.Fatal("Backspace left the view")
	}
	d.key("esc")
	d.mustShow("> Update")
}

// TestUpdateViewDecliningKeepsHome covers AC8: rejecting the summary leaves the
// home unchanged with the message inside the view.
func TestUpdateViewDecliningKeepsHome(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	before := collectFiles(t, env.aHome)
	for _, reject := range [][]string{{"n"}, {"esc"}, {"backspace"}} {
		_, d := env.openApp(t, 80, 24)
		setSource(d, env.repo)
		d.key("enter")
		d.mustShow("[Apply]")
		d.key(reject...)
		d.mustShow("Cancelled. No changes applied.", "Source", "Revision")
		d.mustNotShow("[Apply]")
	}
	if got := collectFiles(t, env.aHome); len(got) != len(before) {
		t.Fatal("a declined update changed the home")
	}
	for rel, data := range before {
		if !bytes.Equal(collectFiles(t, env.aHome)[rel], data) {
			t.Fatalf("%s changed", rel)
		}
	}
}

// TestUpdateViewWriteBlocksKeysAndShowsResult covers AC2/AC8: while the apply
// runs the keys do nothing; afterwards the view shows the result.
func TestUpdateViewWriteBlocksKeysAndShowsResult(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	m, d := env.openApp(t, 80, 24)
	setSource(d, env.repo)
	d.key("enter")
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the apply is not marked as a write")
	}
	d.mustShow("Applying")
	before := d.screen()
	d.key("esc", "backspace", "enter", "ctrl+c", "n", "left")
	if d.quit {
		t.Fatal("Ctrl-C quit during a write")
	}
	if after := d.screen(); after != before {
		t.Fatalf("keys changed the screen during a write:\n%s", after)
	}
	d.hold = false
	d.release()
	d.mustShow("Hive updated", "Source", "Revision")
	d.mustNotShow("Applying")
}

// TestUpdateViewFits covers AC9: the view, its summary and its errors fit 80x24
// and 120x40 with a long source path.
func TestUpdateViewFits(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			env := newUpdateTwinEnv(t, false)
			_, d := env.openApp(t, width, height)
			assertFits(t, d, width, height)
			setSource(d, env.repo)
			typeText(d, strings.Repeat("x", 120))
			assertFits(t, d, width, height)
			d.key("enter") // the long path is not a checkout: an error
			assertFits(t, d, width, height)
			for range 120 {
				d.key("backspace")
			}
			d.key("enter")
			d.mustShow("[Apply]")
			assertFits(t, d, width, height)
		})
	}
}

// TestUpdateViewLongFieldShowsEllipsisAtTheCut covers a text field whose
// content scrolls: the cut edge shows "…", at the start while the tail is
// visible and at the end while the head is.
func TestUpdateViewLongFieldShowsEllipsisAtTheCut(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	_, d := env.openApp(t, 80, 24)
	sourceRow := func() string {
		for _, l := range d.lines() {
			if strings.Contains(l, "Source  ") {
				return strings.TrimRight(l, " ")
			}
		}
		t.Fatalf("no Source row:\n%s", d.screen())
		return ""
	}
	setSource(d, "/start"+strings.Repeat("-middle", 25)+"/end")
	row := sourceRow()
	if !strings.Contains(row, "Source    …") || !strings.HasSuffix(row, "/end") {
		t.Fatalf("the cut start is not marked with an ellipsis: %q", row)
	}
	d.key("home")
	row = sourceRow()
	if !strings.Contains(row, "Source    /start") || !strings.HasSuffix(row, "…") {
		t.Fatalf("the cut end is not marked with an ellipsis: %q", row)
	}
	assertFits(t, d, 80, 24)
	// A field that fits shows no ellipsis.
	d.key("end")
	for range 200 {
		d.key("backspace")
	}
	typeText(d, "/short")
	if row := sourceRow(); strings.Contains(row, "…") {
		t.Fatalf("a short field shows an ellipsis: %q", row)
	}
}

// TestUpdateViewEmptyFieldsSayWhichIsEmpty covers the view's own check: an
// empty Source or Revision gets a clear message, not the command's error with a
// leading gap.
func TestUpdateViewEmptyFieldsSayWhichIsEmpty(t *testing.T) {
	env := newUpdateTwinEnv(t, false)
	_, d := env.openApp(t, 80, 24)
	d.key("backspace") // Source: "." -> ""
	d.key("enter")
	d.mustShow("Source is empty")
	d.mustNotShow("Resolving", "is not a Git")
	setSource(d, env.repo)
	d.key("tab")
	for range len("HEAD") {
		d.key("backspace")
	}
	d.key("enter")
	d.mustShow("Revision is empty")
	d.mustNotShow("Resolving", "invalid --rev")
}

// hasReverse reports whether s draws a reversed cell, the cursor.
var reverseSGR = regexp.MustCompile("\x1b\\[(?:[0-9;]*;)?7(?:;[0-9;]*)?m")

func hasReverse(s string) bool { return reverseSGR.MatchString(s) }

// sourceField extracts the Source row's field text (plain) and its raw form from
// the app's screen.
func sourceField(t *testing.T, d *appDriver) (plain, raw string) {
	t.Helper()
	for i, l := range d.lines() {
		if strings.Contains(l, "Source  ") {
			raw = strings.Split(d.raw(), "\n")[i]
			return strings.TrimPrefix(strings.TrimPrefix(l, "> "), "  ")[len("Source    "):], raw
		}
	}
	t.Fatalf("no Source row:\n%s", d.screen())
	return "", ""
}

// TestInputViewCursorAtTheEndOfALongValueKeepsTheCursorAndNoTrailingEllipsis
// covers the field with the cursor at the end of content longer than the
// field: the start is marked, nothing is claimed to the right, the cursor cell
// is drawn, and the field stays within its width (80 columns leave 66).
func TestInputViewCursorAtTheEndOfALongValueKeepsTheCursorAndNoTrailingEllipsis(t *testing.T) {
	const fieldWidth = 66
	values := map[string]string{
		"distinct":            strings.Repeat("0123456789", 8),
		"repetitive":          strings.Repeat("Zeta", 40),
		"ends with spaces":    strings.Repeat("path-segment/", 8) + "  ",
		"just past the width": strings.Repeat("a", fieldWidth+1),
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			env := newUpdateTwinEnv(t, false)
			_, d := env.openApp(t, 80, 24)
			setSource(d, value)
			d.key("end")
			typeText(d, "xy")
			plain, raw := sourceField(t, d)
			if strings.HasSuffix(strings.TrimRight(plain, " "), "…") {
				t.Fatalf("a false trailing ellipsis with the cursor at the end: %q", plain)
			}
			if !strings.HasPrefix(plain, "…") {
				t.Fatalf("the cut start is not marked: %q", plain)
			}
			if !hasReverse(raw) {
				t.Fatalf("the cursor cell is not drawn: %q", raw)
			}
			if w := lipgloss.Width(plain); w > fieldWidth {
				t.Fatalf("the field is %d columns wide, its width is %d: %q", w, fieldWidth, plain)
			}
			assertFits(t, d, 80, 24)
		})
	}
}

// TestInputViewMarksEachCutOnlyWhereTextExists covers the cursor moving through
// a long repetitive value: the trailing marker appears only with text to the
// right, the leading one only with text to the left.
func TestInputViewMarksEachCutOnlyWhereTextExists(t *testing.T) {
	th := newAppTheme(true, false)
	in := textinput.New()
	in.Prompt = ""
	in.SetWidth(20)
	in.SetValue(strings.Repeat("Zeta", 20))
	in.Focus()
	for _, tc := range []struct {
		name       string
		pos        int
		start, end bool
	}{
		{"at the start", 0, false, true},
		{"in the middle", 40, true, true},
		{"at the end", 80, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in.SetCursor(tc.pos)
			view := inputView(in, &th)
			plain := stripANSI(view)
			if got := strings.HasPrefix(plain, "…"); got != tc.start {
				t.Errorf("leading ellipsis = %v, want %v: %q", got, tc.start, plain)
			}
			if got := strings.HasSuffix(strings.TrimRight(plain, " "), "…"); got != tc.end {
				t.Errorf("trailing ellipsis = %v, want %v: %q", got, tc.end, plain)
			}
			if w := lipgloss.Width(view); w > 20 {
				t.Errorf("the field is %d columns wide, its width is 20: %q", w, plain)
			}
			if !hasReverse(view) {
				t.Errorf("no cursor cell drawn: %q", view)
			}
		})
	}
	t.Run("a value that fits has no marker", func(t *testing.T) {
		in.SetValue("short")
		in.SetCursor(5)
		if plain := stripANSI(inputView(in, &th)); strings.Contains(plain, "…") {
			t.Fatalf("a short value shows an ellipsis: %q", plain)
		}
	})
}
