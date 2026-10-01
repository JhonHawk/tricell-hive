package management

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	"tricell-hive/integrations/agents"
)

// Synthetic values from testdata/agent-profiles.json that the assertions pin.
const (
	syntheticProfiles      = "testdata/agent-profiles.json"
	claudeExecModel        = "syn-claude-exec"
	claudeNextModel        = "syn-claude-next"
	codexExecModel         = "syn-codex-exec"
	codexNextModel         = "syn-codex-next"
	opencodeExecModel      = "syn-oc/exec"
	opencodeReasoningModel = "syn-oc/reason"
	opencodeExecEffort     = "medium"
	syntheticExecEffort    = "medium" // the execution profile's effort on Claude, Codex and Pi
	opencodeRoleEffort     = "max"
)

// swapModel replaces every occurrence of a synthetic model in the profiles
// file, and fails when the constant is absent so the swap cannot silently do
// nothing.
func swapModel(t *testing.T, profiles, from, to string) string {
	t.Helper()
	quoted := `"` + from + `"`
	if !strings.Contains(profiles, quoted) {
		t.Fatalf("synthetic profiles do not contain %s", quoted)
	}
	return strings.ReplaceAll(profiles, quoted, `"`+to+`"`)
}

var allAgentHosts = []string{"claude", "codex", "cursor", "grok", "opencode", "pi"}

func modelsRole(name, profile, effort string) string {
	front := "---\nname: " + name + "\ndescription: Test role\nmodel_profile: " + profile + "\naccess_profile: observe\n"
	if effort != "" {
		front += "effort: " + effort + "\n"
	}
	return front + "---\nUse evidence.\n"
}

// modelsSource writes the synthetic agent profiles and three roles into o.Source:
// one on the execution profile, one on reasoning with its own effort, and one
// that inherits.
func modelsSource(t *testing.T, o Options) {
	t.Helper()
	profiles, err := os.ReadFile(syntheticProfiles)
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, agents.ProfilesSource), string(profiles))
	put(t, filepath.Join(o.Source, "content/agents/design/plain-role.md"), modelsRole("plain-role", "execution", ""))
	put(t, filepath.Join(o.Source, "content/agents/review/deep-role.md"), modelsRole("deep-role", "reasoning", "max"))
	put(t, filepath.Join(o.Source, "content/agents/quality/inherit-role.md"), modelsRole("inherit-role", "inherit", "low"))
}

// installedField reads a scalar frontmatter or TOML field of an installed agent.
func installedField(text, key string) string {
	m := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `(?:: | = )"(.*)"$`).FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return m[1]
}

func modelRowsByKey(rows []ModelRow) map[string]ModelRow {
	out := map[string]ModelRow{}
	for _, r := range rows {
		out[r.Host+"/"+r.Role] = r
	}
	return out
}

func TestEffectiveModelsMatchInstalledAgentFiles(t *testing.T) {
	o := setup(t)
	o.Hosts = allAgentHosts
	modelsSource(t, o)
	apply(t, plan(t, "install", o))

	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3*len(allAgentHosts) {
		t.Fatalf("got %d rows, want %d: %+v", len(rows), 3*len(allAgentHosts), rows)
	}
	for i := 1; i < len(rows); i++ {
		a, b := rows[i-1], rows[i]
		if a.Host > b.Host || (a.Host == b.Host && a.Role >= b.Role) {
			t.Fatalf("rows are not sorted by host then role: %+v then %+v", a, b)
		}
	}
	state := stateFor(t, o)
	checked := 0
	for _, record := range state.Records {
		if record.Target.Kind != "agent" {
			continue
		}
		role := strings.TrimSuffix(filepath.Base(record.Target.Source), ".md")
		if len(record.Consumers) != 1 {
			t.Fatalf("agent %s has %d consumers, want 1", record.Target.Path, len(record.Consumers))
		}
		host := record.Consumers[0].Host
		row, ok := modelRowsByKey(rows)[host+"/"+role]
		if !ok {
			t.Fatalf("no row for %s/%s", host, role)
		}
		text := get(t, record.Target.Path)
		var model, effort string
		switch row.Host {
		case "codex":
			model, effort = installedField(text, "model"), installedField(text, "model_reasoning_effort")
		case "pi":
			model, effort = installedField(text, "model"), installedField(text, "thinking")
		case "opencode":
			model = installedField(text, "model")
			// OpenCode carries its level as a #variant suffix of the model.
			if base, variant, found := strings.Cut(model, "#"); found {
				model, effort = base, variant
			}
		default:
			model, effort = installedField(text, "model"), installedField(text, "effort")
		}
		wantModel, wantEffort := row.Model, row.Effort
		if row.Host == "opencode" {
			if base, variant, found := strings.Cut(wantModel, "#"); found {
				wantModel, wantEffort = base, variant
			}
		}
		if model != wantModel || effort != wantEffort {
			t.Errorf("%s/%s: file has model %q effort %q, row has %q %q", row.Host, role, model, effort, wantModel, wantEffort)
		}
		checked++
	}
	if checked != len(rows) {
		t.Fatalf("compared %d installed files against %d rows", checked, len(rows))
	}
}

func TestEffectiveModelsRoleEffortWinsOnClaudeCodexAndPi(t *testing.T) {
	o := setup(t)
	o.Hosts = allAgentHosts
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	byKey := modelRowsByKey(rows)
	for _, host := range []string{"claude", "codex", "pi"} {
		if r := byKey[host+"/deep-role"]; r.Effort != "max" || r.Profile != "reasoning" {
			t.Errorf("%s/deep-role = %+v, want the role's own max effort", host, r)
		}
		if r := byKey[host+"/inherit-role"]; r.Effort != "low" || r.Profile != "inherit" {
			t.Errorf("%s/inherit-role = %+v, want the role's own low effort", host, r)
		}
		if r := byKey[host+"/plain-role"]; r.Effort != syntheticExecEffort || r.Profile != "execution" {
			t.Errorf("%s/plain-role = %+v, want the execution profile's %s effort", host, r, syntheticExecEffort)
		}
	}
	for _, host := range []string{"grok", "opencode", "cursor"} {
		if r := byKey[host+"/deep-role"]; r.Effort != "" {
			t.Errorf("%s/deep-role effort = %q, want none", host, r.Effort)
		}
	}
	if r := byKey["grok/plain-role"]; r.Model != "" {
		t.Errorf("grok model = %q, want the host default (empty)", r.Model)
	}
	if r := byKey["opencode/plain-role"]; r.Model != opencodeExecModel+"#"+opencodeExecEffort {
		t.Errorf("opencode model = %q", r.Model)
	}
	// On OpenCode a role's effort replaces the profile's variant.
	if r := byKey["opencode/deep-role"]; r.Model != opencodeReasoningModel+"#"+opencodeRoleEffort {
		t.Errorf("opencode/deep-role model = %q, want the role's max variant", r.Model)
	}
}

func TestEffectiveModelsUseTheRecordsOwnRelease(t *testing.T) {
	o := setup(t)
	modelsSource(t, o)
	o.Hosts = []string{"claude"}
	apply(t, plan(t, "install", o))

	// A second release changes both models, but only Codex moves to it.
	profilesPath := filepath.Join(o.Source, agents.ProfilesSource)
	profiles := swapModel(t, get(t, profilesPath), claudeExecModel, claudeNextModel)
	profiles = swapModel(t, profiles, codexExecModel, codexNextModel)
	put(t, profilesPath, profiles)
	o.Hosts = []string{"codex"}
	apply(t, plan(t, "install", o))

	o.Hosts = nil
	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	byKey := modelRowsByKey(rows)
	if got := byKey["claude/plain-role"].Model; got != claudeExecModel {
		t.Errorf("claude stayed on the older release, model = %q, want %q", got, claudeExecModel)
	}
	if got := byKey["codex/plain-role"].Model; got != codexNextModel {
		t.Errorf("codex is on the newer release, model = %q, want %q", got, codexNextModel)
	}
	state := stateFor(t, o)
	releases := map[string]bool{}
	for _, r := range state.Records {
		if r.Target.Kind == "agent" {
			releases[r.Release] = true
		}
	}
	if len(releases) != 2 {
		t.Fatalf("test setup: agents should span two releases, got %d", len(releases))
	}
}

func TestEffectiveModelsFiltersByHostsAndHome(t *testing.T) {
	o := setup(t)
	modelsSource(t, o)
	o.Hosts = []string{"claude", "codex"}
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	rows, err := EffectiveModels(o)
	if err != nil || len(rows) != 3 {
		t.Fatalf("rows = %+v, %v; want the three Codex rows", rows, err)
	}
	other := o
	other.Home = filepath.Join(filepath.Dir(o.Home), "other-home")
	other.Hosts = nil
	if err := os.MkdirAll(other.Home, 0700); err != nil {
		t.Fatal(err)
	}
	if rows, err = EffectiveModels(other); err != nil || len(rows) != 0 {
		t.Fatalf("another home shows %+v, %v; want none", rows, err)
	}
}

func TestEffectiveModelsWithNoHostsIsEmptyAndWritesNothing(t *testing.T) {
	o := setup(t)
	rows, err := EffectiveModels(o)
	if err != nil || rows == nil || len(rows) != 0 {
		t.Fatalf("EffectiveModels = %#v, %v; want an empty list and no error", rows, err)
	}
	absent(t, o.StateDir)
}

func TestEffectiveModelsReportsAMissingSnapshot(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude"}
	modelsSource(t, o)
	p := plan(t, "install", o)
	apply(t, p)
	if err := os.Remove(filepath.Join(o.StateDir, "releases", p.Release.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if _, err := EffectiveModels(o); err == nil {
		t.Fatal("a missing release snapshot was not reported")
	}
}

// twoReleaseClaudeState installs Claude on one release, moves Codex to a
// second release that changes the Claude model, and returns the options plus
// the two release IDs (older, newer).
func twoReleaseClaudeState(t *testing.T) (Options, string, string) {
	t.Helper()
	o := setup(t)
	modelsSource(t, o)
	o.Hosts = []string{"claude"}
	first := plan(t, "install", o)
	apply(t, first)
	profilesPath := filepath.Join(o.Source, agents.ProfilesSource)
	put(t, profilesPath, swapModel(t, get(t, profilesPath), claudeExecModel, claudeNextModel))
	o.Hosts = []string{"codex"}
	second := plan(t, "install", o)
	apply(t, second)
	o.Hosts = nil
	return o, first.Release.ID, second.Release.ID
}

// addShadowedCopies records, for every Claude agent, a second record at a path
// the current layout does not use, on the given release.
func addShadowedCopies(t *testing.T, o Options, release string) {
	t.Helper()
	s := stateFor(t, o)
	for path, r := range s.Records {
		if r.Target.Kind != "agent" || len(r.Consumers) != 1 || r.Consumers[0].Host != "claude" {
			continue
		}
		old := r
		old.Target.Path = filepath.Join(filepath.Dir(path), "retired-layout", filepath.Base(path))
		old.Release = release
		s.Records[old.Target.Path] = old
	}
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
}

func TestEffectiveModelsPreferTheCurrentDestinationOverAShadowedRecord(t *testing.T) {
	o, _, newer := twoReleaseClaudeState(t)
	// The shadowed copies sit on the newer release, so release age alone
	// would pick them; the record at the current destination is on the older
	// release and must still win.
	addShadowedCopies(t, o, newer)
	for i := 0; i < 50; i++ {
		rows, err := EffectiveModels(o)
		if err != nil {
			t.Fatal(err)
		}
		if got := modelRowsByKey(rows)["claude/plain-role"].Model; got != claudeExecModel {
			t.Fatalf("iteration %d: claude/plain-role model = %q, want %q from the current destination", i, got, claudeExecModel)
		}
	}
}

func TestEffectiveModelsFallBackToTheNewestReleaseWhenNoRecordIsCurrent(t *testing.T) {
	o, older, newer := twoReleaseClaudeState(t)
	s := stateFor(t, o)
	// Move every Claude agent record away from the current destination and
	// add a copy on the other release, so both are shadowed.
	for path, r := range s.Records {
		if r.Target.Kind != "agent" || len(r.Consumers) != 1 || r.Consumers[0].Host != "claude" {
			continue
		}
		delete(s.Records, path)
		r.Target.Path = filepath.Join(filepath.Dir(path), "retired-a", filepath.Base(path))
		r.Release = older
		s.Records[r.Target.Path] = r
		b := r
		b.Target.Path = filepath.Join(filepath.Dir(path), "retired-b", filepath.Base(path))
		b.Release = newer
		s.Records[b.Target.Path] = b
	}
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
	when := time.Now()
	for id, at := range map[string]time.Time{older: when.Add(-time.Hour), newer: when} {
		if err := os.Chtimes(filepath.Join(o.StateDir, "releases", id+".json"), at, at); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 50; i++ {
		rows, err := EffectiveModels(o)
		if err != nil {
			t.Fatal(err)
		}
		if got := modelRowsByKey(rows)["claude/plain-role"].Model; got != claudeNextModel {
			t.Fatalf("iteration %d: claude/plain-role model = %q, want %q from the newest release", i, got, claudeNextModel)
		}
	}
}

// TestEffectiveModelsRowsCarryTheGroupOfTheirSourcePath fixes ModelRow.Group
// (#46, T8): the folder under content/agents/ that holds the role's source.
func TestEffectiveModelsRowsCarryTheGroupOfTheirSourcePath(t *testing.T) {
	o := setup(t)
	o.Hosts = allAgentHosts
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"plain-role": "design", "deep-role": "review", "inherit-role": "quality"}
	for _, r := range rows {
		if r.Group != want[r.Role] {
			t.Errorf("%s/%s: group %q, want %q", r.Host, r.Role, r.Group, want[r.Role])
		}
	}
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
}
