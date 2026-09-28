// Voice is an optional, off-by-default tone layer on top of Hive's
// communication rules, rendered into its own managed block (VoiceBegin /
// VoiceEnd) placed after the Hive block. This file owns the voice catalogue:
// listing the available voices and rendering one voice's text. See
// design.md "Generación del texto". It does not yet plan, apply, or recover
// a voice block; a later change wires VoiceSetting/VoiceSpan into Plan and
// State.
package management

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// VoiceInfo is one entry of ListVoices: a voice's ID and the first
// description line of its source file.
type VoiceInfo struct {
	ID          string
	Description string
}

const preambleFile = "preamble.md"

// voiceRenderVersion is part of RenderVoice's sourceHash. Bump it whenever
// this file's generated address/intensity wording (addressLine,
// intensityLine) changes, so plan install (a later change) regenerates a
// stored VoiceSpan after such a wording change even though the source .md
// files under content/voices/ did not change.
const voiceRenderVersion = "v1"

var voiceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ListVoices returns the voices available under source's content/voices/
// directory: every "*.md" file except preamble.md, its ID being the file
// stem and its description the first non-empty line of its text, sorted by
// ID. An empty or absent directory yields no voices, not an error, only when
// the directory is simply empty; a missing content/voices/ altogether is
// reported as the underlying read error.
func ListVoices(source string) ([]VoiceInfo, error) {
	dir := filepath.Join(source, VoicesSource)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var voices []VoiceInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		if id == "preamble" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		voices = append(voices, VoiceInfo{ID: id, Description: firstNonEmptyLine(data)})
	}
	sort.Slice(voices, func(i, j int) bool { return voices[i].ID < voices[j].ID })
	return voices, nil
}

func firstNonEmptyLine(data []byte) string {
	for _, line := range strings.Split(string(data), "\n") {
		if t := strings.TrimSpace(line); t != "" {
			return t
		}
	}
	return ""
}

// RenderVoice renders one voice's block text: the shared preamble, the
// voice's own description, an address line and an intensity line, per
// design.md "Generación del texto". It validates entirely before returning
// any error, so it never partially renders. Address defaults to "none" and
// Intensity to "subtle" when left empty.
//
// sourceHash identifies exactly what this body was rendered from: the
// preamble's bytes, the voice file's bytes, and voiceRenderVersion, each
// separated by a single 0x00 byte before hashing with SHA-256 (hex-encoded).
// The 0x00 separators, which cannot occur in either UTF-8 markdown source
// (validated elsewhere as valid UTF-8 by validateRelease for release
// content, and never produced by an editor for these hand-authored files),
// keep the byte ranges from colliding at their boundaries the way naive
// concatenation could. It deliberately excludes the user's address,
// intensity and name choice: a later change compares that choice
// separately, against State.VoiceSetting, since a choice change alone
// never needs the preamble or voice file re-read.
func RenderVoice(source string, v VoiceSetting) (body []byte, sourceHash string, err error) {
	if v.Address == "" {
		v.Address = "none"
	}
	if v.Intensity == "" {
		v.Intensity = "subtle"
	}
	if v.Address != "sir" && v.Address != "name" && v.Address != "none" {
		return nil, "", fmt.Errorf("address must be sir, name or none")
	}
	if v.Intensity != "subtle" && v.Intensity != "marked" {
		return nil, "", fmt.Errorf("intensity must be subtle or marked")
	}
	if v.Address == "name" {
		if err := validateVoiceName(v.Name); err != nil {
			return nil, "", err
		}
	} else if v.Name != "" {
		return nil, "", fmt.Errorf("name must be empty unless address is name")
	}
	if v.ID == "preamble" || !voiceIDPattern.MatchString(v.ID) {
		return nil, "", fmt.Errorf("unknown voice %q", v.ID)
	}
	dir := filepath.Join(source, VoicesSource)
	preamble, err := os.ReadFile(filepath.Join(dir, preambleFile))
	if err != nil {
		return nil, "", fmt.Errorf("reading voice preamble: %w", err)
	}
	voiceText, err := os.ReadFile(filepath.Join(dir, v.ID+".md"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("unknown voice %q", v.ID)
		}
		return nil, "", err
	}
	if err := rejectVoiceMarkers(preamble); err != nil {
		return nil, "", fmt.Errorf("voice preamble: %w", err)
	}
	if err := rejectVoiceMarkers(voiceText); err != nil {
		return nil, "", fmt.Errorf("voice %s: %w", v.ID, err)
	}

	var buf bytes.Buffer
	buf.WriteString(strings.TrimRight(string(preamble), "\n"))
	buf.WriteString("\n\n")
	buf.WriteString(strings.TrimRight(string(voiceText), "\n"))
	buf.WriteString("\n\n")
	buf.WriteString(addressLine(v))
	buf.WriteString("\n")
	buf.WriteString(intensityLine(v.Intensity))
	buf.WriteString("\n")

	h := sha256.New()
	h.Write(preamble)
	h.Write([]byte{0})
	h.Write(voiceText)
	h.Write([]byte{0})
	h.Write([]byte(voiceRenderVersion))
	return buf.Bytes(), hex.EncodeToString(h.Sum(nil)), nil
}

// addressLine writes the treatment line per design.md's three forms. The
// "sir" form is deliberately phrased for the model reading it at runtime,
// since the block is generated once but read in whatever session language
// is then active: it names the *kind* of form to use (the formal address,
// gender-neutral elsewhere), not a fixed word, because a fixed English
// "sir" would be wrong in, say, a Spanish session that wants "señor".
func addressLine(v VoiceSetting) string {
	switch v.Address {
	case "sir":
		return "Address: use the formal address the user's session language calls for (for example, \"sir\" in English), and do not mark gender anywhere else in the message."
	case "name":
		return fmt.Sprintf("Address: address the user by their given name, %s.", v.Name)
	default:
		return "Address: use no form of address for the user."
	}
}

func intensityLine(intensity string) string {
	if intensity == "marked" {
		return "Intensity: let this voice show in most messages."
	}
	return "Intensity: let this voice show only in word choice and occasional turns of phrase."
}

// validateVoiceName checks a user-chosen given name before it can reach a
// generated instruction file (see security-boundaries.md: this is the one
// external input the voice layer writes into such a file). Letters
// (including accented ones), spaces, apostrophes and hyphens are accepted,
// 1 to 40 characters, with at least one letter so a name of only spaces,
// apostrophes or hyphens is rejected; this excludes newlines and every
// character the four managed-block markers use, without a separate marker
// check.
func validateVoiceName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required when address is name")
	}
	runes := []rune(name)
	if len(runes) > 40 {
		return fmt.Errorf("name must be at most 40 characters")
	}
	hasLetter := false
	for _, r := range runes {
		if unicode.IsLetter(r) {
			hasLetter = true
			continue
		}
		if r == ' ' || r == '\'' || r == '-' {
			continue
		}
		return fmt.Errorf("name contains an unsupported character")
	}
	if !hasLetter {
		return fmt.Errorf("name must contain at least one letter")
	}
	return nil
}

func rejectVoiceMarkers(b []byte) error {
	for _, m := range []string{Begin, End, VoiceBegin, VoiceEnd} {
		if bytes.Contains(b, []byte(m)) {
			return fmt.Errorf("contains a reserved marker")
		}
	}
	return nil
}
