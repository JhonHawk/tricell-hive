package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestArchiveDeterministicAndPreservesExecutable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "hive-test")
	os.MkdirAll(root, 0755)
	os.WriteFile(filepath.Join(root, "install.sh"), []byte("#!/bin/sh\n"), 0755)
	out := t.TempDir()
	a := filepath.Join(out, "one.tar.gz")
	b := filepath.Join(out, "two.tar.gz")
	if err := writeArchive(root, a); err != nil {
		t.Fatal(err)
	}
	if err := writeArchive(root, b); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(a)
	second, _ := os.ReadFile(b)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("archive is not reproducible")
	}
	f, err := os.Open(a)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	if _, err := tr.Next(); err != nil {
		t.Fatal(err)
	}
	h, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if h.Name != "hive-test/install.sh" || h.Mode != 0755 {
		t.Fatalf("bad archive header: %+v", h)
	}
	if writeArchive(root, a) == nil {
		t.Fatal("overwrote existing archive")
	}
}

func TestReserveVersionRefusesAnExistingVersion(t *testing.T) {
	out := t.TempDir()
	dir, err := reserveVersion(out, "1.2.3")
	if err != nil {
		t.Fatal(err)
	}
	published := filepath.Join(dir, "darwin-arm64", "hive")
	if err := os.MkdirAll(filepath.Dir(published), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(published, []byte("published"), 0755); err != nil {
		t.Fatal(err)
	}
	// A rebuild of the same label must fail before any artifact is rewritten,
	// or a new binary would sit beside the old package, checksum and index.
	if _, err := reserveVersion(out, "1.2.3"); err == nil {
		t.Fatal("reserved an existing version")
	}
	if data, _ := os.ReadFile(published); string(data) != "published" {
		t.Fatalf("existing artifact changed: %q", data)
	}
}

func TestCopyRejectsLinks(t *testing.T) {
	src := t.TempDir()
	os.Symlink("missing", filepath.Join(src, "link"))
	if copyTree(src, filepath.Join(t.TempDir(), "dest")) == nil {
		t.Fatal("accepted linked source")
	}
}

func TestUnsupportedMacOSIntelPlatformRejected(t *testing.T) {
	if err := run([]string{"--out", t.TempDir(), "--platforms", "darwin/amd64"}); err == nil {
		t.Fatal("accepted unsupported macOS Intel platform")
	}
}

func TestArchiveLabelUsesProductVersion(t *testing.T) {
	if got, want := archiveLabel("1.2.3-rc.1", "linux", "arm64"), "hive-1.2.3-rc.1-linux-arm64"; got != want {
		t.Fatalf("archiveLabel() = %q, want %q", got, want)
	}
}

func TestFrozenInputsIncludeRuntimeAndLocalDependencies(t *testing.T) {
	inputs, err := frozenInputs(filepath.Clean("../.."), []string{"darwin/arm64"})
	if err != nil {
		t.Fatal(err)
	}
	set := map[string]bool{}
	for _, input := range inputs {
		set[input] = true
	}
	for _, required := range []string{"go.mod", "install.sh", "content", "integrations", "tooling/cli", "tooling/version", "tooling/distribution"} {
		if !set[required] {
			t.Fatalf("missing frozen input %q", required)
		}
	}
}

func TestPackageArchiveContainsLicenseAndNotices(t *testing.T) {
	platform := runtime.GOOS + "/" + runtime.GOARCH
	supported := false
	for _, known := range platforms {
		if known == platform {
			supported = true
		}
	}
	if !supported {
		t.Skipf("host platform %s is not a package target", platform)
	}
	out := t.TempDir()
	if err := run([]string{"--source", filepath.Clean("../.."), "--out", out, "--platforms", platform}); err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(out, "versions", "*", "*", "*.tar.gz"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("want exactly one archive, got %v (err %v)", archives, err)
	}
	f, err := os.Open(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	top := strings.TrimSuffix(filepath.Base(archives[0]), ".tar.gz")
	found := map[string]int64{}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading archive %s: %v", archives[0], err)
		}
		found[h.Name] = h.Size
	}
	for _, name := range []string{"LICENSE", "THIRD_PARTY_NOTICES.md"} {
		size, ok := found[top+"/"+name]
		if !ok || size == 0 {
			t.Errorf("archive lacks a non-empty %s/%s", top, name)
		}
	}
}
