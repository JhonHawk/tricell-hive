package content

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// linkTarget matches the destination of a Markdown link. Release link
// validation checks that relative targets exist but skips heading anchors, so
// a renamed heading would silently break every link that points to it.
var linkTarget = regexp.MustCompile(`\]\(([^)\s]+)\)`)

var headingLine = regexp.MustCompile(`^#{1,6}\s+(.+?)\s*#*\s*$`)

// headingSlug follows GitHub's anchor rule: lowercase, drop every character
// that is not a letter, digit, space, hyphen or underscore, and turn spaces
// into hyphens.
func headingSlug(heading string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(heading) {
		switch {
		case r == ' ':
			b.WriteRune('-')
		case r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// fileAnchors returns the anchors a Markdown file defines through its
// headings, outside fenced code, with GitHub's suffixes for repeated ones.
func fileAnchors(t *testing.T, path string) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	anchors := map[string]bool{}
	seen := map[string]int{}
	inFence := false
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		m := headingLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		slug := headingSlug(m[1])
		if n := seen[slug]; n > 0 {
			anchors[slug+"-"+strconv.Itoa(n)] = true
		} else {
			anchors[slug] = true
		}
		seen[slug]++
	}
	return anchors
}

// linkSources lists the distributed Markdown files whose links a host reads.
func linkSources(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir("../../content", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestContentLinkAnchorsResolve(t *testing.T) {
	cache := map[string]map[string]bool{}
	for _, src := range linkSources(t) {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		inFence := false
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = !inFence
				continue
			}
			if inFence {
				continue
			}
			for _, m := range linkTarget.FindAllStringSubmatch(line, -1) {
				target, anchor, ok := strings.Cut(m[1], "#")
				if !ok || anchor == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
					continue
				}
				var path string
				switch {
				case target == "":
					path = src
				case strings.HasPrefix(target, "skill:"):
					path = filepath.Join("../../content/skills", filepath.FromSlash(strings.TrimPrefix(target, "skill:")))
				case strings.HasSuffix(target, ".md"):
					path = filepath.Join(filepath.Dir(src), filepath.FromSlash(target))
				default:
					continue
				}
				if _, err := os.Stat(path); err != nil {
					t.Errorf("%s links %s, but %s does not exist", src, m[1], path)
					continue
				}
				if cache[path] == nil {
					cache[path] = fileAnchors(t, path)
				}
				if !cache[path][anchor] {
					t.Errorf("%s links %s, but %s has no heading with anchor #%s", src, m[1], path, anchor)
				}
			}
		}
	}
}
