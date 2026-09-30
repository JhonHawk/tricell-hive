// Voice is an optional, off-by-default tone layer on top of Hive's
// communication rules, rendered into its own managed block (VoiceBegin /
// VoiceEnd) placed after the Hive block. This file owns the voice catalogue
// (listing the available voices and rendering one voice's text, see
// design.md "Generación del texto"), the standalone "voice set"/"voice off"
// plan (BuildVoicePlan), and install/remove's own voice-change generation
// (addVoiceChanges, called from BuildPlan).
package management

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"tricell-hive/integrations/target"
	"unicode"
)

// VoiceInfo is one entry of ListVoices: a voice's ID and the first
// description line of its source file.
type VoiceInfo struct {
	ID          string
	Description string
}

// CurrentVoice returns the home's own active voice setting, or nil when no
// voice is set. It reads state the same read-only way Status does
// (normalize + readState) but needs no host list: a voice is one per home,
// not one per host (design.md "La interfaz"). It never writes and adds no
// on-disk schema of its own — state.Voice is already recorded by
// BuildVoicePlan/Apply. The caller in tooling/cli (the Voice view) uses this
// instead of parsing Status's own formatted "id (address, intensity)"
// string, which cannot recover Name at all (formatVoiceStatus never
// includes it).
func CurrentVoice(o Options) (*VoiceSetting, error) {
	_, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	return state.Voice, nil
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
// validateVoiceSettingShape checks a VoiceSetting's own fields (not whether
// its ID actually exists in a given source): the ID pattern, an address in
// {sir, name, none}, an intensity in {subtle, marked}, and a Name required
// and valid exactly when Address is "name". RenderVoice uses it before
// rendering; validatePlan uses it to reject a hand-crafted or tampered
// plan's VoiceSetting the same way, since a saved plan may never have gone
// through RenderVoice at all.
func validateVoiceSettingShape(v VoiceSetting) error {
	if v.Address != "sir" && v.Address != "name" && v.Address != "none" {
		return fmt.Errorf("address must be sir, name or none")
	}
	if v.Intensity != "subtle" && v.Intensity != "marked" {
		return fmt.Errorf("intensity must be subtle or marked")
	}
	if v.Address == "name" {
		if err := validateVoiceName(v.Name); err != nil {
			return err
		}
	} else if v.Name != "" {
		return fmt.Errorf("name must be empty unless address is name")
	}
	if v.ID == "preamble" || !voiceIDPattern.MatchString(v.ID) {
		return fmt.Errorf("unknown voice %q", v.ID)
	}
	return nil
}

func RenderVoice(source string, v VoiceSetting) (body []byte, sourceHash string, err error) {
	if v.Address == "" {
		v.Address = "none"
	}
	if v.Intensity == "" {
		v.Intensity = "subtle"
	}
	if err := validateVoiceSettingShape(v); err != nil {
		return nil, "", err
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

// voiceRecordFromSpan adapts a VoiceSpan into the *Record shape owned and
// transform expect (files.go), so the same block mechanics that splice the
// Hive block, given hiveMarkers, also splice the voice block, given
// voiceMarkers. Only Managed is used: a voice span has no mode, leading
// separator or CreatedFile bookkeeping of its own.
func voiceRecordFromSpan(path string, span *VoiceSpan) *Record {
	if span == nil {
		return nil
	}
	return &Record{Target: target.Target{Path: path, Kind: "block"}, Managed: span.Managed}
}

// composeVoiceStep applies one voice-block transform (insert, replace or
// remove, depending on which of before/after is nil) to base. A first-time
// insert (before == nil) does not go through transform's generic insert
// path, which appends at EOF: the voice block belongs immediately after the
// Hive block's END line (design.md "Ubicación"), regardless of what text a
// user wrote after it, so it uses insertVoiceSpan instead. Replace and
// remove find the existing voice block by its own markers wherever it is,
// so transform already handles them position-independently.
//
// A change marked Gone is first verified against base: it fails closed
// unless vc.Before is really missing there. It then never removes or
// replaces anything: After == nil writes nothing, and After inserts the span
// as a first-time voice block. Without Gone the step behaves as it always
// did, so "voice off" still removes a block that is present.
func composeVoiceStep(path string, base snapshot, vc VoiceChange) (snapshot, error) {
	if vc.Gone {
		if vc.Before == nil {
			return snapshot{}, fmt.Errorf("%s: voice block marked gone without a registered span", path)
		}
		if err := verifyGone(base, *voiceRecordFromSpan(path, vc.Before), voiceMarkers); err != nil {
			return snapshot{}, err
		}
		if vc.After == nil {
			return base, nil
		}
		return insertVoiceSpan(base, vc.After.Managed)
	}
	if vc.Before == nil && vc.After != nil {
		return insertVoiceSpan(base, vc.After.Managed)
	}
	return transform(base, voiceRecordFromSpan(path, vc.Before), voiceRecordFromSpan(path, vc.After), voiceMarkers)
}

// insertVoiceSpan splices managed (a complete VoiceBegin..VoiceEnd block,
// trailing newline included) into s immediately after the Hive block's END
// line, with no separator: managedBlock already ends the Hive block with a
// newline, so whatever follows — nothing, or text a user wrote — already
// starts on its own line. s must already contain exactly one well-formed
// Hive block, which every caller guarantees (voice is only ever written to
// a file that already carries one).
func insertVoiceSpan(s snapshot, managed []byte) (snapshot, error) {
	a, b, err := blockRange(s.Data, hiveMarkers)
	if err != nil {
		return snapshot{}, err
	}
	if a < 0 {
		return snapshot{}, fmt.Errorf("voice block conflict: no Hive block to attach to")
	}
	out := make([]byte, 0, len(s.Data)+len(managed))
	out = append(out, s.Data[:b]...)
	out = append(out, managed...)
	out = append(out, s.Data[b:]...)
	mode := s.Mode
	if !s.Exists {
		mode = 0600
	}
	return snapshot{Exists: true, Data: out, Mode: mode}, nil
}

// checkVoiceConflict rejects, before any write, a hand-edited voice span
// (a registered span whose current bytes no longer match) or an
// unregistered voice block (markers present with no registered span),
// naming the voice block explicitly (see design.md "Conflicto"). A registered
// span that is missing, because the block or its whole file was deleted by
// hand, is not a conflict: gone reports it, so callers mark their change Gone.
func checkVoiceConflict(path string, s snapshot, hasSpan bool, existing VoiceSpan) (gone bool, err error) {
	if hasSpan {
		if err := owned(s, *voiceRecordFromSpan(path, &existing), voiceMarkers); err != nil {
			if isMissing(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}
	a, _, err := blockRange(s.Data, voiceMarkers)
	if err != nil {
		return false, fmt.Errorf("voice block conflict in %s: %w", path, err)
	}
	if a >= 0 {
		return false, fmt.Errorf("voice block conflict in %s: unregistered voice block present", path)
	}
	return false, nil
}

// voiceTargetPaths returns, sorted, every path in state.Records that carries
// a Hive block for a user-scope consumer at c.Home: the files voice
// operates on (see design.md "Rutas de la voz" — sourced from the block
// Records, not resolve(), so a Codex/Grok CLAUDE_CONFIG_DIR override is
// still the file the CLI actually reads).
func voiceTargetPaths(c target.Config, state State) []string {
	var paths []string
	for path, r := range state.Records {
		if r.Target.Kind != "block" {
			continue
		}
		for _, cons := range r.Consumers {
			if cons.Scope == "user" && cons.Context == c.Home {
				paths = append(paths, path)
				break
			}
		}
	}
	sort.Strings(paths)
	return paths
}

// filterUserConsumers narrows a consumer list to this scope's user-home
// ones: the consumer set a VoiceChange/VoiceSpan registers.
func filterUserConsumers(cons []Consumer, c target.Config) []Consumer {
	var out []Consumer
	for _, x := range cons {
		if x.Scope == "user" && x.Context == c.Home {
			out = append(out, x)
		}
	}
	return sortedConsumers(out)
}

// voiceConsumersForPath narrows a block record's consumers to this scope's
// user-home ones: the consumer set a VoiceChange/VoiceSpan at that path
// registers.
func voiceConsumersForPath(c target.Config, state State, path string) []Consumer {
	r, ok := state.Records[path]
	if !ok {
		return nil
	}
	return filterUserConsumers(r.Consumers, c)
}

// formatVoiceStatus renders a voice setting for a status row, e.g. "jarvis (sir, subtle)".
func formatVoiceStatus(v VoiceSetting) string {
	return fmt.Sprintf("%s (%s, %s)", v.ID, v.Address, v.Intensity)
}

// BuildVoicePlan builds a "voice set" or "voice off" plan: Action "voice",
// zero Changes, and one VoiceChange per instruction file that already
// carries a managed Hive block for a registered user-scope host (see
// design.md "Operaciones"). Voice is always user scope regardless of
// o.Scope: proposal.md excludes project scope from this change.
func BuildVoicePlan(action string, o Options, v VoiceSetting) (Plan, error) {
	var p Plan
	if action != "set" && action != "off" {
		return p, fmt.Errorf("action must be set or off")
	}
	o.Scope = "user"
	if action == "set" {
		// Normalize before storing or rendering, so State.Voice (and every
		// display derived from it, e.g. status's Voice field) never carries
		// an empty Address/Intensity, and repeating the same choice — spelled
		// out or left to default — is recognized as identical. RenderVoice
		// defaults these too, but only for its own rendering; the choice
		// stored here must match what was rendered.
		if v.Address == "" {
			v.Address = "none"
		}
		if v.Intensity == "" {
			v.Intensity = "subtle"
		}
	}
	c, dir, err := normalize(o)
	if err != nil {
		return p, err
	}
	if err := checkOnboarding(dir); err != nil {
		return p, err
	}
	if pend, err := read(filepath.Join(dir, "pending.json")); err != nil {
		return p, err
	} else if pend.Exists {
		return p, fmt.Errorf("unfinished operation: recover first")
	}
	state, sh, err := readState(dir)
	if err != nil {
		return p, err
	}
	paths := voiceTargetPaths(c, state)
	if len(paths) == 0 {
		return p, fmt.Errorf("Hive not installed")
	}

	var body []byte
	var srcHash string
	if action == "set" {
		body, srcHash, err = RenderVoice(o.Source, v)
		if err != nil {
			return Plan{}, err
		}
	}

	p = Plan{Version: stateVersion, Action: "voice", Config: c, StateDir: dir, StateHash: sh}
	if action == "set" {
		setting := v
		p.VoiceSetting = &setting
	}

	for _, path := range paths {
		s, err := read(path)
		if err != nil {
			return Plan{}, err
		}
		existing, hasSpan := state.VoiceSpans[path]
		if a, _, blockErr := blockRange(s.Data, hiveMarkers); blockErr != nil || a < 0 {
			// The Hive block was deleted by hand. "voice set" skips the file:
			// install restores the block first. "voice off" still removes a
			// voice block left behind, or drops the record of one that is gone
			// too (nothing is written); with no record there it skips the file.
			if blockErr != nil {
				return Plan{}, fmt.Errorf("%s: %w", path, blockErr)
			}
			if action == "set" {
				p.VoiceSkipped = append(p.VoiceSkipped, path)
				continue
			}
			if !hasSpan {
				continue
			}
		}
		gone, err := checkVoiceConflict(path, s, hasSpan, existing)
		if err != nil {
			return Plan{}, err
		}
		var before *VoiceSpan
		if hasSpan {
			span := existing
			before = &span
		}
		// Every path always gets a VoiceChange, even a no-op one (Before
		// equal to After, byte for byte) when nothing would change there:
		// BuildPlan's own Change list works the same way, and it is what
		// lets validatePlan require at least one voice change for a "voice"
		// action plan while PlanUnchanged, not an empty p.Voice, is what
		// tells a caller there is nothing to apply.
		var after *VoiceSpan
		if action == "set" {
			managed := managedBlock(body, s.Data, voiceMarkers)
			after = &VoiceSpan{Managed: managed, SourceHash: srcHash, Consumers: voiceConsumersForPath(c, state, path)}
		}
		p.Voice = append(p.Voice, VoiceChange{Path: path, Consumers: voiceConsumersForPath(c, state, path), Expected: finger(s), Before: before, After: after, Gone: gone})
	}
	if len(p.Voice) == 0 {
		return Plan{}, fmt.Errorf("no instruction file has a managed Hive block; run hive install first")
	}
	p.ID = planID(p)
	return p, nil
}

// addVoiceChanges extends an install or remove Plan (already built by
// BuildPlan) with voice changes, per design.md "Operaciones", restricted to
// p.Config.Scope == "user" (voice never applies to project scope — see
// proposal.md). It walks p.Changes rather than every registered voice path,
// so it only ever touches files this specific plan's p.Hosts already
// resolved, and it checks every touched block path for an unregistered
// voice block regardless of whether a voice is even active, before
// generating anything. It only reads; BuildPlan calls it before p.ID is
// computed.
func addVoiceChanges(p *Plan, o Options, state State) error {
	if p.Config.Scope != "user" {
		return nil
	}
	switch p.Action {
	case "install":
		return addVoiceChangesForInstall(p, o, state)
	case "remove":
		return addVoiceChangesForRemove(p, state)
	default:
		return nil
	}
}

// addVoiceChangesForInstall first rejects an unregistered voice block on any
// path this install touches, then, if a voice is active and the source
// carries content/voices/ (not a --release install, which carries no voice
// source on disk), regenerates a voice span whose freshly rendered text no
// longer matches the stored one for every path in p.Changes, including a
// path gaining its Hive block for the very first time in this same plan
// (Before nil, After rendered, Expected equal to that Change's Expected —
// e.g. a host newly added while a voice is already active). A missing span
// counts as differing; the user's choice is not part of RenderVoice's
// source hash, so the full rendered bytes are compared, not just the hash.
// A voice that cannot be rendered at all (its source file renamed or
// removed, or a reserved marker in its text) does not fail the plan: it
// leaves every stored span untouched and records p.VoiceWarning instead.
func addVoiceChangesForInstall(p *Plan, o Options, state State) error {
	for _, ch := range p.Changes {
		if ch.Target.Kind != "block" || ch.After == nil {
			continue
		}
		s, err := overlayRead(*p, ch.Target, ch.Replaces != nil)
		if err != nil {
			return err
		}
		existing, hasSpan := state.VoiceSpans[ch.Target.Path]
		if _, err := checkVoiceConflict(ch.Target.Path, s, hasSpan, existing); err != nil {
			return err
		}
	}
	// Accepted limit: with no active voice, a pinned release (no voice source
	// on disk) or a voice that cannot be rendered, nothing regenerates the
	// voice text, so a voice block deleted by hand is not restored here; its
	// span stays registered and status shows it as drift.
	if state.Voice == nil || o.ReleaseID != "" {
		return nil
	}
	if info, err := os.Stat(filepath.Join(o.Source, VoicesSource)); err != nil || !info.IsDir() {
		return nil
	}
	body, srcHash, err := RenderVoice(o.Source, *state.Voice)
	if err != nil {
		// The active voice can no longer be rendered (its file renamed or
		// removed, or a reserved marker introduced into its text): this must
		// not block install/update. Leave every stored span untouched and
		// surface a warning instead of failing.
		p.VoiceWarning = fmt.Sprintf("Voice %q not found or invalid in this source; voice files left unchanged (%v).", state.Voice.ID, err)
		return nil
	}
	for _, ch := range p.Changes {
		if ch.Target.Kind != "block" || ch.After == nil {
			continue
		}
		s, err := overlayRead(*p, ch.Target, ch.Replaces != nil)
		if err != nil {
			return err
		}
		existing, hasSpan := state.VoiceSpans[ch.Target.Path]
		gone, err := checkVoiceConflict(ch.Target.Path, s, hasSpan, existing)
		if err != nil {
			return err
		}
		managed := managedBlock(body, s.Data, voiceMarkers)
		newConsumers := filterUserConsumers(ch.After.Consumers, p.Config)
		// A shared file's consumer set can widen (a host joins) even when the
		// rendered bytes don't change; still record that, so status for the
		// newly joined host lists the voice row right away.
		if !gone && hasSpan && bytes.Equal(existing.Managed, managed) && reflect.DeepEqual(existing.Consumers, newConsumers) {
			continue
		}
		var before *VoiceSpan
		if hasSpan {
			span := existing
			before = &span
		}
		after := &VoiceSpan{Managed: managed, SourceHash: srcHash, Consumers: newConsumers}
		p.Voice = append(p.Voice, VoiceChange{Path: ch.Target.Path, Consumers: after.Consumers, Expected: ch.Expected, Before: before, After: after, Gone: gone})
	}
	return nil
}

// addVoiceChangesForRemove first rejects an unregistered voice block on any
// path this remove touches, then drops the voice span of a path whose last
// Hive-block consumer this plan retires, and narrows (rather than drops)
// the span's registered Consumers on a partial retirement of a shared
// file, so status for the retired host stops listing that voice row even
// though the file and its voice text are otherwise untouched.
func addVoiceChangesForRemove(p *Plan, state State) error {
	for _, ch := range p.Changes {
		if ch.Target.Kind != "block" {
			continue
		}
		s, err := overlayRead(*p, ch.Target, false)
		if err != nil {
			return err
		}
		existing, hasSpan := state.VoiceSpans[ch.Target.Path]
		gone, err := checkVoiceConflict(ch.Target.Path, s, hasSpan, existing)
		if err != nil {
			return err
		}
		if !hasSpan {
			continue
		}
		// A span deleted by hand is dropped without writing, also when other
		// hosts still use the file: the next install regenerates it.
		if ch.After == nil || gone {
			span := existing
			p.Voice = append(p.Voice, VoiceChange{Path: ch.Target.Path, Consumers: existing.Consumers, Expected: ch.Expected, Before: &span, After: nil, Gone: gone})
			continue
		}
		narrowed := filterUserConsumers(ch.After.Consumers, p.Config)
		if reflect.DeepEqual(narrowed, existing.Consumers) {
			continue
		}
		before, after := existing, existing
		after.Consumers = narrowed
		p.Voice = append(p.Voice, VoiceChange{Path: ch.Target.Path, Consumers: narrowed, Expected: ch.Expected, Before: &before, After: &after})
	}
	return nil
}
