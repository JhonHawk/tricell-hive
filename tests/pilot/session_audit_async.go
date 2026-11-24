package main

// The asynchronous card returns before the answer, so it is the channel only
// where no waiting card is offered, after the round is written in the text.
// Unattended delegation offers no waiting card, so from its activation only
// the round written first is judged. A child is not judged: its parent owns
// the questions.
func asyncCardWhenWaitingOffered(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "async_card_when_waiting_offered", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	unattended := unattendedActivation(s.Trace.Events)
	var failed, passed []string
	for i, e := range s.Trace.Events {
		if e.Kind != "question" || e.Tool != "request_user_input_async" {
			continue
		}
		offered := s.QuestionOffered && (unattended < 0 || i < unattended)
		if offered || !roundWrittenBefore(s, i) {
			failed = append(failed, regressionEvidence(e))
		} else {
			passed = append(passed, regressionEvidence(e))
		}
	}
	return statusOf(c, failed, passed)
}

func roundWrittenBefore(s auditSession, i int) bool {
	for j := i - 1; j >= 0 && s.Turn[j] == s.Turn[i]; j-- {
		e := s.Trace.Events[j]
		if e.Kind == "text" && e.Role == "assistant" && isTextRound(e.Text) {
			return true
		}
	}
	return false
}
