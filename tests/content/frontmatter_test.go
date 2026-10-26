package content

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFrontmatterPlainScalars rejects a plain (unquoted) frontmatter value
// that strict YAML parsers refuse: one containing ": " or " #", or ending in
// ":". Some hosts drop a skill whose frontmatter fails to parse, so such a
// value must be quoted.
func TestFrontmatterPlainScalars(t *testing.T) {
	checked := 0
	err := filepath.WalkDir("../../content", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		if !strings.HasPrefix(text, "---\n") {
			return nil
		}
		end := strings.Index(text[4:], "\n---\n")
		if end < 0 {
			return nil
		}
		checked++
		rel := filepath.ToSlash(path[len("../../"):])
		for i, line := range strings.Split(text[4:4+end], "\n") {
			key, value, found := strings.Cut(line, ": ")
			if !found || key == "" || strings.HasPrefix(key, " ") {
				continue
			}
			value = strings.TrimSpace(value)
			if value == "" || strings.ContainsAny(value[:1], "\"'|>[{") {
				continue
			}
			if strings.Contains(value, ": ") || strings.Contains(value, " #") || strings.HasSuffix(value, ":") {
				t.Errorf("%s:%d: unquoted %q value is not valid YAML; quote it", rel, i+2, key)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked == 0 {
		t.Fatal("no frontmatter found under content")
	}
}
