package main

import (
	"slices"
	"testing"
)

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
