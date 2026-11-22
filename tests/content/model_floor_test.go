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
// table aligned with the role profiles, so a profile change or a dropped row
// cannot leave the floor stale unnoticed, and its writing-target rows aligned
// with the models «Two sets» names.

const (
	modelFloorPath    = "../../_support/docs/harness-engineering/model-floor.md"
	agentProfilesPath = "../../integrations/agent-profiles.json"
)

// floorProfiles are the role profiles whose named models need a floor row.
var floorProfiles = []string{"execution", "focused", "reasoning", "verifier"}

var floorStatuses = map[string]bool{"provisional": true, "qualified": true}

// fullModelID matches a backquoted versioned model ID, such as `claude-opus-5-5`.
var fullModelID = regexp.MustCompile("`[a-z0-9.]+(-[a-z0-9.]+)+`")

type floorRow struct {
	host, model, origin, target, status string
	efforts                             map[string]bool
}

type expectedModel struct {
	host, model, effort string
}

func TestModelFloor(t *testing.T) {
	doc := readFloorFile(t, modelFloorPath)
	expected := profileModels(t, readFloorFile(t, agentProfilesPath))
	for _, err := range checkFloor(doc, expected) {
		t.Error(err)
	}
}

func TestModelFloorDetectsDrift(t *testing.T) {
	doc := readFloorFile(t, modelFloorPath)
	expected := profileModels(t, readFloorFile(t, agentProfilesPath))
	cases := map[string]func(string) string{
		"row removed": func(d string) string {
			return dropFloorLine(d, "| codex | `gpt-6.1-terra` |")
		},
		"effort changed": func(d string) string {
			return strings.Replace(d, "| claude | `sonnet` | `high` |", "| claude | `sonnet` | `low` |", 1)
		},
		"stale writing-target row": func(d string) string {
			return strings.Replace(d, "| codex | `gpt-6.1-terra` |", "| codex | `gpt-6.1-terra` |\n| codex | `gpt-5.9-old` | `medium` | `focused` | yes | provisional | none | — |\n| codex | `gpt-6.1-terra-x` |", 1)
		},
		"target row «Two sets» does not list": func(d string) string {
			return strings.Replace(d, "`focused`; tier below | no |", "`focused`; tier below | yes |", 1)
		},
		"listed target not marked": func(d string) string {
			return strings.Replace(d, "| claude | `opus` | `high` | `reasoning`, `verifier` | yes |", "| claude | `opus` | `high` | `reasoning`, `verifier` | no |", 1)
		},
		"listed target without a row": func(d string) string {
			return strings.Replace(d, "- **Writing target.** `opus` `gpt-6.2-sol`", "- **Writing target.** `opus` `gpt-6.2-sol` `gpt-7-nova`", 1)
		},
		"alias with a placeholder ID": func(d string) string {
			return regexp.MustCompile("(?m)^\\| `sonnet` \\| [^|]+\\|").ReplaceAllString(d, "| `sonnet` | `tbd` |")
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
	rows := floorTable(floorSection(doc, "## Floor"))
	if len(rows) == 0 {
		return []error{fmt.Errorf("model-floor.md has no table under «## Floor»")}
	}
	floor := map[string]floorRow{}
	for _, cells := range rows {
		if len(cells) < 8 {
			errs = append(errs, fmt.Errorf("floor row %q has %d cells, want 8", strings.Join(cells, " | "), len(cells)))
			continue
		}
		r := floorRow{host: cells[0], model: floorUnquote(cells[1]), origin: cells[3], target: cells[4], status: cells[5], efforts: map[string]bool{}}
		for _, e := range strings.Split(cells[2], ",") {
			if e = floorUnquote(strings.TrimSpace(e)); e != "" && e != "—" {
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
	}
	// The rows marked as writing target are exactly the models the «Writing
	// target» bullet of «Two sets» lists, so a stale or wrongly promoted row
	// fails, and so does a listed model without its row.
	targets := writingTargets(floorSection(doc, "## Two sets"))
	if len(targets) == 0 {
		errs = append(errs, fmt.Errorf("«Two sets» lists no writing-target model"))
	}
	marked := map[string]bool{}
	for key, r := range floor {
		switch r.target {
		case "yes":
			marked[r.model] = true
			if !targets[r.model] {
				errs = append(errs, fmt.Errorf("floor row %s marks a writing target that «Two sets» does not list", key))
			}
		case "no":
		default:
			errs = append(errs, fmt.Errorf("floor row %s: writing target %q, want yes or no", key, r.target))
		}
	}
	for model := range targets {
		if !marked[model] {
			errs = append(errs, fmt.Errorf("«Two sets» lists %s as a writing target, but no floor row marks it", model))
		}
	}
	aliases := floorTable(floorSection(doc, "## Aliases"))
	if len(aliases) == 0 {
		errs = append(errs, fmt.Errorf("model-floor.md has no table under «## Aliases»"))
	}
	seen := map[string]bool{}
	for _, cells := range aliases {
		if len(cells) < 4 {
			errs = append(errs, fmt.Errorf("alias row %q has %d cells, want 4", strings.Join(cells, " | "), len(cells)))
			continue
		}
		alias := floorUnquote(cells[0])
		seen[alias] = true
		if !fullModelID.MatchString(cells[1]) {
			errs = append(errs, fmt.Errorf("alias %s names no full model ID", alias))
		}
		if !regexp.MustCompile(`\d{4}-\d{2}-\d{2}`).MatchString(cells[3]) {
			errs = append(errs, fmt.Errorf("alias %s has no verification date", alias))
		}
	}
	// A model name without a hyphen or a provider path is an alias, such as
	// opus or sonnet, and needs its resolution recorded.
	for _, r := range floor {
		if !strings.ContainsAny(r.model, "-/") && !seen[r.model] {
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
		for _, name := range floorProfiles {
			p := h.Models[name]
			if p.Model == "" || p.Model == "inherit" {
				continue
			}
			out = append(out, expectedModel{host: host, model: p.Model, effort: p.Effort})
		}
	}
	if len(out) == 0 {
		t.Fatal("agent-profiles.json names no model in the floor profiles")
	}
	sort.Slice(out, func(i, j int) bool { return out[i].host+out[i].model < out[j].host+out[j].model })
	return out
}

// writingTargets returns the backquoted models of the «Writing target» bullet.
func writingTargets(section string) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(section, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- **Writing target.**") {
			for _, m := range regexp.MustCompile("`([^`]+)`").FindAllStringSubmatch(line, -1) {
				out[m[1]] = true
			}
		}
	}
	return out
}

// floorSection returns the text from heading to the next heading of the same level.
func floorSection(doc, heading string) string {
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

// floorTable returns the body rows of the first Markdown table in text.
func floorTable(text string) [][]string {
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

func floorUnquote(s string) string { return strings.Trim(strings.TrimSpace(s), "`") }

func dropFloorLine(doc, prefix string) string {
	var kept []string
	for _, line := range strings.Split(doc, "\n") {
		if !strings.HasPrefix(line, prefix) {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

func readFloorFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
