// Package theme centralizes Waycord's visual tokens, avoiding hard-coded
// colors scattered throughout individual widgets.
package theme

import "image/color"

type Palette struct {
	Canvas, Rail, Sidebar, Surface, Surface2       color.NRGBA
	Text, Muted, Subtle, Accent, AccentDim, Border color.NRGBA
	Positive, Warning                              color.NRGBA
}

func Hex(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

var Dark = Palette{
	Canvas: Hex(0x10141D), Rail: Hex(0x0B0E15), Sidebar: Hex(0x141925),
	Surface: Hex(0x1B2230), Surface2: Hex(0x252D3E),
	Text: Hex(0xF0F3FA), Muted: Hex(0xA0AABD), Subtle: Hex(0x697487),
	Accent: Hex(0x9585FA), AccentDim: Hex(0x342E58), Border: Hex(0x2B3444),
	Positive: Hex(0x60D5A4), Warning: Hex(0xE9BD79),
}
