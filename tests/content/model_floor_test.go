package content

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The support floor document lists the models distributed guidance is written
// for and the candidates whose observed failures count toward a correction
// (root AGENTS.md «Guidance structure» and «Measurement»). These tests keep its
// tables aligned with the role profiles and the flash-class pilot list, so a
// profile change or a dropped row cannot leave the floor stale unnoticed.

const (
	modelFloorPath    = "../../_support/docs/harness-engineering/model-floor.md"
	agentProfilesPath = "../../integrations/agent-profiles.json"
	flashClassPath    = "../../tests/pilot/flash_class.go"
)

// writingTargetProfiles are the role profiles whose named models form the
// writing target.
var writingTargetProfiles = []string{"execution", "focused", "reasoning", "verifier"}

var floorStatuses = map[string]bool{"provisional": true, "qualified": true}

type floorRow struct {
	host, model, origin, target, status string
	efforts                             map[string]bool
}

type expectedModel struct {
	host, model, effort string
	target              bool
}

func TestModelFloor(t *testing.T) {
	doc := readFile(t, modelFloorPath)
	expected := append(profileModels(t, readFile(t, agentProfilesPath)), flashModels(t, readFile(t, flashClassPath))...)
	for _, err := range checkFloor(doc, expected) {
		t.Error(err)
	}
}

func TestModelFloorDetectsDrift(t *testing.T) {
	doc := readFile(t, modelFloorPath)
	expected := append(profileModels(t, readFile(t, agentProfilesPath)), flashModels(t, readFile(t, flashClassPath))...)
	cases := map[string]func(string) string{
		"row removed": func(d string) string {
			return dropLine(d, "| codex | `gpt-6.1-terra` |")
		},
		"effort changed": func(d string) string {
			return strings.Replace(d, "| claude | `sonnet` | `high` |", "| claude | `sonnet` | `low` |", 1)
		},
		"alias without ID": func(d string) string {
			return regexp.MustCompile("(?m)^\\| `sonnet` \\| [^|]+\\|").ReplaceAllString(d, "| `sonnet` | |")
		},
	}
	for name, mutate := range cases {
		mutated := mutate(doc)
		if mutated == doc {
			t.Fatalf("%s: the mutation did not change the document; update the test to the table's current shape", name)
		}
		if len(checkFloor(mutated, expected)) == 0 {
			t.Errorf("%s: checkFloor accepted a document it should reject", name)
		}
	}
}

// checkFloor returns one error per row the floor table or the alias table
// lacks or states wrongly.
func checkFloor(doc string, expected []expectedModel) []error {
	var errs []error
	rows := parseTable(section(doc, "## Floor"))
	if len(rows) == 0 {
		return []error{fmt.Errorf("model-floor.md has no table under «## Floor»")}
	}
	floor := map[string]floorRow{}
	for _, cells := range rows {
		if len(cells) < 8 {
			errs = append(errs, fmt.Errorf("floor row %q has %d cells, want 8", strings.Join(cells, " | "), len(cells)))
			continue
		}
		r := floorRow{host: cells[0], model: unquote(cells[1]), origin: cells[3], target: cells[4], status: cells[5], efforts: map[string]bool{}}
		for _, e := range strings.Split(cells[2], ",") {
			if e = unquote(strings.TrimSpace(e)); e != "" && e != "—" {
				r.efforts[e] = true
			}
		}
		if !floorStatuses[r.status] {
			errs = append(errs, fmt.Errorf("floor row %s %s: status %q, want provisional or qualified", r.host, r.model, r.status))
		}
		floor[r.host+" "+r.model] = r
	}
	for _, want := range expected {
		r, ok := floor[want.host+" "+want.model]
		if !ok {
			errs = append(errs, fmt.Errorf("floor table lacks %s %s", want.host, want.model))
			continue
		}
		if want.effort != "" && !r.efforts[want.effort] {
			errs = append(errs, fmt.Errorf("floor row %s %s lacks effort %s", want.host, want.model, want.effort))
		}
		if want.target && r.target != "yes" {
			errs = append(errs, fmt.Errorf("floor row %s %s: writing target %q, want yes", want.host, want.model, r.target))
		}
	}
	aliases := parseTable(section(doc, "## Aliases"))
	if len(aliases) == 0 {
		errs = append(errs, fmt.Errorf("model-floor.md has no table under «## Aliases»"))
	}
	seen := map[string]bool{}
	for _, cells := range aliases {
		if len(cells) < 4 {
			errs = append(errs, fmt.Errorf("alias row %q has %d cells, want 4", strings.Join(cells, " | "), len(cells)))
			continue
		}
		alias := unquote(cells[0])
		seen[alias] = true
		if !strings.Contains(cells[1], "`") {
			errs = append(errs, fmt.Errorf("alias %s names no full model ID", alias))
		}
		if !regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).MatchString(cells[3]) {
			errs = append(errs, fmt.Errorf("alias %s has no verification date", alias))
		}
	}
	for _, r := range floor {
		if r.host == "claude" && !strings.Contains(r.model, "-") && !seen[r.model] {
			errs = append(errs, fmt.Errorf("floor alias %s has no row under «## Aliases»", r.model))
		}
	}
	return errs
}

func profileModels(t *testing.T, raw string) []expectedModel {
	t.Helper()
	var profiles struct {
		Hosts map[string]struct {
			Models map[string]struct {
				Model  string `json:"model"`
				Effort string `json:"effort"`
			} `json:"models"`
		} `json:"hosts"`
	}
	if err := json.Unmarshal([]byte(raw), &profiles); err != nil {
		t.Fatal(err)
	}
	var out []expectedModel
	for host, h := range profiles.Hosts {
		for _, name := range writingTargetProfiles {
			p := h.Models[name]
			if p.Model == "" || p.Model == "inherit" {
				continue
			}
			out = append(out, expectedModel{host: host, model: p.Model, effort: p.Effort, target: true})
		}
	}
	if len(out) == 0 {
		t.Fatal("agent-profiles.json names no model in the writing-target profiles")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].host+out[i].model < out[j].host+out[j].model })
	return out
}

// flashModels reads the flash-class list from the pilot source, since it is an
// unexported variable of another package.
func flashModels(t *testing.T, src string) []expectedModel {
	t.Helper()
	m := regexp.MustCompile(`var flashModels = \[\]string\{([^}]*)\}`).FindStringSubmatch(src)
	if m == nil {
		t.Fatal("flash_class.go has no flashModels list")
	}
	var out []expectedModel
	for _, q := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(m[1], -1) {
		model, effort, _ := strings.Cut(q[1], "#")
		out = append(out, expectedModel{host: "opencode", model: model, effort: effort})
	}
	if len(out) == 0 {
		t.Fatal("flashModels is empty")
	}
	return out
}

// section returns the text from heading to the next heading of the same level.
func section(doc, heading string) string {
	i := strings.Index(doc, "\n"+heading+"\n")
	if i < 0 {
		return ""
	}
	rest := doc[i+len(heading)+2:]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// parseTable returns the body rows of the first Markdown table in text.
func parseTable(text string) [][]string {
	var rows [][]string
	inTable := false
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			if inTable {
				break
			}
			continue
		}
		if !inTable {
			inTable = true
			continue // header
		}
		if strings.HasPrefix(line, "| ---") || strings.HasPrefix(line, "|---") {
			continue
		}
		var cells []string
		for _, c := range strings.Split(strings.Trim(line, "|"), "|") {
			cells = append(cells, strings.TrimSpace(c))
		}
		rows = append(rows, cells)
	}
	return rows
}

func unquote(s string) string { return strings.Trim(strings.TrimSpace(s), "`") }

func dropLine(doc, prefix string) string {
	var kept []string
	for _, line := range strings.Split(doc, "\n") {
		if !strings.HasPrefix(line, prefix) {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
