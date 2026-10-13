package main

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// Regression: lexical confinement followed by RemoveAll can traverse replaced
// ancestors and remove an unrelated directory outside the owned shadow home.
func TestFlashClassCleanupRejectsReplacedAncestors(t *testing.T) {
	for _, component := range []string{".local", ".local/share", ".local/share/opencode", ".cache", "home", "home directory", "leaf directory"} {
		t.Run(component, func(t *testing.T) {
			fakeOpenCodeOnPath(t)
			home := t.TempDir()
			g := &guidanceVariant{shadowHome: home, host: "opencode", integration: "opencode", shadowDataDir: filepath.Join(home, ".local/share/opencode")}
			if err := g.importCredential(); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(home, ".cache"), 0700); err != nil {
				t.Fatal(err)
			}
			g.disposableNativeRoots = []string{filepath.Join(home, ".local"), filepath.Join(home, ".cache")}
			victim := t.TempDir()
			target := filepath.Join(home, component)
			if component == "home" || component == "home directory" {
				target = home
				t.Cleanup(func() { _ = os.RemoveAll(home + "-original") })
			}
			if component == "leaf directory" {
				target = g.shadowDataDir
			}
			// Ancestor replacement points to a victim containing the exact suffix
			// that the previous path-based cleanup would recursively delete.
			for _, suffix := range []string{"share/opencode", "opencode", "."} {
				if err := os.MkdirAll(filepath.Join(victim, suffix), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(victim, suffix, "keep"), []byte("Owned test victim.\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Rename(target, target+"-original"); err != nil {
				t.Fatal(err)
			}
			if component == "home directory" || component == "leaf directory" {
				if err := os.Mkdir(target, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(target, "replacement-keep"), []byte("Preserve replacement.\n"), 0600); err != nil {
					t.Fatal(err)
				}
			} else if err := os.Symlink(victim, target); err != nil {
				t.Fatal(err)
			}
			g.cleanup()
			for _, suffix := range []string{"share/opencode", "opencode", "."} {
				b, err := os.ReadFile(filepath.Join(victim, suffix, "keep"))
				if err != nil || string(b) != "Owned test victim.\n" {
					t.Fatal("cleanup modified an external victim")
				}
			}
			if g.report.ShadowCredentialWarning == "" {
				t.Fatal("uncertain cleanup reported success")
			}
			if component == "home directory" || component == "leaf directory" {
				b, err := os.ReadFile(filepath.Join(target, "replacement-keep"))
				if err != nil || string(b) != "Preserve replacement.\n" {
					t.Fatal("uncertain replacement data was deleted")
				}
			}
		})
	}
}

// Regression: checking ancestor paths before a path-based recursive delete
// leaves a race window. Concurrent local swaps must never reach the victim.
func TestFlashClassCleanupConcurrentAncestorReplacement(t *testing.T) {
	for iteration := 0; iteration < 12; iteration++ {
		fakeOpenCodeOnPath(t)
		home := t.TempDir()
		g := &guidanceVariant{shadowHome: home, host: "opencode", integration: "opencode", shadowDataDir: filepath.Join(home, ".local/share/opencode")}
		if err := g.importCredential(); err != nil {
			t.Fatal(err)
		}
		victim := t.TempDir()
		keep := filepath.Join(victim, "share/opencode/keep")
		if err := os.MkdirAll(filepath.Dir(keep), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, []byte("Keep raced victim.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		local := filepath.Join(home, ".local")
		backup := filepath.Join(home, ".local-original")
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < 150; j++ {
				if os.Rename(local, backup) == nil {
					_ = os.Symlink(victim, local)
					_ = os.Remove(local)
					_ = os.Rename(backup, local)
				}
			}
		}()
		close(start)
		g.cleanup()
		wg.Wait()
		b, err := os.ReadFile(keep)
		if err != nil || string(b) != "Keep raced victim.\n" {
			t.Fatal("concurrent cleanup traversed into victim")
		}
	}
}
