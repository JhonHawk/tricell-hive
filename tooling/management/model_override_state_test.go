package management

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var transactionID = regexp.MustCompile(`"transaction": "[0-9a-f]{32}"`)

var updateStateGolden = flag.Bool("update-state-golden", false, "rewrite testdata/state_without_overrides.golden")

const stateGoldenPath = "testdata/state_without_overrides.golden"

// installedFiveHosts installs the synthetic roles on the five base hosts with
// a product version, so receipts exist, and returns the options and the plan
// that was applied.
func installedFiveHosts(t *testing.T) (Options, Plan) {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
	modelsSource(t, o)
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.0\n")
	p := plan(t, "install", o)
	apply(t, p)
	return o, p
}

// normalized replaces the per-run temporary root so the bytes compare across runs.
func normalized(t *testing.T, o Options, b []byte) []byte {
	t.Helper()
	root := filepath.Dir(o.Home)
	b = bytes.ReplaceAll(b, []byte(root), []byte("<root>"))
	return transactionID.ReplaceAll(b, []byte(`"transaction": "<id>"`))
}

// TestModelOverrideStateWithoutOverridesIsUnchanged characterizes the stored
// state and the plan for an installation that has no override: both encode
// exactly as they did before overrides existed.
func TestModelOverrideStateWithoutOverridesIsUnchanged(t *testing.T) {
	o, p := installedFiveHosts(t)
	stored, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	// The Grok hook holds the checkout path of the test executable and the
	// temporary home, which no normalization can make stable, so the golden
	// file leaves it out. grok_hook_test.go covers it.
	hookDir := filepath.Dir(grokHookPath(o))
	delete(stored.Records, grokHookPath(o))
	var dirs []string
	for _, dir := range stored.CreatedDirs {
		if dir != hookDir {
			dirs = append(dirs, dir)
		}
	}
	stored.CreatedDirs = dirs
	for key, receipt := range stored.Installations {
		var resources []InstalledResource
		for _, res := range receipt.Resources {
			if res.Path != grokHookPath(o) {
				resources = append(resources, res)
			}
		}
		receipt.Resources = resources
		stored.Installations[key] = receipt
	}
	state := encode(stored)
	planned := p
	planned.ID = ""
	planned.Executable = ""
	planned.Changes = nil
	for _, ch := range p.Changes {
		if ch.Target.Kind != "hook" {
			planned.Changes = append(planned.Changes, ch)
		}
	}
	got := append(append([]byte("=== state.json ===\n"), normalized(t, o, state)...), "=== plan ===\n"...)
	got = append(got, normalized(t, o, encode(planned))...)
	if *updateStateGolden {
		if err := os.WriteFile(stateGoldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(stateGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("state or plan encoding changed without overrides; rerun with -update-state-golden only for an intended change")
	}
	if strings.Contains(string(got), "model_overrides") {
		t.Fatal("an installation without overrides encodes a model_overrides field")
	}
}
