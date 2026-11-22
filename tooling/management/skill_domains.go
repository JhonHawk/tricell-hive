package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// A skill may declare domain folders in assets/domains.json: first-level
// folders of the skill, beside the four fixed ones, whose files install like
// those of references/. The manifest is the only list of them, so the manager
// and the content tests read the same one, and a folder nobody listed never
// reaches a host.
const domainsManifestKey = "domains"

var (
	domainsManifestSource = regexp.MustCompile(`^content/skills/([a-z0-9]+(-[a-z0-9]+)*)/assets/domains\.json$`)
	domainFolderName      = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	// fixedSkillFolders are collected for every skill; agents holds the policy
	// file the manager generates for a hidden skill.
	fixedSkillFolders = []string{"references", "scripts", "assets", "playbooks"}
	reservedDomains   = map[string]bool{"references": true, "scripts": true, "assets": true, "playbooks": true, "agents": true}
)

func domainManifestError(source string, reason string) error {
	return fmt.Errorf("invalid domain manifest %s: %s", source, reason)
}

// parseDomainManifest reads a manifest strictly: one JSON object whose only
// key is "domains", an array of distinct kebab-case folder names that are not
// reserved.
func parseDomainManifest(data []byte) ([]string, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return nil, fmt.Errorf("want a JSON object")
	}
	var names []string
	seenKey := false
	for dec.More() {
		tok, err := dec.Token()
		key, ok := tok.(string)
		if err != nil || !ok {
			return nil, fmt.Errorf("want a JSON object")
		}
		if key != domainsManifestKey {
			return nil, fmt.Errorf("unknown key %q: the only key is %q", key, domainsManifestKey)
		}
		if seenKey {
			return nil, fmt.Errorf("key %q appears twice", key)
		}
		seenKey = true
		if err := dec.Decode(&names); err != nil || names == nil {
			return nil, fmt.Errorf("key %q: want an array of folder names", key)
		}
	}
	if tok, err := dec.Token(); err != nil || tok != json.Delim('}') {
		return nil, fmt.Errorf("want a JSON object")
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("text after the object")
	}
	if !seenKey {
		return nil, fmt.Errorf("missing key %q", domainsManifestKey)
	}
	seen := map[string]bool{}
	for _, name := range names {
		switch {
		case reservedDomains[name]:
			return nil, fmt.Errorf("domain %q is a reserved folder name", name)
		case !domainFolderName.MatchString(name):
			return nil, fmt.Errorf("domain %q is not a kebab-case folder name", name)
		case seen[name]:
			return nil, fmt.Errorf("domain %q appears twice", name)
		}
		seen[name] = true
	}
	return names, nil
}

// skillDomainFolders reads the manifest of the skill whose entrypoint is
// skillEntry in a source checkout; a skill without one has no domains.
func skillDomainFolders(source, skillEntry string) ([]string, error) {
	rel, err := filepath.Rel(source, filepath.Join(filepath.Dir(skillEntry), "assets", "domains.json"))
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(source, rel))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names, err := parseDomainManifest(data)
	if err != nil {
		return nil, domainManifestError(filepath.ToSlash(rel), err.Error())
	}
	return names, nil
}

// releaseDomains returns the declared domains of each skill of a release, by
// skill name, checked against the release's own files: every listed folder
// holds a file, and every file in a first-level folder of a skill belongs to a
// fixed or a listed one.
func releaseDomains(r Release) (map[string]map[string]bool, error) {
	out := map[string]map[string]bool{}
	manifests := map[string]string{}
	for _, f := range r.Files {
		m := domainsManifestSource.FindStringSubmatch(f.Path)
		if m == nil {
			continue
		}
		names, err := parseDomainManifest(f.Data)
		if err != nil {
			return nil, domainManifestError(f.Path, err.Error())
		}
		set := map[string]bool{}
		for _, n := range names {
			set[n] = true
		}
		out[m[1]] = set
		manifests[m[1]] = f.Path
	}
	holds := map[string]bool{}
	for _, f := range r.Files {
		parts := strings.Split(f.Path, "/")
		if len(parts) < 5 || parts[0] != "content" || parts[1] != "skills" {
			continue
		}
		skill, folder := parts[2], parts[3]
		if reservedFolderIsFixed(folder) {
			continue
		}
		if !out[skill][folder] {
			return nil, domainManifestError("content/skills/"+skill+"/assets/domains.json", fmt.Sprintf("%s is in folder %q, which the manifest does not list", f.Path, folder))
		}
		holds[skill+"/"+folder] = true
	}
	for skill, set := range out {
		for folder := range set {
			if !holds[skill+"/"+folder] {
				return nil, domainManifestError(manifests[skill], fmt.Sprintf("domain %q is not a folder of the skill that holds a file", folder))
			}
		}
	}
	return out, nil
}

func reservedFolderIsFixed(folder string) bool {
	for _, f := range fixedSkillFolders {
		if f == folder {
			return true
		}
	}
	return false
}
