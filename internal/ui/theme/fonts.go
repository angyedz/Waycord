package theme

import (
	"fmt"

	emoji "eliasnaur.com/font/noto/emoji/color"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

const TextFamilies = "Go, Waycord Emoji, sans-serif"

// NewShaper includes a bundled color-emoji fallback on Linux and Windows.
func NewShaper() (*text.Shaper, error) {
	face, err := opentype.Parse(emoji.TTF)
	if err != nil {
		return nil, fmt.Errorf("load emoji font: %w", err)
	}
	collection := append(gofont.Collection(), font.FontFace{Font: font.Font{Typeface: "Waycord Emoji"}, Face: face})
	return text.NewShaper(text.WithCollection(collection)), nil
}
