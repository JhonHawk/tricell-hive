package main

import "testing"

func TestRejectIgnoredDestinationOptionsOnApply(t *testing.T) {
	if err := run([]string{"apply", "--plan", "anything", "--home", "/temporary"}); err == nil {
		t.Fatal("apply silently ignored alternate home")
	}
}
func TestHelpAndRequiredScope(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"plan", "install", "--hosts", "codex"}); err == nil {
		t.Fatal("missing scope accepted")
	}
}
