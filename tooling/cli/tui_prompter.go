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
// preferring ThemeBase when NO_COLOR is set and otherwise ThemeCharm with
// the background lipgloss.HasDarkBackground (charm.land/lipgloss/v2 v2.0.1,
// the version this module's go.sum pins) detects.
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
	styles := huh.ThemeCharm(isDark)
	p.theme = huh.ThemeFunc(func(bool) *huh.Styles { return styles })
	return p
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

// runForm runs one single-group form through this prompter's own
// accessible/theme/IO settings, translating cancellation the same way for
// every field type (design.md "Cancelar").
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
func (p *huhPrompter) runForm(form *huh.Form) (cancelled bool, err error) {
	form = form.WithAccessible(p.accessible).WithInput(p.in).WithOutput(p.out).WithKeyMap(formKeyMap)
	if p.theme != nil {
		form = form.WithTheme(p.theme)
	}
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

// Confirm implements prompter.Confirm. Without allowBack it is a plain
// huh.Confirm defaulting to Cancel, exactly like confirmInstall's own
// empty-Enter default; with allowBack it is a three-option Select (Apply,
// Back, Cancel), also defaulting to Cancel (design.md "El `prompter` de
// `huh`").
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
	field := huh.NewSelect[installDecision]().
		Title(prompt).
		Options(
			huh.NewOption("Apply", installApply),
			huh.NewOption("Back", installBack),
			huh.NewOption("Cancel", installCancelled),
		).
		Value(&choice)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return installCancelled, err
	}
	if cancelled {
		return installCancelled, nil
	}
	return choice, nil
}

// SelectHosts implements prompter.SelectHosts with a MultiSelect: no
// validator (a validator on MultiSelect re-prompts forever at the end of
// input — design.md "Contexto verificado") and nothing preselected, exactly
// like the text wizard's own selectInstallerHosts (install.go). Cancelling
// or submitting with nothing selected both mean "no changes", matching the
// text wizard's own empty-Enter cancellation.
func (p *huhPrompter) SelectHosts(candidates []hostCandidate) ([]string, bool, error) {
	if len(candidates) == 0 {
		return nil, false, nil
	}
	options := make([]huh.Option[string], len(candidates))
	for i, c := range candidates {
		options[i] = huh.NewOption(hostCandidateLabel(c), c.Name)
	}
	var selected []string
	field := huh.NewMultiSelect[string]().
		Title("Select CLI hosts").
		Options(options...).
		Value(&selected)
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

// SelectProviders implements prompter.SelectProviders: a MultiSelect for the
// optional capabilities (the same no-validator, nothing-preselected rule as
// SelectHosts), then this prompter's own ProviderVersion for every selected
// capability that is not ManualOnly, matching selectProviderRequests's own
// shape (install.go).
func (p *huhPrompter) SelectProviders(offers []providerOffer) ([]providerRequest, bool, error) {
	if len(offers) == 0 {
		return nil, true, nil
	}
	options := make([]huh.Option[string], len(offers))
	for i, o := range offers {
		options[i] = huh.NewOption(providerOfferLabel(o), o.ID)
	}
	var chosen []string
	field := huh.NewMultiSelect[string]().
		Title("Select optional capabilities").
		Options(options...).
		Value(&chosen)
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

// selectMenuEntry presents the interface's own fixed seven-entry menu
// (design.md "La interfaz"). Quit is always its default, so an empty Enter
// or the end of input in accessible mode selects Quit instead of huh
// indexing option -1 and panicking (design.md "Contexto verificado").
func (p *huhPrompter) selectMenuEntry(status string) (menuEntry, bool, error) {
	choice := menuQuit
	options := make([]huh.Option[menuEntry], len(menuLabels))
	for i, label := range menuLabels {
		options[i] = huh.NewOption(label, menuEntry(i))
	}
	field := huh.NewSelect[menuEntry]().
		Title(status).
		Options(options...).
		Value(&choice)
	cancelled, err := p.runForm(huh.NewForm(huh.NewGroup(field)))
	if err != nil {
		return menuQuit, false, err
	}
	if cancelled {
		return menuQuit, true, nil
	}
	return choice, false, nil
}
