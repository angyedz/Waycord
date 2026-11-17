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
	Canvas: Hex(0x141218), Rail: Hex(0x0F0D13), Sidebar: Hex(0x1D1B20),
	Surface: Hex(0x2B2930), Surface2: Hex(0x36343B),
	Text: Hex(0xE7E0E9), Muted: Hex(0xCAC4D0), Subtle: Hex(0xAAA4B1),
	Accent: Hex(0xD0BCFF), AccentDim: Hex(0x4A4458), Border: Hex(0x49454F),
	Positive: Hex(0xA8DAB5), Warning: Hex(0xE9BD79),
}
