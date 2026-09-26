package main

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"reflect"
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
