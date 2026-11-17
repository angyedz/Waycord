package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
)

// slidingProfile keeps the contents at their normal width while the viewport
// and chat width transition together.
func (a *App) slidingProfile(progress float32) layout.Widget {
	return fixedWidth(unit.Dp(210*progress), func(gtx layout.Context) layout.Dimensions {
		visible := gtx.Constraints.Max.X
		full := gtx.Dp(210)
		area := clip.Rect{Max: image.Pt(visible, gtx.Constraints.Max.Y)}.Push(gtx.Ops)
		offset := op.Offset(image.Pt(visible-full, 0)).Push(gtx.Ops)
		child := gtx
		child.Constraints.Min.X, child.Constraints.Max.X = full, full
		a.directProfile(child)
		offset.Pop()
		area.Pop()
		return layout.Dimensions{Size: image.Pt(visible, gtx.Constraints.Max.Y)}
	})
}

func (a *App) railButton(gtx layout.Context, button *widget.Clickable, label string, selected bool) layout.Dimensions {
	background, text := a.palette.Surface, a.palette.Muted
	if selected {
		background, text = a.palette.AccentDim, a.palette.Text
	}
	background = a.interactionColor(gtx, button, background, a.palette.Surface2)
	return button.Layout(gtx, fixedWidth(42, fixedHeight(42, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, background, 16, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, a.text(label, 17, text))
		})
	})))
}

// directProfile is shown only for the currently selected direct conversation.
func (a *App) directProfile(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	return layout.Inset{Top: 4, Right: 14}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Sidebar, 18, inset(16, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(a.text("mira", 19, a.palette.Text)),
				vspace(5),
				layout.Rigid(a.text("●  Online", 12, a.palette.Positive)),
			)
		}))
	})
}
