package version

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValid(t *testing.T) {
	for _, value := range []string{"dev", "0.0.0", "1.2.3", "12.34.56-rc.1", "1.2.3-alpha-9"} {
		if !Valid(value) {
			t.Fatalf("Valid(%q) = false", value)
		}
	}
	for _, value := range []string{"", "v1.2.3", "1.2", "1.2.3+build", "01.2.3", "1.2.3-01", "1.2.3-", "1.2.3/escape"} {
		if Valid(value) {
			t.Fatalf("Valid(%q) = true", value)
		}
	}
}

func TestReadSourceFile(t *testing.T) {
	root := t.TempDir()
	value, present, err := ReadSourceFile(root)
	if err != nil || present || value != "" {
		t.Fatalf("absent: ReadSourceFile() = %q, %v, %v", value, present, err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("  1.2.3  \n"), 0600); err != nil {
		t.Fatal(err)
	}
	value, present, err = ReadSourceFile(root)
	if err != nil || !present || value != "1.2.3" {
		t.Fatalf("present with whitespace: ReadSourceFile() = %q, %v, %v", value, present, err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("   \n"), 0600); err != nil {
		t.Fatal(err)
	}
	value, present, err = ReadSourceFile(root)
	if err != nil || !present || value != "" {
		t.Fatalf("empty: ReadSourceFile() = %q, %v, %v", value, present, err)
	}
	unreadable := t.TempDir()
	if err := os.Mkdir(filepath.Join(unreadable, "VERSION"), 0700); err != nil {
		t.Fatal(err)
	}
	if value, present, err = ReadSourceFile(unreadable); err == nil || present {
		t.Fatalf("unreadable: ReadSourceFile() = %q, %v, %v, want a read error", value, present, err)
	}
}

func TestReadSource(t *testing.T) {
	root := t.TempDir()
	got, err := ReadSource(root)
	if err != nil || got != "dev" {
		t.Fatalf("missing ReadSource() = %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("1.2.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = ReadSource(root)
	if err != nil || got != "1.2.3" {
		t.Fatalf("ReadSource() = %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte("bad\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadSource(root); err == nil {
		t.Fatal("accepted invalid VERSION")
	}
}
