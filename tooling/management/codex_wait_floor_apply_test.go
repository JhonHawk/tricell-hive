package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func wfState(t *testing.T, o Options) State {
	t.Helper()
	s, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func wfMode(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func TestApplyCodexWaitFloorCreatesTheFileWith0600(t *testing.T) {
	o := wfOptions(t)
	apply(t, plan(t, "install", o))
	if got := get(t, wfConfig(o)); got != wfHeader+"\n"+wfMin+"\n"+wfDefault+"\n" {
		t.Fatalf("config = %q", got)
	}
	if m := wfMode(t, wfConfig(o)); m != 0600 {
		t.Fatalf("mode = %v", m)
	}
	absent(t, wfConfig(o)+".hive-backup")
	rec := wfState(t, o).CodexWaitFloor
	if rec == nil || !rec.CreatedFile || !rec.CreatedTable || rec.BackupCreated || rec.Path != wfConfig(o) || len(rec.Lines) != 3 {
		t.Fatalf("record = %#v", rec)
	}
}

func TestApplyCodexWaitFloorInsertsAndBacksUp(t *testing.T) {
	o := wfOptions(t)
	original := "model = \"x\"\n[mcp_servers.x.env]\nTOKEN = \"SUPERSECRET\"\n"
	put(t, wfConfig(o), original) // put writes 0640
	apply(t, plan(t, "install", o))
	want := original + "\n" + wfHeader + "\n" + wfMin + "\n" + wfDefault + "\n"
	if got := get(t, wfConfig(o)); got != want {
		t.Fatalf("config = %q", got)
	}
	if m := wfMode(t, wfConfig(o)); m != 0640 {
		t.Fatalf("the user's mode changed: %v", m)
	}
	bk := wfConfig(o) + ".hive-backup"
	if get(t, bk) != original || wfMode(t, bk) != 0600 {
		t.Fatalf("backup = %q mode %v", get(t, bk), wfMode(t, bk))
	}
	rec := wfState(t, o).CodexWaitFloor
	if rec == nil || rec.CreatedFile || !rec.CreatedTable || !rec.BackupCreated || !rec.Separator {
		t.Fatalf("record = %#v", rec)
	}
	// Nothing Hive persists next to the state holds the config bytes.
	filepath.WalkDir(o.StateDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if b, _ := os.ReadFile(path); strings.Contains(string(b), "SUPERSECRET") {
				t.Errorf("%s holds config.toml bytes", path)
			}
		}
		return nil
	})
}

func TestApplyCodexWaitFloorKeepsAnExistingBackup(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	put(t, wfConfig(o)+".hive-backup", "oldest\n")
	apply(t, plan(t, "install", o))
	if get(t, wfConfig(o)+".hive-backup") != "oldest\n" {
		t.Fatal("an existing backup was overwritten")
	}
	if rec := wfState(t, o).CodexWaitFloor; rec == nil || rec.BackupCreated {
		t.Fatalf("Hive claims a backup it did not create: %#v", rec)
	}
}

func TestApplyCodexWaitFloorInsertsOnlyMissingKeys(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), wfHeader+"\nenabled = true\n"+wfMin+"\n")
	apply(t, plan(t, "install", o))
	if got := get(t, wfConfig(o)); got != wfHeader+"\nenabled = true\n"+wfMin+"\n"+wfDefault+"\n" {
		t.Fatalf("config = %q", got)
	}
	rec := wfState(t, o).CodexWaitFloor
	if rec == nil || rec.CreatedTable || len(rec.Lines) != 1 || rec.Lines[0] != wfDefault {
		t.Fatalf("record = %#v", rec)
	}
}

func TestApplyCodexWaitFloorAlreadyAppliedWritesAndRecordsNothing(t *testing.T) {
	o := wfOptions(t)
	body := wfHeader + "\n" + wfMin + "\n" + wfDefault + "\n"
	put(t, wfConfig(o), body)
	apply(t, plan(t, "install", o))
	if get(t, wfConfig(o)) != body {
		t.Fatal("the file changed")
	}
	absent(t, wfConfig(o)+".hive-backup")
	if rec := wfState(t, o).CodexWaitFloor; rec != nil {
		t.Fatalf("Hive claims ownership of the user's lines: %#v", rec)
	}
}

func TestApplyRejectsAStaleCodexWaitFloorPlan(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	p := plan(t, "install", o)
	put(t, wfConfig(o), "model = \"y\"\n")
	if _, err := (Engine{}).Apply(p); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("want a stale plan error, got %v", err)
	}
	if strings.Contains(get(t, wfConfig(o)), "multi_agent_v2") {
		t.Fatal("a stale plan wrote the file")
	}
}

func TestApplyCodexWaitFloorFromASavedPlan(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	path := filepath.Join(filepath.Dir(o.Home), "plan.json")
	if err := SavePlan(path, plan(t, "install", o)); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPlan(path)
	if err != nil {
		t.Fatal(err)
	}
	apply(t, p)
	if !strings.Contains(get(t, wfConfig(o)), wfDefault) {
		t.Fatal("the saved plan did not write the floor")
	}
}

func TestCodexWaitFloorRecordSurvivesLaterInstalls(t *testing.T) {
	o := wfOptions(t)
	apply(t, plan(t, "install", o))
	want := wfState(t, o).CodexWaitFloor
	o.CodexWaitFloor = false
	put(t, filepath.Join(o.Source, GlobalSource), "# Rules\nChanged.\n")
	apply(t, plan(t, "install", o))
	got := wfState(t, o).CodexWaitFloor
	if got == nil || got.Path != want.Path || strings.Join(got.Lines, "|") != strings.Join(want.Lines, "|") || !got.CreatedFile {
		t.Fatalf("record after a later install = %#v, want %#v", got, want)
	}
}

func TestPlanUnchangedCountsTheCodexWaitFloor(t *testing.T) {
	o := wfOptions(t)
	apply(t, plan(t, "install", o))
	same1, err := PlanUnchanged(plan(t, "install", o))
	if err != nil || !same1 {
		t.Fatalf("after applying: unchanged = %v, err = %v", same1, err)
	}
	o2 := wfOptions(t)
	o2.CodexWaitFloor = false
	apply(t, plan(t, "install", o2))
	o2.CodexWaitFloor = true
	if u, err := PlanUnchanged(plan(t, "install", o2)); err != nil || u {
		t.Fatalf("a plan that would write: unchanged = %v, err = %v", u, err)
	}
}

func TestStoredCodexWaitFloor(t *testing.T) {
	o := wfOptions(t)
	if rec, err := StoredCodexWaitFloor(o); err != nil || rec != nil {
		t.Fatalf("before any install: %#v, %v", rec, err)
	}
	put(t, wfConfig(o), "model = \"x\"\n")
	apply(t, plan(t, "install", o))
	rec, err := StoredCodexWaitFloor(o)
	if err != nil || rec == nil || !rec.BackupCreated || rec.CreatedFile || rec.Path != wfConfig(o) || len(rec.Lines) != 3 {
		t.Fatalf("after install: %#v, %v", rec, err)
	}
	rec.Lines[0] = "changed"
	if again, _ := StoredCodexWaitFloor(o); again.Lines[0] != wfHeader {
		t.Fatal("the accessor returned shared memory")
	}
	apply(t, plan(t, "remove", wfRemoveOptions(o)))
	if rec, err := StoredCodexWaitFloor(o); err != nil || rec != nil {
		t.Fatalf("after remove: %#v, %v", rec, err)
	}
}

func TestApplyCodexWaitFloorRefusesAFileChangedAfterPlanning(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(s string) error {
		if s == "prepared" {
			put(t, wfConfig(o), "model = \"changed in between\"\n")
		}
		return nil
	}}).Apply(p)
	if err == nil || !strings.Contains(err.Error(), "concurrent change") {
		t.Fatalf("want a concurrent change refusal, got %v", err)
	}
	if strings.Contains(get(t, wfConfig(o)), "multi_agent_v2") {
		t.Fatal("the write went through")
	}
	absent(t, wfConfig(o)+".hive-backup")
}

func TestApplyCodexWaitFloorVerifiesTheFileAfterWriting(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(s string) error {
		if s == "codex-lines-write" {
			put(t, wfConfig(o), "somebody else wrote this\n")
		}
		return nil
	}}).Apply(p)
	if err == nil || !strings.Contains(err.Error(), "post-write verification failed") {
		t.Fatalf("want a post-write verification failure, got %v", err)
	}
	// The interrupted transaction is recoverable and leaves the other writer's file alone.
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if get(t, wfConfig(o)) != "somebody else wrote this\n" {
		t.Fatal("recover touched a file without the lines")
	}
}
