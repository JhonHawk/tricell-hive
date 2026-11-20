package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// waitLoopWord marks a script that may carry a shell loop.
	waitLoopWord = regexp.MustCompile(`\b(?:while|until|for)\b`)
	// nestedSleep is a sleep inside a quoted script, such as bash -c or ssh.
	nestedSleep = regexp.MustCompile(`(?:^|[\s;&|({])(?:\S*/)?sleep(?:\s|$)`)
	// Ways a script sets a clock or a counter. The forms that assign from an
	// expression count only when the expression reads the variable itself, the
	// clock or $SECONDS.
	clockAssignments   = regexp.MustCompile("^(\\w+)=(?:\\$\\(date\\s+\\+%s|[\"']?\\$\\{?SECONDS\\b)")
	stepAssignments    = regexp.MustCompile(`^\(\(\s*(?:(?:\+\+|--)(\w+)|(\w+)\s*(?:\+\+|--|[-+*/]=))`)
	exprAssignments    = []*regexp.Regexp{regexp.MustCompile(`^(\w+)=\$\(\((.*)`), regexp.MustCompile(`^(\w+)=\$\(expr\s(.*)`), regexp.MustCompile(`^\(\(\s*(\w+)\s*=([^=].*)`)}
	letAssignment      = regexp.MustCompile("^let\\s+[\"']?(?:\\+\\+|--)?(\\w+)(.*)")
	cStyleVariable     = regexp.MustCompile(`^\(\((\w+)`)
	elapsedOperand     = regexp.MustCompile(`\$\{?SECONDS\b|date \+%s`)
	elapsedBareOperand = regexp.MustCompile(`\bSECONDS\b`)
	// waitKeywords precede a command inside a compound command.
	waitKeywords = map[string]bool{"do": true, "then": true, "else": true, "elif": true, "if": true, "while": true, "until": true, "!": true, "{": true}
)

// runsInBackground reports whether a shell call was launched so the host does
// not wait for it: Claude Code's run_in_background, Grok's block_until_ms of 0.
func runsInBackground(host string, e traceEvent) bool {
	in := decodeInput(e)
	switch host {
	case "claude":
		background, _ := in["run_in_background"].(bool)
		return background
	case "grok":
		ms, ok := in["block_until_ms"].(float64)
		return ok && ms == 0
	}
	return false
}

// unboundedBackgroundWait fails when a shell command launched in the
// background holds a while, until or for loop with sleep and no deadline. It
// does not judge whether the loop's exit condition covers every outcome.
// Codex's unified exec does not mark which command runs in the background, so
// it reports not_observed.
func unboundedBackgroundWait(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "unbounded_background_wait", Status: "not_observed", Evidence: []string{}}
	if s.Host != "claude" && s.Host != "grok" {
		return c
	}
	events := s.Trace.Events
	var bounded string
	var unbounded []string
	for i := range events {
		e := &events[i]
		if e.Kind != "shell" || !runsInBackground(s.Host, *e) || !succeeded(events, i) {
			continue
		}
		withDeadline, without := waitLoopCounts(e.Command)
		switch {
		case without > 0:
			unbounded = append(unbounded, regressionEvidence(*e))
		case withDeadline > 0 && bounded == "":
			bounded = regressionEvidence(*e)
		}
	}
	switch {
	case len(unbounded) > 0:
		c.Status, c.Evidence = "fail", unbounded
	case bounded != "":
		c.Status, c.Evidence = "pass", []string{bounded}
	}
	return c
}

// waitLoopCounts counts the loops with sleep in a command, split by whether
// each has a deadline.
func waitLoopCounts(command string) (withDeadline, without int) {
	scanWaitLoops(shellSegments(command, 0), false, 0, &withDeadline, &without)
	return withDeadline, without
}

type openLoop struct {
	start, doAt int
	sleeps      bool
	// nested holds the segment ranges of the loops closed inside this one, which
	// its own deadline search skips.
	nested [][2]int
}

// scanWaitLoops walks the segments of one script. A loop is a wait loop when
// its body holds a sleep at any depth. Its deadline is a leading timeout that
// wraps its script (passed down as wrapped), a finite for, or a comparison of
// a counter or the clock in its own condition or body. A timeout inside a loop
// body, or around a nested inner loop, is not a deadline for the loop.
func scanWaitLoops(segs [][]string, wrapped bool, depth int, withDeadline, without *int) {
	segs = normalizeSegments(segs)
	counters := counterVariables(segs)
	var stack []*openLoop
	markSleep := func() {
		for _, l := range stack {
			l.sleeps = true
		}
	}
	for i, seg := range segs {
		cmd := commandOf(seg)
		rest, timeout := unwrapLauncher(cmd)
		for _, script := range runnerScripts(rest) {
			if nestedSleep.MatchString(script) {
				markSleep()
			}
			if depth < 3 && waitLoopWord.MatchString(script) {
				scanWaitLoops(shellSegments(script, 0), wrapped || timeout, depth+1, withDeadline, without)
			}
		}
		switch strings.TrimRight(seg[0], "&)") {
		case "while", "until", "for":
			stack = append(stack, &openLoop{start: i, doAt: -1})
		case "do":
			if n := len(stack); n > 0 && stack[n-1].doAt < 0 {
				stack[n-1].doAt = i
			}
		case "done":
			for len(stack) > 0 && stack[len(stack)-1].doAt < 0 {
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				continue
			}
			l := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.nested = append(parent.nested, [2]int{l.start, i})
			}
			if !l.sleeps {
				continue
			}
			if wrapped || loopHasDeadline(segs, l, i, counters) {
				*withDeadline++
			} else {
				*without++
			}
			continue
		}
		if len(cmd) > 0 && filepath.Base(strings.Trim(cmd[0], "()")) == "sleep" {
			markSleep()
		}
	}
}

// unwrapLauncher drops the launchers that run another command with the same
// arguments (timeout, nohup, env, exec, command, setsid, nice, stdbuf), their
// flags and leading assignments, and reports whether one of them was timeout
// or gtimeout.
func unwrapLauncher(cmd []string) (rest []string, timeout bool) {
	valued := map[string]bool{"-s": true, "-k": true, "-n": true, "-u": true, "--signal": true, "--kill-after": true}
	for len(cmd) > 0 {
		t := cmd[0]
		base := filepath.Base(t)
		switch base {
		case "timeout", "gtimeout", "nohup", "env", "exec", "command", "setsid", "nice", "stdbuf":
			if strings.Contains(t, "=") {
				return cmd, timeout
			}
			timeout = timeout || base == "timeout" || base == "gtimeout"
			cmd = cmd[1:]
			for len(cmd) > 0 && strings.HasPrefix(cmd[0], "-") {
				flag := cmd[0]
				cmd = cmd[1:]
				if valued[flag] && len(cmd) > 0 {
					cmd = cmd[1:]
				}
			}
			if base == "timeout" || base == "gtimeout" {
				if len(cmd) > 0 {
					cmd = cmd[1:]
				}
			}
			continue
		}
		if strings.Contains(t, "=") && !strings.HasPrefix(t, "-") {
			cmd = cmd[1:]
			continue
		}
		return cmd, timeout
	}
	return cmd, timeout
}

var shellCommandFlag = regexp.MustCompile(`^-[a-z]*c[a-z]*$`)

// runnerScripts returns the script a command hands to a shell (-c) or to ssh.
// It reads no other quoted argument, so an echoed sentence is not a script.
func runnerScripts(cmd []string) []string {
	if len(cmd) == 0 {
		return nil
	}
	switch filepath.Base(cmd[0]) {
	case "bash", "sh", "zsh", "dash":
		for i := 1; i+1 < len(cmd); i++ {
			if shellCommandFlag.MatchString(cmd[i]) {
				return []string{cmd[i+1]}
			}
		}
	case "ssh":
		valued := map[string]bool{"-o": true, "-i": true, "-p": true, "-l": true, "-F": true, "-J": true, "-L": true, "-R": true, "-D": true}
		hostSeen := false
		var script []string
		for i := 1; i < len(cmd); i++ {
			switch {
			case hostSeen:
				script = append(script, cmd[i])
			case strings.HasPrefix(cmd[i], "-"):
				if valued[cmd[i]] {
					i++
				}
			default:
				hostSeen = true
			}
		}
		if len(script) > 0 {
			return []string{strings.Join(script, " ")}
		}
	}
	return nil
}

// commandOf drops the compound-command keywords that precede a command.
func commandOf(seg []string) []string {
	for len(seg) > 0 && waitKeywords[seg[0]] {
		seg = seg[1:]
	}
	return seg
}

// counterVariables names the variables the script uses as a counter or a
// clock: incremented or decremented, assigned from an expression that reads
// the variable, the clock or $SECONDS. A variable set from an unrelated
// expression is not one.
func counterVariables(segs [][]string) map[string]bool {
	names := map[string]bool{}
	selfOrClock := func(name, expr string) {
		if regexp.MustCompile(`\b`+name+`\b`).MatchString(expr) || elapsedOperand.MatchString(expr) || elapsedBareOperand.MatchString(expr) {
			names[name] = true
		}
	}
	for _, seg := range segs {
		text := strings.Join(commandOf(seg), " ")
		if m := clockAssignments.FindStringSubmatch(text); m != nil {
			names[m[1]] = true
		}
		if m := stepAssignments.FindStringSubmatch(text); m != nil {
			names[m[1]+m[2]] = true
		}
		for _, re := range exprAssignments {
			if m := re.FindStringSubmatch(text); m != nil {
				selfOrClock(m[1], m[2])
			}
		}
		if m := letAssignment.FindStringSubmatch(text); m != nil {
			rest := strings.TrimSpace(m[2])
			if strings.HasPrefix(rest, "=") {
				selfOrClock(m[1], rest[1:])
			} else {
				names[m[1]] = true
			}
		}
	}
	return names
}

// loopHasDeadline applies the deadlines that hold without a timeout wrapper: a
// for over a finite list or with a counter condition, and a comparison of a
// counter or the clock in a segment of the loop's own condition or body.
func loopHasDeadline(segs [][]string, l *openLoop, end int, counters map[string]bool) bool {
	header := segs[l.start:l.doAt]
	if h := header[0]; h[0] == "for" {
		switch {
		case len(h) > 1 && strings.HasPrefix(h[1], "(("):
			if len(header) == 3 {
				withVariable := map[string]bool{}
				for name := range counters {
					withVariable[name] = true
				}
				name := ""
				if m := cStyleVariable.FindStringSubmatch(h[1]); m != nil {
					name = m[1]
				} else if h[1] == "((" && len(h) > 2 {
					name = strings.SplitN(h[2], "=", 2)[0]
				}
				if name != "" {
					withVariable[name] = true
				}
				if comparesToLimit(append([]string{"(("}, header[1]...), withVariable) {
					return true
				}
			}
		case len(h) > 2 && h[2] == "in":
			return len(h) > 3
		case len(h) <= 2:
			return true
		}
	}
next:
	for j := l.start; j < end; j++ {
		for _, r := range l.nested {
			if j >= r[0] && j <= r[1] {
				continue next
			}
		}
		if comparesToLimit(segs[j], counters) {
			return true
		}
	}
	return false
}

// comparesToLimit reports whether a test segment compares the clock or a
// counter with something, in either direction. A segment that only prints the
// value is not a test.
func comparesToLimit(seg []string, counters map[string]bool) bool {
	cmd := commandOf(seg)
	if len(cmd) == 0 {
		return false
	}
	arithmetic := strings.HasPrefix(cmd[0], "((")
	last := cmd[len(cmd)-1]
	if cmd[0] != "[" && cmd[0] != "[[" && cmd[0] != "test" && !arithmetic && last != "]" && last != "]]" {
		return false
	}
	compared := false
	for _, t := range cmd {
		switch t {
		case "-lt", "-le", "-gt", "-ge", "-eq", "-ne", "<", "<=", ">", ">=", "==", "!=":
			compared = true
		default:
			compared = compared || (arithmetic && (strings.ContainsAny(t, "<>") || strings.Contains(t, "==") || strings.Contains(t, "!=")))
		}
	}
	if !compared {
		return false
	}
	text := " " + strings.Join(cmd, " ") + " "
	if elapsedOperand.MatchString(text) || (arithmetic && elapsedBareOperand.MatchString(text)) {
		return true
	}
	for name := range counters {
		if regexp.MustCompile(`\$\{?`+name+`\b|\$\(\([^)]*\b`+name+`\b`).MatchString(text) || (arithmetic && regexp.MustCompile(`\b`+name+`\b`).MatchString(text)) {
			return true
		}
	}
	return false
}

// normalizeSegments puts each loop start and each `do` at the head of its own
// segment, so the scan finds them: it splits at a background `&`, splits a `do`
// that follows a C-style for header's `))`, and peels `then while :`, `(while :`
// and `{ until ...`.
func normalizeSegments(segs [][]string) [][]string {
	var split [][]string
	for _, seg := range segs {
		var cur []string
		flush := func() {
			if len(cur) > 0 {
				split = append(split, cur)
			}
			cur = nil
		}
		for _, t := range seg {
			switch {
			case t == "&":
				flush()
			case len(t) > 1 && strings.HasSuffix(t, "&") && !strings.HasSuffix(t, "&&") && !strings.HasSuffix(t, ">&"):
				cur = append(cur, strings.TrimSuffix(t, "&"))
				flush()
			case t == "do" && len(cur) > 0 && strings.HasSuffix(cur[len(cur)-1], "))"):
				flush()
				cur = append(cur, t)
			default:
				cur = append(cur, t)
			}
		}
		flush()
	}
	return peelLoopStarts(split)
}

// peelLoopStarts puts a loop that starts after a keyword or a group opener,
// such as `then while :`, `(while :` or `{ until ...`, at the head of its own
// segment.
func peelLoopStarts(segs [][]string) [][]string {
	isLoop := func(t string) bool { return t == "while" || t == "until" || t == "for" }
	out := make([][]string, 0, len(segs))
	for _, seg := range segs {
		seg = append([]string(nil), seg...)
		for len(seg) > 0 {
			head := seg[0]
			switch {
			case head == "(" || head == "{":
				seg = seg[1:]
				continue
			case strings.HasPrefix(head, "(") && isLoop(head[1:]):
				seg[0] = head[1:]
			case (head == "then" || head == "do" || head == "else" || head == "elif" || head == "!") && len(seg) > 1:
				next := strings.TrimLeft(seg[1], "({")
				if next == "" && len(seg) > 2 {
					next = seg[2]
				}
				if isLoop(next) {
					out = append(out, seg[:1])
					seg = seg[1:]
					continue
				}
			}
			break
		}
		if len(seg) > 0 {
			out = append(out, seg)
		}
	}
	return out
}
