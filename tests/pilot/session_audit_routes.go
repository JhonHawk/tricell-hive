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
	flowCommand = regexp.MustCompile(`(?:^|[\s\x60(>])(?:[/$@]|/skill:)flow-(?:build|plan|figma-design|verify|close|deploy)\b`)
	// asksTheUser marks a message that waits on the user's answer or approval:
	// a question mark that ends a sentence, outside code and links, or the
	// words for approval.
	asksTheUser = regexp.MustCompile(`(?i)\?(?:\s|\*|$)|aprobaci[oó]n|approval|approve`)
	codeOrLink  = regexp.MustCompile("(?s)\x60\x60\x60.*?\x60\x60\x60|\x60[^\x60\n]*\x60|https?://\\S+")
	// bareReply is a prompt that only answers what the agent asked: a letter,
	// a number or an acceptance, as a reply to a question or a proposal reads.
	bareReply = regexp.MustCompile(`(?i)^\s*(?:[a-e]|\d{1,2}|ok|okay|s[ií]|yes|dale|va|listo|adelante|go|de acuerdo|con los recomendados|approved?|aprobad[oa])\s*[.!]?\s*$`)
	// looseRoundLine is a decision asked in one line: its ID or a session
	// question's fixed header at the start, then its recommendation, such as
	// "- **D1 · Delivery.** I recommend hold" or "**Review** — I recommend …",
	// in the first person. A recap or an analysis that cites a decision and
	// what a rule recommends is not a round.
	looseRoundLine = regexp.MustCompile(`(?mi)^[\s#>*-]*\**\s*(?:D\d+|Entrega|Delivery|Revisi[oó]n(?: adversarial)?|Review|Adversarial review)\b\**\s*[·—–:.-][^\n]*(?:\brecomiendo\b|\bI recommend\b)[^\n]*(?:\belige\b|\bchoose\b|\bsi prefieres\b|\bif you prefer\b|\bor\b|\bo\b)`)
	situationLine  = regexp.MustCompile(`(?mi)^[\s*_>]*(?:situation|situaci[oó]n)\s*:`)
	// uiSource is a file that renders an interface; tests are not.
	uiSource   = regexp.MustCompile(`(?i)\.(?:tsx|jsx|vue|svelte|css|scss|html)$`)
	testSource = regexp.MustCompile(`(?i)(?:^|/)(?:e2e|tests?|__tests__)/|\.(?:test|spec|e2e)\.`)
	// notAppSource holds pages that render no app screen: support material,
	// reports and documents.
	notAppSource = regexp.MustCompile(`(?i)(?:^|/)(?:_support|reports?|docs)/`)
	// readCommand is a shell command that prints a file, as a skill load does;
	// a test or a listing of the path is not a read.
	readCommand = regexp.MustCompile(`(?:^|[\s;&|(])(?:cat|sed|head|tail|nl|bat|less|more|awk|rg|grep)\b`)
	// inPlaceEdit opens a shell segment that rewrites its file arguments, and
	// redirectTarget is the file a segment's output is written into.
	inPlaceEdit    = regexp.MustCompile(`^\s*(?:sed\s+-i|perl\s+-\w*i|tee\b)`)
	redirectTarget = regexp.MustCompile(`(?:^|[^2&>])>>?\s*([^&\s>|;]+)`)
	commandParts   = regexp.MustCompile(`&&|\|\||[;|\n]`)
	fileArgument   = regexp.MustCompile(`[\w./()\[\]@-]+\.\w+`)
	// spawnRole is the role a spawn call names, such as agent_type: "hive-review-ux".
	spawnRole = regexp.MustCompile(`spawn_agent\(\s*\{[^}]*?\b(?:agent_type|subagent_type|agent|role)\s*:\s*["'\x60](hive-[a-z-]+)`)
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
		return flowSkillRead.MatchString(e.Command) && readCommand.MatchString(e.Command)
	}
	return false
}

// child reports whether a transcript belongs to a delegated child, whose
// parent owns the questions, the routes and the reviews.
func child(s auditSession) bool { return s.Subagent || s.Sidechain }

// succeeded reports whether the call at i did not end in an error, such as a
// tool call the user rejected.
func succeeded(events []traceEvent, i int) bool {
	for j := i + 1; j < len(events); j++ {
		if events[j].Kind == "tool_result" && events[j].ID == events[i].ID {
			return events[j].Success == nil || *events[j].Success
		}
	}
	return true
}

// flowEnteredWithoutRoute checks how a flow's skill came to be loaded. It
// fails when the load follows, in the same turn, a question card's answer, or
// when the user's prompt is a bare reply (a letter, a number or an acceptance)
// to a message of the agent that asked a question or an approval without a
// next-step block or a flow command. A longer prompt after such a message is
// flagged for review: it may be a new request that names the change, which
// only judgment tells. A typed flow command passes, and a flow loaded in the
// first turn is the direct continuation and is not judged.
func flowEnteredWithoutRoute(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "flow_entered_without_route", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	judged := map[int]bool{}
	var failed, review, passed []string
	for i, e := range events {
		turn := s.Turn[i]
		if judged[turn] || !readsFlowSkill(e) {
			continue
		}
		judged[turn] = true
		card, prev := -1, -1
		for j := i - 1; j >= 0; j-- {
			if s.Turn[j] == turn && events[j].Kind == "question" && card < 0 {
				card = j
			}
			// The agent's last output before this turn: its message, or a card
			// whose answer arrives as the prompt of this turn.
			if s.Turn[j] < turn && (events[j].Kind == "question" || events[j].Kind == "text" && events[j].Role == "assistant") {
				prev = j
				break
			}
		}
		prompt := s.Prompts[turn]
		switch {
		case flowCommand.MatchString(prompt):
			passed = append(passed, regressionEvidence(e)+" typed command")
		case card >= 0:
			failed = append(failed, fmt.Sprintf("%s after the card at line %d", regressionEvidence(e), events[card].Line))
		case prev < 0:
		case events[prev].Kind == "text" && (strings.Contains(events[prev].Text, "▶") || flowCommand.MatchString(events[prev].Text)):
			passed = append(passed, regressionEvidence(e)+" after a next-step block")
		case events[prev].Kind == "question" || asksTheUser.MatchString(codeOrLink.ReplaceAllString(events[prev].Text, "")):
			line := fmt.Sprintf("%s after the question at line %d", regressionEvidence(e), events[prev].Line)
			if bareReply.MatchString(prompt) {
				failed = append(failed, line)
			} else {
				review = append(review, line)
			}
		}
	}
	if len(failed) == 0 && len(review) > 0 {
		c.Status, c.Evidence = "review", review
		return c
	}
	return statusOf(c, failed, passed)
}

// roundFormat reports, for a message, whether it asks decisions in text and
// whether every one of them has the full format. A decision is a block under a
// "D<n> —" title with lettered options, a row of a table round, or a loose
// one-line round; each needs its own Situation line and starred option.
func roundFormat(text string) (asks, full bool) {
	full = true
	if isLooseRound(text) || len(tableRounds(text)) > 0 {
		asks, full = true, false
	}
	for _, block := range decisionBlocks(text) {
		if !lettered(block) {
			continue
		}
		asks = true
		full = full && fullRoundFormat(block)
	}
	return asks, full
}

// lettered reports whether a block lists options A and B on lines of their own.
func lettered(block string) bool {
	letters := map[byte]bool{}
	for _, line := range optionLine.FindAllString(block, -1) {
		if l := strings.TrimLeft(line, " \t\r\n#>*-"); l != "" {
			letters[l[0]] = true
		}
	}
	return letters['A'] && letters['B']
}

// decisionRoundFormat fails when a message asks decisions in text and one of
// them lacks the full format (a Situation line and the starred option), and no
// question card follows it in its turn. A round the card carries is judged by
// text_round_without_card instead.
func decisionRoundFormat(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "decision_round_format", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	var failed, passed []string
	for i, e := range events {
		if e.Kind != "text" || e.Role != "assistant" {
			continue
		}
		asks, full := roundFormat(e.Text)
		if !asks {
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
		case full:
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
	if child(s) {
		return c
	}
	events := s.Trace.Events
	var edit, review *traceEvent
	for i := range events {
		e := &events[i]
		if !succeeded(events, i) {
			continue
		}
		if edit == nil && editsInterface(*e, s.Cwd) {
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

// editsInterface reports whether an event writes a file of the project that
// renders an interface: a write call, or a shell command that rewrites one.
func editsInterface(e traceEvent, cwd string) bool {
	var paths []string
	switch {
	case e.Kind == "write":
		paths = []string{e.Path}
	case e.Kind == "shell":
		paths = shellWrites(e.Command)
	}
	for _, p := range paths {
		p = filepath.ToSlash(p)
		if uiSource.MatchString(p) && !testSource.MatchString(p) && !notAppSource.MatchString(p) && inProject(p, cwd) {
			return true
		}
	}
	return false
}

// shellWrites returns the files a shell command writes: the arguments of an
// in-place edit or of tee, and the target of each output redirection.
func shellWrites(cmd string) []string {
	var paths []string
	for _, seg := range commandParts.Split(cmd, -1) {
		if inPlaceEdit.MatchString(seg) {
			paths = append(paths, fileArgument.FindAllString(seg, -1)...)
		}
		for _, m := range redirectTarget.FindAllStringSubmatch(seg, -1) {
			paths = append(paths, m[1])
		}
	}
	return paths
}

// inProject reports whether a written path belongs to the session's project:
// relative, or under its working directory. A page written to a temporary
// folder is not the project's interface.
func inProject(path, cwd string) bool {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return !strings.HasPrefix(clean, "..")
	}
	return cwd != "" && strings.HasPrefix(clean, filepath.Clean(cwd)+string(filepath.Separator))
}

func launchesUXReview(e traceEvent) bool {
	switch e.Tool {
	case "spawn_agent":
		return e.Text == "hive-review-ux" || roleArgument(e) == "hive-review-ux"
	case "Agent", "Task", "spawn_subagent":
		return roleArgument(e) == "hive-review-ux"
	}
	return false
}

// roleArgument returns the role a child-launching call names in its arguments.
func roleArgument(e traceEvent) string {
	m := decodeInput(e)
	for _, k := range []string{"subagent_type", "agent_type", "agent", "role"} {
		if v := str(m[k]); v != "" {
			return v
		}
	}
	return ""
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
