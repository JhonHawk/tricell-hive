package viewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
)

func frontendFixture(t *testing.T) (fstest.MapFS, bundleManifest) {
	t.Helper()
	files := fstest.MapFS{
		"assets/main.js":    {Data: []byte("export {};")},
		"assets/styles.css": {Data: []byte("body{color:white}")},
		"index.html":        {Data: []byte(`<script type="module" src="/assets/app/assets/main.js"></script>`)},
	}
	m := bundleManifest{Version: 1, SourceDigest: strings.Repeat("a", 64), Entrypoint: "index.html"}
	for _, name := range []string{"assets/main.js", "assets/styles.css", "index.html"} {
		sum := sha256.Sum256(files[name].Data)
		m.Files = append(m.Files, bundleFile{Path: name, SHA256: hex.EncodeToString(sum[:]), MIME: bundleMIME(name)})
	}
	return files, m
}

func TestFrontendBundleRejectsInvalidOrStaleOutput(t *testing.T) {
	cases := []struct {
		name   string
		change func(fstest.MapFS, *bundleManifest)
	}{
		{"version", func(_ fstest.MapFS, m *bundleManifest) { m.Version = 2 }},
		{"source digest", func(_ fstest.MapFS, m *bundleManifest) { m.SourceDigest = "invalid" }},
		{"entrypoint", func(_ fstest.MapFS, m *bundleManifest) { m.Entrypoint = "other.html" }},
		{"missing entrypoint", func(_ fstest.MapFS, m *bundleManifest) { m.Files = m.Files[:2] }},
		{"duplicate file", func(_ fstest.MapFS, m *bundleManifest) { m.Files = append(m.Files, m.Files[2]) }},
		{"unsorted files", func(_ fstest.MapFS, m *bundleManifest) { m.Files[0], m.Files[1] = m.Files[1], m.Files[0] }},
		{"traversal", func(_ fstest.MapFS, m *bundleManifest) { m.Files[0].Path = "../main.js" }},
		{"backslash", func(_ fstest.MapFS, m *bundleManifest) { m.Files[0].Path = `assets\main.js` }},
		{"MIME", func(_ fstest.MapFS, m *bundleManifest) { m.Files[0].MIME = "text/plain" }},
		{"hash", func(_ fstest.MapFS, m *bundleManifest) { m.Files[0].SHA256 = strings.Repeat("0", 64) }},
		{"modified asset", func(files fstest.MapFS, _ *bundleManifest) {
			files["assets/main.js"] = &fstest.MapFile{Data: []byte("changed")}
		}},
		{"missing asset", func(files fstest.MapFS, _ *bundleManifest) { delete(files, "assets/main.js") }},
		{"extra file", func(files fstest.MapFS, _ *bundleManifest) {
			files["extra.js"] = &fstest.MapFile{Data: []byte("extra")}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files, manifest := frontendFixture(t)
			tc.change(files, &manifest)
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			files["bundle-manifest.json"] = &fstest.MapFile{Data: raw}
			if _, err := loadFrontendBundle(files); err == nil {
				t.Fatal("invalid bundle accepted")
			}
		})
	}
	files, manifest := frontendFixture(t)
	raw, _ := json.Marshal(manifest)
	files["bundle-manifest.json"] = &fstest.MapFile{Data: raw}
	if _, err := loadFrontendBundle(files); err != nil {
		t.Fatalf("valid bundle: %v", err)
	}
	files["bundle-manifest.json"] = &fstest.MapFile{Data: append(raw, []byte(` {"extra":true}`)...)}
	if _, err := loadFrontendBundle(files); err == nil {
		t.Error("trailing JSON accepted")
	}
	delete(files, "bundle-manifest.json")
	if _, err := loadFrontendBundle(files); err == nil {
		t.Error("missing manifest accepted")
	}
}

func TestServerServesReactBundle(t *testing.T) {
	s := newSrvEnv(t)
	raw, err := staticFiles.ReadFile("static/app/bundle-manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Entrypoint string `json:"entrypoint"`
		Files      []struct {
			Path string `json:"path"`
			MIME string `json:"mime"`
		} `json:"files"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, file := range manifest.Files {
		want, err := staticFiles.ReadFile("static/app/" + file.Path)
		if err != nil {
			t.Fatal(err)
		}
		for _, method := range []string{http.MethodGet, http.MethodHead} {
			w := s.do(method, "/assets/app/"+file.Path, reqOpts{})
			if w.Code != http.StatusOK || w.Header().Get("Content-Type") != file.MIME {
				t.Errorf("%s %s: status %d, MIME %q", method, file.Path, w.Code, w.Header().Get("Content-Type"))
			}
			if method == http.MethodGet && !bytes.Equal(w.Body.Bytes(), want) {
				t.Errorf("wrong bytes for %s", file.Path)
			}
			if method == http.MethodHead && w.Body.Len() != 0 {
				t.Errorf("HEAD body for %s", file.Path)
			}
			if w.Header().Get("Content-Security-Policy") != contentSecurityPolicy {
				t.Error("bundle changed CSP")
			}
		}
	}
	entry, _ := staticFiles.ReadFile("static/app/" + manifest.Entrypoint)
	if got := s.do(http.MethodGet, "/", reqOpts{}); !bytes.Equal(got.Body.Bytes(), entry) {
		t.Error("root does not serve the React entrypoint")
	}
	for _, target := range []string{"/assets/app/bundle-manifest.json", "/assets/app/missing.js", "/assets/app/../server.go", "/assets/app/assets/source.js.map", "/assets/app//index.html"} {
		if got := s.do(http.MethodGet, target, reqOpts{}); got.Code != http.StatusNotFound {
			t.Errorf("undeclared %s: %d", target, got.Code)
		}
	}
	if strings.Contains(string(entry), "localhost:") {
		t.Error("production entrypoint references a development server")
	}
}
