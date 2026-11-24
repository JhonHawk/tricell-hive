package management

import (
	"fmt"
	"sort"
	"strings"

	"tricell-hive/integrations/antigravity"
	"tricell-hive/integrations/mdlinks"
	"tricell-hive/integrations/target"
)

// Antigravity truncates every rule file above antigravity.MaxRuleBytes, so the
// global guidance reaches it as numbered rule parts, hive-01.md, hive-02.md…
// Each part is an ordinary block record of GlobalSource whose file name carries
// its 1-based part number: no record field is added, so a state written before
// this host existed loads unchanged. Every part holds a whole number of `##`
// sections of the rewritten guidance, and the voice block lives in part 1 only.

// antigravityVoiceReserve is the room part 1 keeps for the voice block. The
// largest block the shipped voices can render (longest voice, a 40-letter name
// of 4-byte characters, marked intensity) is checked against it by a test; a
// voice that outgrows it fails the plan instead of overflowing the file.
const antigravityVoiceReserve = 2500

// antigravityPartSize is the size of a rule file that holds body as its Hive
// block: the frontmatter Hive writes plus the managed block.
func antigravityPartSize(body string) int {
	return len(antigravity.RuleFrontmatter) + len(managedBlock([]byte(body), nil, hiveMarkers))
}

type guidanceUnit struct {
	heading string
	text    string
}

// splitGuidanceUnits cuts text into the text before the first `##` heading and
// one unit per `##` section. Only headings outside fenced code blocks count.
func splitGuidanceUnits(text string) []guidanceUnit {
	var units []guidanceUnit
	cur := guidanceUnit{heading: "the text before the first ## heading"}
	var fenceChar byte
	fenceLen := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if line == "" {
			continue
		}
		if fenceChar != 0 {
			if closesFence(line, fenceChar, fenceLen) {
				fenceChar, fenceLen = 0, 0
			}
		} else if ch, n, ok := opensFence(line); ok {
			fenceChar, fenceLen = ch, n
		} else if strings.HasPrefix(line, "## ") {
			if cur.text != "" {
				units = append(units, cur)
			}
			cur = guidanceUnit{heading: strings.TrimSpace(line)}
		}
		cur.text += line
	}
	if cur.text != "" {
		units = append(units, cur)
	}
	return units
}

func fenceRun(line string, ch byte) (indent, n int, rest string) {
	for indent < len(line) && indent < 4 && line[indent] == ' ' {
		indent++
	}
	for indent+n < len(line) && line[indent+n] == ch {
		n++
	}
	return indent, n, line[indent+n:]
}

func opensFence(line string) (byte, int, bool) {
	for _, ch := range []byte{'`', '~'} {
		indent, n, rest := fenceRun(line, ch)
		if indent > 3 || n < 3 {
			continue
		}
		if ch == '`' && strings.Contains(rest, "`") {
			continue
		}
		return ch, n, true
	}
	return 0, 0, false
}

func closesFence(line string, ch byte, open int) bool {
	indent, n, rest := fenceRun(line, ch)
	return indent <= 3 && n >= open && strings.TrimSpace(rest) == ""
}

// splitAntigravityGuidance cuts the rewritten guidance into rule parts.
// Units are packed greedily, in order: a part takes the next section while the
// file it would produce stays within antigravity.RuleBudget (part 1 also keeps
// antigravityVoiceReserve). The concatenation of the parts is text itself. A
// section that cannot start a part within the budget fails, named.
func splitAntigravityGuidance(text string) ([]string, error) {
	var parts []string
	var cur string
	limit := func(part int) int {
		if part == 1 {
			return antigravity.RuleBudget - antigravityVoiceReserve
		}
		return antigravity.RuleBudget
	}
	for _, u := range splitGuidanceUnits(text) {
		if cur != "" && antigravityPartSize(cur+u.text) <= limit(len(parts)+1) {
			cur += u.text
			continue
		}
		if cur != "" {
			parts = append(parts, cur)
		}
		if size := antigravityPartSize(u.text); size > limit(len(parts)+1) {
			return nil, fmt.Errorf("Antigravity rule part %d: section %q is %d bytes with the frontmatter and block markers, over the %d-byte limit; split it with a ## heading", len(parts)+1, u.heading, size, limit(len(parts)+1))
		}
		cur = u.text
	}
	if cur != "" {
		parts = append(parts, cur)
	}
	if len(parts) > antigravity.MaxRuleParts {
		return nil, fmt.Errorf("Antigravity rule parts: %d parts exceed the %d the file names allow", len(parts), antigravity.MaxRuleParts)
	}
	return parts, nil
}

// antigravityGuidanceParts returns the rule parts of release r as Antigravity
// receives them, with the skill links rewritten for its shared store.
func antigravityGuidanceParts(c target.Config, r *Release) ([]string, error) {
	text := mdlinks.RewriteSkillLinks(string(payload(r, GlobalSource)), skillsDirFor("antigravity", c))
	return splitAntigravityGuidance(text)
}

// antigravityRuleTargets lists the rule part targets of a plan. An install
// takes the count from the release it writes; any other caller takes the parts
// state records for this home. Surplus parts of an install are not listed:
// desiredResources retires them as absent entries.
func antigravityRuleTargets(c target.Config, hosts []string, state State, rel *Release) ([]target.Target, error) {
	if c.Scope != "user" {
		return nil, nil
	}
	selected := false
	for _, h := range hosts {
		selected = selected || h == "antigravity"
	}
	if !selected {
		return nil, nil
	}
	var numbers []int
	if rel != nil {
		parts, err := antigravityGuidanceParts(c, rel)
		if err != nil {
			return nil, err
		}
		for i := range parts {
			numbers = append(numbers, i+1)
		}
	} else {
		want := Consumer{Host: "antigravity", Scope: "user", Context: c.Home}
		for _, r := range state.Records {
			if n, ok := antigravity.RulePartNumber(c.AntigravityHome, r.Target.Path); ok && r.Target.Kind == "block" && hasConsumer(r.Consumers, want) {
				numbers = append(numbers, n)
			}
		}
		sort.Ints(numbers)
	}
	var out []target.Target
	for _, n := range numbers {
		out = append(out, target.Target{Source: GlobalSource, Path: antigravity.RulePath(c.AntigravityHome, n), Kind: "block", Host: "antigravity", Scope: c.Scope, Context: c.Home})
	}
	return out, nil
}

// antigravityRulePart returns the text of the rule part a block target holds,
// and false for any other block.
func antigravityRulePart(c target.Config, r *Release, t target.Target) (string, bool, error) {
	n, ok := antigravity.RulePartNumber(c.AntigravityHome, t.Path)
	if !ok || t.Kind != "block" || t.Source != GlobalSource {
		return "", false, nil
	}
	parts, err := antigravityGuidanceParts(c, r)
	if err != nil {
		return "", true, err
	}
	if n > len(parts) {
		return "", true, fmt.Errorf("Antigravity rule part %d does not exist: the guidance has %d", n, len(parts))
	}
	return parts[n-1], true, nil
}

// skipsVoice reports whether path is an Antigravity rule part after the
// first, which never carries the voice block.
func skipsVoice(c target.Config, path string) bool {
	n, ok := antigravity.RulePartNumber(c.AntigravityHome, path)
	return ok && n > 1
}

// checkAntigravityVoiceFits fails when the voice block would push rule part 1,
// frontmatter and Hive block included, over the size Antigravity truncates at.
func checkAntigravityVoiceFits(c target.Config, path string, hive Record, voice []byte) error {
	if n, ok := antigravity.RulePartNumber(c.AntigravityHome, path); !ok || n != 1 {
		return nil
	}
	if size := len(hive.Leading) + len(hive.Managed) + len(voice); size > antigravity.MaxRuleBytes {
		return fmt.Errorf("%s: the voice block makes the rule file %d bytes, over Antigravity's %d-byte limit", path, size, antigravity.MaxRuleBytes)
	}
	return nil
}
