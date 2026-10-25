package codex

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

const (
	hdr  = "[features.multi_agent_v2]"
	kMin = "min_wait_timeout_ms = 300000"
	kDef = "default_wait_timeout_ms = 600000"
)

func TestWaitFloorEdits(t *testing.T) {
	cases := []struct {
		name         string
		exists       bool
		in           string
		state        WaitFloorState
		out          string
		lines        []string
		createdTable bool
		createdFile  bool
		separator    bool
	}{
		{name: "no file", in: "", state: WaitFloorAbsent, out: hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{hdr, kMin, kDef}, createdTable: true, createdFile: true},
		{name: "empty existing file", exists: true, in: "", state: WaitFloorAbsent, out: hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{hdr, kMin, kDef}, createdTable: true},
		{name: "file without table", exists: true, in: "model = \"x\"\n", state: WaitFloorAbsent,
			out:   "model = \"x\"\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{hdr, kMin, kDef}, createdTable: true, separator: true},
		{name: "file without final newline", exists: true, in: "model = \"x\"", state: WaitFloorAbsent,
			out:   "model = \"x\"\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{hdr, kMin, kDef}, createdTable: true, separator: true},
		{name: "both values already set", exists: true, in: hdr + "\nmin_wait_timeout_ms = 300000\ndefault_wait_timeout_ms = 600000\n", state: WaitFloorApplied,
			out: hdr + "\nmin_wait_timeout_ms = 300000\ndefault_wait_timeout_ms = 600000\n"},
		{name: "applied with underscores and comment", exists: true, in: hdr + " # hive\nmin_wait_timeout_ms = 300_000 # x\ndefault_wait_timeout_ms = 600000\nmax_wait_timeout_ms = 3600000\n", state: WaitFloorApplied,
			out: hdr + " # hive\nmin_wait_timeout_ms = 300_000 # x\ndefault_wait_timeout_ms = 600000\nmax_wait_timeout_ms = 3600000\n"},
		{name: "BOM directly before the table header", exists: true, in: "\xef\xbb\xbf" + hdr + "\n" + kMin + "\n" + kDef + "\n", state: WaitFloorApplied,
			out: "\xef\xbb\xbf" + hdr + "\n" + kMin + "\n" + kDef + "\n"},
		{name: "BOM directly before the table header with one key", exists: true, in: "\xef\xbb\xbf" + hdr + "\n" + kMin + "\n", state: WaitFloorMissing,
			out:   "\xef\xbb\xbf" + hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{kDef}},
		{name: "table without the keys", exists: true, in: hdr + "\nenabled = true\n\n[tui]\nx = 1\n", state: WaitFloorMissing,
			out:   hdr + "\nenabled = true\n" + kMin + "\n" + kDef + "\n\n[tui]\nx = 1\n",
			lines: []string{kMin, kDef}},
		{name: "table with only min", exists: true, in: hdr + "\nmin_wait_timeout_ms = 300000\n", state: WaitFloorMissing,
			out:   hdr + "\nmin_wait_timeout_ms = 300000\n" + kDef + "\n",
			lines: []string{kDef}},
		{name: "table with only default", exists: true, in: hdr + "\ndefault_wait_timeout_ms = 600000\n", state: WaitFloorMissing,
			out:   hdr + "\ndefault_wait_timeout_ms = 600000\n" + kMin + "\n",
			lines: []string{kMin}},
		{name: "empty table at end without newline", exists: true, in: hdr, state: WaitFloorMissing,
			out:   hdr + "\n" + kMin + "\n" + kDef + "\n",
			lines: []string{kMin, kDef}},
		{name: "table keys followed by comment before next table", exists: true, in: hdr + "\nenabled = true\n# next\n[tui]\n", state: WaitFloorMissing,
			out:   hdr + "\nenabled = true\n" + kMin + "\n" + kDef + "\n# next\n[tui]\n",
			lines: []string{kMin, kDef}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exists := tc.exists
			state, edit, err := WaitFloor([]byte(tc.in), exists || tc.in != "")
			if tc.name == "no file" {
				state, edit, err = WaitFloor(nil, false)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if state != tc.state {
				t.Fatalf("state = %v, want %v", state, tc.state)
			}
			got := string(edit.Apply([]byte(tc.in)))
			if got != tc.out {
				t.Fatalf("output:\n%q\nwant:\n%q", got, tc.out)
			}
			if !reflect.DeepEqual(edit.Record.Lines, tc.lines) {
				t.Fatalf("lines = %q, want %q", edit.Record.Lines, tc.lines)
			}
			if edit.Record.CreatedTable != tc.createdTable || edit.Record.CreatedFile != tc.createdFile || edit.Record.Separator != tc.separator {
				t.Fatalf("record = %+v", edit.Record)
			}
		})
	}
}

func TestWaitFloorRefusals(t *testing.T) {
	big := "# " + strings.Repeat("x", 1024*1024) + "\n"
	cases := []struct {
		name  string
		in    string
		state WaitFloorState
	}{
		{name: "different min", in: hdr + "\nmin_wait_timeout_ms = 10000\n", state: WaitFloorConflict},
		{name: "higher min", in: hdr + "\nmin_wait_timeout_ms = 400000\n", state: WaitFloorConflict},
		{name: "different default", in: hdr + "\ndefault_wait_timeout_ms = 1000\n", state: WaitFloorConflict},
		{name: "max below default", in: hdr + "\nmax_wait_timeout_ms = 599999\n", state: WaitFloorConflict},
		{name: "max at default is fine", in: hdr + "\nmax_wait_timeout_ms = 600000\n", state: WaitFloorMissing},
		{name: "non integer value", in: hdr + "\nmin_wait_timeout_ms = \"300000\"\n", state: WaitFloorUnsupported},
		{name: "duplicate key", in: hdr + "\nmin_wait_timeout_ms = 300000\nmin_wait_timeout_ms = 300000\n", state: WaitFloorUnsupported},
		{name: "duplicate table", in: hdr + "\nenabled = true\n[tui]\n" + hdr + "\n", state: WaitFloorUnsupported},
		{name: "quoted duplicate table", in: hdr + "\n[features.\"multi_agent_v2\"]\n", state: WaitFloorUnsupported},
		{name: "dotted root key", in: "features.multi_agent_v2.min_wait_timeout_ms = 300000\n", state: WaitFloorUnsupported},
		{name: "dotted root table value", in: "features.multi_agent_v2 = { min_wait_timeout_ms = 1 }\n", state: WaitFloorUnsupported},
		{name: "inline features table", in: "features = { multi_agent_v2 = true }\n", state: WaitFloorUnsupported},
		{name: "array of tables", in: "[[features.multi_agent_v2]]\nmin_wait_timeout_ms = 1\n", state: WaitFloorUnsupported},
		{name: "array of features tables", in: "[[features]]\nx = 1\n", state: WaitFloorUnsupported},
		{name: "multi_agent_v2 value under features", in: "[features]\nmulti_agent_v2 = true\n", state: WaitFloorUnsupported},
		{name: "multi_agent_v2 inline under features", in: "[features]\nmulti_agent_v2 = { enabled = true }\n", state: WaitFloorUnsupported},
		{name: "multi_agent_v2 dotted under features", in: "[features]\nmulti_agent_v2.min_wait_timeout_ms = 1\n", state: WaitFloorUnsupported},
		{name: "quoted key under features", in: "[features]\n\"multi_agent_v2\" = true\n", state: WaitFloorUnsupported},
		{name: "subtable of the table", in: "[features.multi_agent_v2.extra]\nx = 1\n", state: WaitFloorUnsupported},
		{name: "triple double quotes", in: "x = \"\"\"\nmulti\n\"\"\"\n", state: WaitFloorUnsupported},
		{name: "triple single quotes", in: "x = '''\nmulti\n'''\n", state: WaitFloorUnsupported},
		{name: "file over 1 MiB", in: big, state: WaitFloorUnsupported},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state, edit, err := WaitFloor([]byte(tc.in), true)
			if state != tc.state {
				t.Fatalf("state = %v, want %v (err %v)", state, tc.state, err)
			}
			refused := tc.state == WaitFloorConflict || tc.state == WaitFloorUnsupported
			var ref *WaitFloorRefusal
			if refused {
				if !errors.As(err, &ref) || ref.State != tc.state || ref.Cause == "" {
					t.Fatalf("want a refusal with a cause, got %v", err)
				}
				if len(edit.Insert) != 0 {
					t.Fatalf("a refusal must not carry an edit")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestWaitFloorRefusalsLeakNoFileText(t *testing.T) {
	secret := "SUPERSECRET-token-value"
	for _, in := range []string{
		"[mcp_servers.x.env]\nTOKEN = \"" + secret + "\"\n" + hdr + "\nmin_wait_timeout_ms = \"" + secret + "\"\n",
		hdr + "\nmin_wait_timeout_ms = 1 # " + secret + "\n",
		"features = { a = \"" + secret + "\" }\n",
		"x = \"\"\"" + secret + "\"\"\"\n",
	} {
		_, _, err := WaitFloor([]byte(in), true)
		if err == nil {
			t.Fatalf("expected a refusal for %q", in)
		}
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaks file text: %v", err)
		}
	}
}

func TestWaitFloorPreservesFormatting(t *testing.T) {
	cases := []struct{ name, in, out string }{
		{name: "CRLF", in: "model = \"x\"\r\n[tui]\r\nx = 1\r\n",
			out: "model = \"x\"\r\n[tui]\r\nx = 1\r\n\r\n" + hdr + "\r\n" + kMin + "\r\n" + kDef + "\r\n"},
		{name: "CRLF without final newline", in: "model = \"x\"\r\nx = 1",
			out: "model = \"x\"\r\nx = 1\r\n\r\n" + hdr + "\r\n" + kMin + "\r\n" + kDef + "\r\n"},
		{name: "CRLF table keys", in: hdr + "\r\nenabled = true\r\n[tui]\r\n",
			out: hdr + "\r\nenabled = true\r\n" + kMin + "\r\n" + kDef + "\r\n[tui]\r\n"},
		{name: "BOM and comments", in: "\xef\xbb\xbf# keep me\n" + hdr + " # trailing\nenabled = true # note\n",
			out: "\xef\xbb\xbf# keep me\n" + hdr + " # trailing\nenabled = true # note\n" + kMin + "\n" + kDef + "\n"},
		{name: "BOM only", in: "\xef\xbb\xbf", out: "\xef\xbb\xbf" + hdr + "\n" + kMin + "\n" + kDef + "\n"},
		{name: "quoted header", in: "[\"features.multi_agent_v2\"]\nx = 1\n",
			out: "[\"features.multi_agent_v2\"]\nx = 1\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n"},
		{name: "header-like text in a comment", in: "# " + hdr + "\nx = 1\n",
			out: "# " + hdr + "\nx = 1\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n"},
		{name: "blank lines at the end", in: "x = 1\n\n\n",
			out: "x = 1\n\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, edit, err := WaitFloor([]byte(tc.in), true)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(edit.Apply([]byte(tc.in))); got != tc.out {
				t.Fatalf("output:\n%q\nwant:\n%q", got, tc.out)
			}
		})
	}
}

func TestUndoWaitFloorRoundTrip(t *testing.T) {
	cases := []struct {
		name, in, out string
		exists        bool
		deleteFile    bool
	}{
		{name: "no file", in: "", deleteFile: true},
		{name: "empty existing file", in: "", exists: true},
		{name: "file with other content", in: "model = \"x\"\n[tui]\nx = 1\n", exists: true},
		{name: "file without final newline", in: "model = \"x\"", exists: true, out: "model = \"x\"\n"},
		{name: "CRLF", in: "model = \"x\"\r\n[tui]\r\nx = 1\r\n", exists: true},
		{name: "BOM", in: "\xef\xbb\xbfmodel = \"x\"\n", exists: true},
		{name: "table without keys", in: hdr + "\nenabled = true\n\n[tui]\nx = 1\n", exists: true},
		{name: "table with only min", in: hdr + "\nmin_wait_timeout_ms = 300000\n", exists: true},
		{name: "blank lines at the end", in: "x = 1\n\n\n", exists: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, edit, err := WaitFloor([]byte(tc.in), tc.exists)
			if err != nil {
				t.Fatal(err)
			}
			applied := edit.Apply([]byte(tc.in))
			res := UndoWaitFloor(applied, edit.Record)
			want := tc.out
			if want == "" {
				want = tc.in
			}
			if string(res.Content) != want {
				t.Fatalf("undo:\n%q\nwant:\n%q", res.Content, want)
			}
			if res.DeleteFile != tc.deleteFile {
				t.Fatalf("DeleteFile = %v, want %v", res.DeleteFile, tc.deleteFile)
			}
			if len(res.Kept) != 0 || res.Skipped != "" {
				t.Fatalf("unexpected leftovers: %+v", res)
			}
			// A second undo finds nothing to remove and changes nothing.
			again := UndoWaitFloor(res.Content, edit.Record)
			if string(again.Content) != string(res.Content) || len(again.Removed) != 0 {
				t.Fatalf("second undo changed the file: %+v", again)
			}
		})
	}
}

func TestUndoWaitFloorKeepsWhatChanged(t *testing.T) {
	created := WaitFloorRecord{Lines: []string{hdr, kMin, kDef}, CreatedTable: true, Separator: true}
	cases := []struct {
		name       string
		in         string
		rec        WaitFloorRecord
		out        string
		removed    []string
		kept       []string
		headerKept bool
		deleteFile bool
		skipped    bool
	}{
		{name: "user changed min", in: "x = 1\n\n" + hdr + "\nmin_wait_timeout_ms = 1000\n" + kDef + "\n", rec: created,
			out: "x = 1\n\n" + hdr + "\nmin_wait_timeout_ms = 1000\n", removed: []string{"default_wait_timeout_ms"}, kept: []string{"min_wait_timeout_ms"}, headerKept: true},
		{name: "user removed a line", in: "x = 1\n\n" + hdr + "\n" + kDef + "\n", rec: created,
			out: "x = 1\n", removed: []string{"default_wait_timeout_ms"}, kept: []string{"min_wait_timeout_ms"}},
		{name: "user added a key to the table", in: "x = 1\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\nenabled = true\n", rec: created,
			out: "x = 1\n\n" + hdr + "\nenabled = true\n", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}, headerKept: true},
		{name: "user comment in the table", in: hdr + "\n# mine\n" + kMin + "\n" + kDef + "\n", rec: WaitFloorRecord{Lines: created.Lines, CreatedTable: true},
			out: hdr + "\n# mine\n", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}, headerKept: true},
		{name: "trailing comment on our line", in: hdr + "\n" + kMin + " # keep?\n" + kDef + "\n", rec: WaitFloorRecord{Lines: created.Lines, CreatedTable: true},
			out: "", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}},
		{name: "table owned by the user keeps its header", in: hdr + "\nenabled = true\n" + kMin + "\n" + kDef + "\n", rec: WaitFloorRecord{Lines: []string{kMin, kDef}},
			out: hdr + "\nenabled = true\n", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}},
		{name: "only the registered key goes", in: hdr + "\n" + kMin + "\n" + kDef + "\n", rec: WaitFloorRecord{Lines: []string{kDef}},
			out: hdr + "\n" + kMin + "\n", removed: []string{"default_wait_timeout_ms"}},
		{name: "Codex rewrote the file without the table", in: "model = \"y\"\n", rec: created,
			out: "model = \"y\"\n", kept: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}, skipped: true},
		{name: "table unreadable now", in: hdr + "\n" + kMin + "\n" + hdr + "\n", rec: created,
			out: hdr + "\n" + kMin + "\n" + hdr + "\n", kept: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}, skipped: true},
		{name: "file created by Hive but the user added content", in: "model = \"y\"\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n",
			rec: WaitFloorRecord{Lines: created.Lines, CreatedTable: true, CreatedFile: true, Separator: true},
			out: "model = \"y\"\n", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}},
		{name: "file created by Hive and left empty", in: hdr + "\n" + kMin + "\n" + kDef + "\n",
			rec: WaitFloorRecord{Lines: created.Lines, CreatedTable: true, CreatedFile: true},
			out: "", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}, deleteFile: true},
		{name: "blank before the header is not ours", in: "x = 1\n\n" + hdr + "\n" + kMin + "\n" + kDef + "\n",
			rec: WaitFloorRecord{Lines: created.Lines, CreatedTable: true},
			out: "x = 1\n\n", removed: []string{"min_wait_timeout_ms", "default_wait_timeout_ms"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := UndoWaitFloor([]byte(tc.in), tc.rec)
			if string(res.Content) != tc.out {
				t.Fatalf("content:\n%q\nwant:\n%q", res.Content, tc.out)
			}
			if !reflect.DeepEqual(res.Removed, tc.removed) || !reflect.DeepEqual(res.Kept, tc.kept) {
				t.Fatalf("removed %q kept %q, want %q and %q", res.Removed, res.Kept, tc.removed, tc.kept)
			}
			if res.HeaderKept != tc.headerKept || res.DeleteFile != tc.deleteFile || (res.Skipped != "") != tc.skipped {
				t.Fatalf("result = %+v", res)
			}
		})
	}
}

func TestWaitFloorRefusesAHeaderLikeLineInsideTheTable(t *testing.T) {
	for name, in := range map[string]string{
		"array element":            hdr + "\nx = [\n[1, 2]\n]\n",
		"array element with comma": hdr + "\nx = [\n[1, 2],\n]\n",
		"unterminated":             hdr + "\n[oops\n",
	} {
		t.Run(name, func(t *testing.T) {
			state, _, err := WaitFloor([]byte(in), true)
			if state != WaitFloorUnsupported || err == nil {
				t.Fatalf("state = %v, err = %v", state, err)
			}
		})
	}
	// A real header after the table still ends it.
	if state, _, err := WaitFloor([]byte(hdr+"\nenabled = true\n[tui]\nx = 1\n"), true); state != WaitFloorMissing || err != nil {
		t.Fatalf("state = %v, err = %v", state, err)
	}
}
