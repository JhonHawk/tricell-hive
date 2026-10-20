package viewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type bundleManifest struct {
	Version      int          `json:"version"`
	SourceDigest string       `json:"sourceDigest"`
	Entrypoint   string       `json:"entrypoint"`
	Files        []bundleFile `json:"files"`
}

type bundleFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	MIME   string `json:"mime"`
}

type frontendAsset struct {
	data []byte
	mime string
}

type frontendBundle struct {
	entrypoint string
	files      map[string]frontendAsset
}

func bundleMIME(name string) string {
	switch path.Ext(name) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	}
	return ""
}

// loadFrontendBundle accepts only the build's declared, verified files. It does
// not expose a directory server or require the frontend toolchain at runtime.
func loadFrontendBundle(source fs.FS) (*frontendBundle, error) {
	raw, err := fs.ReadFile(source, "bundle-manifest.json")
	if err != nil {
		return nil, fmt.Errorf("frontend manifest: %w", err)
	}
	var manifest bundleManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("frontend manifest: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF || manifest.Version != 1 || manifest.Entrypoint != "index.html" || !keyPattern.MatchString(manifest.SourceDigest) {
		return nil, errors.New("invalid frontend manifest")
	}
	bundle := &frontendBundle{entrypoint: manifest.Entrypoint, files: make(map[string]frontendAsset)}
	previous := ""
	for _, file := range manifest.Files {
		if !fs.ValidPath(file.Path) || strings.Contains(file.Path, "\\") || file.Path <= previous || !keyPattern.MatchString(file.SHA256) || file.MIME == "" || file.MIME != bundleMIME(file.Path) {
			return nil, errors.New("invalid frontend asset declaration")
		}
		previous = file.Path
		data, err := fs.ReadFile(source, file.Path)
		if err != nil {
			return nil, fmt.Errorf("frontend asset %s: %w", file.Path, err)
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != file.SHA256 {
			return nil, fmt.Errorf("frontend asset hash mismatch: %s", file.Path)
		}
		bundle.files[file.Path] = frontendAsset{data: data, mime: file.MIME}
	}
	if _, ok := bundle.files[bundle.entrypoint]; !ok {
		return nil, errors.New("frontend entrypoint is missing")
	}
	err = fs.WalkDir(source, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || name == "bundle-manifest.json" {
			return nil
		}
		if _, ok := bundle.files[name]; !ok {
			return fmt.Errorf("undeclared frontend asset: %s", name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (s *Server) frontendAsset(w http.ResponseWriter, r *http.Request, name string) {
	asset, ok := s.frontend.files[name]
	if !ok {
		httpError(w, http.StatusNotFound, "not found")
		return
	}
	if !allow(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	w.Header().Set("Content-Type", asset.mime)
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(asset.data))
}
