package main

import (
	"os/exec"
	"path/filepath"
	"slices"
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
