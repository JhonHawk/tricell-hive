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

// packetResult reads the text of a packet read's result: the members its last
// end line names, and whether every one of them has its exact heading. Without
// an end line the members are unknown and nothing is counted.
func packetResult(text string) (members []string, complete bool) {
	ends := packetEnds(text)
	if len(ends) == 0 {
		return nil, false
	}
	last := ends[len(ends)-1]
	return last.members, last.complete
}

type packetEnd struct {
	members  []string
	complete bool
}

// packetEnds reads every end line of a result, since one shell command may
// print several packets: for each, the members it names and whether each of
// them has its exact heading in the result.
func packetEnds(text string) []packetEnd {
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
	var ends []packetEnd
	for _, l := range lines {
		m := packetEndLine.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		end := packetEnd{complete: true}
		for _, name := range strings.Split(m[1], ",") {
			name = strings.TrimSpace(name)
			end.members = append(end.members, name)
			if !headings[name] {
				end.complete = false
			}
		}
		ends = append(ends, end)
	}
	return ends
}

// packetFilter names the shell programs that print only the lines they match:
// their output can hold every heading and the end line without the leaves'
// text, so they never read a packet whole.
var packetFilter = regexp.MustCompile(`^(awk|grep|egrep|fgrep|rg)$`)

// readsPacketFile reports whether a call reads an installed packet: the Read
// tool on its path, or a shell segment of a program that prints whole files.
func readsPacketFile(e traceEvent) bool {
	switch e.Kind {
	case "read":
		return packetFilePath.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		for _, args := range shellSegments(e.Command, 0) {
			if len(args) > 0 && !packetFilter.MatchString(filepath.Base(args[0])) && shellReadsPattern(args, packetFilePath) {
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
	ends := packetEnds(text)
	if len(ends) == 0 {
		return false, true
	}
	named := false
	for _, end := range ends {
		if slices.Contains(end.members, leafMemberPrefix+leaf) {
			if end.complete {
				return true, false
			}
			named = true
		}
	}
	return false, named
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
