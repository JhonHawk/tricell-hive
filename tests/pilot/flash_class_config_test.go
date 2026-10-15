package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Regression: rejecting supported compatible entries prevents the installed
// native configuration (legacy Engram plus a V2 codegraph server) from running.
func TestFlashClassOpenCodeCompatibleCombinedConfiguration(t *testing.T) {
	real, shadow := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "opencode.json"), []byte(`{"mcp":{"engram":{"type":"local","command":["engram","mcp","--tools=agent"],"enabled":true},"servers":{"codegraph":{"type":"local","command":["codegraph"]}}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowOpenCodeConfig(shadow, real, "/private/fixture/engram-data"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(shadow, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg map[string]any
	if json.Unmarshal(b, &cfg) != nil {
		t.Fatal("shadow config malformed")
	}
	servers := object(object(cfg["mcp"])["servers"])
	if len(servers) != 1 || object(servers["engram"])["type"] != "local" || len(list(object(servers["engram"])["command"])) != 3 {
		t.Fatal("compatible launch command was not converted to isolated V2 configuration")
	}
}

// Regression: identical supported entries have no unresolved precedence and
// must not acquire a new prohibition solely because both forms are present.
func TestFlashClassOpenCodeAcceptsEquivalentDefinitions(t *testing.T) {
	real, shadow := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "opencode.json"), []byte(`{"mcp":{"engram":{"type":"local","command":["engram","mcp"],"enabled":true},"servers":{"engram":{"type":"local","command":["engram","mcp"],"disabled":false}}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowOpenCodeConfig(shadow, real, "/private/fixture/engram-data"); err != nil {
		t.Fatal(err)
	}
}

// Regression: accepting an ambiguous server map or a disabled/remote server would
// produce a shadow run with no verified local memory route.
func TestFlashClassOpenCodeRejectsUnusableMemoryDefinitions(t *testing.T) {
	for name, body := range map[string]string{
		"ambiguous":          `{"mcp":{"engram":{"type":"local","command":["engram","mcp"]},"servers":{"engram":{"type":"local","command":["other","mcp"]}}}}`,
		"disabled":           `{"mcp":{"servers":{"engram":{"type":"local","command":["engram","mcp"],"disabled":true}}}}`,
		"remote":             `{"mcp":{"servers":{"engram":{"type":"remote","command":["engram","mcp"]}}}}`,
		"empty executable":   `{"mcp":{"servers":{"engram":{"type":"local","command":[""]}}}}`,
		"malformed":          `{"private":"FAKE_PRIVATE_MARKER",`,
		"orphan comma":       `{"mcp":{"servers":{,}}}`,
		"wrong command type": `{"mcp":{"servers":{"engram":{"type":"local","command":"FAKE_PRIVATE_MARKER"}}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			real, shadow := t.TempDir(), t.TempDir()
			if err := os.WriteFile(filepath.Join(real, "opencode.json"), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			err := writeShadowOpenCodeConfig(shadow, real, "/private/fixture/engram-data")
			if err == nil {
				t.Fatal("unusable native memory definition accepted")
			}
			if strings.Contains(err.Error(), "FAKE_PRIVATE_MARKER") {
				t.Fatal("private input reached diagnostic")
			}
			if _, err := os.Stat(filepath.Join(shadow, "opencode.json")); !os.IsNotExist(err) {
				t.Fatal("failed validation wrote shadow configuration")
			}
		})
	}
}

// Regression: selecting a file arbitrarily when both exist silently compares
// a different native configuration than the one the host may load.
func TestFlashClassOpenCodeRefusesAmbiguousNativeConfiguration(t *testing.T) {
	real, shadow := t.TempDir(), t.TempDir()
	for _, name := range []string{"opencode.json", "opencode.jsonc"} {
		if err := os.WriteFile(filepath.Join(real, name), []byte(`{"mcp":{"servers":{"engram":{"type":"local","command":["engram","mcp"]}}}}`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeShadowOpenCodeConfig(shadow, real, "/private/fixture/engram-data"); err == nil {
		t.Fatal("ambiguous native configuration accepted")
	}
	if _, err := os.Stat(filepath.Join(shadow, "opencode.json")); !os.IsNotExist(err) {
		t.Fatal("ambiguous configuration wrote a shadow file")
	}
}

// Regression: reading only opencode.json prevents native JSONC configurations
// from supplying the same launch definition; stripping comments inside quoted
// strings corrupts executable paths and arguments.
func TestFlashClassOpenCodeReadsJSONCWithoutCopyingPrivateFields(t *testing.T) {
	real, shadow := t.TempDir(), t.TempDir()
	body := `{
  // Native V2 server map.
  "mcp": {"servers": {"engram": {
    "type": "local",
    "command": ["/fixture/engram", "mcp", "https://example.test/a//b",],
    "environment": {"PRIVATE": "FAKE_PRIVATE_MARKER"},
  },},},
  "plugins": ["private-plugin"],
}`
	if err := os.WriteFile(filepath.Join(real, "opencode.jsonc"), []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	if err := writeShadowOpenCodeConfig(shadow, real, "/private/fixture/engram-data"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(shadow, "opencode.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "FAKE_PRIVATE_MARKER") || strings.Contains(string(got), "private-plugin") {
		t.Fatal("unrequested native settings copied")
	}
	if !strings.Contains(string(got), "https://example.test/a//b") {
		t.Fatal("quoted command argument corrupted")
	}
}
