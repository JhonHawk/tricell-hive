package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// full_suite_after_text_fix and review_rounds_per_candidate read the shell
// commands, edits and review launches of a session in order, without running
// anything. Both resolve a doubt toward review or pass, never toward fail: a
// write whose target cannot be resolved, or a first suite run that cannot be
// shown to pass, is a review.

var (
	textExtensions = map[string]bool{".md": true, ".mdx": true, ".txt": true}
	// A run is shown to pass by a PIPESTATUS line whose first status is 0 or
	// an exit=0 line in its output.
	statusUse   = regexp.MustCompile(`\$\{PIPESTATUS\[0\]\}|\$\?|\$\{\?\}`)
	patchedFile = regexp.MustCompile(`(?m)^\*\*\* (?:Update|Add|Delete) File: (.+)$|^\*\*\* Move to: (.+)$`)
	typedReview = regexp.MustCompile(`(?:^|[\s>])/(?:[a-z0-9-]+:)?code-review\b`)
	// goValueFlags take their value as the next word.
	goValueFlags  = map[string]bool{"-count": true, "-timeout": true, "-parallel": true, "-p": true, "-cpu": true, "-tags": true, "-ldflags": true, "-coverprofile": true, "-covermode": true, "-coverpkg": true, "-shuffle": true, "-vet": true, "-o": true, "-gcflags": true, "-mod": true}
	goFilterFlags = map[string]bool{"-run": true, "-skip": true, "-bench": true, "-fuzz": true, "-list": true}
	pytestFilters = map[string]bool{"-k": true, "-m": true, "--lf": true, "--ff": true, "--deselect": true, "--ignore": true, "--co": true, "--collect-only": true}
	// interpreterWrite marks a script that writes a file; interpreters are the
	// commands that run one.
	interpreterWrite = regexp.MustCompile(`write_text|write_bytes|\.write\(|writeFile|File\.write|open\([^)]*,\s*["'][wa]|shutil\.|os\.(?:rename|remove|replace|unlink)|\.unlink\(|\.rename\(`)
	interpreters     = map[string]bool{"python": true, "python3": true, "node": true, "ruby": true, "deno": true, "bun": true}
	// treeWriters change the working tree when called with the flag they name.
	treeWriters = map[string]string{"gofmt": "-w", "goimports": "-w", "prettier": "--write", "eslint": "--fix", "ruff": "--fix"}
)

// rerunItem is one step of a session that matters to a criterion below, in
// the order the session did it.
type rerunItem struct {
	// kind: run (an unfiltered suite run), text (a change to text files only),
	// other (a change to anything else, a rebase, merge, pull or a switch),
	// unknown (a write whose target cannot be resolved), boundary (a push, a
	// merge of a pull request, a new branch or a new worktree).
	kind string
	// branch marks a change of branch, which restarts the review count.
	branch bool
	// event is the index of the event that holds the step.
	event int
	// dir is the working folder of a run, "" when it is not known.
	dir string
	// plain: a run whose success is the suite's own, with no pipe or
	// background around it.
	plain bool
	// mark is the text the command prints to show the status of this run: its
	// echo of $? right after it, or of ${PIPESTATUS[0]} after its pipeline,
	// with the status replaced by 0.
	mark string
}

func fullSuiteAfterTextFix(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "full_suite_after_text_fix", Status: "not_observed", Evidence: []string{}}
	events := s.Trace.Events
	var failed, review, passed []string
	prev := -1
	var prevItem rerunItem
	other, unknown := false, false
	for _, it := range rerunItems(s) {
		switch it.kind {
		case "other":
			other = true
		case "unknown":
			unknown = true
		case "boundary":
			prev, other, unknown = -1, false, false
		case "run":
			if prev >= 0 && it.dir != "" && it.dir == prevItem.dir {
				evidence := fmt.Sprintf("%s after the run at line %d", regressionEvidence(events[it.event]), events[prevItem.event].Line)
				switch {
				case other:
					passed = append(passed, evidence)
				case unknown:
					review = append(review, evidence+" with a write whose target is unknown")
				case !runShownToPass(s, prevItem):
					review = append(review, evidence+" and the first run is not shown to pass")
				default:
					failed = append(failed, evidence)
				}
			}
			prev, prevItem, other, unknown = it.event, it, false, false
		}
	}
	return rerunStatus(c, failed, review, passed)
}

// rerunStatus: a fail wins, then a review, then a pass.
func rerunStatus(c criterionAssessment, failed, review, passed []string) criterionAssessment {
	if len(failed) == 0 && len(review) > 0 {
		c.Status, c.Evidence = "review", review
		return c
	}
	return statusOf(c, failed, passed)
}

// runShownToPass reports whether a run's own status shows it passed: the
// success of a command with no pipe or background around the suite, or the
// line its status echo printed with 0. A run launched in the background, or
// with no result, is not shown: the host only reports that it started.
func runShownToPass(s auditSession, run rerunItem) bool {
	events := s.Trace.Events
	if runsInBackground(s.Host, events[run.event]) {
		return false
	}
	answered := false
	for _, e := range events[run.event+1:] {
		if e.Kind == "tool_result" && e.ID == events[run.event].ID {
			answered = true
			break
		}
	}
	if !answered {
		return false
	}
	if run.plain && succeeded(events, run.event) {
		return true
	}
	if run.mark == "" {
		return false
	}
	for _, line := range strings.Split(resultText(events, run.event), "\n") {
		line = strings.TrimSpace(line)
		if line == run.mark || len(run.mark) > 1 && strings.Contains(line, run.mark) {
			return true
		}
	}
	return false
}

// rerunItems lists, in order, the suite runs, changes and boundaries of a
// session. The working folder carries from one shell call to the next.
func rerunItems(s auditSession) []rerunItem {
	var items []rerunItem
	cwd := s.Cwd
	for i, e := range s.Trace.Events {
		switch e.Kind {
		case "write", "delete":
			if e.Success != nil && !*e.Success {
				continue
			}
			for _, kind := range editKinds(e, cwd) {
				items = append(items, rerunItem{kind: kind, event: i})
			}
		case "shell":
			var more []rerunItem
			more, cwd = shellItems(e.Command, cwd)
			for _, it := range more {
				it.event = i
				items = append(items, it)
			}
		}
	}
	return items
}

// editKinds classifies what an edit tool changes: one kind per file.
func editKinds(e traceEvent, cwd string) []string {
	paths := []string{e.Path}
	if e.Path == "" {
		paths = nil
		for _, m := range patchedFile.FindAllStringSubmatch(e.Text, -1) {
			paths = append(paths, strings.TrimSpace(m[1]+m[2]))
		}
		if len(paths) == 0 {
			return []string{"unknown"}
		}
	}
	var kinds []string
	for _, p := range paths {
		if k := pathKind(p, cwd); k != "" {
			kinds = append(kinds, k)
		}
	}
	return kinds
}

// pathKind classifies a path a change writes: "" for a file in the host's
// temporary location, text, other, or unknown when it cannot be resolved.
func pathKind(p, cwd string) string {
	switch {
	case p == "":
		return "unknown"
	case !strings.ContainsAny(p, "`") && !outsideTemp(p, cwd):
		return ""
	case strings.ContainsAny(p, "$`") || strings.HasPrefix(p, "~") || interpHole.MatchString(p):
		return "unknown"
	case textExtensions[strings.ToLower(filepath.Ext(p))]:
		return "text"
	}
	return "other"
}

// shellItems reads one shell command, through shellSegments and without
// counting text inside arguments or heredocs, and returns its steps and the
// working folder after it ("" when it is not known).
func shellItems(command, cwd string) ([]rerunItem, string) {
	var items, scripted []rerunItem
	literals := pythonWritePath.FindAllStringSubmatch(command, -1)
	for _, p := range literals {
		if k := pathKind(p[1], cwd); k != "" {
			scripted = append(scripted, rerunItem{kind: k})
		}
	}
	body := expandAssignments(inlineShellHeredocs(command))
	segments := shellSegments(body, 0)
	if segments == nil {
		return append(items, rerunItem{kind: "unknown"}), ""
	}
	segmentOf := map[int]int{}
	for si, segment := range segments {
		for _, args := range splitBackground(segment) {
			var more []rerunItem
			more, cwd = simpleItems(args, cwd)
			for _, it := range more {
				if it.kind == "run" {
					segmentOf[len(items)] = si
				}
				items = append(items, it)
			}
		}
	}
	if !hasBackground(body) {
		for i, si := range segmentOf {
			items[i].mark = statusMark(segments, si)
		}
	}
	if runsInterpreter(segments) && len(interpreterWrite.FindAllString(command, -1)) > len(literals) {
		scripted = append(scripted, rerunItem{kind: "unknown"})
	}
	// A script's writes cannot be placed among the segments, so a command that
	// also runs the suite is unknown on both sides of the run.
	if len(scripted) > 0 {
		hasRun := false
		for _, it := range items {
			hasRun = hasRun || it.kind == "run"
		}
		if hasRun {
			items = append(append([]rerunItem{{kind: "unknown"}}, items...), rerunItem{kind: "unknown"})
		} else {
			items = append(scripted, items...)
		}
	}
	tail := lastStatement(body)
	pipelined := hasPipeOrBackground(body)
	tailRuns := 0
	for _, segment := range shellSegments(tail, 0) {
		if _, ok := suiteRun(segment, cwd); ok {
			tailRuns++
		}
	}
	for i := len(items) - 1; i >= 0 && tailRuns > 0; i-- {
		if items[i].kind == "run" {
			items[i].plain = !pipelined
			tailRuns--
		}
	}
	return items, cwd
}

// echoMark returns what an echo or printf segment prints for a status, with
// the status replaced by 0, or "" when the segment prints none.
func echoMark(segment []string) string {
	if name := filepath.Base(segment[0]); name != "echo" && name != "printf" {
		return ""
	}
	var words []string
	for _, w := range segment[1:] {
		if m := redirectWord.FindStringSubmatch(w); m != nil && w != m[1] {
			break
		}
		words = append(words, w)
	}
	text := strings.Join(words, " ")
	if !statusUse.MatchString(text) {
		return ""
	}
	return strings.TrimSpace(statusUse.ReplaceAllString(text, "0"))
}

// statusMark returns what the command prints to show the status of the suite
// run in segment si: an echo or printf of $? in the next segment, or of
// ${PIPESTATUS[0]} in a later one when the suite is piped, with the status
// replaced by 0. It is "" when no echo is tied to the run, or when another
// status echo of the command prints the same text, so the output could not
// tell which command it came from.
func statusMark(segments [][]string, si int) string {
	mark := ""
	for next := si + 1; next < len(segments) && mark == ""; next++ {
		if echoMark(segments[next]) == "" {
			continue
		}
		if next > si+1 && !strings.Contains(strings.Join(segments[next], " "), "PIPESTATUS") {
			return ""
		}
		mark = echoMark(segments[next])
	}
	if mark == "" {
		return ""
	}
	same := 0
	for _, segment := range segments {
		if echoMark(segment) == mark {
			same++
		}
	}
	if same > 1 {
		return ""
	}
	return mark
}

func runsInterpreter(segments [][]string) bool {
	for _, segment := range segments {
		for i := 0; i < len(segment) && i < 3; i++ {
			if interpreters[filepath.Base(segment[i])] {
				return true
			}
		}
	}
	return false
}

func simpleItems(args []string, cwd string) ([]rerunItem, string) {
	var items []rerunItem
	var words []string
	for i := 0; i < len(args); i++ {
		a := strings.TrimRight(strings.TrimLeft(args[i], "({"), ")}")
		if a == "" {
			continue
		}
		if a == "<" || strings.HasPrefix(a, "<") && !strings.HasPrefix(a, "<<") {
			if a == "<" {
				i++
			}
			continue
		}
		if m := redirectWord.FindStringSubmatch(a); m != nil && a != m[1] {
			target := m[1]
			if target == "" && i+1 < len(args) {
				i++
				target = args[i]
			}
			if target != "" && !strings.HasPrefix(target, "&") {
				if k := pathKind(target, cwd); k != "" {
					items = append(items, rerunItem{kind: k})
				}
			}
			continue
		}
		words = append(words, a)
	}
	xargs := false
	for len(words) > 0 {
		switch name := filepath.Base(words[0]); {
		case strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-"):
			words = words[1:]
		case shellWrappers[name]:
			xargs = xargs || name == "xargs"
			words = words[1:]
			for len(words) > 0 && (strings.HasPrefix(words[0], "-") || strings.Contains(words[0], "=")) {
				words = words[1:]
			}
		default:
			more, next := namedItems(words, cwd, xargs)
			return append(items, more...), next
		}
	}
	return items, cwd
}

func namedItems(words []string, cwd string, xargs bool) ([]rerunItem, string) {
	name, args := filepath.Base(words[0]), words[1:]
	if interpHole.MatchString(words[0]) {
		return []rerunItem{{kind: "unknown"}}, cwd
	}
	if dir, ok := suiteRun(words, cwd); ok {
		return []rerunItem{{kind: "run", dir: dir}}, cwd
	}
	targets := func(paths []string) []rerunItem {
		var items []rerunItem
		for _, p := range paths {
			if xargs {
				return []rerunItem{{kind: "unknown"}}
			}
			if k := pathKind(p, cwd); k != "" {
				items = append(items, rerunItem{kind: k})
			}
		}
		return items
	}
	switch {
	case name == "cd":
		return nil, nextFolder(args, cwd)
	case fileWriters[name]:
		return targets(writerTargets(name, args)), cwd
	case (name == "sed" || name == "perl") && hasInPlace(args):
		return targets(inPlaceFiles(args)), cwd
	case name == "find" && findChanges(args):
		return []rerunItem{{kind: "unknown"}}, cwd
	case name == "git":
		return gitItems(words, cwd), cwd
	case name == "gh":
		if len(args) >= 2 && args[0] == "pr" && args[1] == "merge" {
			return []rerunItem{{kind: "boundary"}}, cwd
		}
		return nil, cwd
	case name == "patch" || name == "black" && !containsField(args, "--check") && !containsField(args, "--diff"):
		return []rerunItem{{kind: "unknown"}}, cwd
	case treeWriters[name] != "" && containsField(args, treeWriters[name]):
		return []rerunItem{{kind: "unknown"}}, cwd
	case name == "go" && len(args) > 0 && (args[0] == "generate" || args[0] == "fix" || args[0] == "get" || args[0] == "mod"):
		return []rerunItem{{kind: "unknown"}}, cwd
	}
	return nil, cwd
}

// gitItems reads a git command: a push is a boundary, a commit or an add
// changes nothing, a new branch or worktree is a boundary, a rebase, merge,
// pull, checkout, switch, cherry-pick, am, apply, reset or restore is another
// change, and the rest that rewrite the tree have a target that is not shown.
func gitItems(words []string, cwd string) []rerunItem {
	_, sub, _ := gitInvocation(words, cwd)
	if sub >= len(words) {
		return nil
	}
	name, rest := words[sub], words[sub+1:]
	switch name {
	case "push":
		return []rerunItem{{kind: "boundary"}}
	case "checkout", "switch":
		for _, a := range rest {
			switch a {
			case "-b", "-B", "-c", "-C", "--create", "--force-create", "--orphan", "--detach", "-d":
				return []rerunItem{{kind: "boundary", branch: true}}
			}
		}
		return []rerunItem{{kind: "other", branch: true}}
	case "rebase", "merge", "pull", "cherry-pick", "am", "apply", "reset", "restore":
		return []rerunItem{{kind: "other"}}
	case "worktree":
		if len(rest) > 0 && rest[0] == "add" {
			return []rerunItem{{kind: "boundary", branch: true}}
		}
	case "branch":
		if gitWork(words, cwd) {
			for _, a := range rest {
				if strings.HasPrefix(a, "-") {
					return nil
				}
			}
			return []rerunItem{{kind: "boundary", branch: true}}
		}
	case "stash", "revert", "rm", "mv", "clean":
		return []rerunItem{{kind: "unknown"}}
	}
	return nil
}

// suiteRun recognizes an unfiltered suite run and returns the working folder
// it runs in: "go test ./..." (flags such as -count and -race allowed, -run
// and a sub-package pattern filter it), "npm test", "pnpm test" and "pytest"
// with no argument.
func suiteRun(words []string, cwd string) (string, bool) {
	if len(words) == 0 {
		return "", false
	}
	for len(words) > 0 && (shellWrappers[filepath.Base(words[0])] || strings.Contains(words[0], "=") && !strings.HasPrefix(words[0], "-")) {
		words = words[1:]
		for len(words) > 0 && strings.HasPrefix(words[0], "-") {
			words = words[1:]
		}
	}
	if len(words) == 0 {
		return "", false
	}
	dir := cwd
	in := func(p string) {
		switch {
		case filepath.IsAbs(p):
			dir = filepath.Clean(p)
		case dir != "":
			dir = filepath.Join(dir, p)
		}
	}
	name, args := filepath.Base(words[0]), words[1:]
	switch name {
	case "go":
		if len(args) >= 2 && args[0] == "-C" {
			in(args[1])
			args = args[2:]
		}
		if len(args) == 0 || args[0] != "test" {
			return "", false
		}
		pattern := false
		for i := 1; i < len(args); i++ {
			flag, _, _ := strings.Cut(args[i], "=")
			switch {
			case goFilterFlags[flag]:
				return "", false
			case strings.HasPrefix(args[i], "-"):
				if goValueFlags[args[i]] {
					i++
				}
			case args[i] == "./...":
				pattern = true
			default:
				return "", false
			}
		}
		return dir, pattern
	case "npm", "pnpm":
		for len(args) >= 2 && (args[0] == "--dir" || args[0] == "-C" || args[0] == "--prefix") {
			in(args[1])
			args = args[2:]
		}
		return dir, len(args) == 1 && args[0] == "test"
	case "pytest", "python", "python3":
		if name != "pytest" {
			if len(args) < 2 || args[0] != "-m" || args[1] != "pytest" {
				return "", false
			}
			args = args[2:]
		}
		for _, a := range args {
			flag, _, _ := strings.Cut(a, "=")
			if !strings.HasPrefix(a, "-") || pytestFilters[flag] {
				return "", false
			}
		}
		return dir, true
	}
	return "", false
}

// lastStatement is the text after the last unquoted ";" or newline.
func lastStatement(command string) string {
	var quote rune
	escaped := false
	cut := 0
	for i, ch := range command {
		switch {
		case escaped:
			escaped = false
		case ch == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if ch == quote {
				quote = 0
			}
		case ch == '\'' || ch == '"':
			quote = ch
		case ch == ';' || ch == '\n':
			if strings.TrimSpace(command[i+1:]) != "" {
				cut = i + 1
			}
		}
	}
	return command[cut:]
}

// hasBackground reports whether a command sends something to the background
// with an unquoted "&".
func hasBackground(command string) bool {
	return hasPipeOrBackground(strings.NewReplacer("||", " ", "|", " ").Replace(command))
}

// hasPipeOrBackground reports whether a command has an unquoted "|", "||" or
// a "&" that is neither "&&" nor part of a redirection, so the exit status of
// the suite run is not the one of the whole command.
func hasPipeOrBackground(command string) bool {
	var quote rune
	escaped := false
	chars := []rune(command)
	for i, ch := range chars {
		switch {
		case escaped:
			escaped = false
		case ch == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if ch == quote {
				quote = 0
			}
		case ch == '\'' || ch == '"':
			quote = ch
		case ch == '|':
			return true
		case ch == '&':
			prev, next := rune(0), rune(0)
			if i > 0 {
				prev = chars[i-1]
			}
			if i+1 < len(chars) {
				next = chars[i+1]
			}
			if next == '&' || prev == '&' || prev == '>' || next == '>' {
				continue
			}
			return true
		}
	}
	return false
}

// reviewRoundsPerCandidate fails when a main Claude Code session launches more
// than two code review rounds with no push, merge, new load of flow-build or
// flow-plan, or change of branch between them. The launches of one message are
// one round. A round after one the user typed as /code-review is a review: the
// user may have asked for it.
func reviewRoundsPerCandidate(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "review_rounds_per_candidate", Status: "not_observed", Evidence: []string{}}
	if s.Host != "claude" || child(s) {
		return c
	}
	events := s.Trace.Events
	resets := map[int]bool{}
	for _, it := range rerunItems(s) {
		if it.kind == "boundary" || it.branch {
			resets[it.event] = true
		}
	}
	type round struct {
		line  int
		typed bool
	}
	var rounds []round
	var failed, review []string
	seen := false
	counted := map[string]bool{}
	add := func(r round, label string) {
		seen = true
		rounds = append(rounds, r)
		if len(rounds) <= 2 {
			return
		}
		typed := false
		for _, p := range rounds {
			typed = typed || p.typed
		}
		evidence := fmt.Sprintf("stdout.jsonl:%d %s round %d since the last push, merge, flow load or branch change", r.line, label, len(rounds))
		if typed {
			review = append(review, evidence+", after a /code-review the user typed")
		} else {
			failed = append(failed, evidence)
		}
	}
	turn := -1
	for i, e := range events {
		if s.Turn[i] != turn {
			turn = s.Turn[i]
			if typedReview.MatchString(s.Prompts[turn]) {
				add(round{e.Line, true}, "typed /code-review")
			}
		}
		switch {
		case resets[i] || loadedFlow(e) == "flow-build" || loadedFlow(e) == "flow-plan":
			rounds = nil
		case reviewLaunch(e) && succeeded(events, i):
			if e.Message != "" {
				if counted[e.Message] {
					continue
				}
				counted[e.Message] = true
			}
			add(round{e.Line, false}, e.Tool)
		}
	}
	if !seen {
		return c
	}
	return rerunStatus(c, failed, review, []string{"two or fewer rounds per candidate"})
}

// reviewLaunch reports whether an event launches a code review round: the
// code-review skill or a hive-review-code child.
func reviewLaunch(e traceEvent) bool {
	switch {
	case e.Kind == "skill_invocation":
		return e.Text == "code-review" || strings.HasSuffix(e.Text, ":code-review")
	case e.Tool == "Agent" || e.Tool == "Task":
		return roleArgument(e) == "hive-review-code"
	}
	return false
}
