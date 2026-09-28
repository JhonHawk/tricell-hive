// tui_screens.go implements the Status, Update, Releases and Voice screens
// (design.md "La interfaz"). Status is read-only; Update, Releases and Voice
// each reuse an existing plan/apply core (updateWith, BuildPlan("install",
// ...) for a rollback, BuildVoicePlan+voicePlanWith) through the huh
// prompter, the same "separar las preguntas de la lógica" shape tui_hosts.go
// already follows for Install/Remove.
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"tricell-hive/tooling/management"
)

// statusScreen implements Status (design.md "La interfaz"): one row per
// registered user-scope host, read-only. It takes p only for signature
// symmetry with the other screens in runMenuEntry's dispatch table; it never
// prompts.
func statusScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Status ==")
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		fmt.Fprintln(out, "No CLI hosts are registered. Choose Install CLIs.")
		return nil
	}
	so := o
	so.Hosts = hosts
	entries, err := management.Status(so)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		fmt.Fprintln(out, statusLine(host, entries))
	}
	return nil
}

// statusLine renders one host's own row: its release (short ID), product
// version, how many of its resources are in drift, and the voice active for
// it, each "-" or "off" when unknown, so the row's shape never depends on
// which fields happen to be populated (design.md "La interfaz": "la release
// (ID corto), la versión del producto, cuántos recursos están en drift y la
// voz activa"). A voice row's own Host is always empty (management.Status);
// its Consumers list which hosts render it, so membership there is what
// attributes it to this host, not StatusEntry.Host.
func statusLine(host string, entries []management.StatusEntry) string {
	release, version, voice := "", "", ""
	drift := 0
	for _, e := range entries {
		if e.Kind == "voice" {
			for _, c := range e.Consumers {
				if c.Host == host && e.Voice != "" {
					voice = e.Voice
				}
			}
			continue
		}
		if e.Host != host {
			continue
		}
		if e.Release != "" {
			release = e.Release
		}
		if e.ProductVersion != "" {
			version = e.ProductVersion
		}
		if e.Status == "drift" {
			drift++
		}
	}
	if release == "" {
		release = "-"
	} else {
		release = shortHash(release)
	}
	if version == "" {
		version = "-"
	}
	if voice == "" {
		voice = "off"
	}
	return fmt.Sprintf("%s · release %s · version %s · drift %d · voice %s", host, release, version, drift, voice)
}

// updateScreen implements Update (design.md "La interfaz"): the source and
// revision Inputs (default "." and "HEAD"), a "Resolving <rev>…" notice
// before the potentially slow Git/extraction work updateWith does, and
// updateWith itself with this session's own huh prompter driving its
// Confirm. interactive is always true here: the huh prompter can always
// confirm, in a real terminal or in accessible mode, unlike
// installTerminal's own interactive flag (which reflects whether stdin/
// stdout are a real terminal at all).
func updateScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Update ==")
	source, rev, ok, err := p.SourceAndRevision()
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	fmt.Fprintf(out, "Resolving %s…\n", rev)
	f := updateFlags{Source: source, Rev: rev, Home: o.Home, StateDir: o.StateDir}
	// mentionDryRunFlag is false: this path has no --dry-run flag of its own
	// to suggest (T4 fix round item 4), unlike hive update's own call in
	// update.go's update(), which keeps it true unchanged.
	return updateWith(f, out, true, p, false)
}

// releasesScreen implements Releases (design.md "La interfaz"): the
// filterable, fixed-height Select of retained releases, newest first
// (management.Releases's own order), with the currently installed one
// marked. Choosing it prints "Already installed"; choosing any other one
// starts rollbackFlow.
func releasesScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Releases ==")
	entries, err := management.Releases(o)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Fprintln(out, "No releases are retained yet")
		return nil
	}
	installedID, err := installedReleaseID(o)
	if err != nil {
		return err
	}
	chosen, ok, err := p.SelectRelease(entries, installedID)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	if installedID != "" && chosen == installedID {
		fmt.Fprintln(out, "Already installed")
		return nil
	}
	return rollbackFlow(o, chosen, out, p)
}

// installedReleaseID names the release ID currently installed for the
// registered user-scope hosts, reusing tui.go's own latestRelease the menu's
// status line already relies on. It returns "" (not an error) when no host
// is registered yet: Releases still lists retained snapshots in that case,
// it just cannot mark one "installed".
func installedReleaseID(o management.Options) (string, error) {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return "", err
	}
	if len(hosts) == 0 {
		return "", nil
	}
	so := o
	so.Hosts = hosts
	entries, err := management.Status(so)
	if err != nil {
		return "", err
	}
	id, _ := latestRelease(entries)
	return id, nil
}

// rollbackFlow builds a "install" plan pinned to releaseID for the
// registered hosts, shows the same showInstallSummary preview Install CLIs
// uses (including its voice warning) and confirms before applying
// (design.md "La interfaz", Releases row). A release the current manager
// cannot validate (for example a retired agent frontmatter field —
// _support/docs/architecture/deployment-manager.md's own "unsupported agent
// field" example) fails BuildPlan with one of releaseValidationErrorPrefixes'
// own messages; only then is the error wrapped with the documented downgrade
// hint before returning, so runMenu's own generic "print the error and go
// back to the menu" (AC10) shows both without this flow ever needing its own
// special-cased error branch. An unrelated BuildPlan failure (T4 fix round
// item 3: for example "explicit hosts required" when nothing is registered,
// or "unfinished operation: recover first") is returned unwrapped: the hint
// only makes sense for a problem with releaseID's own retained snapshot.
func rollbackFlow(o management.Options, releaseID string, out io.Writer, p prompter) error {
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return err
	}
	ro := o
	ro.Hosts = hosts
	ro.ReleaseID = releaseID
	plan, err := management.BuildPlan("install", ro)
	if err != nil {
		if isReleaseValidationError(err) {
			return fmt.Errorf("%w; plan that downgrade with the manager from the commit that produced it", err)
		}
		return err
	}
	unchanged, err := management.PlanUnchanged(plan)
	if err != nil {
		return err
	}
	// mentionDryRunFlag is false, the same as installScreen's own call
	// (install.go): this path has no --dry-run flag of its own to suggest.
	showInstallSummary(out, plan, onboardingPreview{}, false, unchanged, false)
	decision, err := p.Confirm("Apply these changes?", false)
	if err != nil {
		return err
	}
	if decision != installApply {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	result, err := (management.Engine{}).Apply(plan)
	if err != nil {
		return err
	}
	reportApplyResult(out, "Hive rolled back", result)
	return nil
}

// releaseValidationErrorPrefixes are every error text tooling/management's
// own loadRelease/validateRelease (plan.go) can produce while checking a
// pinned ReleaseID's own retained snapshot — the closed set read directly
// from that function's source, current as of this change. rollbackFlow's
// own downgrade hint (T4 fix round item 3) applies only when a BuildPlan
// failure starts with one of these, never to an unrelated failure such as
// "explicit hosts required" or "unfinished operation: recover first", which
// loadRelease is never reached to produce.
var releaseValidationErrorPrefixes = []string{
	"invalid release ID",
	"invalid release fingerprint or file list",
	"invalid source ",
	"missing skill entrypoint:",
	"duplicate agent name:",
	"agent release missing renderer",
	"unsupported agent field ",
	"unsupported agent renderer",
	"reserved source delimiter",
	"nonportable personal path in ",
}

// isReleaseValidationError reports whether err's own message starts with one
// of releaseValidationErrorPrefixes.
func isReleaseValidationError(err error) bool {
	msg := err.Error()
	for _, prefix := range releaseValidationErrorPrefixes {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

// formatReleaseLabel renders one release's own Select option label at or
// under releaseLabelWidth columns (design.md "La interfaz": "etiquetas de 78
// columnas o menos"): its short ID, date, first commit (or "-" when the
// release predates commit tracking) and its consuming hosts, truncated to
// whatever room is left, with an "(installed)" marker for the one currently
// applied.
const releaseLabelWidth = 78

func formatReleaseLabel(e management.ReleaseEntry, installed bool) string {
	shortID := shortHash(e.ID)
	date := e.LastWrittenAt
	if len(date) > 10 {
		date = date[:10]
	}
	commit := "-"
	if len(e.Commits) > 0 {
		commit = shortHash(e.Commits[0])
	}
	marker := ""
	if installed {
		marker = " (installed)"
	}
	prefix := fmt.Sprintf("%s  %s  %s  ", shortID, date, commit)
	budget := releaseLabelWidth - utf8.RuneCountInString(prefix) - utf8.RuneCountInString(marker)
	if budget < 0 {
		budget = 0
	}
	return prefix + truncateJoined(releaseHostNames(e.Consumers), budget) + marker
}

// releaseHostNames is a release's own unique, sorted consumer host names.
func releaseHostNames(consumers []management.Consumer) []string {
	seen := map[string]bool{}
	var names []string
	for _, c := range consumers {
		if !seen[c.Host] {
			seen[c.Host] = true
			names = append(names, c.Host)
		}
	}
	sort.Strings(names)
	return names
}

// truncateJoined joins names with ", " and truncates the result to width
// runes via truncateRunes (never splitting a name's own first character off
// with nothing to show for it).
func truncateJoined(names []string, width int) string {
	return truncateRunes(strings.Join(names, ", "), width)
}

// truncateRunes returns s unchanged if it fits within width runes,
// otherwise truncates it to width RUNES with a trailing "…". It counts and
// slices by rune, not by byte: "…" (U+2026) is three UTF-8 bytes but one
// terminal column, and every caller's own width here is a column budget —
// slicing the underlying string by byte index would silently make a long
// label two bytes (not columns) over budget.
func truncateRunes(s string, width int) string {
	if utf8.RuneCountInString(s) <= width {
		return s
	}
	if width <= 1 {
		return strings.Repeat(".", max(width, 0))
	}
	return string([]rune(s)[:width-1]) + "…"
}

// voiceScreen implements Voice (design.md "La interfaz"): show the active
// voice, offer ListVoices' own catalogue plus Off, then (for a chosen voice)
// its address/name/intensity, before voicePlanWith applies the result with
// this session's own huh prompter — the same core hive voice set/off use.
func voiceScreen(o management.Options, out io.Writer, p *huhPrompter) error {
	fmt.Fprintln(out, "== Voice ==")
	hosts, err := management.RegisteredHosts(o)
	if err != nil {
		return err
	}
	if len(hosts) == 0 {
		fmt.Fprintln(out, "No CLI hosts are registered. Choose Install CLIs.")
		return nil
	}
	so := o
	so.Hosts = hosts
	entries, err := management.Status(so)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "Active voice: %s\n", activeVoiceLine(entries))

	voices, err := management.ListVoices(o.Source)
	if err != nil {
		return err
	}
	choice, ok, err := p.SelectVoiceOrOff(voices)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	if choice == "" {
		plan, err := management.BuildVoicePlan("off", o, management.VoiceSetting{})
		if err != nil {
			return err
		}
		return voicePlanWith(plan, "Voice turned off", out, true, false, "", p)
	}
	address, name, intensity, ok, err := p.VoiceDetails()
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(out, "Cancelled. No changes applied.")
		return nil
	}
	v := management.VoiceSetting{ID: choice, Address: address, Name: name, Intensity: intensity}
	plan, err := management.BuildVoicePlan("set", o, v)
	if err != nil {
		return err
	}
	return voicePlanWith(plan, "Voice set", out, true, false, "", p)
}

// voiceLabelWidth bounds voiceOptionLabel to one rendered line (T4 fix
// round item 1): huh's own Select wraps an option whose rendered width
// exceeds the field's width, and fieldHeight's own n+1 model assumes
// exactly one line per option. 76 leaves two columns of margin inside an
// 80-column terminal for the field's own cursor/indent.
const voiceLabelWidth = 76

// voiceOptionLabel renders one ListVoices entry as "id — description",
// truncating the description (never the ID) so the whole label fits within
// voiceLabelWidth columns. Before this, a full-length description (real
// voices like jarvis or mentor carry one- or two-sentence descriptions) could
// wrap to two or more rendered lines; fieldHeight(len(options)) only ever
// reserves one line per option, so a long description pushed later options
// out of the viewport entirely — confirmed directly: at width 80, jarvis and
// mentor's own full descriptions wrapped, hiding both from a fresh 80x24
// render along with everything after them.
func voiceOptionLabel(v management.VoiceInfo) string {
	prefix := v.ID + " — "
	budget := voiceLabelWidth - utf8.RuneCountInString(prefix)
	if budget < 0 {
		budget = 0
	}
	return prefix + truncateRunes(voiceDisplayDescription(v.Description), budget)
}

// activeVoiceLine is Voice's own "show the active voice" line: the same
// "id (address, intensity)" text Status's voice row already carries
// (management.formatVoiceStatus), or "off" when no voice entry is active.
func activeVoiceLine(entries []management.StatusEntry) string {
	for _, e := range entries {
		if e.Kind == "voice" && e.Voice != "" {
			return e.Voice
		}
	}
	return "off"
}
