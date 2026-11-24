package main

// asyncCardWhenWaitingOffered checks each call to request_user_input_async of a
// main session. The asynchronous card does not wait for the answer, so it is
// the channel only where the host offers no waiting card, and then only after
// the round is written in the text of the same turn. A call fails when the
// session offers request_user_input in any turn, or when no assistant text
// earlier in its turn is a round. A child session is not judged: its parent
// owns the questions. The evidence names the line and the tool, never the text.
func asyncCardWhenWaitingOffered(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "async_card_when_waiting_offered", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	var failed, passed []string
	for i, e := range s.Trace.Events {
		if e.Kind != "question" || e.Tool != "request_user_input_async" {
			continue
		}
		if s.QuestionOffered || !roundWrittenBefore(s, i) {
			failed = append(failed, regressionEvidence(e))
		} else {
			passed = append(passed, regressionEvidence(e))
		}
	}
	return statusOf(c, failed, passed)
}

// roundWrittenBefore reports whether an assistant text before the event at i,
// in its turn, is a round written in the text.
func roundWrittenBefore(s auditSession, i int) bool {
	for j := i - 1; j >= 0 && s.Turn[j] == s.Turn[i]; j-- {
		e := s.Trace.Events[j]
		if e.Kind == "text" && e.Role == "assistant" && isTextRound(e.Text) {
			return true
		}
	}
	return false
}
