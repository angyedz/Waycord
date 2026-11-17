package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

// slidingMembers only lays out the member list while it is visible or moving.
func (a *App) slidingMembers(progress float32) layout.Widget {
	return fixedWidth(unit.Dp(210*progress), func(gtx layout.Context) layout.Dimensions {
		visible, full := gtx.Constraints.Max.X, gtx.Dp(210)
		area := clip.Rect{Max: image.Pt(visible, gtx.Constraints.Max.Y)}.Push(gtx.Ops)
		offset := op.Offset(image.Pt(visible-full, 0)).Push(gtx.Ops)
		child := gtx
		child.Constraints.Min.X, child.Constraints.Max.X = full, full
		a.membersPanel(child)
		offset.Pop()
		area.Pop()
		return layout.Dimensions{Size: image.Pt(visible, gtx.Constraints.Max.Y)}
	})
}

func (a *App) membersPanel(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 4, Right: 14, Bottom: 14}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Sidebar, 16, padding(layout.Inset{Top: 16, Left: 14, Right: 14, Bottom: 16}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(a.text("ONLINE — 3", 11, a.palette.Subtle)),
				vspace(12),
				layout.Rigid(a.memberRow("mira")), vspace(10),
				layout.Rigid(a.memberRow("angyedz")), vspace(10),
				layout.Rigid(a.memberRow("k66")),
			)
		}))
	})
}

func (a *App) memberRow(name string) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(fixedWidth(30, fixedHeight(30, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.AccentDim, 15, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, a.text(firstLetter(name), 13, a.palette.Text))
				})
			}))),
			hspace(9),
			layout.Flexed(1, a.text(name, 14, a.palette.Muted)),
		)
	}
}
