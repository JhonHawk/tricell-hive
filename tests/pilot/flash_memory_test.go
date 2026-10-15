package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Regression: deleting an ID without exact fixture ownership can remove human
// records. Missing/duplicate IDs or another project's export must be blocked.
func TestFlashClassMemoryProvenance(t *testing.T) {
	for _, raw := range []string{
		`{"version":"3","observations":[{"id":1,"project":"human"}]}`,
		`{"version":"3","sessions":[{"id":"one"}]}`,
		`{"version":"3","observations":[{"id":1,"project":"fixture"},{"id":1,"project":"fixture"}]}`,
		`{"version":"3","prompts":[{"id":0,"project":"fixture"}]}`,
		`{"private":"FAKE_PRIVATE_MARKER",`,
	} {
		if _, err := parseFlashMemoryRecords([]byte(raw), "fixture"); err == nil || strings.Contains(err.Error(), "FAKE_PRIVATE_MARKER") {
			t.Fatal("unowned records accepted or private fields exposed")
		}
	}
}

// Regression: following a replaced directory or database symlink during export
// could read everyday history before the later cleanup ownership check runs.
func TestFlashClassMemoryRejectsSymlinkBeforeExport(t *testing.T) {
	_, log := fakeFlashMemoryCLI(t)
	out := t.TempDir()
	g, err := prepareFlashMemoryGuard("fixture", out, os.Environ())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(g.removeTemporaryExports)
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
	g.ownedMemory = &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}
	outside := filepath.Join(t.TempDir(), "human.db")
	if err := os.WriteFile(outside, []byte("FAKE_PRIVATE_CONTENT"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(dir, "engram.db")); err != nil {
		t.Fatal(err)
	}
	if g.finish(isolatedMemoryEnvironment(os.Environ(), dir), time.Now().Add(time.Second)) == nil {
		t.Fatal("symlinked runtime store accepted")
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(strings.TrimSpace(string(data)), "\n")) != 1 {
		t.Fatal("unsafe runtime path reached export or deletion")
	}
}

func fakeFlashMemoryCLI(t *testing.T) (string, string) {
	t.Helper()
	bin := t.TempDir()
	state := filepath.Join(t.TempDir(), "state.json")
	log := filepath.Join(t.TempDir(), "calls.jsonl")
	script := `#!/usr/bin/env python3
import json,os,pathlib,sys
args=sys.argv[1:];data=os.environ.get('ENGRAM_DATA_DIR','everyday')
with open(os.environ['FAKE_MEMORY_LOG'],'a') as f: f.write(json.dumps({'args':args,'store':data})+'\n')
state=pathlib.Path(os.environ['FAKE_MEMORY_STATE'])
records=json.loads(state.read_text()) if state.exists() else {'version':'3','sessions':[{'id':'child-session','project':'fixture'}],'observations':[{'id':101,'project':'fixture'}],'prompts':[{'id':202,'project':'fixture'}]}
if args[0]=='export':
 if data.endswith('engram-data'): doc=records
 elif os.environ.get('FAKE_LEAK')=='1': doc={'version':'3','observations':[{'id':909,'project':'fixture'}]}
 else: doc={'version':'3','sessions':None,'observations':[],'prompts':None}
 pathlib.Path(args[1]).write_text(json.dumps(doc));sys.exit(0)
if args[0]=='delete':
 if os.environ.get('FAKE_DELETE_FAIL')=='1': sys.exit(1)
 if args[1]=='session': key,id='sessions',args[2]
 elif args[1]=='prompt': key,id='prompts',int(args[2])
 else: key,id='observations',int(args[1])
 records[key]=[x for x in records[key] if x['id']!=id];state.write_text(json.dumps(records));sys.exit(0)
sys.exit(1)
`
	if err := os.WriteFile(filepath.Join(bin, "engram"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_MEMORY_STATE", state)
	t.Setenv("FAKE_MEMORY_LOG", log)
	return state, log
}

// Regression: exported exact IDs do not bind a later CLI invocation to the same
// store. Until that contract is available, preserve the store and block cleanup.
func TestFlashClassMemoryExactCleanupAndLeakDetection(t *testing.T) {
	for _, mode := range []string{"clean", "everyday leak", "cleanup failure", "routing leak"} {
		t.Run(mode, func(t *testing.T) {
			_, log := fakeFlashMemoryCLI(t)
			out := t.TempDir()
			g, err := prepareFlashMemoryGuard("fixture", out, os.Environ())
			if err != nil {
				t.Fatal(err)
			}
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
			g.ownedMemory = &memoryIsolation{outputDir: out, dataDir: dir, marker: marker, ownedInfo: info}
			t.Cleanup(g.removeTemporaryExports)
			env := isolatedMemoryEnvironment(os.Environ(), engramDataDir(out))
			if mode == "everyday leak" {
				g.everydayEnv = append(g.everydayEnv, "FAKE_LEAK=1")
			}
			if mode == "cleanup failure" {
				env = append(env, "FAKE_DELETE_FAIL=1")
			}
			if mode == "routing leak" {
				env = isolatedMemoryEnvironment(os.Environ(), "/human/store")
			}
			err = g.finish(env, time.Now().Add(5*time.Second))
			if err == nil {
				t.Fatal("unsafe or incomplete cleanup accepted")
			}
			if mode == "clean" && !strings.Contains(err.Error(), "atomic memory store binding unavailable") {
				t.Fatal("missing atomic contract was not reported", err)
			}
			b, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			deletes := 0
			for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
				var call struct {
					Args  []string
					Store string
				}
				if json.Unmarshal([]byte(line), &call) != nil {
					t.Fatal("fake CLI log malformed")
				}
				if call.Args[0] == "delete" {
					deletes++
					if call.Store != engramDataDir(out) {
						t.Fatal("cleanup touched everyday memory")
					}
				}
				if call.Store != engramDataDir(out) && strings.Contains(strings.Join(call.Args, " "), "--all") {
					t.Fatal("everyday history was exported")
				}
			}
			if deletes != 0 {
				t.Fatal("unbound store performed deletions")
			}
			if _, err := os.Stat(g.temporary); !os.IsNotExist(err) {
				t.Fatal("temporary memory export survived cleanup")
			}
		})
	}
}
