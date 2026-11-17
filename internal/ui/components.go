package ui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func rounded(gtx layout.Context, bg color.NRGBA, radius unit.Dp, child layout.Widget) layout.Dimensions {
	return layout.Background{}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, gtx.Dp(radius)).Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, bg)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}, child)
}
func inset(all unit.Dp, child layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions { return layout.UniformInset(all).Layout(gtx, child) }
}
func padding(p layout.Inset, child layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions { return p.Layout(gtx, child) }
}
func hspace(x unit.Dp) layout.FlexChild { return layout.Rigid(layout.Spacer{Width: x}.Layout) }
func vspace(x unit.Dp) layout.FlexChild { return layout.Rigid(layout.Spacer{Height: x}.Layout) }
func fixedWidth(x unit.Dp, child layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		w := gtx.Dp(x)
		if w > gtx.Constraints.Max.X {
			w = gtx.Constraints.Max.X
		}
		gtx.Constraints.Min.X = w
		gtx.Constraints.Max.X = w
		return child(gtx)
	}
}
func fixedHeight(x unit.Dp, child layout.Widget) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		h := gtx.Dp(x)
		if h > gtx.Constraints.Max.Y {
			h = gtx.Constraints.Max.Y
		}
		gtx.Constraints.Min.Y = h
		gtx.Constraints.Max.Y = h
		return child(gtx)
	}
}
func (a *App) text(value string, size unit.Sp, col color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		l := material.Label(a.theme, size, value)
		l.Color = col
		return l.Layout(gtx)
	}
}
func (a *App) clickLabel(gtx layout.Context, b *widget.Clickable, title string, selected bool) layout.Dimensions {
	bg := a.palette.Sidebar
	fg := a.palette.Muted
	if selected {
		bg = a.palette.AccentDim
		fg = a.palette.Text
	}
	return b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, bg, 10, inset(12, a.text(title, 14, fg)))
	})
}
func (a *App) pill(gtx layout.Context, b *widget.Clickable, title string, active bool) layout.Dimensions {
	bg := a.palette.Surface2
	fg := a.palette.Muted
	if active {
		bg = a.palette.AccentDim
		fg = a.palette.Text
	}
	return b.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, bg, 10, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 8, Left: 14, Right: 14}.Layout(gtx, a.text(title, 13, fg))
		})
	})
}
