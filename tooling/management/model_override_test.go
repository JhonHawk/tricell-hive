package management

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"tricell-hive/integrations/agents"
	"tricell-hive/tooling/legacy"
)

// Release values the synthetic profiles give plain-role (execution) on Claude.
const (
	plainRole    = "plain-role"
	inheritRole  = "inherit-role"
	claudeEffort = "medium"
)

// snapshotHome hashes every regular file under the test home.
func snapshotHome(t *testing.T, o Options) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(o.Home, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(b)
			out[path] = hex.EncodeToString(sum[:])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func changedPaths(before, after map[string]string) []string {
	var out []string
	for p, h := range after {
		if before[p] != h {
			out = append(out, p)
		}
	}
	for p := range before {
		if _, ok := after[p]; !ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

// agentFile returns where host's role file is installed.
func agentFile(t *testing.T, o Options, host, role string) string {
	t.Helper()
	for path, r := range stateFor(t, o).Records {
		if r.Target.Kind == "agent" && strings.TrimSuffix(filepath.Base(r.Target.Source), ".md") == role && len(r.Consumers) == 1 && r.Consumers[0].Host == host {
			return path
		}
	}
	t.Fatalf("no %s agent file for %s", role, host)
	return ""
}

func buildModels(t *testing.T, o Options, host string, next map[string]ModelOverride) Plan {
	t.Helper()
	p, err := BuildModelsPlan(o, host, next)
	if err != nil {
		t.Fatalf("BuildModelsPlan(%s, %+v): %v", host, next, err)
	}
	return p
}

func stateBytes(t *testing.T, o Options) string {
	t.Helper()
	return get(t, filepath.Join(o.StateDir, "state.json"))
}

// versionStatuses returns the distinct version states Status reports per CLI.
func versionStatuses(t *testing.T, o Options) map[string]string {
	t.Helper()
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for _, en := range entries {
		if en.VersionStatus == "" {
			continue
		}
		if prev, ok := out[en.Host]; ok && prev != en.VersionStatus {
			t.Fatalf("%s reports both %s and %s", en.Host, prev, en.VersionStatus)
		}
		out[en.Host] = en.VersionStatus
	}
	return out
}

func allVerified(t *testing.T, o Options) {
	t.Helper()
	got := versionStatuses(t, o)
	if len(got) != len(o.Hosts) {
		t.Fatalf("version status for %v, want every host of %v", got, o.Hosts)
	}
	for host, status := range got {
		if status != "verified" {
			t.Fatalf("%s is %s, want verified", host, status)
		}
	}
}

// --- AC1: one file changes, in each host's own format -----------------------

func TestModelsPlanOverrideChangesOnlyTheRoleFile(t *testing.T) {
	cases := []struct {
		host string
		set  ModelOverride
		want []string
	}{
		{"claude", ModelOverride{Model: "opus", Effort: "max"}, []string{`model: "opus"`, `effort: "max"`}},
		{"codex", ModelOverride{Model: "gpt-x", Effort: "high"}, []string{`model = "gpt-x"`, `model_reasoning_effort = "high"`}},
		{"pi", ModelOverride{Model: "p/m", Effort: "xhigh"}, []string{`model: "p/m"`, `thinking: "xhigh"`}},
		{"opencode", ModelOverride{Model: "o/m", Effort: "low"}, []string{`model: "o/m#low"`}},
		{"grok", ModelOverride{Model: "g-1"}, []string{`model: "g-1"`}},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			o, _ := installedFiveHosts(t)
			target := agentFile(t, o, c.host, plainRole)
			before := snapshotHome(t, o)
			p := buildModels(t, o, c.host, map[string]ModelOverride{plainRole: c.set})
			if !reflect.DeepEqual(p.Hosts, []string{c.host}) || p.Action != "install" {
				t.Fatalf("plan = %s for %v, want an install for %s only", p.Action, p.Hosts, c.host)
			}
			apply(t, p)
			after := snapshotHome(t, o)
			if got := changedPaths(before, after); !reflect.DeepEqual(got, []string{target}) {
				t.Fatalf("changed files = %v, want only %s", got, target)
			}
			text := get(t, target)
			for _, want := range c.want {
				if !strings.Contains(text, want+"\n") {
					t.Errorf("%s has no line %q:\n%s", target, want, text)
				}
			}
			if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, map[string]map[string]ModelOverride{c.host: {plainRole: c.set}}) {
				t.Fatalf("stored overrides = %+v", got)
			}
			allVerified(t, o)
		})
	}
}

// --- AC2: the override survives every reinstall -------------------------------

// newerRelease edits the synthetic profiles and the product version so the
// next install is a new release of a new product version.
func newerRelease(t *testing.T, o Options, version string) {
	t.Helper()
	profilesPath := filepath.Join(o.Source, agents.ProfilesSource)
	put(t, profilesPath, swapModel(t, get(t, profilesPath), claudeExecModel, claudeNextModel))
	put(t, filepath.Join(o.Source, "VERSION"), version+"\n")
}

func TestModelOverrideSurvivesUpdateAndKeepsReceiptsVerified(t *testing.T) {
	o, first := installedFiveHosts(t)
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	allVerified(t, o)

	newerRelease(t, o, "1.2.1")
	update := plan(t, "install", o)
	if update.Release.ID == first.Release.ID {
		t.Fatal("test setup: the profile change did not make a new release")
	}
	if !reflect.DeepEqual(update.ModelOverrides, map[string]map[string]ModelOverride{"claude": {plainRole: {Effort: "max"}}}) {
		t.Fatalf("update plan carries %+v, want the state's override", update.ModelOverrides)
	}
	apply(t, update)
	text := get(t, agentFile(t, o, "claude", plainRole))
	for _, want := range []string{`model: "` + claudeNextModel + `"`, `effort: "max"`} {
		if !strings.Contains(text, want+"\n") {
			t.Errorf("after update the claude role has no line %q:\n%s", want, text)
		}
	}
	if !strings.Contains(get(t, agentFile(t, o, "claude", "deep-role")), `model: "syn-claude-reason"`) {
		t.Error("an unrelated role lost its release model")
	}
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, map[string]map[string]ModelOverride{"claude": {plainRole: {Effort: "max"}}}) {
		t.Fatalf("override after update = %+v", got)
	}
	allVerified(t, o)
}

func TestModelOverrideSurvivesRollbackWithTheSameVersionStateAsWithout(t *testing.T) {
	rollback := func(withOverride bool) (map[string]string, string, Options) {
		o, first := installedFiveHosts(t)
		if withOverride {
			apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Model: "opus", Effort: "max"}}))
		}
		newerRelease(t, o, "1.2.1")
		apply(t, plan(t, "install", o))
		back := o
		back.ReleaseID = first.Release.ID
		apply(t, plan(t, "install", back))
		return versionStatuses(t, o), get(t, agentFile(t, o, "claude", plainRole)), o
	}
	without, _, _ := rollback(false)
	with, text, o := rollback(true)
	if !reflect.DeepEqual(with, without) {
		t.Fatalf("version state after rollback is %v with an override, %v without", with, without)
	}
	for _, want := range []string{`model: "opus"`, `effort: "max"`} {
		if !strings.Contains(text, want+"\n") {
			t.Errorf("after rollback the claude role has no line %q:\n%s", want, text)
		}
	}
	if len(stateFor(t, o).ModelOverrides["claude"]) != 1 {
		t.Fatalf("override after rollback = %+v", stateFor(t, o).ModelOverrides)
	}
}

// --- AC3: removing overrides ---------------------------------------------------

func TestModelsPlanEmptySetRestoresACleanInstall(t *testing.T) {
	o, _ := installedFiveHosts(t)
	cleanFile := get(t, agentFile(t, o, "claude", plainRole))
	cleanState := stateBytes(t, o)
	cleanHome := snapshotHome(t, o)

	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Model: "opus", Effort: "max"}, "deep-role": {Effort: "low"}}))
	if get(t, agentFile(t, o, "claude", plainRole)) == cleanFile {
		t.Fatal("test setup: the override changed nothing")
	}
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{}))
	if got := get(t, agentFile(t, o, "claude", plainRole)); got != cleanFile {
		t.Fatalf("after reset the role differs from a clean install:\n%s\nwant\n%s", got, cleanFile)
	}
	if got := changedPaths(cleanHome, snapshotHome(t, o)); len(got) != 0 {
		t.Fatalf("files differ from the clean install: %v", got)
	}
	if got := stateBytes(t, o); got != cleanState {
		t.Fatalf("state after reset differs from the clean state:\n%s", got)
	}
	allVerified(t, o)
}

func TestModelOverrideRemoveDropsOnlyTheRemovedCLIsOverrides(t *testing.T) {
	o, _ := installedFiveHosts(t)
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	apply(t, buildModels(t, o, "codex", map[string]ModelOverride{plainRole: {Effort: "high"}}))

	removal := o
	removal.Hosts = []string{"codex"}
	rp := plan(t, "remove", removal)
	if len(rp.ModelOverrides) != 0 {
		t.Fatalf("a remove plan carries overrides: %+v", rp.ModelOverrides)
	}
	apply(t, rp)
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, map[string]map[string]ModelOverride{"claude": {plainRole: {Effort: "max"}}}) {
		t.Fatalf("overrides after removing codex = %+v, want claude's only", got)
	}
	for _, host := range []string{"claude", "grok", "opencode", "pi"} {
		removal.Hosts = []string{host}
		apply(t, plan(t, "remove", removal))
	}
	if strings.Contains(stateBytes(t, o), "model_overrides") {
		t.Fatalf("removing every CLI left overrides:\n%s", stateBytes(t, o))
	}
}

// --- changes that only touch the state -----------------------------------------

func TestModelOverrideStateOnlyChangesAreSaved(t *testing.T) {
	t.Run("set equal to the release value", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		before := snapshotHome(t, o)
		p := buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: claudeEffort}})
		if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
			t.Fatalf("PlanUnchanged = %v, %v; an override that only changes the state is a change", unchanged, err)
		}
		if res := apply(t, p); res == "unchanged" {
			t.Fatal("Apply dropped a state-only change")
		}
		if got := changedPaths(before, snapshotHome(t, o)); len(got) != 0 {
			t.Fatalf("an override equal to the release value changed files: %v", got)
		}
		if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, map[string]map[string]ModelOverride{"claude": {plainRole: {Effort: claudeEffort}}}) {
			t.Fatalf("stored overrides = %+v", got)
		}
		allVerified(t, o)
	})
	t.Run("reset of an override that is not applied", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		clean := stateBytes(t, o)
		s := stateFor(t, o)
		s.ModelOverrides = map[string]map[string]ModelOverride{"claude": {"retired-role": {Model: "x"}}}
		if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
			t.Fatal(err)
		}
		// An install that keeps the override leaves it in place.
		keep := plan(t, "install", o)
		if keep.ModelOverrides["claude"]["retired-role"].Model != "x" {
			t.Fatalf("install plan dropped the unapplied override: %+v", keep.ModelOverrides)
		}
		p := buildModels(t, o, "claude", nil)
		if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
			t.Fatalf("PlanUnchanged = %v, %v; dropping an unapplied override is a change", unchanged, err)
		}
		apply(t, p)
		if got := stateBytes(t, o); got != clean {
			t.Fatalf("state after the reset differs from the clean state:\n%s", got)
		}
	})
	t.Run("unchanged when nothing differs", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		p := buildModels(t, o, "claude", nil)
		if unchanged, err := PlanUnchanged(p); err != nil || !unchanged {
			t.Fatalf("PlanUnchanged = %v, %v; an empty set over no overrides changes nothing", unchanged, err)
		}
	})
}

func TestModelOverrideVoiceAndRemovePlansKeepOtherOverrides(t *testing.T) {
	o, _ := installedFiveHosts(t)
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	apply(t, buildModels(t, o, "codex", map[string]ModelOverride{plainRole: {Model: "gpt-x"}}))
	want := stateFor(t, o).ModelOverrides

	voiceSource(t, o)
	vp := voicePlan(t, "set", o, jarvisSirSubtle)
	if len(vp.ModelOverrides) != 0 {
		t.Fatalf("a voice plan carries overrides: %+v", vp.ModelOverrides)
	}
	apply(t, vp)
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, want) {
		t.Fatalf("overrides after a voice plan = %+v, want %+v", got, want)
	}

	removal := o
	removal.Hosts = []string{"grok"}
	apply(t, plan(t, "remove", removal))
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, want) {
		t.Fatalf("overrides after removing another CLI = %+v, want %+v", got, want)
	}
	if !strings.Contains(get(t, agentFile(t, o, "codex", plainRole)), `model = "gpt-x"`) {
		t.Fatal("the codex role lost its override")
	}
}

func TestModelOverrideProjectScopeNeitherCarriesNorTouchesUserOverrides(t *testing.T) {
	o, _ := installedFiveHosts(t)
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	want := stateFor(t, o).ModelOverrides

	project := o
	project.Scope = "project"
	project.Root = filepath.Join(o.Home, "repo")
	project.Hosts = []string{"claude"}
	if err := os.MkdirAll(project.Root, 0700); err != nil {
		t.Fatal(err)
	}
	p := plan(t, "install", project)
	if len(p.ModelOverrides) != 0 {
		t.Fatalf("a project-scope plan carries user overrides: %+v", p.ModelOverrides)
	}
	apply(t, p)
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, want) {
		t.Fatalf("a project install changed the user overrides: %+v", got)
	}
	apply(t, plan(t, "remove", project))
	if got := stateFor(t, o).ModelOverrides; !reflect.DeepEqual(got, want) {
		t.Fatalf("a project remove changed the user overrides: %+v", got)
	}
}

// --- OpenCode ------------------------------------------------------------------

func TestModelOverrideOpenCodeEffortOnInheritProfileReplacesTheVariant(t *testing.T) {
	o, _ := installedFiveHosts(t)
	target := agentFile(t, o, "opencode", inheritRole)
	if !strings.Contains(get(t, target), `model: "syn-oc/inherit#max"`) {
		t.Fatalf("test setup: the inherit profile no longer carries #max:\n%s", get(t, target))
	}
	apply(t, buildModels(t, o, "opencode", map[string]ModelOverride{inheritRole: {Effort: "high"}}))
	text := get(t, target)
	if !strings.Contains(text, `model: "syn-oc/inherit#high"`+"\n") || strings.Count(text, "#") != 1 {
		t.Fatalf("OpenCode file does not carry exactly #high:\n%s", text)
	}
	if got := stateFor(t, o).ModelOverrides["opencode"][inheritRole]; got != (ModelOverride{Effort: "high"}) {
		t.Fatalf("stored override = %+v, want an effort-only override", got)
	}
}

// --- the effective model --------------------------------------------------------

func TestEffectiveModelsMarksOnlyOverriddenRows(t *testing.T) {
	o, _ := installedFiveHosts(t)
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Model: "opus", Effort: "max"}}))
	apply(t, buildModels(t, o, "opencode", map[string]ModelOverride{inheritRole: {Effort: "high"}}))
	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	byKey := modelRowsByKey(rows)
	if r := byKey["claude/"+plainRole]; r.Model != "opus" || r.Effort != "max" || !r.Override {
		t.Errorf("claude plain-role = %+v, want the override marked", r)
	}
	if r := byKey["opencode/"+inheritRole]; r.Model != "syn-oc/inherit#high" || !r.Override {
		t.Errorf("opencode inherit-role = %+v, want the #high override marked", r)
	}
	for key, r := range byKey {
		if key != "claude/"+plainRole && key != "opencode/"+inheritRole && r.Override {
			t.Errorf("%s is marked as overridden", key)
		}
	}
	// The rows agree with the installed files.
	text := get(t, agentFile(t, o, "claude", plainRole))
	if !strings.Contains(text, `model: "opus"`) || !strings.Contains(text, `effort: "max"`) {
		t.Fatalf("claude file disagrees with the row:\n%s", text)
	}
}

func TestEffectiveModelsKeepsAnUnappliedOverrideOutOfTheRows(t *testing.T) {
	o, _ := installedFiveHosts(t)
	s := stateFor(t, o)
	s.ModelOverrides = map[string]map[string]ModelOverride{"claude": {"retired-role": {Model: "x"}}}
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
	rows, err := EffectiveModels(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3*5 {
		t.Fatalf("got %d rows, want 15 and none for the retired role", len(rows))
	}
	for _, r := range rows {
		if r.Override {
			t.Fatalf("%+v is marked overridden", r)
		}
	}
}

// --- rejections ------------------------------------------------------------------

func TestModelsPlanRefusesWhatItCannotDoSafely(t *testing.T) {
	t.Run("a resource deleted by hand", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		if err := os.Remove(sharedPath(o)); err != nil {
			t.Fatal(err)
		}
		stateBefore, homeBefore := stateBytes(t, o), snapshotHome(t, o)
		_, err := BuildModelsPlan(o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}})
		if err == nil || !strings.Contains(err.Error(), "hive install") || !strings.Contains(err.Error(), "hive doctor") {
			t.Fatalf("error = %v, want a pointer to hive install or hive doctor", err)
		}
		if stateBytes(t, o) != stateBefore || len(changedPaths(homeBefore, snapshotHome(t, o))) != 0 {
			t.Fatal("a refused plan wrote something")
		}
	})
	t.Run("an agent file deleted by hand", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		if err := os.Remove(agentFile(t, o, "claude", plainRole)); err != nil {
			t.Fatal(err)
		}
		if _, err := BuildModelsPlan(o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}); err == nil || !strings.Contains(err.Error(), "hive install") {
			t.Fatalf("error = %v, want a pointer to hive install", err)
		}
	})
	t.Run("a non-agent resource that would change", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		s := stateFor(t, o)
		changed := 0
		for path, r := range s.Records {
			// Codex's own instruction block, which no other CLI shares.
			if r.Target.Kind == "block" && len(r.Consumers) == 1 && r.Consumers[0].Host == "codex" {
				r.Release = strings.Repeat("a", 64)
				s.Records[path] = r
				changed++
			}
		}
		if changed != 1 {
			t.Fatalf("test setup: found %d codex-only blocks, want 1", changed)
		}
		if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
			t.Fatal(err)
		}
		if _, err := BuildModelsPlan(o, "codex", map[string]ModelOverride{plainRole: {Effort: "max"}}); err == nil || !strings.Contains(err.Error(), "not an agent file") || !strings.Contains(err.Error(), "hive install") {
			t.Fatalf("error = %v, want a refusal that names the non-agent file and points to hive install", err)
		}
	})
	t.Run("agents from more than one release", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		s := stateFor(t, o)
		path := agentFile(t, o, "claude", plainRole)
		r := s.Records[path]
		r.Release = strings.Repeat("b", 64)
		s.Records[path] = r
		if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
			t.Fatal(err)
		}
		_, err := BuildModelsPlan(o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}})
		if err == nil || !strings.Contains(err.Error(), "Run hive update first; claude has agents from more than one release") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("a role the release does not have", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		_, err := BuildModelsPlan(o, "claude", map[string]ModelOverride{"ghost-role": {Model: "x"}})
		if err == nil || !strings.Contains(err.Error(), `role "ghost-role" is not in the release installed for claude`) {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("a CLI that is not registered", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		for host, want := range map[string]string{"cursor": "cursor is not installed", "vim": "unsupported host"} {
			if _, err := BuildModelsPlan(o, host, map[string]ModelOverride{plainRole: {Model: "x"}}); err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error = %v, want %q", host, err, want)
			}
		}
	})
	t.Run("invalid overrides", func(t *testing.T) {
		o, _ := installedFiveHosts(t)
		for _, c := range []struct {
			host string
			set  ModelOverride
			want string
		}{
			{"claude", ModelOverride{Model: "a b"}, "model may only contain"},
			{"grok", ModelOverride{Effort: "high"}, "grok does not support effort"},
			{"claude", ModelOverride{Effort: "extreme"}, "effort must be one of"},
			{"claude", ModelOverride{}, "needs a model or an effort"},
			{"pi", ModelOverride{Effort: "ultra"}, "not accepted by pi"},
		} {
			stateBefore := stateBytes(t, o)
			_, err := BuildModelsPlan(o, c.host, map[string]ModelOverride{plainRole: c.set})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("%s %+v: error = %v, want %q", c.host, c.set, err, c.want)
			}
			if stateBytes(t, o) != stateBefore {
				t.Error("a refused plan wrote the state")
			}
		}
	})
	t.Run("a CLI with no agents", func(t *testing.T) {
		o := setup(t)
		o.Hosts = []string{"claude"}
		apply(t, plan(t, "install", o))
		if _, err := BuildModelsPlan(o, "claude", map[string]ModelOverride{plainRole: {Model: "x"}}); err == nil || !strings.Contains(err.Error(), "no agent roles") {
			t.Fatalf("error = %v, want a refusal naming the missing agent roles", err)
		}
	})
}

func TestModelsPlanRefusalCoversLegacyAndMigration(t *testing.T) {
	o, _ := installedFiveHosts(t)
	good := buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}})
	if err := modelsPlanRefusal(good); err != nil {
		t.Fatalf("a plan that only rewrites agents is refused: %v", err)
	}
	withLegacy := good
	withLegacy.Legacy = append(withLegacy.Legacy, legacy.Edit{Path: "/x"})
	migration := good
	migration.Migration = &MigrationReceipt{}
	for name, p := range map[string]Plan{"legacy": withLegacy, "migration": migration} {
		if err := modelsPlanRefusal(p); err == nil || !strings.Contains(err.Error(), "hive install") {
			t.Errorf("%s plan: error = %v, want a pointer to hive install", name, err)
		}
	}
}

func TestStoredModelOverridesIsAnEmptyCopyOfTheState(t *testing.T) {
	o, _ := installedFiveHosts(t)
	got, err := StoredModelOverrides(o)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("StoredModelOverrides on a clean state = %#v, %v; want an empty, non-nil map", got, err)
	}
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	got, err = StoredModelOverrides(o)
	if err != nil || !reflect.DeepEqual(got, map[string]map[string]ModelOverride{"claude": {plainRole: {Effort: "max"}}}) {
		t.Fatalf("StoredModelOverrides = %+v, %v", got, err)
	}
	got["claude"][plainRole] = ModelOverride{Model: "mutated"}
	if again, _ := StoredModelOverrides(o); again["claude"][plainRole].Model != "" {
		t.Fatal("the returned map aliases the state")
	}
}
