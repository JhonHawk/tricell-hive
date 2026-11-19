package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// flowSkillRead names the installed SKILL.md of a flow that a next-step
	// block starts, under a host's skills folder such as ~/.agents/skills; a
	// read of the repository's own content/skills is research, not a start.
	flowSkillRead = regexp.MustCompile(`/\.[a-z]+/skills/(flow-build|flow-plan|flow-figma-design)/SKILL\.md`)
	// flowCommand is a flow's command in any host's syntax, as a next-step block
	// writes it or the user types it.
	flowCommand = regexp.MustCompile(`(?:^|[\s\x60(])(?:[/$@]|/skill:)flow-(?:build|plan|figma-design|verify|close|deploy)\b`)
	// asksTheUser marks a message that waits on the user's answer or approval.
	asksTheUser = regexp.MustCompile(`(?i)\?|aprobaci[oó]n|approval|approve`)
	// looseRoundLine is a decision asked in one line: its ID or a session
	// question's fixed header at the start, then its recommendation, such as
	// "- **D1 · Delivery.** I recommend hold" or "**Review** — I recommend …",
	// in the first person. A recap or an analysis that cites a decision and
	// what a rule recommends is not a round.
	looseRoundLine = regexp.MustCompile(`(?mi)^[\s#>*-]*\**\s*(?:D\d+|Entrega|Delivery|Revisi[oó]n(?: adversarial)?|Review|Adversarial review)\b\**\s*[·—–:.-][^\n]*(?:\brecomiendo\b|\bI recommend\b|\(recommended\)|\(recomendad[ao]\))`)
	situationLine  = regexp.MustCompile(`(?mi)^[\s*_>]*(?:situation|situaci[oó]n)\s*:`)
	// uiSource is a file that renders an interface; tests are not.
	uiSource     = regexp.MustCompile(`(?i)\.(?:tsx|jsx|vue|svelte|css|scss|html)$`)
	testSource   = regexp.MustCompile(`(?i)(?:^|/)(?:e2e|tests?|__tests__)/|\.(?:test|spec|e2e)\.`)
	hiveRoleName = regexp.MustCompile(`hive-[a-z]+(?:-[a-z]+)*`)
)

// isLooseRound reports whether a message asks one or more decisions in text,
// in the full format or not: a decision ID or a session question's fixed header
// at the start of a line, with a recommendation.
func isLooseRound(text string) bool {
	return looseRoundLine.MatchString(text)
}

func fullRoundFormat(text string) bool {
	return roundOptionStar.MatchString(text) && situationLine.MatchString(text)
}

// readsFlowSkill reports whether an event loads a flow's skill.
func readsFlowSkill(e traceEvent) bool {
	switch e.Kind {
	case "skill_invocation":
		return e.Text == "flow-build" || e.Text == "flow-plan" || e.Text == "flow-figma-design"
	case "read", "search":
		return flowSkillRead.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		return flowSkillRead.MatchString(e.Command)
	}
	return false
}

// flowEnteredWithoutRoute fails when a flow's skill is loaded in the turn that
// follows a message of the agent that asked a question or an approval without
// a next-step block or a flow command, unless the user typed a flow command. A
// flow loaded in the turn of a request is the direct continuation and is not
// judged.
func flowEnteredWithoutRoute(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "flow_entered_without_route", Status: "not_observed", Evidence: []string{}}
	if s.Subagent {
		return c
	}
	events := s.Trace.Events
	judged := map[int]bool{}
	var failed, passed []string
	for i, e := range events {
		turn := s.Turn[i]
		if judged[turn] || !readsFlowSkill(e) {
			continue
		}
		judged[turn] = true
		prev := -1
		for j := i - 1; j >= 0; j-- {
			if s.Turn[j] < turn && events[j].Kind == "text" && events[j].Role == "assistant" {
				prev = j
				break
			}
		}
		if prev < 0 {
			continue
		}
		switch {
		case flowCommand.MatchString(s.Prompts[turn]):
			passed = append(passed, regressionEvidence(e)+" typed command")
		case strings.Contains(events[prev].Text, "▶") || flowCommand.MatchString(events[prev].Text):
			passed = append(passed, regressionEvidence(e)+" after a next-step block")
		case asksTheUser.MatchString(events[prev].Text):
			failed = append(failed, fmt.Sprintf("%s after the question at line %d", regressionEvidence(e), events[prev].Line))
		}
	}
	return statusOf(c, failed, passed)
}

// decisionRoundFormat fails when a message asks decisions in text without the
// full format (a Situation line and the starred option) and no question card
// follows it in its turn. A round the card carries is judged by
// text_round_without_card instead.
func decisionRoundFormat(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "decision_round_format", Status: "not_observed", Evidence: []string{}}
	events := s.Trace.Events
	var failed, passed []string
	for i, e := range events {
		if e.Kind != "text" || e.Role != "assistant" || !(isTextRound(e.Text) || isLooseRound(e.Text)) {
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
		case fullRoundFormat(e.Text):
			passed = append(passed, regressionEvidence(e))
		default:
			failed = append(failed, regressionEvidence(e))
		}
	}
	return statusOf(c, failed, passed)
}

// uiChangeWithoutUXReview fails when a session edits a file that renders an
// interface and launches no hive-review-ux child. A child session is not
// judged: its parent owns the review.
func uiChangeWithoutUXReview(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "ui_change_without_ux_review", Status: "not_observed", Evidence: []string{}}
	if s.Subagent {
		return c
	}
	var edit, review *traceEvent
	for i := range s.Trace.Events {
		e := &s.Trace.Events[i]
		path := filepath.ToSlash(e.Path)
		if edit == nil && e.Kind == "write" && uiSource.MatchString(path) && !testSource.MatchString(path) && inProject(path, s.Cwd) {
			edit = e
		}
		if review == nil && launchesUXReview(*e) {
			review = e
		}
	}
	switch {
	case edit == nil:
		return c
	case review == nil:
		c.Status, c.Evidence = "fail", []string{regressionEvidence(*edit)}
	default:
		c.Status, c.Evidence = "pass", []string{regressionEvidence(*review)}
	}
	return c
}

// inProject reports whether a written path belongs to the session's project:
// relative, or under its working directory. A page written to a temporary
// folder is not the project's interface.
func inProject(path, cwd string) bool {
	if !filepath.IsAbs(path) {
		return true
	}
	return cwd != "" && strings.HasPrefix(path, filepath.ToSlash(cwd)+"/")
}

func launchesUXReview(e traceEvent) bool {
	switch e.Tool {
	case "spawn_agent", "Agent", "Task":
		return strings.Contains(e.Text+string(e.Input), "hive-review-ux")
	}
	return false
}

func statusOf(c criterionAssessment, failed, passed []string) criterionAssessment {
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(passed) > 0:
		c.Status, c.Evidence = "pass", passed
	}
	return c
}
