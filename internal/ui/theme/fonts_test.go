package theme

import (
	"testing"

	giofont "gioui.org/font"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"

	emoji "eliasnaur.com/font/noto/emoji/color"
	"gioui.org/font/opentype"
	gotext "github.com/go-text/typesetting/font"
)

func TestEmojiFallbackIncludesBitmapGlyph(t *testing.T) {
	parsed, err := opentype.Parse(emoji.TTF)
	if err != nil {
		t.Fatal(err)
	}
	face := parsed.Face()
	glyph, ok := face.NominalGlyph('😊')
	if !ok {
		t.Fatal("emoji glyph unavailable")
	}
	if _, ok := face.GlyphData(glyph).(gotext.GlyphBitmap); !ok {
		t.Fatal("emoji glyph has no bitmap")
	}
	if _, err := NewShaper(); err != nil {
		t.Fatal(err)
	}
}

func TestEmojiShapingHasVisibleBounds(t *testing.T) {
	shaper, err := NewShaper()
	if err != nil {
		t.Fatal(err)
	}
	shaper.LayoutString(text.Parameters{Font: giofont.Font{Typeface: giofont.Typeface(TextFamilies)}, PxPerEm: fixed.I(16), MaxWidth: 500}, "😊")
	glyph, ok := shaper.NextGlyph()
	if !ok || glyph.Bounds.Empty() {
		t.Fatalf("emoji has no visible glyph: %+v", glyph)
	}
}

func TestTextAndEmojiUseReadableGlyphMetrics(t *testing.T) {
	shaper, err := NewShaper()
	if err != nil {
		t.Fatal(err)
	}
	params := text.Parameters{Font: giofont.Font{Typeface: giofont.Typeface(TextFamilies)}, PxPerEm: fixed.I(14), MaxWidth: 500}
	for _, value := range []string{"#", "0", "9"} {
		shaper.LayoutString(params, value)
		glyph, ok := shaper.NextGlyph()
		if !ok || glyph.Bounds.Empty() || glyph.Advance > fixed.I(12) {
			t.Errorf("unexpected text glyph for %q: %+v", value, glyph)
		}
	}
}
