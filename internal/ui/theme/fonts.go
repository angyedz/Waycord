package theme

import (
	"fmt"

	emoji "eliasnaur.com/font/noto/emoji/color"
	"eliasnaur.com/font/roboto/robotomedium"
	"eliasnaur.com/font/roboto/robotoregular"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/font/opentype"
	"gioui.org/text"
)

const TextFamilies = "Waycord Sans, Waycord Emoji, sans-serif"

// NewShaper includes a bundled color-emoji fallback on Linux and Windows.
func NewShaper() (*text.Shaper, error) {
	face, err := opentype.Parse(emoji.TTF)
	if err != nil {
		return nil, fmt.Errorf("load emoji font: %w", err)
	}
	regular, err := opentype.Parse(robotoregular.TTF)
	if err != nil {
		return nil, fmt.Errorf("load regular font: %w", err)
	}
	medium, err := opentype.Parse(robotomedium.TTF)
	if err != nil {
		return nil, fmt.Errorf("load medium font: %w", err)
	}
	collection := append(gofont.Collection(),
		font.FontFace{Font: font.Font{Typeface: "Waycord Sans"}, Face: regular},
		font.FontFace{Font: font.Font{Typeface: "Waycord Sans", Weight: font.Medium}, Face: medium},
		font.FontFace{Font: font.Font{Typeface: "Waycord Emoji"}, Face: face},
	)
	return text.NewShaper(text.WithCollection(collection)), nil
}
