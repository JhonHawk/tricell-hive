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
	text := "CLIs\n  claude  detected  Hive release 4f8d96b9ce4f (verified)  CLI version 2.1.284 (Claude Code)\n    /a/very/long/path/that/does/not/fit/in/forty/columns/at/all/AGENTS.md\nshort"
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
	if !strings.HasSuffix(strings.TrimSpace(got[1])+" "+joined, "CLI version 2.1.284 (Claude Code)") {
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
