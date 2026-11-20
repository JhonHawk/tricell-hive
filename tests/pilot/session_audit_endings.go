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
	// "A — …", "B) …", "- **C** — …".
	letterItem = regexp.MustCompile(`^\s*(?:[-*+]\s+)?(?:\*\*)?[A-Z](?:\*\*)?\s*(?:[—–.)\-:]|\s\()`)
	// italicLabel opens a list item whose label is in italics, not bold.
	italicLabel = regexp.MustCompile(`^(?:\*([^*\s][^*]*)\*|_([^_\s][^_]*)_)`)
	// boldQuestion is a heading line that asks in bold, which the reply block
	// accepts in place of the arrow.
	boldQuestion = regexp.MustCompile(`\*\*[^*\n]*\?[^*\n]*\*\*`)
	singleLetter = regexp.MustCompile(`^[A-Z]$`)
)

func readsReportEndings(e traceEvent) bool {
	switch e.Kind {
	case "read", "search":
		return reportEndingsRead.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		return reportEndingsRead.MatchString(e.Command) && readCommand.MatchString(e.Command)
	}
	return false
}

// trailingReplyList returns the list items a text ends with and the text above
// them. Indented lines continue the item above; a blank or other line stops.
func trailingReplyList(text string) (items []string, above string) {
	lines := strings.Split(strings.TrimRight(text, " \t\r\n"), "\n")
	first := len(lines)
	for i := len(lines) - 1; i >= 0; i-- {
		l := lines[i]
		switch {
		case listItem.MatchString(l) || letterItem.MatchString(l):
			items = append([]string{l}, items...)
			first = i
		case strings.TrimSpace(l) != "" && (strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t")):
		default:
			return items, strings.Join(lines[:first], "\n")
		}
	}
	return items, strings.Join(lines[:first], "\n")
}

// replyListWithoutBlock reports whether a text ends with replies offered as a
// list the reply block does not allow: two or more items labeled in italics
// by something other than a single letter, or lettered items with neither a
// ▶ nor a bold question above them. listed tells whether it ends with such a
// list at all.
func replyListWithoutBlock(text string) (bad, listed bool) {
	items, above := trailingReplyList(text)
	if len(items) < 2 {
		return false, false
	}
	lettered := 0
	for _, l := range items {
		if letterItem.MatchString(l) {
			lettered++
		}
	}
	if lettered >= 2 {
		return !strings.Contains(above, "▶") && !boldQuestion.MatchString(above), true
	}
	labeled := 0
	for _, l := range items {
		m := listItem.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if e := italicLabel.FindStringSubmatch(strings.TrimSpace(m[1])); e != nil {
			label := e[1] + e[2]
			if !singleLetter.MatchString(strings.TrimSpace(label)) {
				labeled++
			}
		}
	}
	return labeled >= 2, labeled >= 2
}

// analysisReplyBlock checks how a main-session turn ends when it offers the
// user replies. The last assistant text of each turn fails when it ends with
// italic-labeled list items or with lettered items and no ▶ or bold question
// heading; the evidence says whether the installed report-endings rule was
// read earlier. A turn that ends with a question card or a flow command, a
// decision round written in text and a child session are not judged.
func analysisReplyBlock(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "analysis_reply_block", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	type turnState struct{ text, card int }
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
		if flowCommand.MatchString(e.Text) || isTextRound(e.Text) {
			continue
		}
		bad, listed := replyListWithoutBlock(e.Text)
		if !listed {
			continue
		}
		if !bad {
			passed = append(passed, regressionEvidence(e)+" reply block")
			continue
		}
		rule := "report-endings rule not read"
		if read >= 0 && read < st.text {
			rule = "report-endings rule was read"
		}
		failed = append(failed, fmt.Sprintf("%s replies listed without the reply block; %s", regressionEvidence(e), rule))
	}
	return statusOf(c, failed, passed)
}
