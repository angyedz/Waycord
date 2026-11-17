package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"
)

func (a *App) railButton(gtx layout.Context, button *widget.Clickable, label string, selected bool) layout.Dimensions {
	background, text := a.palette.Surface, a.palette.Muted
	if selected {
		background, text = a.palette.AccentDim, a.palette.Text
	}
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
