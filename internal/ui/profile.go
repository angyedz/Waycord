package ui

import (
	"gioui.org/layout"

	"github.com/angyedz/Waycord/internal/ui/profile"
)

func (a *App) profile(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 6, Right: 14, Bottom: 16, Left: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Sidebar, 24, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.AccentDim, 24, fixedHeight(92, inset(18, a.text("W / COMMUNITY", 11, a.palette.Text))))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: 16, Top: 12}.Layout(gtx, profile.View{Theme: a.theme, Text: a.palette.Text, Muted: a.palette.Muted, Accent: a.palette.Positive}.Layout)
				}),
			)
		})
	})
}
