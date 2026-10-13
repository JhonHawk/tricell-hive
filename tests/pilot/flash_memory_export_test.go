package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFlashMemoryContainerFullExport(t *testing.T) {
	for _, mode := range []string{"empty", "owned", "unexpected", "default"} {
		t.Run(mode, func(t *testing.T) {
			bin := t.TempDir()
			log := filepath.Join(t.TempDir(), "calls.json")
			script := `#!/usr/bin/env python3
import json,os,pathlib,sys
args=sys.argv[1:]
pathlib.Path(os.environ['FAKE_EXPORT_LOG']).write_text(json.dumps(args))
mode=os.environ['FAKE_EXPORT_MODE']
if mode=='empty' and '--project' in args: sys.exit(1)
doc={'version':'0.2.0','sessions':None,'observations':[],'prompts':None}
if mode in ('owned','default'): doc['observations']=[{'id':101,'project':'fixture'}]
if mode=='unexpected' and '--all' in args: doc['observations']=[{'id':102,'project':'other-project'}]
pathlib.Path(args[1]).write_text(json.dumps(doc))
`
			if err := os.WriteFile(filepath.Join(bin, "engram"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("FAKE_EXPORT_LOG", log)
			t.Setenv("FAKE_EXPORT_MODE", mode)
			out := t.TempDir()
			dir := engramDataDir(out)
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(dir)
			if err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(dir, ".pilot-owner")
			if err := os.WriteFile(marker, []byte(memoryOwnerMarker), 0600); err != nil {
				t.Fatal(err)
			}
			g := &flashMemoryGuard{project: "fixture", output: out, temporary: t.TempDir(), disposable: mode != "default", ownedMemory: &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}}
			t.Cleanup(g.removeTemporaryExports)
			env := isolatedMemoryEnvironment(os.Environ(), dir)
			if mode == "default" {
				records, err := g.export(env, g.project, time.Now().Add(time.Second))
				if err != nil || len(records.Observations) != 1 {
					t.Fatal("default export changed", err)
				}
				b, _ := os.ReadFile(log)
				if !strings.Contains(string(b), "--project") || strings.Contains(string(b), "--all") {
					t.Fatal("default arguments changed")
				}
				return
			}
			r := result{MemoryStoreMode: disposableContainer, FlashClass: &flashClassReport{}}
			finishFlashMemoryResult(&r, g, env, time.Now().Add(time.Second))
			writeRunResult(filepath.Join(out, "run.json"), r)
			b, err := os.ReadFile(filepath.Join(out, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report map[string]any
			if err := json.Unmarshal(b, &report); err != nil {
				t.Fatal(err)
			}
			flash := report["FlashClass"].(map[string]any)
			records := flash["MemoryRecords"].(map[string]any)
			counts, ok := records["exported_counts"].(map[string]any)
			if !ok {
				t.Error("exported counts missing")
			} else {
				want := float64(1)
				if mode == "empty" {
					want = 0
				}
				if counts["observations"] != want || counts["sessions"] != float64(0) || counts["prompts"] != float64(0) {
					t.Errorf("wrong exported counts: %+v", counts)
				}
			}
			if mode == "unexpected" {
				if flash["Memory"] != "blocked" || !strings.Contains(flash["Error"].(string), "provenance unverified") {
					t.Error("unexpected project silently accepted")
				}
				if !strings.Contains(string(b), "other-project") {
					t.Error("unexpected project missing from result")
				}
			} else if flash["Memory"] != deferredVolumeDestruction {
				t.Errorf("expected deferred state, got %v", flash["Memory"])
			}
			raw, err := os.ReadFile(filepath.Join(out, "memory-export.json"))
			if err != nil {
				t.Fatal(err)
			}
			if mode == "unexpected" && !strings.Contains(string(raw), "other-project") {
				t.Error("unexpected record evidence lost")
			}
			calls, _ := os.ReadFile(log)
			if !strings.Contains(string(calls), "--all") || strings.Contains(string(calls), "--project") {
				t.Error("container did not export whole store")
			}
		})
	}
}
