package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// changeFolderSegment is the `changes/<id>` segment of a path; the id
	// `archive` is the archive folder, not a change.
	changeFolderSegment = regexp.MustCompile("changes/([^/\\s'\"`\\\\;&|<>()]+)")
	// teeTarget is the file a `tee` writes.
	teeTarget = regexp.MustCompile(`\btee\s+(?:-a\s+)?["']?([^\s"';&|<>]+)`)
	// scriptedWrite is a Python call that writes a file: a Path method or open
	// with a write or append mode.
	scriptedWrite = regexp.MustCompile(`\.write_(?:text|bytes)\(|\bopen\([^)]*,\s*["'][wa]`)
)

// changeFolderIDs returns the `changes/<id>` segments a text names, outside
// the archive.
func changeFolderIDs(text string) []string {
	var ids []string
	for _, m := range changeFolderSegment.FindAllStringSubmatch(text, -1) {
		if m[1] != "archive" {
			ids = append(ids, "changes/"+m[1])
		}
	}
	return ids
}

// changeFolderWrite is one write to a change folder.
type changeFolderWrite struct {
	index    int
	progress bool // a tasks.py done, checkpoint or execution record
}

// resolvedCommand returns the command of an event whose run the session
// resolved and whose success is known: Codex's completed command item with its
// exit code, or a shell call of any host that reported success. A command whose
// success is unknown is not one; the two sources are not paired, since a
// command counted twice changes no verdict.
func resolvedCommand(e traceEvent) (string, bool) {
	if (e.Kind != "shell" && e.Kind != "exec_item") || e.Success == nil || !*e.Success || e.Command == "" {
		return "", false
	}
	return expandAssignments(e.Command), true
}

// shellAssignment is a `NAME=value` assignment where a command can start.
var shellAssignment = regexp.MustCompile(`(?:^|[;&|(\n]\s*)([A-Za-z_][A-Za-z0-9_]*)=("[^"\n]*"|'[^'\n]*'|[^\s;&|)"']+)`)

// expandAssignments replaces `$NAME` and `${NAME}` after a `NAME=value`
// assignment of the same command with its value, so a path built from a
// variable matches as written. A value made by a substitution is left as is.
func expandAssignments(command string) string {
	for from := 0; from < len(command); {
		loc := shellAssignment.FindStringSubmatchIndex(command[from:])
		if loc == nil {
			break
		}
		name := command[from+loc[2] : from+loc[3]]
		value := strings.Trim(command[from+loc[4]:from+loc[5]], "\"'")
		end := from + loc[1]
		if !strings.Contains(value, "$(") && !strings.Contains(value, "`") {
			rest := strings.NewReplacer("${"+name+"}", value, "$"+name, value).Replace(command[end:])
			command = command[:end] + rest
		}
		from = end
	}
	return command
}

// folderRoot matches a path that is a change folder itself.
var folderRoot = regexp.MustCompile(`changes/([^/]+)/?$`)

// folderRenames returns the change folders a command moves with mv or git mv,
// source and destination, resolving relative paths against a preceding cd of
// the same command. A move into the archive, or of anything but a change
// folder, is not one.
func folderRenames(command string) [][2]string {
	var moves [][2]string
	cwd := ""
	resolve := func(p string) string {
		if filepath.IsAbs(p) {
			return filepath.Clean(p)
		}
		return filepath.Join(cwd, p)
	}
	for _, args := range shellSegments(command, 0) {
		if len(args) >= 2 && args[0] == "cd" {
			cwd = resolve(args[1])
			continue
		}
		var operands []string
		if isGit, sub, _ := gitInvocation(args, cwd); isGit && sub < len(args) && args[sub] == "mv" {
			operands = args[sub+1:]
		} else if len(args) > 0 && args[0] == "mv" {
			operands = args[1:]
		} else {
			continue
		}
		var paths []string
		for _, a := range operands {
			if !strings.HasPrefix(a, "-") {
				paths = append(paths, a)
			}
		}
		if len(paths) != 2 {
			continue
		}
		src, dst := folderRoot.FindStringSubmatch(resolve(paths[0])), folderRoot.FindStringSubmatch(resolve(paths[1]))
		if src != nil && dst != nil && src[1] != "archive" && dst[1] != "archive" && src[1] != dst[1] {
			moves = append(moves, [2]string{"changes/" + src[1], "changes/" + dst[1]})
		}
	}
	return moves
}

// commandWrites returns the change folders a command writes by a known form
// and whether the write is only a tasks.py progress record: tasks.py done,
// checkpoint or execution, a redirection or tee, and a command that names the
// folder and writes through a Python call. A read names the folder without
// writing it and is not one.
func commandWrites(command string) (content, progress []string) {
	for _, args := range shellSegments(command, 0) {
		for i, a := range args {
			if filepath.Base(a) != "tasks.py" || i+2 >= len(args) {
				continue
			}
			switch args[i+2] {
			case "done", "checkpoint", "execution":
				progress = append(progress, changeFolderIDs(args[i+1])...)
			}
		}
	}
	for _, m := range redirectPath.FindAllStringSubmatch(command, -1) {
		content = append(content, changeFolderIDs(m[1])...)
	}
	for _, m := range teeTarget.FindAllStringSubmatch(command, -1) {
		content = append(content, changeFolderIDs(m[1])...)
	}
	if scriptedWrite.MatchString(command) {
		content = append(content, changeFolderIDs(command)...)
	}
	return content, progress
}

// namesFolder reports whether any argument of a segment other than a commit
// message names the folder.
func namesFolder(args []string, folder string) bool {
	for i, a := range args {
		if i > 0 && (args[i-1] == "-m" || args[i-1] == "--message" || args[i-1] == "-F") {
			continue
		}
		for _, id := range changeFolderIDs(a) {
			if id == folder {
				return true
			}
		}
	}
	return false
}

// commitsFolder reports whether a command that succeeded commits the folder:
// a `git commit` or git_state.py commit segment that names it, or that follows
// a `git add` segment of the same command that names it, or a
// `tasks.py <folder> versioned`. A commit of other paths in the same command as
// a tasks.py record is not one.
func commitsFolder(command, folder string) bool {
	added := false
	for _, args := range shellSegments(command, 0) {
		isGit, sub, _ := gitInvocation(args, "")
		if isGit && sub < len(args) {
			switch args[sub] {
			case "add":
				added = added || namesFolder(args, folder)
			case "commit":
				if added || namesFolder(args, folder) {
					return true
				}
			}
		}
		for i, a := range args {
			switch filepath.Base(a) {
			case "git_state.py":
				if i+1 < len(args) && args[i+1] == "commit" && (added || namesFolder(args, folder)) {
					return true
				}
			case "tasks.py":
				if i+2 < len(args) && args[i+2] == "versioned" && namesFolder(args[i+1:i+2], folder) {
					return true
				}
			}
		}
	}
	return false
}

// lastMergeIndex returns the index of the last event that completes a merge,
// whether or not it deletes the branch: a successful `gh pr merge`, or a
// successful `git merge` followed by a successful `git push` that names base;
// -1 when there is none.
func lastMergeIndex(events []traceEvent, base string) int {
	last := -1
	pendingGitMerge := false
	for i, e := range events {
		command, ok := resolvedCommand(e)
		if !ok {
			continue
		}
		for _, args := range shellSegments(command, 0) {
			if isGhPrSubcommand(args, "merge") {
				last = i
				continue
			}
			isGit, sub, _ := gitInvocation(args, "")
			if !isGit || sub >= len(args) {
				continue
			}
			switch args[sub] {
			case "merge":
				pendingGitMerge = true
			case "push":
				if pendingGitMerge && pushTargetsBase(args[sub+1:], base) {
					last = i
					pendingGitMerge = false
				}
			}
		}
	}
	return last
}

// changeFolderUncommitted checks that a close which merged left the change
// folder it wrote versioned. The folders are the `changes/<id>` segments
// outside the archive of the paths the session wrote through a patch, a write
// tool or a command of a known writing form. After the first write to a folder,
// a successful commit command that names it, or `tasks.py <folder> versioned`
// that exited 0, is a pass. Without either, assistant text after the last merge
// is a fail naming the folder; no such text, or a folder whose only writes are
// tasks.py progress records after the last merge (it came versioned from
// another session and its edits wait for the archive), is a review. A session
// that wrote no change folder or merged nothing is not observed.
func changeFolderUncommitted(s auditSession, base string) criterionAssessment {
	c := criterionAssessment{Criterion: "change_folder_uncommitted", Status: "not_observed", Evidence: []string{}}
	events := s.Trace.Events
	merge := lastMergeIndex(events, base)
	writes := map[string][]changeFolderWrite{}
	var order []string
	note := func(folder string, w changeFolderWrite) {
		if len(writes[folder]) == 0 {
			order = append(order, folder)
		}
		writes[folder] = append(writes[folder], w)
	}
	for i, e := range events {
		switch e.Kind {
		case "write", "write_item", "delete":
			for _, id := range changeFolderIDs(e.Path) {
				note(id, changeFolderWrite{index: i})
			}
		case "shell", "exec_item":
			command, ok := resolvedCommand(e)
			if !ok {
				continue
			}
			content, progress := commandWrites(command)
			for _, id := range content {
				note(id, changeFolderWrite{index: i})
			}
			for _, id := range progress {
				note(id, changeFolderWrite{index: i, progress: true})
			}
			for _, mv := range folderRenames(command) {
				moved := writes[mv[0]]
				delete(writes, mv[0])
				for k, f := range order {
					if f == mv[0] {
						order = append(order[:k], order[k+1:]...)
						break
					}
				}
				for _, w := range moved {
					note(mv[1], w)
				}
				note(mv[1], changeFolderWrite{index: i})
			}
		}
	}
	if merge < 0 || len(order) == 0 {
		return c
	}
	textAfterMerge := false
	for _, e := range events[merge+1:] {
		if e.Kind == "text" && e.Role == "assistant" && strings.TrimSpace(e.Text) != "" {
			textAfterMerge = true
			break
		}
	}
	var failed, review, passed []string
	for _, folder := range order {
		ws := writes[folder]
		first, onlyLateProgress := ws[0].index, true
		for _, w := range ws {
			if w.index < first {
				first = w.index
			}
			if !w.progress || w.index < merge {
				onlyLateProgress = false
			}
		}
		covered := false
		for _, e := range events[first+1:] {
			if command, ok := resolvedCommand(e); ok && commitsFolder(command, folder) {
				covered = true
				break
			}
		}
		line := fmt.Sprintf("stdout.jsonl:%d %s %s", events[first].Line, events[first].Kind, folder)
		switch {
		case covered:
			passed = append(passed, line+" committed after the first write")
		case onlyLateProgress:
			review = append(review, line+" only progress records after the last merge")
		case textAfterMerge:
			failed = append(failed, line+" not committed before the close report")
		default:
			review = append(review, line+" no assistant text after the last merge")
		}
	}
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(review) > 0:
		c.Status, c.Evidence = "review", review
	default:
		c.Status, c.Evidence = "pass", passed
	}
	return c
}
