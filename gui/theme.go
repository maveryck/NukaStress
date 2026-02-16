package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var (
	pipGreen         = color.NRGBA{R: 0x6A, G: 0xD8, B: 0x7A, A: 0xFF}
	pipGreenSoft     = color.NRGBA{R: 0xB2, G: 0xE8, B: 0xBE, A: 0xFF}
	nuclearRed       = color.NRGBA{R: 0xFF, G: 0x44, B: 0x2A, A: 0xFF}
	quantumYellow    = color.NRGBA{R: 0x8C, G: 0xFF, B: 0x66, A: 0xFF}
	wastelandBg      = color.NRGBA{R: 0x06, G: 0x0A, B: 0x06, A: 0xFF}
	wastelandPanel   = color.NRGBA{R: 0x0F, G: 0x16, B: 0x0F, A: 0xFF}
	wastelandPanelHi = color.NRGBA{R: 0x1A, G: 0x28, B: 0x1A, A: 0xFF}
	wastelandBorder  = color.NRGBA{R: 0x2F, G: 0x5A, B: 0x2F, A: 0xFF}
	wastelandHover   = color.NRGBA{R: 0x2A, G: 0x4B, B: 0x2F, A: 0xFF}
)

type PipBoyTheme struct{}

func (PipBoyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return wastelandBg
	case theme.ColorNameForeground:
		return pipGreenSoft
	case theme.ColorNamePrimary:
		return pipGreen
	case theme.ColorNameInputBackground:
		return wastelandPanel
	case theme.ColorNameButton:
		return wastelandPanelHi
	case theme.ColorNameMenuBackground:
		return wastelandPanel
	case theme.ColorNameOverlayBackground:
		return wastelandPanel
	case theme.ColorNameHover:
		return wastelandHover
	case theme.ColorNamePressed:
		return wastelandBorder
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x68, G: 0xAE, B: 0x68, A: 0xFF}
	case theme.ColorNameSeparator:
		return wastelandBorder
	case theme.ColorNameScrollBar:
		return wastelandBorder
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x1E, G: 0x3D, B: 0x1E, A: 0xFF}
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 0x53, G: 0x72, B: 0x53, A: 0xFF}
	case theme.ColorNameError:
		return nuclearRed
	case theme.ColorNameWarning:
		return quantumYellow
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (PipBoyTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Bold && pipBoyBoldFont != nil {
		return pipBoyBoldFont
	}
	if style.Monospace {
		return theme.DefaultTextMonospaceFont()
	}
	if pipBoyDisplayFont != nil && style.Bold {
		return pipBoyDisplayFont
	}
	return theme.DefaultTextMonospaceFont()
}

func (PipBoyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (PipBoyTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return theme.DefaultTheme().Size(name) + 1
	}
	return theme.DefaultTheme().Size(name)
}
