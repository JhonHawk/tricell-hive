package hostevent

import (
	"path/filepath"
	"regexp"
	"strings"
)

// HookTimeoutSeconds is the explicit timeout every JSON hook entry carries. It
// is above the 1.5 s default budget of Claude Code's SessionEnd and equal to
// the whole-run budget of `hive hook`, so the hook finishes before its host
// gives up on it.
const HookTimeoutSeconds = 5

// Command is the command line a host's hook file registers for one event:
// the hive binary as one single-quoted shell word, then `hook <host> <event>`.
// The hosts run it with a shell, so the path is quoted; the manager checks the
// path before it is written.
func Command(executable, word, event string) string {
	return shellQuote(executable) + " hook " + word + " " + event
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

var commandPattern = regexp.MustCompile(`^'((?:[^']|'\\'')*)' hook ([a-z]+) ([A-Za-z]+)$`)

// ParseCommand splits a command Command produced. ok is false for any other
// text, so a foreign hook is never taken for Hive's.
func ParseCommand(command string) (executable, word, event string, ok bool) {
	m := commandPattern.FindStringSubmatch(command)
	if m == nil {
		return "", "", "", false
	}
	executable = strings.ReplaceAll(m[1], `'\''`, "'")
	if !filepath.IsAbs(executable) {
		return "", "", "", false
	}
	return executable, m[2], m[3], true
}

// IsHiveCommand reports whether command is a hook command of Hive for the host
// word and event: either exactly one of the commands the manager recorded, or
// the registered shape with an absolute path to a binary named exactly "hive"
// (so a moved binary is still recognized and replaced, and a lookalike such as
// hivectl or a foreign hook that merely mentions "hook" never is; a binary under
// another name is recognized only through the recorded commands).
func IsHiveCommand(command, word, event string, recorded map[string]bool) bool {
	if recorded[command] {
		return true
	}
	exe, w, e, ok := ParseCommand(command)
	return ok && w == word && e == event && filepath.Base(exe) == "hive"
}
