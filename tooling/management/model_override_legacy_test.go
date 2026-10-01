package management

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// legacyCodexInstructions returns the legacy catalog's Codex AGENTS.md bytes.
func legacyCodexInstructions(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../legacy/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Files []struct {
			Root, Path string
			Data       []byte
		}
	}
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	for _, f := range c.Files {
		if f.Root == "codex" && f.Path == "AGENTS.md" {
			return string(f.Data)
		}
	}
	t.Fatal("the legacy catalog has no Codex AGENTS.md")
	return ""
}

func TestModelsPlanRefusesAPendingLegacyMigrationAndOtherCLIsStillWork(t *testing.T) {
	o, _ := installedFiveHosts(t)
	// A legacy configuration reappears in Codex's instruction file.
	put(t, filepath.Join(o.Home, ".codex", "AGENTS.md"), legacyCodexInstructions(t))
	stateBefore, homeBefore := stateBytes(t, o), snapshotHome(t, o)

	_, err := BuildModelsPlan(o, "codex", map[string]ModelOverride{plainRole: {Effort: "high"}})
	if err == nil {
		t.Fatal("BuildModelsPlan accepted a CLI with a pending legacy migration")
	}
	if !strings.Contains(err.Error(), "hive install") || !strings.Contains(err.Error(), "hive doctor") {
		t.Fatalf("error = %v, want a pointer to hive install or hive doctor", err)
	}
	if stateBytes(t, o) != stateBefore || len(changedPaths(homeBefore, snapshotHome(t, o))) != 0 {
		t.Fatal("a refused plan wrote something")
	}
	// Another CLI is unaffected.
	apply(t, buildModels(t, o, "claude", map[string]ModelOverride{plainRole: {Effort: "max"}}))
	if !strings.Contains(get(t, agentFile(t, o, "claude", plainRole)), `effort: "max"`) {
		t.Fatal("the override on another CLI was not applied")
	}
}
