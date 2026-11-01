package opencode

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
)

func TestContextLoaderSnapshotRendering(t *testing.T) {
	template := []byte("const data = " + contextSnapshotMarker + ";\n")
	files := map[string][]byte{
		"content/skills/flow-plan/SKILL.md":     []byte("\ufeff\"'` \\ 雪😀\r\n" + contextSnapshotMarker + "\n\u2028\x00"),
		"content/skills/harness-audit/SKILL.md": []byte("# Harness\n"),
		"content/skills/unused/SKILL.md":        []byte("Not a selected role"),
	}
	script, metadata, err := RenderContextLoader(template, files)
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Schema     int                                             `json:"schema"`
		References []struct{ Role, Resource, Text, SHA256 string } `json:"references"`
	}
	if err := json.Unmarshal(bytes.TrimSuffix(bytes.TrimPrefix(script, []byte("const data = ")), []byte(";\n")), &data); err != nil {
		t.Fatal(err)
	}
	if data.Schema != 2 || len(data.References) != 2 {
		t.Fatal("snapshot must have exactly the two fixed roles")
	}
	var provenance struct {
		Schema     int                 `json:"schema"`
		References []map[string]string `json:"references"`
	}
	if err := json.Unmarshal(metadata, &provenance); err != nil {
		t.Fatal(err)
	}
	if provenance.Schema != 2 || len(provenance.References) != 2 {
		t.Fatal("invalid provenance schema")
	}
	wantReferences := []struct{ role, resource string }{
		{"hive-review-harness", "harness-audit/SKILL.md"},
		{"hive-review-plan", "flow-plan/SKILL.md"},
	}
	for i, mapping := range wantReferences {
		ref := data.References[i]
		want := files["content/skills/"+mapping.resource]
		digest := sha256.Sum256(want)
		if ref.Role != mapping.role || ref.Resource != mapping.resource || !bytes.Equal([]byte(ref.Text), want) || ref.SHA256 != hex.EncodeToString(digest[:]) {
			t.Fatal("snapshot lost identity, exact bytes or digest")
		}
		meta := provenance.References[i]
		if len(meta) != 3 || meta["role"] != ref.Role || meta["resource"] != ref.Resource || meta["sha256"] != ref.SHA256 {
			t.Fatal("metadata must contain only matching provenance")
		}
	}
	second, secondMeta, err := RenderContextLoader(template, files)
	if err != nil || !bytes.Equal(script, second) || !bytes.Equal(metadata, secondMeta) {
		t.Fatal("rendering must be deterministic")
	}
}
