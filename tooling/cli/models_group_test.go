package main

import (
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// Tests of the role groups in `hive models` and of --group on `hive models set`
// and `hive models reset` (#46, T8, D7-A). Every home and source is synthetic:
// the synthetic source has the design group (hive-design-architecture and
// plain-role) and the quality group (inherit-role).

const modelsRowsGoldenPath = "testdata/models_rows.golden"

func (e modelsEnv) stored(t *testing.T) map[string]map[string]management.ModelOverride {
	t.Helper()
	s, err := management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// withoutGroupHeaders drops the group header lines, which are the only lines
// of `hive models` that name a group and carry no role.
func withoutGroupHeaders(text string) string {
	var keep []string
	for _, l := range strings.Split(text, "\n") {
		f := strings.Fields(l)
		if len(f) > 0 && (f[0] == "design" || f[0] == "quality") {
			continue
		}
		keep = append(keep, l)
	}
	return strings.Join(keep, "\n")
}

func TestModelsCommandPrintsOneHeaderPerGroupInOrder(t *testing.T) {
	e := newModelsEnv(t)
	text := e.modelsText(t)
	claude := text[strings.Index(text, "claude\n"):strings.Index(text, "\ncodex\n")]
	lines := strings.Split(strings.TrimRight(claude, "\n"), "\n")
	want := []string{
		"claude",
		"  Role                      Profile     Model                     Effort",
		"  design  mixed  mixed",
		"  hive-design-architecture  reasoning   syn-claude-reason         high",
		"  plain-role                execution   syn-claude-exec           medium",
		"  quality  inherit (parent session)  low",
		"  inherit-role              inherit     inherit (parent session)  low",
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("claude block:\n%s\nwant:\n%s", strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
	// grok has neither a model nor an effort per role: the groups still show.
	if !strings.Contains(text, "  design  host default  -\n") || !strings.Contains(text, "  quality  host default  -\n") {
		t.Fatalf("grok headers missing:\n%s", text)
	}
}

func TestModelsCommandHeaderShowsTheCommonValuesAndTheMarker(t *testing.T) {
	e := newModelsEnv(t)
	if _, err := e.write("set", true, "y\n", "--host", "claude", "--group", "design", "--model", "opus", "--effort", "high"); err != nil {
		t.Fatal(err)
	}
	text := e.modelsText(t)
	if !strings.Contains(text, "  design *  opus  high\n") {
		t.Fatalf("a group whose roles share one setting shows it, marked:\n%s", text)
	}
}

// sortedRoleLines sorts the role lines of every CLI block, because the base
// text lists roles by name and the grouped text lists them by group first. The
// CLI name and the column header stay where they are.
func sortedRoleLines(text string) string {
	blocks := strings.Split(text, "\n\n")
	for i, b := range blocks {
		lines := strings.Split(b, "\n")
		if len(lines) > 2 {
			sort.Strings(lines[2:])
		}
		blocks[i] = strings.Join(lines, "\n")
	}
	return strings.Join(blocks, "\n\n")
}

func TestModelsCommandRowsWithoutHeadersAreTheBaseText(t *testing.T) {
	e := newModelsEnv(t)
	got := sortedRoleLines(withoutGroupHeaders(e.modelsText(t)))
	base, err := os.ReadFile(modelsRowsGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	want := sortedRoleLines(string(base))
	if got != want {
		t.Fatalf("rows changed once the headers are removed:\n%s\nwant:\n%s", got, want)
	}
}

func TestModelsCommandSetGroupWritesTheGivenPartsOnEveryRole(t *testing.T) {
	e := newModelsEnv(t)
	out, err := e.write("set", true, "y\n", "--host", "claude", "--group", "design", "--model", "opus")
	if err != nil {
		t.Fatalf("set --group: %v\n%s", err, out)
	}
	st := e.stored(t)["claude"]
	want := map[string]management.ModelOverride{
		"hive-design-architecture": {Model: "opus"},
		"plain-role":               {Model: "opus"},
	}
	if !reflect.DeepEqual(st, want) {
		t.Fatalf("stored = %+v, want %+v", st, want)
	}
	if !strings.Contains(e.read(t, e.agentFile("claude", "plain-role")), `model: "opus"`) {
		t.Fatal("plain-role file does not carry the model")
	}
	if strings.Contains(out, "Replaces the own override") {
		t.Fatalf("no role had an override, yet:\n%s", out)
	}
}

func TestModelsCommandSetGroupReplacesTheOwnOverrideAndNamesTheRole(t *testing.T) {
	e := newModelsEnv(t)
	if _, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--effort", "max"); err != nil {
		t.Fatal(err)
	}
	out, err := e.write("set", true, "y\n", "--host", "claude", "--group", "design", "--model", "opus")
	if err != nil {
		t.Fatalf("set --group: %v\n%s", err, out)
	}
	if !strings.HasPrefix(out, "Change the Design group on claude (2 roles)\n\n") || !strings.Contains(out, "\nReplaces the own override of plain-role.\n") {
		t.Fatalf("summary lacks the title or the replaced override:\n%s", out)
	}
	if got := e.stored(t)["claude"]["plain-role"]; got != (management.ModelOverride{Model: "opus"}) {
		t.Fatalf("plain-role stored = %+v, want the model alone: the effort override is replaced", got)
	}
}

func TestModelsCommandSetGroupEqualsTheRolesOneByOne(t *testing.T) {
	cases := []struct {
		name string
		host string
		args []string
	}{
		{"claude model and effort", "claude", []string{"--model", "opus", "--effort", "high"}},
		{"claude effort only", "claude", []string{"--effort", "low"}},
		{"opencode model only", "opencode", []string{"--model", "x/y"}},
		{"opencode effort", "opencode", []string{"--model", "x/y", "--effort", "low"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a, b := newModelsEnv(t), newModelsEnv(t)
			if out, err := a.write("set", true, "y\n", append([]string{"--host", c.host, "--group", "design"}, c.args...)...); err != nil {
				t.Fatalf("group: %v\n%s", err, out)
			}
			for _, role := range []string{"hive-design-architecture", "plain-role"} {
				if out, err := b.write("set", true, "y\n", append([]string{"--host", c.host, "--role", role}, c.args...)...); err != nil {
					t.Fatalf("role %s: %v\n%s", role, err, out)
				}
			}
			if sa, sb := a.stored(t), b.stored(t); !reflect.DeepEqual(sa, sb) {
				t.Fatalf("overrides differ: group %+v, roles %+v", sa, sb)
			}
			for _, role := range []string{"hive-design-architecture", "plain-role", "inherit-role"} {
				fa, fb := a.read(t, a.agentFile(c.host, role)), b.read(t, b.agentFile(c.host, role))
				if fa != fb {
					t.Errorf("%s file differs:\n%s\nvs\n%s", role, fa, fb)
				}
			}
			if a.modelsText(t) != b.modelsText(t) {
				t.Errorf("hive models differs:\n%s\nvs\n%s", a.modelsText(t), b.modelsText(t))
			}
		})
	}
}

func TestModelsCommandResetGroupRemovesTheOverridesOfTheGroup(t *testing.T) {
	e := newModelsEnv(t)
	clean := e.snapshot(t)
	for _, args := range [][]string{
		{"--host", "claude", "--group", "design", "--model", "opus", "--effort", "high"},
		{"--host", "claude", "--role", "inherit-role", "--effort", "max"},
	} {
		if out, err := e.write("set", true, "y\n", args...); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
	}
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--group", "design", "--only", "effort"); err != nil {
		t.Fatalf("reset --only effort: %v\n%s", err, out)
	}
	st := e.stored(t)["claude"]
	want := map[string]management.ModelOverride{
		"hive-design-architecture": {Model: "opus"},
		"plain-role":               {Model: "opus"},
		"inherit-role":             {Effort: "max"},
	}
	if !reflect.DeepEqual(st, want) {
		t.Fatalf("after reset --only effort: %+v, want %+v", st, want)
	}
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--group", "design"); err != nil {
		t.Fatalf("reset --group: %v\n%s", err, out)
	}
	if got := e.stored(t)["claude"]; !reflect.DeepEqual(got, map[string]management.ModelOverride{"inherit-role": {Effort: "max"}}) {
		t.Fatalf("after reset --group: %+v; only the other group's override should remain", got)
	}
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--group", "quality"); err != nil {
		t.Fatalf("reset quality: %v\n%s", err, out)
	}
	now := e.snapshot(t)
	for path, text := range clean {
		if strings.HasPrefix(path, e.stateDir) && !strings.HasSuffix(path, "state.json") {
			continue
		}
		if now[path] != text {
			t.Errorf("%s differs from the clean installation", path)
		}
	}
}

func TestModelsCommandResetGroupWithNothingSetChangesNothing(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	out, err := e.write("reset", true, "y\n", "--host", "claude", "--group", "design")
	if err != nil || !strings.Contains(out, "Nothing to change") {
		t.Fatalf("err = %v, output %q", err, out)
	}
	if !sameFiles(before, e.snapshot(t)) {
		t.Fatal("a reset with nothing to remove wrote something")
	}
}

func TestModelsCommandGroupRejectionsWriteNothing(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	cases := []struct {
		name string
		sub  string
		args []string
		want string
	}{
		{"set with role and group", "set", []string{"--host", "claude", "--group", "design", "--role", "plain-role", "--model", "x"}, "--group and --role cannot be combined"},
		{"reset with role and group", "reset", []string{"--host", "claude", "--group", "design", "--role", "plain-role"}, "--group and --role cannot be combined"},
		{"reset all with group", "reset", []string{"--host", "claude", "--group", "design", "--all"}, "--all cannot be combined with --group"},
		{"unknown group on set", "set", []string{"--host", "claude", "--group", "ghost", "--model", "x"}, `group "ghost" is unknown for claude; known groups: design, quality`},
		{"unknown group on reset", "reset", []string{"--host", "claude", "--group", "ghost"}, `group "ghost" is unknown for claude; known groups: design, quality`},
		{"group set without a part", "set", []string{"--host", "claude", "--group", "design"}, "needs --model or --effort"},
		{"group with an invalid effort", "set", []string{"--host", "grok", "--group", "design", "--effort", "high"}, "grok does not support effort"},
		{"unregistered CLI", "set", []string{"--host", "cursor", "--group", "design", "--model", "x"}, "cursor is not installed"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := e.write(c.sub, true, "y\n", c.args...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error = %v (output %q), want %q", err, out, c.want)
			}
			if !sameFiles(before, e.snapshot(t)) {
				t.Fatal("a rejected command wrote something")
			}
		})
	}
}

// --- OpenCode keeps the shown effort when only the model changes ------------

func TestModelsCommandOpenCodeRoleModelKeepsTheShownEffort(t *testing.T) {
	e := newModelsEnv(t)
	if out, err := e.write("set", true, "y\n", "--host", "opencode", "--role", "inherit-role", "--model", "x/y"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := e.stored(t)["opencode"]["inherit-role"]; got != (management.ModelOverride{Model: "x/y", Effort: "max"}) {
		t.Fatalf("stored = %+v, want model x/y with the shown effort max", got)
	}
	if !strings.Contains(e.read(t, e.agentFile("opencode", "inherit-role")), `model: "x/y#max"`) {
		t.Fatalf("file:\n%s", e.read(t, e.agentFile("opencode", "inherit-role")))
	}
}

func TestModelsCommandOpenCodeGroupWithMixedEffortsKeepsEachRolesEffort(t *testing.T) {
	e := newModelsEnv(t)
	if out, err := e.write("set", true, "y\n", "--host", "opencode", "--group", "design", "--model", "x/y"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	want := map[string]management.ModelOverride{
		"hive-design-architecture": {Model: "x/y", Effort: "high"},
		"plain-role":               {Model: "x/y", Effort: "medium"},
	}
	if got := e.stored(t)["opencode"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("stored = %+v, want %+v", got, want)
	}
	for role, variant := range map[string]string{"hive-design-architecture": "high", "plain-role": "medium"} {
		if !strings.Contains(e.read(t, e.agentFile("opencode", role)), `model: "x/y#`+variant+`"`) {
			t.Errorf("%s file lacks x/y#%s", role, variant)
		}
	}
}

func TestModelsCommandOpenCodeModelWithAnEffortWritesOnlyTheGivenParts(t *testing.T) {
	e := newModelsEnv(t)
	if out, err := e.write("set", true, "y\n", "--host", "opencode", "--role", "inherit-role", "--model", "x/y", "--effort", "low"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := e.stored(t)["opencode"]["inherit-role"]; got != (management.ModelOverride{Model: "x/y", Effort: "low"}) {
		t.Fatalf("stored = %+v", got)
	}
}
