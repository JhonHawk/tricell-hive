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

// stagesAll reports whether a `git add` takes every change: -A, --all or `.`.
func stagesAll(args []string) bool {
	for _, a := range args {
		if a == "-A" || a == "--all" || a == "." {
			return true
		}
	}
	return false
}

// commitsAll reports whether a `git commit` takes every tracked change: -a, in
// a short-flag cluster or alone, or --all.
func commitsAll(args []string) bool {
	for _, a := range args {
		if a == "--all" || (strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && strings.Contains(a, "a") && !strings.HasPrefix(a, "-m")) {
			return true
		}
	}
	return false
}

// folderCoverage walks the commands of a session once and says, per folder,
// whether it was committed at or after its last content write: strong when a
// commit names the folder, when it follows a `git add` naming it that no commit
// of the same repository has used yet (the add may be in an earlier command), or
// when `tasks.py <folder> versioned` exited 0; weak when only a `git add -A`,
// `git add .` or `git commit -a` of the same repository could have taken it,
// which the audit cannot tell. A repository is the directory a command resolves
// to: the `-C` or `--repo` directory, else the directory of the last `cd` of the
// command, else cwd. A commit uses up the staged state of its repository only.
func folderCoverage(events []traceEvent, since map[string]int, cwd string) (strong, weak map[string]bool) {
	strong, weak = map[string]bool{}, map[string]bool{}
	staged := map[string]map[string]int{}
	weakAt := map[string]int{}
	for i, e := range events {
		command, ok := resolvedCommand(e)
		if !ok {
			continue
		}
		dir := cwd
		for _, args := range shellSegments(command, 0) {
			if len(args) >= 2 && args[0] == "cd" {
				dir = resolveDir(dir, args[1])
				continue
			}
			isGit, sub, gitDir := gitInvocation(args, dir)
			if isGit && sub < len(args) && args[sub] == "add" {
				repo := cleanDir(gitDir)
				for f := range since {
					if namesFolder(args, f) {
						if staged[repo] == nil {
							staged[repo] = map[string]int{}
						}
						staged[repo][f] = i
					}
				}
				if stagesAll(args) {
					weakAt[repo] = i
				}
				continue
			}
			repo := cleanDir(gitDir)
			isCommit, all := false, false
			if isGit && sub < len(args) && args[sub] == "commit" {
				isCommit, all = true, commitsAll(args[sub+1:])
			}
			for k, a := range args {
				switch filepath.Base(a) {
				case "git_state.py":
					if k+1 < len(args) && args[k+1] == "commit" {
						isCommit, repo = true, cleanDir(dir)
						for m, v := range args {
							if v == "--repo" && m+1 < len(args) {
								repo = cleanDir(resolveDir(dir, args[m+1]))
							}
						}
					}
				case "tasks.py":
					if k+2 < len(args) && args[k+2] == "versioned" {
						for f, from := range since {
							if i >= from && namesFolder(args[k+1:k+2], f) {
								strong[f] = true
							}
						}
					}
				}
			}
			if !isCommit {
				continue
			}
			for f, from := range since {
				at, wasStaged := staged[repo][f]
				w, hasWeak := weakAt[repo]
				switch {
				case i < from:
				case namesFolder(args, f) || (wasStaged && at >= from):
					strong[f] = true
				case all || (hasWeak && w >= from):
					weak[f] = true
				}
			}
			delete(staged, repo)
			delete(weakAt, repo)
		}
	}
	return strong, weak
}

// resolveDir resolves a directory argument against the current one; $PWD is the
// current one.
func resolveDir(cwd, dir string) string {
	switch dir {
	case "$PWD", "${PWD}", ".":
		return cwd
	}
	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}
	return filepath.Join(cwd, dir)
}

// cleanDir normalizes a repository directory so equal paths compare equal.
func cleanDir(dir string) string {
	return filepath.Clean(dir)
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
// tool or a command of a known writing form. A folder passes when a successful
// commit follows its last content write before the last merge and names it, or
// follows a `git add` that did, in any earlier command, or when
// `tasks.py <folder> versioned` exited 0 after that write. A broad `git add -A`,
// `git add .` or `git commit -a` makes it a review. Without any of them,
// assistant text after the last merge is a fail naming the folder; no such
// text, or a folder whose only writes are tasks.py progress records after the
// last merge (it came versioned from another session and its edits wait for the
// archive), is a review. A session that wrote no change folder or merged
// nothing is not observed.
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
	first, onlyLateProgress, since := map[string]int{}, map[string]bool{}, map[string]int{}
	for _, folder := range order {
		ws := writes[folder]
		first[folder], onlyLateProgress[folder] = ws[0].index, true
		for _, w := range ws {
			if w.index < first[folder] {
				first[folder] = w.index
			}
			if !w.progress || w.index < merge {
				onlyLateProgress[folder] = false
			}
		}
		// The commit must follow the last content write before the merge,
		// or the first write when the folder has only progress records.
		since[folder] = first[folder]
		for _, w := range ws {
			if !w.progress && w.index < merge && w.index > since[folder] {
				since[folder] = w.index
			}
		}
	}
	strong, weak := folderCoverage(events, since, s.Cwd)
	var failed, review, passed []string
	for _, folder := range order {
		line := fmt.Sprintf("stdout.jsonl:%d %s %s", events[first[folder]].Line, events[first[folder]].Kind, folder)
		switch {
		case strong[folder]:
			passed = append(passed, line+" committed after its last write")
		case weak[folder]:
			review = append(review, line+" only a broad add or commit -a could have taken it")
		case onlyLateProgress[folder]:
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
