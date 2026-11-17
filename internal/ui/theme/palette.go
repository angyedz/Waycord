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

var Midnight = Palette{
	Canvas: Hex(0x09080C), Rail: Hex(0x050508), Sidebar: Hex(0x121116),
	Surface: Hex(0x1C1A21), Surface2: Hex(0x2B2930),
	Text: Hex(0xF0ECF4), Muted: Hex(0xCBC5D0), Subtle: Hex(0xA7A1AD),
	Accent: Hex(0xD0BCFF), AccentDim: Hex(0x40394D), Border: Hex(0x38353E),
	Positive: Hex(0xA8DAB5), Warning: Hex(0xE9BD79),
}

var Light = Palette{
	Canvas: Hex(0xF7F5FA), Rail: Hex(0xEAE7EF), Sidebar: Hex(0xF0EDF5),
	Surface: Hex(0xFFFFFF), Surface2: Hex(0xE6E1EA),
	Text: Hex(0x201A24), Muted: Hex(0x49454F), Subtle: Hex(0x625B66),
	Accent: Hex(0x6750A4), AccentDim: Hex(0xEADDFF), Border: Hex(0xCAC4D0),
	Positive: Hex(0x1A6728), Warning: Hex(0x7A5900),
}

func ForName(name string) (Palette, bool) {
	switch name {
	case "dark":
		return Dark, true
	case "midnight":
		return Midnight, true
	case "light":
		return Light, true
	default:
		return Palette{}, false
	}
}
