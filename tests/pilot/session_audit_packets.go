package main

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// A stage packet is the file a flow tells the agent to read whole: the text of
// several rule leaves, each under a heading `## <member>`, ending in a line
// `<!-- hive packet end: <key> · <member>,<member>,… -->` that names every
// member. A read of the packet counts as a read of a leaf only when its result
// holds that end line and an exact heading for every member it names, so a
// result cut at its end, or cut in the middle by the host, is not counted, and
// a section a leaf owns, such as `## Change folder`, is never mistaken for the
// member `hive-core/references/change-folder`.
var (
	// packetFilePath names an installed packet: a flow's stage file under any
	// host's skills folder. The repository holds none.
	packetFilePath = regexp.MustCompile(`(?:^|[^\w.-])[^\s"'` + "`" + `]*/skills/[a-z0-9]+(?:-[a-z0-9]+)*/stage(?:-[a-z]+)?\.md(?:$|[^\w./-])`)
	packetEndLine  = regexp.MustCompile(`^<!-- hive packet end: stage(?:-[a-z]+)? · (.+) -->\s*$`)
	// lineNumberPrefix is what Claude Code's Read tool puts before each line,
	// a number and an arrow (or a tab). Other hosts' prefixes were not measured:
	// a result that carries one the pattern does not strip reads as incomplete.
	lineNumberPrefix = regexp.MustCompile(`^\s*\d+(?:\t|→)`)
)

const leafMemberPrefix = "hive-core/references/"

// packetResult reads the text of a packet read's result: the members its end
// line names, and whether every one of them has its exact heading. Without the
// end line the members are unknown and nothing is counted.
func packetResult(text string) (members []string, complete bool) {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = lineNumberPrefix.ReplaceAllString(strings.TrimRight(l, "\r"), "")
	}
	headings := map[string]bool{}
	for _, l := range lines {
		if m, ok := strings.CutPrefix(strings.TrimRight(l, " \t"), "## "); ok {
			headings[m] = true
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		m := packetEndLine.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		for _, name := range strings.Split(m[1], ",") {
			members = append(members, strings.TrimSpace(name))
		}
		for _, name := range members {
			if !headings[name] {
				return members, false
			}
		}
		return members, true
	}
	return nil, false
}

// readsPacketFile reports whether a call reads an installed packet: the Read
// tool on its path, or a shell segment of a program that prints files.
func readsPacketFile(e traceEvent) bool {
	switch e.Kind {
	case "read":
		return packetFilePath.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		for _, args := range shellSegments(e.Command, 0) {
			if shellReadsPattern(args, packetFilePath) {
				return true
			}
		}
	}
	return false
}

// resultTextOf returns the text of the result of the call at i, joined by the
// call's ID among the events that follow it.
func resultTextOf(events []traceEvent, i int) (string, bool) {
	id := events[i].ID
	if id == "" {
		return "", false
	}
	for _, e := range events[i+1:] {
		if e.Kind == "tool_result" && e.ID == id {
			if e.Success != nil && !*e.Success {
				return "", false
			}
			return e.Text, true
		}
	}
	return "", false
}

// packetNamesLeaf judges the call at i as a read of the leaf through a packet.
// read: the packet's result is complete and names the leaf. unverified: the
// call reads a packet that may hold the leaf but whose result does not show
// it whole: no result to join, no end line, or a heading missing. A call that
// is no packet read, a failed one, or a packet whose end line does not name the
// leaf gives neither.
func packetNamesLeaf(events []traceEvent, i int, leaf string) (read, unverified bool) {
	e := events[i]
	if (e.Success != nil && !*e.Success) || !readsPacketFile(e) {
		return false, false
	}
	text, ok := resultTextOf(events, i)
	if !ok {
		return false, true
	}
	members, complete := packetResult(text)
	if len(members) == 0 {
		return false, true
	}
	if !slices.Contains(members, leafMemberPrefix+leaf) {
		return false, false
	}
	return complete, !complete
}

// readsReportEndingsAt is readsReportEndings for the call at i, which also
// counts a complete packet that holds the report-endings leaf.
func readsReportEndingsAt(events []traceEvent, i int) bool {
	if readsReportEndings(events[i]) {
		return true
	}
	read, _ := packetNamesLeaf(events, i, "report-endings")
	return read
}
