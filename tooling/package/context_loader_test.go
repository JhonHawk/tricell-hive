package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/distribution"
	"tricell-hive/tooling/management"
)

func TestOpenCodeContextLoaderOfflinePackage(t *testing.T) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	out := t.TempDir()
	if err := run([]string{"--source", "../..", "--out", out, "--platforms", platform}); err != nil {
		t.Fatal(err)
	}
	matches, err := filepath.Glob(filepath.Join(out, "versions", "*", "*", "*.tar.gz"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("archive: %v %v", matches, err)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	root, err := distribution.Extract(data, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	root, err = target.Canonical(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := distribution.VerifyIfPackaged(root); err != nil {
		t.Fatal(err)
	}
	source := "integrations/opencode/context-loader.js"
	original, err := os.ReadFile(filepath.Join("../..", source))
	if err != nil {
		t.Fatal(err)
	}
	packaged, err := os.ReadFile(filepath.Join(root, source))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(packaged, original) {
		t.Fatal("offline plugin bytes differ")
	}
	manifest, err := distribution.ReadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Version != 2 {
		t.Fatal("new package must use manifest 2")
	}
	for name := range manifest.Files {
		if strings.Contains(name, "context-loader") && name != source {
			t.Fatalf("offline package contains a second editorial or generated plugin resource: %s", name)
		}
	}
	// The plan registers `hive hook`, and the manager refuses the temporary path
	// of a test binary: use the checked-in stand-in the management tests use.
	standIn, err := filepath.EvalSymlinks("../management/testdata/hive-test-executable")
	if err != nil {
		t.Fatal(err)
	}
	standIn, err = filepath.Abs(standIn)
	if err != nil {
		t.Fatal(err)
	}
	previous := management.HiveExecutable
	management.HiveExecutable = func() (string, error) { return standIn, nil }
	t.Cleanup(func() { management.HiveExecutable = previous })
	// Prepare only in a fictional home; never execute the packaged installer.
	p, err := management.BuildPlan("install", management.Options{Scope: "user", Home: filepath.Join(t.TempDir(), "home"), Source: root, Hosts: []string{"opencode"}})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, ch := range p.Changes {
		if ch.Target.Kind == "plugin" {
			count++
			if filepath.Dir(ch.Target.Path) != filepath.Join(p.Config.OpenCodeHome, "hive", "context-loader") {
				t.Fatal("plugin is not delivered as an inactive directory")
			}
			if filepath.Base(ch.Target.Path) != "index.js" && filepath.Base(ch.Target.Path) != "context-loader.json" {
				t.Fatal("unexpected plugin filename")
			}
			if ch.After == nil {
				t.Fatal("missing plugin payload")
			}
			if strings.HasSuffix(ch.Target.Path, ".js") && bytes.Equal(ch.After.Managed, packaged) {
				t.Fatal("installed JS is unrendered editorial template")
			}
		}
	}
	if count != 2 {
		t.Fatalf("extracted package did not prepare inactive pair: %d", count)
	}
}
