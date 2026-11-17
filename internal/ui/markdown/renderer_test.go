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

func TestDiscordSpoilersAreMarkedAndKeepNestedFormatting(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("before ||hidden **bold**|| after")
	if err != nil {
		t.Fatal(err)
	}
	var before, hidden, bold, after bool
	var spoilerID int
	for _, span := range spans {
		switch strings.TrimSpace(span.Style.Content) {
		case "before":
			before = !span.Spoiler
		case "hidden":
			hidden = span.Spoiler
			spoilerID = span.SpoilerID
		case "bold":
			bold = span.Spoiler && span.Style.Font.Weight == font.Bold && span.SpoilerID == spoilerID
		case "after":
			after = !span.Spoiler
		}
	}
	if !before || !hidden || !bold || !after || spoilerID < 0 {
		t.Fatalf("spoiler content or nested formatting was lost: %+v", spans)
	}
	if got := PlainText(spans); got != "before hidden bold after" {
		t.Fatalf("spoiler text must remain selectable: %q", got)
	}
}

func TestDiscordSpoilerDelimiterIsLiteralWhenUnmatchedOrEscaped(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("||unfinished\n\\||literal\\||")
	if err != nil {
		t.Fatal(err)
	}
	plain := PlainText(spans)
	if !strings.Contains(plain, "||unfinished") || !strings.Contains(plain, "||literal||") {
		t.Fatalf("unmatched or escaped spoiler delimiters disappeared: %q", plain)
	}
	for _, span := range spans {
		if span.Spoiler {
			t.Fatalf("unmatched or escaped spoiler was activated: %+v", span)
		}
	}
}

func TestDiscordSpoilerDelimiterIsIgnoredInsideCode(t *testing.T) {
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("`||inline||`\n\n```\n||fenced||\n```")
	if err != nil {
		t.Fatal(err)
	}
	if got := PlainText(spans); !strings.Contains(got, "||inline||") || !strings.Contains(got, "||fenced||") {
		t.Fatalf("code delimiters were altered: %q", got)
	}
	for _, span := range spans {
		if span.Spoiler {
			t.Fatalf("code content was marked as a spoiler: %+v", span)
		}
	}
}

func TestDisplayStylesMaskSpoilersUntilRevealed(t *testing.T) {
	mask := color.NRGBA{R: 100, G: 100, B: 110, A: 255}
	renderer := New("Waycord Sans", color.NRGBA{A: 255}, color.NRGBA{B: 255, A: 255})
	spans, err := renderer.Render("public ||secret **phrase**||")
	if err != nil {
		t.Fatal(err)
	}
	masked := strings.Builder{}
	styles := DisplayStyles(spans, nil, mask)
	for i, span := range spans {
		if span.Spoiler {
			masked.WriteString(styles[i].Content)
			if strings.Contains(styles[i].Content, "secret") || strings.Contains(styles[i].Content, "phrase") {
				t.Fatalf("unrevealed spoiler text is visible: %+v", styles[i])
			}
			if styles[i].Color != mask || !styles[i].Interactive {
				t.Fatalf("spoiler mask lost its appearance or click target: %+v", styles[i])
			}
		}
	}
	if strings.Contains(masked.String(), "secret") || PlainText(spans) != "public secret phrase" {
		t.Fatalf("masking should hide only the display copy while preserving selectable content: masked=%q plain=%q", masked.String(), PlainText(spans))
	}
	revealed := map[int]bool{0: true}
	styles = DisplayStyles(spans, revealed, mask)
	var visible string
	for i, span := range spans {
		if span.Spoiler {
			visible += styles[i].Content
			if styles[i].Color == mask || !strings.Contains(styles[i].Content, "secret") && !strings.Contains(styles[i].Content, "phrase") {
				t.Fatalf("revealed spoiler remained masked: %+v", styles[i])
			}
		}
	}
	if visible != "secret phrase" {
		t.Fatalf("revealed content changed: %q", visible)
	}
}
