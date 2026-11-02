package content

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

// skillDescriptionLimit is the longest skill description, in characters, that
// the Agent Skills format accepts; a host may reject or truncate a longer one,
// which would hide when to use the skill.
const skillDescriptionLimit = 1024

// frontmatterScalar returns the single-line scalar value of key in a document's
// frontmatter, with one level of YAML quoting removed.
func frontmatterScalar(text, key string) (string, bool, error) {
	if !strings.HasPrefix(text, "---\n") {
		return "", false, nil
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return "", false, nil
	}
	for _, line := range strings.Split(text[4:4+end], "\n") {
		value, found := strings.CutPrefix(line, key+":")
		if !found {
			continue
		}
		value = strings.TrimSpace(value)
		switch {
		case strings.HasPrefix(value, `"`):
			unquoted, err := strconv.Unquote(value)
			return unquoted, true, err
		case strings.HasPrefix(value, "'"):
			if len(value) < 2 || !strings.HasSuffix(value, "'") {
				return "", true, strconv.ErrSyntax
			}
			return strings.ReplaceAll(value[1:len(value)-1], "''", "'"), true, nil
		case value == "|" || value == ">" || strings.HasPrefix(value, "|") || strings.HasPrefix(value, ">"):
			return "", true, strconv.ErrSyntax
		}
		return value, true, nil
	}
	return "", false, nil
}

// TestSkillFrontmatterNameAndDescription checks that each skill's `name`
// equals its folder name (a host resolves the skill by that name) and that its
// `description` is present and within the limit.
func TestSkillFrontmatterNameAndDescription(t *testing.T) {
	paths, err := filepath.Glob("../../content/skills/*/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no skills found under content/skills")
	}
	for _, path := range paths {
		rel := filepath.ToSlash(path[len("../../"):])
		folder := filepath.Base(filepath.Dir(path))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		name, ok, err := frontmatterScalar(text, "name")
		switch {
		case err != nil:
			t.Errorf("%s: `name` is not a single-line scalar", rel)
		case !ok:
			t.Errorf("%s: frontmatter has no `name`", rel)
		case name != folder:
			t.Errorf("%s: `name` is %q, want the folder name %q", rel, name, folder)
		}
		description, ok, err := frontmatterScalar(text, "description")
		switch {
		case err != nil:
			t.Errorf("%s: `description` is not a single-line scalar", rel)
		case !ok || strings.TrimSpace(description) == "":
			t.Errorf("%s: frontmatter has no non-empty `description`", rel)
		case utf8.RuneCountInString(description) > skillDescriptionLimit:
			t.Errorf("%s: `description` is %d characters, over the %d-character limit", rel, utf8.RuneCountInString(description), skillDescriptionLimit)
		}
	}
}
