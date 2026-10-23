package viewer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func frontendSourceDigest(source fs.FS) (string, error) {
	var names []string
	excluded := map[string]bool{"node_modules": true, ".git": true, "coverage": true, "playwright-report": true, "test-results": true, ".DS_Store": true}
	err := fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if name == "." {
			return nil
		}
		if excluded[entry.Name()] || strings.HasPrefix(entry.Name(), ".env") || strings.HasSuffix(entry.Name(), ".local") {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("frontend source symlink: %s", name)
		}
		if !entry.IsDir() {
			names = append(names, name)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	digest := sha256.New()
	for _, name := range names {
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return "", err
		}
		sum := sha256.Sum256(data)
		io.WriteString(digest, name+"\x00"+hex.EncodeToString(sum[:])+"\n")
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func TestFrontendSourceDigest(t *testing.T) {
	files := fstest.MapFS{
		"package.json":              {Data: []byte("{}")},
		"src/main.ts":               {Data: []byte("export {};\n")},
		".env.local":                {Data: []byte("not-a-real-credential")},
		"node_modules/ignored.js":   {Data: []byte("ignored")},
		"config.local/ignored.json": {Data: []byte("ignored")},
	}
	const want = "2842a7e0708469e6a6cb235392caf5458fc5ce77f8a2c69c48980edf292d3475"
	got, err := frontendSourceDigest(files)
	if err != nil || got != want {
		t.Fatalf("digest = %s, %v; want %s", got, err, want)
	}
	files["src/main.ts"] = &fstest.MapFile{Data: []byte("changed")}
	if got, err = frontendSourceDigest(files); err != nil || got == want {
		t.Fatalf("changed source not detected: %s, %v", got, err)
	}
	files["linked.ts"] = &fstest.MapFile{Mode: fs.ModeSymlink, Data: []byte("src/main.ts")}
	if _, err := frontendSourceDigest(files); err == nil {
		t.Error("source symlink accepted")
	}
}

func TestFrontendBundleMatchesSources(t *testing.T) {
	raw, err := staticFiles.ReadFile("static/app/bundle-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest bundleManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	digest, err := frontendSourceDigest(os.DirFS("frontend"))
	if err != nil {
		t.Fatal(err)
	}
	if digest != manifest.SourceDigest {
		t.Fatal("frontend sources are stale; rebuild the bundle")
	}
}
