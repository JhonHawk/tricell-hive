package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"tricell-hive/tooling/legacy"
	"tricell-hive/tooling/management"
)

// A managed file edited by hand is refused by the text commands with one
// sentence that names the path once and says what to do (issues #65 and #66).

func editedSharedSkillEnv(t *testing.T) (updateEnv, string) {
	t.Helper()
	env := driftedEnv(t, []string{"claude", "codex"})
	return env, env.sharedSkillPath()
}

func requireChangedFileRefusal(t *testing.T, err error, path string) {
	t.Helper()
	if err == nil {
		t.Fatal("the command accepted an edited managed file")
	}
	if want := (&legacy.ModifiedFileError{Path: path}).Error(); err.Error() != want {
		t.Errorf("refusal = %q, want %q", err.Error(), want)
	}
	if n := strings.Count(err.Error(), path); n != 1 {
		t.Errorf("the path appears %d times in %q", n, err.Error())
	}
	var changed *management.ManagedFileChangedError
	if !errors.As(err, &changed) || changed.Path != path {
		t.Errorf("errors.As did not find a ManagedFileChangedError for %s in %T", path, err)
	}
}

func TestPlanRemoveRefusesAnEditedSharedSkillInPlainWords(t *testing.T) {
	env, path := editedSharedSkillEnv(t)
	err := run([]string{"plan", "remove", "--scope", "user", "--home", env.home, "--source", env.repo, "--hosts", "claude,codex", "--state-dir", env.stateDir})
	requireChangedFileRefusal(t, err, path)
	if _, statErr := os.Stat(path); statErr != nil {
		t.Errorf("the edited skill was touched: %v", statErr)
	}
}

func TestInstallCommandRefusesAnEditedSharedSkillForCursorInPlainWords(t *testing.T) {
	env, path := editedSharedSkillEnv(t)
	var out bytes.Buffer
	args := []string{"--home", env.home, "--state-dir", env.stateDir, "--source", env.repo, "--hosts", "cursor"}
	err := installWithDependencies(args, strings.NewReader("y\n"), &out, true, hostsTestDeps(coreOnlyAdapterFactory))
	requireChangedFileRefusal(t, err, path)
}
