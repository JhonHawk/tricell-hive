package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

// shellWork and its helpers decide, without running anything, whether a shell
// line changes a file outside the host's temporary location or delivers work.
// analysis_reply_block uses it to leave work turns unjudged, so every doubt
// resolves toward work: a line it cannot tokenize, a path it cannot resolve
// and a writer fed by xargs all count as work.

var (
	redirectWord = regexp.MustCompile(`^(?:[0-9]*|&)>>?(.*)$`)
	inPlaceFlag  = regexp.MustCompile(`^(?:-[a-zA-Z]*i\S*|--in-place\S*)$`)
	// shellWrappers run the command that follows them.
	shellWrappers = map[string]bool{"env": true, "sudo": true, "time": true, "command": true, "nohup": true, "exec": true, "nice": true, "xargs": true}
	// fileWriters change the files they name; destinationOnly writers change
	// only their last operand or their -t value.
	fileWriters     = map[string]bool{"tee": true, "cp": true, "mv": true, "rm": true, "rmdir": true, "mkdir": true, "touch": true, "ln": true, "install": true, "truncate": true}
	destinationOnly = map[string]bool{"cp": true, "ln": true, "install": true}
	// writerValueFlags take a value that is not a file the command changes.
	writerValueFlags = map[string]bool{"-s": true, "--size": true, "-r": true, "--reference": true, "-m": true, "--mode": true, "-o": true, "-g": true, "-S": true, "--suffix": true, "-d": true}
	// gitAlwaysWork are git subcommands that change the working tree, the
	// history or a remote whatever their arguments.
	gitAlwaysWork = map[string]bool{"commit": true, "push": true, "merge": true, "apply": true, "am": true, "checkout": true, "switch": true, "reset": true, "restore": true, "rm": true, "mv": true, "add": true, "rebase": true, "cherry-pick": true, "revert": true, "pull": true, "clean": true}
	gitListFlags  = map[string]bool{"-l": true, "--list": true, "-a": true, "--all": true, "-r": true, "--remotes": true, "--contains": true, "--no-contains": true, "--merged": true, "--no-merged": true, "--points-at": true, "--show-current": true, "-v": true, "-vv": true, "-n": true}
)

// shellWork reports whether a shell line changes a file outside the host's
// temporary location or delivers work.
func shellWork(command string) bool {
	for _, p := range pythonWritePath.FindAllStringSubmatch(command, -1) {
		if !tempPath.MatchString(p[1]) {
			return true
		}
	}
	segments := shellSegments(inlineShellHeredocs(command), 0)
	if segments == nil {
		return true
	}
	cwd := ""
	for _, segment := range segments {
		for _, args := range splitBackground(segment) {
			work, next := commandWork(args, cwd)
			if work {
				return true
			}
			cwd = next
		}
	}
	return false
}

// splitBackground splits a simple command at each "&" that sends a command to
// the background.
func splitBackground(args []string) [][]string {
	var out [][]string
	start := 0
	for i, a := range args {
		if a == "&" {
			out = append(out, args[start:i])
			start = i + 1
		}
	}
	return append(out, args[start:])
}

// commandWork reports whether one simple command does work, and the working
// folder after it: "" when the folder is not known.
func commandWork(args []string, cwd string) (bool, string) {
	outside := func(p string) bool { return outsideTemp(p, cwd) }
	var words []string
	for i := 0; i < len(args); i++ {
		a := strings.TrimRight(strings.TrimLeft(args[i], "({"), ")}")
		if a == "" {
			continue
		}
		if m := redirectWord.FindStringSubmatch(a); m != nil && a != m[1] {
			target := m[1]
			if target == "" && i+1 < len(args) {
				i++
				target = args[i]
			}
			if target != "" && !strings.HasPrefix(target, "&") && outside(target) {
				return true, cwd
			}
			continue
		}
		words = append(words, a)
	}
	xargs := false
	for len(words) > 0 {
		name := filepath.Base(words[0])
		switch {
		case strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-"):
			words = words[1:]
		case shellWrappers[name]:
			xargs = xargs || name == "xargs"
			words = words[1:]
			for len(words) > 0 && (strings.HasPrefix(words[0], "-") || strings.Contains(words[0], "=")) {
				words = words[1:]
			}
		case interpHole.MatchString(words[0]):
			return true, cwd
		default:
			return namedWork(filepath.Base(words[0]), words, cwd, xargs, outside)
		}
	}
	return false, cwd
}

func namedWork(name string, words []string, cwd string, xargs bool, outside func(string) bool) (bool, string) {
	args := words[1:]
	switch {
	case name == "cd":
		return false, nextFolder(args, cwd)
	case fileWriters[name]:
		if xargs {
			return true, cwd
		}
		for _, p := range writerTargets(name, args) {
			if outside(p) {
				return true, cwd
			}
		}
	case (name == "sed" || name == "perl") && hasInPlace(args):
		if xargs {
			return true, cwd
		}
		for _, p := range inPlaceFiles(args) {
			if outside(p) {
				return true, cwd
			}
		}
	case name == "find":
		if findChanges(args) {
			paths := 0
			for _, a := range args {
				if strings.HasPrefix(a, "-") || a == "!" || a == "(" {
					break
				}
				paths++
				if outside(a) {
					return true, cwd
				}
			}
			return paths == 0 && outside("."), cwd
		}
	case name == "git":
		return gitWork(words, cwd), cwd
	case name == "gh":
		return len(args) >= 2 && args[0] == "pr" && (args[1] == "create" || args[1] == "merge"), cwd
	default:
		for i, a := range words {
			if filepath.Base(a) == "git_state.py" && containsField(words[i+1:], "commit") {
				return true, cwd
			}
		}
	}
	return false, cwd
}

// nextFolder is the working folder after cd: known only when the target is an
// absolute path, or relative to a known folder.
func nextFolder(args []string, cwd string) string {
	if len(args) != 1 || strings.ContainsAny(args[0], "$`~") || interpHole.MatchString(args[0]) || args[0] == "-" {
		return ""
	}
	if filepath.IsAbs(args[0]) {
		return filepath.Clean(args[0])
	}
	if cwd == "" {
		return ""
	}
	return filepath.Join(cwd, args[0])
}

// outsideTemp reports whether a path a command writes lies outside the host's
// temporary location, or cannot be resolved.
func outsideTemp(p, cwd string) bool {
	switch {
	case p == "/dev/null" || p == "/dev/stdout" || p == "/dev/stderr":
		return false
	case tempPath.MatchString(p):
		return false
	case strings.ContainsAny(p, "$`") || strings.HasPrefix(p, "~") || interpHole.MatchString(p):
		return true
	case filepath.IsAbs(p):
		return !tempPath.MatchString(filepath.Clean(p))
	case cwd == "":
		return true
	}
	return !tempPath.MatchString(filepath.Join(cwd, p))
}

// writerTargets are the files a writer changes: every operand, or only the
// destination of cp, ln and install.
func writerTargets(name string, args []string) []string {
	var operands []string
	destination := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-t" || a == "--target-directory":
			if i+1 < len(args) {
				i++
				destination = args[i]
			}
		case strings.HasPrefix(a, "--target-directory="):
			destination = strings.TrimPrefix(a, "--target-directory=")
		case writerValueFlags[a] && !(name == "ln" && a == "-s") && !(name == "rm" && a == "-r") && !(name == "rmdir" && a == "-r"):
			i++
		case strings.HasPrefix(a, "-"):
		default:
			operands = append(operands, a)
		}
	}
	if destinationOnly[name] {
		if destination != "" {
			return []string{destination}
		}
		if len(operands) > 0 {
			return operands[len(operands)-1:]
		}
		return nil
	}
	if destination != "" {
		operands = append(operands, destination)
	}
	return operands
}

func hasInPlace(args []string) bool {
	for _, a := range args {
		if inPlaceFlag.MatchString(a) {
			return true
		}
	}
	return false
}

// inPlaceFiles are the files a sed or perl edit in place rewrites: its
// operands, less the script when no -e or -f option gives it.
func inPlaceFiles(args []string) []string {
	var operands []string
	scriptGiven := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-e" || a == "-f" || a == "--expression" || a == "--file" || (strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && strings.HasSuffix(a, "e") && len(a) > 2):
			scriptGiven = true
			i++
		case strings.HasPrefix(a, "--expression=") || strings.HasPrefix(a, "--file="):
			scriptGiven = true
		case strings.HasPrefix(a, "-"):
		default:
			operands = append(operands, a)
		}
	}
	if !scriptGiven && len(operands) > 0 {
		return operands[1:]
	}
	return operands
}

// findChanges reports whether a find command deletes files or runs a writer
// on them.
func findChanges(args []string) bool {
	for i, a := range args {
		if a == "-delete" {
			return true
		}
		if (a == "-exec" || a == "-execdir" || a == "-ok" || a == "-okdir") && i+1 < len(args) {
			name := filepath.Base(args[i+1])
			if fileWriters[name] || ((name == "sed" || name == "perl") && hasInPlace(args[i+2:])) {
				return true
			}
		}
	}
	return false
}

// gitWork reports whether a git command changes the working tree, the
// history or a remote.
func gitWork(words []string, cwd string) bool {
	_, sub, _ := gitInvocation(words, cwd)
	if sub >= len(words) {
		return false
	}
	name, rest := words[sub], words[sub+1:]
	operand, listing := false, false
	for _, a := range rest {
		if gitListFlags[a] {
			listing = true
		}
		if !strings.HasPrefix(a, "-") {
			operand = true
		}
	}
	switch name {
	case "stash":
		return len(rest) == 0 || (rest[0] != "list" && rest[0] != "show")
	case "branch", "tag":
		for _, a := range rest {
			switch a {
			case "-d", "-D", "--delete", "-m", "-M", "--move", "-c", "-C", "--copy", "-f", "--force":
				return true
			}
		}
		return operand && !listing
	case "worktree":
		return len(rest) > 0 && rest[0] != "list"
	}
	return gitAlwaysWork[name]
}

func containsField(fields []string, want string) bool {
	for _, f := range fields {
		if f == want {
			return true
		}
	}
	return false
}

// shellAssignment is a `NAME=value` assignment where a command can start.
var shellAssignment = regexp.MustCompile(`(?:^|[;&|(\n]\s*)([A-Za-z_][A-Za-z0-9_]*)=("[^"\n]*"|'[^'\n]*'|[^\s;&|)"']+)`)

// maxExpandedCommand bounds the size of a command after expansion: a chain of
// assignments that each repeat the one before doubles it at every step.
const maxExpandedCommand = 64 << 10

// expandAssignments replaces `$NAME` and `${NAME}` after a `NAME=value`
// assignment of the same command with its value, so a path built from a
// variable matches as written. A value made by a substitution is left as is. A
// command that grows past maxExpandedCommand, or starts past it, is returned
// unexpanded.
func expandAssignments(command string) string {
	if len(command) > maxExpandedCommand {
		return command
	}
	expanded := command
	for from := 0; from < len(expanded); {
		loc := shellAssignment.FindStringSubmatchIndex(expanded[from:])
		if loc == nil {
			break
		}
		name := expanded[from+loc[2] : from+loc[3]]
		value := strings.Trim(expanded[from+loc[4]:from+loc[5]], "\"'")
		end := from + loc[1]
		if !strings.Contains(value, "$(") && !strings.Contains(value, "`") {
			rest := strings.NewReplacer("${"+name+"}", value, "$"+name, value).Replace(expanded[end:])
			if end+len(rest) > maxExpandedCommand {
				return command
			}
			expanded = expanded[:end] + rest
		}
		from = end
	}
	return expanded
}
