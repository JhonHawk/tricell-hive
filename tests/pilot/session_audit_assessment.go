package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ticketKey is a Jira or Linear key in a prompt, such as CAM-20. Its
	// prefix is checked against standardPrefixes, which name standards and
	// algorithms that look like keys.
	ticketKey        = regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])([A-Z]{2,}[A-Z0-9]*)-\d+\b`)
	standardPrefixes = map[string]bool{
		"UTF": true, "SHA": true, "ISO": true, "RFC": true, "CVE": true, "CWE": true, "MD": true,
		"AES": true, "RSA": true, "TLS": true, "SSL": true, "HTTP": true, "IEEE": true, "ECMA": true,
		"BCP": true, "PEP": true, "GPT": true, "WCAG": true, "ARIA": true, "ES": true,
	}
	// githubIssue is an issue number written next to the word issue or ticket.
	githubIssue = regexp.MustCompile(`(?i)\b(?:issue|ticket)\b[^\n#]{0,20}#\d+|#\d+[^\n#]{0,20}\b(?:issue|ticket)\b`)
	// hiveModeRead names the installed SKILL.md of hive-mode under a host's
	// skills folder; a read of the repository's own content/skills is research.
	hiveModeRead = regexp.MustCompile(`/\.[a-z]+/skills/hive-mode/SKILL\.md`)
	// ticketStatusRead is the playbook that reports a ticket's status, not an
	// assessment of it.
	ticketStatusRead = regexp.MustCompile(`hive-mode/playbooks/ticket-status\.md`)
	// hiveModeCommand is the hive-mode command typed at the start of the
	// prompt; the host injects the skill, so no event of the turn loads it.
	hiveModeCommand  = regexp.MustCompile(`^\s*(?:<command-message>[^<]*</command-message>\s*)?(?:<command-name>)?(?:[/$@]|/skill:)hive-mode(?:\s|<|$)`)
	recommendedLabel = regexp.MustCompile(`(?i)\(★\)|\((?:recommended|recomendad[oa])\)`)
)

// namesTicket reports whether a prompt names a ticket: a Jira or Linear key,
// or a GitHub issue number next to the word issue or ticket.
func namesTicket(prompt string) bool {
	for _, m := range ticketKey.FindAllStringSubmatch(prompt, -1) {
		if !standardPrefixes[m[1]] {
			return true
		}
	}
	return githubIssue.MatchString(prompt)
}

// loadsHiveMode reports whether an event loads the hive-mode skill: a skill
// call, or a read of its SKILL.md, also inside a shell command that reads
// several files.
func loadsHiveMode(e traceEvent) bool {
	switch e.Kind {
	case "skill_invocation":
		return e.Text == "hive-mode"
	case "read", "search":
		return hiveModeRead.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		return hiveModeRead.MatchString(e.Command) && readCommand.MatchString(e.Command)
	}
	return false
}

func readsTicketStatus(e traceEvent) bool {
	switch e.Kind {
	case "read", "search":
		return ticketStatusRead.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		return ticketStatusRead.MatchString(e.Command)
	}
	return false
}

// ticketAssessmentEnding checks how the assessment of a ticket closes. A turn
// is an assessment when the user's prompt names a ticket and the agent loads
// hive-mode in it; a turn that follows a flow command or reads the
// ticket-status playbook is not. The assessment must end with a question
// card, a next-step block or a flow command: without one it fails, and with an
// open question in prose it is flagged for review.
func ticketAssessmentEnding(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "ticket_assessment_ending", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	type turnState struct {
		load, closing, card int
		status              bool
	}
	turns := map[int]*turnState{}
	var order []int
	for i, e := range events {
		turn := s.Turn[i]
		st := turns[turn]
		if st == nil {
			st = &turnState{load: -1, closing: -1, card: -1}
			turns[turn] = st
			order = append(order, turn)
		}
		if st.load < 0 && hiveModeCommand.MatchString(s.Prompts[turn]) {
			st.load = i
		}
		switch {
		case readsTicketStatus(e):
			st.status = true
		case st.load < 0 && loadsHiveMode(e):
			st.load = i
		case st.load >= 0 && e.Kind == "question":
			st.card = i
		case st.load >= 0 && e.Kind == "text" && e.Role == "assistant":
			st.closing = i
		}
	}
	var failed, review, passed []string
	for _, turn := range order {
		st := turns[turn]
		prompt := s.Prompts[turn]
		if st.load < 0 || st.status || st.closing < 0 && st.card < 0 || !namesTicket(prompt) || flowCommand.MatchString(prompt) {
			continue
		}
		if st.card > st.closing {
			passed = append(passed, regressionEvidence(events[st.card])+" card")
			continue
		}
		e := events[st.closing]
		switch {
		case strings.Contains(e.Text, "▶") || flowCommand.MatchString(e.Text):
			passed = append(passed, fmt.Sprintf("%s next-step block", regressionEvidence(e)))
		case asksTheUser.MatchString(codeOrLink.ReplaceAllString(e.Text, "")):
			review = append(review, regressionEvidence(e))
		default:
			failed = append(failed, regressionEvidence(e))
		}
	}
	if len(failed) == 0 && len(review) > 0 {
		c.Status, c.Evidence = "review", review
		return c
	}
	return statusOf(c, failed, passed)
}

// cardAnswers counts the questions of a session's cards: how many were
// answered with one of their options, and how many of those chose the
// recommended one. It is informative and never a criterion.
type cardAnswers struct {
	Answered    int `json:"answered"`
	Recommended int `json:"recommended"`
}

// cardAnswersOf reads every question card and its result. An empty card, a
// free-text answer and a multiple choice do not count as answered with an
// option.
func cardAnswersOf(s auditSession) cardAnswers {
	var out cardAnswers
	events := s.Trace.Events
	for i, e := range events {
		if e.Kind != "question" || len(e.Input) == 0 {
			continue
		}
		result, ok := cardResult(events, i)
		if !ok {
			continue
		}
		var in struct {
			Questions []struct {
				ID          string
				Question    string
				MultiSelect bool
				Options     []struct{ Label string }
			}
		}
		if json.Unmarshal(e.Input, &in) != nil {
			continue
		}
		for _, q := range in.Questions {
			if q.MultiSelect {
				continue
			}
			var labels []string
			for _, o := range q.Options {
				labels = append(labels, o.Label)
			}
			label := chosenLabel(s.Host, result, q.ID, q.Question, labels)
			if label == "" {
				continue
			}
			out.Answered++
			if recommendedLabel.MatchString(label) {
				out.Recommended++
			}
		}
	}
	return out
}

// cardResult returns the text of the result of the call at qi.
func cardResult(events []traceEvent, qi int) (string, bool) {
	for i := qi + 1; i < len(events); i++ {
		if events[i].Kind == "tool_result" && events[i].ID == events[qi].ID {
			return events[i].Text, true
		}
	}
	return "", false
}

// chosenLabel returns the option label a result records for a question, or ""
// when the answer is empty, free text or absent. Claude writes each answer as
// "question"="label"; Codex answers {"answers":{"<id>":{"answers":["label"]}}}.
func chosenLabel(host, result, id, question string, labels []string) string {
	switch host {
	case "claude":
		for _, l := range labels {
			if strings.Contains(result, fmt.Sprintf("%q=%q", question, l)) || strings.Contains(result, `"`+question+`"="`+l+`"`) {
				return l
			}
		}
	case "codex":
		var v struct {
			Answers map[string]struct{ Answers []string }
		}
		if json.Unmarshal([]byte(strings.TrimSpace(result)), &v) != nil {
			return ""
		}
		given := v.Answers[id].Answers
		if len(given) != 1 {
			return ""
		}
		for _, l := range labels {
			if given[0] == l {
				return l
			}
		}
	}
	return ""
}

var (
	// cardOptionLetter opens an option label of a question card: a capital
	// letter, a space, a dash and a space, as "A - (★) Accept".
	cardOptionLetter = regexp.MustCompile(`^[A-Z] - `)
	// cardOptionMark is the recommended mark of a label: the star right after
	// its letter, or the host's "(Recommended)" at its end. A star quoted later
	// in another option's name is not a mark.
	cardOptionMark = regexp.MustCompile(`^[A-Z] - \(★\)|(?i)\(recommended\)\s*$`)
)

// cardOptionFormat checks the option labels of every waiting question card a
// session shows, answered or not: each opens with its letter and a dash, and
// a single choice marks its first option, and only it, as recommended. Codex's
// request_user_input_async does not wait for the answer and is not judged. The
// evidence gives the card's line and the question's number in it, never its
// text or its ID, which the model writes.
func cardOptionFormat(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "card_option_format", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	var failed, passed []string
	for _, e := range s.Trace.Events {
		if e.Kind != "question" || e.Tool == "request_user_input_async" {
			continue
		}
		for n, raw := range list(decodeInput(e)["questions"]) {
			q := object(raw)
			var labels []string
			for _, o := range list(q["options"]) {
				if label := str(object(o)["label"]); label != "" {
					labels = append(labels, label)
				} else if label := str(o); label != "" {
					labels = append(labels, label)
				}
			}
			if len(labels) == 0 {
				continue
			}
			name := fmt.Sprintf("question %d", n+1)
			multi, _ := q["multiSelect"].(bool)
			if reason := cardOptionsOutOfFormat(labels, multi); reason != "" {
				failed = append(failed, fmt.Sprintf("%s %s: %s", regressionEvidence(e), name, reason))
			} else {
				passed = append(passed, fmt.Sprintf("%s %s", regressionEvidence(e), name))
			}
		}
	}
	return statusOf(c, failed, passed)
}

// cardOptionsOutOfFormat says why a question's option labels leave the format,
// or "" when they keep it.
func cardOptionsOutOfFormat(labels []string, multi bool) string {
	for _, l := range labels {
		if !cardOptionLetter.MatchString(l) {
			return "an option without its letter and dash"
		}
	}
	if multi {
		return ""
	}
	var marked []int
	for i, l := range labels {
		if cardOptionMark.MatchString(l) {
			marked = append(marked, i)
		}
	}
	switch {
	case len(marked) == 0:
		return "no option marked as recommended"
	case len(marked) > 1 || marked[0] != 0:
		return "the recommended mark is not on the first option alone"
	}
	return ""
}
