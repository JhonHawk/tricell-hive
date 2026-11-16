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

func TestChildEnvironmentShadowsGitHubCredentials(t *testing.T) {
	got := withoutGitHubCredentials([]string{"PATH=/bin", "GH_TOKEN=a", "GITHUB_TOKEN=b", "GH_ENTERPRISE_TOKEN=c", "GITHUB_ENTERPRISE_TOKEN=d", "GH_CONFIG_DIR=/real", "GH_HOST=example.test", "OTHER=value"}, "/private/gh")
	want := []string{"PATH=/bin", "GH_HOST=example.test", "OTHER=value", "GH_CONFIG_DIR=/private/gh", "GH_TOKEN=" + githubPlaceholderToken, "GH_ENTERPRISE_TOKEN=" + githubPlaceholderToken}
	if !slices.Equal(got, want) {
		t.Fatalf("child environment keeps GitHub credentials or loses a setting: %v", got)
	}
}

// fakeGitHubCLI puts a gh in PATH: `auth status` exits with statusExit and
// `auth token` prints token, both only under the trimmed environment (an empty
// configuration directory and the placeholder in GH_TOKEN); otherwise `auth
// status` exits 0, which the preflight refuses.
func fakeGitHubCLI(t *testing.T, statusExit, token string) {
	t.Helper()
	bin := t.TempDir()
	script := "#!/bin/sh\n" +
		"[ -d \"$GH_CONFIG_DIR\" ] && [ -z \"$(ls -A \"$GH_CONFIG_DIR\")\" ] && [ \"$GH_TOKEN\" = '" + githubPlaceholderToken + "' ] && [ -z \"$GITHUB_TOKEN\" ] || exit 0\n" +
		"[ \"$1 $2\" = 'auth token' ] && { printf '%s\\n' " + token + "; exit 0; }\n" +
		"exit " + statusExit + "\n"
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestGitHubPreflightRecordsTheChildSession(t *testing.T) {
	fakeGitHubCLI(t, "4", `"$GH_TOKEN"`)
	out := t.TempDir()
	// A fixed parent environment: the machine's own GH_ and GITHUB_ variables
	// must not reach the recorded names.
	parent := []string{"PATH=" + os.Getenv("PATH"), "GH_TOKEN=secret-value", "GITHUB_TOKEN=secret-value", "GH_HOST=example.test"}
	env, report, err := prepareGitHubIsolation(out, parent)
	if err != nil {
		t.Fatal(err)
	}
	if !report.CLIFound || report.AuthStatusExitCode != 4 || report.AuthToken != "placeholder_only" {
		t.Fatalf("preflight did not record gh's answers: %+v", report)
	}
	if !slices.Equal(report.Variables, []string{"GH_CONFIG_DIR", "GH_ENTERPRISE_TOKEN", "GH_HOST", "GH_TOKEN"}) {
		t.Fatalf("recorded GitHub variables: %v", report.Variables)
	}
	if !slices.Contains(env, "GH_CONFIG_DIR="+filepath.Join(out, report.ConfigDirectory)) || strings.Contains(strings.Join(env, "\n"), "secret-value") {
		t.Fatal("returned environment is not the trimmed one")
	}
	if _, _, err := prepareGitHubIsolation(out, parent); err == nil {
		t.Fatal("reused a prior gh configuration directory")
	}
}

func TestGitHubPreflightRefusesAnotherToken(t *testing.T) {
	fakeGitHubCLI(t, "1", "token-from-the-keychain")
	if _, _, err := prepareGitHubIsolation(t.TempDir(), []string{"PATH=" + os.Getenv("PATH")}); err == nil {
		t.Fatal("launched although gh auth token returns a token other than the placeholder")
	}
}

func TestGitHubPreflightRefusesASignedInSession(t *testing.T) {
	fakeGitHubCLI(t, "0", `"$GH_TOKEN"`)
	if _, _, err := prepareGitHubIsolation(t.TempDir(), []string{"PATH=" + os.Getenv("PATH")}); err == nil {
		t.Fatal("launched with a gh session that still answers as signed in")
	}
}

func TestGitHubPreflightWithoutTheCLI(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, report, err := prepareGitHubIsolation(t.TempDir(), []string{"PATH=" + os.Getenv("PATH")})
	if err != nil || report.CLIFound || report.AuthStatusExitCode != 0 || report.AuthToken != "" {
		t.Fatalf("missing gh must be recorded, not fail the run: %+v, %v", report, err)
	}
}
