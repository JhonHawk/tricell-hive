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

// latestDelta is the most recent archived delta of one requirement. SameDay
// holds the deltas of earlier folders of the same day whose order against it
// is unknown: one of the two has no time, or both have the same time.
// Duplicate is the report for a folder that lists the requirement more than
// once; such a delta takes part in no comparison.
type latestDelta struct {
	Capability string
	Folder     string
	deltaBlock
	SameDay   []deltaBlock
	Duplicate string
}

// archiveTime is the HHMM segment of an archive folder named
// YYYY-MM-DD-HHMM-<change-id>, or "" for a name without a valid time.
func archiveTime(folder string) string {
	if len(folder) < 16 || folder[10] != '-' || folder[15] != '-' {
		return ""
	}
	hhmm := folder[11:15]
	for _, c := range hhmm {
		if c < '0' || c > '9' {
			return ""
		}
	}
	if hhmm[:2] > "23" || hhmm[2:] > "59" {
		return ""
	}
	return hhmm
}

// unorderedSameDay reports whether two archive folders of one day have no
// known order between them.
func unorderedSameDay(a, b string) bool {
	if closingDate(a) != closingDate(b) {
		return false
	}
	ta, tb := archiveTime(a), archiveTime(b)
	return ta == "" || tb == "" || ta == tb
}

// duplicateOps returns, for each requirement the blocks of one delta file list
// more than once, the operations it appears under, in file order.
func duplicateOps(blocks []deltaBlock) map[string][]string {
	ops := map[string][]string{}
	for _, b := range blocks {
		ops[b.Name] = append(ops[b.Name], b.Op)
	}
	for name, o := range ops {
		if len(o) < 2 {
			delete(ops, name)
		}
	}
	return ops
}

// satisfiedBy reports whether a spec block, present or not, agrees with delta b.
func satisfiedBy(b deltaBlock, current string, present bool) bool {
	if b.Op == "REMOVED" {
		return !present
	}
	return present && normalizeBlock(current) == normalizeBlock(b.Block)
}

// closingDate is the date prefix (YYYY-MM-DD) of an archive folder name.
func closingDate(folder string) string {
	if len(folder) < 10 {
		return folder
	}
	return folder[:10]
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
// folder-name order, which follows the closing date and then the time when the
// name carries one, and keeps, per capability and requirement, the delta of the
// last folder. A requirement its folder lists more than once keeps the folder's
// first block with its Duplicate report, and no block of it reaches SameDay.
func collectLatestDeltas(root string) ([]latestDelta, error) {
	paths, err := filepath.Glob(filepath.Join(root, "changes", "archive", "*", "specs", "*", "spec.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths) // the folder name leads with the closing date and time
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
		blocks := parseDeltaBlocks(string(data))
		dups := duplicateOps(blocks)
		seen := map[string]bool{}
		for _, b := range blocks {
			if seen[b.Name] {
				continue
			}
			seen[b.Name] = true
			d := latestDelta{Capability: capability, Folder: folder, deltaBlock: b}
			if ops, ok := dups[b.Name]; ok {
				d.Duplicate = fmt.Sprintf("%s/changes/archive/%s/specs/%s/spec.md: requirement %q listed %d times (%s)", openspecDisplay, folder, capability, b.Name, len(ops), strings.Join(ops, ", "))
			}
			key := capability + "/" + b.Name
			if i, ok := index[key]; ok {
				prev := out[i]
				if unorderedSameDay(prev.Folder, folder) {
					d.SameDay = append([]deltaBlock{}, prev.SameDay...)
					if prev.Duplicate == "" {
						d.SameDay = append(d.SameDay, prev.deltaBlock)
					}
				}
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
// the current spec. It returns one message per mismatch, and one per
// requirement whose latest delta lists it more than once, reported whatever the
// spec or an active change says.
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
		if d.Duplicate != "" {
			failures = append(failures, d.Duplicate)
			continue
		}
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
		matched := satisfiedBy(d.deltaBlock, current, present)
		for _, b := range d.SameDay {
			if !matched && satisfiedBy(b, current, present) {
				matched = true
			}
		}
		if matched {
			continue
		}
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

// Two folders of one day order by their YYYY-MM-DD-HHMM time when both carry
// different times, so the spec must match the later one.
func TestSpecDeltaSameDateOrdersByTime(t *testing.T) {
	files := map[string]string{
		"changes/archive/2026-01-01-1403-a/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nFirst.\n",
		"changes/archive/2026-01-01-2003-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nSecond.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nFirst.\n",
	}
	wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "2026-01-01-2003-b", "DIFFERS")
	files["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\nSecond.\n"
	if got := runSpecDeltaCheck(t, files); len(got) != 0 {
		t.Fatalf("spec matching the later timed delta reported: %q", got)
	}
}

// Two folders of one day where either has no time, or both have the same
// time, do not fix their order, so the spec may match the delta of either one.
func TestSpecDeltaSameDateAcceptsEitherDelta(t *testing.T) {
	// "2575" is not a valid HHMM, so that name has no time.
	for _, names := range [][2]string{{"2026-01-01-1403-a", "2026-01-01-b"}, {"2026-01-01-1200-a", "2026-01-01-1200-b"}, {"2026-01-01-2003-a", "2026-01-01-2575-b"}} {
		files := map[string]string{
			"changes/archive/" + names[0] + "/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nFirst.\n",
			"changes/archive/" + names[1] + "/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nSecond.\n",
			"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nFirst.\n",
		}
		if got := runSpecDeltaCheck(t, files); len(got) != 0 {
			t.Fatalf("%v: spec matching an unordered same-day delta reported: %q", names, got)
		}
	}

	files := map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nFirst.\n",
		"changes/archive/2026-01-01-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nSecond.\n",
	}
	for _, text := range []string{"First.", "Second."} {
		files["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\n" + text + "\n"
		if got := runSpecDeltaCheck(t, files); len(got) != 0 {
			t.Fatalf("spec matching the same-date delta %q reported: %q", text, got)
		}
	}
	files["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\nThird.\n"
	wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "2026-01-01-b", "DIFFERS")
	// An earlier date still loses to a later one.
	files["changes/archive/2025-12-31-z/specs/cap/spec.md"] = "## MODIFIED Requirements\n\n### Requirement: Alpha\nThird.\n"
	wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "2026-01-01-b", "DIFFERS")

	// A same-day REMOVED counts as one of the candidates.
	removed := map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## REMOVED Requirements\n\n### Requirement: Alpha\nRetired.\n",
		"changes/archive/2026-01-01-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nSecond.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Other\nOther.\n",
	}
	if got := runSpecDeltaCheck(t, removed); len(got) != 0 {
		t.Fatalf("spec matching the same-date REMOVED reported: %q", got)
	}
	removed["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\nThird.\n"
	wantOneFailure(t, runSpecDeltaCheck(t, removed), "Alpha", "2026-01-01-b", "DIFFERS")
}

// A requirement one delta lists twice leaves its outcome to the order the
// sections are applied in. The repeat is reported, in either file order and
// whatever the spec or an active change says, while its folder holds the latest
// delta of that requirement; a folder sorted after it supersedes it.
func TestSpecDeltaSameFolderDuplicateFails(t *testing.T) {
	dup := map[string]string{
		"removed-then-added": "## REMOVED Requirements\n\n### Requirement: Alpha\nReplaced.\n\n## ADDED Requirements\n\n### Requirement: Alpha\nNew text.\n",
		"added-then-removed": "## ADDED Requirements\n\n### Requirement: Alpha\nNew text.\n\n## REMOVED Requirements\n\n### Requirement: Alpha\nReplaced.\n",
	}
	for name, delta := range dup {
		t.Run(name, func(t *testing.T) {
			files := map[string]string{
				"changes/archive/2026-01-01-a/specs/cap/spec.md": delta,
				"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nNew text.\n",
			}
			wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "2026-01-01-a", "listed 2 times", "ADDED", "REMOVED")

			files["changes/live/specs/cap/spec.md"] = "## MODIFIED Requirements\n\n### Requirement: Alpha\nLive.\n"
			wantOneFailure(t, runSpecDeltaCheck(t, files), "2026-01-01-a", "listed 2 times")
			delete(files, "changes/live/specs/cap/spec.md")

			// A later folder supersedes the duplicate.
			files["changes/archive/2026-02-01-c/specs/cap/spec.md"] = "## MODIFIED Requirements\n\n### Requirement: Alpha\nNew text.\n"
			if got := runSpecDeltaCheck(t, files); len(got) != 0 {
				t.Fatalf("superseded duplicate reported: %q", got)
			}
			// The later folder is still compared with the spec.
			files["specs/cap/spec.md"] = specHeader + "### Requirement: Alpha\nOther text.\n"
			wantOneFailure(t, runSpecDeltaCheck(t, files), "Alpha", "2026-02-01-c", "DIFFERS")
		})
	}
}

// The blocks of a duplicated requirement take part in no comparison, so they
// cannot hide the mismatch of an unordered same-day folder that supersedes it.
func TestSpecDeltaDuplicateHidesNoMismatch(t *testing.T) {
	got := runSpecDeltaCheck(t, map[string]string{
		"changes/archive/2026-01-01-a/specs/cap/spec.md": "## ADDED Requirements\n\n### Requirement: Alpha\nX.\n\n## MODIFIED Requirements\n\n### Requirement: Alpha\nY.\n",
		"changes/archive/2026-01-01-b/specs/cap/spec.md": "## MODIFIED Requirements\n\n### Requirement: Alpha\nZ.\n",
		"specs/cap/spec.md": specHeader + "### Requirement: Alpha\nX.\n",
	})
	wantOneFailure(t, got, "Alpha", "2026-01-01-b", "DIFFERS")
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
