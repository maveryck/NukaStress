package gui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/agaaler.ttf
var agaalerTTF []byte

//go:embed assets/pixel_square_bold10.ttf
var pixelSquareBold10 []byte

var pipBoyBoldFont = fyne.NewStaticResource("agaaler.ttf", agaalerTTF)
var pipBoyDisplayFont = fyne.NewStaticResource("pixel_square_bold10.ttf", pixelSquareBold10)
