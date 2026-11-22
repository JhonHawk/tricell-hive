package main

import (
	"regexp"
	"strings"
)

var (
	// statesTopic is the persistent-memory topic of a tracker's states.
	statesTopic = regexp.MustCompile(`tracker/states/`)
	// noMemories is Engram's answer to a search that found nothing.
	noMemories = regexp.MustCompile(`(?i)no memories found`)
	// readOnlyTool is a tracker tool that only reads, such as one that lists a
	// ticket's transitions; its name may still contain "transition".
	readOnlyTool = regexp.MustCompile(`(?i)(?:^|_)(?:get|list|read|search|fetch|view)[A-Z_]`)
	// acliTransition is acli's ticket move, which leafActionsIn does not know.
	acliTransition = regexp.MustCompile(`\bacli\s+jira\s+workitem\s+transition\b`)
	// memoryWrite is a write to persistent memory, by tool or by Engram's CLI.
	memoryWriteTool = regexp.MustCompile(`(?i)mem_save|mem_update`)
	memoryWriteCLI  = regexp.MustCompile(`\bengram\s+save\b`)
	// flowVerifyRead is the installed flow-verify skill, read or invoked.
	flowVerifyRead = regexp.MustCompile(`/\.[a-z]+/skills/flow-verify/SKILL\.md`)
	// tryItHeading is the heading of the block a report carries for a change
	// that can be seen or run, in the languages Hive's sessions use.
	tryItHeading = regexp.MustCompile(`(?mi)^[\s#>*_-]*(?:pru[eé]balo(?: t[uú])?|try it)\b`)
)

// eventInput is the raw text of an event's arguments and command.
func eventInput(e traceEvent) string {
	return string(e.Input) + "\n" + e.Command
}

// trackerStatesCacheWritten fails when a session searches the states topic
// and finds nothing, then moves a ticket, and never writes the topic.
func trackerStatesCacheWritten(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "tracker_states_cache_written", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	missing, written := false, false
	var move *traceEvent
	for i := range events {
		e := &events[i]
		in := eventInput(*e)
		switch {
		case (e.Kind == "tool" || e.Kind == "mcp_call") && strings.Contains(e.Tool, "mem_search") && statesTopic.MatchString(in):
			if noMemories.MatchString(e.Text) || noMemories.MatchString(resultText(events, i)) {
				missing = true
			}
		case (e.Kind == "tool" || e.Kind == "mcp_call") && memoryWriteTool.MatchString(e.Tool) && statesTopic.MatchString(in):
			written = true
		case (e.Kind == "shell" || e.Kind == "exec_item") && memoryWriteCLI.MatchString(e.Command) && statesTopic.MatchString(e.Command):
			written = true
		case missing && move == nil && movesTicket(*e, s.Cwd):
			move = e
		}
	}
	switch {
	case !missing || move == nil:
	case written:
		c.Status, c.Evidence = "pass", []string{regressionEvidence(*move)}
	default:
		c.Status, c.Evidence = "fail", []string{regressionEvidence(*move)}
	}
	return c
}

// movesTicket reports whether an event changes a ticket's state: acli's
// transition, or what leafActionsIn counts as a ticket move, read the same way
// from Codex's completed MCP calls and executed commands.
func movesTicket(e traceEvent, cwd string) bool {
	switch e.Kind {
	case "mcp_call":
		e.Kind = "tool"
	case "exec_item":
		e.Kind = "shell"
	}
	if e.Kind == "shell" && acliTransition.MatchString(e.Command) {
		return true
	}
	if e.Kind == "tool" && (e.Memory || readOnlyTool.MatchString(e.Tool)) {
		return false
	}
	for _, a := range leafActionsIn(e, cwd) {
		if a.class == "ticket move" {
			return true
		}
	}
	return false
}

// launchesRole reports whether an event starts a child of the role, which
// edits the interface on the session's behalf.
func launchesRole(e traceEvent, role string) bool {
	switch e.Tool {
	case "spawn_agent":
		return e.Text == role || roleArgument(e) == role
	case "Agent", "Task", "spawn_subagent":
		return roleArgument(e) == role
	}
	return false
}

// resultText is the text of the tool result that answers the call at i.
func resultText(events []traceEvent, i int) string {
	if events[i].ID == "" {
		return ""
	}
	for j := i + 1; j < len(events); j++ {
		if events[j].Kind == "tool_result" && events[j].ID == events[i].ID {
			return events[j].Text
		}
	}
	return ""
}

// tryItBlockInReport fails when a session edits a file that renders an
// interface, loads flow-verify, and its last message carries no Try it block.
func tryItBlockInReport(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "try_it_block_in_report", Status: "not_observed", Evidence: []string{}}
	if child(s) {
		return c
	}
	events := s.Trace.Events
	edited, verified := false, false
	var last *traceEvent
	for i := range events {
		e := &events[i]
		switch {
		case succeeded(events, i) && (editsInterface(*e, s.Cwd) || launchesRole(*e, "hive-build-frontend") || launchesRole(*e, "hive-build-mobile")):
			edited = true
		case e.Kind == "skill_invocation" && strings.Contains(e.Text+string(e.Input), "flow-verify"):
			verified = true
		case (e.Kind == "read" || e.Kind == "search") && flowVerifyRead.MatchString(e.Path):
			verified = true
		case e.Kind == "shell" && flowVerifyRead.MatchString(e.Command) && readCommand.MatchString(e.Command):
			verified = true
		case e.Kind == "text" && e.Role == "assistant":
			last = e
		}
	}
	switch {
	case !edited || !verified || last == nil:
	case tryItHeading.MatchString(last.Text):
		c.Status, c.Evidence = "pass", []string{regressionEvidence(*last)}
	default:
		c.Status, c.Evidence = "fail", []string{regressionEvidence(*last)}
	}
	return c
}
