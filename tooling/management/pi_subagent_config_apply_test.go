package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pcEagerFile = "{\n  \"toolActivation\": \"eager\"\n}\n"

func pcJournal(t *testing.T, o Options, id string) journal {
	t.Helper()
	var j journal
	if err := decodeFile(filepath.Join(o.StateDir, "transactions", id+".json"), &j); err != nil {
		t.Fatal(err)
	}
	return j
}

// AC1: apply creates the file with 0600, the folders it creates with 0700, and
// records the path, the value, the file and the folders it created.
func TestApplyPiConfigCreatesTheFile(t *testing.T) {
	o := pcOptions(t)
	apply(t, plan(t, "install", o))
	if got := get(t, pcConfig(o)); got != pcEagerFile {
		t.Fatalf("config = %q", got)
	}
	if m := pcMode(t, pcConfig(o)); m != 0600 {
		t.Fatalf("file mode = %v", m)
	}
	if m := pcMode(t, filepath.Dir(pcConfig(o))); m != 0700 {
		t.Fatalf("folder mode = %v", m)
	}
	absent(t, pcConfig(o)+".hive-backup")
	rec := pcState(t, o).PiSubagentConfig
	if rec == nil || rec.Path != pcConfig(o) || rec.Value != `"eager"` || !rec.CreatedFile || rec.BackupCreated {
		t.Fatalf("record = %#v", rec)
	}
	// Only the folder this step creates is its own; the transaction's own folders
	// stay with the transaction's bookkeeping.
	for _, d := range rec.CreatedDirs {
		if !strings.HasPrefix(d, pcHome(o)+string(os.PathSeparator)) {
			t.Fatalf("claims a folder outside the Pi home: %s", d)
		}
	}
	if len(rec.CreatedDirs) == 0 || rec.CreatedDirs[len(rec.CreatedDirs)-1] != filepath.Dir(pcConfig(o)) {
		t.Fatalf("created folders = %v", rec.CreatedDirs)
	}
}

// AC1: a created file in a Pi home that does not exist yet.
func TestApplyPiConfigCreatesAMissingPiHome(t *testing.T) {
	o := pcOptions(t)
	if err := os.RemoveAll(pcHome(o)); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if get(t, pcConfig(o)) != pcEagerFile {
		t.Fatal("no file")
	}
}

// AC2: only the key is added; other keys, values, order and the file's mode stay;
// a backup is made when none exists.
func TestApplyPiConfigInsertsOnlyTheKey(t *testing.T) {
	o := pcOptions(t)
	original := "{\n  \"zeta\": 1,\n  \"alpha\": {\"nested\": [1, 2]},\n  \"mid\": \"x\"\n}\n"
	put(t, pcConfig(o), original) // put writes 0640
	apply(t, plan(t, "install", o))
	want := strings.TrimSuffix(original, "\n}\n") + ",\n  \"toolActivation\": \"eager\"\n}\n"
	if got := get(t, pcConfig(o)); got != want {
		t.Fatalf("config = %q, want %q", got, want)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(get(t, pcConfig(o))), &root); err != nil || len(root) != 4 {
		t.Fatalf("not a valid object with four keys: %v", err)
	}
	if m := pcMode(t, pcConfig(o)); m != 0640 {
		t.Fatalf("the user's mode changed: %v", m)
	}
	bk := pcConfig(o) + ".hive-backup"
	if get(t, bk) != original || pcMode(t, bk) != 0600 {
		t.Fatalf("backup = %q mode %v", get(t, bk), pcMode(t, bk))
	}
	rec := pcState(t, o).PiSubagentConfig
	if rec == nil || rec.CreatedFile || !rec.BackupCreated || len(rec.CreatedDirs) != 0 {
		t.Fatalf("record = %#v", rec)
	}
}

func TestApplyPiConfigKeepsAnExistingBackup(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	put(t, pcConfig(o)+".hive-backup", "oldest")
	apply(t, plan(t, "install", o))
	if get(t, pcConfig(o)+".hive-backup") != "oldest" {
		t.Fatal("an existing backup was overwritten")
	}
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.BackupCreated {
		t.Fatalf("Hive claims a backup it did not create: %#v", rec)
	}
}

func TestApplyPiConfigIntoAnEmptyObject(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), "{}")
	apply(t, plan(t, "install", o))
	if got := get(t, pcConfig(o)); got != pcEagerFile {
		t.Fatalf("config = %q", got)
	}
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.CreatedFile || !rec.BackupCreated {
		t.Fatalf("record = %#v", rec)
	}
}

// AC3: a later install finds the key Hive wrote: applied, no write, record kept.
func TestApplyPiConfigAlreadyAppliedCarriesTheRecordForward(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	apply(t, plan(t, "install", o))
	body, before := get(t, pcConfig(o)), get(t, filepath.Join(o.StateDir, "state.json"))
	recBefore := pcState(t, o).PiSubagentConfig
	p := plan(t, "install", o)
	if st := p.PiSubagentConfig; st == nil || st.Status != PiConfigApplied {
		t.Fatalf("step = %#v", st)
	}
	unchanged, err := PlanUnchanged(p)
	if err != nil || !unchanged {
		t.Fatalf("unchanged = %v, err = %v", unchanged, err)
	}
	if got := apply(t, p); got != "unchanged" {
		t.Fatalf("apply = %q", got)
	}
	if get(t, pcConfig(o)) != body || get(t, filepath.Join(o.StateDir, "state.json")) != before {
		t.Fatal("a repeated install wrote")
	}
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.Path != recBefore.Path || !rec.BackupCreated {
		t.Fatalf("the record was not carried forward: %#v", rec)
	}
}

// AC3: another change in the same transaction still carries the record forward.
func TestApplyPiConfigRecordSurvivesAnotherWrite(t *testing.T) {
	o := pcOptions(t)
	apply(t, plan(t, "install", o))
	put(t, filepath.Join(o.Source, GlobalSource), "# Rules\nChanged rules.\n")
	apply(t, plan(t, "install", o))
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || !rec.CreatedFile {
		t.Fatalf("the record was dropped: %#v", rec)
	}
	if get(t, pcConfig(o)) != pcEagerFile {
		t.Fatal("the file changed")
	}
}

// AC7: an update whose only difference is this step is not "unchanged".
func TestPlanUnchangedKnowsThePiConfigStep(t *testing.T) {
	o := pcOptions(t)
	apply(t, plan(t, "install", o))
	if unchanged, err := PlanUnchanged(plan(t, "install", o)); err != nil || !unchanged {
		t.Fatalf("a fully applied install = %v, %v", unchanged, err)
	}
	// The user takes the key out; everything else is as installed.
	put(t, pcConfig(o), "{}")
	p := plan(t, "install", o)
	if st := p.PiSubagentConfig; st == nil || st.Status != PiConfigWrite {
		t.Fatalf("step = %#v", st)
	}
	unchanged, err := PlanUnchanged(p)
	if err != nil || unchanged {
		t.Fatalf("the step is the only difference: unchanged = %v, err = %v", unchanged, err)
	}
	apply(t, p)
	if !strings.Contains(get(t, pcConfig(o)), `"toolActivation": "eager"`) {
		t.Fatal("the step was not applied")
	}
}

func TestPlanUnchangedRejectsAStalePiConfig(t *testing.T) {
	o := pcOptions(t)
	apply(t, plan(t, "install", o))
	put(t, pcConfig(o), "{}")
	p := plan(t, "install", o)
	put(t, pcConfig(o), `{"changed": true}`)
	if _, err := PlanUnchanged(p); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyRejectsAStalePiConfigPlan(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	p := plan(t, "install", o)
	put(t, pcConfig(o), `{"a": 2}`)
	if _, err := (Engine{}).Apply(p); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("err = %v", err)
	}
	if get(t, pcConfig(o)) != `{"a": 2}` {
		t.Fatal("a stale plan wrote")
	}
}

// AC8: the legacy installer owned forceTopLevelAsync in the same file; the
// migration edits it first, and the step starts from what it leaves.
func pcSeedLegacyAsync(t *testing.T, o Options, config string) {
	t.Helper()
	put(t, filepath.Join(pcHome(o), ".hive-deploy-manifest.json"),
		`{"schemaVersion":1,"scope":"pi","managedFiles":{},"managedConfig":{"extensions/subagent/config.json":{"forceTopLevelAsync":true}}}`)
	put(t, pcConfig(o), config)
}

func TestApplyPiConfigAfterTheLegacyMigrationOfTheSameFile(t *testing.T) {
	for name, tc := range map[string]struct{ legacy, want string }{
		"other keys stay":  {`{"forceTopLevelAsync":true,"theme":"dark"}`, `{"theme":"dark","toolActivation": "eager"}`},
		"only the legacy":  {`{"forceTopLevelAsync":true}`, pcEagerFile},
		"legacy after key": {`{"theme":"dark","forceTopLevelAsync":true}`, `{"theme":"dark","toolActivation": "eager"}`},
	} {
		t.Run(name, func(t *testing.T) {
			o := pcOptions(t)
			pcSeedLegacyAsync(t, o, tc.legacy)
			p := plan(t, "install", o)
			if len(p.Legacy) == 0 {
				t.Fatal("the migration does not touch the file")
			}
			st := p.PiSubagentConfig
			if st == nil || st.Status != PiConfigWrite {
				t.Fatalf("step = %#v", st)
			}
			disk, _ := read(pcConfig(o))
			if st.Expected == finger(disk) {
				t.Fatal("the step was computed from the file before the migration")
			}
			apply(t, p)
			got := get(t, pcConfig(o))
			if got != tc.want {
				t.Fatalf("config = %q, want %q", got, tc.want)
			}
			if strings.Contains(got, "forceTopLevelAsync") {
				t.Fatal("the legacy key survived")
			}
			// The migration's own copy of the file is the one a recovery would use.
			if bk := get(t, pcConfig(o)+".hive-backup"); strings.Contains(bk, "toolActivation") {
				t.Fatalf("backup = %q", bk)
			}
		})
	}
}

func TestApplyPiConfigLegacyMigrationRecoversTheOriginalFile(t *testing.T) {
	for _, stage := range []string{"pi-config-backup", "pi-config-write", "pi-config-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := pcOptions(t)
			legacy := `{"forceTopLevelAsync":true}`
			pcSeedLegacyAsync(t, o, legacy)
			if _, err := wfFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, pcConfig(o)); got != legacy {
				t.Fatalf("config = %q, want the legacy file back", got)
			}
			absent(t, pcConfig(o)+".hive-backup")
		})
	}
}

// AC6: a failed apply is recovered by key, whatever else changed in the file.
func TestRecoverUndoesTheInsertedKey(t *testing.T) {
	for _, stage := range []string{"pi-config-backup", "pi-config-write", "pi-config-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := pcOptions(t)
			original := "{\n  \"a\": 1\n}\n"
			put(t, pcConfig(o), original)
			if _, err := wfFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			want := original
			if stage != "pi-config-backup" {
				// pi-subagents edits another key before the recovery.
				put(t, pcConfig(o), strings.Replace(get(t, pcConfig(o)), `"a": 1`, `"a": 2`, 1))
				want = strings.Replace(original, `"a": 1`, `"a": 2`, 1)
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, pcConfig(o)); got != want {
				t.Fatalf("config = %q, want %q", got, want)
			}
			absent(t, pcConfig(o)+".hive-backup")
			absent(t, filepath.Join(o.StateDir, "pending.json"))
			if pcState(t, o).PiSubagentConfig != nil {
				t.Fatal("the record survived a recovery")
			}
		})
	}
}

func TestRecoverRemovesTheFileAndFoldersHiveCreated(t *testing.T) {
	for _, stage := range []string{"pi-config-write", "pi-config-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := pcOptions(t)
			if _, err := wfFail(stage).Apply(plan(t, "install", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, pcConfig(o))
			absent(t, filepath.Dir(pcConfig(o)))
			if pcState(t, o).PiSubagentConfig != nil {
				t.Fatal("the record survived a recovery")
			}
		})
	}
}

// A user who put the same value in after the plan is not undone by the recovery
// of a transaction that never wrote it.
func TestRecoverLeavesTheBackupWhenTheKeyWasNeverWritten(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	if _, err := wfFail("pi-config-backup").Apply(plan(t, "install", o)); err == nil {
		t.Fatal("expected the failpoint")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	absent(t, pcConfig(o)+".hive-backup")
	if get(t, pcConfig(o)) != `{"a": 1}` {
		t.Fatal("the file changed")
	}
}

// The journal never holds the file's other content.
func TestPiConfigJournalHoldsNoFileBytes(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"secretToken": "SUPERSECRET"}`)
	id := apply(t, plan(t, "install", o))
	filepath.WalkDir(o.StateDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if b, _ := os.ReadFile(path); strings.Contains(string(b), "SUPERSECRET") {
				t.Errorf("%s holds config.json bytes", path)
			}
		}
		return nil
	})
	if j := pcJournal(t, o, id); j.PiConfig == nil || !j.PiConfig.Done {
		t.Fatalf("journal entry = %#v", j.PiConfig)
	}
}

// AC9: a release update neither plans nor applies the step.
func TestReleaseUpdateRejectsThePiConfigStep(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"pi"})
	var core Plan
	if err := json.Unmarshal(p.CorePlan, &core); err != nil {
		t.Fatal(err)
	}
	if core.PiSubagentConfig != nil {
		t.Fatalf("the core plan of a release update carries %#v", core.PiSubagentConfig)
	}
	if err := validatePlan(core, pcState(t, o)); err != nil {
		t.Fatal(err)
	}
	core.PiSubagentConfig = &PiSubagentConfigStep{Status: PiConfigWrite, Path: piConfigPath(core), Key: "toolActivation", Value: `"eager"`}
	core.ID = planID(core)
	if err := validatePlan(core, pcState(t, o)); err == nil || !strings.Contains(err.Error(), "release update") {
		t.Fatalf("validatePlan accepted the step: %v", err)
	}
	p.CorePlan, _ = json.Marshal(core)
	p.CorePlanID = core.ID
	p, _ = FreezeUpdatePlan(p)
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err == nil {
		t.Fatal("the coordinator accepted a release update that carries the step")
	}
	// The coordinator's own check, apart from validatePlan.
	if _, err := updateCore(p); err == nil || !strings.Contains(err.Error(), "providers") {
		t.Fatalf("updateCore = %v", err)
	}
}

func TestReleaseUpdateNeverWritesTheKey(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"pi"})
	put(t, pcConfig(o), "{}")
	if _, err := (Engine{}).ApplyReleaseUpdate(p); err != nil {
		t.Fatal(err)
	}
	if get(t, pcConfig(o)) != "{}" {
		t.Fatal("a release update wrote the key")
	}
}

// AC1 for a package installation: the core plan carries the step, applies it, and
// the providers' unchanged-core check then holds.
func TestPackageInstallationWritesThePiConfigAndProvidersStayUnchanged(t *testing.T) {
	o, p := releaseUpdateFixture(t, []string{"pi"})
	if err := os.Remove(pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pcConfig(o) + ".hive-backup"); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	st := pcState(t, o)
	st.PiSubagentConfig = nil
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	core, err := BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	if core.PiSubagentConfig == nil || core.PiSubagentConfig.Status != PiConfigWrite {
		t.Fatalf("step = %#v", core.PiSubagentConfig)
	}
	parent, err := PreparePackageInstallation(core, p.Package, p.Command, "install-pi-config")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = (Engine{}).ApplyReleaseUpdate(parent); err != nil {
		t.Fatal(err)
	}
	if get(t, pcConfig(o)) != pcEagerFile {
		t.Fatalf("config = %q", get(t, pcConfig(o)))
	}
	core, err = BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	if core.PiSubagentConfig == nil || core.PiSubagentConfig.Status != PiConfigApplied {
		t.Fatalf("step after the install = %#v", core.PiSubagentConfig)
	}
	s, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := packageProvidersPreflight(core, s); err != nil {
		t.Fatalf("providers preflight: %v", err)
	}
	// A rebuilt plan that still has the write to do is not an unchanged core.
	put(t, pcConfig(o), "{}")
	st = pcState(t, o)
	st.PiSubagentConfig = nil
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	core, err = BuildPackageInstallPlan(o, p.Package)
	if err != nil {
		t.Fatal(err)
	}
	s, _, _ = readState(o.StateDir)
	if err := packageProvidersPreflight(core, s); err == nil {
		t.Fatal("providers ran with the key still to write")
	}
}

func TestStoredPiSubagentConfig(t *testing.T) {
	o := pcOptions(t)
	if rec, err := StoredPiSubagentConfig(o); err != nil || rec != nil {
		t.Fatalf("before an install: %#v, %v", rec, err)
	}
	apply(t, plan(t, "install", o))
	rec, err := StoredPiSubagentConfig(o)
	if err != nil || rec == nil || rec.Path != pcConfig(o) || !rec.CreatedFile {
		t.Fatalf("record = %#v, %v", rec, err)
	}
	rec.CreatedFile = false
	if again, _ := StoredPiSubagentConfig(o); !again.CreatedFile {
		t.Fatal("the accessor returns the stored record, not a copy")
	}
}
