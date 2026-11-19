package main

import (
	"regexp"
	"strings"
)

var (
	// proseChoice is a question in prose that offers the user alternatives,
	// such as "¿Prefieres `interactive` o `hold`?" or "Do you want A or B?".
	proseChoice = regexp.MustCompile(`(?i)¿[^?¿\n]*\s[ou]\s[^?¿\n]*\?|(?:^|[.!:]\s+)(?:do|would|should|shall|can|could|which)\b[^?\n]*\sor\s[^?\n]*\?`)
	fencedCode  = regexp.MustCompile("(?s)```.*?```")
	// sessionChoice names what a session question decides, which is always a
	// round; another choice in prose may be a question about what the user
	// wants, which the text may carry.
	sessionChoice = regexp.MustCompile(`(?i)\b(?:entrega|delivery|revisi[oó]n|review|seguridad|security|interactive|hold|ready|automatic|direct-base)\b`)
)

// proseChoiceWithoutCard checks the choices asked in prose while a waiting card
// is available and none follows in that turn. A delivery, review or security
// choice is a session question, so it fails; any other choice may be a question
// about what the user wants, so it is flagged for review.
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
	var failed, review, passed []string
	for i, e := range events {
		if e.Kind != "text" || e.Role != "assistant" || (unattended >= 0 && i > unattended) {
			continue
		}
		text := fencedCode.ReplaceAllString(e.Text, "")
		choices := proseChoice.FindAllString(text, -1)
		if len(choices) == 0 || strings.Contains(e.Text, "▶") {
			continue
		}
		carded := false
		for j := i + 1; j < len(events) && s.Turn[j] == s.Turn[i]; j++ {
			if events[j].Kind == "question" {
				carded = true
				break
			}
		}
		switch {
		case carded:
			passed = append(passed, regressionEvidence(e))
		case sessionChoice.MatchString(strings.Join(choices, " ")):
			failed = append(failed, regressionEvidence(e))
		default:
			review = append(review, regressionEvidence(e))
		}
	}
	if len(failed) == 0 && len(review) > 0 {
		c.Status, c.Evidence = "review", review
		return c
	}
	return statusOf(c, failed, passed)
}
