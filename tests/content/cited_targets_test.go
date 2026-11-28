package content

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// citedLink matches a Markdown link to a Markdown file, relative or through
// the skill: locator, followed by the section or step the text cites next to
// it: «Section» right after the link, or "step N" in the link text or right
// after it.
var citedLink = regexp.MustCompile(`\[([^\]]*)\]\(((?:skill:)?[^)#\s]+\.md)(#[^)\s]*)?\)(\s*«([^»]+)»|\s+steps?\s+(\d+))?`)

var citedBefore = regexp.MustCompile(`«([^»]+)»\s+of\s+(?:the\s+)?\[[^\]]*\]\(((?:skill:)?[^)#\s]+\.md)(?:#[^)\s]*)?\)`)

var stepInLinkText = regexp.MustCompile(`\bsteps?\s+(\d+)\b`)

// TestCitedSectionsAndStepsExist fails when guidance cites a section or a
// step of a file it links and that file has no such section or step. A
// section exists as a heading with that text, or as a bold label at the start
// of an item or paragraph followed by ".", ":" or "," (inside or outside the
// bold); a step exists as a numbered item "N." at the start of a line in the
// linked file. A section is cited as «Section» right after the link or as
// «Section» of [link](file.md). It checks only that the target
// exists, not that it says what the citation attributes to it.
func TestCitedSectionsAndStepsExist(t *testing.T) {
	root := "../../content"
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		rel, _ := filepath.Rel(root, path)
		check := func(at int, targetLink, section, step string) {
			target := resolveCitedTarget(path, targetLink)
			line := strings.Count(text[:at], "\n") + 1
			body, err := os.ReadFile(target)
			if err != nil {
				return // a missing file is skill_links_test's and the installer's finding
			}
			if section != "" && !hasCitedSection(string(body), section) {
				t.Errorf("content/%s:%d cites «%s» of %s, which has no heading or bold label with that name: link the file that holds it, with its anchor, or rename the citation", rel, line, section, target)
			}
			if step != "" && !hasCitedStep(string(body), step) {
				t.Errorf("content/%s:%d cites step %s of %s, which has no numbered item %s. at the start of a line: cite the step that exists, or link the file that holds the rule", rel, line, step, target, step)
			}
		}
		for _, m := range citedLink.FindAllStringSubmatchIndex(text, -1) {
			linkText := text[m[2]:m[3]]
			var section, step string
			if m[10] >= 0 {
				section = strings.TrimSpace(text[m[10]:m[11]])
			}
			if m[12] >= 0 {
				step = text[m[12]:m[13]]
			} else if s := stepInLinkText.FindStringSubmatch(linkText); s != nil {
				step = s[1]
			}
			if section != "" || step != "" {
				check(m[0], text[m[4]:m[5]], section, step)
			}
		}
		for _, m := range citedBefore.FindAllStringSubmatchIndex(text, -1) {
			check(m[0], text[m[4]:m[5]], strings.TrimSpace(text[m[2]:m[3]]), "")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func resolveCitedTarget(from, link string) string {
	if strings.HasPrefix(link, "skill:") {
		return filepath.Join("../../content/skills", filepath.FromSlash(strings.TrimPrefix(link, "skill:")))
	}
	return filepath.Join(filepath.Dir(from), filepath.FromSlash(link))
}

func hasCitedSection(body, name string) bool {
	q := regexp.QuoteMeta(name)
	heading := regexp.MustCompile(`(?m)^#+\s+` + q + `\s*$`)
	label := regexp.MustCompile(`(?m)^\s*(?:[-*]\s+|\d+\.\s+)?\*\*` + q + `(?:[.:,]\*\*|\*\*[.:,])`)
	return heading.MatchString(body) || label.MatchString(body)
}

func hasCitedStep(body, n string) bool {
	return regexp.MustCompile(fmt.Sprintf(`(?m)^%s\.\s`, regexp.QuoteMeta(n))).MatchString(body)
}
