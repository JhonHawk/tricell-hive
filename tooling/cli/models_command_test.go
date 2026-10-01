package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// Tests of `hive models`, `hive models set` and `hive models reset` (#46, T2).
// Every home, state directory and source is synthetic.

var updateModelsGolden = flag.Bool("update-models-golden", false, "rewrite testdata/models_text.golden")

const modelsGoldenPath = "testdata/models_text.golden"

// modelsEnv is a synthetic home with the synthetic agent roles installed on
// five CLIs.
type modelsEnv struct {
	home, stateDir, source string
	hosts                  []string
}

func newModelsEnv(t *testing.T) modelsEnv {
	t.Helper()
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := modelsEnv{
		home:     filepath.Join(base, "home"),
		stateDir: filepath.Join(base, "state"),
		source:   modelsTestSource(t),
		hosts:    []string{"claude", "codex", "grok", "opencode", "pi"},
	}
	if err := os.MkdirAll(e.home, 0700); err != nil {
		t.Fatal(err)
	}
	installDirect(t, e.home, e.stateDir, e.source, e.hosts)
	return e
}

func (e modelsEnv) common() []string {
	return []string{"--home", e.home, "--state-dir", e.stateDir}
}

// TestModelsCommandWithoutOverridesPrintsTheBaseText characterizes the text of
// `hive models` for an installation with no override.
func TestModelsCommandWithoutOverridesPrintsTheBaseText(t *testing.T) {
	e := newModelsEnv(t)
	var out bytes.Buffer
	if err := runModels(e.common(), &out); err != nil {
		t.Fatal(err)
	}
	if *updateModelsGolden {
		if err := os.WriteFile(modelsGoldenPath, out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(modelsGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("hive models output changed without overrides:\n%s\nwant:\n%s", out.String(), want)
	}
}

var _ = management.Options{}
