package pi

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClassifySubagentConfig(t *testing.T) {
	cases := []struct {
		name   string
		data   string
		exists bool
		want   SubagentConfigStatus
		value  string
	}{
		{"absent", "", false, ConfigAbsent, ""},
		{"empty object", "{}", true, KeyMissing, ""},
		{"other keys", `{"forceTopLevelAsync": true, "x": 1}`, true, KeyMissing, ""},
		{"auto", `{"toolActivation": "auto"}`, true, KeyPresent, `"auto"`},
		{"eager", `{"a": 1, "toolActivation": "eager"}`, true, KeyPresent, `"eager"`},
		{"null", `{"toolActivation": null}`, true, KeyPresent, `null`},
		{"duplicate keys", `{"a": 1, "a": 2}`, true, ConfigUnsupported, ""},
		{"invalid json", `{"a": `, true, ConfigUnsupported, ""},
		{"array", `[1, 2]`, true, ConfigUnsupported, ""},
		{"trailing data", `{} {}`, true, ConfigUnsupported, ""},
		{"empty file", ``, true, ConfigUnsupported, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, value := ClassifySubagentConfig([]byte(c.data), c.exists)
			if got != c.want || value != c.value {
				t.Fatalf("got %v %q, want %v %q", got, value, c.want, c.value)
			}
		})
	}
}

func TestInsertAndRemoveToolActivationKeepOtherMembers(t *testing.T) {
	cases := []struct {
		name, in, inserted string
	}{
		{"new file", "", "{\n  \"toolActivation\": \"eager\"\n}\n"},
		{"empty object", "{}\n", "{\n  \"toolActivation\": \"eager\"\n}\n"},
		{"indented", "{\n  \"forceTopLevelAsync\": true,\n  \"b\": [1, 2]\n}\n", "{\n  \"forceTopLevelAsync\": true,\n  \"b\": [1, 2],\n  \"toolActivation\": \"eager\"\n}\n"},
		{"compact", `{"z":1,"a":{"k":"v"}}`, `{"z":1,"a":{"k":"v"},"toolActivation": "eager"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := InsertToolActivation([]byte(c.in))
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != c.inserted {
				t.Fatalf("insert:\n%s\nwant:\n%s", out, c.inserted)
			}
			if st, v := ClassifySubagentConfig(out, true); st != KeyPresent || v != `"eager"` {
				t.Fatalf("inserted file classifies as %v %q", st, v)
			}
			back, removed, err := RemoveToolActivation(out)
			if err != nil || !removed {
				t.Fatalf("remove: %v %v", removed, err)
			}
			if c.in != "" && strings.TrimSpace(c.in) != "{}" && string(back) != c.in {
				t.Fatalf("remove did not restore the original:\n%s\nwant:\n%s", back, c.in)
			}
			if st, _ := ClassifySubagentConfig(back, true); st != KeyMissing {
				t.Fatalf("removed file classifies as %v", st)
			}
		})
	}
}

func TestRemoveToolActivationFirstAndOnlyMember(t *testing.T) {
	out, removed, err := RemoveToolActivation([]byte(`{"toolActivation": "eager", "a": 1}`))
	if err != nil || !removed || string(out) != `{"a": 1}` {
		t.Fatalf("got %q %v %v", out, removed, err)
	}
	out, removed, err = RemoveToolActivation([]byte("{\n  \"toolActivation\": \"eager\"\n}\n"))
	if err != nil || !removed {
		t.Fatalf("got %v %v", removed, err)
	}
	if !SubagentConfigEmpty(out) {
		t.Fatalf("only member removed, file not empty: %q", out)
	}
	if SubagentConfigEmpty([]byte(`{"a": 1}`)) {
		t.Fatal("non-empty object reported empty")
	}
	if _, removed, _ := RemoveToolActivation([]byte(`{"a": 1}`)); removed {
		t.Fatal("removed a missing key")
	}
}

func TestInsertToolActivationRefusesUnsupported(t *testing.T) {
	for _, in := range []string{`[1]`, `{"a":1,"a":2}`, `{"toolActivation":"auto"}`, `{`} {
		if _, err := InsertToolActivation([]byte(in)); err == nil {
			t.Fatalf("insert accepted %q", in)
		}
	}
}

func TestCheckSubagentConfigTarget(t *testing.T) {
	home := t.TempDir()
	if exists, err := CheckSubagentConfigTarget(home); err != nil || exists {
		t.Fatalf("missing folders: %v %v", exists, err)
	}
	dir := filepath.Join(home, "extensions", "subagent")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if exists, err := CheckSubagentConfigTarget(home); err != nil || !exists {
		t.Fatalf("regular file: %v %v", exists, err)
	}

	big := filepath.Join(t.TempDir(), "big")
	bigDir := filepath.Join(big, "extensions", "subagent")
	_ = os.MkdirAll(bigDir, 0700)
	_ = os.WriteFile(filepath.Join(bigDir, "config.json"), make([]byte, MaxSubagentConfigBytes+1), 0600)
	assertRefused(t, big)

	link := t.TempDir()
	_ = os.MkdirAll(filepath.Join(link, "extensions", "subagent"), 0700)
	_ = os.Symlink(path, filepath.Join(link, "extensions", "subagent", "config.json"))
	assertRefused(t, link)

	linkedDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(linkedDir, "extensions"), 0700)
	_ = os.Symlink(dir, filepath.Join(linkedDir, "extensions", "subagent"))
	assertRefused(t, linkedDir)

	readOnly := t.TempDir()
	if err := os.Chmod(readOnly, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(readOnly, 0700) })
	assertRefused(t, readOnly)
}

func assertRefused(t *testing.T, home string) {
	t.Helper()
	_, err := CheckSubagentConfigTarget(home)
	var r *SubagentConfigRefusal
	if !errors.As(err, &r) {
		t.Fatalf("%s: want a refusal, got %v", home, err)
	}
}
