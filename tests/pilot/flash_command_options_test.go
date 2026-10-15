package main

import "testing"

func TestFlashCommandRejectsUnlistedOptions(t *testing.T) {
	commands := []string{"rg --pre /tmp/program needle /fixture", "git diff --no-index --output /fixture/out /fixture/a /fixture/b", "rg --pre=/tmp/program needle /fixture", "grep --exclude-from /fixture/list needle /fixture", "find /fixture -fprint /fixture/out", "git diff --output=/fixture/out", "git diff --ext-diff", "git show --textconv", "ls --color", "cat -n file", "pwd -P", "wc --files0-from file", "rg -e needle", "find . -execdir program", "find . -delete", "git log --format format", "git status --porcelain", "git rev-parse --git-dir", "git rev-parse main", "head -n", "tail -n -1", "head -n word", "git log -n", "git log -n 1x", "find . -name", "find . -name -delete", "find . -type -fprint", "find . -maxdepth --help", "ls --", "head -n=10 file", "rg -n=1 needle file"}
	for i, command := range commands {
		if flashOperationCommand(command) != "" {
			t.Errorf("unlisted or malformed option accepted at index %d", i)
		}
	}
}

func TestFlashCommandKeepsMinimalOptions(t *testing.T) {
	commands := []string{"ls -l -a -la -1 /fixture", "cat /fixture/plan.md", "head -n 10 file", "tail -n 0 file", "wc -l -w -c file", "pwd", "rg -n -i -l -c -F -w needle /fixture", "grep -n -i -l -c -F -w needle file", "find /fixture -name plan.md -iname readme.md -type f -maxdepth 2", "git status --short -s", "git log --oneline -n 10", "git diff --stat /fixture/a", "git show --stat HEAD", "git rev-parse HEAD", "git rev-parse --show-toplevel"}
	for i, command := range commands {
		if flashOperationCommand(command) != command {
			t.Errorf("listed option rejected at index %d", i)
		}
	}
}
