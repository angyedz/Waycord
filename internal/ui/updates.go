package ui

import (
	"gioui.org/layout"

	"github.com/angyedz/Waycord/internal/version"
)

func (a *App) updates(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 24, Bottom: 24, Left: 30, Right: 30}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(a.text("Updates & recovery", 26, a.palette.Text)),
			vspace(10),
			layout.Rigid(a.text("Signed releases. Atomic version slots. Automatic rollback on failed startup.", 14, a.palette.Muted)),
			vspace(26),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 16, inset(22, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("CURRENT BUILD", 11, a.palette.Subtle)), vspace(11),
						layout.Rigid(a.text("Waycord  "+version.Version, 21, a.palette.Text)), vspace(9),
						layout.Rigid(a.text("Development build · No signed update feed configured", 13, a.palette.Warning)),
						vspace(21), layout.Rigid(a.text("Release channel (design preview)", 14, a.palette.Text)), vspace(12),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return a.pill(gtx, &a.channelOptions[0], "Stable", a.updateChannel == 0)
								}),
								hspace(8),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return a.pill(gtx, &a.channelOptions[1], "Beta", a.updateChannel == 1)
								}),
								hspace(8),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return a.pill(gtx, &a.channelOptions[2], "Nightly", a.updateChannel == 2)
								}),
							)
						}),
					)
				}))
			}),
			vspace(18),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Sidebar, 14, inset(20, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("How OTA works", 17, a.palette.Text)), vspace(10),
						layout.Rigid(a.text("01   Verify Ed25519 manifest signature", 13, a.palette.Muted)), vspace(7),
						layout.Rigid(a.text("02   Download and SHA-256 verify the binary", 13, a.palette.Muted)), vspace(7),
						layout.Rigid(a.text("03   Stage in a separate version slot", 13, a.palette.Muted)), vspace(7),
						layout.Rigid(a.text("04   Launch; promote only after the first frame", 13, a.palette.Muted)), vspace(7),
						layout.Rigid(a.text("05   Revert to the previous build if startup fails", 13, a.palette.Muted)),
					)
				}))
			}),
			vspace(20),
			layout.Rigid(a.text("Installation & update checks are currently available through waycord-updater CLI.", 12, a.palette.Subtle)),
		)
	})
}
