package management

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"
	"tricell-hive/integrations/mdlinks"
)

// These checks cover authored Markdown instruction links, not arbitrary prose,
// generated assets, remote URLs, heading anchors or runtime project inputs.
var inlineResourceLink = regexp.MustCompile(`\[[^\]\n]*\]\(\s*(<[^>\n]+>|[^\s)]+)(?:\s+"[^"\n]*")?\s*\)`)
var definedResourceLink = regexp.MustCompile(`(?m)^ {0,3}\[[^\]\n]+\]:\s*(<[^>\n]+>|[^\s]+)`)
var personalResourcePath = regexp.MustCompile(`(?:/Users/|/home/|/Volumes/|[A-Za-z]:\\Users\\)`)

func instructionMarkdown(source string) bool {
	return strings.HasSuffix(source, ".md") && (strings.HasPrefix(source, "content/agents/") ||
		(strings.HasPrefix(source, "content/skills/") && (strings.HasSuffix(source, "/SKILL.md") || strings.Contains(source, "/references/"))))
}

func validateInstructionReferences(r Release) error {
	entries := make(map[string]bool, len(r.Files))
	for _, f := range r.Files {
		entries[f.Path] = true
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
			if err := validateInstructionLink(f.Path, raw, entries); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateInstructionLink(source, raw string, entries map[string]bool) error {
	fail := func(reason string) error {
		return fmt.Errorf("instruction reference %s -> %s: %s", source, raw, reason)
	}
	if strings.HasPrefix(raw, "#") {
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fail("invalid destination")
	}
	if u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "mailto" {
		return nil
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
		return nil
	}
	if strings.HasSuffix(u.Path, "/") || strings.HasSuffix(u.Opaque, "/") {
		for entry := range entries {
			if strings.HasPrefix(entry, strings.TrimSuffix(resource, "/")+"/") {
				return nil
			}
		}
	}
	return fail("target is not included in this release")
}
