package main

import (
	"regexp"
	"strings"
)

var (
	// sentenceEnd ends a sentence: a stop followed by a space. A dot inside a
	// word, as in a file name, does not end it.
	sentenceEnd = regexp.MustCompile(`([.!?:])\s+`)
	// alternatives offers the user options, "x o y" or "x or y".
	alternatives = regexp.MustCompile(`(?i)\s(?:o|u|or)\s`)
	// confirmation asks a yes or no about one option, such as «¿Confirmas…?».
	confirmation = regexp.MustCompile(`(?i)\b(?:confirm[ao]s?|confirm|apruebas|approve|aceptas|accept)\b`)
	// sessionChoice names what a session question decides: the delivery mode,
	// the code review or the security review. It is always a round; another
	// choice in prose may be a question about what the user wants.
	sessionChoice = regexp.MustCompile("(?i)\\bentrega\\b|revisi[oó]n (?:del? )?(?:c[oó]digo|seguridad)|/code-review|\\bdelivery\\b|\\b(?:code|security) review\\b|`(?:interactive|automatic|ready|hold|direct-base)`|\\b(?:interactive|direct-base)\\b")
	fencedCode    = regexp.MustCompile("(?s)```.*?```")
	// nextStepBlock is the block that ends a report, from its ▶ heading on.
	nextStepBlock = regexp.MustCompile(`(?s)[^\n]*▶.*$`)
)

// proseChoices returns the questions of a message that ask the user to choose:
// alternatives in prose, or a session question asked as a yes or no.
func proseChoices(text string) (choices []string, session bool) {
	text = nextStepBlock.ReplaceAllString(fencedCode.ReplaceAllString(text, ""), "")
	for _, q := range proseQuestions(text) {
		offers := alternatives.MatchString(q)
		isSession := sessionChoice.MatchString(q) && (offers || confirmation.MatchString(q))
		if offers || isSession {
			choices = append(choices, q)
			session = session || isSession
		}
	}
	return choices, session
}

// proseQuestions returns each sentence of the text that ends in a question
// mark, from its ¿ when it has one. A round's title line asks its own question;
// text_round_without_card judges the round, so its title is skipped.
func proseQuestions(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		if roundTitleLine.MatchString(line) {
			continue
		}
		for _, s := range strings.Split(sentenceEnd.ReplaceAllString(line, "$1\n"), "\n") {
			s = strings.TrimSpace(s)
			if !strings.HasSuffix(s, "?") {
				continue
			}
			if i := strings.LastIndex(s, "¿"); i >= 0 {
				s = s[i:]
			}
			out = append(out, s)
		}
	}
	return out
}

// proseChoiceWithoutCard checks the choices asked in prose while a waiting card
// is available and none follows in that turn. A delivery, review or security
// choice is a session question, so it fails; any other choice may be a question
// about what the user wants, so it is flagged for review. The text the last
// card of the turn requires when it closed empty is not judged.
func proseChoiceWithoutCard(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "prose_choice_without_card", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	offered := s.QuestionOffered
	for _, e := range events {
		if e.Kind == "question" {
			offered = true
		}
	}
	if !offered {
		return c
	}
	unattended := unattendedActivation(events)
	var failed, review []string
	for i, e := range events {
		if e.Kind != "text" || e.Role != "assistant" || (unattended >= 0 && i > unattended) {
			continue
		}
		choices, session := proseChoices(e.Text)
		if len(choices) == 0 || cardFollowsOrClosedEmpty(s, i) {
			continue
		}
		if session {
			failed = append(failed, regressionEvidence(e))
		} else {
			review = append(review, regressionEvidence(e))
		}
	}
	if len(failed) == 0 && len(review) > 0 {
		c.Status, c.Evidence = "review", review
		return c
	}
	return statusOf(c, failed, nil)
}

// cardFollowsOrClosedEmpty reports whether a question follows the message at i
// in its turn, or the last card before it in that turn closed empty, which lets
// the round be written in the text, as text_round_without_card reads it.
func cardFollowsOrClosedEmpty(s auditSession, i int) bool {
	events := s.Trace.Events
	last := -1
	for j := range events {
		if events[j].Kind != "question" || s.Turn[j] != s.Turn[i] {
			continue
		}
		if j > i {
			return true
		}
		last = j
	}
	return last >= 0 && cardResultIsEmpty(s.Host, events, last, i)
}
