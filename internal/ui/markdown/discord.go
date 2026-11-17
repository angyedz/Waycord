package markdown

import (
	"strings"
	"unicode/utf8"

	"gioui.org/x/richtext"
)

const (
	strikeMarker    rune = '\ue000'
	underlineMarker rune = '\ue001'
	unmatchedStrike rune = '\ue002'
	unmatchedUnder  rune = '\ue003'
)

func preprocessDiscord(source string) string {
	var out []rune
	open := make(map[rune]int, 2)
	inlineTicks := 0
	fenced := false
	for i := 0; i < len(source); {
		if strings.HasPrefix(source[i:], "```") {
			fenced = !fenced
			out = append(out, '`', '`', '`')
			i += 3
			continue
		}
		if source[i] == '`' {
			ticks := 1
			for i+ticks < len(source) && source[i+ticks] == '`' {
				ticks++
			}
			if !fenced {
				if inlineTicks == 0 {
					inlineTicks = ticks
				} else if inlineTicks == ticks {
					inlineTicks = 0
				}
			}
			for range ticks {
				out = append(out, '`')
			}
			i += ticks
			continue
		}
		if fenced || inlineTicks > 0 {
			r, size := utf8.DecodeRuneInString(source[i:])
			out = append(out, r)
			i += size
			continue
		}
		if source[i] == '\\' {
			out = append(out, '\\')
			i++
			if i < len(source) && (strings.HasPrefix(source[i:], "~~") || strings.HasPrefix(source[i:], "__")) {
				out = append(out, rune(source[i]), rune(source[i+1]))
				i += 2
			}
			continue
		}
		var marker rune
		var delimiter string
		switch {
		case strings.HasPrefix(source[i:], "~~"):
			marker, delimiter = strikeMarker, "~~"
		case strings.HasPrefix(source[i:], "__"):
			marker, delimiter = underlineMarker, "__"
		}
		if marker != 0 {
			if _, exists := open[marker]; exists {
				out = append(out, marker)
				delete(open, marker)
			} else {
				open[marker] = len(out)
				out = append(out, marker)
			}
			i += len(delimiter)
			continue
		}
		r, size := utf8.DecodeRuneInString(source[i:])
		out = append(out, r)
		i += size
	}
	unmatched := make(map[int]string, len(open))
	for marker, index := range open {
		if marker == strikeMarker {
			unmatched[index] = string(unmatchedStrike)
		} else {
			unmatched[index] = string(unmatchedUnder)
		}
	}
	var result strings.Builder
	for index, char := range out {
		if delimiter, exists := unmatched[index]; exists {
			result.WriteString(delimiter)
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

type Span struct {
	Style     richtext.SpanStyle
	Strike    bool
	Underline bool
}

func applyDiscordDecorations(spans []richtext.SpanStyle) []Span {
	active := make(map[rune]bool, 2)
	var result []Span
	for _, span := range spans {
		start := 0
		for offset, marker := range span.Content {
			if marker != strikeMarker && marker != underlineMarker && marker != unmatchedStrike && marker != unmatchedUnder {
				continue
			}
			appendStyledSpan(&result, span, span.Content[start:offset], active)
			switch marker {
			case strikeMarker, underlineMarker:
				active[marker] = !active[marker]
			case unmatchedStrike:
				appendStyledSpan(&result, span, "~~", active)
			case unmatchedUnder:
				appendStyledSpan(&result, span, "__", active)
			}
			start = offset + utf8.RuneLen(marker)
		}
		appendStyledSpan(&result, span, span.Content[start:], active)
	}
	return result
}

func appendStyledSpan(result *[]Span, source richtext.SpanStyle, content string, active map[rune]bool) {
	if content == "" {
		return
	}
	span := source.DeepCopy()
	span.Content = content
	*result = append(*result, Span{Style: span, Strike: active[strikeMarker], Underline: active[underlineMarker]})
}
