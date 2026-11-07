package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

func piActivationRow(t *testing.T, o management.Options, f *doctorFake) integrationRow {
	t.Helper()
	rows, _ := collectIntegrationRows(o, f.deps())
	return rowByID(t, rows, piSubagentActivationID)
}

func writePiSubagentConfig(t *testing.T, home, body string) string {
	t.Helper()
	dir := filepath.Join(home, ".pi", "agent", "extensions", "subagent")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPiSubagentActivationRowStates(t *testing.T) {
	o, f, home, _ := integrationsFixtureHosts(t, "claude,pi")

	// The fixture's install for Pi already wrote the key and its record.
	r := piActivationRow(t, o, f)
	if r.Found != "eager" || !r.HiveWrote || r.Next != "Nothing to do." {
		t.Fatalf("written by Hive: %+v", r)
	}

	path := writePiSubagentConfig(t, home, `{"secretish": "do-not-show", "toolActivation": "auto"}`)
	r = piActivationRow(t, o, f)
	text := strings.Join(r.detailLines(), "\n")
	if r.Found != "set by you" || r.HiveWrote || !strings.Contains(text, `"auto"`) || strings.Contains(text, "do-not-show") {
		t.Fatalf("user value: found %q, detail %s", r.Found, text)
	}

	writePiSubagentConfig(t, home, `[1]`)
	if r = piActivationRow(t, o, f); r.Found != "not read" || !strings.Contains(r.Next, "yourself") {
		t.Fatalf("unsupported: %+v", r)
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if r = piActivationRow(t, o, f); r.Found != "not set" || !strings.Contains(r.Next, "hive install") {
		t.Fatalf("absent: %+v", r)
	}
}

func TestPiSubagentActivationRowOnlyWithPi(t *testing.T) {
	o, f, _, _ := integrationsFixtureHosts(t, "claude")
	rows, _ := collectIntegrationRows(o, f.deps())
	for _, r := range rows {
		if r.ID == piSubagentActivationID {
			t.Fatal("row shown without Pi registered")
		}
	}
}

// With Hive's state unreadable, an eager key cannot be attributed to the user.
func TestPiSubagentActivationRowWithUnreadableState(t *testing.T) {
	o, f, _, stateDir := integrationsFixtureHosts(t, "claude,pi")
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := piActivationRow(t, o, f)
	if r.Found != "not checked" || r.HiveWrote {
		t.Fatalf("row = %+v", r)
	}
}
