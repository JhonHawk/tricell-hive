package viewer

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// The embedded module and stylesheet must be exact copies of the ones the
// offline report uses (go:embed cannot reach outside the package), AC15.
func TestAssetsEmbeddedCopiesMatchTheSkillAssetsByteForByte(t *testing.T) {
	for _, name := range []string{"walkthrough.mjs", "walkthrough.css"} {
		want, err := os.ReadFile(filepath.Join("..", "..", "content", "skills", "flow-walkthrough", "assets", name))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join("static", name))
		if err != nil {
			t.Fatalf("missing copy: %v", err)
		}
		if len(want) == 0 {
			t.Fatalf("%s: source asset is empty", name)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("tooling/viewer/static/%s differs from content/skills/flow-walkthrough/assets/%s; copy it again", name, name)
		}
		embedded, err := staticFiles.ReadFile("static/" + name)
		if err != nil || !bytes.Equal(embedded, want) {
			t.Errorf("embedded %s differs from the skill asset (err=%v)", name, err)
		}
	}
}
