package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMemoryEnvironmentReplacesInheritedStoreAndCloud(t *testing.T) {
	got := isolatedMemoryEnvironment([]string{"PATH=/bin", "ENGRAM_DATA_DIR=/real", "ENGRAM_PROJECT=real", "ENGRAM_CLOUD_AUTOSYNC=1", "ENGRAM_CLOUD_TOKEN=secret", "ENGRAM_CLOUD_SERVER=https://cloud", "ENGRAM_DATABASE_URL=postgres://real", "OTHER=value"}, "/private/test")
	joined := strings.Join(got, "\n")
	for _, bad := range []string{"/real", "secret", "https://cloud", "postgres:", "ENGRAM_PROJECT=", "ENGRAM_CLOUD_AUTOSYNC=1"} {
		if strings.Contains(joined, bad) {
			t.Fatalf("inherited memory setting survived: %s", bad)
		}
	}
	for _, want := range []string{"PATH=/bin", "OTHER=value", "ENGRAM_DATA_DIR=/private/test", "ENGRAM_CLOUD_AUTOSYNC=0"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing setting %s", want)
		}
	}
}

func TestMemoryPreflightAndOwnedCleanup(t *testing.T) {
	bin := t.TempDir()
	script := "#!/bin/sh\n[ \"$ENGRAM_CLOUD_AUTOSYNC\" = 0 ] || exit 1\n[ -z \"$ENGRAM_PROJECT\" ] || exit 1\n[ -d \"$ENGRAM_DATA_DIR\" ] || exit 1\necho 'No projects found.'\n"
	if err := os.WriteFile(filepath.Join(bin, "engram"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out := t.TempDir()
	m, _, err := prepareMemoryIsolation(out, append(os.Environ(), "ENGRAM_PROJECT=real", "ENGRAM_CLOUD_AUTOSYNC=1"))
	if err != nil || !m.report.PreflightVerifiedEmpty {
		t.Fatalf("preflight failed: %v", err)
	}
	if _, _, err := prepareMemoryIsolation(out, os.Environ()); err == nil {
		t.Fatal("reused a prior store")
	}
	if got := m.cleanup(); got != "isolated_database_removed" {
		t.Fatal(got)
	}
	if _, err := os.Stat(m.dataDir); !os.IsNotExist(err) {
		t.Fatal("store survived cleanup")
	}
}

func TestMemoryCleanupPreservesReplacement(t *testing.T) {
	out := t.TempDir()
	dir := filepath.Join(out, "engram-data")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(dir)
	m := &memoryIsolation{outputDir: out, dataDir: dir, marker: filepath.Join(dir, ".pilot-owner"), ownedInfo: info}
	if err := os.Rename(dir, filepath.Join(out, "old")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(m.marker, []byte(memoryOwnerMarker), 0600); err != nil {
		t.Fatal(err)
	}
	if got := m.cleanup(); got != "cleanup_refused_replaced_directory" {
		t.Fatal(got)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatal("replacement removed")
	}
}

func TestMemoryPreflightRejectsNonemptyOrAmbiguousOutput(t *testing.T) {
	for _, output := range []string{"", "No projects found.\nreal-project", "Projects: real"} {
		if emptyEngramProjects([]byte(output)) {
			t.Fatalf("accepted ambiguous preflight: %q", output)
		}
	}
}

func TestEngramNativeHTTPStoreAndCleanup(t *testing.T) {
	if _, err := exec.LookPath("engram"); err != nil {
		t.Skip("native Engram CLI unavailable")
	}
	m, env, err := prepareMemoryIsolation(t.TempDir(), os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.cleanup() })
	env, err = m.startHTTP(env, m.outputDir)
	if err != nil || !m.report.HTTPStoreVerified {
		t.Fatalf("HTTP isolation: %v", err)
	}
	cmd := exec.Command("engram", "save", "Isolated fixture marker", "Synthetic native isolation test", "--project", "hive-isolation-test")
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		t.Fatal("native isolated write failed")
	}
	m.stopHTTP()
	m.captureStats()
	if !m.report.RuntimeStatsVerified || m.report.RuntimeObservations != 1 {
		t.Fatalf("write did not reach isolated store: %+v", m.report)
	}
	if got := m.cleanup(); got != "isolated_database_removed" {
		t.Fatal(got)
	}
}

func TestInitializationErrorIsNotAnObservedMemoryRead(t *testing.T) {
	ok := true
	events := []traceEvent{
		{ID: "m1", Tool: "mem_context", Kind: "tool", Memory: true, Success: &ok},
		{ID: "m1", Kind: "tool_result", Text: "gentle-engram could not initialize the Engram memory provider: ownership mismatch", Success: &ok},
	}
	if observedMemoryRead(events) {
		t.Fatal("provider initialization failure counted as memory read")
	}
	events = append(events, traceEvent{ID: "m2", Tool: "mem_search", Kind: "tool", Memory: true, Success: &ok}, traceEvent{ID: "m2", Kind: "tool_result", Text: "No memories found", Success: &ok})
	if !observedMemoryRead(events) {
		t.Fatal("successful empty search not observed")
	}
}

func TestCodexMCPExplicitlyReceivesPrivateStore(t *testing.T) {
	args := codexMemoryOverrides([]string{"ENGRAM_DATA_DIR=/private/run", "ENGRAM_URL=http://127.0.0.1:12345", "ENGRAM_PORT=12345", "ENGRAM_CLOUD_TOKEN=must-not-forward"}, "/fixture")
	joined := strings.Join(args, "\n")
	if !strings.Contains(joined, `plugins."engram@engram".mcp_servers.engram.enabled=false`) || strings.Contains(joined, `plugins."engram@engram".enabled=false`) {
		t.Fatal("duplicate plugin MCP must be disabled without disabling runtime hooks")
	}
	for _, expected := range []string{`mcp_servers.engram.env.ENGRAM_DATA_DIR="/private/run"`, `mcp_servers.engram.env.ENGRAM_URL="http://127.0.0.1:12345"`, `mcp_servers.engram.env.ENGRAM_CLOUD_AUTOSYNC="0"`, `mcp_servers.engram.env.ENGRAM_CLOUD_TOKEN=""`, `mcp_servers.engram.cwd="/fixture"`} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("missing native override %s", expected)
		}
	}
	if strings.Contains(joined, "must-not-forward") {
		t.Fatal("cloud credential forwarded")
	}
}

func TestProjectDetectionAndStructuredErrorsAreNotStoreReads(t *testing.T) {
	ok := true
	for _, tc := range []struct{ tool, text string }{
		{"mem_current_project", `{"project":"fixture","project_source":"git_root"}`},
		{"mem_search", `{"error_code":"unknown_project","message":"not found"}`},
		{"mem_context", ""},
	} {
		events := []traceEvent{
			{ID: "m", Tool: tc.tool, Kind: "tool", Memory: true, Success: &ok},
			{ID: "m", Kind: "tool_result", Text: tc.text, Success: &ok},
		}
		if observedMemoryRead(events) {
			t.Fatalf("%s falsely counted as successful store read", tc.tool)
		}
	}
}

func TestCodexMemoryOverridesBelongToExecSubcommand(t *testing.T) {
	for _, prefix := range [][]string{{"exec", "--json", "-"}, {"-a", "never", "exec", "--json", "-"}} {
		args, err := codexMemoryArgs(prefix, []string{"ENGRAM_DATA_DIR=/private/store"}, "/fixture")
		if err != nil {
			t.Fatal(err)
		}
		execSeen := false
		for _, arg := range args {
			if arg == "exec" {
				execSeen = true
			}
			if strings.HasPrefix(arg, "mcp_servers.") && !execSeen {
				t.Fatal("memory configuration belongs after exec, not to the root CLI")
			}
		}
		if args[len(args)-1] != "-" || !strings.Contains(strings.Join(args, " "), `ENGRAM_DATA_DIR="/private/store"`) {
			t.Fatal("lost stdin prompt or memory configuration")
		}
	}
	if _, err := codexMemoryArgs([]string{"mcp", "get", "engram"}, nil, "/fixture"); err == nil {
		t.Fatal("accepted non-exec invocation")
	}
}
