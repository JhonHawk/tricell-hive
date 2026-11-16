package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestChildEnvironmentDirectoryMatchesProcess(t *testing.T) {
	dir := t.TempDir()
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("/bin/sh", "-c", `test "$PWD" = "$(pwd -P)" && test -z "$OLDPWD"`)
	cmd.Dir = dir
	cmd.Env = environmentInDirectory([]string{"PWD=/wrong", "OLDPWD=/stale", "PATH=/bin:/usr/bin"}, dir)
	if err := cmd.Run(); err != nil {
		t.Fatal("child PWD disagrees with fixture directory", err)
	}
}

func TestCodexPermissionModes(t *testing.T) {
	r := result{Host: "codex", ModelRequested: "gpt-5.6-terra", Cwd: "/fixture", Effort: "medium"}
	args, err := launchArgs(r, "/output", "prompt")
	if err != nil || !slices.Contains(args, "workspace-write") || !slices.Contains(args, "never") || slices.Contains(args, "--dangerously-bypass-approvals-and-sandbox") {
		t.Fatalf("default permissions changed: %v, %v", args, err)
	}
	r.CodexBypassSandbox = true
	args, err = launchArgs(r, "/output", "prompt")
	if err != nil || !slices.Contains(args, "--dangerously-bypass-approvals-and-sandbox") || slices.Contains(args, "-s") || slices.Contains(args, "-a") {
		t.Fatalf("explicit mode has conflicting permissions: %v, %v", args, err)
	}
	r.Host = "claude"
	if _, err = launchArgs(r, "/output", "prompt"); err == nil {
		t.Fatal("accepted Codex permission mode for another host")
	}
}

func TestGrokPreservesNativePermissionSelection(t *testing.T) {
	args, err := launchArgs(result{Host: "grok", ModelRequested: "grok-4.6", Cwd: "/fixture"}, "/output", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	for _, override := range []string{"--permission-mode", "--allow", "--deny", "--always-approve", "--sandbox", "--tools", "--disallowed-tools"} {
		if slices.Contains(args, override) {
			t.Fatalf("pilot overrides native Grok permissions with %s", override)
		}
	}
	if !slices.Contains(args, "-p") || !slices.Contains(args, "streaming-messages-json") {
		t.Fatal("lost measurement transport")
	}
}

func TestCodexNativeDispatchKeepsParentThread(t *testing.T) {
	for _, bypass := range []bool{false, true} {
		r := result{Host: "codex", Case: "native-agent-dispatch", ModelRequested: "gpt-5.6-terra", Cwd: "/fixture", CodexBypassSandbox: bypass}
		args, err := launchArgs(r, "/output", "prompt")
		if err != nil || slices.Contains(args, "--ephemeral") {
			t.Fatalf("native dispatch needs a persisted parent thread: %v, %v", args, err)
		}
		r.Case = "smoke"
		args, err = launchArgs(r, "/output", "prompt")
		if err != nil || !slices.Contains(args, "--ephemeral") {
			t.Fatalf("single-agent smoke lost ephemeral mode: %v, %v", args, err)
		}
	}
}

func TestChildEnvironmentDropsGitHubCredentials(t *testing.T) {
	got := withoutGitHubCredentials([]string{"PATH=/bin", "GH_TOKEN=a", "GITHUB_TOKEN=b", "GH_ENTERPRISE_TOKEN=c", "GITHUB_ENTERPRISE_TOKEN=d", "GH_CONFIG_DIR=/real", "GH_HOST=example.test", "OTHER=value"}, "/private/gh")
	want := []string{"PATH=/bin", "GH_HOST=example.test", "OTHER=value", "GH_CONFIG_DIR=/private/gh"}
	if !slices.Equal(got, want) {
		t.Fatalf("child environment keeps GitHub credentials or loses a setting: %v", got)
	}
}

// The fake gh exits 4 only for `auth status` under the trimmed environment, so
// the recorded code proves the preflight ran gh with the child's own settings.
func TestGitHubPreflightRecordsTheChildSession(t *testing.T) {
	bin := t.TempDir()
	script := "#!/bin/sh\n[ \"$1 $2\" = 'auth status' ] || exit 0\n[ -z \"$GH_TOKEN$GITHUB_TOKEN$GH_ENTERPRISE_TOKEN$GITHUB_ENTERPRISE_TOKEN\" ] || exit 0\n[ -d \"$GH_CONFIG_DIR\" ] || exit 0\n[ -z \"$(ls -A \"$GH_CONFIG_DIR\")\" ] || exit 0\nexit 4\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out := t.TempDir()
	env, report, err := prepareGitHubIsolation(out, append(os.Environ(), "GH_TOKEN=secret-value", "GITHUB_TOKEN=secret-value", "GH_HOST=example.test"))
	if err != nil {
		t.Fatal(err)
	}
	if !report.CLIFound || report.AuthStatusExitCode != 4 || report.PreflightCommand != "gh auth status" {
		t.Fatalf("preflight did not record gh's answer: %+v", report)
	}
	if !slices.Equal(report.Variables, []string{"GH_CONFIG_DIR", "GH_HOST"}) {
		t.Fatalf("recorded GitHub variables: %v", report.Variables)
	}
	if !slices.Contains(env, "GH_CONFIG_DIR="+filepath.Join(out, report.ConfigDirectory)) || strings.Contains(strings.Join(env, "\n"), "secret-value") {
		t.Fatal("returned environment is not the trimmed one")
	}
	if _, _, err := prepareGitHubIsolation(out, os.Environ()); err == nil {
		t.Fatal("reused a prior gh configuration directory")
	}
}

func TestGitHubPreflightRefusesASignedInSession(t *testing.T) {
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if _, _, err := prepareGitHubIsolation(t.TempDir(), os.Environ()); err == nil {
		t.Fatal("launched with a gh session that still answers as signed in")
	}
}

func TestGitHubPreflightWithoutTheCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, report, err := prepareGitHubIsolation(t.TempDir(), os.Environ())
	if err != nil || report.CLIFound || report.AuthStatusExitCode != 0 {
		t.Fatalf("missing gh must be recorded, not fail the run: %+v, %v", report, err)
	}
}
