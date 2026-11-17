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
	Canvas: Hex(0x17191F), Rail: Hex(0x101217), Sidebar: Hex(0x1C1F26),
	Surface: Hex(0x242830), Surface2: Hex(0x303540),
	Text: Hex(0xECEEF2), Muted: Hex(0xB4BAC5), Subtle: Hex(0x89919F),
	Accent: Hex(0x91B9AD), AccentDim: Hex(0x33443F), Border: Hex(0x353A44),
	Positive: Hex(0x8CCAB0), Warning: Hex(0xE9BD79),
}
