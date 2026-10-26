package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

func writeCodexConfig(t *testing.T, home, content string) string {
	t.Helper()
	path := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func floorRow(t *testing.T, o management.Options, f *doctorFake) integrationRow {
	t.Helper()
	rows, errText := collectIntegrationRows(o, f.deps())
	if errText != "" {
		t.Fatalf("errText = %q", errText)
	}
	return rowByID(t, rows, "codex-wait-floor")
}

func TestWaitFloorRowShownOnlyWithCodexRegistered(t *testing.T) {
	o, f, _, _ := integrationsFixtureHosts(t, "claude,codex")
	r := floorRow(t, o, f)
	if r.Name != "Codex wait floor" || r.Source != "github.com/openai/codex" {
		t.Fatalf("row = %+v", r)
	}
	o2, f2, _, _ := integrationsFixtureHosts(t, "claude")
	rows, _ := collectIntegrationRows(o2, f2.deps())
	for _, r := range rows {
		if r.ID == "codex-wait-floor" {
			t.Fatalf("row shown without codex registered: %+v", r)
		}
	}
	if strings.Contains(integrationsText(o2, f2.deps()), "Codex wait floor") {
		t.Fatal("the section mentions the wait floor without Codex registered")
	}
}

func TestWaitFloorRowReadsTheConfigAndNeverShowsItsText(t *testing.T) {
	for name, tc := range map[string]struct {
		config, found, evidence string
	}{
		"no file":    {"", "not applied", "not set"},
		"applied":    {floorApplied, "applied", "set"},
		"one key":    {"[features.multi_agent_v2]\nmin_wait_timeout_ms = 300000\n", "partial", "default_wait_timeout_ms"},
		"user value": {floorConflict, "cannot edit", "min_wait_timeout_ms has a different value"},
	} {
		t.Run(name, func(t *testing.T) {
			o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
			if tc.config != "" {
				writeCodexConfig(t, home, tc.config)
			}
			r := floorRow(t, o, f)
			if r.Found != tc.found {
				t.Fatalf("found = %q, want %q (evidence %q)", r.Found, tc.found, r.Evidence)
			}
			text := strings.Join(r.detailLines(), "\n")
			if !strings.Contains(text, tc.evidence) {
				t.Fatalf("detail lacks %q:\n%s", tc.evidence, text)
			}
			if strings.Contains(text, "hunter2") {
				t.Fatalf("file text reached the row:\n%s", text)
			}
		})
	}
}

func TestWaitFloorRowSaysWhetherTheBackupExists(t *testing.T) {
	o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
	path := writeCodexConfig(t, home, floorApplied)
	if text := strings.Join(floorRow(t, o, f).detailLines(), "\n"); !strings.Contains(text, "config.toml.hive-backup: not present") {
		t.Fatalf("detail = %s", text)
	}
	if err := os.WriteFile(path+".hive-backup", []byte("model = \"x\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if text := strings.Join(floorRow(t, o, f).detailLines(), "\n"); !strings.Contains(text, "config.toml.hive-backup: present") {
		t.Fatalf("detail = %s", text)
	}
}

func TestWaitFloorRowSaysWhoWroteTheLines(t *testing.T) {
	o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
	writeCodexConfig(t, home, floorApplied)
	if text := strings.Join(floorRow(t, o, f).detailLines(), "\n"); !strings.Contains(text, "Written by Hive: no") {
		t.Fatalf("detail = %s", text)
	}
	// An install that sets the wait floor leaves Hive's record in its state.
	if err := os.Remove(filepath.Join(home, ".codex", "config.toml")); err != nil {
		t.Fatal(err)
	}
	p, err := management.BuildPlan("install", management.Options{Scope: "user", StateDir: o.StateDir, Source: copyMinimalSource(t), Hosts: []string{"claude", "codex"}, CodexWaitFloor: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
	r := floorRow(t, o, f)
	if text := strings.Join(r.detailLines(), "\n"); r.Found != "applied" || !strings.Contains(text, "Written by Hive: yes") {
		t.Fatalf("found = %q, detail = %s", r.Found, text)
	}
	if !strings.Contains(r.Next, "removes only the lines Hive wrote") {
		t.Fatalf("next with a record = %s", r.Next)
	}
}

func TestWaitFloorRowWarnsWhenTheTableExistsAndCodexIsOlderThan0164(t *testing.T) {
	for name, tc := range map[string]struct {
		config, version string
		warns           bool
	}{
		"table, old codex":  {floorApplied, "codex-cli 0.163.0", true},
		"table, new codex":  {floorApplied, "codex-cli 0.164.0", false},
		"no table, old":     {"model = \"x\"\n", "codex-cli 0.163.0", false},
		"table, no version": {floorApplied, "", false},
	} {
		t.Run(name, func(t *testing.T) {
			o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
			writeCodexConfig(t, home, tc.config)
			if tc.version != "" {
				f.install("codex", tc.version)
			}
			r := floorRow(t, o, f)
			text := strings.Join(r.detailLines(), "\n")
			if warns := strings.Contains(text, "older than 0.164.0"); warns != tc.warns {
				t.Fatalf("warns = %v, want %v:\n%s", warns, tc.warns, text)
			}
		})
	}
}

func TestWaitFloorRowIsNotPartOfOnboardingAndNamesTheCommands(t *testing.T) {
	o, f, _, _ := integrationsFixtureHosts(t, "claude,codex")
	r := floorRow(t, o, f)
	if r.Record != recordNotTracked {
		t.Fatalf("record = %q", r.Record)
	}
	if !strings.Contains(r.Next, "hive install") || !strings.Contains(r.Next, "hive update") {
		t.Fatalf("next = %q", r.Next)
	}
}

func TestWaitFloorNextStepForAppliedWithoutAHiveRecordSaysPlanRemoveLeavesIt(t *testing.T) {
	o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
	writeCodexConfig(t, home, floorApplied) // set by hand: no Hive record
	r := floorRow(t, o, f)
	if r.Found != "applied" {
		t.Fatalf("found = %q", r.Found)
	}
	for _, want := range []string{"not written by Hive", "hive plan remove", "leaves it", "[features.multi_agent_v2]"} {
		if !strings.Contains(r.Next, want) {
			t.Fatalf("next lacks %q: %s", want, r.Next)
		}
	}
	if strings.Contains(r.Next, "removes only the lines Hive wrote") {
		t.Fatalf("next promises a removal that would do nothing: %s", r.Next)
	}
}

func TestWaitFloorRowSaysWhenTheRecordIsForAnotherPathAndDoesNotMixEvidence(t *testing.T) {
	o, f, home, _ := integrationsFixtureHosts(t, "claude,codex")
	p, err := management.BuildPlan("install", management.Options{Scope: "user", StateDir: o.StateDir, Source: copyMinimalSource(t), Hosts: []string{"claude", "codex"}, CodexWaitFloor: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p); err != nil {
		t.Fatal(err)
	}
	// The recorded file now has a backup; CODEX_HOME then moves elsewhere.
	if err := os.WriteFile(filepath.Join(home, ".codex", "config.toml.hive-backup"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	t.Setenv("CODEX_HOME", other)
	r := floorRow(t, o, f)
	text := strings.Join(r.detailLines(), "\n")
	for _, want := range []string{"record is for another path", "hive plan remove will skip it"} {
		if !strings.Contains(text, want) {
			t.Fatalf("detail lacks %q:\n%s", want, text)
		}
	}
	if r.Found != "not applied" || r.HiveWrote {
		t.Fatalf("found = %q hiveWrote = %v", r.Found, r.HiveWrote)
	}
	if strings.Contains(text, "lines in") || strings.Contains(text, "hive-backup: present") {
		t.Fatalf("evidence from the recorded path leaked into the current one:\n%s", text)
	}
}
