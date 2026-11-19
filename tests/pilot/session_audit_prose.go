package main

import "regexp"

var (
	// proseQuestion is one question in prose: a Spanish ¿…? or an English
	// sentence that ends in a question mark, from a line start or a sentence end.
	proseQuestion = regexp.MustCompile(`(?m)¿[^?¿\n]*\?|(?:^|[.!:]\s+)[^.!?¿\n]*\?`)
	// alternatives offers the user options, "x o y" or "x or y".
	alternatives = regexp.MustCompile(`(?i)\s(?:o|u|or)\s`)
	// sessionChoice names what a session question decides: the delivery mode,
	// the code review or the security review. It is always a round; another
	// choice in prose may be a question about what the user wants.
	sessionChoice = regexp.MustCompile("(?i)\\bentrega\\b|revisi[oó]n (?:de )?(?:c[oó]digo|seguridad)|\\bdelivery\\b|\\b(?:code|security) review\\b|`(?:interactive|automatic|ready|hold|direct-base)`|\\b(?:interactive|direct-base)\\b")
	fencedCode    = regexp.MustCompile("(?s)```.*?```")
	// nextStepBlock is the block that ends a report, from its ▶ heading on.
	nextStepBlock = regexp.MustCompile(`(?s)[^\n]*▶.*$`)
)

// proseChoices returns the questions of a message that ask the user to choose:
// alternatives in prose, or a session question asked as a yes or no.
func proseChoices(text string) (choices []string, session bool) {
	text = nextStepBlock.ReplaceAllString(fencedCode.ReplaceAllString(text, ""), "")
	for _, q := range proseQuestion.FindAllString(text, -1) {
		isSession := sessionChoice.MatchString(q)
		if alternatives.MatchString(q) || isSession {
			choices = append(choices, q)
			session = session || isSession
		}
	}
	return choices, session
}

// proseChoiceWithoutCard checks the choices asked in prose while a waiting card
// is available and none follows in that turn. A delivery, review or security
// choice is a session question, so it fails; any other choice may be a question
// about what the user wants, so it is flagged for review. A round in the full
// text format, which text_round_without_card judges, and the text a card that
// closed empty requires, are not judged here.
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
		if asks, _ := roundFormat(e.Text); asks {
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
// in its turn, or an earlier card of that turn closed empty, which lets the
// round be written in the text.
func cardFollowsOrClosedEmpty(s auditSession, i int) bool {
	events := s.Trace.Events
	for j := range events {
		if events[j].Kind != "question" || s.Turn[j] != s.Turn[i] {
			continue
		}
		if j > i || cardResultIsEmpty(s.Host, events, j, i) {
			return true
		}
	}
	return false
}
