package opencode

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"
)

const contextSnapshotMarker = "__HIVE_CONTEXT_SNAPSHOT_DATA__"

var contextReferences = [...]struct{ role, resource string }{
	{"hive-review-harness", "harness-audit/SKILL.md"},
	{"hive-review-plan", "flow-plan/SKILL.md"},
}

// RenderContextLoader derives both inactive resources solely from the release
// payload. Metadata is local provenance, not an authenticated runtime capability.
func RenderContextLoader(template []byte, files map[string][]byte) ([]byte, []byte, error) {
	if !utf8.Valid(template) || bytes.Count(template, []byte(contextSnapshotMarker)) != 1 {
		return nil, nil, fmt.Errorf("invalid OpenCode context template: expected valid UTF-8 and one snapshot marker")
	}
	type identity struct {
		Role     string `json:"role"`
		Resource string `json:"resource"`
		SHA256   string `json:"sha256"`
	}
	type reference struct {
		Role     string `json:"role"`
		Resource string `json:"resource"`
		Text     string `json:"text"`
		SHA256   string `json:"sha256"`
	}
	data := struct {
		Schema     int         `json:"schema"`
		References []reference `json:"references"`
	}{Schema: 2}
	metadata := struct {
		Schema     int        `json:"schema"`
		References []identity `json:"references"`
	}{Schema: 2}
	for _, mapping := range contextReferences {
		owner := "content/skills/" + strings.SplitN(mapping.resource, "/", 2)[0] + "/SKILL.md"
		if entry, ok := files[owner]; !ok || !utf8.Valid(entry) || len(bytes.TrimSpace(entry)) == 0 {
			return nil, nil, fmt.Errorf("missing or invalid OpenCode snapshot owner: %s", owner)
		}
		source := "content/skills/" + mapping.resource
		text, ok := files[source]
		if !ok || !utf8.Valid(text) || len(bytes.TrimSpace(text)) == 0 {
			return nil, nil, fmt.Errorf("missing or invalid OpenCode snapshot reference: %s", source)
		}
		digest := sha256.Sum256(text)
		sum := hex.EncodeToString(digest[:])
		data.References = append(data.References, reference{mapping.role, mapping.resource, string(text), sum})
		metadata.References = append(metadata.References, identity{mapping.role, mapping.resource, sum})
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	provenance, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	// Replace exactly once; marker-like resource text is data, never a new hole.
	return bytes.Replace(template, []byte(contextSnapshotMarker), encoded, 1), append(provenance, '\n'), nil
}
