package main

import (
	"fmt"
	"strings"
	"testing"
)

const (
	pktPath    = "/h/.agents/skills/flow-build/stage.md"
	pktGit     = "hive-core/references/git-commit"
	pktQuest   = "hive-core/references/questions"
	pktChange  = "hive-core/references/change-folder"
	pktEndLine = "<!-- hive packet end: stage · " + pktGit + "," + pktQuest + " -->"
)

func pktText(extraSection string) string {
	return "## " + pktGit + "\n\n# Git commit\n\nRules.\n" + extraSection + "\n## " + pktQuest + "\n\n# Questions\n\nRules.\n\n" + pktEndLine + "\n"
}

func numberedLines(text, sep string) string {
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, l := range lines {
		lines[i] = fmt.Sprintf("%6d%s%s", i+1, sep, l)
	}
	return strings.Join(lines, "\n")
}

func TestPacketResultMembers(t *testing.T) {
	full := pktText("")
	for _, tc := range []struct {
		name, text string
		members    []string
		complete   bool
		endSeen    bool
	}{
		{"plain text", full, []string{pktGit, pktQuest}, true, true},
		{"line numbers with arrows", numberedLines(full, "→"), []string{pktGit, pktQuest}, true, true},
		{"line numbers with tabs", numberedLines(full, "\t"), []string{pktGit, pktQuest}, true, true},
		{"cut at the end", strings.Split(full, "## "+pktQuest)[0], nil, false, false},
		{"cut in the middle, end line intact", strings.Split(full, "## "+pktQuest)[0] + "\n…truncated…\n\n" + pktEndLine + "\n", []string{pktGit, pktQuest}, false, true},
		{"a member heading with a longer name is not the member", strings.ReplaceAll(full, "## "+pktQuest, "## "+pktQuest+"-extra"), []string{pktGit, pktQuest}, false, true},
		{"empty result", "", nil, false, false},
		{"launch acknowledgment", "ok", nil, false, false},
		{"text after the end line", full + "<system-reminder>x</system-reminder>\n", []string{pktGit, pktQuest}, true, true},
	} {
		members, complete := packetResult(tc.text)
		if complete != tc.complete || (tc.endSeen && strings.Join(members, ",") != strings.Join(tc.members, ",")) || (!tc.endSeen && len(members) != 0) {
			t.Errorf("%s: members %v complete %v, want %v %v", tc.name, members, complete, tc.members, tc.complete)
		}
	}
}

func TestPacketOwnSectionIsNotAMember(t *testing.T) {
	events := []traceEvent{
		{Line: 1, Kind: "read", ID: "r1", Path: pktPath},
		{Line: 2, Kind: "tool_result", ID: "r1", Text: numberedLines(pktText("\n## Change folder\n\nOwn section.\n"), "→")},
	}
	if read, _ := packetNamesLeaf(events, 0, "git-commit"); !read {
		t.Fatal("a complete packet does not count for its member")
	}
	if read, unverified := packetNamesLeaf(events, 0, "change-folder"); read || unverified {
		t.Fatalf("the leaf's own %q section counted as the member %s: read=%v unverified=%v", "## Change folder", pktChange, read, unverified)
	}
}

func TestPacketReadJoinsItsResultByID(t *testing.T) {
	good := numberedLines(pktText(""), "→")
	cut := "     1→## " + pktGit + "\n"
	events := []traceEvent{
		{Line: 1, Kind: "read", ID: "a", Path: pktPath},
		{Line: 2, Kind: "shell", ID: "b", Command: "git status"},
		{Line: 3, Kind: "tool_result", ID: "b", Text: good},
		{Line: 4, Kind: "tool_result", ID: "a", Text: cut},
	}
	if read, unverified := packetNamesLeaf(events, 0, "git-commit"); read || !unverified {
		t.Fatalf("another call's complete result was joined to the read: read=%v unverified=%v", read, unverified)
	}
	events[3].Text = good
	if read, _ := packetNamesLeaf(events, 0, "git-commit"); !read {
		t.Fatal("the read's own result was not joined")
	}
	// A result after the cut-off is not seen: the action at that point came first.
	if read, unverified := packetNamesLeaf(events[:3], 0, "git-commit"); read || !unverified {
		t.Fatalf("a result that is not yet in the trace counted: read=%v unverified=%v", read, unverified)
	}
	// A failed read and a read with no ID are no read.
	no := false
	failed := []traceEvent{{Line: 1, Kind: "read", ID: "a", Path: pktPath, Success: &no}, {Line: 2, Kind: "tool_result", ID: "a", Text: good, Success: &no}}
	if read, unverified := packetNamesLeaf(failed, 0, "git-commit"); read || unverified {
		t.Fatalf("a failed read counted: read=%v unverified=%v", read, unverified)
	}
	noID := []traceEvent{{Line: 1, Kind: "read", Path: pktPath}, {Line: 2, Kind: "tool_result", Text: good}}
	if read, unverified := packetNamesLeaf(noID, 0, "git-commit"); read || !unverified {
		t.Fatalf("a result with no ID counted: read=%v unverified=%v", read, unverified)
	}
}

func TestPacketReadByShellNeedsAReadingProgram(t *testing.T) {
	good := pktText("")
	for cmd, want := range map[string]bool{
		"cat " + pktPath:                      true,
		"sed -n 1,400p " + pktPath:            true,
		"git commit -m 'see " + pktPath + "'": false,
		"cat > " + pktPath + " <<'EOF'":       false,
	} {
		events := []traceEvent{{Line: 1, Kind: "shell", ID: "a", Command: cmd}, {Line: 2, Kind: "tool_result", ID: "a", Text: good}}
		if read, _ := packetNamesLeaf(events, 0, "git-commit"); read != want {
			t.Errorf("%q: read=%v, want %v", cmd, read, want)
		}
	}
}

func TestReportEndingsReadThroughAPacket(t *testing.T) {
	text := "## hive-core/references/report-endings\n\n# Report endings\n\nRules.\n\n<!-- hive packet end: stage · hive-core/references/report-endings -->\n"
	events := []traceEvent{{Line: 1, Kind: "read", ID: "a", Path: "/h/.claude/skills/flow-close/stage.md"}, {Line: 2, Kind: "tool_result", ID: "a", Text: numberedLines(text, "→")}}
	if !readsReportEndingsAt(events, 0) {
		t.Fatal("a complete packet that holds report-endings is not a read of it")
	}
	events[1].Text = "     1→## hive-core/references/report-endings\n"
	if readsReportEndingsAt(events, 0) {
		t.Fatal("a truncated packet counted as a read of report-endings")
	}
	events[1].Text = strings.ReplaceAll(text, "report-endings", "close-report")
	if readsReportEndingsAt(events, 0) {
		t.Fatal("a packet without report-endings counted")
	}
}

func TestCostCountsAPacketReadAsAHiveRead(t *testing.T) {
	for _, tc := range []struct{ tool, path, command string }{
		{"Read", "/h/.agents/skills/flow-verify/stage-entry.md", ""},
		{"Read", "/h/.claude/skills/flow-build/stage.md", ""},
		{"Bash", "", "cat /h/.agents/skills/flow-build/stage.md"},
	} {
		if got := hiveCall(tc.tool, tc.path, tc.command); got != "read" {
			t.Errorf("%+v: %q, want read", tc, got)
		}
	}
	if got := hiveCall("Read", "/repo/content/skills/flow-build/stage.md", ""); got != "" {
		t.Errorf("a repository file counted: %q", got)
	}
}

func TestPacketReadOfTwoPacketsCountsEachMember(t *testing.T) {
	exit := "## hive-core/references/review-findings\n\n# Review findings\n\nRules.\n\n<!-- hive packet end: stage-exit · hive-core/references/review-findings -->\n"
	events := []traceEvent{
		{Line: 1, Kind: "shell", ID: "s", Command: "cat " + pktPath + " /h/.agents/skills/flow-verify/stage-exit.md"},
		{Line: 2, Kind: "tool_result", ID: "s", Text: pktText("") + exit},
	}
	for _, leaf := range []string{"git-commit", "review-findings"} {
		if read, unverified := packetNamesLeaf(events, 0, leaf); !read || unverified {
			t.Errorf("%s: one cat of two packets: read=%v unverified=%v, want read", leaf, read, unverified)
		}
	}
}

func TestPacketFilteredByAShellFilterIsNoRead(t *testing.T) {
	for _, cmd := range []string{"grep -e '^## ' -e 'packet end' " + pktPath, "rg '^## ' " + pktPath, "awk '/^## /' " + pktPath} {
		if readsPacketFile(traceEvent{Kind: "shell", Command: cmd}) {
			t.Errorf("%q counted as a whole read of the packet", cmd)
		}
	}
	if !readsPacketFile(traceEvent{Kind: "shell", Command: "cat " + pktPath}) {
		t.Error("cat of the packet is not a read")
	}
}

func TestPacketUnpaddedTabNumbersAreStripped(t *testing.T) {
	lines := strings.Split(strings.TrimRight(pktText(""), "\n"), "\n")
	for i, l := range lines {
		lines[i] = fmt.Sprintf("%d\t%s", i+1, l)
	}
	if _, complete := packetResult(strings.Join(lines, "\n")); !complete {
		t.Fatal("Claude Code's unpadded form, 1\\t…, is not read as a complete packet")
	}
}

func TestPacketWithoutEndLineIsUnverifiedOnlyForItsManifestMembers(t *testing.T) {
	cut := numberedLines(pktText(""), "→")
	cut = strings.Split(cut, "\n")[0]
	for _, tc := range []struct {
		name   string
		events []traceEvent
	}{
		{"shell head no end line", []traceEvent{{Line: 1, Kind: "shell", ID: "a", Command: "head -20 /h/.agents/skills/hive-mode/stage.md"}, {Line: 2, Kind: "tool_result", ID: "a", Text: cut}}},
		{"no result", []traceEvent{{Line: 1, Kind: "read", ID: "a", Path: "/h/.agents/skills/hive-mode/stage.md"}}},
	} {
		if read, unverified := packetNamesLeaf(tc.events, 0, "git-commit"); read || unverified {
			t.Errorf("%s: hive-mode's packet does not list git-commit: read=%v unverified=%v", tc.name, read, unverified)
		}
		if _, unverified := packetNamesLeaf(tc.events, 0, "questions"); !unverified {
			t.Errorf("%s: hive-mode's packet lists questions, want unverified", tc.name)
		}
	}
	flowBuild := []traceEvent{{Line: 1, Kind: "read", ID: "a", Path: pktPath}}
	if _, unverified := packetNamesLeaf(flowBuild, 0, "git-commit"); !unverified {
		t.Error("flow-build's packet lists git-commit, want unverified without a result")
	}
}

func TestPacketReadThroughAFilterOrPatternScriptIsNoRead(t *testing.T) {
	for _, cmd := range []string{
		"cat " + pktPath + " | grep -E '^## |hive packet end'",
		"sed -n '/^## /p;/packet end/p' " + pktPath,
		"sed '/^## /!d' " + pktPath,
		"cat " + pktPath + " | head -400",
	} {
		if readsPacketFile(traceEvent{Kind: "shell", Command: cmd}) {
			t.Errorf("%q counted as a whole read of the packet", cmd)
		}
	}
	for _, cmd := range []string{"sed -n '1,400p' " + pktPath, "sed -n 1,400p " + pktPath, "cat " + pktPath, "cat " + pktPath + " && git status"} {
		if !readsPacketFile(traceEvent{Kind: "shell", Command: cmd}) {
			t.Errorf("%q is a whole read of the packet", cmd)
		}
	}
}

func TestPacketResultWithATruncationMarkerIsIncomplete(t *testing.T) {
	full := pktText("")
	for name, marker := range map[string]string{
		"codex lines":      "…12 lines truncated…",
		"codex tokens":     "... 3400 tokens truncated ...",
		"claude persisted": "<persisted-output>\nOutput too large (60KB). Full output saved to: /x\n</persisted-output>",
		"claude too large": "Output too large (60KB).",
	} {
		text := full + marker + "\n"
		if _, complete := packetResult(text); complete {
			t.Errorf("%s: a result with a truncation marker counted as complete", name)
		}
	}
}
