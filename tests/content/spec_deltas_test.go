package content

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// openspecRoot is the repository's OpenSpec home, relative to this package.
const openspecRoot = "../../_support/openspec"

// openspecDisplay prefixes paths in failure messages so they read as
// repository-relative.
const openspecDisplay = "_support/openspec"

// deltaBlock is one requirement block of a delta file: the operation of the
// section it sits in, the requirement name, and its text from the
// "### Requirement:" heading up to the next heading of that level or higher.
type deltaBlock struct {
	Op    string
	Name  string
	Block string
}

// latestDelta is the most recent archived delta of one requirement.
type latestDelta struct {
	Capability string
	Folder     string
	deltaBlock
}

// parseDeltaBlocks returns the requirement blocks of a delta file. Text before
// the first "## ADDED|MODIFIED|REMOVED Requirements" heading is a preamble and
// yields no block.
func parseDeltaBlocks(text string) []deltaBlock {
	var blocks []deltaBlock
	op := ""
	cur := -1
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "## ") {
			cur = -1
			op = ""
			for _, o := range []string{"ADDED", "MODIFIED", "REMOVED"} {
				if strings.TrimSpace(line) == "## "+o+" Requirements" {
					op = o
				}
			}
			continue
		}
		if name, ok := requirementName(line); ok {
			cur = -1
			if op != "" {
				blocks = append(blocks, deltaBlock{Op: op, Name: name, Block: line})
				cur = len(blocks) - 1
			}
			continue
		}
		if strings.HasPrefix(line, "### ") {
			cur = -1
			continue
		}
		if cur >= 0 {
			blocks[cur].Block += "\n" + line
		}
	}
	return blocks
}

// requirementName reports the name of a "### Requirement: <name>" heading.
func requirementName(line string) (string, bool) {
	const prefix = "### Requirement:"
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	return strings.TrimSpace(line[len(prefix):]), true
}

// parseSpecBlocks returns the requirement blocks of a spec file by name.
func parseSpecBlocks(text string) map[string]string {
	out := map[string]string{}
	cur := ""
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			cur = ""
		}
		if name, ok := requirementName(line); ok {
			cur = name
			out[cur] = line
			continue
		}
		if cur != "" {
			out[cur] += "\n" + line
		}
	}
	return out
}

// normalizeBlock collapses every run of whitespace to one space.
func normalizeBlock(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// collectLatestDeltas walks changes/archive/*/specs/*/spec.md under root in
// folder-name order and keeps, per capability and requirement, the last delta.
func collectLatestDeltas(root string) ([]latestDelta, error) {
	paths, err := filepath.Glob(filepath.Join(root, "changes", "archive", "*", "specs", "*", "spec.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths) // the folder name leads with the closing date
	index := map[string]int{}
	var out []latestDelta
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		capDir := filepath.Dir(path)
		capability := filepath.Base(capDir)
		folder := filepath.Base(filepath.Dir(filepath.Dir(capDir)))
		for _, b := range parseDeltaBlocks(string(data)) {
			d := latestDelta{Capability: capability, Folder: folder, deltaBlock: b}
			key := capability + "/" + b.Name
			if i, ok := index[key]; ok {
				out[i] = d
				continue
			}
			index[key] = len(out)
			out = append(out, d)
		}
	}
	return out, nil
}

// activeChangeRequirements returns the "capability/requirement" keys touched by
// changes/<id>/specs/ outside archive/.
func activeChangeRequirements(root string) (map[string]bool, error) {
	paths, err := filepath.Glob(filepath.Join(root, "changes", "*", "specs", "*", "spec.md"))
	if err != nil {
		return nil, err
	}
	keys := map[string]bool{}
	for _, path := range paths {
		capDir := filepath.Dir(path)
		if filepath.Base(filepath.Dir(filepath.Dir(capDir))) == "archive" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, b := range parseDeltaBlocks(string(data)) {
			keys[filepath.Base(capDir)+"/"+b.Name] = true
		}
	}
	return keys, nil
}

// checkSpecDeltas compares the latest archived delta of every requirement with
// the current spec and returns one message per mismatch.
func checkSpecDeltas(root string) ([]string, error) {
	deltas, err := collectLatestDeltas(root)
	if err != nil {
		return nil, err
	}
	active, err := activeChangeRequirements(root)
	if err != nil {
		return nil, err
	}
	specs := map[string]map[string]string{}
	var failures []string
	for _, d := range deltas {
		if active[d.Capability+"/"+d.Name] {
			continue
		}
		blocks, ok := specs[d.Capability]
		if !ok {
			blocks = map[string]string{}
			data, err := os.ReadFile(filepath.Join(root, "specs", d.Capability, "spec.md"))
			if err == nil {
				blocks = parseSpecBlocks(string(data))
			} else if !os.IsNotExist(err) {
				return nil, err
			}
			specs[d.Capability] = blocks
		}
		where := fmt.Sprintf("%s/specs/%s/spec.md: requirement %q (%s in %s/changes/archive/%s)", openspecDisplay, d.Capability, d.Name, d.Op, openspecDisplay, d.Folder)
		current, present := blocks[d.Name]
		switch {
		case d.Op == "REMOVED" && present:
			failures = append(failures, where+": still present")
		case d.Op == "REMOVED":
		case !present:
			failures = append(failures, where+": MISSING")
		case normalizeBlock(current) != normalizeBlock(d.Block):
			failures = append(failures, where+": DIFFERS from the delta")
		}
	}
	return failures, nil
}

func writeTreeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runSpecDeltaCheck(t *testing.T, files map[string]string) []string {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		writeTreeFile(t, root, rel, content)
	}
	got, err := checkSpecDeltas(root)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func wantOneFailure(t *testing.T, got []string, parts ...string) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("got %d failures %q, want exactly 1", len(got), got)
	}
	for _, p := range parts {
		if !strings.Contains(got[0], p) {
			t.Errorf("failure %q does not contain %q", got[0], p)
		}
	}
}

const specHeader = "# cap\n\n## Purpose\n\nText.\n\n## Requirements\n\n"

func TestSpecDeltaAddedMissing(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha SHALL exist.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Other\nOther.\n",
	})
	wantOneFailure(t, got, "cap", "Alpha", "ADDED", "2026-01-01-a", "MISSING")
}

func TestSpecDeltaSpecFileAbsent(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha SHALL exist.\n",
	})
	wantOneFailure(t, got, "cap", "Alpha", "MISSING")
}

func TestSpecDeltaTextDiffers(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nAlpha SHALL exist now.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha SHALL exist.\n",
	})
	wantOneFailure(t, got, "cap", "Alpha", "MODIFIED", "2026-01-01-a", "DIFFERS")
}

func TestSpecDeltaWhitespaceOnlyDifferencePasses(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha  SHALL\nexist.\n\n#### Scenario: S\n- **WHEN** x\n- **THEN** y\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha SHALL exist.\n\n\n#### Scenario: S\n- **WHEN**   x\n- **THEN** y\n\n",
	})
	if len(got) != 0 {
		t.Fatalf("whitespace-only difference reported: %q", got)
	}
}

func TestSpecDeltaRemovedHeadingStillPresent(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## REMOVED Requirements\n\n### Requirement: Alpha\nRetired because it is obsolete.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha SHALL exist.\n",
	})
	wantOneFailure(t, got, "cap", "Alpha", "REMOVED", "2026-01-01-a", "still present")
}

func TestSpecDeltaRemovedHeadingAbsentPasses(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## REMOVED Requirements\n\n### Requirement: Alpha\nRetired because it is obsolete.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Other\nOther.\n",
	})
	if len(got) != 0 {
		t.Fatalf("absent REMOVED requirement reported: %q", got)
	}
}

func TestSpecDeltaRemovedWithoutSpecFilePasses(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## REMOVED Requirements\n\n### Requirement: Alpha\nRetired.\n",
	})
	if len(got) != 0 {
		t.Fatalf("REMOVED with no spec file reported: %q", got)
	}
}

func TestSpecDeltaLaterModifiedReplacesAdded(t *testing.T) {
	files := map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nVersion one.\n",
		"changes/archive/2026-02-01-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nVersion two.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nVersion two.\n",
	}
	if got := runSpecDeltaCheck(t, files); len(got) != 0 {
		t.Fatalf("spec matching the latest delta reported: %q", got)
	}
	files["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\nVersion one.\n"
	wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "MODIFIED", "2026-02-01-b", "DIFFERS")
}

func TestSpecDeltaSameDateOrdersByFolderName(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nFirst.\n",
		"changes/archive/2026-01-01-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nSecond.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nSecond.\n",
	})
	if len(got) != 0 {
		t.Fatalf("later folder name did not win: %q", got)
	}
}

func TestSpecDeltaActiveChangeRequirementIsSkipped(t *testing.T) {
	files := map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n\n### Requirement: Beta\nBeta.\n",
		"changes/live-change/specs/cap/spec.md":          "## MODIFIED Requirements\n\n### Requirement: Alpha\nAlpha changed by the live change.\n",
		"specs/cap/spec.md":                              specHeader + "### Requirement: Beta\nBeta.\n",
	}
	if got := runSpecDeltaCheck(t, files); len(got) != 0 {
		t.Fatalf("requirement touched by an active change reported: %q", got)
	}
	// Only the touched requirement is skipped.
	files["specs/cap/spec.md"] = specHeader + "### Requirement: Other\nOther.\n"
	wantOneFailure(t, runSpecDeltaCheck(t, files), "Beta", "MISSING")
}

func TestSpecDeltaActiveChangeIsPerCapability(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/one/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n",
		"changes/live-change/specs/two/spec.md":          "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha elsewhere.\n",
	})
	wantOneFailure(t, got, "one", "Alpha", "MISSING")
}

func TestSpecDeltaArchiveFolderIsNotAnActiveChange(t *testing.T) {
	// Only changes/<id>/ outside archive/ is active. Here the sole delta is
	// archived, so nothing may be skipped.
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n",
	})
	wantOneFailure(t, got, "Alpha", "MISSING")
}

func TestSpecDeltaUnappliedFolderIsNotRead(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/unapplied-specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n",
	})
	if len(got) != 0 {
		t.Fatalf("unapplied-specs delta read as applied: %q", got)
	}
}

func TestSpecDeltaPreambleIsIgnored(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "# cap\n\n### Requirement: Ghost\nBefore any operation heading.\n\n## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha.\n",
	})
	if len(got) != 0 {
		t.Fatalf("preamble read as a requirement: %q", got)
	}
}

func TestSpecDeltaLastRequirementEndsAtEndOfFile(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n\n### Requirement: Omega\nOmega tail text.",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha.\n\n### Requirement: Omega\nOmega tail text, changed.",
	})
	wantOneFailure(t, got, "Omega", "DIFFERS")
}

func TestSpecDeltaBlockEndsAtNextSection(t *testing.T) {
	// The spec's requirement is followed by another "##" section; its text
	// must not absorb that section.
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nAlpha.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nAlpha.\n\n## Notes\n\nTrailing section.\n",
	})
	if len(got) != 0 {
		t.Fatalf("trailing section absorbed into the requirement: %q", got)
	}
}

func TestSpecDeltaParseBlocks(t *testing.T) {
	blocks := parseDeltaBlocks("preamble\n\n## ADDED Requirements\n\n### Requirement: A\nbody a\n\n## REMOVED Requirements\n\n### Requirement: B\nwhy\n")
	if len(blocks) != 2 {
		t.Fatalf("got %d blocks, want 2: %+v", len(blocks), blocks)
	}
	if blocks[0].Op != "ADDED" || blocks[0].Name != "A" || !strings.Contains(blocks[0].Block, "body a") {
		t.Errorf("first block = %+v", blocks[0])
	}
	if blocks[1].Op != "REMOVED" || blocks[1].Name != "B" || strings.Contains(blocks[1].Block, "body a") {
		t.Errorf("second block = %+v", blocks[1])
	}
}

// TestSpecDeltasApplied checks the repository: the latest archived delta of
// every requirement must match the current spec, so a closed change that
// marked its delta applied without applying it fails here.
func TestSpecDeltasApplied(t *testing.T) {
	got, err := checkSpecDeltas(openspecRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, msg := range got {
		t.Error(msg)
	}
}
