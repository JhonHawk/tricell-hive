package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/target"
)

func grokHookPath(o Options) string {
	return filepath.Join(o.Home, ".grok", "hooks", "hive-read-whole-references.json")
}

func grokOptions(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"grok"}
	return o
}

// withHiveExecutable points the hook at path for one test and restores the
// suite default afterwards.
func withHiveExecutable(t *testing.T, path string) {
	t.Helper()
	old := HiveExecutable
	HiveExecutable = func() (string, error) { return path, nil }
	t.Cleanup(func() { HiveExecutable = old })
}

// copyExecutable returns a regular file outside the temporary directory that
// differs from the suite default, so reinstalling with it changes the command.
func copyExecutable(t *testing.T, name string) string {
	t.Helper()
	dir, err := testdataDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return path
}

// testdataDir is the package's testdata directory with symbolic links
// resolved, so a checkout reached through a link still passes the hook check.
func testdataDir() (string, error) {
	dir, err := filepath.Abs("testdata")
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(dir)
}

func canonicalTemp(t *testing.T) string {
	t.Helper()
	dir, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func suiteExecutable(t *testing.T) string {
	t.Helper()
	exe, err := HiveExecutable()
	if err != nil {
		t.Fatal(err)
	}
	return exe
}

func hookRecord(t *testing.T, o Options) Record {
	t.Helper()
	r, ok := stateFor(t, o).Records[grokHookPath(o)]
	if !ok {
		t.Fatal("hook is not recorded in the state")
	}
	return r
}

// AC3: install writes the hook, records it as a hook and every later read of
// the saved state succeeds.
func TestGrokHookInstallWritesAndRecordsTheFile(t *testing.T) {
	o := grokOptions(t)
	exe := suiteExecutable(t)
	p := plan(t, "install", o)
	if p.Executable != exe {
		t.Fatalf("plan executable = %q, want %q", p.Executable, exe)
	}
	apply(t, p)

	var f struct {
		Hooks struct {
			PreToolUse []struct {
				Matcher string `json:"matcher"`
				Hooks   []struct {
					Type    string `json:"type"`
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"PreToolUse"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(get(t, grokHookPath(o))), &f); err != nil {
		t.Fatalf("hook is not JSON: %v", err)
	}
	if len(f.Hooks.PreToolUse) != 1 || f.Hooks.PreToolUse[0].Matcher != "read_file" || len(f.Hooks.PreToolUse[0].Hooks) != 1 {
		t.Fatalf("hook groups: %+v", f.Hooks.PreToolUse)
	}
	want := "'" + exe + "' grok-hook read-whole --skills '" + filepath.Join(o.Home, ".agents", "skills") + "'"
	if got := f.Hooks.PreToolUse[0].Hooks[0]; got.Type != "command" || got.Command != want {
		t.Fatalf("command = %q, want %q", got.Command, want)
	}
	info, err := os.Stat(grokHookPath(o))
	if err != nil || info.Mode().Perm() != 0o644 {
		t.Fatalf("hook mode = %v (%v), want 0644", info.Mode().Perm(), err)
	}
	r := hookRecord(t, o)
	if r.Target.Kind != "hook" || r.Target.Source != grok.ReadHookSource || r.Mode != 0o644 || !r.CreatedFile {
		t.Fatalf("record = %+v", r)
	}
	if len(r.Consumers) != 1 || r.Consumers[0] != (Consumer{"grok", "user", o.Home}) {
		t.Fatalf("consumers = %+v", r.Consumers)
	}

	// The saved state is read again by status, update and remove.
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, en := range entries {
		if en.Path == grokHookPath(o) {
			found = true
			if en.Kind != "hook" || en.Status != "installed" {
				t.Fatalf("status entry = %+v", en)
			}
		}
	}
	if !found {
		t.Fatalf("status has no hook entry: %+v", entries)
	}
	if got := apply(t, plan(t, "install", o)); got != "unchanged" {
		t.Fatalf("update = %q, want unchanged", got)
	}
	if hookRecord(t, o).Mode != 0o644 {
		t.Fatal("update lost the hook mode")
	}
}

// AC4: remove deletes only the hook, even when hooks/ already held a file of
// another tool, and deletes the directory it created when it ends up empty.
func TestGrokHookRemoveKeepsForeignHooks(t *testing.T) {
	for _, preexisting := range []bool{true, false} {
		name := "hooks directory created by Hive"
		if preexisting {
			name = "hooks directory existed before"
		}
		t.Run(name, func(t *testing.T) {
			o := grokOptions(t)
			foreign := filepath.Join(o.Home, ".grok", "hooks", "engram-subagent-stop.json")
			if preexisting {
				put(t, foreign, `{"hooks":{}}`)
			}
			apply(t, plan(t, "install", o))
			if !preexisting {
				put(t, foreign, `{"hooks":{}}`)
			}
			apply(t, plan(t, "remove", o))
			absent(t, grokHookPath(o))
			if get(t, foreign) != `{"hooks":{}}` {
				t.Fatal("foreign hook file changed")
			}
			if _, ok := stateFor(t, o).Records[grokHookPath(o)]; ok {
				t.Fatal("hook is still recorded after remove")
			}
		})
	}
	t.Run("directory Hive created is removed when empty", func(t *testing.T) {
		o := grokOptions(t)
		apply(t, plan(t, "install", o))
		apply(t, plan(t, "remove", o))
		absent(t, filepath.Dir(grokHookPath(o)))
	})
}

func TestGrokHookForeignFileAtThePathCollides(t *testing.T) {
	o := grokOptions(t)
	put(t, grokHookPath(o), `{"hooks":{"PreToolUse":[]}}`)
	_, err := BuildPlan("install", o)
	if err == nil || !strings.Contains(err.Error(), "unowned hook") {
		t.Fatalf("foreign hook file error = %v, want an unowned hook collision", err)
	}
	if get(t, grokHookPath(o)) != `{"hooks":{"PreToolUse":[]}}` {
		t.Fatal("foreign hook file was changed")
	}
}

func TestGrokHookEditedManagedFileIsRejected(t *testing.T) {
	o := grokOptions(t)
	apply(t, plan(t, "install", o))
	edited := get(t, grokHookPath(o)) + " "
	put(t, grokHookPath(o), edited)
	for _, action := range []string{"install", "remove"} {
		_, err := BuildPlan(action, o)
		var changed *ManagedFileChangedError
		if !errors.As(err, &changed) || changed.Path != grokHookPath(o) {
			t.Fatalf("%s over an edited hook: error = %v, want ManagedFileChangedError", action, err)
		}
	}
	if get(t, grokHookPath(o)) != edited {
		t.Fatal("edited hook was overwritten")
	}
}

// AC5: another binary path rewrites the command.
func TestGrokHookReinstallWithAnotherExecutableRewritesTheCommand(t *testing.T) {
	o := grokOptions(t)
	first := suiteExecutable(t)
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, grokHookPath(o)), "'"+first+"' grok-hook") {
		t.Fatal("first install did not use the first executable")
	}
	second := copyExecutable(t, "hive-test-executable-second")
	withHiveExecutable(t, second)
	p := plan(t, "install", o)
	if p.Executable != second {
		t.Fatalf("plan executable = %q, want %q", p.Executable, second)
	}
	if got := apply(t, p); got == "unchanged" {
		t.Fatal("changing the executable did not change the plan")
	}
	text := get(t, grokHookPath(o))
	if !strings.Contains(text, "'"+second+"' grok-hook") || strings.Contains(text, "'"+first+"' grok-hook") {
		t.Fatalf("command was not rewritten:\n%s", text)
	}
	if string(hookRecord(t, o).Managed) != text {
		t.Fatal("recorded payload differs from the written file")
	}
}

func TestGrokHookModelsPlanWithTheSameExecutableDoesNotRefuse(t *testing.T) {
	o := grokOptions(t)
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	p, err := BuildModelsPlan(o, "grok", map[string]ModelOverride{"plain-role": {Model: "g-1"}})
	if err != nil {
		t.Fatalf("models plan refused: %v", err)
	}
	apply(t, p)

	withHiveExecutable(t, copyExecutable(t, "hive-test-executable-moved"))
	_, err = BuildModelsPlan(o, "grok", map[string]ModelOverride{"plain-role": {Model: "g-2"}})
	if err == nil || !strings.Contains(err.Error(), "is not an agent file and would change") {
		t.Fatalf("models plan after the binary moved: %v, want the existing refusal", err)
	}
}

// AC7 (planning): a binary started by go run lives under the temporary
// directory and disappears, so no hook may point at it.
func TestGrokHookPlanRejectsATemporaryExecutable(t *testing.T) {
	o := grokOptions(t)
	temp := filepath.Join(t.TempDir(), "hive")
	if err := os.WriteFile(temp, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	withHiveExecutable(t, temp)
	_, err := BuildPlan("install", o)
	if err == nil || !strings.Contains(err.Error(), "go build") {
		t.Fatalf("temporary executable error = %v, want a message asking for go build", err)
	}
	absent(t, grokHookPath(o))

	// Removing never needs the binary.
	if _, err := BuildPlan("remove", o); err != nil {
		t.Fatalf("remove with a temporary executable: %v", err)
	}
}

func TestGrokHookPlanForOtherHostsNeedsNoExecutable(t *testing.T) {
	o := setup(t) // codex and claude
	HiveExecutableBroken := func() (string, error) { return "", errors.New("must not be called") }
	old := HiveExecutable
	HiveExecutable = HiveExecutableBroken
	t.Cleanup(func() { HiveExecutable = old })
	p := plan(t, "install", o)
	if p.Executable != "" {
		t.Fatalf("executable = %q on a plan without Grok", p.Executable)
	}
	if strings.Contains(string(encode(p)), `"executable"`) {
		t.Fatal("a plan without Grok must not encode the executable field")
	}
}

// AC7 (apply): a plan edited by hand cannot make the manager write a command
// that points at a path that is not a regular absolute file.
func TestGrokHookApplyRejectsAnUnusableExecutable(t *testing.T) {
	dir, err := testdataDir()
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"relative":    "testdata/hive-test-executable",
		"missing":     filepath.Join(dir, "does-not-exist"),
		"directory":   dir,
		"not clean":   dir + "/../testdata/hive-test-executable",
		"empty":       "",
		"temporary":   filepath.Join(t.TempDir(), "hive"),
		"outside tmp": "", // replaced below by a symlink
	}
	link := filepath.Join(dir, "hive-test-executable-link")
	os.Remove(link)
	if err := os.Symlink(filepath.Join(dir, "hive-test-executable"), link); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(link) })
	cases["symlink"] = link
	delete(cases, "outside tmp")
	if err := os.WriteFile(cases["temporary"], []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, exe := range cases {
		t.Run(name, func(t *testing.T) {
			o := grokOptions(t)
			p := plan(t, "install", o)
			p.Executable = exe
			p.ID = planID(p)
			if _, err := (Engine{}).Apply(p); err == nil {
				t.Fatalf("apply accepted executable %q", exe)
			}
			absent(t, grokHookPath(o))
			if _, err := os.Stat(filepath.Join(o.StateDir, "pending.json")); err == nil {
				t.Fatal("a rejected plan left a pending transaction")
			}
		})
	}
}

// A hand-built plan that points at another regular file regenerates other
// bytes than the ones it froze.
func TestGrokHookApplyRejectsAnExecutableThatDiffersFromTheFrozenPayload(t *testing.T) {
	o := grokOptions(t)
	p := plan(t, "install", o)
	p.Executable = copyExecutable(t, "hive-test-executable-edited")
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("apply accepted an executable that does not match the frozen payload")
	}
	absent(t, grokHookPath(o))
}

func TestGrokHookRemovePlanRejectsAnExecutable(t *testing.T) {
	o := grokOptions(t)
	apply(t, plan(t, "install", o))
	p := plan(t, "remove", o)
	p.Executable = suiteExecutable(t)
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("remove plan with an executable was applied")
	}
	if _, err := os.Stat(grokHookPath(o)); err != nil {
		t.Fatalf("hook was removed by a rejected plan: %v", err)
	}
}

// A plan saved before this field existed has no executable and keeps its ID.
func TestPlanWithoutExecutableKeepsItsID(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	if p.Executable != "" {
		t.Fatal("setup must not include Grok")
	}
	path := filepath.Join(canonicalTemp(t), "plan.json")
	if err := SavePlan(path, p); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadPlan(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != p.ID || loaded.ID != planID(loaded) {
		t.Fatalf("plan ID changed across save and load: %s vs %s", loaded.ID, p.ID)
	}
	apply(t, loaded)

	// With Grok, the executable round-trips and is part of the ID.
	g := grokOptions(t)
	gp := plan(t, "install", g)
	gpath := filepath.Join(canonicalTemp(t), "grok-plan.json")
	if err := SavePlan(gpath, gp); err != nil {
		t.Fatal(err)
	}
	gl, err := LoadPlan(gpath)
	if err != nil || gl.Executable != gp.Executable || gl.ID != gp.ID {
		t.Fatalf("grok plan round trip: %+v, %v", gl.Executable, err)
	}
	other := gp
	other.Executable += "x"
	if planID(other) == gp.ID {
		t.Fatal("the executable is not part of the plan ID")
	}
}

func TestStateRejectsAHandEditedHookRecord(t *testing.T) {
	o := grokOptions(t)
	apply(t, plan(t, "install", o))
	good := hookRecord(t, o)

	mutations := map[string]func(*Record){
		"source":   func(r *Record) { r.Target.Source = "content/skills/x/SKILL.md" },
		"consumer": func(r *Record) { r.Consumers = []Consumer{{"codex", "user", o.Home}} },
		"scope":    func(r *Record) { r.Consumers = []Consumer{{"grok", "project", o.Home}} },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			s := stateFor(t, o)
			r := good
			r.Consumers = append([]Consumer{}, good.Consumers...)
			mutate(&r)
			s.Records[grokHookPath(o)] = r
			if err := normalizeState(&s); err == nil {
				t.Fatal("normalizeState accepted a corrupted hook record")
			}
		})
	}
	t.Run("parent directory", func(t *testing.T) {
		s := stateFor(t, o)
		other := filepath.Join(o.Home, ".grok", "elsewhere", grok.ReadHookName)
		r := good
		r.Target.Path = other
		s.Records = map[string]Record{other: r}
		if err := normalizeState(&s); err == nil {
			t.Fatal("normalizeState accepted a hook outside a hooks directory")
		}
	})
	t.Run("file name", func(t *testing.T) {
		s := stateFor(t, o)
		other := filepath.Join(filepath.Dir(grokHookPath(o)), "other.json")
		r := good
		r.Target.Path = other
		s.Records = map[string]Record{other: r}
		if err := normalizeState(&s); err == nil {
			t.Fatal("normalizeState accepted a hook with another file name")
		}
	})
	t.Run("intact", func(t *testing.T) {
		s := stateFor(t, o)
		if err := normalizeState(&s); err != nil {
			t.Fatal(err)
		}
	})
}

// The constant is generated, never a file of a release.
func TestReleaseRejectsTheGeneratedHookSource(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	r := *p.Release
	r.Files = append(append([]Payload{}, r.Files...), Payload{Path: grok.ReadHookSource, Data: []byte("{}")})
	r.ID = releaseID(r)
	if err := validateRelease(r); err == nil {
		t.Fatal("validateRelease accepted the generated hook source as a release file")
	}
}

func TestStateSourcesIgnoreTheGeneratedHookSource(t *testing.T) {
	o := grokOptions(t)
	apply(t, plan(t, "install", o))
	for _, source := range stateSources(stateFor(t, o)) {
		if source == grok.ReadHookSource {
			t.Fatal("stateSources returned the generated hook source")
		}
	}
}

// RequiredHosts previews every recorded resource with nextRecord on a plan
// that carries no executable; once the hook is recorded that must not fail.
func TestRequiredHostsWorksAfterTheGrokHookIsRecorded(t *testing.T) {
	o := grokOptions(t)
	apply(t, plan(t, "install", o))
	hosts, err := RequiredHosts(o)
	if err != nil {
		t.Fatalf("RequiredHosts after the hook was recorded: %v", err)
	}
	if len(hosts) != 1 || hosts[0] != "grok" {
		t.Fatalf("hosts = %v, want [grok]", hosts)
	}
	if got := apply(t, plan(t, "install", o)); got != "unchanged" {
		t.Fatalf("second install = %q", got)
	}
	// A moved binary changes the hook's bytes, so the preview renders it and
	// runs the shared-resource scope gate instead of skipping it.
	withHiveExecutable(t, copyExecutable(t, "hive-test-executable-required"))
	hosts, err = RequiredHosts(o)
	if err != nil {
		t.Fatalf("RequiredHosts after the binary moved: %v", err)
	}
	if len(hosts) != 1 || hosts[0] != "grok" {
		t.Fatalf("hosts after the binary moved = %v, want [grok]", hosts)
	}
	withHiveExecutable(t, "")
	if _, err := RequiredHosts(o); err == nil {
		t.Fatal("RequiredHosts rendered the hook without an executable")
	}
}

func TestGrokHookExecutableNeedsAnExecuteBit(t *testing.T) {
	dir, err := testdataDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "hive-test-executable-noexec")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(path) })
	if err := checkHookExecutable(path); err == nil || !strings.Contains(err.Error(), "execute") {
		t.Fatalf("file without an execute bit: %v", err)
	}
	withHiveExecutable(t, path)
	if _, err := BuildPlan("install", grokOptions(t)); err == nil || !strings.Contains(err.Error(), "execute permission") {
		t.Fatalf("plan with a non-executable binary: %v", err)
	}
}

func TestSuiteExecutableIsMarkedExecutable(t *testing.T) {
	if err := checkHookExecutable(suiteExecutable(t)); err != nil {
		t.Fatal(err)
	}
}
