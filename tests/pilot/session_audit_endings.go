package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// reportEndingsRead names the installed SKILL.md of hive-rule-report-endings
	// under a host's hidden skills folder; a read of the repository's own
	// content/skills is research.
	reportEndingsRead = regexp.MustCompile(`/\.[a-z]+/skills/hive-rule-report-endings/SKILL\.md`)
	listItem          = regexp.MustCompile(`^\s*(?:[-*+]|\d+[.)])\s+(.*)$`)
	// letterItem is an option written with a letter, as the reply block sets:
	// "A — …", "B) …", "- **C** …". A letter joined to a word, as in "E-mail",
	// is not one.
	letterItem = regexp.MustCompile(`^\s*(?:[-*+]\s+)?(?:\*\*([A-Z])\*\*(?:\s|$)|([A-Z])(?:\s*[—–]\s|\s+-\s|[.):]\s|\s\())`)
	// italicLabel opens a list item whose label is in italics, not bold; an
	// underscore label holds no slash, so a path such as _support/x is not one.
	italicLabel = regexp.MustCompile(`^(?:\*([^*\s][^*]*)\*|_([^_\s/][^_/]*)_)(?:\s|[:.,;—–-]|$)`)
	// boldQuestion is a heading line that asks in bold, which the reply block
	// accepts in place of the arrow.
	boldQuestion = regexp.MustCompile(`\*\*[^*\n]*\?[^*\n]*\*\*`)
	singleLetter = regexp.MustCompile(`^[A-Z]$`)
	// deliveryCommand is a shell command that delivers work, which makes the
	// turn a work turn rather than an analysis.
	deliveryCommand = regexp.MustCompile(`\bgit\s+(?:-C\s+\S+\s+)?(?:commit|push|merge)\b|\bgh\s+pr\s+(?:create|merge)\b`)
	// tempPath is the host's temporary location, where an analysis may keep
	// scratch files without becoming a work turn.
	tempPath = regexp.MustCompile(`^(?:/private)?/tmp/|^/var/folders/|/scratchpad/`)
)

// reportMinLines is the fewest non-blank lines a last text needs to be read as
// a report that owes the reply block; a shorter one is a plain answer.
const reportMinLines = 6

// workEvent reports whether an event changes files or delivers work outside
// the host's temporary location.
func workEvent(e traceEvent) bool {
	switch e.Kind {
	case "write", "delete":
		return !tempPath.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		command := e.Command + "\n" + e.CommandTemplate
		if deliveryCommand.MatchString(command) {
			return true
		}
		for _, p := range shellWritePaths(command) {
			if !tempPath.MatchString(p) {
				return true
			}
		}
	}
	return false
}

// bareReport reports whether a last text is a report of reportMinLines or more
// non-blank lines that ends with no reply block: no ▶ anywhere, no flow
// command, and no question in bold among its last lines.
func bareReport(text string) bool {
	if strings.Contains(text, "▶") || flowCommand.MatchString(text) {
		return false
	}
	var lines []string
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	if len(lines) < reportMinLines {
		return false
	}
	return !boldQuestion.MatchString(strings.Join(lines[len(lines)-3:], "\n"))
}

func readsReportEndings(e traceEvent) bool {
	switch e.Kind {
	case "read", "search":
		return reportEndingsRead.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		return reportEndingsRead.MatchString(e.Command) && readCommand.MatchString(e.Command)
	}
	return false
}

// replyEnding is how a text ends when it offers replies: the list items, the
// lines above them that may head the list (after the last --- rule, or the
// two just above it), and the tail from that heading to the end.
type replyEnding struct {
	items   []string
	heading string
	tail    string
}

func isReplyItem(l string) bool { return listItem.MatchString(l) || letterItem.MatchString(l) }

// trailingReplyList reads the list a text ends with. A short closing
// paragraph after it, such as the waiting line, is skipped; indented lines
// continue the item above, and blank lines between items do not end the list.
func trailingReplyList(text string) replyEnding {
	lines := strings.Split(strings.TrimRight(text, " \t\r\n"), "\n")
	end := len(lines)
	if blank := lastBlank(lines); blank > 0 && end-blank-1 <= 2 && !anyItem(lines[blank+1:]) && isReplyItem(lastNonBlank(lines[:blank])) {
		end = blank
	}
	first := end
	for i := end - 1; i >= 0; i-- {
		l := lines[i]
		switch {
		case isReplyItem(l):
			first = i
		case strings.TrimSpace(l) == "" && isReplyItem(lastNonBlank(lines[:i])):
		case strings.TrimSpace(l) != "" && (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")):
		default:
			i = -1
		}
	}
	var ending replyEnding
	for _, l := range lines[first:end] {
		if isReplyItem(l) {
			ending.items = append(ending.items, l)
		}
	}
	// The ending is what follows the last --- rule above the list, as the
	// next-step block and the close footer write it; without a rule, it is the
	// up to two non-blank lines just above the list.
	top := -1
	for i := first - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "---" {
			top = i + 1
			break
		}
	}
	if top < 0 {
		top = first
		for n := 0; top > 0 && n < 2; {
			top--
			if strings.TrimSpace(lines[top]) != "" {
				n++
			}
		}
	}
	ending.heading = strings.Join(lines[top:first], "\n")
	// The tail starts at a ▶ heading or a rule when one opens the ending,
	// otherwise at the list: a body sentence above it is not part of it.
	start := first
	for i := top; i < first; i++ {
		if strings.Contains(lines[i], "▶") || (i == top && i > 0 && strings.TrimSpace(lines[i-1]) == "---") {
			start = i
			break
		}
	}
	ending.tail = strings.Join(lines[start:], "\n")
	return ending
}

func lastBlank(lines []string) int {
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "" {
			return i
		}
	}
	return -1
}

func lastNonBlank(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return lines[i]
		}
	}
	return ""
}

func anyItem(lines []string) bool {
	for _, l := range lines {
		if isReplyItem(l) {
			return true
		}
	}
	return false
}

// letteredRun counts the items lettered A, B, C… in order from A.
func letteredRun(items []string) int {
	n := 0
	for _, l := range items {
		m := letterItem.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if m[1]+m[2] != string(rune('A'+n)) {
			return n
		}
		n++
	}
	return n
}

// replyListWithoutBlock reports whether a text ends with replies offered as a
// list the reply block does not allow: two or more items labeled in italics
// by something other than a single letter, or lettered items with neither a
// ▶ nor a bold question in the lines just above them. listed tells whether it
// ends with such a list at all.
func replyListWithoutBlock(text string) (bad, listed bool) {
	ending := trailingReplyList(text)
	if len(ending.items) < 2 {
		return false, false
	}
	if letteredRun(ending.items) >= 2 {
		return !strings.Contains(ending.heading, "▶") && !boldQuestion.MatchString(ending.heading), true
	}
	labeled := 0
	for _, l := range ending.items {
		m := listItem.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if e := italicLabel.FindStringSubmatch(strings.TrimSpace(m[1])); e != nil {
			if !singleLetter.MatchString(strings.TrimSpace(e[1] + e[2])) {
				labeled++
			}
		}
	}
	return labeled >= 2, labeled >= 2
}

// endingExempt reports whether a text's ending hands over to a flow, which
// the next-step block governs: a flow command in the lines that head its
// final list, in the list itself or after it.
func endingExempt(text string) bool {
	return flowCommand.MatchString(trailingReplyList(text).tail)
}

// analysisReplyBlock checks how a main-session turn ends. The last assistant
// text of each turn fails when it ends with italic-labeled list items or with
// lettered items and no ▶ or bold question heading, or when it offers no
// replies at all in a report (bareReport) of a turn that loaded hive-mode,
// changed no file and delivered nothing; the evidence says whether the installed report-endings
// rule was read earlier. A turn that ends with a question card or a flow
// command, a decision round written in text and a child session are not
// judged.
func analysisReplyBlock(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "analysis_reply_block", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	type turnState struct {
		text, card   int
		worked, hive bool
	}
	turns := map[int]*turnState{}
	var order []int
	read := -1
	for i, e := range events {
		turn := s.Turn[i]
		st := turns[turn]
		if st == nil {
			st = &turnState{text: -1, card: -1}
			turns[turn] = st
			order = append(order, turn)
		}
		st.worked = st.worked || workEvent(e)
		st.hive = st.hive || loadsHiveMode(e)
		switch {
		case e.Kind == "question":
			st.card = i
		case e.Kind == "text" && e.Role == "assistant":
			st.text = i
		case read < 0 && readsReportEndings(e):
			read = i
		}
	}
	var failed, passed []string
	for _, turn := range order {
		st := turns[turn]
		if st.text < 0 || st.card > st.text {
			continue
		}
		e := events[st.text]
		if endingExempt(e.Text) || isTextRound(e.Text) {
			continue
		}
		rule := "report-endings rule not read"
		if read >= 0 && read < st.text {
			rule = "report-endings rule was read"
		}
		bad, listed := replyListWithoutBlock(e.Text)
		if !listed {
			if st.hive && !st.worked && bareReport(e.Text) {
				failed = append(failed, fmt.Sprintf("%s analysis ends without a reply block; %s", regressionEvidence(e), rule))
			}
			continue
		}
		if !bad {
			passed = append(passed, regressionEvidence(e)+" reply block")
			continue
		}
		failed = append(failed, fmt.Sprintf("%s replies listed without the reply block; %s", regressionEvidence(e), rule))
	}
	return statusOf(c, failed, passed)
}
