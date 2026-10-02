package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

const grokCursorNoticeMarker = "Grok also loads Cursor's copy of the Hive guidance"

func noticePlan(t *testing.T, hosts ...string) management.Plan {
	t.Helper()
	home := t.TempDir()
	return management.Plan{
		Action:   "install",
		Hosts:    hosts,
		StateDir: filepath.Join(home, ".hive-state"),
		Config: target.Config{
			Scope: "user", Home: home, Synthetic: true,
			GrokHome: filepath.Join(home, ".grok"), CursorHome: filepath.Join(home, ".cursor"),
		},
	}
}

func writeGrokConfig(t *testing.T, p management.Plan, content string) {
	t.Helper()
	if err := os.MkdirAll(p.Config.GrokHome, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Config.GrokHome, "config.toml"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func summaryOf(p management.Plan) string {
	var out bytes.Buffer
	showInstallSummary(&out, p, onboardingPreview{}, true, false, true)
	return out.String()
}

func TestGrokCursorNoticePrintedWhenBothHostsAndSettingOn(t *testing.T) {
	t.Setenv("GROK_CURSOR_AGENTS_ENABLED", "")
	os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED")
	for name, config := range map[string]string{"unset": "", "true": "[compat.cursor]\nagents = true\n"} {
		t.Run(name, func(t *testing.T) {
			p := noticePlan(t, "grok", "cursor")
			if config != "" {
				writeGrokConfig(t, p, config)
			}
			got := summaryOf(p)
			want := "Grok also loads Cursor's copy of the Hive guidance (" + filepath.Join(p.Config.CursorHome, "AGENTS.md") + "), about 10,700 tokens per session. To skip it, add [compat.cursor] agents = false to " + filepath.Join(p.Config.GrokHome, "config.toml") + "; your own text in " + filepath.Join(p.Config.CursorHome, "AGENTS.md") + " then stops loading in Grok."
			if !strings.Contains(got, want+"\n") {
				t.Fatalf("missing notice %q in:\n%s", want, got)
			}
			if strings.Index(got, grokCursorNoticeMarker) > strings.Index(got, "Close these CLI sessions") {
				t.Fatalf("notice must precede the closing line:\n%s", got)
			}
		})
	}
}

func TestGrokCursorNoticeOmittedWhenDisabled(t *testing.T) {
	os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED")
	p := noticePlan(t, "grok", "cursor")
	writeGrokConfig(t, p, "[compat.cursor]\nagents = false\n")
	if got := summaryOf(p); strings.Contains(got, grokCursorNoticeMarker) {
		t.Fatalf("notice printed although compat.cursor.agents = false:\n%s", got)
	}

	p = noticePlan(t, "grok", "cursor")
	p.Config.Synthetic = false
	t.Setenv("GROK_CURSOR_AGENTS_ENABLED", "false")
	if got := summaryOf(p); strings.Contains(got, grokCursorNoticeMarker) {
		t.Fatalf("notice printed although GROK_CURSOR_AGENTS_ENABLED=false:\n%s", got)
	}

	// A synthetic home ignores the process environment, so the notice returns.
	p.Config.Synthetic = true
	if got := summaryOf(p); !strings.Contains(got, grokCursorNoticeMarker) {
		t.Fatalf("synthetic home must ignore the environment:\n%s", got)
	}
}

func TestGrokCursorNoticeOmittedWithoutBothHosts(t *testing.T) {
	os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED")
	for _, hosts := range [][]string{{"grok"}, {"cursor"}, {"claude", "codex"}} {
		if got := summaryOf(noticePlan(t, hosts...)); strings.Contains(got, grokCursorNoticeMarker) {
			t.Fatalf("notice printed for hosts %v:\n%s", hosts, got)
		}
	}
}

func TestGrokCursorNoticeOmittedOnMalformedSetting(t *testing.T) {
	os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED")
	p := noticePlan(t, "grok", "cursor")
	writeGrokConfig(t, p, "[compat.cursor]\nagents = maybe\n")
	got := summaryOf(p)
	if strings.Contains(got, grokCursorNoticeMarker) || !strings.Contains(got, "Close these CLI sessions") {
		t.Fatalf("malformed setting must only omit the line:\n%s", got)
	}

	p = noticePlan(t, "grok", "cursor")
	if err := os.MkdirAll(filepath.Join(p.Config.GrokHome, "config.toml"), 0700); err != nil { // a directory is unreadable as a file
		t.Fatal(err)
	}
	if got := summaryOf(p); strings.Contains(got, grokCursorNoticeMarker) {
		t.Fatalf("unreadable config must omit the line:\n%s", got)
	}
}

func TestGrokCursorNoticeCountsRegisteredHosts(t *testing.T) {
	os.Unsetenv("GROK_CURSOR_AGENTS_ENABLED")
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	var out bytes.Buffer
	args := []string{"--home", home, "--hosts", "cursor", "--source", minimalTestSource(t), "--state-dir", stateDir}
	if err := install(args, strings.NewReader("\ny\n"), &out, true); err != nil {
		t.Fatalf("install cursor: %v\n%s", err, out.String())
	}
	p := noticePlan(t, "grok")
	p.StateDir = stateDir
	p.Config.Home = home
	p.Config.GrokHome = filepath.Join(home, ".grok")
	p.Config.CursorHome = filepath.Join(home, ".cursor")
	if got := summaryOf(p); !strings.Contains(got, grokCursorNoticeMarker) {
		t.Fatalf("a registered Cursor install plus a Grok plan must print the notice:\n%s", got)
	}
}
