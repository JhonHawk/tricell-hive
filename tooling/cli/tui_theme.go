// tui_theme.go holds the full-screen application's own lipgloss styles
// (design.md "Tema"): light/dark pairs chosen by the detected terminal
// background, and a color-free variant for NO_COLOR. Selection, checkboxes and
// buttons are always drawn with symbols (">", "[x]", "[ ]", "[Apply]"), so no
// state depends on color alone.
package main

import (
	"charm.land/bubbles/v2/help"
	"charm.land/lipgloss/v2"
)

// appTheme is every style the application draws with. Each style that renders
// text is checked for WCAG AA contrast by TestThemeContrastMeetsWCAGAA.
type appTheme struct {
	Text    lipgloss.Style // ordinary rows and body text
	Muted   lipgloss.Style // descriptions, the status line, hints
	Title   lipgloss.Style // a view's heading
	Accent  lipgloss.Style // the selected row
	Success lipgloss.Style // a completed operation
	Danger  lipgloss.Style // an error or a warning
	// ButtonOn paints its own background: the chosen button of a pair.
	ButtonOn  lipgloss.Style
	ButtonOff lipgloss.Style
	Help      help.Styles
}

// themeColors is one branch (dark or light terminal background) of the
// palette. Every foreground clears 4.5:1 against pure black and #1e1e1e
// (dark) or pure white and #f0f0f0 (light).
type themeColors struct {
	text, muted, accent, success, danger string
	buttonFG, buttonBG                   string
}

var (
	darkColors = themeColors{
		text: "#e4e4e4", muted: "#a8a8a8", accent: "#afafff", success: "#5fd787", danger: "#ff8787",
		buttonFG: "#1a1a1a", buttonBG: "#afafff",
	}
	lightColors = themeColors{
		text: "#262626", muted: "#585858", accent: "#3a3ab8", success: "#006b3c", danger: "#b3001b",
		buttonFG: "#ffffff", buttonBG: "#3a3ab8",
	}
)

// newAppTheme builds the theme for a dark or light background. With noColor
// (NO_COLOR is set) no style sets a color; only the bold attribute remains.
func newAppTheme(isDark, noColor bool) appTheme {
	plain := lipgloss.NewStyle()
	if noColor {
		return appTheme{
			Text: plain, Muted: plain, Title: plain.Bold(true), Accent: plain.Bold(true),
			Success: plain, Danger: plain.Bold(true),
			ButtonOn: plain.Bold(true), ButtonOff: plain,
			Help: help.Styles{
				ShortKey: plain, ShortDesc: plain, ShortSeparator: plain, Ellipsis: plain,
				FullKey: plain, FullDesc: plain, FullSeparator: plain,
			},
		}
	}
	c := lightColors
	if isDark {
		c = darkColors
	}
	fg := func(hex string) lipgloss.Style { return plain.Foreground(lipgloss.Color(hex)) }
	return appTheme{
		Text:      fg(c.text),
		Muted:     fg(c.muted),
		Title:     fg(c.accent).Bold(true),
		Accent:    fg(c.accent).Bold(true),
		Success:   fg(c.success),
		Danger:    fg(c.danger).Bold(true),
		ButtonOn:  plain.Foreground(lipgloss.Color(c.buttonFG)).Background(lipgloss.Color(c.buttonBG)).Bold(true),
		ButtonOff: fg(c.text),
		Help: help.Styles{
			ShortKey: fg(c.muted), ShortDesc: fg(c.muted), ShortSeparator: fg(c.muted), Ellipsis: fg(c.muted),
			FullKey: fg(c.muted), FullDesc: fg(c.muted), FullSeparator: fg(c.muted),
		},
	}
}
