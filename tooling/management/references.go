package management

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"tricell-hive/integrations/mdlinks"
)

// These checks cover authored Markdown instruction links, including ATX heading
// anchors (#fragment), not arbitrary prose, generated assets, remote URLs or
// runtime project inputs.
var inlineResourceLink = regexp.MustCompile(`\[[^\]\n]*\]\(\s*(<[^>\n]+>|[^\s)]+)(?:\s+"[^"\n]*")?\s*\)`)
var definedResourceLink = regexp.MustCompile(`(?m)^ {0,3}\[[^\]\n]+\]:\s*(<[^>\n]+>|[^\s]+)`)
var personalResourcePath = regexp.MustCompile(`(?:/Users/|/home/|/Volumes/|[A-Za-z]:\\Users\\)`)
var atxHeading = regexp.MustCompile(`^ {0,3}#{1,6}[ \t]+(.+?)(?:[ \t]+#+)?[ \t]*$`)
var inlineMarkdownLink = regexp.MustCompile(`!?\[([^\]\n]*)\]\([^)\n]*\)`)
var slugDropped = regexp.MustCompile(`[^\p{L}\p{M}\p{N}\p{Pc} -]`)

func instructionMarkdown(source string) bool {
	return strings.HasSuffix(source, ".md") && (source == GlobalSource || strings.HasPrefix(source, "content/agents/") ||
		(strings.HasPrefix(source, "content/skills/") && (strings.HasSuffix(source, "/SKILL.md") || strings.Contains(source, "/references/"))))
}

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
		if base == "" {
			continue
		}
		slug := base
		for anchors[slug] {
			counts[base]++
			slug = base + "-" + strconv.Itoa(counts[base])
		}
		anchors[slug] = true
	}
	return anchors
}

func validateInstructionReferences(r Release) error {
	entries := make(map[string]bool, len(r.Files))
	contents := make(map[string]string, len(r.Files))
	anchors := make(map[string]map[string]bool)
	for _, f := range r.Files {
		entries[f.Path] = true
		if strings.HasSuffix(f.Path, ".md") {
			contents[f.Path] = string(f.Data)
		}
	}
	hasAnchor := func(target, fragment string) bool {
		if !strings.HasSuffix(target, ".md") {
			return false
		}
		set, ok := anchors[target]
		if !ok {
			set = headingAnchors(contents[target])
			anchors[target] = set
		}
		return set[fragment]
	}
	for _, f := range r.Files {
		if !instructionMarkdown(f.Path) {
			continue
		}
		body := mdlinks.OutsideFences(string(f.Data))
		if personalResourcePath.MatchString(body) {
			return fmt.Errorf("nonportable personal path in %s", f.Path)
		}
		matches := inlineResourceLink.FindAllStringSubmatch(body, -1)
		matches = append(matches, definedResourceLink.FindAllStringSubmatch(body, -1)...)
		for _, m := range matches {
			raw := strings.Trim(m[1], "<>")
			if err := validateInstructionLink(f.Path, raw, entries, hasAnchor); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateInstructionLink(source, raw string, entries map[string]bool, hasAnchor func(target, fragment string) bool) error {
	fail := func(reason string) error {
		return fmt.Errorf("instruction reference %s -> %s: %s", source, raw, reason)
	}
	if strings.HasPrefix(raw, "#") {
		frag, err := url.PathUnescape(raw[1:])
		if err != nil || frag == "" || !hasAnchor(source, frag) {
			return fail("missing heading anchor")
		}
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fail("invalid destination")
	}
	if u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "mailto" {
		return nil
	}
	hasHash := strings.Contains(raw, "#")
	if hasHash && u.Fragment == "" {
		return fail("missing heading anchor")
	}
	var resource string
	if u.Scheme == "skill" {
		// Logical resource identity, never a native tool or filesystem URI.
		resource, err = url.PathUnescape(u.Opaque)
		if err != nil || resource == "" || strings.ContainsAny(resource, "\\?#") || path.Clean(resource) != strings.TrimSuffix(resource, "/") || strings.HasPrefix(resource, "/") {
			return fail("invalid skill locator")
		}
		parts := strings.Split(resource, "/")
		if len(parts) < 2 || parts[0] == ".." || !entries["content/skills/"+parts[0]+"/SKILL.md"] {
			return fail("missing owning skill")
		}
		resource = "content/skills/" + resource
	} else {
		if u.Scheme != "" || u.Host != "" || u.RawQuery != "" {
			return fail("unsupported local destination")
		}
		resource = u.Path
		if resource == "" || strings.HasPrefix(resource, "/") || strings.HasPrefix(resource, "~") || strings.Contains(resource, "\\") {
			return fail("use a portable relative or skill locator")
		}
		if strings.HasPrefix(source, "content/agents/") {
			return fail("agents must use skill:owner/path locators, not source-relative paths")
		}
		resource = path.Join(path.Dir(source), resource)
	}
	if !strings.HasPrefix(resource, "content/skills/") {
		return fail("target escapes distributed skills")
	}
	if entries[resource] {
		if hasHash {
			frag, err := url.PathUnescape(u.Fragment)
			if err != nil || frag == "" || !hasAnchor(resource, frag) {
				return fail("missing heading anchor")
			}
		}
		return nil
	}
	if strings.HasSuffix(u.Path, "/") || strings.HasSuffix(u.Opaque, "/") {
		if hasHash {
			return fail("missing heading anchor")
		}
		for entry := range entries {
			if strings.HasPrefix(entry, strings.TrimSuffix(resource, "/")+"/") {
				return nil
			}
		}
	}
	return fail("target is not included in this release")
}
