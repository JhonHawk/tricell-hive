// doctor_project_changes.go holds the two warn-only checks of the change
// records under the project's Specs path: acceptance-criteria coverage and
// change folder naming. They only read; neither blocks.
package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	// changeKebabPattern is an active change folder name: English kebab-case.
	changeKebabPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	// changeArchivePattern is an archived entry: a date, then kebab-case.
	changeArchivePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-[a-z0-9]+(-[a-z0-9]+)*$`)
	// acDefinitionPattern matches a criterion defined in proposal.md: `- AC1 — …`.
	acDefinitionPattern = regexp.MustCompile(`^\s*[-*]\s+AC(\d+)\b`)
	// closesLinePattern matches a tasks.md line that lists closed criteria.
	closesLinePattern = regexp.MustCompile(`\*\*Closes:?\*\*:?(.*)$`)
	// acReferencePattern matches `AC3`, and the ranges `AC11–AC15` and `AC20–21`.
	acReferencePattern = regexp.MustCompile(`AC(\d+)(?:\s*[-–—]\s*(?:AC)?(\d+))?`)
)

const maxChangeDirs = 500 // entries examined per directory

// changeFindings returns the warnings for the change records under the Specs
// value of the section: "" or a missing changes folder yields none.
func changeFindings(root, specsValue string) []string {
	path, _, _ := strings.Cut(specsValue, " · ")
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	changes := filepath.Join(path, "changes")
	var out []string
	for _, name := range subdirs(changes) {
		if name == "archive" {
			continue
		}
		if !changeKebabPattern.MatchString(name) {
			out = append(out, "Specs: change folder "+quoted(name)+" is not English kebab-case (lowercase letters, digits and hyphens)")
		}
		out = append(out, uncoveredCriteria(name, filepath.Join(changes, name))...)
	}
	for _, name := range subdirs(filepath.Join(changes, "archive")) {
		if !changeArchivePattern.MatchString(name) {
			out = append(out, "Specs: archived change "+quoted(name)+" does not match YYYY-MM-DD-<kebab-case>")
		}
	}
	return out
}

func quoted(name string) string { return `"` + shown(name) + `"` }

// subdirs lists the sorted directory names in dir; an unreadable or missing
// directory has none.
func subdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() && len(names) < maxChangeDirs {
			names = append(names, e.Name())
		}
	}
	return names
}

// uncoveredCriteria warns about the criteria of proposal.md that no Closes line
// of tasks.md names. A folder without both files is skipped.
func uncoveredCriteria(id, dir string) []string {
	proposal, err1 := readLimited(filepath.Join(dir, "proposal.md"))
	tasks, err2 := readLimited(filepath.Join(dir, "tasks.md"))
	if err1 != nil || err2 != nil {
		return nil
	}
	closed := map[int]bool{}
	for _, line := range strings.Split(tasks, "\n") {
		m := closesLinePattern.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		for _, r := range acReferencePattern.FindAllStringSubmatch(m[1], -1) {
			from, _ := strconv.Atoi(r[1])
			to := from
			if r[2] != "" {
				to, _ = strconv.Atoi(r[2])
			}
			for n := from; n <= to && n-from < 1000; n++ {
				closed[n] = true
			}
		}
	}
	var missing []int
	seen := map[int]bool{}
	for _, line := range strings.Split(proposal, "\n") {
		if m := acDefinitionPattern.FindStringSubmatch(line); m != nil {
			n, _ := strconv.Atoi(m[1])
			if !closed[n] && !seen[n] {
				missing = append(missing, n)
			}
			seen[n] = true
		}
	}
	if len(missing) == 0 {
		return nil
	}
	sort.Ints(missing)
	names := make([]string, len(missing))
	for i, n := range missing {
		names[i] = "AC" + strconv.Itoa(n)
	}
	return []string{"Specs: change " + quoted(id) + ": " + strings.Join(names, ", ") + " in proposal.md not named by any **Closes:** line in tasks.md"}
}

// readLimited reads a regular file up to hiveFileLimit bytes.
func readLimited(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", os.ErrInvalid
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(data) > hiveFileLimit {
		data = data[:hiveFileLimit]
	}
	return string(data), nil
}
