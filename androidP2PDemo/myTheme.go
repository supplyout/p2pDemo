package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"image/color"
)

var (
	darkPrimary   = &color.RGBA{R: 0x00, G: 0x97, B: 0xA7, A: 255}
	lightPrimary  = &color.RGBA{R: 0xB2, G: 0xEB, B: 0xF2, A: 255}
	primary       = &color.RGBA{R: 0x00, G: 0xBC, B: 0xD4, A: 255}
	accent        = &color.RGBA{R: 0x03, G: 0xA9, B: 0xF4, A: 255}
	textPrimary   = &color.RGBA{R: 0x21, G: 0x21, B: 0x21, A: 255}
	textSecondary = &color.RGBA{R: 0x75, G: 0x75, B: 0x75, A: 255}
	white         = &color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 255}
	background    = &color.RGBA{R: 233, G: 233, B: 227, A: 255}
	dividerColor  = &color.RGBA{R: 0xBD, G: 0xBD, B: 0xBD, A: 255}
	errorColor    = &color.RGBA{R: 0xD3, G: 0x2F, B: 0x2F, A: 255}
)

type myTheme struct {
}

func (m myTheme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	return theme.DarkTheme().Color(c, v)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return resourceSourceHanSansCNMediumMiniTtf
}

func (m myTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(n)
}

func (m myTheme) Size(s fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(s)
}

func NewMyTheme() fyne.Theme {

	return &myTheme{}
}
