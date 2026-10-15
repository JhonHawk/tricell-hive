package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Regression: a queued batch, automatic retry or failure short-circuit that
// leaves its partner running violates the two-root bound and cleanup contract.
func TestFlashClassControllerUsesFixedScheduleAndStopsWithoutRetry(t *testing.T) {
	for _, failure := range []string{"none", "terminal", "operations"} {
		failed := failure != "none"
		root := t.TempDir()
		schedule, _ := json.Marshal(flashSchedule())
		if err := os.WriteFile(filepath.Join(root, "schedule.json"), schedule, 0600); err != nil {
			t.Fatal(err)
		}
		script := `#!/usr/bin/env python3
import fcntl,json,pathlib,sys,time
root=pathlib.Path(__file__).parent
if '--print-flash-schedule' in sys.argv: print((root/'schedule.json').read_text());sys.exit(0)
out=pathlib.Path(sys.argv[sys.argv.index('--out')+1]);out.mkdir()
with open(root/'state.json','a+') as f:
 fcntl.flock(f,fcntl.LOCK_EX);f.seek(0);text=f.read();state=json.loads(text) if text else {'active':0,'max':0,'runs':[]}
 state['active']+=1;state['max']=max(state['max'],state['active']);state['runs'].append(out.name);f.seek(0);f.truncate();f.write(json.dumps(state));f.flush()
time.sleep(.02)
failure=(root/'fail').read_text() if (root/'fail').exists() else 'none'
terminal='process_error' if failure=='terminal' else 'completed'
(out/'run.json').write_text(json.dumps({'Terminal':terminal,'OutsideChanges':[],'FlashClass':{'OperationSafety':'not_observed' if failure=='operations' else 'pass'}}))
with open(root/'state.json','r+') as f:
 fcntl.flock(f,fcntl.LOCK_EX);state=json.load(f);state['active']-=1;f.seek(0);f.truncate();f.write(json.dumps(state));f.flush()
`
		runner := filepath.Join(root, "runner")
		if err := os.WriteFile(runner, []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
		if failed {
			if err := os.WriteFile(filepath.Join(root, "fail"), []byte(failure), 0600); err != nil {
				t.Fatal(err)
			}
		}
		cmd := exec.Command("python3", "../fixtures/flash-class-guidance/control.py", "--runner", runner, "--source", root, "--baseline", root, "--out", filepath.Join(root, "batch"), "--execute")
		output, err := cmd.CombinedOutput()
		if failed && err == nil {
			t.Fatal("failed batch accepted")
		}
		if !failed && err != nil {
			t.Fatalf("fake controller failed: %v %s", err, output)
		}
		data, err := os.ReadFile(filepath.Join(root, "state.json"))
		if err != nil {
			t.Fatal(err)
		}
		var state struct {
			Active, Max int
			Runs        []string
		}
		if json.Unmarshal(data, &state) != nil {
			t.Fatal("fake runner state malformed")
		}
		if state.Active != 0 || state.Max > 2 {
			t.Fatal("controller left an active root or exceeded concurrency")
		}
		if (!failed && len(state.Runs) != 48) || (failed && (len(state.Runs) < 1 || len(state.Runs) > 2)) {
			t.Fatalf("unexpected root count after bounded scheduling: %d", len(state.Runs))
		}
		seen := map[string]bool{}
		for _, run := range state.Runs {
			if seen[run] {
				t.Fatal("controller retried a root")
			}
			seen[run] = true
		}
	}
}
