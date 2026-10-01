package main

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the Models view's editing (#46, T3): the cursor, the panel, the
// confirmation and the reset, each compared with the command it stands for.

const modelsEditHosts = "claude,codex,grok,opencode,pi"

type modelsEditEnv struct {
	m                      *appModel
	d                      *appDriver
	v                      *modelsView
	home, stateDir, source string
	hosts                  string
	deps                   installDependencies
	runner                 catalogRunner // the fake model-list runner; nil leaves the real selection
}

// newModelsEditEnv installs source for hosts and opens the Models view.
func newModelsEditEnv(t *testing.T, source, hosts string, width, height int) modelsEditEnv {
	t.Helper()
	return newModelsEditEnvWith(t, source, hosts, width, height, nil)
}

// newModelsEditEnvWith is newModelsEditEnv with a fake model-list runner.
func newModelsEditEnvWith(t *testing.T, source, hosts string, width, height int, runner catalogRunner) modelsEditEnv {
	t.Helper()
	e := modelsEditEnv{source: source, hosts: hosts, deps: hostsTestDeps(coreOnlyAdapterFactory), runner: runner}
	e.home, e.stateDir = newHostsTestHome(t)
	installViaText(t, e.home, e.stateDir, source, hosts, "y\n", e.deps)
	return e.open(t, width, height)
}

func (e modelsEditEnv) open(t *testing.T, width, height int) modelsEditEnv {
	t.Helper()
	cfg := hostsAppConfig(t, e.home, e.stateDir, e.source, e.deps)
	cfg.catalogRunner = e.runner
	e.m, e.d = newTestApp(t, cfg, width, height)
	before := len(e.m.stack)
	e.d.send(pushViewMsg{v: newModelsView(e.m.cfg)}) // the app's cfg carries the shared catalog cache
	v, ok := e.m.top().(*modelsView)
	if !ok || len(e.m.stack) != before+1 {
		t.Fatalf("top view is %T", e.m.top())
	}
	e.v = v
	e.d.screen() // render once before typing, as openMenuEntry does
	return e
}

// twin is a second home with the same installation, for the command side.
func (e modelsEditEnv) twin(t *testing.T) (home, stateDir string) {
	t.Helper()
	home, stateDir = newHostsTestHome(t)
	installViaText(t, home, stateDir, e.source, e.hosts, "y\n", e.deps)
	return home, stateDir
}

// store applies overrides directly, as an earlier `hive models set` would have.
func store(t *testing.T, home, stateDir, host string, next map[string]management.ModelOverride) {
	t.Helper()
	o := management.Options{Home: home, StateDir: stateDir}
	p, err := management.BuildModelsPlan(o, host, next)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
}

// command runs `hive models <sub>` on a home as a person would, confirming.
func command(t *testing.T, home, stateDir, sub string, args ...string) {
	t.Helper()
	var out bytes.Buffer
	args = append(append([]string{}, args...), "--home", home, "--state-dir", stateDir)
	if err := modelsWrite(sub, args, strings.NewReader("y\n"), &out, true); err != nil {
		t.Fatalf("hive models %s %v: %v\n%s", sub, args, err, out.String())
	}
}

var (
	listRowPattern  = regexp.MustCompile(`^(> |  )(?:  )?([a-z][a-z0-9-]*)( \*)?  +\S`) // roles sit two columns under their group header
	panelRowPattern = regexp.MustCompile(`^(> |  )(Model|Effort|Role) +(.*)$`)
)

// cursorRole is the role on the line marked ">" in the table.
func cursorRole(t *testing.T, d *appDriver) string {
	t.Helper()
	found := ""
	for _, line := range d.lines() {
		if m := listRowPattern.FindStringSubmatch(line); m != nil && m[1] == "> " && m[2] != "Role" {
			if found != "" {
				t.Fatalf("two lines carry the cursor:\n%s", d.screen())
			}
			found = m[2]
		}
	}
	if found == "" {
		// The cursor can rest on a group header (T9); no role is selected then.
		for _, line := range d.lines() {
			if strings.HasPrefix(line, "> ") && !panelRowPattern.MatchString(line) {
				return ""
			}
		}
		t.Fatalf("no line carries the cursor:\n%s", d.screen())
	}
	return found
}

// selectRole moves the cursor down to role.
func selectRole(t *testing.T, d *appDriver, role string) {
	t.Helper()
	for range 60 {
		if cursorRole(t, d) == role {
			return
		}
		d.key("down")
	}
	t.Fatalf("cannot reach %s:\n%s", role, d.screen())
}

// panelField returns the text of a panel row and whether it holds the cursor.
func panelField(t *testing.T, d *appDriver, label string) (string, bool) {
	t.Helper()
	for _, line := range d.lines() {
		if m := panelRowPattern.FindStringSubmatch(line); m != nil && m[2] == label {
			text := strings.TrimSpace(m[3])
			if label == "Model" { // the value opens a list: "value ▸", and a hint on the focused row
				text = strings.TrimSpace(strings.TrimSuffix(text, "→ choose"))
				text = strings.TrimSpace(strings.TrimSuffix(text, "▸"))
			}
			return text, m[1] == "> "
		}
	}
	t.Fatalf("no %s row in the panel:\n%s", label, d.screen())
	return "", false
}

// mustNoPanel fails when the edit panel's Model or Effort row is on screen.
func mustNoPanel(t *testing.T, d *appDriver) {
	t.Helper()
	for _, line := range d.lines() {
		if m := panelRowPattern.FindStringSubmatch(line); m != nil && m[2] != "Role" {
			t.Fatalf("the edit panel is open:\n%s", d.screen())
		}
	}
}

// clearModelField turns Model into its text field, through the list's "Other…"
// entry, and empties the field. The panel's Model is a choice since T9, so every
// test that types a model first goes through this.
func clearModelField(d *appDriver) {
	d.t.Helper()
	d.key("right", "end", "enter")
	for range 60 {
		d.key("backspace")
	}
}

type snapshotFiles struct{ home, state map[string][]byte }

func snap(t *testing.T, e modelsEditEnv) snapshotFiles {
	t.Helper()
	return snapshotFiles{collectFiles(t, e.home), collectFiles(t, e.stateDir)}
}

func sameSnapshot(a, b snapshotFiles) bool {
	eq := func(x, y map[string][]byte) bool {
		if len(x) != len(y) {
			return false
		}
		for k, v := range x {
			if !bytes.Equal(y[k], v) {
				return false
			}
		}
		return true
	}
	return eq(a.home, b.home) && eq(a.state, b.state)
}

// --- the list ------------------------------------------------------------------

// groupedRoles orders the synthetic catalogue the way the grouped table lists
// it: by group (design, generated, quality), then by role.
func groupedRoles(roles []string) []string {
	group := func(r string) string {
		switch r {
		case longRoleName, "plain-role":
			return "design"
		case "inherit-role":
			return "quality"
		}
		return "generated"
	}
	out := append([]string(nil), roles...)
	sort.Slice(out, func(i, j int) bool {
		if gi, gj := group(out[i]), group(out[j]); gi != gj {
			return gi < gj
		}
		return out[i] < out[j]
	})
	return out
}

func TestModelsViewCursorMovesWithArrowsAndReachesTheLastRole(t *testing.T) {
	source, roles := syntheticCatalogueSource(t)
	roles = groupedRoles(roles)
	e := newModelsEditEnv(t, source, "claude", 80, 24)
	// The cursor starts on a group header (T9) and visits the headers too, so
	// the roles are reached in order with the headers skipped.
	if got := cursorRole(t, e.d); got != "" {
		t.Fatalf("the cursor starts on role %s, want a group header", got)
	}
	for i := 0; i < len(roles); i++ {
		for range 3 { // at most one header sits between two roles
			e.d.key("down")
			if cursorRole(t, e.d) != "" {
				break
			}
		}
		if got := cursorRole(t, e.d); got != roles[i] {
			t.Fatalf("the %dth role is %s, want %s", i+1, got, roles[i])
		}
	}
	assertFits(t, e.d, 80, 24)
	e.d.key("down")
	if got := cursorRole(t, e.d); got != roles[len(roles)-1] {
		t.Fatalf("the cursor ran past the last role: %s", got)
	}
	e.d.key("up", "up") // the Quality header sits between the last two roles
	if got := cursorRole(t, e.d); got != roles[len(roles)-2] {
		t.Fatalf("up from the last role reaches %s", got)
	}
}

func TestModelsViewPageKeysStillScrollAndTheCursorStaysInView(t *testing.T) {
	source, roles := syntheticCatalogueSource(t)
	roles = groupedRoles(roles)
	e := newModelsEditEnv(t, source, "claude", 80, 24)
	e.d.key("pgdown")
	e.d.mustNotShow(roles[0] + " ")
	if got := cursorRole(t, e.d); got == roles[0] {
		t.Fatalf("the cursor stayed on a row that scrolled away: %s", got)
	}
	e.d.key("end")
	e.d.mustShow(fmt.Sprintf("lines %d-%d of %d", len(roles)+5-14, len(roles)+5, len(roles)+5)) // 3 group headers and 2 separators
	if got := cursorRole(t, e.d); got != roles[len(roles)-1] {
		t.Fatalf("End leaves the cursor on %s", got)
	}
	e.d.key("home")
	if got := cursorRole(t, e.d); got != "" { // Home selects the first item, a header
		t.Fatalf("Home leaves the cursor on %s, want the first header", got)
	}
	e.d.key("down")
	if got := cursorRole(t, e.d); got != roles[0] {
		t.Fatalf("Home leaves the cursor on %s", got)
	}
}

func TestModelsViewMarksAnOverriddenRoleAndKeepsTheMarkNextToTheCursor(t *testing.T) {
	source := modelsTestSource(t)
	e := newModelsEditEnv(t, source, modelsEditHosts, 80, 24)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Effort: "max"}})
	e.d.key("r")
	e.d.mustShow("plain-role *")
	selectRole(t, e.d, "plain-role")
	line := ""
	for _, l := range e.d.lines() {
		if strings.HasPrefix(l, ">   plain-role *") {
			line = l
		}
	}
	if line == "" || !strings.Contains(line, "max") {
		t.Fatalf("the cursor line does not carry the mark and the override:\n%s", e.d.screen())
	}
	// The Design header carries the mark too, because one of its roles is overridden.
	e.d.mustShow("Design *")
	marked := 0
	for _, l := range e.d.lines() {
		if m := listRowPattern.FindStringSubmatch(l); m != nil && m[3] != "" {
			marked++
		}
	}
	if marked != 1 {
		t.Fatalf("%d role rows carry the mark, want 1:\n%s", marked, e.d.screen())
	}
}

func TestModelsViewFooterHasTheTwoLinesAndTheFixedRowsStaySix(t *testing.T) {
	source, roles := syntheticCatalogueSource(t)
	e := newModelsEditEnv(t, source, "claude", 80, 24)
	e.d.mustShow(
		"Enter edits the selected role or group; x resets one marked *.",
		"Defaults come from integrations/agent-profiles.json in the release.",
	)
	e.d.mustNotShow("To change a model or effort")
	// 24 rows less the chrome and the six fixed rows leave 15 table rows; the
	// three group headers and two separators are table rows too.
	e.d.mustShow(fmt.Sprintf("lines 1-15 of %d", len(roles)+5))
	assertFits(t, e.d, 80, 24)
}

func TestModelsViewHelpBarFitsEightyColumns(t *testing.T) {
	source := modelsTestSource(t)
	e := newModelsEditEnv(t, source, "claude", 80, 24)
	lines := e.d.lines()
	help := lines[len(lines)-1]
	for _, want := range []string{"↑/↓ role", "enter edit", "x reset", "←/→ CLI", "r reload", "esc back", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Errorf("help bar %q lacks %q", help, want)
		}
	}
	e.d.key("enter")
	lines = e.d.lines()
	help = lines[len(lines)-1]
	for _, want := range []string{"↑/↓ field", "←/→ effort", "enter review", "esc close", "ctrl+c quit"} {
		if !strings.Contains(help, want) {
			t.Errorf("panel help bar %q lacks %q", help, want)
		}
	}
}

// --- the panel -----------------------------------------------------------------

func TestModelsViewEnterOpensThePanelWithTheEffectiveModel(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if got, cursor := panelField(t, e.d, "Model"); got != "syn-claude-exec" || !cursor {
		t.Fatalf("Model = %q (cursor %v)\n%s", got, cursor, e.d.screen())
	}
	if got, _ := panelField(t, e.d, "Effort"); got != "< medium >" {
		t.Fatalf("Effort = %q\n%s", got, e.d.screen())
	}
	if got, _ := panelField(t, e.d, "Role"); !strings.HasPrefix(got, "plain-role") {
		t.Fatalf("Role = %q", got)
	}
	assertFits(t, e.d, 80, 24)

	// OpenCode shows the model without its #variant and the variant as the effort.
	e.d.key("esc", "right", "right", "right") // claude -> codex -> grok -> opencode
	if host := selectedModelsHost(t, e.d); host != "opencode" {
		t.Fatalf("host = %s", host)
	}
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	if got, _ := panelField(t, e.d, "Model"); got != "syn-oc/exec" {
		t.Fatalf("OpenCode Model = %q", got)
	}
	if got, _ := panelField(t, e.d, "Effort"); got != "< medium >" {
		t.Fatalf("OpenCode Effort = %q", got)
	}
}

func TestModelsViewPanelOnGrokSaysEffortIsNotSupported(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	e.d.key("right", "right")
	if host := selectedModelsHost(t, e.d); host != "grok" {
		t.Fatalf("host = %s", host)
	}
	e.d.key("enter")
	if got, _ := panelField(t, e.d, "Effort"); got != "not supported by grok" {
		t.Fatalf("Effort = %q\n%s", got, e.d.screen())
	}
	e.d.key("down")
	if _, cursor := panelField(t, e.d, "Model"); !cursor {
		t.Fatal("the cursor left Model for a field that cannot be edited")
	}
	if _, cursor := panelField(t, e.d, "Effort"); cursor {
		t.Fatal("the unsupported Effort field took the cursor")
	}
}

func TestModelsViewPanelKeysEscBackspaceAndTypedLetters(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	seq := e.v.seq
	e.d.key("enter")
	// Model is a choice until "Other…" makes it a text field (T9). There,
	// Backspace deletes text; r and x are typed, not commands.
	e.d.key("right", "end", "enter")
	e.d.key("backspace", "backspace", "r", "x")
	if got, _ := panelField(t, e.d, "Model"); got != "syn-claude-exrx" {
		t.Fatalf("Model = %q, want the typed letters in the field", got)
	}
	if e.v.seq != seq || e.m.top() != view(e.v) {
		t.Fatal("a letter typed into the field acted as a command")
	}
	// Backspace on Effort (not a text field) leaves the panel and the view open.
	e.d.key("down", "backspace")
	if e.m.top() != view(e.v) {
		t.Fatalf("Backspace on Effort left the view; top is %T", e.m.top())
	}
	if _, cursor := panelField(t, e.d, "Effort"); !cursor {
		t.Fatalf("the panel closed or lost the Effort cursor:\n%s", e.d.screen())
	}
	// Esc closes only the panel.
	e.d.key("esc")
	if e.m.top() != view(e.v) {
		t.Fatalf("Esc left the view; top is %T", e.m.top())
	}
	mustNoPanel(t, e.d)
	if cursorRole(t, e.d) != "plain-role" {
		t.Fatal("the cursor moved when the panel closed")
	}
	e.d.key("esc")
	if e.m.top() == view(e.v) {
		t.Fatal("Esc in the list did not leave the view")
	}
}

func TestModelsViewPanelEnterWithoutEditsSaysNothingToChange(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)
	e.d.key("enter", "enter")
	e.d.mustShow("Nothing to change")
	if e.m.top() != view(e.v) {
		t.Fatal("a confirmation opened for a change of nothing")
	}
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("files changed")
	}
}

func TestModelsViewConfirmationNamesTheChangeAndCancelChangesNothing(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)
	e.d.key("enter")
	clearModelField(e.d)
	typeText(e.d, "opus")
	e.d.key("down", "right") // medium -> high
	e.d.key("enter")
	e.d.mustShow(
		"Change plain-role on claude",
		"plain-role  syn-claude-exec → opus  medium → high",
		"Writes 1 file in ~/.claude/agents",
		"Open sessions keep the previous model until they restart.",
		"[Apply]",
	)
	if !strings.Contains(e.d.screen(), "Writes 1 file in ~/.claude/agents") {
		t.Fatalf("the confirmation does not name the directory with ~:\n%s", e.d.screen())
	}
	assertFits(t, e.d, 80, 24)
	e.d.key("n")
	e.d.mustShow("Cancelled. No changes applied.")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("Cancel changed files")
	}
	// The same through the Cancel button and Esc.
	e.d.key("enter")
	e.d.key("esc")
	e.d.mustShow("Cancelled. No changes applied.")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("Esc changed files")
	}
}

func TestModelsViewApplyEqualsTheMatchingCommand(t *testing.T) {
	type edit struct {
		name  string
		host  string // number of right presses from claude
		hops  int
		role  string
		prior map[string]management.ModelOverride // stored before editing, both sides
		keys  func(d *appDriver)
		cmd   func(t *testing.T, home, state string)
	}
	setCmd := func(host, role string, args ...string) func(*testing.T, string, string) {
		return func(t *testing.T, home, state string) {
			command(t, home, state, "set", append([]string{"--host", host, "--role", role}, args...)...)
		}
	}
	resetCmd := func(host, role string, args ...string) func(*testing.T, string, string) {
		return func(t *testing.T, home, state string) {
			command(t, home, state, "reset", append([]string{"--host", host, "--role", role}, args...)...)
		}
	}
	cases := []edit{
		{name: "a new model", host: "claude", role: "plain-role",
			keys: func(d *appDriver) { clearModelField(d); typeText(d, "opus") },
			cmd:  setCmd("claude", "plain-role", "--model", "opus")},
		{name: "only the effort", host: "claude", role: "plain-role",
			keys: func(d *appDriver) { d.key("down", "right") },
			cmd:  setCmd("claude", "plain-role", "--effort", "high")},
		{name: "an emptied model", host: "claude", role: "plain-role",
			prior: map[string]management.ModelOverride{"plain-role": {Model: "opus"}},
			keys:  func(d *appDriver) { clearModelField(d) },
			cmd:   resetCmd("claude", "plain-role", "--only", "model")},
		{name: "the effort back to the release default", host: "claude", role: "plain-role",
			prior: map[string]management.ModelOverride{"plain-role": {Model: "opus", Effort: "max"}},
			keys:  func(d *appDriver) { d.key("down", "left", "left", "left", "left", "left") }, // max -> release default
			cmd:   resetCmd("claude", "plain-role", "--only", "effort")},
		{name: "a model on OpenCode keeps the shown effort", host: "opencode", hops: 3, role: "plain-role",
			keys: func(d *appDriver) { clearModelField(d); typeText(d, "x/y") },
			cmd:  setCmd("opencode", "plain-role", "--model", "x/y", "--effort", "medium")},
		{name: "a model on an OpenCode inherit role keeps #max as its effort", host: "opencode", hops: 3, role: "inherit-role",
			keys: func(d *appDriver) { clearModelField(d); typeText(d, "x/y") },
			cmd:  setCmd("opencode", "inherit-role", "--model", "x/y", "--effort", "max")},
		{name: "both on Pi", host: "pi", hops: 4, role: "plain-role",
			keys: func(d *appDriver) { clearModelField(d); typeText(d, "p/m"); d.key("down", "right", "right") },
			cmd:  setCmd("pi", "plain-role", "--model", "p/m", "--effort", "xhigh")},
		{name: "a model on Grok", host: "grok", hops: 2, role: "plain-role",
			keys: func(d *appDriver) { clearModelField(d); typeText(d, "g-1") },
			cmd:  setCmd("grok", "plain-role", "--model", "g-1")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
			twinHome, twinState := e.twin(t)
			if c.prior != nil {
				store(t, e.home, e.stateDir, c.host, c.prior)
				store(t, twinHome, twinState, c.host, c.prior)
				e.d.key("r")
			}
			for range c.hops {
				e.d.key("right")
			}
			selectRole(t, e.d, c.role)
			e.d.key("enter")
			c.keys(e.d)
			e.d.key("enter")
			e.d.mustShow("[Apply]")
			e.d.key("enter")
			e.d.mustShow("Open sessions keep the previous model until they restart")
			c.cmd(t, twinHome, twinState)
			assertTwin(t, e.home, e.stateDir, twinHome, twinState)
			// The panel closed and the cursor stayed on the role.
			mustNoPanel(t, e.d)
			if got := cursorRole(t, e.d); got != c.role {
				t.Fatalf("the cursor is on %s after applying, want %s", got, c.role)
			}
			if e.m.top() != view(e.v) {
				t.Fatalf("top view is %T", e.m.top())
			}
		})
	}
}

func TestModelsViewXResetsAnOverriddenRoleLikeTheCommand(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	twinHome, twinState := e.twin(t)
	prior := map[string]management.ModelOverride{"plain-role": {Model: "opus", Effort: "max"}, "inherit-role": {Effort: "high"}}
	store(t, e.home, e.stateDir, "claude", prior)
	store(t, twinHome, twinState, "claude", prior)
	e.d.key("r")
	selectRole(t, e.d, "hive-design-architecture")
	before := snap(t, e)
	e.d.key("x") // not overridden: nothing happens
	if e.m.top() != view(e.v) || !sameSnapshot(before, snap(t, e)) {
		t.Fatal("x on a role without an override did something")
	}
	selectRole(t, e.d, "plain-role")
	e.d.key("x")
	e.d.mustShow("Reset plain-role on claude", "plain-role  opus → syn-claude-exec  max → medium", "[Apply]")
	e.d.key("n")
	e.d.mustShow("Cancelled. No changes applied.")
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("Cancel changed files")
	}
	e.d.key("x", "y")
	e.d.mustShow("Open sessions keep the previous model until they restart")
	command(t, twinHome, twinState, "reset", "--host", "claude", "--role", "plain-role")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
	if got := cursorRole(t, e.d); got != "plain-role" {
		t.Fatalf("the cursor is on %s after the reset", got)
	}
	e.d.mustNotShow("plain-role *")
	e.d.mustShow("inherit-role *") // the other override stays
}

func TestModelsViewValidationErrorTakesOneRow(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, size[0], size[1])
			selectRole(t, e.d, "plain-role")
			before := snap(t, e)
			e.d.key("enter")
			clearModelField(e.d)
			typeText(e.d, strings.Repeat("a", 100)+" b")
			e.d.key("enter")
			e.d.mustShow("model may only contain")
			if size[0] == 80 && !strings.Contains(e.d.screen(), "…") {
				t.Fatalf("the long error is not cut with an ellipsis:\n%s", e.d.screen())
			}
			rows := 0
			for _, l := range e.d.lines() {
				if strings.Contains(l, "model") && strings.Contains(l, "override") {
					rows++
				}
			}
			if rows != 1 {
				t.Fatalf("the error spans %d rows:\n%s", rows, e.d.screen())
			}
			assertFits(t, e.d, size[0], size[1])
			if !sameSnapshot(before, snap(t, e)) {
				t.Fatal("a rejected change wrote something")
			}
			// The panel stays open with what was typed.
			if _, cursor := panelField(t, e.d, "Model"); !cursor {
				t.Fatal("the panel closed")
			}
		})
	}
}

func TestModelsViewFitsEveryScreenAtBothSizes(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			source, _ := syntheticCatalogueSource(t)
			e := newModelsEditEnv(t, source, "claude,opencode", size[0], size[1])
			store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Model: "opus", Effort: "max"}})
			e.d.key("r")
			assertFits(t, e.d, size[0], size[1])
			selectRole(t, e.d, "plain-role")
			e.d.key("enter")
			assertFits(t, e.d, size[0], size[1])
			e.d.key("down", "left")
			e.d.key("enter")
			e.d.mustShow("[Apply]")
			assertFits(t, e.d, size[0], size[1])
			e.d.key("n", "esc")
			selectRole(t, e.d, "plain-role")
			e.d.key("x")
			assertFits(t, e.d, size[0], size[1])
		})
	}
}

// TestModelsViewCapturesForHandoff prints the list, the panel and the
// confirmation at 80x24 for a person to review the composition.
func TestModelsViewCapturesForHandoff(t *testing.T) {
	source, _ := syntheticCatalogueSource(t)
	e := newModelsEditEnv(t, source, "claude", 80, 24)
	store(t, e.home, e.stateDir, "claude", map[string]management.ModelOverride{"plain-role": {Model: "opus", Effort: "max"}})
	e.d.key("r")
	t.Logf("LIST, first render\n%s", e.d.screen())
	selectRole(t, e.d, "plain-role")
	t.Logf("LIST, cursor on the overridden role\n%s", e.d.screen())
	e.d.key("enter")
	t.Logf("PANEL\n%s", e.d.screen())
	e.d.key("down", "left", "left")
	e.d.key("enter")
	t.Logf("CONFIRMATION\n%s", e.d.screen())
}

// TestModelsViewFailedApplyShowsTheErrorOrOpensRecovery checks the two ends of a
// failed write: with nothing pending the error stays in the panel, and with an
// operation pending the recovery view opens, as in the Voice view.
func TestModelsViewFailedApplyShowsTheErrorOrOpensRecovery(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	v := e.v
	cmd, act := v.onApplied(modelsAppliedMsg{err: fmt.Errorf("the disk is full")})
	if cmd != nil || act.nav != navPop {
		t.Fatalf("a plain failure answered %+v", act)
	}
	e.d.mustShow("the disk is full")
	_, act = v.onApplied(modelsAppliedMsg{err: fmt.Errorf("interrupted"), pending: management.PendingCore})
	if act.nav != navPush || act.pops != 1 {
		t.Fatalf("a pending operation answered %+v", act)
	}
	if _, ok := act.push.(*recoveryView); !ok {
		t.Fatalf("pushed %T, want the recovery view", act.push)
	}
}

func TestModelsViewUpdatedMessageDependsOnWhetherAFileChanged(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	_, _ = e.v.onApplied(modelsAppliedMsg{filesChanged: true})
	if !strings.Contains(e.v.message, "Open sessions keep the previous model until they restart.") {
		t.Fatalf("message = %q", e.v.message)
	}
	_, _ = e.v.onApplied(modelsAppliedMsg{filesChanged: false})
	if strings.Contains(e.v.message, "Open sessions") || !strings.Contains(e.v.message, "no agent file changed") {
		t.Fatalf("message = %q", e.v.message)
	}
}

// A request that is not empty but whose plan changes nothing says so, opens no
// confirmation and writes nothing.
func TestModelsViewNonEmptyRequestWithNoEffectSaysNothingToChange(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	selectRole(t, e.d, "plain-role")
	before := snap(t, e)
	// The effort goes to "release default" while no override holds one: the
	// panel asks to drop a part that is not stored.
	e.d.key("enter", "down", "left", "left")
	if got, _ := panelField(t, e.d, "Effort"); got != "< release default >" {
		t.Fatalf("Effort = %q, want the release default", got)
	}
	e.d.key("enter")
	e.d.mustShow("Nothing to change")
	if e.m.top() != view(e.v) {
		t.Fatalf("a confirmation opened for a plan that changes nothing; top is %T", e.m.top())
	}
	if !sameSnapshot(before, snap(t, e)) {
		t.Fatal("files changed")
	}
}

func TestModelsViewTwoHundredCharacterValidationErrorStaysOneRow(t *testing.T) {
	model := strings.Repeat("a", 100) + " " + strings.Repeat("b", 99) // 200 characters, one of them a space
	if len(model) != 200 {
		t.Fatalf("test setup: %d characters", len(model))
	}
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, size[0], size[1])
			selectRole(t, e.d, "plain-role")
			before := snap(t, e)
			e.d.key("enter")
			clearModelField(e.d)
			typeText(e.d, model)
			e.d.key("enter")
			e.d.mustShow("model may only contain")
			rows := 0
			for _, l := range e.d.lines() {
				if strings.Contains(l, "may only contain") {
					rows++
				}
			}
			if rows != 1 {
				t.Fatalf("the error spans %d rows:\n%s", rows, e.d.screen())
			}
			assertFits(t, e.d, size[0], size[1])
			if !sameSnapshot(before, snap(t, e)) {
				t.Fatal("a rejected change wrote something")
			}
		})
	}
}

func TestModelsViewTrimsSpacesAroundATypedModel(t *testing.T) {
	e := newModelsEditEnv(t, modelsTestSource(t), modelsEditHosts, 80, 24)
	twinHome, twinState := e.twin(t)
	selectRole(t, e.d, "plain-role")
	e.d.key("enter")
	clearModelField(e.d)
	typeText(e.d, "  opus  ")
	e.d.key("enter")
	e.d.mustShow("syn-claude-exec → opus", "[Apply]")
	e.d.key("enter")
	e.d.mustShow("Open sessions keep the previous model until they restart")
	command(t, twinHome, twinState, "set", "--host", "claude", "--role", "plain-role", "--model", "opus")
	assertTwin(t, e.home, e.stateDir, twinHome, twinState)
}
