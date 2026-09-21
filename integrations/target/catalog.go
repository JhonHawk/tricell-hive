package target

import (
	"path/filepath"
	"strings"
)

// ExpandSkills retains the historical default for callers without a catalogue.
func ExpandSkills(base []Target, sources []string) []Target {
	if sources == nil {
		sources = []string{"content/skills/workspace-conventions/SKILL.md"}
	}
	var out []Target
	for _, t := range base {
		if t.Kind == "block" {
			t.Source = "content/guidance/global.md"
			out = append(out, t)
			continue
		}
		for _, source := range sources {
			parts := strings.Split(source, "/")
			if len(parts) < 4 {
				continue
			}
			n := t
			n.Source = source
			if t.Kind == "symlink" {
				if len(parts) != 4 || parts[3] != "SKILL.md" {
					continue
				}
				n.Path = filepath.Join(filepath.Dir(t.Path), parts[2])
			} else {
				n.Path = filepath.Join(filepath.Dir(filepath.Dir(t.Path)), parts[2], filepath.Join(parts[3:]...))
			}
			if t.LinkTarget != "" {
				n.LinkTarget = filepath.Join(filepath.Dir(t.LinkTarget), parts[2])
			}
			out = append(out, n)
		}
	}
	return out
}
