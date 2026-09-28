// tui_prompter.go implements the interactive interface's own prompter with
// charm.land/huh/v2 forms (design.md "El `prompter` de `huh`").
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"tricell-hive/tooling/management"
)

// oneByteReader reads at most one byte per Read call from its underlying
// reader and latches whether it has seen the end of input. In accessible
// mode, huh.Form.RunAccessible builds a fresh bufio.Scanner for every field
// (design.md "Contexto verificado"): a scanner that reads ahead in bigger
// chunks would silently consume the next field's own answer along with the
// current one. huh's accessible runner also never reports EOF or
// cancellation itself — every field's own error is discarded — so this
// reader is the only place the interface can observe it. Exactly one
// instance wraps the session's real input; no other bufio.Reader ever reads
// it directly.
type oneByteReader struct {
	r   io.Reader
	eof bool
}

func (o *oneByteReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err := o.r.Read(p[:1])
	if errors.Is(err, io.EOF) {
		o.eof = true
	}
	return n, err
}

// huhPrompter implements prompter with charm.land/huh/v2 forms: the
// interactive interface's own question layer. In accessible mode, in is
// always the session's single *oneByteReader; in the normal terminal mode,
// theme holds the once-resolved Charm or base theme (design.md "El
// `prompter` de `huh`").
type huhPrompter struct {
	in         io.Reader
	out        io.Writer
	accessible bool
	theme      huh.Theme
	eofReader  *oneByteReader
}

// newHuhPrompter builds the interface's own prompter. In accessible mode it
// wraps in in a oneByteReader and never resolves a theme, since huh's
// accessible runner is plain text; otherwise it resolves the theme once,
// preferring ThemeBase when NO_COLOR is set and otherwise
// charmThemeForDetectedBackground with the background
// lipgloss.HasDarkBackground (charm.land/lipgloss/v2 v2.0.1, the version
// this module's go.sum pins) detects — not ThemeCharm directly, which would
// render unselected options unreadable on a real dark background (T6 fix
// round F1; see charmThemeForDetectedBackground's own doc comment).
func newHuhPrompter(accessible bool, in io.Reader, out io.Writer) *huhPrompter {
	p := &huhPrompter{out: out, accessible: accessible}
	if accessible {
		p.eofReader = &oneByteReader{r: in}
		p.in = p.eofReader
		// T2 fix round item 6: accessible mode is plain text for screen
		// readers and scripted input. huh's own default theme (ThemeCharm)
		// still colors field titles with ANSI SGR codes even though
		// RunAccessible never renders huh's full-screen view — lipgloss's
		// color-profile detection is not scoped to the writer actually
		// passed in here. Pin ThemeBase, whose Title/Description styles
		// carry no color (theme.go), unconditionally: not only when
		// NO_COLOR is set.
		p.theme = huh.ThemeFunc(huh.ThemeBase)
		return p
	}
	p.in = in
	if os.Getenv("NO_COLOR") != "" {
		p.theme = huh.ThemeFunc(huh.ThemeBase)
		return p
	}
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
	styles := charmThemeForDetectedBackground(isDark)
	p.theme = huh.ThemeFunc(func(bool) *huh.Styles { return styles })
	return p
}

// charmThemeForDetectedBackground is newHuhPrompter's own real-terminal
// theme choice (T6 fix round F1, a blocker; N1 follow-up). huh v2.0.3's
// ThemeCharm has a handful of its lightDark(...) branches backwards
// relative to what lipgloss v2.0.1's LightDark(isDark)(light, dark)
// (color.go:205, returns dark when isDark is true) actually resolves them
// to: theme.go:144's normalFg = lightDark(Color("252"), Color("235")) gives
// Focused.Option/UnselectedOption color 235 (~1.1:1) on an actually dark
// background instead of 252, and TextInput.Placeholder (238/248, same
// shape) is backwards the same way — both verified directly against the
// pinned versions. indigo (Title/Description/Directory), green
// (SelectedOption's own dark-branch value) and every other named color in
// ThemeCharm are already oriented correctly, so the fix here is NOT
// ThemeCharm(!isDark): an earlier round tried exactly that blanket flip and
// it broke indigo along with everything else already correct (N1) — Title
// on a real dark background dropped from ANSI 99 (5.1:1 against black) to
// 62 (4.1:1). Instead, take huh's own ThemeCharm(isDark) — correct for
// everything except the specific fields above — and override only those,
// with the branch each color was actually meant for (confirmed by direct
// WCAG contrast computation, TestThemeContrastMeetsWCAGAA): the lighter of
// the pair for a dark background, the darker for a light one. It also
// covers two fields ThemeCharm always under-contrasts regardless of isDark:
// SelectedOption/TextInput.Cursor's own light-branch green (#02BA84, only
// 2.51:1 against white) needs a darker green on light backgrounds; and
// FocusedButton's cream-on-fuchsia (N2, huh's own pre-existing defect,
// 2.26:1) needs a near-black foreground on the same fuchsia background
// (9.14:1), independent of isDark since the fuchsia itself is not
// lightDark-branched. TextInput.Text carries no explicit color in
// ThemeCharm at all (it inherits the terminal's own default, which this
// package cannot verify), so it is given the same explicit per-branch color
// as Option here, both for consistency and so it has a real, checkable
// value. Blurred and Group copy Focused's own fields by value inside
// ThemeCharm, before this function ever sees the result, so every override
// below is applied to Focused, Blurred and (for Description) Group
// separately — mutating styles.Focused afterward would not reach them.
func charmThemeForDetectedBackground(isDark bool) *huh.Styles {
	styles := huh.ThemeCharm(isDark)

	text := lipgloss.Color("235")
	placeholder := lipgloss.Color("238")
	description := lipgloss.Color("237")
	blurredButtonBG := lipgloss.Color("252")
	if isDark {
		text = lipgloss.Color("252")
		placeholder = lipgloss.Color("248")
		description = lipgloss.Color("250")
		blurredButtonBG = lipgloss.Color("237")
	}
	// N2: a near-black foreground on ThemeCharm's own fuchsia background,
	// replacing cream (#FFFDF5) — 9.14:1 against #F780E2, vs. cream's own
	// 2.26:1. The fuchsia itself is a plain lipgloss.Color, not
	// lightDark-branched, so one override serves both isDark branches.
	focusedButtonFG := lipgloss.Color("#1a1a1a")

	for _, fs := range []*huh.FieldStyles{&styles.Focused, &styles.Blurred} {
		fs.Option = fs.Option.Foreground(text)
		fs.UnselectedOption = fs.UnselectedOption.Foreground(text)
		fs.Description = fs.Description.Foreground(description)
		fs.TextInput.Text = fs.TextInput.Text.Foreground(text)
		fs.TextInput.Placeholder = fs.TextInput.Placeholder.Foreground(placeholder)
		fs.BlurredButton = fs.BlurredButton.Foreground(text).Background(blurredButtonBG)
		fs.FocusedButton = fs.FocusedButton.Foreground(focusedButtonFG)
		fs.Next = fs.Next.Foreground(focusedButtonFG)
	}
	styles.Group.Description = styles.Focused.Description

	if !isDark {
		// ThemeCharm's own green (#02BA84) is correctly the same value on
		// both branches (not lightDark-branched at all for SelectedOption),
		// but that one value only has enough contrast against a DARK
		// background (8.36:1); against a light one it is 2.51:1. Darkening
		// it only for the light branch (isDark's own dark-branch green
		// already passes unchanged) fixes both without touching the dark
		// branch's own already-compliant value.
		lightGreen := lipgloss.Color("#017A57")
		styles.Focused.SelectedOption = styles.Focused.SelectedOption.Foreground(lightGreen)
		styles.Focused.TextInput.Cursor = styles.Focused.TextInput.Cursor.Foreground(lightGreen)
		styles.Blurred.SelectedOption = styles.Blurred.SelectedOption.Foreground(lightGreen)
		styles.Blurred.TextInput.Cursor = styles.Blurred.TextInput.Cursor.Foreground(lightGreen)
	}

	return styles
}

// formKeyMap extends huh's own default keymap so Esc, not only Ctrl-C,
// cancels a form (AC10: "Ctrl-C o Esc"). huh v2.0.3 binds Quit to ctrl+c
// only (keymap.go's NewDefaultKeyMap); Form.Update checks key.Matches(msg,
// f.keymap.Quit) before the active field ever sees a key press
// (form.go's own Update, the tea.KeyPressMsg case), so replacing just this
// one binding is enough for every field type used here. One instance is
// shared by every form in the session; nothing here mutates it afterward.
var formKeyMap = func() *huh.KeyMap {
	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	return k
}()

// cancelledResult reports whether a form's outcome means the operator
// cancelled — Ctrl-C, Esc (huh.ErrUserAborted) or, in accessible mode, the
// end of input — which the prompter always turns into ok=false or
// installCancelled, never a propagated error (design.md "Cancelar").
func cancelledResult(err error, eof bool) bool {
	return eof || errors.Is(err, huh.ErrUserAborted)
}

// configureForm applies this prompter's own accessible/IO/theme/keymap
// settings to form, using the session's shared formKeyMap — the exact
// configuration runForm's own form.Run() then executes. Extracted into its
// own method so a test can verify the esc-quits-the-form wiring (T2 fix
// round item 2; T3 fix-round leftover (a): drive a real key.Msg through the
// configured form's Update) without driving a full interactive Bubble Tea
// program.
func (p *huhPrompter) configureForm(form *huh.Form) *huh.Form {
	return p.configureFormWithKeyMap(form, formKeyMap)
}

// configureFormWithKeyMap is configureForm's own core, taking an explicit
// keymap instead of the session's shared formKeyMap: releasesSelectKeyMap
// (T4's Esc-filter decision, tui_prompter.go) is the one caller that needs a
// different one.
func (p *huhPrompter) configureFormWithKeyMap(form *huh.Form, keymap *huh.KeyMap) *huh.Form {
	form = form.WithAccessible(p.accessible).WithInput(p.in).WithOutput(p.out).WithKeyMap(keymap)
	if p.theme != nil {
		form = form.WithTheme(p.theme)
	}
	return form
}

// runForm runs one single-group form through this prompter's own
// accessible/theme/IO settings and the session's shared formKeyMap,
// translating cancellation the same way for every field type (design.md
// "Cancelar"). runFormWithKeyMap is its own explicit-keymap counterpart.
func (p *huhPrompter) runForm(form *huh.Form) (cancelled bool, err error) {
	return p.runConfiguredForm(p.configureForm(form))
}

// runFormWithKeyMap is runForm's own counterpart for a field that needs a
// keymap other than the session's shared formKeyMap (releasesSelectKeyMap:
// T4's Esc-filter decision).
func (p *huhPrompter) runFormWithKeyMap(form *huh.Form, keymap *huh.KeyMap) (cancelled bool, err error) {
	return p.runConfiguredForm(p.configureFormWithKeyMap(form, keymap))
}

// runConfiguredForm is runForm's and runFormWithKeyMap's shared core, run
// once configureForm/configureFormWithKeyMap has already applied this
// prompter's own accessible/IO/theme/keymap settings to form.
//
// T2 fix round item 1: huh v2.0.3's accessible PromptString returns the
// last *invalid* answer, unfiltered, once real end-of-input follows it
// (internal/accessibility.PromptString: an invalid entry re-prompts, but at
// true EOF it falls back to that stale text instead of the field's default,
// since its own empty-string fallback only triggers when the leftover text
// is empty). Select and MultiSelect's own RunAccessible then index that
// text's parsed integer into their options slice with no bounds check
// (field_select.go's `s.options.val[choice-1]`), so an invalid-answer-then-
// EOF sequence such as "8", "0\n" or "abc\n" panics with an out-of-range
// index instead of ever returning to Go code. Since our own eofReader is
// the only thing that can observe genuine end-of-input, a panic that
// coincides with it is exactly this known huh defect, not a bug in this
// package: recover it and treat it as the same cancellation an EOF alone
// already produces. Any other panic is left to propagate — recovering
// unconditionally would silently hide a real programming error as
// "Cancelled".
func (p *huhPrompter) runConfiguredForm(form *huh.Form) (cancelled bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			if p.eofReader != nil && p.eofReader.eof {
				cancelled, err = true, nil
				return
			}
			panic(r)
		}
	}()
	runErr := form.Run()
	eof := p.eofReader != nil && p.eofReader.eof
	if cancelledResult(runErr, eof) {
		return true, nil
	}
	return false, runErr
}

// maxFieldHeight caps the height requested for a field listing many
// options, so a long list never asks for more than fits comfortably in a
// small terminal (T6's own review criteria use 80x24); nothing built in
// this package currently reaches it.
const maxFieldHeight = 15

// fieldHeight returns the height to request via .Height(...) for a
// single-line-title field listing n options, so its viewport shows every
// option without truncation (T3 fix round item 1). huh v2.0.3's
// MultiSelect.updateViewportSize subtracts the title's own rendered height
// from the viewport height it auto-computes from the options content when
// no explicit height is set (field_multiselect.go:495-514), silently
// hiding the last option — verified directly against the pinned version:
// with 6 hosts, "pi" never appears in .View()'s output. Select's own
// auto-sizing does not have this bug, but every field here still requests
// an explicit height for the same reason and the same n+1 shape, so a
// future change to any of them (a description line, say) cannot
// reintroduce it silently. n+1 compensates for exactly one title line.
func fieldHeight(n int) int {
	if n+1 > maxFieldHeight {
		return maxFieldHeight
	}
	return n + 1
}

// confirmBackSelectField builds Confirm's own three-option Select (Apply,
// Back, Cancel), defaulting to Cancel, with an explicit height so every
// option renders (T3 fix round item 1). Extracted so a test can render it
// directly with .View() without driving a full form.
func confirmBackSelectField(prompt string, value *installDecision) *huh.Select[installDecision] {
	options := []huh.Option[installDecision]{
		huh.NewOption("Apply", installApply),
		huh.NewOption("Back", installBack),
		huh.NewOption("Cancel", installCancelled),
	}
	return huh.NewSelect[installDecision]().
		Title(prompt).
		Options(options...).
		Value(value).
		Height(fieldHeight(len(options)))
}

// Confirm implements prompter.Confirm. Without allowBack it is a plain
// huh.Confirm defaulting to Cancel, exactly like confirmInstall's own
// empty-Enter default; with allowBack it is confirmBackSelectField's own
// three-option Select, also defaulting to Cancel (design.md "El `prompter`
// de `huh`").
func (p *huhPrompter) Confirm(prompt string, allowBack bool) (installDecision, error) {
	if !allowBack {
		apply := false
		field := huh.NewConfirm().
			Title(prompt).
			Affirmative("Apply").
			Negative("Cancel").
			Value(&apply)
		cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
		if err != nil {
			return installCancelled, err
		}
		if cancelled || !apply {
			return installCancelled, nil
		}
		return installApply, nil
	}
	choice := installCancelled
	field := confirmBackSelectField(prompt, &choice)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return installCancelled, err
	}
	if cancelled {
		return installCancelled, nil
	}
	return choice, nil
}

// hostsMultiSelectField builds SelectHosts' own MultiSelect: no validator (a
// validator on MultiSelect re-prompts forever at the end of input —
// design.md "Contexto verificado"), nothing preselected, and an explicit
// height so every option renders (T3 fix round item 1: huh v2.0.3's
// MultiSelect silently truncates its last option otherwise). Extracted so a
// test can render it directly with .View() without driving a full form.
func hostsMultiSelectField(candidates []hostCandidate, value *[]string) *huh.MultiSelect[string] {
	options := make([]huh.Option[string], len(candidates))
	for i, c := range candidates {
		options[i] = huh.NewOption(hostCandidateLabel(c), c.Name)
	}
	return huh.NewMultiSelect[string]().
		Title("Select CLI hosts").
		Options(options...).
		Value(value).
		Height(fieldHeight(len(options)))
}

// SelectHosts implements prompter.SelectHosts. Cancelling or submitting with
// nothing selected both mean "no changes", matching the text wizard's own
// empty-Enter cancellation.
func (p *huhPrompter) SelectHosts(candidates []hostCandidate) ([]string, bool, error) {
	if len(candidates) == 0 {
		return nil, false, nil
	}
	var selected []string
	field := hostsMultiSelectField(candidates, &selected)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return nil, false, err
	}
	if cancelled || len(selected) == 0 {
		return nil, false, nil
	}
	sort.Strings(selected)
	return selected, true, nil
}

// hostCandidateLabel mirrors selectInstallerHosts's own per-candidate status
// text (install.go), so the huh prompter's own options read the same as the
// text wizard's.
func hostCandidateLabel(c hostCandidate) string {
	status := "not detected"
	switch {
	case c.Detected:
		status = "executable detected"
	case c.Registered:
		status = "registered by Hive"
	case c.Legacy:
		status = "legacy installation detected"
	}
	return fmt.Sprintf("%s (%s)", c.Name, status)
}

// providersMultiSelectField builds SelectProviders' own MultiSelect: the
// same no-validator, nothing-preselected rule as hostsMultiSelectField, with
// the same explicit height (T3 fix round item 1). Extracted so a test can
// render it directly with .View() without driving a full form.
func providersMultiSelectField(offers []providerOffer, value *[]string) *huh.MultiSelect[string] {
	options := make([]huh.Option[string], len(offers))
	for i, o := range offers {
		options[i] = huh.NewOption(providerOfferLabel(o), o.ID)
	}
	return huh.NewMultiSelect[string]().
		Title("Select optional capabilities").
		Options(options...).
		Value(value).
		Height(fieldHeight(len(options)))
}

// SelectProviders implements prompter.SelectProviders, then this prompter's
// own ProviderVersion for every selected capability that is not ManualOnly,
// matching selectProviderRequests's own shape (install.go).
func (p *huhPrompter) SelectProviders(offers []providerOffer) ([]providerRequest, bool, error) {
	if len(offers) == 0 {
		return nil, true, nil
	}
	var chosen []string
	field := providersMultiSelectField(offers, &chosen)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return nil, false, err
	}
	if cancelled {
		return nil, false, nil
	}
	if len(chosen) == 0 {
		return nil, true, nil
	}
	byID := make(map[string]providerOffer, len(offers))
	for _, o := range offers {
		byID[o.ID] = o
	}
	requests := make([]providerRequest, 0, len(chosen))
	for _, id := range chosen {
		offer := byID[id]
		if offer.ManualOnly {
			requests = append(requests, providerRequest{ID: offer.ID})
			continue
		}
		version, ok, err := p.ProviderVersion(offer)
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		requests = append(requests, providerRequest{ID: offer.ID, Version: version})
	}
	return requests, true, nil
}

// providerOfferLabel mirrors selectProviderRequests's own per-offer label
// text (install.go).
func providerOfferLabel(o providerOffer) string {
	version := "exact version required"
	if o.ManualOnly {
		version = "manual instructions only"
	}
	return fmt.Sprintf("%s (%s; %s)", o.Name, o.Source, version)
}

// ProviderVersion implements prompter.ProviderVersion with a required text
// Input, matching readProviderVersion's own "exact version required" rule
// (install.go).
func (p *huhPrompter) ProviderVersion(offer providerOffer) (string, bool, error) {
	var version string
	field := huh.NewInput().
		Title(fmt.Sprintf("Exact version for %s", offer.Name)).
		Validate(func(v string) error {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("exact version required for %s", offer.Name)
			}
			return nil
		}).
		Value(&version)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return "", false, err
	}
	if cancelled || strings.TrimSpace(version) == "" {
		return "", false, nil
	}
	return version, true, nil
}

// menuSelectField builds the menu's own Select, defaulting to Quit (an
// empty Enter or the end of input in accessible mode selects Quit instead
// of huh indexing option -1 and panicking — design.md "Contexto
// verificado"), with an explicit height so every entry renders (T3 fix
// round item 1). Extracted so a test can render it directly with .View()
// without driving a full form.
func menuSelectField(status string, value *menuEntry) *huh.Select[menuEntry] {
	options := make([]huh.Option[menuEntry], len(menuLabels))
	for i, label := range menuLabels {
		options[i] = huh.NewOption(label, menuEntry(i))
	}
	return huh.NewSelect[menuEntry]().
		Title(status).
		Options(options...).
		Value(value).
		Height(fieldHeight(len(options)))
}

// selectMenuEntry presents the interface's own fixed seven-entry menu
// (design.md "La interfaz") via menuSelectField.
func (p *huhPrompter) selectMenuEntry(status string) (menuEntry, bool, error) {
	choice := menuQuit
	field := menuSelectField(status, &choice)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return menuQuit, false, err
	}
	if cancelled {
		return menuQuit, true, nil
	}
	return choice, false, nil
}

// SourceAndRevision implements Update's own two Inputs (design.md "La
// interfaz"): source defaults to "." and revision to "HEAD", exactly
// updateFlags' own defaults (update.go's parseUpdateFlags), so an empty
// Enter in accessible mode (internal/accessibility.PromptString falls back
// to the field's current value, i.e. these defaults) or the field's
// placeholder in a real terminal both keep the same behavior `hive update`
// has with no flags at all.
func (p *huhPrompter) SourceAndRevision() (source, rev string, ok bool, err error) {
	source, rev = ".", "HEAD"
	group := huh.NewGroup(
		huh.NewInput().Title("Source").Description("Git checkout to update from").Value(&source),
		huh.NewInput().Title("Revision").Description("commit-ish to update from").Value(&rev),
	)
	cancelled, err := p.runForm(huh.NewForm(group))
	if err != nil {
		return "", "", false, err
	}
	if cancelled {
		return "", "", false, nil
	}
	return source, rev, true, nil
}

// releasesSelectKeyMap is the Releases screen's own form keymap (T4's
// Esc-filter decision, design.md "La interfaz"): Quit bound only to ctrl+c,
// never esc. huh v2.0.3's Form.Update checks key.Matches(msg, f.keymap.Quit)
// before the active field's own Update ever runs (form.go), so with the
// session's shared formKeyMap (Quit: ctrl+c, esc) an open filter's own esc
// handling (field_select.go's SetFilter/ClearFilter, both bound to esc by
// huh.NewDefaultKeyMap and enabled only while filtering) would never be
// reached: esc would always abort the whole screen first, closing the list
// instead of just closing its filter. Verified directly against the pinned
// version: with this keymap, esc while the release Select is filtering
// clears filtering (Select.GetFiltering() turns false) and the form stays
// StateNormal; esc with no filter open is inert (huh's own Select keymap
// leaves esc unbound to anything else, so nothing happens); ctrl+c still
// aborts the form from either state. Every other field in this package
// keeps the shared formKeyMap (Ctrl-C and Esc both cancel), so this is a
// narrow, deliberate exception to AC10's "Ctrl-C o Esc" for this one field:
// only Ctrl-C cancels Releases' own Select.
var releasesSelectKeyMap = func() *huh.KeyMap {
	k := huh.NewDefaultKeyMap()
	k.Quit = key.NewBinding(key.WithKeys("ctrl+c"))
	return k
}()

// releaseSelectField builds Releases' own Select: filterable (huh's default
// Select keymap already binds "/" to open a filter — no .Filtering(true)
// call, which would instead start the field already inside filter-editing
// mode), a fixed height regardless of how many releases exist (design.md:
// "altura fija", as opposed to hostsMultiSelectField's own per-list
// fieldHeight — a long releases list scrolls or filters instead of growing
// the field), and one option per release via formatReleaseLabel
// (tui_screens.go). Extracted so a test can render or drive it directly.
func releaseSelectField(entries []management.ReleaseEntry, installedID string, value *string) *huh.Select[string] {
	options := make([]huh.Option[string], len(entries))
	for i, e := range entries {
		options[i] = huh.NewOption(formatReleaseLabel(e, e.ID == installedID), e.ID)
	}
	return huh.NewSelect[string]().
		Title("Select a release").
		Options(options...).
		Value(value).
		Height(maxFieldHeight)
}

// SelectRelease presents releaseSelectField through releasesSelectKeyMap
// (T4's Esc-filter decision) instead of the session's shared formKeyMap.
func (p *huhPrompter) SelectRelease(entries []management.ReleaseEntry, installedID string) (string, bool, error) {
	var chosen string
	field := releaseSelectField(entries, installedID, &chosen)
	cancelled, err := p.runFormWithKeyMap(huh.NewForm(huh.NewGroup(field)), releasesSelectKeyMap)
	if err != nil {
		return "", false, err
	}
	if cancelled {
		return "", false, nil
	}
	return chosen, true, nil
}

// voiceSelectField builds Voice's own first Select: every ListVoices entry
// as voiceOptionLabel's own one-line "id — description" (tui_screens.go; a
// full description can run to two sentences and would otherwise wrap across
// several rendered lines, defeating fieldHeight's one-line-per-option model
// — T4 fix round item 1), plus a trailing "Off" option mapped to the empty
// string, a sentinel voiceScreen checks for since a real voice ID is never
// empty.
func voiceSelectField(voices []management.VoiceInfo, value *string) *huh.Select[string] {
	options := make([]huh.Option[string], 0, len(voices)+1)
	for _, v := range voices {
		options = append(options, huh.NewOption(voiceOptionLabel(v), v.ID))
	}
	options = append(options, huh.NewOption("Off", ""))
	return huh.NewSelect[string]().
		Title("Select a voice").
		Options(options...).
		Value(value).
		Height(fieldHeight(len(options)))
}

// SelectVoiceOrOff implements Voice's own first choice (design.md "La
// interfaz"): a voice from ListVoices, or Off.
func (p *huhPrompter) SelectVoiceOrOff(voices []management.VoiceInfo) (string, bool, error) {
	var choice string
	field := voiceSelectField(voices, &choice)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return "", false, err
	}
	if cancelled {
		return "", false, nil
	}
	return choice, true, nil
}

// VoiceDetails implements Voice's own per-voice questions (design.md "La
// interfaz"): address (sir/name/none), a name only when address is "name",
// and intensity (subtle/marked) — management.VoiceSetting's own valid
// values and defaults (voice.go's BuildVoicePlan/RenderVoice). Each question
// is its own single-field form, run in sequence, rather than one multi-field
// group with conditional visibility, since only the second question is ever
// conditional and huh's own group-level field hiding needs no exercise here
// proportional to that.
//
// seed pre-selects each field's own starting value (T6 fix round F5):
// voiceScreen passes the active voice's own current Address/Name/Intensity
// when the chosen voice is the one already active, so pressing Enter
// through every field resubmits the same setting ("Nothing to change")
// instead of silently resetting every field to its bare default the moment
// any one of them is revisited. A zero-value seed (a newly chosen, not
// currently active, voice) keeps today's defaults: "none" and "subtle".
func (p *huhPrompter) VoiceDetails(seed management.VoiceSetting) (address, name, intensity string, ok bool, err error) {
	address = seed.Address
	if address == "" {
		address = "none"
	}
	name = seed.Name
	addressField := huh.NewSelect[string]().
		Title("Address").
		Options(
			huh.NewOption("None", "none"),
			huh.NewOption("Sir", "sir"),
			huh.NewOption("By name", "name"),
		).
		Value(&address).
		Height(fieldHeight(3))
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(addressField)))
	if err != nil {
		return "", "", "", false, err
	}
	if cancelled {
		return "", "", "", false, nil
	}

	if address == "name" {
		nameField := huh.NewInput().
			Title("Name").
			Validate(func(v string) error {
				// Blank input is only invalid when there is no seeded
				// default to fall back to (internal/accessibility.
				// PromptString's own cmp.Or(input, defaultValue) resolves a
				// blank raw entry to defaultValue only after this validator
				// already accepts it — T6 fix round F5 follow-up: a
				// re-selected voice already named seed.Name must accept a
				// blank Enter, not be forced to retype the same name).
				if strings.TrimSpace(v) == "" && seed.Name == "" {
					return fmt.Errorf("a name is required for address \"name\"")
				}
				return nil
			}).
			Value(&name)
		cancelled, err = p.runForm(huh.NewForm(huh.NewGroup(nameField)))
		if err != nil {
			return "", "", "", false, err
		}
		if cancelled {
			return "", "", "", false, nil
		}
	}

	intensity = seed.Intensity
	if intensity == "" {
		intensity = "subtle"
	}
	intensityField := huh.NewSelect[string]().
		Title("Intensity").
		Options(
			huh.NewOption("Subtle", "subtle"),
			huh.NewOption("Marked", "marked"),
		).
		Value(&intensity).
		Height(fieldHeight(2))
	cancelled, err = p.runForm(huh.NewForm(huh.NewGroup(intensityField)))
	if err != nil {
		return "", "", "", false, err
	}
	if cancelled {
		return "", "", "", false, nil
	}
	return address, name, intensity, true, nil
}
