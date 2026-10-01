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
	state, err := os.ReadFile(filepath.Join(o.StateDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	planned := p
	planned.ID = ""
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
