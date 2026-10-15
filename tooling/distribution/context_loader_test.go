package distribution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenCodeContextLoaderManifest(t *testing.T) {
	root := packageFixture(t)
	m, err := ReadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	// Legacy packages remain verifiable without a plugin payload.
	if err := VerifyIfPackaged(root); err != nil {
		t.Fatal(err)
	}
	m.Version = 2
	save := func() {
		t.Helper()
		b, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ManifestName), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	save()
	if err := VerifyIfPackaged(root); err == nil || !strings.Contains(err.Error(), "context-loader.js") {
		t.Fatalf("missing exact plugin source must be rejected: %v", err)
	}
	source := "integrations/opencode/context-loader.js"
	path := filepath.Join(root, source)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("export default {};"), 0600); err != nil {
		t.Fatal(err)
	}
	m.Files[source] = Digest([]byte("export default {};"))
	save()
	if err := VerifyIfPackaged(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if VerifyIfPackaged(root) == nil {
		t.Fatal("accepted changed plugin")
	}
}
