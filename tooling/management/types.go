// Package management owns deterministic file delivery. It never launches models.
package management

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"tricell-hive/integrations/agents"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/legacy"
)

const Begin = "<!-- === TRICELL HIVE RULES:BEGIN === -->"
const End = "<!-- === TRICELL HIVE RULES:END === -->"

// VoiceBegin and VoiceEnd delimit the optional voice block, a second managed
// span in the same instruction file, always placed after the Hive block. See
// design.md "Voice block" (archived heading "Bloque de voz").
const VoiceBegin = "<!-- === TRICELL HIVE VOICE:BEGIN === -->"
const VoiceEnd = "<!-- === TRICELL HIVE VOICE:END === -->"
const GlobalSource = "content/guidance/global.md"
const SkillSource = "content/skills/workspace-conventions/SKILL.md"

// VoicesSource is the catalogue directory holding the shared preamble and
// each voice's description, read directly from the source checkout by
// ListVoices and RenderVoice rather than through a Release payload.
const VoicesSource = "content/voices"

// markers is a begin/end delimiter pair for one managed span, plus the short
// name that error messages use to say which block failed ("Hive" or
// "voice"). blockRange, managedBlock, owned and transform all take one
// explicitly so the same block mechanics serve both the Hive block
// (hiveMarkers) and the voice block (voiceMarkers), without behavior
// differing between the two pairs; hiveMarkers keeps every message byte for
// byte identical to the pre-parametrization text.
type markers struct{ begin, end, name string }

var hiveMarkers = markers{begin: Begin, end: End, name: "Hive"}
var voiceMarkers = markers{begin: VoiceBegin, end: VoiceEnd, name: "voice"}

// VoiceSetting is the user's chosen optional voice layer: which voice, how it
// addresses the user, and how strongly it shows. Not yet wired into Plan or
// State; a later change adds Plan.VoiceSetting and State.Voice.
type VoiceSetting struct {
	ID        string `json:"id"`
	Address   string `json:"address"`
	Name      string `json:"name,omitempty"`
	Intensity string `json:"intensity"`
}

// ModelOverride is one role's model and effort on one CLI, chosen by the user
// and stored in State.ModelOverrides. An empty field keeps the release's value.
type ModelOverride = agents.ModelOverride

// VoiceSpan is one rendered voice block as installed in a file: its managed
// text, the hash of the sources RenderVoice rendered it from, and its
// registered consumers.
type VoiceSpan struct {
	Managed    []byte     `json:"managed"`
	SourceHash string     `json:"source_hash"`
	Consumers  []Consumer `json:"consumers,omitempty"`
}

// VoiceChange is one file's voice-block change within a Plan, parallel to
// Change for the Hive block but addressed by Path rather than by a resolved
// target.Target: voice paths come from the block Records of already
// registered hosts (see design.md "Voice paths"
// (archived heading "Rutas de la voz")), not from resolve(). It carries only the
// managed span, never the whole file, so a plan with voice changes can still be
// saved with --out (see SavePlan). Expected is the whole file's fingerprint
// before this plan's changes, exactly like Change.Expected, checked at apply time
// even when no Change touches the same path in the same plan.
type VoiceChange struct {
	Path      string      `json:"path"`
	Consumers []Consumer  `json:"consumers,omitempty"`
	Expected  Fingerprint `json:"expected"`
	Before    *VoiceSpan  `json:"before,omitempty"`
	After     *VoiceSpan  `json:"after,omitempty"`
	// Gone is set when the voice block was deleted by hand; see Change.Gone.
	Gone bool `json:"gone,omitempty"`
}

type Payload struct {
	Path string `json:"path"`
	Data []byte `json:"data"`
	Mode uint32 `json:"mode,omitempty"`
}
type Release struct {
	ID       string    `json:"id"`
	Files    []Payload `json:"files"`
	Profiles []byte    `json:"profiles,omitempty"`
	Renderer string    `json:"renderer,omitempty"`
}
type Record struct {
	Target      target.Target `json:"target"`
	Managed     []byte        `json:"managed"`
	Mode        uint32        `json:"mode,omitempty"`
	Leading     string        `json:"leading,omitempty"`
	CreatedFile bool          `json:"created_file"`
	Release     string        `json:"release"`
	Consumers   []Consumer    `json:"consumers,omitempty"`
}
type Consumer struct {
	Host    string `json:"host"`
	Scope   string `json:"scope"`
	Context string `json:"context"`
}
type State struct {
	Versions      map[string]ProductIdentity     `json:"versions,omitempty"`
	Installations map[string]InstallationReceipt `json:"installations,omitempty"`
	Version       int                            `json:"version"`
	Records       map[string]Record              `json:"records"`
	CreatedDirs   []string                       `json:"created_dirs,omitempty"`
	Migrations    []MigrationReceipt             `json:"migrations,omitempty"`
	// Voice and VoiceSpans hold the optional voice layer (D13-A: schema stays
	// 6; a state with no voice set encodes identically to before these
	// fields existed). Voice is the user's current choice, or nil when no
	// voice is active. VoiceSpans holds, per installed instruction file path,
	// the rendered span currently written there.
	Voice      *VoiceSetting        `json:"voice,omitempty"`
	VoiceSpans map[string]VoiceSpan `json:"voice_spans,omitempty"`
	// ModelOverrides holds the user's per-role model and effort, keyed CLI then
	// role, for the user scope. Schema stays 6: a state with no override encodes
	// as before, because the field is omitted when empty and kept nil then.
	ModelOverrides map[string]map[string]ModelOverride `json:"model_overrides,omitempty"`
	// PiSubagentsSource is the exact npm source Hive installed, or empty when
	// Hive did not add the package. Schema stays 6 via omitempty.
	PiSubagentsSource string `json:"pi_subagents_source,omitempty"`
}
type Fingerprint struct {
	TreeHash   string `json:"tree_hash,omitempty"`
	Exists     bool   `json:"exists"`
	Hash       string `json:"hash"`
	Mode       uint32 `json:"mode"`
	Kind       string `json:"kind,omitempty"`
	LinkTarget string `json:"link_target,omitempty"`
	FileMode   uint32 `json:"file_mode,omitempty"`
}
type Change struct {
	Target   target.Target `json:"target"`
	Expected Fingerprint   `json:"expected"`
	Before   *Record       `json:"before,omitempty"`
	After    *Record       `json:"after,omitempty"`
	Replaces *Record       `json:"replaces,omitempty"`
	// Gone is set when the managed file or block this change would touch was
	// deleted by hand (owned reports ManagedFileMissing or ManagedBlockMissing).
	// Install and update write it again as a fresh installation would; remove
	// writes nothing and drops the record. Apply re-verifies it against the
	// current snapshot and fails closed when the path is no longer missing.
	Gone bool `json:"gone,omitempty"`
}
type Plan struct {
	Product      *ProductIdentity  `json:"product,omitempty"`
	Version      int               `json:"version"`
	Action       string            `json:"action"`
	Config       target.Config     `json:"config"`
	Hosts        []string          `json:"hosts"`
	StateDir     string            `json:"state_dir"`
	StateHash    string            `json:"state_hash"`
	Release      *Release          `json:"release,omitempty"`
	Changes      []Change          `json:"changes"`
	ID           string            `json:"id"`
	Legacy       []legacy.Edit     `json:"legacy,omitempty"`
	Migration    *MigrationReceipt `json:"migration,omitempty"`
	SourceCommit string            `json:"source_commit,omitempty"`
	// Voice carries this plan's voice-block changes, addressed by Path (see
	// VoiceChange). VoiceSetting is the choice apply will write to
	// State.Voice, consulted only when Action is "voice": non-nil for
	// "voice set", nil for "voice off". install/remove plans always leave
	// VoiceSetting nil; State.Voice then only changes automatically, when a
	// remove plan empties State.VoiceSpans (see prepareTransaction).
	Voice        []VoiceChange `json:"voice,omitempty"`
	VoiceSetting *VoiceSetting `json:"voice_setting,omitempty"`
	// VoiceWarning is set on an install plan when an active voice could not
	// be rendered from this source (its file renamed or removed, or a
	// reserved marker in its text) — install still proceeds, keeping the
	// stored spans untouched, and a caller displays this line rather than
	// failing the whole operation.
	VoiceWarning string `json:"voice_warning,omitempty"`
	// VoiceSkipped lists, on a "voice set" plan, the instruction files that
	// were left out because they no longer hold a managed Hive block (deleted
	// by hand). Nothing is written to them; hive install restores the block
	// first. It is informational, so a caller can tell the user.
	VoiceSkipped []string `json:"voice_skipped,omitempty"`
	// ModelOverrides is, on a user-scope install plan, the complete set of
	// overrides the state holds after the plan: the plan's CLIs carry their new
	// set and every other CLI carries what the state already has. Remove, voice
	// and project-scope plans carry none. Apply stores the plan's CLIs' sets.
	ModelOverrides map[string]map[string]ModelOverride `json:"model_overrides,omitempty"`
	// PiPackage is the pi-subagents step for a user-scope install that includes
	// Pi. It is not a file Change: Pi rewrites settings.json itself.
	PiPackage *PiPackageStep `json:"pi_package,omitempty"`
}

// PackageAction is the planned pi-subagents operation.
type PackageAction string

const (
	PackageInstall PackageAction = "install"
	PackageOmit    PackageAction = "omit"
)

// PiPackageStep is the planned declaration of pi-subagents in Pi user settings.
type PiPackageStep struct {
	Action      PackageAction `json:"action"`
	Source      string        `json:"source,omitempty"`
	AddedByHive bool          `json:"added_by_hive,omitempty"`
}

// ReleaseEntry describes one retained release snapshot for hive releases: its
// identity, when its snapshot was last written, the source commits it is
// known to have come from, and where it is currently installed.
type ReleaseEntry struct {
	ID            string     `json:"id"`
	LastWrittenAt string     `json:"last_written_at"`
	Commits       []string   `json:"commits"`
	Consumers     []Consumer `json:"consumers"`
}

// commitRecord is the on-disk shape of releases/<id>.commits.json: the source
// commits a release is known to have been produced from, each with when it
// was applied. It is separate from the release snapshot itself so recording
// a commit never touches the snapshot's compared-byte-for-byte content.
type commitRecord struct {
	Commits []commitLogEntry `json:"commits"`
}
type commitLogEntry struct {
	Commit    string `json:"commit"`
	AppliedAt string `json:"applied_at"`
}
type Options struct {
	Scope, Home, Root, StateDir, Source, ReleaseID string
	Hosts                                          []string
}
type StatusEntry struct {
	Path, Host, Kind, Status, Release string
	ProductVersion, VersionStatus     string
	Consumers                         []Consumer
	// Voice is set only on a Kind "voice" row: the voice ID, address and
	// intensity, e.g. "jarvis (sir, subtle)" (see design.md "status").
	Voice string `json:"Voice,omitempty"`
}

func hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func encode(v any) []byte {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(b, '\n')
}

// V3 releases were addressed by their file list alone. Keep that identifier
// stable so existing release snapshots remain selectable after the upgrade.
func releaseID(r Release) string {
	if r.Renderer == "" && len(r.Profiles) == 0 {
		return hash(encode(r.Files))
	}
	return hash(encode(struct {
		Files    []Payload `json:"files"`
		Profiles []byte    `json:"profiles"`
		Renderer string    `json:"renderer"`
	}{r.Files, r.Profiles, r.Renderer}))
}
func planID(p Plan) string { p.ID = ""; return hash(encode(p)) }
func emptyState() State    { return State{Version: stateVersion, Records: map[string]Record{}} }
func normalize(o Options) (target.Config, string, error) {
	if o.Scope != "user" && o.Scope != "project" {
		return target.Config{}, "", fmt.Errorf("explicit scope must be user or project")
	}
	synthetic := o.Home != ""
	if o.Home == "" {
		var err error
		o.Home, err = os.UserHomeDir()
		if err != nil {
			return target.Config{}, "", err
		}
	}
	home, err := target.Canonical(o.Home)
	if err != nil {
		return target.Config{}, "", err
	}
	c := target.Config{Scope: o.Scope, Home: home, CodexHome: filepath.Join(home, ".codex"), ClaudeHome: filepath.Join(home, ".claude")}
	if !synthetic {
		if v := os.Getenv("CODEX_HOME"); v != "" {
			c.CodexHome = v
		}
		if v := os.Getenv("CLAUDE_CONFIG_DIR"); v != "" {
			c.ClaudeHome = v
		}
	}
	c.CodexHome, err = target.Canonical(c.CodexHome)
	if err != nil {
		return c, "", err
	}
	c.ClaudeHome, err = target.Canonical(c.ClaudeHome)
	if err != nil {
		return c, "", err
	}
	c, err = target.ExpandHostHomes(c, synthetic)
	if err != nil {
		return c, "", err
	}
	if o.Scope == "project" {
		if o.Root == "" {
			return c, "", fmt.Errorf("project scope requires --root")
		}
		c.Root, err = target.Canonical(o.Root)
		if err != nil {
			return c, "", err
		}
	}
	if o.StateDir == "" {
		o.StateDir, err = DefaultStateDir(home, synthetic)
		if err != nil {
			return c, "", err
		}
	}
	state, err := target.Canonical(o.StateDir)
	if err != nil {
		return c, "", err
	}
	return c, state, target.Safe(state)
}
func validateHosts(hosts []string) ([]string, error) {
	h := append([]string(nil), hosts...)
	sort.Strings(h)
	if len(h) == 0 {
		return nil, fmt.Errorf("explicit hosts required")
	}
	for i, s := range h {
		if s != "codex" && s != "claude" && s != "grok" && s != "pi" && s != "opencode" && s != "cursor" {
			return nil, fmt.Errorf("unsupported host %q", s)
		}
		if i > 0 && s == h[i-1] {
			return nil, fmt.Errorf("duplicate host %q", s)
		}
	}
	return h, nil
}

// NormalizeOptions resolves roots without writing or using real-user overrides for explicit homes.
func NormalizeOptions(o Options) (target.Config, string, error) { return normalize(o) }
func DefaultStateDir(home string, synthetic bool) (string, error) {
	return defaultStateDir(home, synthetic, runtime.GOOS)
}
func defaultStateDir(home string, synthetic bool, platform string) (string, error) {
	old := filepath.Join(home, "Library", "Application Support", "tricell-hive")
	if platform != "linux" {
		return old, nil
	}
	base := filepath.Join(home, ".local", "state")
	if !synthetic && os.Getenv("XDG_STATE_HOME") != "" {
		base = os.Getenv("XDG_STATE_HOME")
		if !filepath.IsAbs(base) {
			return "", fmt.Errorf("XDG_STATE_HOME must be absolute")
		}
	}
	next := filepath.Join(base, "tricell-hive")
	exists := func(p string) (bool, error) {
		_, err := os.Lstat(p)
		if os.IsNotExist(err) {
			return false, nil
		}
		return err == nil, err
	}
	a, err := exists(old)
	if err != nil {
		return "", err
	}
	b, err := exists(next)
	if err != nil {
		return "", err
	}
	if a && b && old != next {
		return "", fmt.Errorf("multiple Hive state directories; select --state-dir explicitly")
	}
	if a {
		return old, nil
	}
	return next, nil
}
func RegisteredHosts(o Options) ([]string, error) {
	c, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	s, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, r := range s.Records {
		for _, v := range r.Consumers {
			if v.Scope == c.Scope && (v.Context == c.Home || (c.Scope == "project" && v.Context == c.Root)) {
				set[v.Host] = true
			}
		}
	}
	var out []string
	for h := range set {
		out = append(out, h)
	}
	sort.Strings(out)
	return out, nil
}
