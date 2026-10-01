// tui_contrast_test.go holds the WCAG 2.1 AA contrast table of the
// full-screen application's own theme (TestThemeContrastMeetsWCAGAA,
// tui_theme.go).
package main

import (
	"image/color"
	"math"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

// srgbChannelToLinear converts one 16-bit-per-channel sRGB component (as
// color.Color.RGBA() returns) to linear light, the first step of the WCAG
// 2.1 relative luminance formula (https://www.w3.org/TR/WCAG21/#dfn-relative-luminance).
func srgbChannelToLinear(v uint32) float64 {
	c := float64(v) / 65535.0
	if c <= 0.03928 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

// relativeLuminance is WCAG 2.1's own L = 0.2126*R + 0.7152*G + 0.0722*B,
// computed directly from whatever color.Color a resolved lipgloss style's
// GetForeground()/GetBackground() returns — an ANSI-256 index, a hex color,
// or lipgloss's own NoColor sentinel (RGBA()'s own defined fallback, pure
// black) — with no separate ANSI-256-to-RGB table of this package's own.
func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return 0.2126*srgbChannelToLinear(r) + 0.7152*srgbChannelToLinear(g) + 0.0722*srgbChannelToLinear(b)
}

// wcagContrastRatio is WCAG 2.1 SC 1.4.3's own (L1+0.05)/(L2+0.05), L1 the
// lighter of the two colors' relative luminance.
func wcagContrastRatio(fg, bg color.Color) float64 {
	l1, l2 := relativeLuminance(fg), relativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

// wcagNormalTextMinimum is WCAG 2.1 AA's own minimum contrast ratio for
// normal-weight text (SC 1.4.3); every field checked here renders normal
// text except Title (bold — see its own check below).
const wcagNormalTextMinimum = 4.5

var (
	refBlack     = lipgloss.Color("#000000")
	refDarkGray  = lipgloss.Color("#1e1e1e")
	refWhite     = lipgloss.Color("#ffffff")
	refLightGray = lipgloss.Color("#f0f0f0")
)

// bothDarkRefs and bothLightRefs are the two representative terminal
// backgrounds a dark or light theme, respectively, must stay readable
// against: pure black/white and a common near-black/near-white terminal
// default.
var (
	bothDarkRefs  = []color.Color{refBlack, refDarkGray}
	bothLightRefs = []color.Color{refWhite, refLightGray}
)

// TestThemeContrastMeetsWCAGAA measures the full-screen application's own
// styles (tui_theme.go) with the same WCAG helpers: every style that renders
// text must reach 4.5:1 against both reference backgrounds of its branch (pure
// black and a common near-black for the dark theme; pure white and a common
// near-white for the light one), and a button, which paints its own
// background, must reach it against that background.
func TestThemeContrastMeetsWCAGAA(t *testing.T) {
	type textCheck struct {
		name string
		get  func(th appTheme) lipgloss.Style
	}
	textChecks := []textCheck{
		{"Text", func(th appTheme) lipgloss.Style { return th.Text }},
		{"Muted", func(th appTheme) lipgloss.Style { return th.Muted }},
		{"Title", func(th appTheme) lipgloss.Style { return th.Title }},
		{"Accent", func(th appTheme) lipgloss.Style { return th.Accent }},
		{"Success", func(th appTheme) lipgloss.Style { return th.Success }},
		{"Danger", func(th appTheme) lipgloss.Style { return th.Danger }},
		{"ButtonOff", func(th appTheme) lipgloss.Style { return th.ButtonOff }},
		{"Help.ShortKey", func(th appTheme) lipgloss.Style { return th.Help.ShortKey }},
		{"Help.ShortDesc", func(th appTheme) lipgloss.Style { return th.Help.ShortDesc }},
		{"Help.ShortSeparator", func(th appTheme) lipgloss.Style { return th.Help.ShortSeparator }},
		{"Help.Ellipsis", func(th appTheme) lipgloss.Style { return th.Help.Ellipsis }},
	}
	for _, isDark := range []bool{true, false} {
		th := newAppTheme(isDark, false)
		refs := bothLightRefs
		if isDark {
			refs = bothDarkRefs
		}
		for _, tc := range textChecks {
			t.Run(themeBranch(isDark)+"/"+tc.name, func(t *testing.T) {
				fg := tc.get(th).GetForeground()
				if _, unset := fg.(lipgloss.NoColor); unset {
					t.Fatalf("%s carries no foreground color", tc.name)
				}
				for _, bg := range refs {
					if r := wcagContrastRatio(fg, bg); r < wcagNormalTextMinimum {
						t.Errorf("%s (%v) vs %v = %.2f, want >= %.1f", tc.name, fg, bg, r, wcagNormalTextMinimum)
					}
				}
			})
		}
		t.Run(themeBranch(isDark)+"/ButtonOn", func(t *testing.T) {
			fg, bg := th.ButtonOn.GetForeground(), th.ButtonOn.GetBackground()
			if r := wcagContrastRatio(fg, bg); r < wcagNormalTextMinimum {
				t.Errorf("ButtonOn foreground (%v) vs its own background (%v) = %.2f, want >= %.1f", fg, bg, r, wcagNormalTextMinimum)
			}
		})
	}
}

func themeBranch(isDark bool) string {
	if isDark {
		return "dark"
	}
	return "light"
}

// TestThemeNoColorStylesCarryNoColor covers NO_COLOR: no style of the
// no-color theme sets a foreground or background color.
func TestThemeNoColorStylesCarryNoColor(t *testing.T) {
	th := newAppTheme(true, true)
	styles := map[string]lipgloss.Style{
		"Text": th.Text, "Muted": th.Muted, "Title": th.Title, "Accent": th.Accent,
		"Success": th.Success, "Danger": th.Danger, "ButtonOn": th.ButtonOn, "ButtonOff": th.ButtonOff,
		"Help.ShortKey": th.Help.ShortKey, "Help.ShortDesc": th.Help.ShortDesc,
		"Help.ShortSeparator": th.Help.ShortSeparator, "Help.Ellipsis": th.Help.Ellipsis,
	}
	var _ help.Styles = th.Help
	for name, st := range styles {
		if _, ok := st.GetForeground().(lipgloss.NoColor); !ok {
			t.Errorf("%s has a foreground color under NO_COLOR: %v", name, st.GetForeground())
		}
		if _, ok := st.GetBackground().(lipgloss.NoColor); !ok {
			t.Errorf("%s has a background color under NO_COLOR: %v", name, st.GetBackground())
		}
	}
}

// TestThemeSurfaceKeepsTextReadable: the picker's panel paints a background, so
// every text style it uses must reach 4.5:1 against that background.
func TestThemeSurfaceKeepsTextReadable(t *testing.T) {
	for _, isDark := range []bool{true, false} {
		th := newAppTheme(isDark, false)
		bg := th.Surface.GetBackground()
		if _, unset := bg.(lipgloss.NoColor); unset {
			t.Fatalf("%s: the surface has no background", themeBranch(isDark))
		}
		for name, st := range map[string]lipgloss.Style{"Text": th.Text, "Muted": th.Muted, "Title": th.Title, "Accent": th.Accent, "Danger": th.Danger} {
			if r := wcagContrastRatio(st.GetForeground(), bg); r < wcagNormalTextMinimum {
				t.Errorf("%s %s on the surface = %.2f, want >= %.1f", themeBranch(isDark), name, r, wcagNormalTextMinimum)
			}
		}
	}
	if _, ok := newAppTheme(true, true).Surface.GetBackground().(lipgloss.NoColor); !ok {
		t.Error("the NO_COLOR surface paints a background")
	}
}
