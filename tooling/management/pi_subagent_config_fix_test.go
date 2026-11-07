package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/legacy"
)

// Finding 1: recovery restores only what the interrupted removal actually did.
func TestRecoverOfARemovalThatDeletedNothingRestoresNothing(t *testing.T) {
	for _, stage := range []string{"pi-config-pending", "state"} {
		t.Run("file already gone/"+stage, func(t *testing.T) {
			o := pcOptions(t)
			pcInstall(t, o)
			if err := os.RemoveAll(filepath.Dir(pcConfig(o))); err != nil {
				t.Fatal(err)
			}
			if _, err := wfFail(stage).Apply(plan(t, "remove", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, pcConfig(o))
			absent(t, filepath.Dir(pcConfig(o)))
		})
		t.Run("key already gone/"+stage, func(t *testing.T) {
			o := pcOptions(t)
			put(t, pcConfig(o), `{"a": 1}`)
			pcInstall(t, o)
			put(t, pcConfig(o), `{"a": 2}`)
			if _, err := wfFail(stage).Apply(plan(t, "remove", o)); err == nil {
				t.Fatal("expected the failpoint")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, pcConfig(o)); got != `{"a": 2}` {
				t.Fatalf("config = %q: the recovery put back a key the removal never removed", got)
			}
		})
	}
}

// Finding 2: a key the removal could not touch keeps its record.
func TestRemovePiConfigKeepsTheRecordWhenItCannotEditTheFile(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	real := filepath.Join(filepath.Dir(o.Home), "real.json")
	put(t, real, pcEagerFile)
	os.Remove(pcConfig(o))
	if err := os.Symlink(real, pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	id := apply(t, plan(t, "remove", o))
	if rec := pcState(t, o).PiSubagentConfig; rec == nil || rec.Path != pcConfig(o) {
		t.Fatalf("the record was dropped: %#v", rec)
	}
	if !strings.Contains(pcNotes(t, o, id), "symbolic link") {
		t.Fatalf("notes = %q", pcNotes(t, o, id))
	}
	// Once the file is regular again a later removal still undoes the key.
	os.Remove(pcConfig(o))
	put(t, pcConfig(o), pcEagerFile)
	apply(t, plan(t, "remove", o))
	absent(t, pcConfig(o))
	if pcState(t, o).PiSubagentConfig != nil {
		t.Fatal("the record survived the second removal")
	}
}

// Finding 3: recovery gives back the exact bytes the removal changed.
func TestRecoverOfARemovalRestoresTheExactBytes(t *testing.T) {
	weird := "{ \"toolActivation\" :\t\"eager\" }\n"
	t.Run("file Hive created", func(t *testing.T) {
		o := pcOptions(t)
		pcInstall(t, o)
		put(t, pcConfig(o), weird)
		if _, err := wfFail("pi-config-write").Apply(plan(t, "remove", o)); err == nil {
			t.Fatal("expected the failpoint")
		}
		if _, err := (Engine{}).Recover(o.StateDir); err != nil {
			t.Fatal(err)
		}
		if got := get(t, pcConfig(o)); got != weird {
			t.Fatalf("config = %q, want %q", got, weird)
		}
	})
	t.Run("existing file", func(t *testing.T) {
		o := pcOptions(t)
		put(t, pcConfig(o), "{ }\n")
		pcInstall(t, o)
		put(t, pcConfig(o), weird)
		if _, err := wfFail("state").Apply(plan(t, "remove", o)); err == nil {
			t.Fatal("expected the failpoint")
		}
		if _, err := (Engine{}).Recover(o.StateDir); err != nil {
			t.Fatal(err)
		}
		if got := get(t, pcConfig(o)); got != weird {
			t.Fatalf("config = %q, want %q", got, weird)
		}
	})
}

// Finding 4: a backup the user replaced is not Hive's to delete.
func TestRemovePiConfigKeepsABackupThatIsNoLongerHives(t *testing.T) {
	o := pcOptions(t)
	put(t, pcConfig(o), `{"a": 1}`)
	pcInstall(t, o)
	put(t, pcConfig(o)+".hive-backup", "the user's own copy")
	id := apply(t, plan(t, "remove", o))
	if get(t, pcConfig(o)+".hive-backup") != "the user's own copy" {
		t.Fatal("a backup that is not Hive's was deleted")
	}
	if get(t, pcConfig(o)) != `{"a": 1}` {
		t.Fatal("the key was not removed")
	}
	if !strings.Contains(pcNotes(t, o, id), "kept config.json.hive-backup") {
		t.Fatalf("notes = %q", pcNotes(t, o, id))
	}
}

// Finding 5: the step judges the file the migration leaves, not the one it replaces.
func TestPlanPiConfigJudgesThePostMigrationFile(t *testing.T) {
	o := pcOptions(t)
	p := plan(t, "install", o)
	real := filepath.Join(filepath.Dir(o.Home), "real.json")
	put(t, real, "{}")
	if err := os.MkdirAll(filepath.Dir(pcConfig(o)), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(real, pcConfig(o)); err != nil {
		t.Fatal(err)
	}
	p.Legacy = []legacy.Edit{{Path: pcConfig(o), Before: []byte("{}"), After: []byte("{}"), Mode: 0600}}
	p.PiSubagentConfig = nil
	if err := planPiSubagentConfig(&p, State{}); err != nil {
		t.Fatal(err)
	}
	if st := p.PiSubagentConfig; st == nil || st.Status != PiConfigWrite {
		t.Fatalf("step = %#v", st)
	}
}

// Finding 7: a removal whose only open item is a record for another Pi home is
// not "unchanged", so its note reaches the caller.
func TestRemoveWithOnlyARecordForAnotherHomeIsNotUnchanged(t *testing.T) {
	o := pcOptions(t)
	pcInstall(t, o)
	elsewhere := filepath.Join(filepath.Dir(o.Home), "other-pi", "config.json")
	st := pcState(t, o)
	st.PiSubagentConfig.Path = elsewhere
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "remove", o))
	p := plan(t, "remove", o)
	if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
		t.Fatalf("unchanged = %v, err = %v", unchanged, err)
	}
	id := apply(t, p)
	if id == "unchanged" || !strings.Contains(pcNotes(t, o, id), elsewhere) {
		t.Fatalf("apply = %q, notes = %q", id, pcNotes(t, o, id))
	}
}

// Finding 8.
func TestShownPiValueIsExported(t *testing.T) {
	if got := ShownPiValue("a\x07b"); got != "a b" {
		t.Fatalf("got %q", got)
	}
	if got := ShownPiValue(strings.Repeat("é", 100)); len([]rune(got)) != 67 || !strings.HasSuffix(got, "...") {
		t.Fatalf("got %q", got)
	}
}
