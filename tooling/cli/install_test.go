package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallDryRunDoesNotCreateStateOrDestinations(t *testing.T) {
	home := t.TempDir()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"install", "--home", home, "--hosts", "codex", "--source", source, "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("dry run wrote into home: %v", entries)
	}
}

func installArgs(t *testing.T, home string) []string {
	t.Helper()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return []string{"--home", home, "--hosts", "codex", "--source", source}
}

func TestInstallConfirmationBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, answer   string
		tty, wantError bool
	}{
		{"cancel", "n\n", true, false}, {"default", "\n", true, false}, {"eof", "y", true, false}, {"nonterminal", "y\n", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			var out bytes.Buffer
			err := install(installArgs(t, home), strings.NewReader(test.answer), &out, test.tty)
			if (err != nil) != test.wantError {
				t.Fatalf("err=%v output=%s", err, out.String())
			}
			entries, err := os.ReadDir(home)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("unconfirmed install wrote files: %v", entries)
			}
		})
	}
}

func TestInstallThenAutoDetectNoop(t *testing.T) {
	home := t.TempDir()
	var first bytes.Buffer
	args := installArgs(t, home)
	if err := install(args, strings.NewReader("y\n"), &first, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(first.String(), "Open new CLI sessions") {
		t.Fatalf("missing restart instruction: %s", first.String())
	}
	instruction := filepath.Join(home, ".codex", "AGENTS.md")
	before, err := os.ReadFile(instruction)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(instruction)
	if err != nil {
		t.Fatal(err)
	}
	var second bytes.Buffer
	// Remove explicit hosts: registered ownership must support offline detection.
	args = append(args[:2], args[4:]...)
	if err := install(args, strings.NewReader(""), &second, false); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(instruction)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(instruction)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || !info.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("no-op rewrote installed file")
	}
	if !strings.Contains(second.String(), "No changes") {
		t.Fatalf("not a no-op: %s", second.String())
	}
}

func TestSyntheticHomeDoesNotDetectRealCLI(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	err = install([]string{"--home", home, "--source", source, "--dry-run"}, strings.NewReader(""), &out, false)
	if err == nil || !strings.Contains(err.Error(), "no CLIs detected") {
		t.Fatalf("unexpected detection: %v", err)
	}
}
