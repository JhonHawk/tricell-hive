package agents

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

const contentGoldenPath = "testdata/render_content.golden"

// renderContent renders every real role under content/agents with the real
// profiles for the six hosts. Unlike TestRenderGolden it follows content, so it
// characterizes Render only until a content edit lands; regenerate it with
// -update-golden when a role or profile changes on purpose.
func renderContent(t *testing.T) []byte {
	t.Helper()
	root := filepath.Join("..", "..")
	profiles, err := os.ReadFile(filepath.Join(root, ProfilesSource))
	if err != nil {
		t.Fatal(err)
	}
	sources, err := filepath.Glob(filepath.Join(root, "content", "agents", "*", "*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Fatal("no roles under content/agents")
	}
	sort.Strings(sources)
	var out bytes.Buffer
	for _, path := range sources {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		canonical := filepath.ToSlash(rel)
		for _, host := range []string{"claude", "codex", "cursor", "grok", "opencode", "pi"} {
			rendered, err := Render(canonical, data, profiles, host, "/skills", nil)
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

func TestRenderContentGolden(t *testing.T) {
	got := renderContent(t)
	if *updateGolden {
		if err := os.WriteFile(contentGoldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(contentGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Render output differs from %s (%d bytes, want %d); rerun with -update-golden only for an intended change", contentGoldenPath, len(got), len(want))
	}
}
