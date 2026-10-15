package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// Regression: reopening the database for each native CLI deletion can apply
// previously exported numeric IDs to a different SQLite store. The unavailable
// atomic binding must block cleanup, not be replaced by another path check.
func TestFlashClassMemoryDatabaseReplacementWithholdsDeletion(t *testing.T) {
	for _, swap := range []string{"after-export", "during-delete"} {
		t.Run(swap, func(t *testing.T) {
			bin, out, victim := t.TempDir(), t.TempDir(), t.TempDir()
			data := engramDataDir(out)
			if err := os.Mkdir(data, 0700); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(data)
			if err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(data, ".pilot-owner")
			if err := os.WriteFile(marker, []byte(memoryOwnerMarker), 0600); err != nil {
				t.Fatal(err)
			}
			create := exec.Command("python3", "-c", `import sqlite3,sys
for path in sys.argv[1:]:
 c=sqlite3.connect(path)
 for table in ['sessions','observations','prompts']:
  c.execute('create table '+table+'(id text primary key,project text,content text)')
  c.execute('insert into '+table+' values(?,?,?)',('own-session' if table=='sessions' else '101','fixture','keep victim' if 'victim' in path else 'fixture data'))
 c.commit();c.close()`, filepath.Join(data, "engram.db"), filepath.Join(victim, "victim.db"))
			if output, err := create.CombinedOutput(); err != nil {
				t.Fatalf("fictitious SQLite setup: %v %s", err, output)
			}
			script := `#!/usr/bin/env python3
import json,os,pathlib,sqlite3,sys
a=sys.argv[1:];root=os.environ.get('ENGRAM_DATA_DIR','');database=pathlib.Path(root)/'engram.db';victim=pathlib.Path(os.environ['FAKE_VICTIM_DB'])
def replace():
 database.rename(database.with_name('original.db'));victim.rename(database)
if a[0]=='export':
 doc={'version':'3','sessions':[],'observations':[],'prompts':[]}
 if root.endswith('engram-data'):
  c=sqlite3.connect(database)
  for table in ['sessions','observations','prompts']:
   doc[table]=[{'id':id if table=='sessions' else int(id),'project':project} for id,project in c.execute('select id,project from '+table)]
  c.close()
 pathlib.Path(a[1]).write_text(json.dumps(doc))
 if '--all' in a and os.environ['FAKE_SWAP']=='after-export': replace()
 sys.exit(0)
if a[0]=='delete':
 table,id=('sessions',a[2]) if a[1]=='session' else ('prompts',a[2]) if a[1]=='prompt' else ('observations',a[1])
 c=sqlite3.connect(database);c.execute('delete from '+table+' where id=?',(id,));c.commit();c.close()
 if table=='observations' and os.environ['FAKE_SWAP']=='during-delete': replace()
 sys.exit(0)
sys.exit(1)
`
			if err := os.WriteFile(filepath.Join(bin, "engram"), []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("FAKE_VICTIM_DB", filepath.Join(victim, "victim.db"))
			t.Setenv("FAKE_SWAP", swap)
			g, err := prepareFlashMemoryGuard("fixture", out, []string{"PATH=" + os.Getenv("PATH"), "FAKE_VICTIM_DB=" + os.Getenv("FAKE_VICTIM_DB"), "FAKE_SWAP=" + swap})
			if err != nil {
				t.Fatal(err)
			}
			g.ownedMemory = &memoryIsolation{outputDir: out, dataDir: data, marker: marker, ownedInfo: info}
			if g.finish(isolatedMemoryEnvironment(os.Environ(), data), time.Now().Add(5*time.Second)) == nil {
				t.Fatal("unbound native deletion reported successful cleanup")
			}
			path := filepath.Join(victim, "victim.db")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				path = filepath.Join(data, "engram.db")
			}
			check := exec.Command("python3", "-c", `import sqlite3,sys
c=sqlite3.connect('file:'+sys.argv[1]+'?mode=ro',uri=True)
assert all(c.execute('select count(*) from '+t).fetchone()[0]==1 for t in ['sessions','observations','prompts'])
c.close()`, path)
			if output, err := check.CombinedOutput(); err != nil {
				t.Fatalf("second fictitious store changed: %v %s", err, output)
			}
		})
	}
}
