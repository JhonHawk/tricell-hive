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

// driftFixture builds the update fixture, then edits (edit != "") or deletes
// (edit == "") one managed file: the user instruction file, or the shared skill
// when skill is true. It returns the path and the bytes Hive wrote.
func driftFixture(t *testing.T, skill bool, edit string) (env updateEnv, path string, original, drifted []byte) {
	t.Helper()
	env, _ = newUpdateFixtureWithPendingCommit(t)
	path = filepath.Join(env.home, ".claude", "CLAUDE.md")
	if skill {
		path = env.sharedSkillPath()
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if edit == "" {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	} else {
		drifted = bytes.Replace(original, []byte(edit), []byte("Edited by hand."), 1)
		if bytes.Equal(drifted, original) {
			t.Fatal("setup: nothing was edited")
		}
		if err := os.WriteFile(path, drifted, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if len(driftRows(t, env)) == 0 {
		t.Fatal("setup: no drift row")
	}
	return env, path, original, drifted
}

func driftInstall(env updateEnv, out *bytes.Buffer) error {
	args := []string{"--home", env.home, "--state-dir", env.stateDir, "--source", env.repo, "--hosts", "codex,claude"}
	return installWithDependencies(args, strings.NewReader("y\n"), out, true, hostsTestDeps(coreOnlyAdapterFactory))
}

func driftRemovePlan(env updateEnv) error {
	_, err := management.BuildPlan("remove", management.Options{Scope: "user", Home: env.home, StateDir: env.stateDir, Hosts: []string{"codex", "claude"}})
	return err
}

// TestEditedManagedFilesAreRefusedAndDeletedOnesReinstalled is the evidence
// behind the repair line of the Installation section (M4). A managed file that
// was edited by hand (an instruction block or a skill) makes `hive update`,
// `hive install` and `hive plan remove` refuse and leave it as it is; only
// putting the file back the way Hive wrote it lets them run again. A managed
// file that was deleted is different: `hive update` and `hive install` put it
// back, and `hive plan remove` completes without recreating it.
func TestEditedManagedFilesAreRefusedAndDeletedOnesReinstalled(t *testing.T) {
	for _, tc := range []struct {
		mode, edit string
		skill      bool
	}{
		{"instruction block edited by hand", "Keep user content.", false},
		{"skill edited by hand", "Preserve evidence.", true},
	} {
		tc := tc
		t.Run(tc.mode, func(t *testing.T) {
			env, path, original, drifted := driftFixture(t, tc.skill, tc.edit)
			wantRows := len(driftRows(t, env))
			unchanged := func(step string) {
				t.Helper()
				got, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(got, drifted) {
					t.Fatalf("%s changed the edited file (err %v)", step, err)
				}
				if rows := driftRows(t, env); len(rows) != wantRows {
					t.Fatalf("%s left %d drift rows, want %d", step, len(rows), wantRows)
				}
			}

			// hive update (its plan is BuildPlan("install") over the chosen commit).
			var out bytes.Buffer
			if err := update(env.args(), strings.NewReader("y\n"), &out, true); err == nil {
				t.Fatalf("hive update succeeded\n%s", out.String())
			}
			unchanged("hive update")

			out.Reset()
			if err := driftInstall(env, &out); err == nil {
				t.Fatalf("hive install succeeded\n%s", out.String())
			}
			unchanged("hive install")

			if err := driftRemovePlan(env); err == nil {
				t.Fatal("hive plan remove succeeded")
			}
			unchanged("hive plan remove")

			// Putting the file back as Hive wrote it is what works.
			if err := os.WriteFile(path, original, 0o600); err != nil {
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

	for _, tc := range []struct {
		mode  string
		skill bool
	}{
		{"instruction file deleted", false},
		{"skill deleted", true},
	} {
		tc := tc
		t.Run(tc.mode, func(t *testing.T) {
			// update and install run over the fixture's newest commit, so a
			// reinstalled skill may carry newer bytes than the deleted one; the
			// file must exist again and Status must be clean.
			reinstalled := func(t *testing.T, env updateEnv, path string, step string) {
				t.Helper()
				if got, err := os.ReadFile(path); err != nil || len(got) == 0 {
					t.Fatalf("%s did not put the deleted file back (err %v)", step, err)
				}
				if rows := driftRows(t, env); len(rows) != 0 {
					t.Fatalf("%s left drift: %+v", step, rows)
				}
			}
			t.Run("hive update", func(t *testing.T) {
				env, path, _, _ := driftFixture(t, tc.skill, "")
				var out bytes.Buffer
				if err := update(env.args(), strings.NewReader("y\n"), &out, true); err != nil {
					t.Fatalf("hive update: %v\n%s", err, out.String())
				}
				reinstalled(t, env, path, "hive update")
			})
			t.Run("hive install", func(t *testing.T) {
				env, path, _, _ := driftFixture(t, tc.skill, "")
				var out bytes.Buffer
				if err := driftInstall(env, &out); err != nil {
					t.Fatalf("hive install: %v\n%s", err, out.String())
				}
				reinstalled(t, env, path, "hive install")
			})
			t.Run("hive plan remove", func(t *testing.T) {
				env, path, _, _ := driftFixture(t, tc.skill, "")
				if err := driftRemovePlan(env); err != nil {
					t.Fatalf("hive plan remove: %v", err)
				}
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("hive plan remove recreated the deleted file (err %v)", err)
				}
			})
		})
	}
}
