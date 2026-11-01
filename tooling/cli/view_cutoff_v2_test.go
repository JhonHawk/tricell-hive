package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type viewCutoffEvidence struct {
	Mode     os.FileMode
	Size     int64
	Modified int64
	Hash     string
}

func viewCutoffManifest(t *testing.T, state string) map[string]viewCutoffEvidence {
	t.Helper()
	out := map[string]viewCutoffEvidence{}
	err := filepath.WalkDir(state, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := os.Lstat(p)
		if err != nil {
			return err
		}
		v := viewCutoffEvidence{Mode: i.Mode(), Size: i.Size(), Modified: i.ModTime().UnixNano()}
		if i.Mode().IsRegular() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			v.Hash = fmt.Sprintf("%x", sha256.Sum256(raw))
		}
		rel, _ := filepath.Rel(state, p)
		out[rel] = v
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func TestViewV3CutoffBeforeStartupEffects(t *testing.T) {
	for _, command := range []string{"start", "serve", "stop"} {
		for _, kind := range []string{"legacy", "unknown", "corrupt"} {
			t.Run(command+"/"+kind, func(t *testing.T) {
				state := t.TempDir()
				dir := filepath.Join(state, "viewer")
				if err := os.Mkdir(dir, 0700); err != nil {
					t.Fatal(err)
				}
				os.WriteFile(filepath.Join(state, "manager.json"), []byte("manager state"), 0600)
				os.WriteFile(filepath.Join(dir, "server.json"), []byte("orphan info must remain"), 0600)
				switch kind {
				case "legacy":
					os.Mkdir(filepath.Join(dir, "timeline-api-v1"), 0700)
				case "unknown":
					os.WriteFile(filepath.Join(dir, "unknown"), nil, 0600)
				case "corrupt":
					os.WriteFile(filepath.Join(dir, "storage.json"), []byte(`{"storage_version":99}`), 0600)
				}
				before := viewCutoffManifest(t, state)
				launched := false
				var output bytes.Buffer
				deps := viewDeps{launch: func([]string, string) (int, error) { launched = true; return 0, nil }}
				var err error
				if command == "start" {
					err = viewStart(newViewPaths(state), viewFlags{port: 0}, &output, deps)
				} else if command == "stop" {
					err = viewStop(newViewPaths(state), &output)
				} else {
					err = viewServe(newViewPaths(state), viewFlags{port: 0}, &output)
				}
				if err == nil {
					t.Fatal("invalid state accepted")
				}
				if kind == "legacy" && !strings.Contains(err.Error(), "cutoff_required") {
					t.Fatal(err)
				}
				if launched {
					t.Fatal("launched on rejected state")
				}
				if !reflect.DeepEqual(before, viewCutoffManifest(t, state)) {
					t.Fatal("rejected state mutated")
				}
			})
		}
	}
}
