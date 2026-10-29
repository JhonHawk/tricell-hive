package management

import (
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"tricell-hive/integrations/mdlinks"
)

// Release validation checks that an instruction link's target exists but not
// its heading anchor, so a renamed heading would silently break every link
// that points to it. This test closes that gap for the repository's own
// content, with the same link grammar and fence rule as release validation.

var atxHeading = regexp.MustCompile(`^ {0,3}#{1,6}[ \t]+(.+?)(?:[ \t]+#+)?[ \t]*$`)
var inlineMarkdownLink = regexp.MustCompile(`!?\[([^\]\n]*)\]\([^)\n]*\)`)
var slugDropped = regexp.MustCompile(`[^\p{L}\p{M}\p{N}\p{Pc} -]`)

// headingAnchors returns the anchors that a Markdown file's headings define
// outside fenced code, following github-slugger: the rendered text, lowercased,
// without punctuation, spaces turned into hyphens, and a numeric suffix for a
// slug already taken.
func headingAnchors(data string) map[string]bool {
	anchors := map[string]bool{}
	counts := map[string]int{}
	for _, line := range strings.Split(mdlinks.OutsideFences(data), "\n") {
		m := atxHeading.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := inlineMarkdownLink.ReplaceAllString(m[1], "$1")
		base := strings.ReplaceAll(slugDropped.ReplaceAllString(strings.ToLower(text), ""), " ", "-")
		slug := base
		for anchors[slug] {
			counts[base]++
			slug = base + "-" + strconv.Itoa(counts[base])
		}
		anchors[slug] = true
	}
	return anchors
}

func TestHeadingAnchorsFollowGitHub(t *testing.T) {
	data := "# Report a running or completed item\n## Step\n## Step\n## Step 1\n" +
		"   ## Indented `code` heading ##\n## See [questions](q.md)\n## Café_au lait\n" +
		"```md\n## Fenced\n```\n~~~\n## Tilde fenced\n~~~\n"
	got := headingAnchors(data)
	for _, want := range []string{"report-a-running-or-completed-item", "step", "step-1", "step-1-1", "indented-code-heading", "see-questions", "café_au-lait"} {
		if !got[want] {
			t.Errorf("missing anchor #%s in %v", want, got)
		}
	}
	for _, unwanted := range []string{"fenced", "tilde-fenced"} {
		if got[unwanted] {
			t.Errorf("fenced heading produced anchor #%s", unwanted)
		}
	}
}

func TestRepositoryInstructionAnchorsResolve(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "content"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		files[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	anchors := map[string]map[string]bool{}
	checked := 0
	for source, data := range files {
		if !instructionMarkdown(source) {
			continue
		}
		body := mdlinks.OutsideFences(data)
		matches := inlineResourceLink.FindAllStringSubmatch(body, -1)
		matches = append(matches, definedResourceLink.FindAllStringSubmatch(body, -1)...)
		for _, m := range matches {
			raw := strings.Trim(m[1], "<>")
			u, err := url.Parse(raw)
			if err != nil || u.Fragment == "" || u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "mailto" {
				continue
			}
			target := source
			switch {
			case u.Scheme == "skill":
				resource, err := url.PathUnescape(u.Opaque)
				if err != nil {
					t.Errorf("%s links %s: invalid skill locator", source, raw)
					continue
				}
				target = "content/skills/" + resource
			case u.Path != "":
				target = path.Join(path.Dir(source), u.Path)
			}
			text, ok := files[target]
			if !ok {
				t.Errorf("%s links %s, but %s is not a Markdown file in content/", source, raw, target)
				continue
			}
			if anchors[target] == nil {
				anchors[target] = headingAnchors(text)
			}
			checked++
			if !anchors[target][u.Fragment] {
				t.Errorf("%s links %s, but %s has no heading with anchor #%s", source, raw, target, u.Fragment)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no anchored instruction links found; the scan is not reading the content")
	}
}
