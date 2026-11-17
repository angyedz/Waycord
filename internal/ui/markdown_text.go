package ui

import (
	"image"
	"image/color"
	"log"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
	"gioui.org/x/richtext"
	"gioui.org/x/styledtext"

	"github.com/angyedz/Waycord/internal/platform"
	uimarkdown "github.com/angyedz/Waycord/internal/ui/markdown"
)

func (a *App) resetMessageTextState(index int) {
	if widget := a.reactionWidgets[index]; widget != nil {
		widget.links = richtext.InteractiveText{}
		widget.spoilers = nil
	}
}

func (a *App) messageBody(gtx layout.Context, index int, source string) layout.Dimensions {
	state := a.reactionWidget(index)
	spans, err := a.messageMarkdown.Render(index, source)
	if err != nil {
		label := material.Label(a.theme, 15, source)
		label.Color = a.palette.Muted
		label.State = &state.body
		return label.Layout(gtx)
	}
	plain := uimarkdown.PlainText(spans)
	if plain == "" {
		plain = source
	}
	for {
		span, event, ok := state.links.Update(gtx)
		if !ok {
			break
		}
		if event.Type == richtext.Click {
			if url, ok := span.Get(uimarkdown.URLKey).(string); ok {
				if err := platform.OpenURL(url); err != nil {
					log.Print("open message link: ", err)
				}
			} else if spoilerID, ok := span.Get(uimarkdown.SpoilerKey).(int); ok {
				if state.spoilers == nil {
					state.spoilers = make(map[int]bool)
				}
				state.spoilers[spoilerID] = !state.spoilers[spoilerID]
			}
		}
	}
	styled := uimarkdown.DisplayStyles(spans, state.spoilers, a.palette.Subtle)
	formatted := richtext.Text(&state.links, a.theme.Shaper, styled...).Layout(gtx)
	decorations := make([]styledtext.SpanStyle, len(styled))
	for index, span := range styled {
		decorations[index] = styledtext.SpanStyle{Font: span.Font, Size: span.Size, Color: color.NRGBA{}, Content: span.Content}
	}
	styledtext.Text(a.theme.Shaper, decorations...).Layout(gtx, func(gtx layout.Context, index int, dims layout.Dimensions) {
		if index >= len(spans) || (!spans[index].Underline && !spans[index].Strike) {
			return
		}
		y := dims.Baseline + gtx.Dp(1)
		if spans[index].Strike {
			y = dims.Baseline * 3 / 5
		}
		thickness := max(1, gtx.Dp(1))
		paint.FillShape(gtx.Ops, spans[index].Style.Color, clip.Rect{Min: image.Pt(0, y), Max: image.Pt(dims.Size.X, y+thickness)}.Op())
	})

	// PassOp lets rich links and the selectable layer receive the same pointer
	// gesture. The selectable layer paints no glyphs, only the selection.
	pass := (pointer.PassOp{}).Push(gtx.Ops)
	selectable := material.Label(a.theme, 15, plain)
	selectable.Color = color.NRGBA{}
	selectable.SelectionColor = color.NRGBA{R: a.palette.Accent.R, G: a.palette.Accent.G, B: a.palette.Accent.B, A: 88}
	selectable.State = &state.body
	state.body.PaintSelectionWhenUnfocused = true
	selectionContext := gtx
	selectionContext.Constraints = layout.Exact(formatted.Size)
	area := clip.Rect{Max: formatted.Size}.Push(gtx.Ops)
	selectable.Layout(selectionContext)
	area.Pop()
	pass.Pop()
	return formatted
}
