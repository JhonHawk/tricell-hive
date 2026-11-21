package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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

// readsPacketFile reports whether a call reads an installed packet: the Read
// tool on its path, or a shell segment of a program that prints whole files
// and whose output reaches the result unfiltered.
func readsPacketFile(e traceEvent) bool {
	switch e.Kind {
	case "read":
		return packetFilePath.MatchString(filepath.ToSlash(e.Path))
	case "shell":
		pieces, piped := pipelineSegments(e.Command)
		for k, piece := range pieces {
			for _, args := range shellSegments(piece, 0) {
				if len(args) == 0 || packetFilter.MatchString(filepath.Base(args[0])) || !shellReadsPattern(args, packetFilePath) {
					continue
				}
				if piped[k] || strings.Contains(piece, "|") {
					continue
				}
				if filepath.Base(args[0]) == "sed" && !sedPrintsWholeLines(args) {
					continue
				}
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

// packetFileKey reads the flow and the manifest key of an installed packet
// path: `<skills>/<flow>/stage.md` is key `stage`, `stage-exit.md` is
// `stage-exit`.
var packetFileKey = regexp.MustCompile(`/skills/([a-z0-9]+(?:-[a-z0-9]+)*)/(stage(?:-[a-z]+)?)\.md(?:$|[^\w./-])`)

// packetRepoRoot finds the repository root that holds `content/skills`: the
// nearest parent of the working directory that has it, else the checkout this
// file was compiled from. The audit has no repository root of its own, and a
// transcript's cwd is the audited session's, not this checkout.
func packetRepoRoot() string {
	var starts []string
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		starts = append(starts, filepath.Dir(file))
	}
	for _, dir := range starts {
		for {
			if st, err := os.Stat(filepath.Join(dir, "content", "skills")); err == nil && st.IsDir() {
				return dir
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

// manifestHolds reads the repository's `content/skills/<flow>/assets/stage.json`
// and reports whether the packet at path lists the leaf. known is false when
// the manifest cannot be read or has no such key, so the caller keeps a packet
// whose members it cannot tell as possibly holding the leaf.
func manifestHolds(path, leaf string) (holds, known bool) {
	m := packetFileKey.FindStringSubmatch(filepath.ToSlash(path) + " ")
	root := packetRepoRoot()
	if m == nil || root == "" {
		return false, false
	}
	data, err := os.ReadFile(filepath.Join(root, "content", "skills", m[1], "assets", "stage.json"))
	if err != nil {
		return false, false
	}
	var manifest map[string][]string
	if json.Unmarshal(data, &manifest) != nil {
		return false, false
	}
	members, ok := manifest[m[2]]
	if !ok {
		return false, false
	}
	return slices.Contains(members, leafMemberPrefix+leaf+".md"), true
}

// packetPathOf returns the packet path a call reads, from the Read tool's
// path or the shell command's text.
func packetPathOf(e traceEvent) string {
	if e.Kind == "read" {
		return e.Path
	}
	return packetFilePath.FindString(e.Command)
}

// mayHoldLeaf reports whether the packet a call reads may hold the leaf, from
// the manifest; a packet whose manifest is unavailable may.
func mayHoldLeaf(e traceEvent, leaf string) bool {
	holds, known := manifestHolds(packetPathOf(e), leaf)
	return holds || !known
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
		return false, mayHoldLeaf(e, leaf)
	}
	ends := packetEnds(text)
	if len(ends) == 0 {
		return false, mayHoldLeaf(e, leaf)
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
