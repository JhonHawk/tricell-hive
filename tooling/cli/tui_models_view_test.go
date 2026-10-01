package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the Models view (#46, T1).

const (
	longRoleName  = "hive-design-architecture" // the longest catalogue role, 24 columns
	modelsProfile = "integrations/agent-profiles.json"
	// modelsProfileFixture is the synthetic profiles file the view tests read;
	// modelsTestSource installs it at modelsProfile in the temporary source.
	modelsProfileFixture = "testdata/agent-profiles.json"
)

func modelsRoleSource(name, profile string) string {
	return modelsRoleSourceWithAccess(name, profile, "observe")
}

// modelsRoleSourceWithAccess is modelsRoleSource with an explicit access profile.
func modelsRoleSourceWithAccess(name, profile, access string) string {
	return "---\nname: " + name + "\ndescription: Test role\nmodel_profile: " + profile + "\naccess_profile: " + access + "\n---\nUse evidence.\n"
}

// modelsTestSource builds a Hive source with the synthetic agent
// profiles (testdata/agent-profiles.json) and three roles: a reasoning one with the longest role name, an
// execution one, and one that inherits its parent session's model.
func modelsTestSource(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	write := func(rel, text string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	profiles, err := os.ReadFile(filepath.FromSlash(modelsProfileFixture))
	if err != nil {
		t.Fatal(err)
	}
	write("content/guidance/global.md", "# Global\n\nMinimal test guidance.\n")
	write(modelsProfile, string(profiles))
	write("content/agents/design/"+longRoleName+".md", modelsRoleSource(longRoleName, "reasoning"))
	write("content/agents/design/plain-role.md", modelsRoleSource("plain-role", "execution"))
	write("content/agents/quality/inherit-role.md", modelsRoleSource("inherit-role", "inherit"))
	return dir
}

// openModelsView opens the application and pushes the Models view over a home
// with the given CLIs installed (none when hosts is empty).
func openModelsView(t *testing.T, hosts string, width, height int) (*appModel, *appDriver, *modelsView, string, string) {
	t.Helper()
	source := modelsTestSource(t)
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	if hosts != "" {
		installViaText(t, home, stateDir, source, hosts, "y\n", deps)
	}
	cfg := hostsAppConfig(t, home, stateDir, source, deps)
	m, d := newTestApp(t, cfg, width, height)
	before := len(m.stack)
	d.send(pushViewMsg{v: newModelsView(m.cfg)})
	v, ok := m.top().(*modelsView)
	if !ok || len(m.stack) != before+1 {
		t.Fatalf("top view is %T", m.top())
	}
	return m, d, v, home, stateDir
}

var hostRowBrackets = regexp.MustCompile(`\[([a-z]+)\]`)

// selectedModelsHost is the CLI shown in brackets on the screen.
func selectedModelsHost(t *testing.T, d *appDriver) string {
	t.Helper()
	for _, line := range d.lines() {
		if m := hostRowBrackets.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	}
	t.Fatalf("no bracketed CLI on the screen:\n%s", d.screen())
	return ""
}

// modelsRowFor returns the columns of the table line that starts with role.
func modelsRowFor(t *testing.T, d *appDriver, role string) []string {
	t.Helper()
	for _, line := range d.lines() {
		// Every table line starts with the two columns of the cursor mark.
		line = strings.TrimPrefix(strings.TrimPrefix(line, "> "), "  ")
		if strings.HasPrefix(line, role+" ") || line == role {
			return regexp.MustCompile(`\s{2,}`).Split(strings.TrimSpace(line), -1)
		}
	}
	t.Fatalf("no table line for %s:\n%s", role, d.screen())
	return nil
}

func TestModelsViewSwitchesCLIWithArrowsAndBracketsTheSelectedOne(t *testing.T) {
	_, d, _, _, _ := openModelsView(t, "claude,codex,grok,opencode", 80, 24)
	d.mustShow("Models")
	if got := selectedModelsHost(t, d); got != "claude" {
		t.Fatalf("selected = %s, want claude first", got)
	}
	d.mustShow("[claude]", "codex", "grok", "opencode")
	d.key("right")
	if got := selectedModelsHost(t, d); got != "codex" {
		t.Fatalf("after right: %s", got)
	}
	d.mustShow("syn-codex-exec")
	d.mustNotShow("syn-claude-exec")
	d.key("left")
	if got := selectedModelsHost(t, d); got != "claude" {
		t.Fatalf("after left: %s", got)
	}
	d.mustShow("syn-claude-exec")
	d.key("left") // the first CLI stays selected
	if got := selectedModelsHost(t, d); got != "claude" {
		t.Fatalf("left at the first CLI moved to %s", got)
	}
	d.key("right", "right", "right", "right")
	if got := selectedModelsHost(t, d); got != "opencode" {
		t.Fatalf("right past the last CLI selected %s", got)
	}
}

func TestModelsViewShowsHostDefaultAndInheritInPlainWords(t *testing.T) {
	_, d, _, _, _ := openModelsView(t, "claude,grok", 80, 24)
	if row := modelsRowFor(t, d, "inherit-role"); len(row) < 3 || row[2] != "inherit (parent session)" {
		t.Fatalf("Claude inherit row = %q\n%s", row, d.screen())
	}
	d.key("right")
	if got := selectedModelsHost(t, d); got != "grok" {
		t.Fatalf("selected %s", got)
	}
	if row := modelsRowFor(t, d, "plain-role"); len(row) < 3 || row[2] != "host default" {
		t.Fatalf("Grok row = %q\n%s", row, d.screen())
	}
	d.mustNotShow("inherit (parent session)")
}

func TestModelsViewSplitsTheOpenCodeVariantIntoTheEffortColumn(t *testing.T) {
	_, d, _, _, _ := openModelsView(t, "claude,opencode", 80, 24)
	d.key("right")
	row := modelsRowFor(t, d, "plain-role")
	if len(row) != 4 || row[2] != "syn-oc/exec" || row[3] != "medium" {
		t.Fatalf("OpenCode execution row = %q, want model without #medium and effort medium\n%s", row, d.screen())
	}
	d.mustNotShow("#medium")
	row = modelsRowFor(t, d, longRoleName)
	if len(row) != 4 || row[2] != "syn-oc/reason" || row[3] != "high" {
		t.Fatalf("OpenCode reasoning row = %q", row)
	}
}

func TestModelsViewShowsRoleEffortAndProfileColumns(t *testing.T) {
	_, d, _, _, _ := openModelsView(t, "claude", 80, 24)
	if row := modelsRowFor(t, d, "plain-role"); len(row) != 4 || row[1] != "execution" || row[2] != "syn-claude-exec" || row[3] != "medium" {
		t.Fatalf("Claude plain row = %q", row)
	}
	d.mustShow("Role", "Profile", "Model", "Effort")
}

func TestModelsViewFooterSaysHowToEditAndWhereDefaultsComeFrom(t *testing.T) {
	_, d, _, _, _ := openModelsView(t, "claude", 80, 24)
	d.mustShow("Enter edits the selected role or group; x resets one marked *.", "Defaults come from integrations/agent-profiles.json in the release.")
	lines := d.lines()
	// The help bar is the last line; the footer is the two lines above it.
	if !strings.Contains(strings.Join(lines[len(lines)-3:len(lines)-1], "\n"), "Defaults come from") {
		t.Fatalf("the footer is not the two lines above the help bar:\n%s", d.screen())
	}
}

func TestModelsViewWithoutRegisteredCLIs(t *testing.T) {
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		_, d, _, _, _ := openModelsView(t, "", size[0], size[1])
		d.mustShow("No CLI hosts are registered.", "Open CLIs from the menu", "hive install")
		d.mustNotShow("Profile", "To change a model or effort", "hive update")
		assertFits(t, d, size[0], size[1])
	}
}

func TestModelsViewCLIWithoutAgents(t *testing.T) {
	_, d, v, _, _ := openModelsView(t, "claude,codex", 80, 24)
	d.send(modelsLoadedMsg{owned: owned{v}, seq: v.seq, hosts: []string{"claude", "codex"}, rows: []management.ModelRow{
		{Host: "claude", Role: "solo", Profile: "execution", Model: "sonnet", Effort: "high"},
	}})
	d.key("right")
	d.mustShow("[codex]", "No agents installed for codex")
	assertFits(t, d, 80, 24)
}

func TestModelsViewColumnsAt80And120Columns(t *testing.T) {
	long := "provider/a-long-model-name-for-the-wide-layout-12345" // 50 runes: cut at 80 columns, whole at 120
	rows := []management.ModelRow{
		{Host: "pi", Role: longRoleName, Profile: "reasoning", Model: long, Effort: "medium"},
		{Host: "pi", Role: "short", Profile: "inherit", Model: "inherit", Effort: "high"},
	}
	for _, size := range []struct{ w, h int }{{80, 24}, {120, 30}} {
		_, d, v, _, _ := openModelsView(t, "pi", size.w, size.h)
		d.send(modelsLoadedMsg{owned: owned{v}, seq: v.seq, hosts: []string{"pi"}, rows: rows})
		assertFits(t, d, size.w, size.h)
		var header, first string
		for _, l := range d.lines() {
			if strings.HasPrefix(l, "  Role") {
				header = l
			}
			if strings.HasPrefix(l, "> "+longRoleName) || strings.HasPrefix(l, "  "+longRoleName) {
				first = l
			}
		}
		if header == "" || first == "" {
			t.Fatalf("%dx%d: table not found:\n%s", size.w, size.h, d.screen())
		}
		// Role is as wide as the longest role, Profile follows at +2, and so on.
		// The table starts after the two columns of the cursor mark.
		if i := strings.Index(header, "Profile"); i != 28 {
			t.Errorf("%dx%d: Profile starts at column %d, want 28", size.w, size.h, i)
		}
		if i := strings.Index(header, "Model"); i != 28+10+2 {
			t.Errorf("%dx%d: Model starts at column %d, want 40", size.w, size.h, i)
		}
		if size.w == 80 {
			if !strings.Contains(first, "…") || strings.Contains(first, long) {
				t.Errorf("80 columns: the long model is not cut with an ellipsis: %q", first)
			}
			if got := len([]rune(first)); got > 80 {
				t.Errorf("row is %d runes: %q", got, first)
			}
			modelCol := []rune(first)[40:]
			cut := strings.SplitN(string(modelCol), "  ", 2)[0]
			if n := len([]rune(cut)); n < 28 {
				t.Errorf("Model column shows %d runes, want at least 28: %q", n, cut)
			}
		} else if !strings.Contains(first, long) {
			t.Errorf("120 columns: the model should show whole: %q", first)
		}
		if !strings.HasPrefix(first[2:], longRoleName+"  reasoning") {
			t.Errorf("the role was cut or misaligned: %q", first)
		}
	}
}

func TestModelsViewScrollsWhenRowsOutgrowTheScreen(t *testing.T) {
	var rows []management.ModelRow
	for i := 0; i < 40; i++ {
		rows = append(rows, management.ModelRow{Host: "claude", Role: "role-" + string(rune('a'+i/26)) + string(rune('a'+i%26)), Profile: "execution", Model: "sonnet", Effort: "high"})
	}
	_, d, v, _, _ := openModelsView(t, "claude", 80, 24)
	d.send(modelsLoadedMsg{owned: owned{v}, seq: v.seq, hosts: []string{"claude"}, rows: rows})
	assertFits(t, d, 80, 24)
	d.mustShow("role-aa", "lines 1-15 of 40")
	d.key("pgdown")
	d.mustNotShow("role-aa ")
	d.key("end")
	d.mustShow("lines 26-40 of 40")
	assertFits(t, d, 80, 24)
}

func TestModelsViewLoadErrorOffersRetryAndRReloads(t *testing.T) {
	_, d, v, _, _ := openModelsView(t, "claude", 80, 24)
	d.send(modelsLoadedMsg{owned: owned{v}, seq: v.seq, err: errors.New("state is unreadable")})
	d.mustShow("Hive's state in", "The models cannot be shown", "r to retry after fixing it", "Detail: state is unreadable")
	// The long state path breaks after a slash, so the words after it may wrap anywhere.
	mustShowFlat(d, "could not be read. Repair or restore its files; hive doctor shows the same problem.")
	d.mustNotShow("Cannot read the models")
	lines := d.lines()
	plain, detail := -1, -1
	for i, l := range lines {
		if plain < 0 && strings.Contains(l, "Repair or restore its files") {
			plain = i
		}
		if detail < 0 && strings.Contains(l, "state is unreadable") {
			detail = i
		}
	}
	if plain < 0 || detail <= plain {
		t.Fatalf("the technical detail must come after the plain words (plain %d, detail %d):\n%s", plain, detail, d.screen())
	}
	assertFits(t, d, 80, 24)
	seq := v.seq
	d.key("r")
	if v.seq != seq+1 {
		t.Fatalf("r did not start a reload (seq %d -> %d)", seq, v.seq)
	}
	d.mustShow("[claude]", "syn-claude-exec")
	d.mustNotShow("could not be read")
}

func TestModelsViewIgnoresAStaleResultAndKeepsTheSelectedCLI(t *testing.T) {
	_, d, v, _, _ := openModelsView(t, "claude,codex", 80, 24)
	d.key("right")
	stale := v.seq - 1
	d.send(modelsLoadedMsg{owned: owned{v}, seq: stale, err: errors.New("late")})
	d.mustNotShow("late")
	d.key("r")
	if got := selectedModelsHost(t, d); got != "codex" {
		t.Fatalf("a reload moved the selection to %s", got)
	}
}

func TestModelsViewEscAndBackspaceReturnToTheMenu(t *testing.T) {
	for _, k := range []string{"esc", "backspace"} {
		m, d, _, _, _ := openModelsView(t, "claude", 80, 24)
		d.key(k)
		if _, ok := m.top().(*menuView); !ok {
			t.Fatalf("%s left %T on top", k, m.top())
		}
	}
}

func TestModelsViewIsReadOnly(t *testing.T) {
	_, d, _, home, stateDir := openModelsView(t, "claude,codex,grok,opencode", 80, 24)
	beforeHome := collectFiles(t, home)
	beforeState := collectFiles(t, stateDir)
	stateJSON, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	d.key("right", "right", "r", "left")
	for path, want := range beforeHome {
		if got, ok := collectFiles(t, home)[path]; !ok || !bytes.Equal(got, want) {
			t.Errorf("home file %s changed", path)
		}
	}
	afterHome, afterState := collectFiles(t, home), collectFiles(t, stateDir)
	if len(afterHome) != len(beforeHome) || len(afterState) != len(beforeState) {
		t.Fatalf("files appeared: home %d -> %d, state %d -> %d", len(beforeHome), len(afterHome), len(beforeState), len(afterState))
	}
	for path, want := range beforeState {
		if !bytes.Equal(afterState[path], want) {
			t.Errorf("state file %s changed", path)
		}
	}
	after, err := os.ReadFile(filepath.Join(stateDir, "state.json"))
	if err != nil || !bytes.Equal(after, stateJSON) {
		t.Fatalf("state.json changed: %v", err)
	}
}

func TestRenderModelsTextSharesTheViewWordsAndCutsNothing(t *testing.T) {
	long := "provider/a-very-long-model-name-that-does-not-fit-in-the-column-1234567890"
	rows := []management.ModelRow{
		{Host: "claude", Role: "b-role", Profile: "inherit", Model: "inherit", Effort: "high"},
		{Host: "grok", Role: "a-role", Profile: "execution"},
		{Host: "opencode", Role: "a-role", Profile: "execution", Model: "x/y#max"},
		{Host: "pi", Role: "a-role", Profile: "reasoning", Model: long, Effort: "medium"},
	}
	var out bytes.Buffer
	renderModelsText([]string{"claude", "grok", "opencode", "pi", "cursor"}, rows, &out)
	text := out.String()
	for _, want := range []string{"claude\n", "grok\n", "opencode\n", "pi\n", "inherit (parent session)", "host default", long, "cursor\n  No agents installed for cursor\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "#max") || !regexp.MustCompile(`x/y\s+max`).MatchString(text) {
		t.Errorf("the OpenCode variant is not split into Effort:\n%s", text)
	}
	if strings.Contains(text, "…") {
		t.Errorf("the text output cut something:\n%s", text)
	}
	for _, l := range strings.Split(text, "\n") {
		if l != strings.TrimRight(l, " ") {
			t.Errorf("trailing spaces: %q", l)
		}
	}
	out.Reset()
	renderModelsText(nil, nil, &out)
	if out.String() != "No CLI hosts are registered. Open CLIs from the menu, or run hive install, to install Hive.\n" {
		t.Errorf("empty output = %q", out.String())
	}
}

// syntheticRoleCount is the number of roles syntheticCatalogueSource builds:
// enough that the Models view scrolls at 80x24, where it shows
// 24 - modelsFixedRows = 18 rows.
const syntheticRoleCount = 22

// syntheticCatalogueSource builds a Hive source with the repository's real
// agent profiles and syntheticRoleCount synthetic roles: modelsTestSource's
// three (including the longest real role name) plus generated ones that mix
// every model profile kind of the real catalogue (reasoning, execution,
// inherit, verifier) with every access profile (observe, implement, verify).
// It reads no real role.
func syntheticCatalogueSource(t *testing.T) (string, []string) {
	t.Helper()
	dir := modelsTestSource(t)
	roles := []string{longRoleName, "plain-role", "inherit-role"}
	profiles := []string{"reasoning", "execution", "inherit", "verifier"}
	accesses := []string{"observe", "implement", "verify"}
	for i := len(roles); i < syntheticRoleCount; i++ {
		name := fmt.Sprintf("hive-role-%02d", i)
		path := filepath.Join(dir, "content", "agents", "generated", name+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(modelsRoleSourceWithAccess(name, profiles[i%len(profiles)], accesses[i%len(accesses)])), 0600); err != nil {
			t.Fatal(err)
		}
		roles = append(roles, name)
	}
	return dir, roles
}

func TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs(t *testing.T) {
	source, roles := syntheticCatalogueSource(t)
	if len(roles) != syntheticRoleCount {
		t.Fatalf("the synthetic catalogue has %d roles, want %d", len(roles), syntheticRoleCount)
	}
	roles = groupedRoles(roles) // the table lists roles by group, then by name
	first, last := roles[0], roles[len(roles)-1]
	deps := hostsTestDeps(coreOnlyAdapterFactory)
	home, stateDir := newHostsTestHome(t)
	installViaText(t, home, stateDir, source, strings.Join(installerHosts, ","), "y\n", deps)
	cfg := hostsAppConfig(t, home, stateDir, source, deps)
	for _, size := range [][2]int{{80, 24}, {120, 40}} {
		t.Run(fmt.Sprintf("%dx%d", size[0], size[1]), func(t *testing.T) {
			m, d := newTestApp(t, cfg, size[0], size[1])
			d.send(pushViewMsg{v: newModelsView(m.cfg)})
			v, ok := m.top().(*modelsView)
			if !ok {
				t.Fatalf("top view is %T", m.top())
			}
			for i := range installerHosts {
				if i > 0 {
					d.key("right")
				}
				host := selectedModelsHost(t, d)
				assertFits(t, d, size[0], size[1])
				modelsRowFor(t, d, first)
				if size[0] == 80 && !v.box.scrollable() {
					t.Fatalf("%s: %d roles do not scroll at %dx%d", host, len(roles), size[0], size[1])
				}
				if v.box.scrollable() {
					for range 40 {
						d.key("down")
					}
					modelsRowFor(t, d, last)
					assertFits(t, d, size[0], size[1])
					for range 40 {
						d.key("up")
					}
					modelsRowFor(t, d, first)
					for range 5 {
						d.key("pgdown")
					}
					modelsRowFor(t, d, last)
					for range 5 {
						d.key("pgup")
					}
					modelsRowFor(t, d, first)
				} else {
					modelsRowFor(t, d, last)
				}
				if size[0] == 80 && !strings.Contains(d.screen(), longRoleName+"  ") {
					t.Errorf("%s: %s is cut at 80 columns:\n%s", host, longRoleName, d.screen())
				}
			}
		})
	}
}
