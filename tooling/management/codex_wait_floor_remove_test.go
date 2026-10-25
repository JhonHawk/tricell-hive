package management

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func wfInstall(t *testing.T, o Options) {
	t.Helper()
	apply(t, plan(t, "install", o))
}

func wfRemoveOptions(o Options) Options { o.CodexWaitFloor = false; return o }

func TestRemoveCodexWaitFloorRestoresAFileHiveCreated(t *testing.T) {
	o := wfOptions(t)
	wfInstall(t, o)
	apply(t, plan(t, "remove", wfRemoveOptions(o)))
	absent(t, wfConfig(o))
	if wfState(t, o).CodexWaitFloor != nil {
		t.Fatal("the record outlived the removal")
	}
}

func TestRemoveCodexWaitFloorRestoresAnExistingFileByLines(t *testing.T) {
	o := wfOptions(t)
	original := "model = \"x\"\n[tui]\nx = 1\n"
	put(t, wfConfig(o), original)
	wfInstall(t, o)
	// Codex edits another part of the file between install and remove.
	put(t, wfConfig(o), strings.Replace(get(t, wfConfig(o)), "model = \"x\"", "model = \"y\"", 1))
	apply(t, plan(t, "remove", wfRemoveOptions(o)))
	if got := get(t, wfConfig(o)); got != "model = \"y\"\n[tui]\nx = 1\n" {
		t.Fatalf("config = %q", got)
	}
	if m := wfMode(t, wfConfig(o)); m != 0640 {
		t.Fatalf("mode = %v", m)
	}
	absent(t, wfConfig(o)+".hive-backup")
}

func TestRemoveCodexWaitFloorKeepsAValueTheUserChanged(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	wfInstall(t, o)
	put(t, wfConfig(o), strings.Replace(get(t, wfConfig(o)), wfMin, "min_wait_timeout_ms = 120000", 1))
	out := apply(t, plan(t, "remove", wfRemoveOptions(o)))
	got := get(t, wfConfig(o))
	if !strings.Contains(got, "min_wait_timeout_ms = 120000") || !strings.Contains(got, wfHeader) || strings.Contains(got, wfDefault) {
		t.Fatalf("config = %q", got)
	}
	if !strings.Contains(out, "min_wait_timeout_ms") || strings.Contains(out, "120000") {
		t.Fatalf("result must name the key and not the value: %q", out)
	}
	// An incomplete revert keeps the backup.
	if get(t, wfConfig(o)+".hive-backup") != "model = \"x\"\n" {
		t.Fatal("the backup must stay after an incomplete revert")
	}
	if wfState(t, o).CodexWaitFloor != nil {
		t.Fatal("the record must be dropped even when a value was kept")
	}
}

func TestRemoveCodexWaitFloorBackupRules(t *testing.T) {
	t.Run("Hive deletes the backup it created", func(t *testing.T) {
		o := wfOptions(t)
		put(t, wfConfig(o), "model = \"x\"\n")
		wfInstall(t, o)
		apply(t, plan(t, "remove", wfRemoveOptions(o)))
		absent(t, wfConfig(o)+".hive-backup")
	})
	t.Run("Hive keeps a backup it did not create", func(t *testing.T) {
		o := wfOptions(t)
		put(t, wfConfig(o), "model = \"x\"\n")
		put(t, wfConfig(o)+".hive-backup", "oldest\n")
		wfInstall(t, o)
		apply(t, plan(t, "remove", wfRemoveOptions(o)))
		if get(t, wfConfig(o)+".hive-backup") != "oldest\n" {
			t.Fatal("a backup Hive did not create was touched")
		}
	})
}

func TestRemoveCodexWaitFloorWhenTheLinesAreGoneReportsAndContinues(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	wfInstall(t, o)
	put(t, wfConfig(o), "model = \"rewritten by Codex\"\n")
	out, err := (Engine{}).Apply(plan(t, "remove", wfRemoveOptions(o)))
	if err != nil {
		t.Fatal(err)
	}
	if get(t, wfConfig(o)) != "model = \"rewritten by Codex\"\n" {
		t.Fatal("the file changed")
	}
	for _, want := range []string{"min_wait_timeout_ms", "default_wait_timeout_ms", "hive-backup"} {
		if !strings.Contains(out, want) {
			t.Fatalf("the note must name %s: %q", want, out)
		}
	}
	if strings.Contains(out, "300000") || strings.Contains(out, "600000") {
		t.Fatalf("the note must not repeat values: %q", out)
	}
	// Lines that vanished before the remove keep the backup, and the note says so truthfully.
	if get(t, wfConfig(o)+".hive-backup") != "model = \"x\"\n" {
		t.Fatal("the backup must stay")
	}
	if _, err := os.Stat(filepath.Join(o.StateDir, "pending.json")); !os.IsNotExist(err) {
		t.Fatal("pending.json left behind")
	}
}

func TestRemoveCodexWaitFloorNeverClaimsAMissingBackup(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	wfInstall(t, o)
	if err := os.Remove(wfConfig(o) + ".hive-backup"); err != nil {
		t.Fatal(err)
	}
	put(t, wfConfig(o), "rewritten\n")
	out := apply(t, plan(t, "remove", wfRemoveOptions(o)))
	if strings.Contains(out, "hive-backup") {
		t.Fatalf("claims a backup that does not exist: %q", out)
	}
}

func TestRemoveCodexWaitFloorClassifiesWhatTheEditorKept(t *testing.T) {
	cases := map[string]struct{ rewrite func(string) string }{
		"dotted form": {func(string) string {
			return "model = \"x\"\nfeatures.multi_agent_v2.min_wait_timeout_ms = 300000\n"
		}},
		"key moved outside the table": {func(s string) string {
			return strings.Replace(s, wfMin+"\n", "", 1) + "[tui]\n" + wfMin + "\n"
		}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			o := wfOptions(t)
			put(t, wfConfig(o), "model = \"x\"\n")
			wfInstall(t, o)
			put(t, wfConfig(o), tc.rewrite(get(t, wfConfig(o))))
			out := apply(t, plan(t, "remove", wfRemoveOptions(o)))
			if !strings.Contains(out, "min_wait_timeout_ms") || !strings.Contains(out, "hive-backup") {
				t.Fatalf("not reported: %q", out)
			}
			if get(t, wfConfig(o)+".hive-backup") != "model = \"x\"\n" {
				t.Fatal("the backup must stay")
			}
		})
	}
}

func TestRemoveWithoutCodexInTheHostsLeavesTheLines(t *testing.T) {
	o := wfOptions(t)
	o.Hosts = []string{"codex", "claude"}
	wfInstall(t, o)
	r := wfRemoveOptions(o)
	r.Hosts = []string{"claude"}
	apply(t, plan(t, "remove", r))
	if !strings.Contains(get(t, wfConfig(o)), wfDefault) || wfState(t, o).CodexWaitFloor == nil {
		t.Fatal("removing another host touched the Codex lines")
	}
}

func TestRecoverUndoesTheWaitFloorByLines(t *testing.T) {
	for _, stage := range []string{"codex-lines-backup", "codex-lines-write", "codex-lines-after", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := wfOptions(t)
			original := "model = \"x\"\n"
			put(t, wfConfig(o), original)
			p := plan(t, "install", o)
			_, err := (Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}).Apply(p)
			if err == nil {
				t.Fatal("expected the failpoint")
			}
			if stage != "codex-lines-backup" {
				// Codex edits another part of the file before recovery.
				put(t, wfConfig(o), strings.Replace(get(t, wfConfig(o)), "model = \"x\"", "model = \"y\"", 1))
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			want := original
			if stage != "codex-lines-backup" {
				want = "model = \"y\"\n"
			}
			if got := get(t, wfConfig(o)); got != want {
				t.Fatalf("config = %q, want %q", got, want)
			}
			absent(t, wfConfig(o)+".hive-backup")
			absent(t, filepath.Join(o.StateDir, "pending.json"))
			if wfState(t, o).CodexWaitFloor != nil {
				t.Fatal("the record survived a recovery")
			}
		})
	}
}

func TestRecoverDoesNotBlockWhenTheLinesAreGone(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(s string) error {
		if s == "codex-lines-after" {
			return errors.New("injected")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("expected the failpoint")
	}
	put(t, wfConfig(o), "something else entirely\n")
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if get(t, wfConfig(o)) != "something else entirely\n" {
		t.Fatal("recover touched a file without the lines")
	}
	absent(t, filepath.Join(o.StateDir, "pending.json"))
}

func TestRecoverDeletesAFileHiveCreated(t *testing.T) {
	o := wfOptions(t)
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(s string) error {
		if s == "codex-lines-write" {
			return errors.New("injected")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("expected the failpoint")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	absent(t, wfConfig(o))
}

func TestCodexLinesJournalHoldsNoConfigBytes(t *testing.T) {
	typ := reflect.TypeOf(linesEntry{})
	for i := 0; i < typ.NumField(); i++ {
		if f := typ.Field(i); f.Type.Kind() == reflect.Slice && f.Type.Elem().Kind() == reflect.Uint8 {
			t.Fatalf("linesEntry.%s can hold the file's bytes", f.Name)
		}
	}
	for _, typ := range []reflect.Type{reflect.TypeOf(CodexWaitFloorStep{}), reflect.TypeOf(CodexWaitFloorRecord{})} {
		if _, ok := typ.FieldByName("Data"); ok {
			t.Fatalf("%s has a Data field", typ.Name())
		}
	}
	o := wfOptions(t)
	put(t, wfConfig(o), "[mcp_servers.x.env]\nTOKEN = \"SUPERSECRET\"\n")
	wfInstall(t, o)
	matches, _ := filepath.Glob(filepath.Join(o.StateDir, "transactions", "*.json"))
	if len(matches) == 0 {
		t.Fatal("no journal")
	}
	for _, m := range matches {
		b := get(t, m)
		if strings.Contains(b, "SUPERSECRET") || !strings.Contains(b, "CodexLines") {
			t.Fatalf("journal %s: holds the secret or lacks the lines entry", m)
		}
	}
}

func TestRemoveCodexWaitFloorDropsTheHeaderWhenTheUserDeletedTheKeys(t *testing.T) {
	o := wfOptions(t)
	original := "model = \"x\"\n"
	put(t, wfConfig(o), original)
	wfInstall(t, o)
	put(t, wfConfig(o), original+"\n"+wfHeader+"\n")
	out := apply(t, plan(t, "remove", wfRemoveOptions(o)))
	if got := get(t, wfConfig(o)); got != original {
		t.Fatalf("config = %q", got)
	}
	// Hive did not remove the keys itself, so the backup stays and the note says so.
	if get(t, wfConfig(o)+".hive-backup") != original || !strings.Contains(out, "hive-backup") {
		t.Fatalf("backup or note wrong: %q", out)
	}
}

func TestRemoveCodexWaitFloorSkipsARecordForAnotherPath(t *testing.T) {
	o := wfOptions(t)
	put(t, wfConfig(o), "model = \"x\"\n")
	wfInstall(t, o)
	// The record names another Codex folder, as if CODEX_HOME changed since.
	elsewhere := filepath.Join(filepath.Dir(o.Home), "other-codex", "config.toml")
	st := wfState(t, o)
	st.CodexWaitFloor.Path = elsewhere
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), st); err != nil {
		t.Fatal(err)
	}
	before := get(t, wfConfig(o))
	out, err := (Engine{}).Apply(plan(t, "remove", wfRemoveOptions(o)))
	if err != nil {
		t.Fatalf("the rest of the uninstall must go on: %v", err)
	}
	if get(t, wfConfig(o)) != before {
		t.Fatal("the current config.toml was touched")
	}
	if !strings.Contains(out, elsewhere) {
		t.Fatalf("the note must name the recorded path: %q", out)
	}
	rec := wfState(t, o).CodexWaitFloor
	if rec == nil || rec.Path != elsewhere {
		t.Fatalf("the record must stay so a later remove can still undo it: %#v", rec)
	}
	absent(t, filepath.Join(o.StateDir, "pending.json"))
	if u, err := PlanUnchanged(plan(t, "remove", wfRemoveOptions(o))); err != nil || !u {
		t.Fatalf("a second remove has nothing to do: unchanged = %v, err = %v", u, err)
	}
}
