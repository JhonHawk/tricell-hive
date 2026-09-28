package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// voiceBudget caps each voice's always-loaded cost: the preamble plus one
// voice text is added to every session that enables it.
const voiceBudget = 3072

var voiceMarkers = []string{
	"<!-- === TRICELL HIVE RULES:BEGIN === -->",
	"<!-- === TRICELL HIVE RULES:END === -->",
	"<!-- === TRICELL HIVE VOICE:BEGIN === -->",
	"<!-- === TRICELL HIVE VOICE:END === -->",
}

func readVoice(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("../../content/voices", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestVoicePreambleKeepsHiveRulesFirst pins the preamble points the voice
// layer relies on: precedence stated by text rather than position, the
// explicit list of what a voice never changes, where tone may appear, and
// when a reader must ignore the voice.
func TestVoicePreambleKeepsHiveRulesFirst(t *testing.T) {
	preamble := readVoice(t, "preamble.md")
	for _, want := range []string{
		"never overrides the Hive rules",
		"follow the Hive rule",
		"The first sentence carries the result",
		"labeled sections",
		"plain everyday words",
		"gloss",
		"without flattery",
		"session language",
		"transitions, closings, and the form of address",
		"If another agent launched you",
		"file, commit, pull request, ticket, specification, or another agent",
		"ignore this voice section entirely",
		"only in transitions",
		"without metaphors",
		"without flattery or deference",
		"keep plain wording",
	} {
		if !strings.Contains(preamble, want) {
			t.Errorf("preamble.md must state %q", want)
		}
	}
}

func TestVoiceTextsExistStayUnmarkedAndWithinBudget(t *testing.T) {
	preamble := readVoice(t, "preamble.md")
	for _, id := range []string{"jarvis", "senior-direct", "mentor"} {
		text := readVoice(t, id+".md")
		for _, marker := range voiceMarkers {
			if strings.Contains(preamble, marker) || strings.Contains(text, marker) {
				t.Errorf("%s or preamble.md contains the reserved marker %q", id, marker)
			}
		}
		if size := len(preamble) + len(text); size > voiceBudget {
			t.Errorf("%s: preamble plus voice is %d bytes, over the %d-byte budget", id, size, voiceBudget)
		}
	}
}
