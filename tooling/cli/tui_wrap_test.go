package main

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestScrollBoxHangingIndentKeepsTheStructure covers the UX review of #46: a
// line of the read-only views that does not fit continues under its own
// indentation plus two spaces, so rows, paths and values keep their shape.
func TestScrollBoxHangingIndentKeepsTheStructure(t *testing.T) {
	box := newScrollBox()
	box.hanging = true
	box.vp.SetWidth(40)
	box.vp.SetHeight(20)
	text := "CLIs\n  claude  detected  Hive 4f8d96b9ce4f (verified)  CLI 2.1.284 (Claude Code)\n    /a/very/long/path/that/does/not/fit/in/forty/columns/at/all/AGENTS.md\nshort"
	box.setText(text)
	lines := strings.Split(box.vp.View(), "\n")
	var got []string
	for _, l := range lines {
		l = strings.TrimRight(l, " ")
		if l == "" {
			continue
		}
		if w := lipgloss.Width(l); w > 40 {
			t.Fatalf("line %q is %d wide", l, w)
		}
		got = append(got, l)
	}
	if got[0] != "CLIs" || !strings.HasPrefix(got[1], "  claude") || got[len(got)-1] != "short" {
		t.Fatalf("lines: %q", got)
	}
	var claude, path []string
	for _, l := range got[2 : len(got)-1] {
		if strings.HasPrefix(l, "    /") && len(path) == 0 {
			path = append(path, l)
			continue
		}
		if len(path) > 0 {
			path = append(path, l)
		} else {
			claude = append(claude, l)
		}
	}
	for _, l := range claude {
		if !strings.HasPrefix(l, "    ") || strings.HasPrefix(l, "     ") {
			t.Errorf("continuation of the CLI row is not indented by 4: %q", l)
		}
	}
	for _, l := range path[1:] {
		if !strings.HasPrefix(l, "      ") {
			t.Errorf("continuation of the path is not indented by 6: %q", l)
		}
	}
	joined := strings.Join(strings.Fields(strings.Join(claude, " ")), " ")
	if !strings.HasSuffix(strings.TrimSpace(got[1])+" "+joined, "CLI 2.1.284 (Claude Code)") {
		t.Errorf("the CLI row lost text: %q + %q", got[1], joined)
	}
}

func TestScrollBoxWithoutHangingIndentIsUnchanged(t *testing.T) {
	box := newScrollBox()
	box.vp.SetWidth(20)
	box.vp.SetHeight(10)
	box.setText("  one two three four five six seven")
	if first := strings.Split(box.vp.View(), "\n")[1]; strings.HasPrefix(first, " ") {
		t.Fatalf("a plain box indented its continuation: %q", first)
	}
}

// longPath is 120 characters and has a "/" every few characters and no "-".
func longPath() string {
	p := "/home/user/Development/projects/tricell/tricell_hive/.claude/worktrees/tui_polish/content/skills/adversarial_research/SKILL.md"
	return p[:120]
}

// assertBreaksAfterSlash checks that the lines rejoin to the original, none
// exceeds width, and every line but the last ends in "/".
func assertBreaksAfterSlash(t *testing.T, lines []string, original string, width int) {
	t.Helper()
	var joined string
	for i, l := range lines {
		l = strings.TrimSpace(l)
		if w := lipgloss.Width(l); w > width {
			t.Errorf("line %q is %d wide, limit %d", l, w, width)
		}
		if i < len(lines)-1 && !strings.HasSuffix(l, "/") {
			t.Errorf("line %d %q breaks in the middle of a name", i, l)
		}
		joined += l
	}
	if joined != original {
		t.Errorf("rejoined = %q, want %q", joined, original)
	}
}

// TestWrapHangingBreaksLongPathsAfterASlash (K2): a path wider than the line
// breaks after a "/" instead of in the middle of a name, keeping the indent.
func TestWrapHangingBreaksLongPathsAfterASlash(t *testing.T) {
	path := longPath()
	for _, width := range []int{40, 80} {
		lines := wrapHanging("    "+path, width)
		if len(lines) < 2 {
			t.Fatalf("width %d: %d lines", width, len(lines))
		}
		if !strings.HasPrefix(lines[0], "    /") {
			t.Errorf("width %d: first line %q lost its indent", width, lines[0])
		}
		for _, l := range lines[1:] {
			if !strings.HasPrefix(l, "      ") || strings.HasPrefix(l, "       ") {
				t.Errorf("width %d: continuation %q is not indented by 6", width, l)
			}
		}
		assertBreaksAfterSlash(t, lines, path, width)
	}
}

// TestWrapLinesBreaksLongPathsAfterASlash (K2): the CLIs view note wraps with
// wrapLines, which breaks a path that does not fit after a "/" too.
func TestWrapLinesBreaksLongPathsAfterASlash(t *testing.T) {
	path := longPath()
	for _, width := range []int{40, 80} {
		assertBreaksAfterSlash(t, wrapLines(path, width), path, width)
		lines := wrapLines("Detail: "+path+" differs from what Hive expects there", width)
		if got := strings.Join(strings.Fields(strings.Join(lines, " ")), ""); !strings.Contains(got, path) {
			t.Errorf("width %d: the path was cut: %q", width, lines)
		}
	}
}

// TestWrapKeepsShortTokensWhole (K2): only a token wider than the line breaks
// at a slash; a path that fits moves whole to the next line, and a token
// without a slash still breaks hard.
func TestWrapKeepsShortTokensWhole(t *testing.T) {
	lines := wrapLines("see the file /a/b/c.md now", 20)
	if strings.Join(lines, "|") != "see the file|/a/b/c.md now" {
		t.Errorf("a token that fits was split: %q", lines)
	}
	word := strings.Repeat("x", 100)
	for _, lines := range [][]string{wrapLines(word, 40), wrapHanging("  "+word, 40)} {
		var joined string
		for _, l := range lines {
			if w := lipgloss.Width(l); w > 40 {
				t.Errorf("line %q is %d wide", l, w)
			}
			joined += strings.TrimSpace(l)
		}
		if len(lines) < 3 || joined != word {
			t.Errorf("a token with no slash was not hard-broken to fit: %q", lines)
		}
	}
}
