package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests for the findings of the independent UX review of the Models view and the
// Project confirmation (#46): H1, M1, M2, M3, L1, L2, L5 and L7.

const (
	longFrom = "github-copilot/gpt-6.1-sol"
	longTo   = "github-copilot/claude-opus-4.7"
)

const longToThinking = "github-copilot/claude-opus-4.7-thinking-high"

func longIDFake() *fakeCatalog {
	f := standardFake()
	f.out["opencode"] = longFrom + "\n" + longTo + "\n" + longToThinking + "\n"
	return f
}

// changeTableLines returns the confirmation's header line, its role rows and
// every line that is too wide, for rows starting with prefix.
func changeTableLines(d *appDriver, prefix string, width int) (headers, rows, wide []string) {
	for _, l := range d.lines() {
		if w := len([]rune(l)); w > width {
			wide = append(wide, l)
		}
		f := strings.Fields(l)
		switch {
		case len(f) >= 3 && f[0] == "Role" && f[1] == "Model" && f[2] == "Effort":
			headers = append(headers, l)
		case len(f) > 0 && strings.HasPrefix(f[0], prefix) && strings.Contains(l, "→"):
			rows = append(rows, l)
		}
	}
	return headers, rows, wide
}

func TestModelsConfirmationFitsEightyColumnsWithLongOpenCodeIDs(t *testing.T) {
	e := newModelsEditEnvWith(t, modelsTestSource(t), "claude,opencode", 80, 24, longIDFake().run)
	store(t, e.home, e.stateDir, "opencode", map[string]management.ModelOverride{longRoleName: {Model: longFrom}})
	e.d.key("r")
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, longRoleName)
	e.d.key("enter", "right")
	pick(e.d, "opus-4.7")
	e.d.key("enter")
	e.d.mustShow("[Apply]")
	headers, rows, wide := changeTableLines(e.d, longRoleName, 80)
	if len(headers) != 1 || len(rows) != 1 || len(wide) != 0 {
		t.Fatalf("headers %q, rows %q, too wide %q:\n%s", headers, rows, wide, e.d.screen())
	}
	if !strings.Contains(rows[0], "gpt-6.1-sol") || !strings.Contains(rows[0], "claude-opus-4.7") || !strings.Contains(rows[0], "→") {
		t.Fatalf("the row lost the model names: %q", rows[0])
	}
	assertFits(t, e.d, 80, 24)
}

func TestModelsConfirmationOfASevenRoleGroupKeepsOneRowPerRole(t *testing.T) {
	source, roles := bulkSource(t)
	e := newModelsEditEnvWith(t, source, "claude,opencode", 80, 24, longIDFake().run)
	prior := map[string]management.ModelOverride{}
	for _, r := range roles {
		prior[r] = management.ModelOverride{Model: longFrom}
	}
	store(t, e.home, e.stateDir, "opencode", prior)
	e.d.key("r")
	toHost(t, e.d, "opencode")
	e.d.key("enter", "right")
	pick(e.d, "thinking")
	e.d.key("enter")
	e.d.mustShow("[Apply]", "Change the Bulk group on opencode (7 roles)")
	headers, rows, wide := changeTableLines(e.d, "bulk-role-", 80)
	if len(headers) != 1 || len(rows) != 7 || len(wide) != 0 {
		t.Fatalf("headers %d, rows %d, too wide %q:\n%s", len(headers), len(rows), wide, e.d.screen())
	}
	assertFits(t, e.d, 80, 24)
}

func TestModelsCommandSummaryKeepsFullIDsAligned(t *testing.T) {
	c := newModelsEnv(t)
	out, err := c.write("set", true, "y\n", "--host", "opencode", "--role", "plain-role", "--model", longTo, "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "syn-oc/exec → "+longTo) {
		t.Fatalf("the command truncated an id:\n%s", out)
	}
}

// --- M1 -------------------------------------------------------------------------

func TestModelsViewSearchWithNoMatchSaysSoAndEnterDoesNothing(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)
	e.d.key("enter", "right")
	typeText(e.d, "zzzz")
	e.d.mustShow("No matching models", "release default", "Other…")
	e.d.key("enter")
	e.d.mustShow("Select model") // still open: nothing was applied
	if e.v.panel.touched {
		t.Fatal("Enter on an empty result chose an entry")
	}
	e.d.key("down")
	if got := pickerSelection(t, e); got != "release default" {
		t.Fatalf("↓ reaches %q, want the bottom entries", got)
	}
	e.d.key("right")
	if got := pickerSelection(t, e); got != "Other…" {
		t.Fatalf("→ reaches %q", got)
	}
	e.d.key("esc", "esc")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("files changed")
	}
}

// --- M2 -------------------------------------------------------------------------

func TestModelsViewListArrivingKeepsTheHighlightedEntryByIdentity(t *testing.T) {
	f := standardFake()
	f.out["codex"] = codexList(map[string][]string{"gpt-5.5": {"low"}, "syn-codex-exec": {"low", "medium"}}, "gpt-5.5", "syn-codex-exec")
	for _, c := range []struct {
		name string
		keys []string
		want string
	}{
		{"the current model", nil, "syn-codex-exec"},
		{"release default", []string{"down"}, "release default"},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := pickerEnv(t, "claude,codex", f, 80, 24)
			toHost(t, e.d, "codex")
			selectRole(t, e.d, "plain-role")
			cmd := e.v.openPanel() // the box opens before the list arrives
			e.d.key("right")
			e.d.key(c.keys...)
			if got := pickerSelection(t, e); got != c.want {
				t.Fatalf("before the list: %q, want %q", got, c.want)
			}
			e.d.run(cmd)
			if got := pickerSelection(t, e); got != c.want {
				t.Fatalf("after the list arrived the highlight moved to %q, want %q", got, c.want)
			}
		})
	}
}

// --- M3 -------------------------------------------------------------------------

func TestModelsViewValidationErrorDropsTheTechnicalPrefixButTheCommandKeepsIt(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	clearModelField(e.d)
	typeText(e.d, "a")
	e.d.key("space")
	typeText(e.d, "b")
	e.d.key("enter")
	e.d.mustShow("model may only contain")
	if strings.Contains(e.d.screen(), "model override for") {
		t.Fatalf("the panel shows the technical prefix:\n%s", e.d.screen())
	}
	for _, l := range e.d.lines() {
		if strings.Contains(l, "may only contain") && !strings.Contains(l, "only contain") {
			t.Fatalf("the actionable part is cut: %q", l)
		}
	}
	c := newModelsEnv(t)
	_, err := c.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--model", "a b")
	if err == nil || !strings.Contains(err.Error(), "model override for claude plain-role:") {
		t.Fatalf("the command lost its prefix: %v", err)
	}
}

// --- L1, L2 ---------------------------------------------------------------------

func TestModelsViewEmptyTextFieldShowsNoStrayEllipsis(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	clearModelField(e.d)
	for _, l := range e.d.lines() {
		if modelRowPattern.MatchString(l) && strings.Contains(l, "…") {
			t.Fatalf("an empty Model field shows %q", l)
		}
	}
}

func TestModelsViewTextModeHelpBarShowsTheTextKeys(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	clearModelField(e.d)
	lines := e.d.lines()
	help := lines[len(lines)-1]
	for _, bad := range []string{"→ choose model", "←/→ effort"} {
		if strings.Contains(help, bad) {
			t.Fatalf("the text-mode help bar %q still says %q", help, bad)
		}
	}
	for _, want := range []string{"type model", "↑/↓ field", "enter review", "esc close", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Errorf("text-mode help bar %q lacks %q", help, want)
		}
	}
	e.d.key("down") // Effort: its own keys come back
	lines = e.d.lines()
	if help = lines[len(lines)-1]; !strings.Contains(help, "←/→ effort") {
		t.Fatalf("the Effort row's help bar %q lacks ←/→ effort", help)
	}
}

// --- L7 -------------------------------------------------------------------------

func TestMenuDescriptionsSayModelsAndProjectAlsoEdit(t *testing.T) {
	want := map[string]string{
		"Models":  "See and change the model and effort of each role",
		"Project": "Check or edit this repository's ## Hive section",
	}
	for _, it := range mainMenuItems {
		if w, ok := want[it.label]; ok {
			if it.desc != w {
				t.Errorf("%s: %q, want %q", it.label, it.desc, w)
			}
			delete(want, it.label)
		}
	}
	if len(want) != 0 {
		t.Fatalf("missing entries: %v", want)
	}
	_, d := newTestApp(t, testAppConfig(t), 80, 24)
	d.mustShow("See and change the model and effort of each role", "Check or edit this repository's ## Hive section")
	assertFits(t, d, 80, 24)
}

// --- L5 -------------------------------------------------------------------------

func TestProjectConfirmationIndentsWrappedLinesUnderTheirItem(t *testing.T) {
	_, d, v := openForm(t, formRepo(t), 80, 24)
	d.key("down", "down") // Tracker
	var words []string
	for i := 0; i < 14; i++ {
		words = append(words, fmt.Sprintf("word%02d", i))
	}
	v.form.inputs[2].SetValue(strings.Join(words, " "))
	d.key("enter")
	d.mustShow("Write AGENTS.md", "After:")
	lines := d.lines()
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "  - Tracker: ") && strings.Contains(l, "word00") {
			start = i
		}
	}
	if start < 0 || start+1 >= len(lines) || !strings.Contains(lines[start+1], "word") {
		t.Fatalf("the Tracker line did not wrap as the test needs:\n%s", d.screen())
	}
	if cont := lines[start+1]; !strings.HasPrefix(cont, "    ") {
		t.Fatalf("the continuation %q is not indented under its item:\n%s", cont, d.screen())
	}
	assertFits(t, d, 80, 24)
}

// TestModelsViewCapturesTheLongIDConfirmationsForHandoff prints the two
// confirmations of the review's H1 at 80x24.
func TestModelsViewCapturesTheLongIDConfirmationsForHandoff(t *testing.T) {
	if os.Getenv("HIVE_CAPTURE") == "" {
		t.Skip("set HIVE_CAPTURE=1 to print the captures")
	}
	e := newModelsEditEnvWith(t, modelsTestSource(t), "claude,opencode", 80, 24, longIDFake().run)
	store(t, e.home, e.stateDir, "opencode", map[string]management.ModelOverride{longRoleName: {Model: longFrom}})
	e.d.key("r")
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, longRoleName)
	e.d.key("enter", "right")
	pick(e.d, "opus-4.7")
	e.d.key("enter")
	fmt.Printf("SINGLE ROLE\n%s\n", e.d.screen())

	source, roles := bulkSource(t)
	g := newModelsEditEnvWith(t, source, "claude,opencode", 80, 24, longIDFake().run)
	prior := map[string]management.ModelOverride{}
	for _, r := range roles {
		prior[r] = management.ModelOverride{Model: longFrom}
	}
	store(t, g.home, g.stateDir, "opencode", prior)
	g.d.key("r")
	toHost(t, g.d, "opencode")
	g.d.key("enter", "right")
	pick(g.d, "thinking")
	g.d.key("enter")
	fmt.Printf("SEVEN ROLES\n%s\n", g.d.screen())
}
