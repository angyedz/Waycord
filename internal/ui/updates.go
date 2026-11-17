package ui

import (
	"gioui.org/layout"

	"github.com/angyedz/Waycord/internal/version"
)

func (a *App) updates(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 16, Bottom: 16, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(a.text("Updates", 24, a.palette.Text)),
			vspace(16),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 16, inset(18, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("Waycord  "+version.Version, 17, a.palette.Text)),
						vspace(8),
						layout.Rigid(a.text("Update checks are unavailable in this preview.", 13, a.palette.Muted)),
					)
				}))
			}),
		)
	})
}
