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
	// spanishQuestion is a ¿…? question, which may wrap a line but never runs
	// past a blank line.
	spanishQuestion = regexp.MustCompile(`¿(?:[^?¿\n]|\n[^?¿\n])*\?`)
	// sessionCard is the header or the question of a card that asks a session
	// question: the delivery, the code review or the security review. A header
	// counts only when it is the session question's name alone, so «Revisión
	// del plan» on a findings card does not.
	sessionCard = regexp.MustCompile(`(?i)^\W*(?:entrega|delivery|revisi[oó]n|review|seguridad|security)\W*$|\bentrega\b|revisi[oó]n (?:del? )?(?:c[oó]digo|seguridad)|\b(?:code|security) review\b|\bdelivery mode\b`)
	// nextStepBlock is the block that ends a report, from its ▶ heading on.
	nextStepBlock = regexp.MustCompile(`(?s)[^\n]*▶.*$`)
	// imperativeConfirm asks the user to confirm without a question mark, such
	// as «Para arrancar, confirma …»: at the start of a sentence or after a
	// comma or colon, so «el log confirma que…» does not count.
	// RE2's \b is ASCII, so «Confirmé» would end at the «é»: the word must end
	// in a non-letter instead.
	imperativeConfirm = regexp.MustCompile(`(?i)(?:^|[,:]\s*)[*_]*(?:conf[ií]rma(?:me|lo|la)?|confirm)(?:[^\p{L}]|$)`)
	// offeredReplies opens a sentence that lists the replies the user may type,
	// and quotedReply is one of them in quotation marks.
	offeredReplies = regexp.MustCompile(`(?i)\b(?:puedes responder|responde|you can reply|reply with)\b`)
	// imperativeReply is the same offer as an instruction to the user, at the
	// start of a sentence or after a comma or colon, so «el endpoint responde
	// «ok»» does not count; it needs a single quoted reply to fail.
	imperativeReply = regexp.MustCompile(`(?i)(?:^|[,:;]\s*)[*_]*(?:puedes responder|responde|you can reply|reply with)\b`)
	quotedReply     = regexp.MustCompile(`«[^»\n]+»|"[^"\n]+"|“[^”\n]+”`)
	// recommends and codeSpan, with startsWithMode, read a sentence that
	// recommends a text in backticks that begins with a delivery mode as a whole
	// word, such as «Recomiendo `automatic, Codex, sí`»: a session question
	// answered for the user, with no options. A span that only contains a mode,
	// such as `gh pr ready` or `threshold`, does not count, and spans are
	// collected whole so the text between two of them is never read as one.
	recommends     = regexp.MustCompile(`(?i)\b(?:recomiendo|i recommend)\b`)
	codeSpan       = regexp.MustCompile("`[^`\\n]*`")
	startsWithMode = regexp.MustCompile("^`\\s*(?:interactive|automatic|ready|hold|direct-base)\\b")
	// leftPending names a decision as still to confirm, such as «queda por
	// confirmar la entrega»: the question is open and nothing asks it.
	leftPending = regexp.MustCompile(`(?i)\b(?:queda|quedan|falta|faltan)\s+(?:por|de)\s+confirmar\b|\bpendientes?\s+de\s+confirm(?:ar|aci[oó]n)\b|\bstill\s+needs?\s+confirm(?:ing|ation)\b|\bawaiting\s+your\s+confirmation\b`)
	// listedMode is a list item whose label is a delivery mode, with or without
	// a letter, such as "- **A — `interactive` (recomendado).**" or
	// "- **Interactiva (recomendada)** — …". Its first group is the letter.
	listedMode = regexp.MustCompile("(?im)^[ \\t]*(?:[-*•+]|\\d+[.)])[ \\t]+[*_`\\s]*(?:([A-Z])[ \\t]*[—–:.)-][ \\t]*[*_`\\s]*)?(?:interactiv[ea]|autom[aá]tic[ao]?|ready|hold|local|direct-base)\\b[^\\n]*")
	// recommendedItem marks the item offered as the recommendation.
	recommendedItem = regexp.MustCompile(`(?i)recomendad[ao]|recommended|★`)
)

// listedModeOptions returns the list items that offer delivery modes as options
// in a message with no round title: two or more items, lettered or one of them
// recommended, ask the delivery question as a list instead of a round. Items
// that only explain each mode's effect, with no letter and no recommendation,
// are a report.
func listedModeOptions(text string) []string {
	if roundTitleLine.MatchString(text) {
		return nil
	}
	matches := listedMode.FindAllStringSubmatch(text, -1)
	if len(matches) < 2 {
		return nil
	}
	offered := false
	var items []string
	for _, m := range matches {
		offered = offered || m[1] != "" || recommendedItem.MatchString(m[0])
		items = append(items, m[0])
	}
	if !offered {
		return nil
	}
	return items
}

func backtickedMode(s string) bool {
	for _, span := range codeSpan.FindAllString(s, -1) {
		if startsWithMode.MatchString(span) {
			return true
		}
	}
	return false
}

// unquoted drops the text in quotation marks, so a report that quotes another
// session's recommendation does not count as one.
func unquoted(s string) string {
	return quotedReply.ReplaceAllString(s, "")
}

// proseChoices returns the questions of a message that ask the user to choose:
// alternatives in prose, or a session question asked as a yes or no.
func proseChoices(text string) (choices []string, session bool) {
	text = nextStepBlock.ReplaceAllString(fencedCode.ReplaceAllString(text, ""), "")
	if items := listedModeOptions(text); len(items) > 0 {
		choices, session = append(choices, strings.Join(items, "\n")), true
	}
	text = spanishQuestion.ReplaceAllStringFunc(text, func(q string) string { return strings.ReplaceAll(q, "\n", " ") })
	for _, q := range proseQuestions(text) {
		offers := alternatives.MatchString(q)
		isSession := sessionChoice.MatchString(q) && (offers || confirmation.MatchString(q))
		if offers || isSession {
			choices = append(choices, q)
			session = session || isSession
		}
	}
	for _, paragraph := range strings.Split(text, "\n\n") {
		named := sessionChoice.MatchString(paragraph)
		for _, s := range proseSentences(paragraph) {
			replies := len(quotedReply.FindAllString(s, -1))
			switch {
			case strings.HasSuffix(s, "?"):
			case imperativeConfirm.MatchString(s) && sessionChoice.MatchString(s):
				choices = append(choices, s)
				session = true
			case recommends.MatchString(unquoted(s)) && backtickedMode(unquoted(s)):
				choices = append(choices, s)
				session = true
			case leftPending.MatchString(unquoted(s)) && sessionChoice.MatchString(unquoted(s)):
				choices = append(choices, s)
			case imperativeReply.MatchString(s) && replies >= 1 && named:
				choices = append(choices, s)
				session = true
			case offeredReplies.MatchString(s) && replies >= 2:
				choices = append(choices, s)
			}
		}
	}
	return choices, session
}

// proseSentences returns each sentence of the text, its markup trimmed. A
// sentence that holds a quoted reply keeps the stops inside the quotes.
func proseSentences(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		masked := quotedReply.ReplaceAllStringFunc(line, func(q string) string {
			return strings.NewReplacer(".", "\x00", "!", "\x01", "?", "\x02", ":", "\x03").Replace(q)
		})
		for _, s := range strings.Split(sentenceEnd.ReplaceAllString(masked, "$1\n"), "\n") {
			s = strings.NewReplacer("\x00", ".", "\x01", "!", "\x02", "?", "\x03", ":").Replace(s)
			if s = strings.TrimRight(strings.TrimSpace(s), "*_"); s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}

// proseQuestions returns each sentence of the text that ends in a question
// mark, from its ¿ when it has one. A round's title line asks its own question;
// round_text_and_card judges the round, so its title is skipped.
func proseQuestions(text string) []string {
	var out []string
	titles := append(starredTitles(text, roundTitleLine.FindAllStringIndex(text, -1)), numberedRoundTitles(text)...)
	offset := 0
	for _, line := range strings.Split(text, "\n") {
		start := offset
		offset += len(line) + 1
		if isRoundTitleLine(titles, start, start+len(line)) {
			continue
		}
		for _, s := range strings.Split(sentenceEnd.ReplaceAllString(line, "$1\n"), "\n") {
			// A question in bold or italics, such as "**¿interactive o hold?**",
			// ends in its markup.
			s = strings.TrimRight(strings.TrimSpace(s), "*_")
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

// proseChoiceWithoutCard checks the choices asked in prose, on any host, when
// no card follows in that turn. A delivery, review or security
// choice is a session question, so it fails; any other choice may be a question
// about what the user wants, so it is flagged for review. The text the last
// card of the turn requires when it closed empty is not judged.
func proseChoiceWithoutCard(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "prose_choice_without_card", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	unattended := unattendedActivation(events)
	var failed, review, passed []string
	for i, e := range events {
		// A card that carries a session question is the positive evidence.
		if unattended >= 0 && i > unattended {
			break
		}
		if e.Kind == "question" && asksSessionQuestion(e) {
			passed = append(passed, regressionEvidence(e))
			break
		}
	}
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
	return statusOf(c, failed, passed)
}

// asksSessionQuestion reports whether a card's header or question, not its
// option descriptions, asks the delivery, the code review or the security
// review.
func asksSessionQuestion(e traceEvent) bool {
	for _, raw := range list(decodeInput(e)["questions"]) {
		q := object(raw)
		if sessionCard.MatchString(str(q["header"])) || sessionCard.MatchString(str(q["question"])) {
			return true
		}
	}
	return false
}

// cardFollowsOrClosedEmpty reports whether a question follows the message at i
// in its turn, or the last card before it in that turn closed empty, which lets
// the round be written in the text.
func cardFollowsOrClosedEmpty(s auditSession, i int) bool {
	if cardAfterInTurn(s, i) >= 0 {
		return true
	}
	last := cardBeforeInTurn(s, i)
	return last >= 0 && cardResultIsEmpty(s.Host, s.Trace.Events, last, i)
}
