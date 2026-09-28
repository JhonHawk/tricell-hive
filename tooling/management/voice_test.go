package management

// TDD tests for the voice catalogue: ListVoices and RenderVoice. They run
// against synthetic content/voices/ trees in temp dirs, never the real
// content/voices/ tree another worker is writing concurrently (see
// design.md "Generación del texto").

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// voiceFixture builds a synthetic source checkout with content/voices/
// containing the given preamble and named voice files, returning the source
// root.
func voiceFixture(t *testing.T, preamble string, voices map[string]string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "content", "voices")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "preamble.md"), []byte(preamble), 0600); err != nil {
		t.Fatal(err)
	}
	for id, text := range voices {
		if err := os.WriteFile(filepath.Join(dir, id+".md"), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

const testPreamble = "Priority: this section never overrides Hive's rules.\n"

// --- ListVoices --------------------------------------------------------------

func TestListVoicesExcludesPreambleAndSorts(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{
		"senior-direct": "Direct and blunt.\nMore detail.\n",
		"jarvis":        "Warm, formal, a little dry.\nMore detail.\n",
	})
	voices, err := ListVoices(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 2 {
		t.Fatalf("expected 2 voices, got %d: %+v", len(voices), voices)
	}
	if voices[0].ID != "jarvis" || voices[1].ID != "senior-direct" {
		t.Fatalf("expected sorted [jarvis senior-direct], got [%s %s]", voices[0].ID, voices[1].ID)
	}
	if voices[0].Description != "Warm, formal, a little dry." {
		t.Fatalf("unexpected description: %q", voices[0].Description)
	}
}

func TestListVoicesDescriptionSkipsBlankLines(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{
		"mentor": "\n\n   \nPatient and encouraging.\nMore detail.\n",
	})
	voices, err := ListVoices(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 1 || voices[0].Description != "Patient and encouraging." {
		t.Fatalf("got %+v", voices)
	}
}

func TestListVoicesEmptyDirectory(t *testing.T) {
	source := voiceFixture(t, testPreamble, nil)
	voices, err := ListVoices(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) != 0 {
		t.Fatalf("expected no voices, got %+v", voices)
	}
}

// --- RenderVoice: success cases ------------------------------------------------

// This asserts the body's exact structure — preamble, then voice text, then
// exactly one Address: line and exactly one Intensity: line, nothing else —
// without hardcoding the generated lines' wording, so it stays meaningful
// if that wording changes (see item 4 of the T1 fix round).
func TestRenderVoiceContainsPreambleThenVoiceThenAddressThenIntensity(t *testing.T) {
	preamble := "Priority: this section never overrides Hive's rules.\n"
	voice := "Warm, formal, a little dry.\n"
	source := voiceFixture(t, preamble, map[string]string{"jarvis": voice})
	body, hash, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "none", Intensity: "subtle"})
	if err != nil {
		t.Fatal(err)
	}
	if hash == "" {
		t.Fatal("expected a non-empty source hash")
	}
	text := string(body)
	trimmedPreamble := strings.TrimRight(preamble, "\n")
	if !strings.HasPrefix(text, trimmedPreamble) {
		t.Fatalf("expected the body to start with the exact preamble bytes, got %q", text)
	}
	rest := strings.TrimPrefix(text, trimmedPreamble)
	rest = strings.TrimPrefix(rest, "\n\n")
	trimmedVoice := strings.TrimRight(voice, "\n")
	if !strings.HasPrefix(rest, trimmedVoice) {
		t.Fatalf("expected the voice text right after the preamble, got %q", rest)
	}
	rest = strings.TrimPrefix(rest, trimmedVoice)
	rest = strings.TrimPrefix(rest, "\n\n")
	rest = strings.TrimRight(rest, "\n")
	lines := strings.Split(rest, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected exactly two trailing lines (address, intensity) and nothing after, got %q", lines)
	}
	if !strings.HasPrefix(lines[0], "Address:") {
		t.Fatalf("expected the first trailing line to start with %q, got %q", "Address:", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Intensity:") {
		t.Fatalf("expected the second trailing line to start with %q, got %q", "Intensity:", lines[1])
	}
}

func TestRenderVoiceDefaultsAddressNoneAndIntensitySubtle(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	body, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToLower(string(body)), "no form of address") {
		t.Fatalf("expected the default none-address line, got %q", body)
	}
	if !strings.Contains(strings.ToLower(string(body)), "word choice") {
		t.Fatalf("expected the default subtle-intensity line, got %q", body)
	}
}

func TestRenderVoiceAddressSirIsGenderNeutral(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	body, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "sir", Intensity: "subtle"})
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(body))
	if !strings.Contains(lower, "formal") {
		t.Fatalf("expected a formal-address instruction, got %q", body)
	}
	if !strings.Contains(lower, "gender") {
		t.Fatalf("expected the address instruction to call out gender neutrality, got %q", body)
	}
}

func TestRenderVoiceAddressNameUsesGivenName(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	body, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: "Juan Manuel", Intensity: "marked"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "Juan Manuel") {
		t.Fatalf("expected the given name in the body, got %q", body)
	}
	if !strings.Contains(strings.ToLower(string(body)), "most messages") {
		t.Fatalf("expected the marked-intensity line, got %q", body)
	}
}

func TestRenderVoiceSourceHashChangesWithVoiceText(t *testing.T) {
	sourceA := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	sourceB := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Different text entirely.\n"})
	_, hashA, err := RenderVoice(sourceA, VoiceSetting{ID: "jarvis"})
	if err != nil {
		t.Fatal(err)
	}
	_, hashB, err := RenderVoice(sourceB, VoiceSetting{ID: "jarvis"})
	if err != nil {
		t.Fatal(err)
	}
	if hashA == hashB {
		t.Fatal("expected different source hashes for different voice text")
	}
}

// The source hash must let T2 detect that a stored VoiceSpan needs
// regeneration after the address/intensity wording generated by this file
// changes, even when the source .md files under content/voices/ did not
// change themselves. It must therefore be more than a hash of the preamble
// and voice bytes alone.
func TestRenderVoiceSourceHashIncludesRenderVersion(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	_, gotHash, err := RenderVoice(source, VoiceSetting{ID: "jarvis"})
	if err != nil {
		t.Fatal(err)
	}
	preamble, err := os.ReadFile(filepath.Join(source, "content", "voices", "preamble.md"))
	if err != nil {
		t.Fatal(err)
	}
	voice, err := os.ReadFile(filepath.Join(source, "content", "voices", "jarvis.md"))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	h.Write(preamble)
	h.Write([]byte{0})
	h.Write(voice)
	withoutVersion := hex.EncodeToString(h.Sum(nil))
	if gotHash == withoutVersion {
		t.Fatal("expected the source hash to also cover a render version, not just preamble+voice bytes")
	}
}

func TestRenderVoiceSourceHashStableForSameContent(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	_, hashA, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "none"})
	if err != nil {
		t.Fatal(err)
	}
	_, hashB, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: "Ana"})
	if err != nil {
		t.Fatal(err)
	}
	if hashA != hashB {
		t.Fatal("expected the same source hash regardless of address/intensity, since it hashes the sources, not the choice")
	}
}

// --- RenderVoice: errors, none of which write anything -------------------------

func TestRenderVoiceUnknownID(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "nonexistent"}); err == nil {
		t.Fatal("expected an error for an unknown voice ID")
	}
}

// preamble.md matches voiceIDPattern and exists, but it is the shared
// preamble, not a voice: ID "preamble" must be rejected as unknown rather
// than rendering the preamble twice (once as the preamble, once as the
// "voice" text).
func TestRenderVoiceRejectsPreambleAsVoiceID(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "preamble"}); err == nil {
		t.Fatal("expected an error for voice ID \"preamble\"")
	}
}

// This test must not pass merely because voiceIDPattern rejects "/" and
// "..": it plants a real, marker-free file exactly where "../evil" would
// resolve to (content/voices/../evil.md -> content/evil.md) if path
// traversal were not blocked, so it would fail on its own — reading that
// file successfully and returning no error — if that ID validation were
// ever removed or weakened to allow this shape.
func TestRenderVoiceUnknownIDRejectsPathTraversal(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if err := os.WriteFile(filepath.Join(source, "content", "evil.md"), []byte("Traversal payload.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "../evil"}); err == nil {
		t.Fatal("expected an error for a path-traversal voice ID")
	}
}

func TestRenderVoiceInvalidAddress(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "maam"}); err == nil {
		t.Fatal("expected an error for an address not in {sir, name, none}")
	}
}

func TestRenderVoiceInvalidIntensity(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Intensity: "loud"}); err == nil {
		t.Fatal("expected an error for an intensity not in {subtle, marked}")
	}
}

func TestRenderVoiceAddressNameWithoutName(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name"}); err == nil {
		t.Fatal("expected an error for address name without a Name")
	}
}

func TestRenderVoiceInvalidNameCharacters(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	cases := []string{"Juan<script>", "Robert; DROP TABLE", "line\nbreak", "Name123"}
	for _, name := range cases {
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: name}); err == nil {
			t.Fatalf("expected an error for invalid name %q", name)
		}
	}
}

func TestRenderVoiceValidNameCharacters(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	cases := []string{"Ana", "Jean-Luc", "O'Brien", "María José"}
	for _, name := range cases {
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: name}); err != nil {
			t.Fatalf("expected %q to be accepted, got %v", name, err)
		}
	}
}

func TestRenderVoiceNameTooLong(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	name := strings.Repeat("a", 41)
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: name}); err == nil {
		t.Fatal("expected an error for a name longer than 40 characters")
	}
}

func TestRenderVoiceNameAtMaxLengthIsAccepted(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	name := strings.Repeat("a", 40)
	if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: name}); err != nil {
		t.Fatalf("expected a 40-character name to be accepted, got %v", err)
	}
}

func TestRenderVoiceRejectsWhitespaceOnlyName(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	for _, name := range []string{" ", "   ", "-", "--", "' '"} {
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: "name", Name: name}); err == nil {
			t.Fatalf("expected an error for a name with no letters: %q", name)
		}
	}
}

// A Name is only ever meaningful with Address "name"; rejecting it otherwise
// keeps an invalid or unintended Name from ever reaching State once a later
// change wires VoiceSetting into it.
func TestRenderVoiceRejectsNameUnlessAddressIsName(t *testing.T) {
	source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n"})
	for _, address := range []string{"none", "sir", ""} {
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis", Address: address, Name: "Ana"}); err == nil {
			t.Fatalf("expected an error for a non-empty Name with address %q", address)
		}
	}
}

// rejectVoiceMarkers checks each of the four markers independently, with no
// pairing requirement, so a single marker line alone (no matching
// begin/end partner) must already be rejected — in either source file.
func TestRenderVoiceRejectsEachMarkerAloneInVoiceText(t *testing.T) {
	for _, marker := range []string{Begin, End, VoiceBegin, VoiceEnd} {
		source := voiceFixture(t, testPreamble, map[string]string{"jarvis": "Warm.\n" + marker + "\n"})
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis"}); err == nil {
			t.Fatalf("expected an error for voice text containing %q alone", marker)
		}
	}
}

func TestRenderVoiceRejectsEachMarkerAloneInPreamble(t *testing.T) {
	for _, marker := range []string{Begin, End, VoiceBegin, VoiceEnd} {
		source := voiceFixture(t, testPreamble+marker+"\n", map[string]string{"jarvis": "Warm.\n"})
		if _, _, err := RenderVoice(source, VoiceSetting{ID: "jarvis"}); err == nil {
			t.Fatalf("expected an error for a preamble containing %q alone", marker)
		}
	}
}
