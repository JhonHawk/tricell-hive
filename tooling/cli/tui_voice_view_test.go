package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the Voice view (T9).

var voiceRowPattern = regexp.MustCompile(`^(> |  )(Voice|Address|Name|Intensity)  +(.*)$`)

type voiceRowState struct {
	cursor bool
	value  string
}

// voiceRows parses the Voice view's rows by label.
func voiceRows(d *appDriver) map[string]voiceRowState {
	rows := map[string]voiceRowState{}
	for _, line := range d.lines() {
		if m := voiceRowPattern.FindStringSubmatch(line); m != nil {
			value := strings.TrimSpace(m[3])
			value = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(value, "<"), ">"))
			rows[m[2]] = voiceRowState{cursor: m[1] == "> ", value: value}
		}
	}
	return rows
}

// voiceFixture is a home with Claude installed and a source with one voice.
type voiceFixture struct {
	home, stateDir, source string
	deps                   installDependencies
}

func newVoiceFixture(t *testing.T) voiceFixture {
	t.Helper()
	f := voiceFixture{source: minimalTestSource(t), deps: defaultInstallDependencies(coreOnlyAdapterFactory)}
	f.home, f.stateDir = newHostsTestHome(t)
	installViaText(t, f.home, f.stateDir, f.source, "claude", "y\n", f.deps)
	return f
}

func (f voiceFixture) command(t *testing.T, v *management.VoiceSetting) {
	t.Helper()
	plan, err := management.BuildVoicePlan("off", management.Options{Source: f.source, Home: f.home, StateDir: f.stateDir}, management.VoiceSetting{})
	if v != nil {
		plan, err = management.BuildVoicePlan("set", management.Options{Source: f.source, Home: f.home, StateDir: f.stateDir}, *v)
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(plan); err != nil {
		t.Fatal(err)
	}
}

func (f voiceFixture) open(t *testing.T, width, height int) (*appModel, *appDriver) {
	t.Helper()
	m, d := newTestApp(t, hostsAppConfig(t, f.home, f.stateDir, f.source, f.deps), width, height)
	openMenuEntry(t, d, "Voice")
	return m, d
}

// TestVoiceViewLoadsTheActiveVoice covers AC6: the rows show Voice, Address,
// Name (only with name) and Intensity, prefilled from the active voice.
func TestVoiceViewLoadsTheActiveVoice(t *testing.T) {
	f := newVoiceFixture(t)
	t.Run("off", func(t *testing.T) {
		_, d := f.open(t, 80, 24)
		rows := voiceRows(d)
		if rows["Voice"].value != "Off" || !rows["Voice"].cursor {
			t.Fatalf("rows = %+v\n%s", rows, d.screen())
		}
		if _, ok := rows["Address"]; ok {
			t.Fatalf("Address shown with the voice off:\n%s", d.screen())
		}
	})
	t.Run("set", func(t *testing.T) {
		f.command(t, &management.VoiceSetting{ID: testVoiceID, Address: "name", Name: "Robin", Intensity: "marked"})
		_, d := f.open(t, 80, 24)
		rows := voiceRows(d)
		if !strings.HasPrefix(rows["Voice"].value, testVoiceID) || rows["Address"].value != "name" || rows["Name"].value != "Robin" || rows["Intensity"].value != "marked" {
			t.Fatalf("rows = %+v\n%s", rows, d.screen())
		}
	})
}

// TestVoiceViewSetsThenTurnsOffLikeCommands covers AC6: ←→ change the values,
// ↑↓ the row, Enter shows the summary; the result equals `hive voice set`, and
// Off equals `hive voice off`.
func TestVoiceViewSetsThenTurnsOffLikeCommands(t *testing.T) {
	twin := newVoiceFixture(t)
	twin.command(t, &management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})

	f := newVoiceFixture(t)
	_, d := f.open(t, 80, 24)
	d.key("right") // Off -> testvoice
	if v := voiceRows(d)["Voice"].value; !strings.HasPrefix(v, testVoiceID) {
		t.Fatalf("Voice = %q\n%s", v, d.screen())
	}
	d.key("down", "right") // Address none -> sir
	if got := voiceRows(d)["Address"]; got.value != "sir" || !got.cursor {
		t.Fatalf("Address row = %+v\n%s", got, d.screen())
	}
	d.key("enter")
	d.mustShow("Voice: "+testVoiceID+" (address sir, intensity subtle)", "Voice files to change", "[Apply]")
	if len(voiceRows(d)) != 0 {
		t.Fatal("the rows are drawn over the summary")
	}
	d.key("enter")
	d.mustShow("Voice set")
	assertTwin(t, f.home, f.stateDir, twin.home, twin.stateDir)

	// Off.
	twin.command(t, nil)
	d.key("up", "left") // back to the Voice row: testvoice -> Off
	if voiceRows(d)["Voice"].value != "Off" {
		t.Fatalf("Voice = %q\n%s", voiceRows(d)["Voice"].value, d.screen())
	}
	d.key("enter")
	d.mustShow("Voice: off", "[Apply]")
	d.key("enter")
	d.mustShow("Voice turned off")
	assertTwin(t, f.home, f.stateDir, twin.home, twin.stateDir)
}

// TestVoiceViewNameRowAndBackspace covers AC6 and AC2: Name appears only with
// address name; typing edits it, Backspace deletes a character and never goes
// back, and the result equals the command with the same name.
func TestVoiceViewNameRowAndBackspace(t *testing.T) {
	twin := newVoiceFixture(t)
	twin.command(t, &management.VoiceSetting{ID: testVoiceID, Address: "name", Name: "Ana", Intensity: "marked"})

	f := newVoiceFixture(t)
	_, d := f.open(t, 80, 24)
	d.key("right", "down", "right") // sir
	if _, ok := voiceRows(d)["Name"]; ok {
		t.Fatalf("Name shown with address sir:\n%s", d.screen())
	}
	d.key("right") // name
	rows := voiceRows(d)
	if rows["Address"].value != "name" {
		t.Fatalf("rows = %+v", rows)
	}
	if _, ok := rows["Name"]; !ok {
		t.Fatalf("Name not shown with address name:\n%s", d.screen())
	}
	d.key("down") // Name row: the text field has the focus
	typeText(d, "Anaa")
	d.key("backspace")
	if voiceRows(d)["Name"].value != "Ana" {
		t.Fatalf("Name = %q\n%s", voiceRows(d)["Name"].value, d.screen())
	}
	d.mustShow("Address", "Intensity")
	d.mustNotShow("Main menu")
	d.key("down", "right") // Intensity subtle -> marked
	if voiceRows(d)["Intensity"].value != "marked" {
		t.Fatalf("rows = %+v", voiceRows(d))
	}
	d.key("enter")
	d.mustShow("(address name, intensity marked)", "[Apply]")
	d.key("enter")
	d.mustShow("Voice set")
	assertTwin(t, f.home, f.stateDir, twin.home, twin.stateDir)
}

// TestVoiceViewNameIsRequiredForAddressName covers the planner's own rule: a
// blank name with address name shows the planner's error inside the view.
func TestVoiceViewNameIsRequiredForAddressName(t *testing.T) {
	f := newVoiceFixture(t)
	before := collectFiles(t, f.home)
	_, d := f.open(t, 80, 24)
	d.key("right", "down", "right", "right", "enter")
	d.mustNotShow("[Apply]")
	d.mustShow("Voice", "Address")
	if got := collectFiles(t, f.home); len(got) != len(before) {
		t.Fatal("a rejected voice changed the home")
	}
	mustShowFlat(d, "name is required")
}

// TestVoiceViewNothingToChange covers AC6: with no change from the active
// voice, the view says there is nothing to apply.
func TestVoiceViewNothingToChange(t *testing.T) {
	f := newVoiceFixture(t)
	f.command(t, &management.VoiceSetting{ID: testVoiceID, Address: "sir", Intensity: "subtle"})
	before := collectFiles(t, f.home)
	_, d := f.open(t, 80, 24)
	d.key("enter")
	d.mustShow("Voice is already set this way")
	d.mustNotShow("[Apply]")
	if got := collectFiles(t, f.home); len(got) != len(before) {
		t.Fatal("the home changed")
	}
}

// TestVoiceViewEmptyStates covers the design table: no CLIs, and a source
// without voices.
func TestVoiceViewEmptyStates(t *testing.T) {
	t.Run("no CLIs", func(t *testing.T) {
		home, stateDir := newHostsTestHome(t)
		_, d := newTestApp(t, hostsAppConfig(t, home, stateDir, minimalTestSource(t), defaultInstallDependencies(coreOnlyAdapterFactory)), 80, 24)
		openMenuEntry(t, d, "Voice")
		d.mustShow("No CLI hosts are registered.")
		d.key("enter")
		d.mustShow("No CLI hosts are registered.")
	})
	t.Run("source without voices", func(t *testing.T) {
		f := newVoiceFixture(t)
		f.source = t.TempDir()
		_, d := f.open(t, 80, 24)
		mustShowFlat(d, "Run hive from a Hive checkout or package, or pass --source")
		if len(voiceRows(d)) != 0 {
			t.Fatal("rows shown without voices")
		}
	})
}

// TestVoiceViewDecliningAndWriteKeepFiles covers AC8 and AC2: rejecting the
// summary changes nothing; during the apply the keys do nothing.
func TestVoiceViewDecliningAndWriteKeepFiles(t *testing.T) {
	f := newVoiceFixture(t)
	before := collectFiles(t, f.home)
	for _, reject := range [][]string{{"n"}, {"esc"}, {"backspace"}} {
		_, d := f.open(t, 80, 24)
		d.key("right", "enter")
		d.mustShow("[Apply]")
		d.key(reject...)
		d.mustShow("Cancelled. No changes applied.", "Voice")
		d.mustNotShow("[Apply]")
	}
	if got := collectFiles(t, f.home); len(got) != len(before) {
		t.Fatal("a declined voice changed the home")
	}

	m, d := f.open(t, 80, 24)
	d.key("right", "enter")
	d.hold = true
	d.key("enter")
	if !m.isWriting() {
		t.Fatal("the apply is not marked as a write")
	}
	d.mustShow("Applying")
	screen := d.screen()
	d.key("esc", "backspace", "enter", "ctrl+c", "left", "down")
	if d.quit || d.screen() != screen {
		t.Fatal("keys acted during a write")
	}
	d.hold = false
	d.release()
	d.mustShow("Voice set", "Voice")
}

// TestVoiceViewFits covers AC9: with several voices with long descriptions and
// a long name, every row fits 80x24 and 120x40, the summary too.
func TestVoiceViewFits(t *testing.T) {
	src := copyMinimalSource(t)
	for i := 1; i <= 8; i++ {
		text := fmt.Sprintf("Voice%d: %s\n", i, strings.Repeat("a very warm and formal voice with a long description ", 4))
		if err := os.WriteFile(filepath.Join(src, "content", "voices", fmt.Sprintf("voice-number-%d.md", i)), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		width, height := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", width, height), func(t *testing.T) {
			f := newVoiceFixture(t)
			f.source = src
			_, d := f.open(t, width, height)
			assertFits(t, d, width, height)
			d.key("right", "right") // voice-number-1: a long description
			assertFits(t, d, width, height)
			if !strings.Contains(d.screen(), "…") {
				t.Fatalf("a long description was cut without an ellipsis:\n%s", d.screen())
			}
			d.key("down", "right", "right", "down")
			typeText(d, "María José de la Torre-Hernández y Fernández de Córdoba del Castillo")
			assertFits(t, d, width, height)
			d.key("enter") // the planner refuses a name over 40 characters
			mustShowFlat(d, "name must be at most 40 characters")
			assertFits(t, d, width, height)
			for range 70 {
				d.key("backspace")
			}
			typeText(d, "María José de la Torre-Hernández")
			d.key("enter")
			d.mustShow("[Apply]")
			assertFits(t, d, width, height)
		})
	}
}

// TestVoiceViewBackspaceWhilePlanningKeepsThePlanWhenNameHasTheFocus covers the
// same rule as the Update view: while the plan builds, Backspace does not
// silently abandon it when a text field has the focus; Esc leaves visibly.
func TestVoiceViewBackspaceWhilePlanningKeepsThePlanWhenNameHasTheFocus(t *testing.T) {
	f := newVoiceFixture(t)
	m, d := f.open(t, 80, 24)
	d.key("right", "down", "right", "right", "down") // Name row, focused
	typeText(d, "Ana")
	_, cmd := m.Update(keyMsg(t, "enter")) // the planning Cmd is not run yet
	d.mustShow("Planning the voice change")
	d.key("backspace")
	d.mustShow("Planning the voice change", "Address")
	d.mustNotShow("Main menu")
	d.run(cmd)
	d.mustShow("[Apply]")

	// With no text field focused, Backspace leaves like Esc, and the late
	// result is dropped.
	_, d = f.open(t, 80, 24)
	d.key("right")
	_, cmd = d.model.Update(keyMsg(t, "enter"))
	d.mustShow("Planning the voice change")
	d.key("backspace")
	d.mustShow("> Voice")
	d.run(cmd)
	d.mustNotShow("[Apply]")
}

// TestVoiceViewEditingClearsThePreviousResult covers the message under the
// rows: once the user edits a value, the result of the last apply goes away.
func TestVoiceViewEditingClearsThePreviousResult(t *testing.T) {
	t.Run("values", func(t *testing.T) {
		f := newVoiceFixture(t)
		_, d := f.open(t, 80, 24)
		d.key("right", "enter", "enter")
		d.mustShow("Voice set")
		d.key("down") // moving between rows is not an edit
		d.mustShow("Voice set")
		d.key("right") // Address: an edit
		d.mustNotShow("Voice set")
	})
	t.Run("name", func(t *testing.T) {
		f := newVoiceFixture(t)
		_, d := f.open(t, 80, 24)
		d.key("right", "down", "right", "right", "down")
		typeText(d, "Ana")
		d.key("enter", "enter")
		d.mustShow("Voice set")
		d.key("down", "up") // back on the Name row, its field focused
		d.key("x")
		d.mustNotShow("Voice set")
	})
	t.Run("the voice", func(t *testing.T) {
		f := newVoiceFixture(t)
		_, d := f.open(t, 80, 24)
		d.key("right", "enter", "enter")
		d.mustShow("Voice set")
		d.key("up", "left")
		d.mustNotShow("Voice set")
	})
}

// TestVoiceViewShowsAnActiveVoiceMissingFromTheSource covers a home whose active
// voice is not among this source's voices: the view shows it, marked as not in
// the source, and Enter does not propose turning the voice off.
func TestVoiceViewShowsAnActiveVoiceMissingFromTheSource(t *testing.T) {
	f := newVoiceFixture(t)
	// Activate a voice that only another source has.
	other := t.TempDir()
	if err := os.CopyFS(other, os.DirFS(f.source)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "content", "voices", "elsewhere.md"), []byte("Elsewhere: only in the other source.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := management.BuildVoicePlan("set", management.Options{Source: other, Home: f.home, StateDir: f.stateDir}, management.VoiceSetting{ID: "elsewhere", Address: "sir", Intensity: "marked"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(plan); err != nil {
		t.Fatal(err)
	}
	before := collectFiles(t, f.home)

	_, d := f.open(t, 80, 24) // f.source has no "elsewhere"
	rows := voiceRows(d)
	if !strings.HasPrefix(rows["Voice"].value, "elsewhere") || !strings.Contains(rows["Voice"].value, "not in this source") {
		t.Fatalf("the active voice is not shown as missing from the source: %+v\n%s", rows, d.screen())
	}
	if rows["Address"].value != "sir" || rows["Intensity"].value != "marked" {
		t.Fatalf("the active voice's settings were lost: %+v", rows)
	}
	d.key("enter")
	d.mustNotShow("Turn the voice off", "Voice: off")
	d.mustNotShow("[Apply]")
	if got := collectFiles(t, f.home); len(got) != len(before) {
		t.Fatal("reviewing changed the home")
	}
	// Off and the source's own voices stay reachable.
	d.key("right")
	if v := voiceRows(d)["Voice"].value; v != "Off" {
		t.Fatalf("→ from the last voice should wrap to Off, got %q", v)
	}
}
