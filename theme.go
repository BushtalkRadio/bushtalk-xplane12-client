package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// BushtalkTheme implements fyne.Theme with Bushtalk Radio's aesthetic
type BushtalkTheme struct{}

var _ fyne.Theme = (*BushtalkTheme)(nil)

// Bushtalk color palette
var (
	// Dark mode (primary theme for desktop app)
	bgDark          = color.NRGBA{R: 15, G: 23, B: 42, A: 255}   // #0f172a - deep slate
	bgDarkSecondary = color.NRGBA{R: 30, G: 41, B: 59, A: 255}   // #1e293b - slate panel
	bgDarkHover     = color.NRGBA{R: 51, G: 65, B: 85, A: 255}   // #334155 - hover state
	textLight       = color.NRGBA{R: 226, G: 232, B: 240, A: 255} // #e2e8f0 - primary text
	textMuted       = color.NRGBA{R: 148, G: 163, B: 184, A: 255} // #94a3b8 - muted text
	accentOrange    = color.NRGBA{R: 251, G: 146, B: 60, A: 255}  // #fb923c - accent
	terracotta      = color.NRGBA{R: 193, G: 120, B: 88, A: 255}  // #c17858 - terracotta
	borderDark      = color.NRGBA{R: 71, G: 85, B: 105, A: 153}   // slate-600 @ 60%
	successGreen    = color.NRGBA{R: 52, G: 211, B: 153, A: 255}  // #34d399
	errorRed        = color.NRGBA{R: 248, G: 113, B: 113, A: 255} // #f87171
)

func (t *BushtalkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return bgDark
	case theme.ColorNameForeground:
		return textLight
	case theme.ColorNameButton:
		return bgDarkSecondary
	case theme.ColorNameDisabledButton:
		return bgDarkHover
	case theme.ColorNameDisabled:
		return textMuted
	case theme.ColorNamePlaceHolder:
		return textMuted
	case theme.ColorNamePrimary:
		return accentOrange
	case theme.ColorNameFocus:
		return accentOrange
	case theme.ColorNameSelection:
		return color.NRGBA{R: 251, G: 146, B: 60, A: 77} // accent @ 30%
	case theme.ColorNameHover:
		return bgDarkHover
	case theme.ColorNameInputBackground:
		return bgDarkSecondary
	case theme.ColorNameInputBorder:
		return borderDark
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 102} // black @ 40%
	case theme.ColorNameSuccess:
		return successGreen
	case theme.ColorNameError:
		return errorRed
	case theme.ColorNameWarning:
		return terracotta
	case theme.ColorNameHeaderBackground:
		return bgDarkSecondary
	case theme.ColorNameMenuBackground:
		return bgDarkSecondary
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 15, G: 23, B: 42, A: 230} // bgDark @ 90%
	case theme.ColorNameScrollBar:
		return borderDark
	case theme.ColorNameSeparator:
		return borderDark
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *BushtalkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *BushtalkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *BushtalkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameScrollBarSmall:
		return 4
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 16
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 8
	case theme.SizeNameSelectionRadius:
		return 6
	}
	return theme.DefaultTheme().Size(name)
}
