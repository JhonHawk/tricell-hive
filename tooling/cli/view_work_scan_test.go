package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectWorkDocuments(t *testing.T) {
	write := func(t *testing.T, root, path string, data []byte) {
		t.Helper()
		name := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("ordered original frozen bytes and extras", func(t *testing.T) {
		root := t.TempDir()
		for _, p := range []string{"2026-11-24-export.plan.md", "proposal.md", "specs/a/spec.md", "notes.md"} {
			write(t, root, p, []byte("# "+p+"\n"))
		}
		docs, err := collectWorkDocuments(root, "plan", []string{"notes.md"}, false)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"2026-11-24-export.plan.md", "notes.md"}
		for i, p := range want {
			if docs[i].Path != p || string(docs[i].Bytes) != "# "+p+"\n" || docs[i].SHA256 != fmt.Sprintf("%x", sha256.Sum256(docs[i].Bytes)) || docs[i].Size != int64(len(docs[i].Bytes)) {
				t.Fatalf("incorrect frozen document: %#v", docs[i])
			}
		}
		write(t, root, "2026-11-24-export.plan.md", []byte("edited"))
		if docs[0].Label != "Plan" || string(docs[0].Bytes) != "# 2026-11-24-export.plan.md\n" {
			t.Fatal("source edit changed frozen bytes")
		}
	})
	t.Run("exact count and byte limits", func(t *testing.T) {
		root := t.TempDir()
		paths := make([]string, 8)
		for i := range paths {
			paths[i] = fmt.Sprintf("document-%d.md", i)
			write(t, root, paths[i], []byte(strings.Repeat("x", 256*1024)))
		}
		docs, err := collectWorkDocuments(root, "research", paths, true)
		if err != nil || len(docs) != 8 {
			t.Fatalf("exact permitted limits rejected: %d, %v", len(docs), err)
		}
	})
	t.Run("explicit nonplan selection replaces schema", func(t *testing.T) {
		root := t.TempDir()
		write(t, root, "evidence.md", []byte("evidence"))
		docs, err := collectWorkDocuments(root, "verify", []string{"evidence.md"}, true)
		if err != nil || len(docs) != 1 || docs[0].Label != "evidence.md" {
			t.Fatalf("%#v %v", docs, err)
		}
	})
	for _, scenario := range []string{"escape", "absolute", "symlink file", "symlink directory", "hardlink", "directory", "invalid UTF8", "oversized", "several plans", "duplicate", "missing required", "non Markdown", "empty explicit", "nonplan default", "enumeration budget"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "2026-11-24-export.plan.md", []byte("plan"))
			stage := "plan"
			extra := []string{}
			explicit := false
			switch scenario {
			case "escape":
				extra = []string{"../outside.md"}
			case "absolute":
				extra = []string{filepath.Join(root, "2026-11-24-export.plan.md")}
			case "symlink file":
				if err := os.Symlink("2026-11-24-export.plan.md", filepath.Join(root, "alias.md")); err != nil {
					t.Fatal(err)
				}
				extra = []string{"alias.md"}
			case "symlink directory":
				os.Mkdir(filepath.Join(root, "real"), 0700)
				write(t, root, "real/note.md", []byte("note"))
				os.Symlink("real", filepath.Join(root, "alias"))
				extra = []string{"alias/note.md"}
			case "hardlink":
				if err := os.Link(filepath.Join(root, "2026-11-24-export.plan.md"), filepath.Join(root, "alias.md")); err != nil {
					t.Fatal(err)
				}
			case "directory":
				os.Mkdir(filepath.Join(root, "dir.md"), 0700)
				extra = []string{"dir.md"}
			case "invalid UTF8":
				write(t, root, "2026-11-24-export.plan.md", []byte{0xff})
			case "oversized":
				write(t, root, "2026-11-24-export.plan.md", []byte(strings.Repeat("x", 256*1024+1)))
			case "several plans":
				write(t, root, "2026-11-25-other.plan.md", []byte("other"))
			case "duplicate":
				extra = []string{"./2026-11-24-export.plan.md"}
			case "missing required":
				os.Remove(filepath.Join(root, "2026-11-24-export.plan.md"))
			case "non Markdown":
				write(t, root, "notes.txt", []byte("note"))
				extra = []string{"notes.txt"}
			case "empty explicit":
				explicit = true
			case "nonplan default":
				stage = "verify"
			case "enumeration budget":
				for i := 0; i < 4100; i++ {
					write(t, root, fmt.Sprintf("irrelevant-%04d.txt", i), []byte("x"))
				}
			}
			docs, err := collectWorkDocuments(root, stage, extra, explicit)
			if err == nil || docs != nil {
				t.Fatalf("unsafe selection returned documents: %d, %v", len(docs), err)
			}
		})
	}
}

// The split read/validate boundary makes hostile changes deterministic without
// timing sleeps or a scanner-only test hook. All changes use real filesystem IO.
func TestWorkDocumentScanRejectsChangedReadSet(t *testing.T) {
	for _, change := range []string{"same-size restored-mtime rewrite", "path replacement", "parent symlink replacement", "new hardlink", "second plan added"} {
		t.Run(change, func(t *testing.T) {
			root := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, "notes"), 0700); err != nil {
				t.Fatal(err)
			}
			path := "notes/n.md"
			if change == "second plan added" {
				path = "2026-11-24-export.plan.md"
			}
			full := filepath.Join(root, path)
			if err := os.WriteFile(full, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			anchor, err := os.OpenRoot(root)
			if err != nil {
				t.Fatal(err)
			}
			defer anchor.Close()
			scan := workDocumentScan{root: anchor}
			defer func() {
				for _, node := range scan.nodes {
					if node.file != nil {
						node.file.Close()
					}
				}
			}()
			if _, err := scan.rememberDirectory("."); err != nil {
				t.Fatal(err)
			}
			if _, err := scan.read(path); err != nil {
				t.Fatal(err)
			}
			before, err := os.Stat(full)
			if err != nil {
				t.Fatal(err)
			}
			switch change {
			case "same-size restored-mtime rewrite":
				if err := os.WriteFile(full, []byte("modified"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Chtimes(full, before.ModTime(), before.ModTime()); err != nil {
					t.Fatal(err)
				}
			case "path replacement":
				if err := os.Rename(full, full+".old"); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte("original"), 0600); err != nil {
					t.Fatal(err)
				}
			case "parent symlink replacement":
				if err := os.Rename(filepath.Join(root, "notes"), filepath.Join(root, "moved")); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink("moved", filepath.Join(root, "notes")); err != nil {
					t.Fatal(err)
				}
			case "new hardlink":
				if err := os.Link(full, filepath.Join(root, "linked.md")); err != nil {
					t.Fatal(err)
				}
			case "second plan added":
				if err := os.WriteFile(filepath.Join(root, "2026-11-25-other.plan.md"), []byte("new"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := scan.validate(); err == nil {
				t.Fatal("changed read set was accepted for freezing")
			}
		})
	}
}
