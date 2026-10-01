package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"tricell-hive/tooling/management"
)

// Tests of the model picker and the role groups in the Models view (#46, T9):
// the list the Model field opens, the shared catalog cache, the group headers
// and the group panel. Every runner is a fake and every home is synthetic.

// fakeCatalog is a catalogRunner with recorded outputs. fail counts the leading
// calls of a binary that fail, to check that a failure is retried.
type fakeCatalog struct {
	mu    sync.Mutex
	out   map[string]string
	fail  map[string]int
	calls []string
}

func (f *fakeCatalog) run(_ context.Context, bin string, _ []string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, bin)
	if f.fail[bin] > 0 {
		f.fail[bin]--
		return nil, errors.New("the " + bin + " listing failed")
	}
	out, ok := f.out[bin]
	if !ok {
		return nil, errors.New("no recorded output for " + bin)
	}
	return []byte(out), nil
}

func (f *fakeCatalog) count(bin string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c == bin {
			n++
		}
	}
	return n
}

func (f *fakeCatalog) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// codexList is a recorded-shape output of `codex debug models`.
func codexList(models map[string][]string, order ...string) string {
	var parts []string
	for _, id := range order {
		var levels []string
		for _, e := range models[id] {
			levels = append(levels, `{"effort":"`+e+`"}`)
		}
		name := ""
		if n, ok := codexNames[id]; ok {
			name = `"display_name":"` + n + `",`
		}
		parts = append(parts, `{"slug":"`+id+`",`+name+`"visibility":"list","supported_reasoning_levels":[`+strings.Join(levels, ",")+`]}`)
	}
	return `{"models":[` + strings.Join(parts, ",") + `]}`
}

// codexNames are the display names the fake `codex debug models` gives.
var codexNames = map[string]string{"gpt-6.1-sol": "GPT-6.1 Sol", "gpt-5.5": "GPT-5.5", "syn-codex-exec": "Synthetic Exec"}

var allEfforts = []string{"low", "medium", "high", "xhigh", "max", "ultra"}

func standardFake() *fakeCatalog {
	return &fakeCatalog{
		fail: map[string]int{},
		out: map[string]string{
			"codex": codexList(map[string][]string{
				"gpt-6.1-sol": allEfforts,
				"gpt-5.5":     {"low", "medium", "high", "xhigh"},
			}, "gpt-6.1-sol", "gpt-5.5"),
			"opencode": "x/y\nopenai/gpt-5.5\nanthropic/claude-sonnet-4.5\n",
			"grok":     "Available models:\n  * grok-4.7 (default)\n  - grok-4.6\n",
			"pi":       "provider model context\np m 1\n",
		},
	}
}

func pickerEnv(t *testing.T, hosts string, f *fakeCatalog, width, height int) modelsEditEnv {
	t.Helper()
	var run catalogRunner
	if f != nil {
		run = f.run
	}
	return newModelsEditEnvWith(t, modelsTestSource(t), hosts, width, height, run)
}

// boxInterior returns the picker box's lines, the top border and the bottom one
// included, without the side borders and with runs of spaces collapsed, or fails
// when the box is not open.
func boxInterior(t *testing.T, d *appDriver) []string {
	t.Helper()
	var out []string
	in := false
	for _, l := range d.lines() {
		switch {
		case strings.HasPrefix(l, "┌"):
			in = true
		case strings.HasPrefix(l, "└"):
			return out
		}
		if in {
			l = strings.TrimSuffix(strings.TrimPrefix(strings.TrimRight(l, " "), "│"), "│")
			out = append(out, strings.Join(strings.Fields(l), " "))
		}
	}
	t.Fatalf("the model picker is not open:\n%s", d.screen())
	return nil
}

// listItems returns what the picker lists: section headers and models in order
// (a "●" prefix marks the current model), then its two bottom entries. The
// borders, search and status rows and the blank separators are left out.
func listItems(t *testing.T, d *appDriver) []string {
	t.Helper()
	in := boxInterior(t, d)
	var items []string
	for i, l := range in {
		if i < 3 || i == len(in)-1 || l == "" { // top border, search, status; the bottom entries
			continue
		}
		items = append(items, l)
	}
	return append(items, "release default", "Other…")
}

func boxStatus(t *testing.T, d *appDriver) string {
	t.Helper()
	return boxInterior(t, d)[2]
}

func mustHaveNoList(t *testing.T, d *appDriver) {
	t.Helper()
	if strings.Contains(d.screen(), "Select model") {
		t.Fatalf("the model picker is open:\n%s", d.screen())
	}
}

// pickerSelection is the entry the picker's cursor is on: a model id, or one of
// the two bottom entries. The highlight is color, so the tests read the state.
func pickerSelection(t *testing.T, e modelsEditEnv) string {
	t.Helper()
	l := e.v.panel.list
	c := e.v.pickerContent()
	switch {
	case l.cur < 0:
		return ""
	case l.cur < len(c.rows):
		return c.rows[l.cur].id
	case l.cur == len(c.rows):
		return "release default"
	}
	return "Other…"
}

// pickDefault chooses "release default", the first of the two bottom entries.
func pickDefault(d *appDriver) {
	d.t.Helper()
	d.key("end", "left", "enter")
}

// pickOther chooses "Other…", the last entry.
func pickOther(d *appDriver) {
	d.t.Helper()
	d.key("end", "enter")
}

// pick filters the open list by text and chooses the first entry.
func pick(d *appDriver, filter string) {
	d.t.Helper()
	typeText(d, filter)
	d.key("enter")
}

// selectedLine is the text of the table line that carries the cursor.
func selectedLine(t *testing.T, d *appDriver) string {
	t.Helper()
	for _, l := range d.lines() {
		if strings.HasPrefix(l, "> ") && !panelRowPattern.MatchString(l) {
			return strings.TrimSpace(strings.TrimPrefix(l, "> "))
		}
	}
	t.Fatalf("no line carries the cursor:\n%s", d.screen())
	return ""
}

// toHost presses → until the CLI is selected.
func toHost(t *testing.T, d *appDriver, host string) {
	t.Helper()
	for range 8 {
		cur := selectedModelsHost(t, d)
		if cur == host {
			return
		}
		if cur > host {
			d.key("left")
		} else {
			d.key("right")
		}
	}
	t.Fatalf("cannot reach %s", host)
}

// --- the selector -----------------------------------------------------------------

func TestModelsViewArrowOpensTheModelListInOrder(t *testing.T) {
	f := standardFake()
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	// The models the CLI's roles use come first, the current one marked; the
	// CLI's own list follows, without repeating them; two entries close the box.
	want := []string{"● syn-claude-exec", "fable", "opus", "sonnet", "haiku", "release default", "Other…"}
	got := listItems(t, e.d)
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("list = %q, want %q\n%s", got, want, e.d.screen())
	}
	e.d.mustShow("Select model", "claude · plain-role", "esc", "Search")
	if got := pickerSelection(t, e); got != "syn-claude-exec" { // the cursor starts on the current model
		t.Fatalf("the cursor is on %q", got)
	}
	if strings.Contains(e.d.screen(), "Role   ") && strings.Contains(e.d.screen(), "Profile") {
		t.Fatalf("the table header is still shown:\n%s", e.d.screen())
	}
	assertFits(t, e.d, 80, 24)
}

func TestModelsViewListDoesNotRepeatAnEffectiveModelTheCLIListsAndKeepsCatalogOrder(t *testing.T) {
	f := standardFake()
	f.out["codex"] = codexList(map[string][]string{"syn-codex-exec": {"low", "medium"}, "gpt-5.5": {"low"}}, "gpt-5.5", "syn-codex-exec")
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	want := "GPT-5.5 gpt-5.5|● Synthetic Exec syn-codex-exec|release default|Other…"
	if got := strings.Join(listItems(t, e.d), "|"); got != want {
		t.Fatalf("list = %q, want %q", got, want)
	}
}

func TestModelsViewLetterOpensTheListAndFiltersCaseInsensitively(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "o")
	e.d.mustShow("Search: o")
	e.d.key("P")
	e.d.mustShow("Search: oP")
	got := listItems(t, e.d)
	if strings.Join(got, "|") != "opus|release default|Other…" {
		t.Fatalf("filtered list = %q", got)
	}
	e.d.key("backspace")
	e.d.mustShow("Search: o")
	if e.m.top() != view(e.v) {
		t.Fatal("Backspace in the list left the view")
	}
}

func TestModelsViewChoosingAModelReturnsToThePanelAndEnterReviews(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	twinHome, twinState := e.twin(t)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pick(e.d, "opus")
	mustHaveNoList(t, e.d)
	if got, cursor := panelField(t, e.d, "Model"); got != "opus" || !cursor {
		t.Fatalf("Model = %q (cursor %v)\n%s", got, cursor, e.d.screen())
	}
	e.d.key("enter")
	e.d.mustShow("Change plain-role on claude", "syn-claude-exec → opus", "[Apply]")
	e.d.key("enter")
	e.d.mustShow("Open sessions keep the previous model until they restart")
	command(t, twinHome, twinState, "set", "--host", "claude", "--role", "plain-role", "--model", "opus")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
}

func TestModelsViewReleaseDefaultEqualsResetOnlyModel(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	twinHome, twinState := e.twin(t)
	prior := map[string]management.ModelOverride{"plain-role": {Model: "opus", Effort: "max"}}
	store(t, e.home, e.stateDir, "claude", prior)
	store(t, twinHome, twinState, "claude", prior)
	e.d.key("r")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pickDefault(e.d)
	if got, _ := panelField(t, e.d, "Model"); got != "release default" {
		t.Fatalf("Model = %q", got)
	}
	e.d.key("enter", "enter")
	e.d.mustShow("Open sessions keep the previous model until they restart")
	command(t, twinHome, twinState, "reset", "--host", "claude", "--role", "plain-role", "--only", "model")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
}

func TestModelsViewEscClosesTheListThenThePanelThenTheView(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)
	e.d.key("enter", "right", "o", "esc")
	mustHaveNoList(t, e.d)
	if got, cursor := panelField(t, e.d, "Model"); got != "syn-claude-exec" || !cursor {
		t.Fatalf("Esc changed the model: %q (cursor %v)", got, cursor)
	}
	e.d.key("esc")
	mustNoPanel(t, e.d)
	if e.m.top() != view(e.v) {
		t.Fatal("the second Esc left the view")
	}
	e.d.key("esc")
	if e.m.top() == view(e.v) {
		t.Fatal("the third Esc did not leave the view")
	}
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("files changed")
	}
}

func TestModelsViewOtherTurnsModelIntoTheTextField(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	twinHome, twinState := e.twin(t)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pickOther(e.d)
	if !e.v.TextFocused() {
		t.Fatalf("the text field has no focus:\n%s", e.d.screen())
	}
	if got, _ := panelField(t, e.d, "Model"); !strings.HasPrefix(got, "syn-claude-exec") {
		t.Fatalf("Model = %q, want the current value in the field", got)
	}
	clearModelField(e.d)
	typeText(e.d, "abc")
	e.d.key("left", "left")
	typeText(e.d, "X")
	if got, _ := panelField(t, e.d, "Model"); !strings.HasPrefix(got, "aXbc") {
		t.Fatalf("← did not move the text cursor: %q", got)
	}
	e.d.key("enter")
	e.d.mustShow("[Apply]")
	e.d.key("enter")
	command(t, twinHome, twinState, "set", "--host", "claude", "--role", "plain-role", "--model", "aXbc")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
}

func TestModelsViewEnterInThePanelStillReviewsOnGrok(t *testing.T) {
	e := pickerEnv(t, "claude,codex,grok", standardFake(), 80, 24)
	toHost(t, e.d, "grok")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	e.d.key("enter")
	e.d.mustShow("Nothing to change")
	e.d.key("right")
	e.d.mustShow("grok-4.6")
	pick(e.d, "4.6")
	e.d.key("enter")
	e.d.mustShow("Change plain-role on grok", "host default → grok-4.6", "[Apply]")
}

func TestModelsViewTwoHundredNineModelsReachTheLastOneWithPageDown(t *testing.T) {
	f := standardFake()
	var lines []string
	for i := 1; i <= 209; i++ {
		lines = append(lines, fmt.Sprintf("p/m-%03d", i))
	}
	f.out["opencode"] = strings.Join(lines, "\n") + "\n"
	e := pickerEnv(t, "claude,opencode", f, 80, 24)
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	e.d.mustNotShow("m-209")
	for range 40 {
		e.d.key("pgdown")
	}
	e.d.mustShow("m-209", "Other…") // shown without its "p/" provider
	e.d.key("home")
	e.d.mustShow("release default")
	e.d.key("end")
	if got := pickerSelection(t, e); got != "Other…" {
		t.Fatalf("End leaves the cursor on %q", got)
	}
	assertFits(t, e.d, 80, 24)
}

// --- efforts by model on Codex -------------------------------------------------

func effortCycle(t *testing.T, d *appDriver) []string {
	t.Helper()
	d.key("down")
	seen := []string{}
	for range 8 {
		v, _ := panelField(t, d, "Effort")
		seen = append(seen, strings.Trim(v, "<> "))
		d.key("right")
	}
	d.key("up")
	return seen
}

func TestModelsViewCodexEffortsFollowTheChosenModel(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	pick(e.d, "gpt-5.5")
	got := strings.Join(effortCycle(t, e.d), ",")
	if strings.Contains(got, "max") || strings.Contains(got, "ultra") || !strings.Contains(got, "xhigh") {
		t.Fatalf("efforts after choosing gpt-5.5: %s", got)
	}
	// A model with every level offers every level again.
	e.d.key("right")
	pick(e.d, "gpt-6.1")
	if got := strings.Join(effortCycle(t, e.d), ","); !strings.Contains(got, "ultra") {
		t.Fatalf("efforts after choosing gpt-6.1-sol: %s", got)
	}
}

func TestModelsViewAnEffortTheNewModelLacksFallsBackToReleaseDefault(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	toHost(t, e.d, "codex")
	store(t, e.home, e.stateDir, "codex", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if got, _ := panelField(t, e.d, "Effort"); got != "< max >" {
		t.Fatalf("Effort = %q", got)
	}
	e.d.key("right")
	pick(e.d, "gpt-5.5")
	if got, _ := panelField(t, e.d, "Effort"); got != "< release default >" {
		t.Fatalf("Effort = %q, want the release default", got)
	}
}

func TestModelsViewACatalogArrivingWithThePanelOpenKeepsTheChosenEffort(t *testing.T) {
	f := standardFake()
	f.out["codex"] = codexList(map[string][]string{"syn-codex-exec": {"low", "medium"}}, "syn-codex-exec")
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)

	// The panel opens before the list arrives.
	cmd := e.v.openPanel()
	if cmd == nil {
		t.Fatal("opening the panel did not start the query")
	}
	e.d.key("down", "right") // medium -> high, a level the model does not list
	if got, _ := panelField(t, e.d, "Effort"); got != "< high >" {
		t.Fatalf("Effort = %q", got)
	}
	e.d.run(cmd)
	if got, _ := panelField(t, e.d, "Effort"); got != "< high >" {
		t.Fatalf("the catalog moved the chosen effort to %q", got)
	}

	// With the effort untouched, a catalog that arrives late asks for nothing.
	e.d.key("esc")
	e.m.cfg.catalog = newModelCatalogCache() // forget the list: a new query will run
	e.v.cfg = e.m.cfg
	cmd = e.v.openPanel()
	e.d.run(cmd)
	e.d.key("enter")
	e.d.mustShow("Nothing to change")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("files changed")
	}
}

// --- loading, failure, and the shared cache -------------------------------------

func TestModelsViewFailedListShowsTheReasonAndKeepsOther(t *testing.T) {
	f := standardFake()
	f.fail["codex"] = 1
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	e.d.mustShow("Model list unavailable: the codex listing failed")
	got := listItems(t, e.d)
	if got[len(got)-1] != "Other…" {
		t.Fatalf("list = %q, want Other… kept", got)
	}
	assertFits(t, e.d, 80, 24)
}

func TestModelsViewCatalogLoadsOncePerCLIAndAFailureIsRetried(t *testing.T) {
	f := standardFake()
	f.fail["codex"] = 1
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if f.count("codex") != 1 {
		t.Fatalf("the first open ran the query %d times", f.count("codex"))
	}
	e.d.key("esc", "enter") // the failure was not kept: the next open retries
	if f.count("codex") != 2 {
		t.Fatalf("the failed query ran %d times, want a retry", f.count("codex"))
	}
	e.d.key("right")
	e.d.mustShow("gpt-5.5")
	e.d.mustNotShow("Model list unavailable")
	e.d.key("esc", "esc", "enter")
	if f.count("codex") != 2 {
		t.Fatalf("a stored list was queried again: %d calls", f.count("codex"))
	}
	// Leaving the view and coming back keeps the list for the session.
	e.d.key("esc", "esc")
	if e.m.top() == view(e.v) {
		t.Fatal("still in the view")
	}
	e.d.send(pushViewMsg{v: newModelsView(e.m.cfg)})
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	e.d.mustShow("gpt-5.5")
	if f.count("codex") != 2 {
		t.Fatalf("a new Models view queried again: %d calls", f.count("codex"))
	}
}

func TestModelsViewSpinnerRunsWhileTheListLoads(t *testing.T) {
	f := standardFake()
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	if e.v.NeedsSpinner() {
		t.Fatal("the view spins before any panel opens")
	}
	cmd := e.v.openPanel()
	if !e.v.NeedsSpinner() {
		t.Fatal("the view does not spin while the list loads")
	}
	e.d.key("right")
	e.d.mustShow("Loading models…")
	e.d.run(cmd)
	if e.v.NeedsSpinner() {
		t.Fatal("the view still spins after the list arrived")
	}
	e.d.mustNotShow("Loading models…")
	e.d.mustShow("gpt-5.5")
}

func TestModelsViewAListThatArrivesAfterLeavingIsKept(t *testing.T) {
	f := standardFake()
	e := pickerEnv(t, "claude,codex", f, 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	cmd := e.v.openPanel()
	e.d.key("esc", "esc") // leave the view while the query runs
	if e.m.top() == view(e.v) {
		t.Fatal("still in the view")
	}
	e.d.run(cmd)
	if _, ok := e.m.cfg.catalog.get("codex"); !ok {
		t.Fatal("the list that finished after leaving was lost")
	}
	e.d.send(pushViewMsg{v: newModelsView(e.m.cfg)})
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if f.count("codex") != 1 {
		t.Fatalf("the query ran %d times", f.count("codex"))
	}
}

func TestModelsViewUnderASyntheticHomeRunsNothingAndOffersTheCurrentValueAndOther(t *testing.T) {
	old := defaultCatalogRunner
	calls := 0
	defaultCatalogRunner = func(context.Context, string, []string) ([]byte, error) {
		calls++
		return nil, errors.New("must not run")
	}
	t.Cleanup(func() { defaultCatalogRunner = old })
	queries := countCatalogQueries(t)
	e := pickerEnv(t, "claude,codex", nil, 80, 24) // no fake: the app's own selection under --home
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	got := strings.Join(listItems(t, e.d), "|")
	if got != "● syn-codex-exec|release default|Other…" {
		t.Fatalf("list = %q", got)
	}
	e.d.mustNotShow("unavailable")
	if calls != 0 || *queries != 0 {
		t.Fatalf("a CLI was listed under --home: %d runner calls, %d queries", calls, *queries)
	}
}

// TestModelsViewBrowsingNeverQueriesAList is AC13 for the view: without an open
// panel nothing is listed, by any path.
func TestModelsViewBrowsingNeverQueriesAList(t *testing.T) {
	f := standardFake()
	queries := countCatalogQueries(t)
	e := pickerEnv(t, "claude,codex,grok,opencode,pi", f, 80, 24)
	for range 4 {
		e.d.key("right")
		e.d.key("down", "down", "up", "pgdown", "pgup", "end", "home")
	}
	e.d.key("r")
	e.d.key("x") // not overridden: nothing happens
	assertFits(t, e.d, 80, 24)
	if f.total() != 0 || *queries != 0 {
		t.Fatalf("browsing the view ran %d listing commands and %d queries", f.total(), *queries)
	}
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if f.count("codex") != 1 || *queries != 1 {
		t.Fatalf("opening a panel ran %d commands, %d queries; want 1 and 1", f.count("codex"), *queries)
	}
}

// --- groups ---------------------------------------------------------------------

func TestModelsViewGroupHeadersAppearInOrderAndTheCursorLandsOnThem(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	lines := e.d.lines()
	design, quality, hive := -1, -1, -1
	for i, l := range lines {
		switch {
		case strings.Contains(l, "Design"):
			design = i
		case strings.Contains(l, "Quality"):
			quality = i
		case strings.Contains(l, longRoleName):
			hive = i
		}
	}
	if !(design >= 0 && design < hive && hive < quality) {
		t.Fatalf("headers out of order (design %d, role %d, quality %d):\n%s", design, hive, quality, e.d.screen())
	}
	if got := selectedLine(t, e.d); !strings.HasPrefix(got, "▾ Design") {
		t.Fatalf("the cursor starts on %q, want the Design header", got)
	}
	e.d.key("down", "down", "down")
	if got := selectedLine(t, e.d); !strings.HasPrefix(got, "▾ Quality") {
		t.Fatalf("after three downs the cursor is on %q, want the Quality header", got)
	}
	// The header sits in the table's columns: the group under Role, and the
	// common model and effort under Model and Effort.
	line := ""
	for _, l := range e.d.lines() {
		if strings.Contains(l, "Design") {
			line = l
		}
	}
	if f := strings.Fields(line); strings.Join(f, " ") != "▾ Design mixed mixed" && strings.Join(f, " ") != "> ▾ Design mixed mixed" {
		t.Fatalf("header line = %q", line)
	}
	head := ""
	for _, l := range e.d.lines() {
		if strings.Contains(l, "Profile") && strings.Contains(l, "Effort") {
			head = l
		}
	}
	col := func(text, sub string) int { return utf8.RuneCountInString(text[:strings.Index(text, sub)]) } // "▾" is one column
	if col(line, "mixed") != col(head, "Model") {
		t.Fatalf("the header's model is not under the Model heading:\n%s", e.d.screen())
	}
}

func TestModelsViewHeaderCarriesTheMarkWhenARoleIsOverridden(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	e.d.mustShow("▾ Design *")
	e.d.mustNotShow("Quality *")
}

func TestModelsViewEnterOnAHeaderOpensTheGroupPanelWithMixedValues(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	e.d.key("enter")
	if got, _ := panelField(t, e.d, "Role"); got != "Design group on claude (2 roles)" {
		t.Fatalf("Role = %q\n%s", got, e.d.screen())
	}
	if got, cursor := panelField(t, e.d, "Model"); got != "mixed" || !cursor {
		t.Fatalf("Model = %q (cursor %v)", got, cursor)
	}
	if got, _ := panelField(t, e.d, "Effort"); got != "< mixed >" {
		t.Fatalf("Effort = %q", got)
	}
	assertFits(t, e.d, 80, 24)
	// Nothing touched: nothing to change.
	e.d.key("enter")
	e.d.mustShow("Nothing to change")
	// A group whose roles agree starts on their common values.
	e.d.key("esc", "down", "down", "down", "enter")
	if got, _ := panelField(t, e.d, "Model"); got != "inherit" {
		t.Fatalf("Quality Model = %q", got)
	}
}

func TestModelsViewGroupApplyEqualsSetGroup(t *testing.T) {
	cases := []struct {
		name string
		host string
		keys func(d *appDriver)
		cmd  []string
	}{
		{"claude model", "claude", func(d *appDriver) { d.key("right"); pick(d, "opus") },
			[]string{"set", "--host", "claude", "--group", "design", "--model", "opus"}},
		{"claude model and effort", "claude", func(d *appDriver) { d.key("right"); pick(d, "opus"); d.key("down", "right", "right", "right") },
			[]string{"set", "--host", "claude", "--group", "design", "--model", "opus", "--effort", "medium"}},
		{"opencode model over mixed efforts", "opencode", func(d *appDriver) { d.key("right"); pick(d, "x/y") },
			[]string{"set", "--host", "opencode", "--group", "design", "--model", "x/y"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := pickerEnv(t, "claude,opencode", standardFake(), 80, 24)
			twinHome, twinState := e.twin(t)
			toHost(t, e.d, c.host)
			e.d.key("enter")
			c.keys(e.d)
			e.d.key("enter")
			e.d.mustShow("[Apply]")
			e.d.key("enter")
			e.d.mustShow("Open sessions keep the previous model until they restart")
			command(t, twinHome, twinState, c.cmd[0], c.cmd[1:]...)
			assertTwin(t, e.home, e.stateDir, twinHome, twinState)
			if e.m.top() != view(e.v) {
				t.Fatalf("top view is %T", e.m.top())
			}
		})
	}
}

func TestModelsViewGroupConfirmationStartsWithTheReplacedOverrides(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	e.d.key("enter", "right")
	pick(e.d, "opus")
	e.d.key("enter")
	lines := e.d.lines()
	replaced, last := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "Replaces the own override of plain-role.") && replaced < 0 {
			replaced = i
		}
		if strings.HasPrefix(l, "plain-role ") {
			last = i
		}
	}
	if replaced < 0 || last < 0 || replaced < last {
		t.Fatalf("the confirmation does not put the replaced overrides after the table:\n%s", e.d.screen())
	}
	assertFits(t, e.d, 80, 24)
}

func TestModelsViewXOnAHeaderEqualsResetGroup(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	twinHome, twinState := e.twin(t)
	prior := map[string]management.ModelOverride{"plain-role": {Model: "opus"}, "hive-design-architecture": {Effort: "low"}, "inherit-role": {Effort: "high"}}
	store(t, e.home, e.stateDir, "claude", prior)
	store(t, twinHome, twinState, "claude", prior)
	e.d.key("r")
	e.d.key("x")
	e.d.mustShow("Reset the Design group on claude (2 roles)", "opus →", "hive-design-architecture", "[Apply]")
	e.d.mustNotShow("inherit-role:")
	e.d.key("y")
	e.d.mustShow("Open sessions keep the previous model until they restart")
	command(t, twinHome, twinState, "reset", "--host", "claude", "--group", "design")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
	e.d.mustNotShow("Design *")
	e.d.key("down", "down", "down")
	e.d.mustShow("Quality *")
}

func TestModelsViewBackspaceOnEffortOrInTheListNeverClosesTheView(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	e.d.key("enter", "down", "backspace")
	if e.m.top() != view(e.v) {
		t.Fatal("Backspace on Effort closed the view")
	}
	e.d.key("up", "right", "backspace", "backspace")
	if e.m.top() != view(e.v) {
		t.Fatal("Backspace in the list closed the view")
	}
	e.d.mustShow("Select model")
}

func TestModelsViewFooterNamesRolesAndGroups(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	e.d.mustShow(
		"Enter edits the selected role or group; x resets one marked *.",
		"Defaults come from integrations/agent-profiles.json in the release.",
	)
}

func TestModelsViewHelpBarWithTheListOpen(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	lines := e.d.lines()
	help := lines[len(lines)-1]
	for _, want := range []string{"type search", "↑/↓ model", "enter choose", "esc close", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Errorf("help bar %q lacks %q", help, want)
		}
	}
}

func TestModelsViewFitsTheListTheMixedGroupAndTheGroupConfirmation(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			f := standardFake()
			var lines []string
			for i := 1; i <= 209; i++ {
				lines = append(lines, fmt.Sprintf("provider-with-a-long-name/model-with-a-long-name-%03d", i))
			}
			f.out["opencode"] = strings.Join(lines, "\n") + "\n"
			e := pickerEnv(t, "claude,opencode", f, size[0], size[1])
			assertFits(t, e.d, size[0], size[1])
			e.d.key("enter") // the group panel, with mixed values
			assertFits(t, e.d, size[0], size[1])
			e.d.key("right")
			assertFits(t, e.d, size[0], size[1])
			pick(e.d, "opus")
			e.d.key("enter")
			e.d.mustShow("[Apply]")
			assertFits(t, e.d, size[0], size[1])
			e.d.key("n", "esc")
			toHost(t, e.d, "opencode")
			selectRole(t, e.d, "plain-role")
			e.d.key("enter", "right")
			e.d.key("pgdown", "pgdown")
			assertFits(t, e.d, size[0], size[1])
			e.d.key("end")
			assertFits(t, e.d, size[0], size[1])
		})
	}
}

// TestModelsViewCapturesTheGroupedScreensForHandoff prints the grouped table,
// the open model list and a group confirmation at 80x24 for a person to review.
func TestModelsViewCapturesThePickerForHandoff(t *testing.T) {
	if os.Getenv("HIVE_CAPTURE") == "" {
		t.Skip("set HIVE_CAPTURE=1 to print the captures")
	}
	f := standardFake()
	f.out["opencode"] = readCatalogString(t, "opencode") + "openai/gpt-4\nanthropic/claude-opus-4.7\nxai/grok-4.5\nsyn-oc/extra\n"
	e := pickerEnv(t, "claude,codex,opencode", f, 120, 40)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	fmt.Printf("EDIT PANEL (Model focused)\n%s\n", e.d.screen())
	e.d.key("right")
	fmt.Printf("CLAUDE PICKER (flat)\n%s\n", e.d.screen())
	e.d.key("esc", "esc")
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	fmt.Printf("OPENCODE PICKER (providers)\n%s\n", e.d.screen())
}

func readCatalogString(t *testing.T, name string) string { return string(readCatalogSample(t, name)) }

// --- table structure and the readable confirmation (correction round 1) -----------

// openModelsEnvWith opens the view over a source, with the theme choice.
func openModelsNoColor(t *testing.T, noColor bool) modelsEditEnv {
	t.Helper()
	e := modelsEditEnv{source: modelsTestSource(t), hosts: "claude,codex", deps: hostsTestDeps(coreOnlyAdapterFactory)}
	e.home, e.stateDir = newHostsTestHome(t)
	installViaText(t, e.home, e.stateDir, e.source, e.hosts, "y\n", e.deps)
	cfg := hostsAppConfig(t, e.home, e.stateDir, e.source, e.deps)
	cfg.NoColor = noColor
	e.m, e.d = newTestApp(t, cfg, 80, 24)
	e.d.send(pushViewMsg{v: newModelsView(e.m.cfg)})
	e.v = e.m.top().(*modelsView)
	e.d.screen()
	return e
}

func TestModelsViewNoColorTableKeepsMarkerIndentAndBlankRow(t *testing.T) {
	e := openModelsNoColor(t, true)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	if colorSGR(e.d.raw()) {
		t.Fatalf("NO_COLOR table emits a color sequence: %q", e.d.raw())
	}
	lines := e.d.lines()
	design, quality := -1, -1
	for i, l := range lines {
		switch {
		case strings.Contains(l, "▾ Design *"):
			design = i
		case strings.Contains(l, "▾ Quality"):
			quality = i
		}
	}
	if design < 0 || quality < 0 {
		t.Fatalf("headers lack the marker or the override star:\n%s", e.d.screen())
	}
	if strings.TrimSpace(lines[quality-1]) != "" {
		t.Fatalf("no blank row before the second group:\n%s", e.d.screen())
	}
	if strings.TrimSpace(lines[design-1]) == "" {
		t.Fatalf("a blank row sits before the first group:\n%s", e.d.screen())
	}
	if !strings.HasPrefix(lines[design+1], "    hive-design-architecture") {
		t.Fatalf("roles are not indented under their header: %q", lines[design+1])
	}
	// The blank row cannot be selected: from the last Design role, down lands on Quality.
	selectRole(t, e.d, "plain-role")
	e.d.key("down")
	if got := selectedLine(t, e.d); !strings.HasPrefix(got, "▾ Quality") {
		t.Fatalf("the cursor is on %q, want the Quality header", got)
	}
	e.d.key("up")
	if cursorRole(t, e.d) != "plain-role" {
		t.Fatal("up from the header did not skip the blank row")
	}
}

func TestModelsViewHeaderStyleDiffersFromRolesInColor(t *testing.T) {
	e := openModelsNoColor(t, false)
	e.d.key("down", "down") // the cursor leaves the header and the first roles
	var header, role string
	for _, l := range strings.Split(e.d.raw(), "\n") {
		switch {
		case strings.Contains(l, "Quality"):
			header = l
		case strings.Contains(l, "inherit-role"):
			role = l
		}
	}
	if !strings.Contains(header, "\x1b[") || strings.Contains(role, "\x1b[1") {
		t.Fatalf("header %q and role %q do not differ in style", header, role)
	}
	if !strings.Contains(header, "\x1b[1") && !strings.Contains(header, ";1") {
		t.Fatalf("the header is not bold: %q", header)
	}
}

// bulkSource adds seven roles to one group, each with an override already.
func bulkSource(t *testing.T) (string, []string) {
	t.Helper()
	dir := modelsTestSource(t)
	var roles []string
	for i := 1; i <= 7; i++ {
		name := fmt.Sprintf("bulk-role-%d", i)
		path := filepath.Join(dir, "content", "agents", "bulk", name+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(modelsRoleSource(name, "execution")), 0o600); err != nil {
			t.Fatal(err)
		}
		roles = append(roles, name)
	}
	return dir, roles
}

func TestModelsViewSevenRoleGroupConfirmationFitsAndScrolls(t *testing.T) {
	source, roles := bulkSource(t)
	e := newModelsEditEnvWith(t, source, "claude", 80, 24, nil)
	prior := map[string]management.ModelOverride{}
	for _, r := range roles {
		prior[r] = management.ModelOverride{Model: "sonnet", Effort: "low"}
	}
	store(t, e.home, e.stateDir, "claude", prior)
	e.d.key("r") // the Bulk header comes first alphabetically, so the cursor starts on it
	e.d.key("enter", "right")
	pick(e.d, "opus")
	e.d.key("enter")
	e.d.mustShow("Change the Bulk group on claude (7 roles)", "[Apply]")
	assertFits(t, e.d, 80, 24)
	e.d.mustShow("bulk-role-7  sonnet → opus", "Open sessions keep the previous model until they restart.")
}

// A group too big for the screen scrolls inside the confirmation.
func TestModelsViewBigGroupConfirmationScrolls(t *testing.T) {
	source, _ := syntheticCatalogueSource(t) // the generated group holds most of the catalogue
	e := newModelsEditEnvWith(t, source, "claude", 80, 24, nil)
	for range 30 {
		if strings.HasPrefix(selectedLine(t, e.d), "▾ Generated") {
			break
		}
		e.d.key("down")
	}
	e.d.key("enter", "right")
	pick(e.d, "opus")
	e.d.key("enter")
	e.d.mustShow("Change the Generated group on claude (", "[Apply]")
	assertFits(t, e.d, 80, 24)
	c, ok := e.m.top().(*confirmView)
	if !ok || !c.box.scrollable() {
		t.Fatalf("the confirmation of a big group does not scroll at 80x24:\n%s", e.d.screen())
	}
	e.d.mustNotShow("Open sessions keep")
	for range 6 {
		e.d.key("pgdown")
	}
	e.d.mustShow("Open sessions keep the previous model until they restart.")
	assertFits(t, e.d, 80, 24)
}

func TestModelsConfirmationReadsAsATableWithTheDirectoryOnce(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 120, 40)
	e.d.key("enter", "right")
	pick(e.d, "opus")
	e.d.key("enter")
	screen := e.d.screen()
	for _, want := range []string{
		"Change the Design group on claude (2 roles)",
		"Role                      Model                     Effort",
		"hive-design-architecture  syn-claude-reason → opus  high",
		"plain-role                syn-claude-exec → opus    medium",
		"Writes 2 files in ~/.claude/agents",
		"Open sessions keep the previous model until they restart.",
	} {
		if !strings.Contains(screen, want) {
			t.Errorf("confirmation lacks %q:\n%s", want, screen)
		}
	}
	if strings.Contains(screen, "File:") || strings.Contains(screen, "claude plain-role:") || strings.Contains(screen, "high → high") {
		t.Fatalf("the old per-role lines are still there:\n%s", screen)
	}
}

func TestAbbreviateHomeUsesTheHomeInUse(t *testing.T) {
	home := t.TempDir()
	real, _ := filepath.EvalSymlinks(home)
	for _, c := range []struct{ path, home, want string }{
		{filepath.Join(home, ".claude", "agents"), home, "~/.claude/agents"},
		{filepath.Join(real, ".claude", "agents"), home, "~/.claude/agents"},
		{real, home, "~"},
		{"/elsewhere/agents", home, "/elsewhere/agents"},
		{home + "-other/agents", home, home + "-other/agents"},
	} {
		if got := abbreviateHome(c.path, c.home); got != c.want {
			t.Errorf("abbreviateHome(%q) = %q, want %q", c.path, got, c.want)
		}
	}
}

// --- the boxed picker (D9-A) ----------------------------------------------------------

func TestModelsViewOpenCodeSectionsFollowTheProvidersWithoutRepeatingInUseModels(t *testing.T) {
	f := standardFake()
	f.out["opencode"] = "openai/gpt-5.5\nsyn-oc/exec\nanthropic/claude-sonnet-4.5\nopenai/gpt-4\nsolo\n"
	e := pickerEnv(t, "claude,opencode", f, 80, 24)
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	want := []string{
		"OpenAI", "gpt-5.5", "gpt-4",
		"syn-oc", "● exec",
		"Anthropic", "claude-sonnet-4.5",
		"opencode models", "solo",
		"release default", "Other…",
	}
	if got := listItems(t, e.d); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("sections:\n%q\nwant:\n%q\n%s", got, want, e.d.screen())
	}
	seen := map[string]int{}
	for _, l := range listItems(t, e.d) {
		f := strings.Fields(strings.TrimPrefix(l, "● "))
		seen[f[0]]++
	}
	if seen["exec"] != 1 {
		t.Fatalf("an in-use model is listed %d times", seen["exec"])
	}
	// What is displayed short is chosen and searched as the full id.
	e.v.panel.list.filter = "openai/gpt-4"
	if got := strings.Join(listItems(t, e.d), "|"); !strings.Contains(got, "OpenAI|gpt-4|") {
		t.Fatalf("a search by the full id finds %q", got)
	}
	e.v.panel.list.filter = ""
	for _, r := range e.v.pickerContent().rows {
		if r.id == "openai/gpt-5.5" {
			return
		}
	}
	t.Fatal("the selectable value is not the full id")
}

func TestModelsViewPiSectionsFollowTheProviders(t *testing.T) {
	f := standardFake()
	f.out["pi"] = "provider model context\nopenai gpt-4 8K\nxai grok-4.3 1M\nopenai gpt-4-turbo 128K\n"
	e := pickerEnv(t, "claude,pi", f, 80, 24)
	toHost(t, e.d, "pi")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	got := strings.Join(listItems(t, e.d), "|")
	for _, want := range []string{"OpenAI|gpt-4|gpt-4-turbo|xAI|grok-4.3"} {
		if !strings.Contains(got, want) {
			t.Fatalf("sections = %q, want %q", got, want)
		}
	}
}

func TestModelsViewCodexRowsShowTheNameAndThenTheID(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	got := strings.Join(listItems(t, e.d), "|")
	if !strings.Contains(got, "GPT-6.1 Sol gpt-6.1-sol|GPT-5.5 gpt-5.5") {
		t.Fatalf("names are not shown before ids: %q", got)
	}
}

func TestModelsViewSearchMatchesNameIDAndProvider(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	toHost(t, e.d, "codex")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	e.v.panel.catalog = []catalogModel{{ID: "m-1", Name: "Zebra Pro", Provider: "acme"}, {ID: "m-2", Provider: "other"}}
	for filter, want := range map[string]string{
		"zebra": "m-1", // the name
		"m-2":   "m-2", // the id
		"ACME":  "m-1", // the provider, whatever its case
	} {
		e.v.panel.list.filter = filter
		got := strings.Join(listItems(t, e.d), "|")
		if !strings.Contains(got, want) || (want == "m-1" && strings.Contains(got, "m-2")) {
			t.Errorf("search %q = %q", filter, got)
		}
	}
	e.v.panel.list.filter = "no-such-model"
	if got := strings.Join(listItems(t, e.d), "|"); got != "No matching models|release default|Other…" {
		t.Fatalf("a search with no match leaves %q", got)
	}
}

func TestModelsViewBottomEntriesAreReachableWithTheArrows(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	n := len(e.v.pickerContent().rows)
	for range n {
		e.d.key("down")
	}
	if got := pickerSelection(t, e); got != "release default" {
		t.Fatalf("past the last model the cursor is on %q", got)
	}
	e.d.key("right")
	if got := pickerSelection(t, e); got != "Other…" {
		t.Fatalf("→ reaches %q", got)
	}
	e.d.key("right", "down")
	if got := pickerSelection(t, e); got != "Other…" {
		t.Fatalf("→ and ↓ at the end move to %q", got)
	}
	e.d.key("left")
	if got := pickerSelection(t, e); got != "release default" {
		t.Fatalf("← reaches %q", got)
	}
	e.d.key("up")
	if got := pickerSelection(t, e); got != "haiku" {
		t.Fatalf("↑ from the entries reaches %q, want the last model", got)
	}
	e.d.key("home")
	if got := pickerSelection(t, e); got != "syn-claude-exec" {
		t.Fatalf("Home reaches %q", got)
	}
}

func TestModelsViewNoColorMarksTheSelectedRowAndEntryWithAnArrow(t *testing.T) {
	e := openModelsNoColor(t, true)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	if colorSGR(e.d.raw()) {
		t.Fatalf("NO_COLOR picker emits a color sequence: %q", e.d.raw())
	}
	count := func() (marked []string) {
		for _, l := range boxInterior(t, e.d) {
			if strings.HasPrefix(l, "> ") {
				marked = append(marked, l)
			}
		}
		return marked
	}
	if got := count(); len(got) != 1 || !strings.HasPrefix(got[0], "> ● syn-claude-exec") {
		t.Fatalf("marked rows = %q\n%s", got, e.d.screen())
	}
	e.d.key("end", "left")
	if got := count(); len(got) != 1 || !strings.HasPrefix(got[0], "> release default") {
		t.Fatalf("marked entries = %q\n%s", got, e.d.screen())
	}
}

func TestModelsViewPickerFitsWith209ModelsAndLongNames(t *testing.T) {
	longName := strings.Repeat("An Extremely Long Display Name ", 3) // capped at 60 by the catalog
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			f := standardFake()
			var oc, codex []string
			names := map[string][]string{}
			for i := 1; i <= 209; i++ {
				oc = append(oc, fmt.Sprintf("provider-with-a-long-name-%d/model-with-a-long-name-%03d", i%4, i))
				id := fmt.Sprintf("gpt-%03d", i)
				codex = append(codex, id)
				names[id] = allEfforts
				codexNames[id] = longName
			}
			t.Cleanup(func() {
				for _, id := range codex {
					delete(codexNames, id)
				}
			})
			f.out["opencode"] = strings.Join(oc, "\n") + "\n"
			f.out["codex"] = codexList(names, codex...)
			e := pickerEnv(t, "claude,codex,opencode", f, size[0], size[1])
			for _, host := range []string{"opencode", "codex"} {
				toHost(t, e.d, host)
				selectRole(t, e.d, "plain-role")
				e.d.key("enter", "right")
				assertFits(t, e.d, size[0], size[1])
				for range 5 {
					e.d.key("pgdown")
					assertFits(t, e.d, size[0], size[1])
				}
				e.d.key("end")
				assertFits(t, e.d, size[0], size[1])
				e.d.key("home")
				typeText(e.d, "model")
				assertFits(t, e.d, size[0], size[1])
				e.d.key("esc", "esc")
			}
		})
	}
}

// --- the borderless picker (correction round 3) ----------------------------------------

func TestModelsViewFlatListsHaveNoSectionHeaders(t *testing.T) {
	f := standardFake()
	f.out["cursor-agent"] = string(readCatalogSample(t, "cursor"))
	for host, want := range map[string]string{
		"claude": "● syn-claude-exec|fable|opus|sonnet|haiku|release default|Other…",
		"codex":  "● syn-codex-exec|GPT-6.1 Sol gpt-6.1-sol|GPT-5.5 gpt-5.5|release default|Other…",
		"grok":   "● grok-4.7|grok-4.6|release default|Other…",
		"cursor": "● Auto auto|Codex 5.3 Low gpt-5.3-codex-low|Codex 5.3 Low Fast gpt-5.3-codex-low-fast|Codex 5.3 gpt-5.3-codex|Codex 5.3 Fast gpt-5.3-codex-fast|Codex 5.3 High gpt-5.3-codex-high|Codex 5.3 High Fast gpt-5.3-codex-high-fast|release default|Other…",
	} {
		t.Run(host, func(t *testing.T) {
			e := pickerEnv(t, "claude,codex,cursor,grok", f, 120, 40)
			toHost(t, e.d, host)
			selectRole(t, e.d, "plain-role")
			// A host default (Grok, Cursor) has no model to mark; give the role one.
			if host == "grok" || host == "cursor" {
				model := map[string]string{"grok": "grok-4.7", "cursor": "auto"}[host]
				store(t, e.home, e.stateDir, host, map[string]management.ModelOverride{"plain-role": {Model: model}})
				e.d.key("r")
				selectRole(t, e.d, "plain-role")
			}
			e.d.key("enter", "right")
			if got := strings.Join(listItems(t, e.d), "|"); got != want {
				t.Fatalf("%s list:\n%q\nwant:\n%q\n%s", host, got, want, e.d.screen())
			}
		})
	}
}

func TestModelsViewProviderHeadersAreReadableWithAFallback(t *testing.T) {
	f := standardFake()
	f.out["opencode"] = "github-copilot/claude-opus-4.7\nxai/grok-4.5\nopencode-go/deepseek-v4\nweird-prov/model-1\namazon-bedrock/titan\n"
	e := pickerEnv(t, "claude,opencode", f, 120, 40)
	toHost(t, e.d, "opencode")
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	got := strings.Join(listItems(t, e.d), "|")
	for _, want := range []string{"GitHub Copilot|claude-opus-4.7", "xAI|grok-4.5", "OpenCode Go|deepseek-v4", "weird-prov|model-1", "Amazon Bedrock|titan"} {
		if !strings.Contains(got, want) {
			t.Errorf("sections %q lack %q", got, want)
		}
	}
	// A blank row separates sections, except before the first one.
	lines := boxInterior(t, e.d)
	for i, l := range lines {
		if l == "GitHub Copilot" && (i < 5 || lines[i-1] != "") {
			t.Errorf("no blank row before a later section: %q", lines[:i+1])
		}
	}
}

func TestModelsViewCursorStartsOnTheCurrentModelOrTheFirstOneWhenMixed(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter", "right")
	if got := pickerSelection(t, e); got != "syn-claude-exec" {
		t.Fatalf("the cursor starts on %q, want the current model", got)
	}
	e.d.key("esc", "esc", "up", "up", "up") // back to the Design header: its roles differ
	e.d.key("enter", "right")
	if got := pickerSelection(t, e); got != "syn-claude-reason" && got != "fable" {
		t.Fatalf("a mixed group starts on %q, want the first model", got)
	}
	for _, l := range listItems(t, e.d) {
		if strings.HasPrefix(l, "●") {
			t.Fatalf("a mixed group marks %q as current", l)
		}
	}
}

// The regression for the picker that ran past the screen's edge: every line of
// the screen fits, and the box stays two columns short of the view's width, so a
// terminal that draws an ambiguous-width character as two columns cannot wrap
// its border.
func TestModelsViewPickerBoxStaysTwoColumnsShortOfTheWidth(t *testing.T) {
	long := strings.Repeat("An Extremely Long Display Name ", 3)
	for _, size := range [][2]int{{80, 24}, {120, 40}, {200, 50}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			f := standardFake()
			var oc []string
			for i := 1; i <= 209; i++ {
				oc = append(oc, fmt.Sprintf("provider-with-a-long-name-%d/model-with-a-long-name-%03d", i%4, i))
			}
			f.out["opencode"] = strings.Join(oc, "\n") + "\n"
			codexNames["gpt-6.1-sol"] = long
			t.Cleanup(func() { codexNames["gpt-6.1-sol"] = "GPT-6.1 Sol" })
			e := pickerEnv(t, "claude,codex,opencode", f, size[0], size[1])
			for _, host := range []string{"claude", "codex", "opencode"} {
				toHost(t, e.d, host)
				selectRole(t, e.d, "plain-role")
				e.d.key("enter", "right")
				for step := 0; step < 4; step++ {
					assertFits(t, e.d, size[0], size[1])
					in, boxLines := false, 0
					for _, l := range e.d.lines() {
						if strings.HasPrefix(l, "┌") {
							in = true
						}
						if in {
							boxLines++
							if w := lipgloss.Width(l); w > size[0]-2 {
								t.Fatalf("a box line is %d wide, want at most %d: %q", w, size[0]-2, l)
							}
						}
						if strings.HasPrefix(l, "└") {
							in = false
						}
					}
					if boxLines < 8 {
						t.Fatalf("the box has %d lines:\n%s", boxLines, e.d.screen())
					}
					e.d.key("pgdown")
				}
				e.d.key("esc", "esc")
			}
		})
	}
}

var modelRowPattern = regexp.MustCompile(`^(> |  )Model +\S`)

func TestModelsViewModelFieldShowsTheChooseIndicatorAndHint(t *testing.T) {
	e := pickerEnv(t, "claude,codex", standardFake(), 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	raw := ""
	for _, l := range e.d.lines() {
		if modelRowPattern.MatchString(l) {
			raw = l
		}
	}
	if !strings.Contains(raw, "syn-claude-exec ▸") || !strings.Contains(raw, "→ choose") {
		t.Fatalf("Model row = %q, want the value, ▸ and the hint while focused", raw)
	}
	e.d.key("down")
	for _, l := range e.d.lines() {
		if modelRowPattern.MatchString(l) && (!strings.Contains(l, "▸") || strings.Contains(l, "→ choose")) {
			t.Fatalf("an unfocused Model row = %q: ▸ stays, the hint goes", l)
		}
	}
	lines := e.d.lines()
	help := lines[len(lines)-1]
	for _, want := range []string{"↑/↓ field", "→ choose model", "←/→ effort", "enter review", "esc close", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Errorf("help bar %q lacks %q", help, want)
		}
	}
	// A mixed group shows its state the same way.
	e.d.key("esc", "up", "up", "up", "up", "enter")
	if got := e.d.screen(); !strings.Contains(got, "mixed ▸") {
		t.Fatalf("a mixed group's Model row lacks \"mixed ▸\":\n%s", got)
	}
}
