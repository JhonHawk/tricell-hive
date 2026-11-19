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
	// ticketMove is a command that changes a ticket's state.
	ticketMove = regexp.MustCompile(`\bacli\s+jira\s+workitem\s+transition\b|\bgh\s+issue\s+(?:close|reopen|edit)\b`)
	// moveTool is an MCP tool that changes a ticket's state.
	moveTool = regexp.MustCompile(`(?i)transition.*issue|issue.*transition|save_issue|update_issue`)
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
		case missing && move == nil && (e.Kind == "shell" || e.Kind == "exec_item") && ticketMove.MatchString(e.Command):
			move = e
		case missing && move == nil && (e.Kind == "tool" || e.Kind == "mcp_call") && !e.Memory && moveTool.MatchString(e.Tool):
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
		case succeeded(events, i) && (editsInterface(*e, s.Cwd) || launchesRole(*e, "hive-build-frontend")):
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
