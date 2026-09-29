package agents

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "rewrite testdata/render.golden")

const goldenPath = "testdata/render.golden"

// renderEverything renders every catalogue role for every host and joins the
// outputs in a stable order, so one file characterizes the whole of Render.
func renderEverything(t *testing.T) []byte {
	t.Helper()
	profiles := repositoryProfiles(t)
	sources, err := filepath.Glob(filepath.Join("..", "..", "content", "agents", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 20 {
		t.Fatalf("catalogue has %d roles, want 20", len(sources))
	}
	sort.Strings(sources)
	var out bytes.Buffer
	for _, source := range sources {
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		canonical := strings.TrimPrefix(filepath.ToSlash(source), "../../")
		for _, host := range []string{"claude", "codex", "cursor", "grok", "opencode", "pi"} {
			rendered, err := Render(canonical, data, profiles, host)
			if err != nil {
				t.Fatalf("Render(%s, %s): %v", host, canonical, err)
			}
			fmt.Fprintf(&out, "=== %s %s (%d bytes) ===\n", host, canonical, len(rendered))
			out.Write(rendered)
			out.WriteString("\n=== end ===\n")
		}
	}
	return out.Bytes()
}

func TestRenderGolden(t *testing.T) {
	got := renderEverything(t)
	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Render output differs from %s (%d bytes, want %d); rerun with -update-golden only for an intended change", goldenPath, len(got), len(want))
	}
}
