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

// packetTruncationMarker matches what a host writes into a result it cut:
// Codex's `…N lines truncated…` and `... N tokens truncated ...`, and Claude
// Code's persisted-output preview, `<persisted-output>` or `Output too large`.
// A result that holds one is never a whole read, even when the end line and
// every heading survived the cut.
var packetTruncationMarker = regexp.MustCompile(`(?i)(?:…|\.\.\.)\s*\d+\s+(?:lines?|tokens?|chars?|characters?|bytes?)\s+truncated\s*(?:…|\.\.\.)|<persisted-output>|Output too large`)

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
	truncated := packetTruncationMarker.MatchString(text)
	var ends []packetEnd
	for _, l := range lines {
		m := packetEndLine.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		end := packetEnd{complete: !truncated}
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

// plainSedRange is the one sed script that prints whole lines by number,
// such as `1,400p`; any other script can select lines by pattern.
var plainSedRange = regexp.MustCompile(`^\d+(?:,(?:\d+|\$))?p$`)

// sedPrintsWholeLines reports whether a sed segment has no script or only a
// plain line-range script. A segment with `-e`, `-f` or any other script
// counts as a filter.
func sedPrintsWholeLines(args []string) bool {
	var positional []string
	for _, a := range args[1:] {
		switch {
		case a == "-n" || a == "-s" || a == "-E" || a == "-r" || a == "--quiet" || a == "--silent":
		case strings.HasPrefix(a, "-"):
			return false
		default:
			positional = append(positional, a)
		}
	}
	if len(positional) == 0 {
		return true
	}
	return packetFilePath.MatchString(positional[0]) || plainSedRange.MatchString(positional[0])
}

// pipelineSegments splits a command at the shell's separators like
// shellSegments and tells, for each piece, whether its output is piped into a
// later command. A piece whose text holds a `|` inside quotes, such as a
// `bash -lc '…'` wrapper, counts as piped, which reads as no whole read.
func pipelineSegments(command string) (pieces []string, piped []bool) {
	var cur strings.Builder
	var quote rune
	escaped := false
	chars := []rune(command)
	flush := func(p bool) {
		pieces = append(pieces, cur.String())
		piped = append(piped, p)
		cur.Reset()
	}
	for i := 0; i < len(chars); i++ {
		ch := chars[i]
		switch {
		case escaped:
			escaped = false
		case ch == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if ch == quote {
				quote = 0
			}
		case ch == '\'' || ch == '"':
			quote = ch
		case ch == ';' || ch == '\n':
			flush(false)
			continue
		case ch == '&' && i+1 < len(chars) && chars[i+1] == '&':
			flush(false)
			i++
			continue
		case ch == '|':
			if i+1 < len(chars) && chars[i+1] == '|' {
				flush(false)
				i++
			} else {
				flush(true)
			}
			continue
		}
		cur.WriteRune(ch)
	}
	flush(false)
	return pieces, piped
}

// packetPassThrough names the shell programs that pass every line they read
// on, whole or as a leading or trailing slice, so a packet piped into them
// still reaches the result as lines a cut marker or a missing end line shows.
var packetPassThrough = regexp.MustCompile(`^(cat|head|tail|nl|less|more|tee)$`)

// quotedSinglePipe matches a lone `|` in a piece, as a `bash -lc '… | …'`
// wrapper holds; `||` is a separator, not a pipe.
var quotedSinglePipe = regexp.MustCompile(`(^|[^|])\|([^|]|$)`)

// readsPacketFile reports whether a call reads an installed packet whole.
func readsPacketFile(e traceEvent) bool {
	whole, _ := packetReadKind(e)
	return whole
}

// packetReadKind judges a call that names a packet. whole: the Read tool on its
// path, or a shell segment of a program that prints whole files whose output
// reaches the result unfiltered or through pass-through programs only. maybe:
// the packet's text reaches the result through a program the audit does not
// know, so the read can be neither counted nor ruled out. A filter such as
// grep, rg, awk or a pattern sed gives neither.
func packetReadKind(e traceEvent) (whole, maybe bool) {
	switch e.Kind {
	case "read":
		return packetFilePath.MatchString(filepath.ToSlash(e.Path)), false
	case "shell":
		pieces, piped := pipelineSegments(e.Command)
		for k, piece := range pieces {
			for _, args := range shellSegments(piece, 0) {
				if len(args) == 0 || packetFilter.MatchString(filepath.Base(args[0])) || !shellReadsPattern(args, packetFilePath) {
					continue
				}
				if filepath.Base(args[0]) == "sed" && !sedPrintsWholeLines(args) {
					continue
				}
				if quotedSinglePipe.MatchString(piece) {
					maybe = true
					continue
				}
				switch downstreamOf(pieces, piped, k) {
				case "pass":
					return true, false
				case "unknown":
					maybe = true
				}
			}
		}
	}
	return false, maybe
}

// downstreamOf classifies what the pipe after piece k does with its output:
// "pass" when it reaches the result whole or through pass-through programs,
// "filter" when a program selects lines by pattern, "unknown" otherwise.
func downstreamOf(pieces []string, piped []bool, k int) string {
	for j := k; j < len(pieces)-1 && piped[j]; j++ {
		segs := shellSegments(pieces[j+1], 0)
		if len(segs) == 0 || len(segs[0]) == 0 {
			return "unknown"
		}
		args := segs[0]
		name := filepath.Base(args[0])
		switch {
		case packetFilter.MatchString(name):
			return "filter"
		case name == "sed":
			if !sedPrintsWholeLines(args) {
				return "filter"
			}
		case !packetPassThrough.MatchString(name):
			return "unknown"
		}
	}
	return "pass"
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
// call reads a packet whose result does not show the leaf whole: no result to
// join, no end line (the audit cannot tell which packet version the session
// read, so any leaf may be in it), a heading missing, or a pipe through a
// program the audit does not know. A call that
// is no packet read, a failed one, or a packet whose end line does not name the
// leaf gives neither.
func packetNamesLeaf(events []traceEvent, i int, leaf string) (read, unverified bool) {
	e := events[i]
	if e.Success != nil && !*e.Success {
		return false, false
	}
	whole, maybe := packetReadKind(e)
	if !whole {
		return false, maybe
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
