package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// waitLoopWord marks a token that may carry a shell loop, such as the script
	// of `bash -c '...'`.
	waitLoopWord = regexp.MustCompile(`\b(?:while|until|for)\b`)
	// counterAssignments are the ways a script keeps a counter or a clock:
	// arithmetic, `date +%s`, $SECONDS, `((i++))` and `let`.
	counterAssignments = []*regexp.Regexp{
		regexp.MustCompile(`^(\w+)=\$\(\(`),
		regexp.MustCompile(`^(\w+)=\$\(date\s+\+%s`),
		regexp.MustCompile("^(\\w+)=[\"']?\\$\\{?SECONDS"),
		regexp.MustCompile(`^(\w+)=\$\(expr\s`),
		regexp.MustCompile(`^\(\(\s*(?:\+\+|--)?(\w+)\s*(?:\+\+|--|[-+*]?=)`),
		regexp.MustCompile("^let\\s+[\"']?(\\w+)"),
	}
	// cStyleVariable is the variable a C-style for initializes.
	cStyleVariable = regexp.MustCompile(`^\(\((\w+)`)
	// elapsedOperand is a comparison operand that reads the clock directly.
	elapsedOperand = regexp.MustCompile(`SECONDS|date \+%s|\$\(\(`)
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
}

// scanWaitLoops walks the segments of one script. A loop whose script sits in
// a segment that starts with timeout or gtimeout (`timeout 600 bash -c '...'`)
// is bounded by it, passed down as wrapped; a timeout inside a loop body is
// not a deadline for the loop.
func scanWaitLoops(segs [][]string, wrapped bool, depth int, withDeadline, without *int) {
	segs = peelLoopStarts(segs)
	counters := counterVariables(segs)
	var stack []*openLoop
	for i, seg := range segs {
		if depth < 3 {
			for _, token := range seg[1:] {
				if strings.ContainsAny(token, " \n;") && waitLoopWord.MatchString(token) {
					scanWaitLoops(shellSegments(token, 0), wrapped || leadsWithTimeout(seg), depth+1, withDeadline, without)
				}
			}
		}
		first := strings.TrimRight(seg[0], "&)")
		switch first {
		case "while", "until", "for":
			stack = append(stack, &openLoop{start: i, doAt: -1})
		case "do":
			if n := len(stack); n > 0 && stack[n-1].doAt < 0 {
				stack[n-1].doAt = i
			}
		case "done":
			n := len(stack)
			if n == 0 {
				continue
			}
			l := stack[n-1]
			stack = stack[:n-1]
			if l.doAt < 0 || !l.sleeps {
				continue
			}
			if wrapped || loopHasDeadline(segs, l, i, counters) {
				*withDeadline++
			} else {
				*without++
			}
			continue
		}
		if cmd := commandOf(seg); len(cmd) > 0 && cmd[0] == "sleep" {
			for _, l := range stack {
				l.sleeps = true
			}
		}
	}
}

// leadsWithTimeout reports whether the segment's command is timeout or
// gtimeout, after leading variable assignments and nohup.
func leadsWithTimeout(seg []string) bool {
	for _, t := range seg {
		if strings.Contains(t, "=") || t == "nohup" {
			continue
		}
		base := filepath.Base(t)
		return base == "timeout" || base == "gtimeout"
	}
	return false
}

// commandOf drops the compound-command keywords that precede a command.
func commandOf(seg []string) []string {
	for len(seg) > 0 && waitKeywords[seg[0]] {
		seg = seg[1:]
	}
	return seg
}

// counterVariables names the variables the script sets from arithmetic, the
// clock or $SECONDS, or increments.
func counterVariables(segs [][]string) map[string]bool {
	names := map[string]bool{}
	for _, seg := range segs {
		text := strings.Join(commandOf(seg), " ")
		for _, re := range counterAssignments {
			if m := re.FindStringSubmatch(text); m != nil {
				names[m[1]] = true
			}
		}
	}
	return names
}

// loopHasDeadline applies the three kinds of deadline that hold without a
// timeout wrapper: a finite for list, a counter or the clock compared in the
// loop condition, and such a comparison in the body that leads to exit or break.
func loopHasDeadline(segs [][]string, l *openLoop, end int, counters map[string]bool) bool {
	header := segs[l.start:l.doAt]
	switch header[0][0] {
	case "for":
		h := header[0]
		switch {
		case len(h) > 1 && strings.HasPrefix(h[1], "(("):
			if len(header) != 3 {
				return false
			}
			withVariable := map[string]bool{}
			for name := range counters {
				withVariable[name] = true
			}
			if m := cStyleVariable.FindStringSubmatch(h[1]); m != nil {
				withVariable[m[1]] = true
			}
			return comparesToLimit(append([]string{"(("}, header[1]...), false, withVariable)
		case len(h) > 2 && h[2] == "in":
			return len(h) > 3
		}
		return true
	case "while":
		for _, h := range header {
			if comparesToLimit(h, false, counters) {
				return true
			}
		}
	case "until":
		for _, h := range header {
			if comparesToLimit(h, true, counters) {
				return true
			}
		}
	}
	for j := l.doAt; j < end; j++ {
		if comparesToLimit(segs[j], true, counters) && leadsToExit(segs, j, end) {
			return true
		}
	}
	return false
}

// comparesToLimit reports whether a test segment compares the clock or a
// counter with a limit. upper selects -ge, -gt, >= and >; otherwise -lt, -le,
// <= and <. A segment that only prints the value is not a test.
func comparesToLimit(seg []string, upper bool, counters map[string]bool) bool {
	cmd := commandOf(seg)
	if len(cmd) == 0 {
		return false
	}
	arithmetic := strings.HasPrefix(cmd[0], "((")
	if cmd[0] != "[" && cmd[0] != "[[" && cmd[0] != "test" && !arithmetic {
		return false
	}
	text := " " + strings.Join(cmd, " ") + " "
	ops := []string{" -lt ", " -le "}
	if upper {
		ops = []string{" -ge ", " -gt "}
	}
	if arithmetic || cmd[0] == "[[" {
		if upper {
			ops = append(ops, ">=", ">")
		} else {
			ops = append(ops, "<=", "<")
		}
	}
	compared := false
	for _, op := range ops {
		compared = compared || strings.Contains(text, op)
	}
	if !compared {
		return false
	}
	if elapsedOperand.MatchString(text) {
		return true
	}
	for name := range counters {
		if regexp.MustCompile(`\$\{?`+name+`\b`).MatchString(text) || (arithmetic && regexp.MustCompile(`\b`+name+`\b`).MatchString(text)) {
			return true
		}
	}
	return false
}

// leadsToExit reports whether the test at segs[j] ends the script or the loop
// when it holds: the branch of an if, or the command joined to it with && or ||.
func leadsToExit(segs [][]string, j, end int) bool {
	exits := func(seg []string) bool {
		cmd := commandOf(seg)
		return len(cmd) > 0 && (cmd[0] == "exit" || cmd[0] == "break")
	}
	if segs[j][0] != "if" && segs[j][0] != "elif" {
		return j+1 < end && exits(segs[j+1])
	}
	for k := j + 1; k < end; k++ {
		switch segs[k][0] {
		case "fi", "else", "elif":
			return false
		}
		if exits(segs[k]) {
			return true
		}
	}
	return false
}

// peelLoopStarts puts a loop that starts after a keyword or a group opener,
// such as `then while :`, `(while :` or `{ until ...`, at the head of its own
// segment, so the scan finds it.
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
