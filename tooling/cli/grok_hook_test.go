package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/integrations/grok/readhook"
)

func grokHookFixture(t *testing.T) (skills, ref string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	skills = filepath.Join(root, "skills")
	ref = filepath.Join(skills, "a", "references", "r.md")
	if err := os.MkdirAll(filepath.Dir(ref), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ref, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return skills, ref
}

func grokHookEvent(ref string) string {
	b, _ := json.Marshal(map[string]any{"toolName": "read_file", "toolInput": map[string]any{"target_file": ref, "limit": 1}})
	return string(b)
}

func TestGrokHookReadWholeRewrites(t *testing.T) {
	skills, ref := grokHookFixture(t)
	var out bytes.Buffer
	runGrokHook([]string{"read-whole", "--skills", skills}, strings.NewReader(grokHookEvent(ref)), &out, readhook.ReadWhole)
	if !strings.Contains(out.String(), `"updatedInput"`) || strings.Contains(out.String(), `"limit"`) {
		t.Errorf("out = %q", out.String())
	}
}

func TestGrokHookInvalidArgumentsAreSilent(t *testing.T) {
	skills, ref := grokHookFixture(t)
	ev := grokHookEvent(ref)
	for name, args := range map[string][]string{
		"no subcommand":      nil,
		"unknown subcommand": {"nope", "--skills", skills},
		"missing skills":     {"read-whole"},
		"unknown flag":       {"read-whole", "--bogus", "x"},
		"extra positional":   {"read-whole", "--skills", skills, "extra"},
		"relative skills":    {"read-whole", "--skills", "rel/skills"},
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			runGrokHook(args, strings.NewReader(ev), &out, readhook.ReadWhole)
			if out.Len() != 0 {
				t.Errorf("out = %q", out.String())
			}
		})
	}
}

func TestGrokHookInvalidJSONIsSilent(t *testing.T) {
	skills, _ := grokHookFixture(t)
	var out bytes.Buffer
	runGrokHook([]string{"read-whole", "--skills", skills}, strings.NewReader("{broken"), &out, readhook.ReadWhole)
	if out.Len() != 0 {
		t.Errorf("out = %q", out.String())
	}
}

func TestGrokHookPanicIsRecovered(t *testing.T) {
	skills, ref := grokHookFixture(t)
	var out bytes.Buffer
	boom := func([]byte, string) []byte { panic("simulated") }
	runGrokHook([]string{"read-whole", "--skills", skills}, strings.NewReader(grokHookEvent(ref)), &out, boom)
	if out.Len() != 0 {
		t.Errorf("out = %q", out.String())
	}
}

func TestGrokHookRunDispatchReturnsNil(t *testing.T) {
	if err := run([]string{"grok-hook", "nope"}); err != nil {
		t.Errorf("run returned %v; the hook must never fail", err)
	}
}
