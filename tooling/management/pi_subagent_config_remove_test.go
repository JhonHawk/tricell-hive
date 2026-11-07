package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pcInstall(t *testing.T, o Options) {
	t.Helper()
	apply(t, plan(t, "install", o))
}

// pcNotes returns the journal's notes and checks that Apply surfaced each of
// them in its result, which is the transaction ID followed by "; " and the
// warnings.
func pcNotes(t *testing.T, o Options, result string) string {
	t.Helper()
	id, warning, _ := strings.Cut(result, "; ")
	j := pcJournal(t, o, id)
	if j.PiConfig == nil {
		return ""
	}
	for _, n := range j.PiConfig.Notes {
		if !strings.Contains(warning, n) {
			t.Fatalf("Apply did not report the note %q; result %q", n, result)
		}
	}
	return strings.Join(j.PiConfig.Notes, "; ")
}

// AC5: a file and a folder Hive created go, with the key, when nothing else is in them.
func TestRemovePiConfigDeletesWhatHiveCreated(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	apply(t, plan(t, "remove", o))
	absent(t, pcConfig(o))
	absent(t, filepath.Dir(pcConfig(o)))
	if pcState(t, o).PiSubagentConfig != nil {
		t.Fatal("the record survived the removal")
	}
}

// AC5: in an existing file only the key goes; bytes, mode and the backup Hive made.
func TestRemovePiConfigRestoresAnExistingFile(t *testing.T) {
	o := pcOptions(t)
	original := "{\n  \"a\": 1,\n  \"b\": [true]\n}\n"
	put(t, pcConfig(o), original)
	pcInstall(t, o)
	apply(t, plan(t, "remove", o))
	if got := get(t, pcConfig(o)); got != original {
		t.Fatalf("config = %q, want the original", got)
	}
	if m := pcMode(t, pcConfig(o)); m != 0640 {
		t.Fatalf("mode = %v", m)
	}
	absent(t, pcConfig(o)+".hive-backup")
	if pcState(t, o).PiSubagentConfig != nil {
		t.Fatal("the record survived the removal")
	}
}

func TestRemovePiConfigKeepsABackupHiveDidNotCreate(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	put(t, pcConfig(o)+".hive-backup", "oldest")
	pcInstall(t, o)
	apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)+".hive-backup") != "oldest" {
		t.Fatal("a backup Hive did not create was removed")
	}
}

// AC5: a file Hive created that now holds other settings stays, without the key.
func TestRemovePiConfigKeepsACreatedFileWithOtherSettings(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	put(t, pcConfig(o), "{\n  \"theme\": \"dark\",\n  \"toolActivation\": \"eager\"\n}\n")
	id := apply(t, plan(t, "remove", o))
	if got := get(t, pcConfig(o)); got != "{\n  \"theme\": \"dark\"\n}\n" {
		t.Fatalf("config = %q", got)
	}
	if !strings.Contains(pcNotes(t, o, id), "holds other settings") {
		t.Fatalf("notes = %q", pcNotes(t, o, id))
	}
}

// AC5: a value the user changed is left, with the file, the folders and the backup.
func TestRemovePiConfigLeavesAValueTheUserChanged(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	pcInstall(t, o)
	changed := `{"a": 1, "toolActivation": "auto"}`
	put(t, pcConfig(o), changed)
	id := apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)) != changed {
		t.Fatal("a user value was removed")
	}
	if get(t, pcConfig(o)+".hive-backup") != `{"a": 1}` {
		t.Fatal("the backup went although the key stayed")
	}
	if n := pcNotes(t, o, id); !strings.Contains(n, "changed") || !strings.Contains(n, "kept config.json.hive-backup") {
		t.Fatalf("notes = %q", n)
	}
	if pcState(t, o).PiSubagentConfig != nil {
		t.Fatal("Hive still claims a key the user owns")
	}
}

func TestRemovePiConfigLeavesAFileHiveCreatedWhenTheUserChangedTheValue(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	put(t, pcConfig(o), `{"toolActivation": "dynamic"}`)
	apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)) != `{"toolActivation": "dynamic"}` {
		t.Fatal("the file or its value changed")
	}
}

func TestRemovePiConfigWhenTheKeyOrFileIsGone(t *testing.T) {
	t.Run("key gone", func(t *testing.T) {
		o := pcOptions(t)
		put(t, pcConfig(o), `{"a": 1}`)
		pcInstall(t, o)
		put(t, pcConfig(o), `{"a": 2}`)
		id := apply(t, plan(t, "remove", o))
		if get(t, pcConfig(o)) != `{"a": 2}` || pcState(t, o).PiSubagentConfig != nil {
			t.Fatal("the removal touched a file without the key")
		}
		if !strings.Contains(pcNotes(t, o, id), "already gone") {
			t.Fatalf("notes = %q", pcNotes(t, o, id))
		}
	})
	t.Run("file gone", func(t *testing.T) {
		o := pcOptions(t)
		pcInstall(t, o)
		if err := removeAll(pcConfig(o)); err != nil {
			t.Fatal(err)
		}
		id := apply(t, plan(t, "remove", o))
		absent(t, pcConfig(o))
		absent(t, filepath.Dir(pcConfig(o)))
		if n := pcNotes(t, o, id); !strings.Contains(n, "nothing to remove; config.json is gone") || strings.Contains(n, "left toolActivation") || pcState(t, o).PiSubagentConfig != nil {
			t.Fatalf("notes = %q", pcNotes(t, o, id))
		}
	})
}

// A file that became a symbolic link is neither read nor written.
func TestRemovePiConfigLeavesAFileItCannotEditSafely(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	real := filepath.Join(filepath.Dir(o.Home), "real.json")
	put(t, real, "{\"toolActivation\": \"eager\"}\n")
	if err := removeAll(pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	if err := symlink(real, pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	id := apply(t, plan(t, "remove", o))
	if get(t, real) != "{\"toolActivation\": \"eager\"}\n" {
		t.Fatal("the link target was edited")
	}
	if !strings.Contains(pcNotes(t, o, id), "symbolic link") {
		t.Fatalf("notes = %q", pcNotes(t, o, id))
	}
}

func TestRemoveWithoutPiInTheHostsLeavesTheKey(t *testing.T) {
	o := pcOptions(t)
	o.Hosts = []string{"pi", "claude"}
	pcInstall(t, o)
	o.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)) != pcEagerFile || pcState(t, o).PiSubagentConfig == nil {
		t.Fatal("removing another host touched the Pi key")
	}
}

// A record for another Pi home is kept and nothing is touched there.
func TestRemovePiConfigSkipsARecordForAnotherPath(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	elsewhere := filepath.Join(filepath.Dir(o.Home), "other-pi", "config.json")
	st := pcState(t, o)
	st.PiSubagentConfig.Path = elsewhere
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	id := apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)) != pcEagerFile {
		t.Fatal("the current file was touched")
	}
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.Path != elsewhere {
		t.Fatalf("the record was dropped: %#v", rec)
	}
	if !strings.Contains(pcNotes(t, o, id), elsewhere) {
		t.Fatalf("notes = %q", pcNotes(t, o, id))
	}
}

func TestPlanUnchangedKnowsTheRemovalOfThePiKey(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	if unchanged, err := PlanUnchanged(plan(t, "remove", o)); err != nil || unchanged {
		t.Fatalf("removal with a record: unchanged = %v, %v", unchanged, err)
	}
	apply(t, plan(t, "remove", o))
	if unchanged, err := PlanUnchanged(plan(t, "remove", o)); err != nil || !unchanged {
		t.Fatalf("second removal: unchanged = %v, %v", unchanged, err)
	}
}

// AC6: an interrupted removal puts the key back, by key, and the state returns
// to what it was.
func TestRecoverOfAnInterruptedRemovalRestoresTheKey(t *testing.T) {
	for _, stage := range []string{"pi-config-pending", "pi-config-write", "state"} {
		t.Run("existing file/"+stage, func(t *testing.T) {
			o := pcOptions(t)
			put(t, pcConfig(o), "{\n  \"a\": 1\n}\n")
			pcInstall(t, o)
			installed := get(t, pcConfig(o))
			recBefore := pcState(t, o).PiSubagentConfig
			if _, err := wfFail(stage).Apply(plan(t, "remove", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if stage != "pi-config-pending" {
				// pi-subagents rewrites another key while the removal is pending.
				put(t, pcConfig(o), strings.Replace(get(t, pcConfig(o)), `"a": 1`, `"a": 2`, 1))
				installed = strings.Replace(installed, `"a": 1`, `"a": 2`, 1)
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, pcConfig(o)); got != installed {
				t.Fatalf("config = %q, want %q", got, installed)
			}
			absent(t, filepath.Join(o.StateDir, "pending.json"))
			if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.Path != recBefore.Path || !rec.BackupCreated {
				t.Fatalf("record = %#v", rec)
			}
			// The inverse works again once recovered.
			apply(t, plan(t, "remove", o))
			if strings.Contains(get(t, pcConfig(o)), "toolActivation") {
				t.Fatal("the key could not be removed after the recovery")
			}
		})
		t.Run("created file/"+stage, func(t *testing.T) {
			o := pcOptions(t)
			pcInstall(t, o)
			if _, err := wfFail(stage).Apply(plan(t, "remove", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, pcConfig(o)); got != pcEagerFile {
				t.Fatalf("config = %q", got)
			}
			if m := pcMode(t, pcConfig(o)); m != 0600 {
				t.Fatalf("mode = %v", m)
			}
			if rec := pcState(t, o).PiSubagentConfig; rec == nil || !rec.CreatedFile {
				t.Fatalf("record = %#v", rec)
			}
		})
	}
}

func removeAll(path string) error       { return os.RemoveAll(path) }
func symlink(target, link string) error { return os.Symlink(target, link) }
