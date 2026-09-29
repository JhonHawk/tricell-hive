package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
)

// driftRows returns the Status rows in drift for the update fixture's hosts.
func driftRows(t *testing.T, env updateEnv) []management.StatusEntry {
	t.Helper()
	entries, err := management.Status(management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir, Hosts: []string{"codex", "claude"}})
	if err != nil {
		t.Fatal(err)
	}
	var drift []management.StatusEntry
	for _, e := range entries {
		if e.Status == "drift" {
			drift = append(drift, e)
		}
	}
	return drift
}

// TestNoCommandRepairsADriftedManagedFile is the evidence behind the repair
// line of the Installation section (M4). When a file Hive wrote no longer
// matches what Hive recorded (an instruction block edited or deleted, or a
// skill edited), `hive update`, `hive install` and `hive plan remove` all
// refuse and leave it as it is; only putting the file back the way Hive wrote
// it lets them run again. So the line cannot name a Hive command that repairs it.
func TestNoCommandRepairsADriftedManagedFile(t *testing.T) {
	for _, tc := range []struct{ mode, edit string }{
		{"instruction block edited by hand", "Keep user content."},
		{"instruction file deleted", ""},
		{"skill edited by hand", "Preserve evidence."},
	} {
		mode := tc.mode
		t.Run(mode, func(t *testing.T) {
			env, _ := newUpdateFixtureWithPendingCommit(t)
			claudeMD := filepath.Join(env.home, ".claude", "CLAUDE.md")
			if strings.HasPrefix(mode, "skill") {
				claudeMD = env.sharedSkillPath()
			}
			original, err := os.ReadFile(claudeMD)
			if err != nil {
				t.Fatal(err)
			}
			var drifted []byte
			deleted := tc.edit == ""
			if deleted {
				if err := os.Remove(claudeMD); err != nil {
					t.Fatal(err)
				}
			} else {
				drifted = bytes.Replace(original, []byte(tc.edit), []byte("Edited by hand."), 1)
				if bytes.Equal(drifted, original) {
					t.Fatal("setup: nothing was edited")
				}
				if err := os.WriteFile(claudeMD, drifted, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			rows := driftRows(t, env)
			if len(rows) == 0 {
				t.Fatal("setup: no drift row")
			}
			wantRows := len(rows)
			unchanged := func(step string) {
				t.Helper()
				got, err := os.ReadFile(claudeMD)
				if deleted {
					if !os.IsNotExist(err) {
						t.Fatalf("%s recreated the deleted file (err %v)", step, err)
					}
				} else if err != nil || !bytes.Equal(got, drifted) {
					t.Fatalf("%s changed the edited file (err %v)", step, err)
				}
				if rows := driftRows(t, env); len(rows) != wantRows {
					t.Fatalf("%s left %d drift rows, want %d", step, len(rows), wantRows)
				}
			}

			// hive update (its plan is BuildPlan("install") over the chosen commit).
			var out bytes.Buffer
			err = update(env.args(), strings.NewReader("y\n"), &out, true)
			if err == nil {
				t.Fatalf("hive update err = %v\n%s", err, out.String())
			}
			unchanged("hive update")

			// hive install over the same checkout.
			out.Reset()
			args := []string{"--home", env.home, "--state-dir", env.stateDir, "--source", env.repo, "--hosts", "codex,claude"}
			err = installWithDependencies(args, strings.NewReader("y\n"), &out, true, hostsTestDeps(coreOnlyAdapterFactory))
			if err == nil {
				t.Fatalf("hive install err = %v\n%s", err, out.String())
			}
			unchanged("hive install")

			// hive plan remove.
			_, err = management.BuildPlan("remove", management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir, Hosts: []string{"codex", "claude"}})
			if err == nil {
				t.Fatalf("hive plan remove err = %v", err)
			}
			unchanged("hive plan remove")

			// Putting the file back as Hive wrote it is what works.
			if err := os.WriteFile(claudeMD, original, 0o600); err != nil {
				t.Fatal(err)
			}
			if rows := driftRows(t, env); len(rows) != 0 {
				t.Fatalf("restoring the file left drift: %+v", rows)
			}
			out.Reset()
			if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
				t.Fatalf("hive update after restoring the file: %v\n%s", err, out.String())
			}
		})
	}
}
