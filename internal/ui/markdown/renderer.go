package markdown

import (
	"container/list"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/unit"
	gioMarkdown "gioui.org/x/markdown"
	"gioui.org/x/richtext"
)

// Renderer adapts the official Gio Markdown renderer to Waycord's compact chat style.
type Renderer struct {
	parser *gioMarkdown.Renderer
}

type cacheEntry struct {
	index  int
	source string
	spans  []Span
}

// Cache keeps rendered spans for recently visible messages only.
type Cache struct {
	renderer *Renderer
	capacity int
	entries  map[int]*list.Element
	order    *list.List
}

const URLKey = gioMarkdown.MetadataURL

func New(face string, body, link color.NRGBA) *Renderer {
	parser := gioMarkdown.NewRenderer()
	parser.Config = gioMarkdown.Config{
		DefaultFont:      font.Font{Typeface: font.Typeface(face), Weight: font.Normal, Style: font.Regular},
		MonospaceFont:    font.Font{Typeface: "monospace"},
		DefaultSize:      unit.Sp(15),
		DefaultColor:     body,
		InteractiveColor: link,
	}
	return &Renderer{parser: parser}
}

func NewCache(renderer *Renderer, capacity int) *Cache {
	if capacity < 1 {
		capacity = 1
	}
	return &Cache{renderer: renderer, capacity: capacity, entries: make(map[int]*list.Element), order: list.New()}
}

func (c *Cache) Render(index int, source string) ([]Span, error) {
	if element := c.entries[index]; element != nil {
		entry := element.Value.(cacheEntry)
		if entry.source == source {
			c.order.MoveToFront(element)
			return entry.spans, nil
		}
		c.order.Remove(element)
		delete(c.entries, index)
	}
	spans, err := c.renderer.Render(source)
	if err != nil {
		return nil, err
	}
	entry := cacheEntry{index: index, source: source, spans: spans}
	c.entries[index] = c.order.PushFront(entry)
	if c.order.Len() > c.capacity {
		last := c.order.Back()
		delete(c.entries, last.Value.(cacheEntry).index)
		c.order.Remove(last)
	}
	return spans, nil
}

func (r *Renderer) Render(source string) ([]Span, error) {
	spans, err := r.parser.Render([]byte(preprocessDiscord(source)))
	if err != nil {
		return nil, err
	}
	return applyDiscordDecorations(spans), nil
}

func PlainText(spans []Span) string {
	var text strings.Builder
	for _, span := range spans {
		text.WriteString(span.Style.Content)
	}
	return text.String()
}

func RichStyles(spans []Span) []richtext.SpanStyle {
	styles := make([]richtext.SpanStyle, len(spans))
	for i := range spans {
		styles[i] = spans[i].Style
	}
	return styles
}
