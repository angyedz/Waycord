package markdown

import (
	"strings"
	"unicode/utf8"

	"gioui.org/x/richtext"
)

const (
	strikeMarker     rune = '\ue000'
	underlineMarker  rune = '\ue001'
	unmatchedStrike  rune = '\ue002'
	unmatchedUnder   rune = '\ue003'
	spoilerMarker    rune = '\ue004'
	unmatchedSpoiler rune = '\ue005'
)

const SpoilerKey = "waycord:spoiler"

func preprocessDiscord(source string) string {
	var out []rune
	open := make(map[rune]int, 3)
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
			if strings.HasPrefix(source[i+1:], "||") {
				out = append(out, '|', '|')
				i += 3
				continue
			}
			out = append(out, '\\')
			i++
			if i < len(source) && (strings.HasPrefix(source[i:], "~~") || strings.HasPrefix(source[i:], "__") || strings.HasPrefix(source[i:], "||")) {
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
		case strings.HasPrefix(source[i:], "||"):
			marker, delimiter = spoilerMarker, "||"
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
		} else if marker == underlineMarker {
			unmatched[index] = string(unmatchedUnder)
		} else {
			unmatched[index] = string(unmatchedSpoiler)
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
	Spoiler   bool
	SpoilerID int
}

func applyDiscordDecorations(spans []richtext.SpanStyle) []Span {
	active := make(map[rune]bool, 2)
	var result []Span
	activeSpoiler := -1
	nextSpoilerID := 0
	for _, span := range spans {
		start := 0
		for offset, marker := range span.Content {
			if marker != strikeMarker && marker != underlineMarker && marker != unmatchedStrike && marker != unmatchedUnder && marker != spoilerMarker && marker != unmatchedSpoiler {
				continue
			}
			appendStyledSpan(&result, span, span.Content[start:offset], active, activeSpoiler)
			switch marker {
			case strikeMarker, underlineMarker:
				active[marker] = !active[marker]
			case spoilerMarker:
				if activeSpoiler >= 0 {
					activeSpoiler = -1
				} else {
					activeSpoiler = nextSpoilerID
					nextSpoilerID++
				}
			case unmatchedStrike:
				appendStyledSpan(&result, span, "~~", active, activeSpoiler)
			case unmatchedUnder:
				appendStyledSpan(&result, span, "__", active, activeSpoiler)
			case unmatchedSpoiler:
				appendStyledSpan(&result, span, "||", active, activeSpoiler)
			}
			start = offset + utf8.RuneLen(marker)
		}
		appendStyledSpan(&result, span, span.Content[start:], active, activeSpoiler)
	}
	return result
}

func appendStyledSpan(result *[]Span, source richtext.SpanStyle, content string, active map[rune]bool, spoilerID int) {
	if content == "" {
		return
	}
	span := source.DeepCopy()
	span.Content = content
	spoiler := spoilerID >= 0
	if spoiler {
		span.Interactive = true
		span.Set(SpoilerKey, spoilerID)
	}
	*result = append(*result, Span{Style: span, Strike: active[strikeMarker], Underline: active[underlineMarker], Spoiler: spoiler, SpoilerID: spoilerID})
}
