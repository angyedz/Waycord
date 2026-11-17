package markdown

import (
	"image/color"
	"strings"
	"testing"

	"gioui.org/font"
)

func TestRenderCommonMessageMarkdown(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{R: 220, G: 220, B: 230, A: 255}, color.NRGBA{R: 150, G: 130, B: 255, A: 255})
	spans, err := renderer.Render("**bold** and *italic* with `code`\n\n> quote\n\n- item")
	if err != nil {
		t.Fatal(err)
	}
	var bold, italic, code, quote, item bool
	for _, span := range spans {
		switch strings.TrimSpace(span.Style.Content) {
		case "bold":
			bold = span.Style.Font.Weight == font.Bold
		case "italic":
			italic = span.Style.Font.Style == font.Italic
		case "code":
			code = span.Style.Font.Typeface == "monospace"
		case "quote":
			quote = true
		case "item":
			item = true
		}
	}
	if !bold || !italic || !code || !quote || !item {
		t.Fatalf("missing markdown styles: bold=%t italic=%t code=%t quote=%t item=%t spans=%+v", bold, italic, code, quote, item, spans)
	}
	plain := PlainText(spans)
	for _, visible := range []string{"bold", "italic", "code", "quote", "item"} {
		if !strings.Contains(plain, visible) {
			t.Errorf("selectable text omitted %q: %q", visible, plain)
		}
	}
}

func TestRenderLinkKeepsItsDestinationMetadataAndLabel(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("[guide](https://example.test/guide)")
	if err != nil {
		t.Fatal(err)
	}
	if got := PlainText(spans); got != "guide" {
		t.Fatalf("unexpected visible link text: %q", got)
	}
	if len(spans) == 0 || spans[0].Style.Color != (color.NRGBA{B: 255, A: 255}) {
		t.Fatalf("link was not styled: %+v", spans)
	}
	if !spans[0].Style.Interactive {
		t.Fatalf("link is not interactive: %+v", spans[0])
	}
}

func TestCacheIsBoundedAndRefreshesChangedMessages(t *testing.T) {
	cache := NewCache(New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255}), 2)
	for index, source := range []string{"one", "two", "three"} {
		if _, err := cache.Render(index, source); err != nil {
			t.Fatal(err)
		}
	}
	if len(cache.entries) != 2 {
		t.Fatalf("cache exceeded its capacity: %d", len(cache.entries))
	}
	spans, err := cache.Render(2, "updated")
	if err != nil {
		t.Fatal(err)
	}
	if got := PlainText(spans); got != "updated" {
		t.Fatalf("stale message render returned: %q", got)
	}
}

func TestDiscordStrikeAndUnderlineSkipInlineCode(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("~~strike~~ __underline__ `~~literal~~` __unfinished")
	if err != nil {
		t.Fatal(err)
	}
	var strike, underline, literal bool
	for _, span := range spans {
		switch span.Style.Content {
		case "strike":
			strike = span.Strike
		case "underline":
			underline = span.Underline
		case "~~literal~~":
			literal = !span.Strike && !span.Underline && span.Style.Font.Typeface == "monospace"
		}
	}
	if !strike || !underline || !literal || !strings.Contains(PlainText(spans), "__unfinished") {
		t.Fatalf("Discord delimiters were not rendered correctly: %+v", spans)
	}
}
