// Package profile presents a local preview profile, independent of application services.
package profile

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
)

// View describes the content and colors of a profile card.
type View struct {
	Theme               *material.Theme
	Text, Muted, Accent color.NRGBA
}

func (v View) label(text string, size unit.Sp, color color.NRGBA) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		label := material.Label(v.Theme, size, text)
		label.Color = color
		return label.Layout(gtx)
	}
}

// Layout renders fixture content; it does not connect to a user account.
func (v View) Layout(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 6, Right: 18, Bottom: 18}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(v.label("COMMUNITY", 10, v.Muted)),
			layout.Rigid(layout.Spacer{Height: 12}.Layout),
			layout.Rigid(v.label("Waycord studio", 19, v.Text)),
			layout.Rigid(layout.Spacer{Height: 6}.Layout),
			layout.Rigid(v.label("A small space to build something better.", 13, v.Muted)),
			layout.Rigid(layout.Spacer{Height: 24}.Layout),
			layout.Rigid(v.label("●  ONLINE IN PREVIEW", 10, v.Accent)),
			layout.Rigid(layout.Spacer{Height: 16}.Layout),
			layout.Rigid(v.label("mira", 14, v.Text)),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(v.label("Design & conversation", 12, v.Muted)),
			layout.Rigid(layout.Spacer{Height: 18}.Layout),
			layout.Rigid(v.label("angyedz", 14, v.Text)),
			layout.Rigid(layout.Spacer{Height: 5}.Layout),
			layout.Rigid(v.label("Building Waycord", 12, v.Muted)),
			layout.Rigid(layout.Spacer{Height: 24}.Layout),
			layout.Rigid(v.label("OFFLINE DEMO", 10, v.Muted)),
			layout.Rigid(layout.Spacer{Height: 7}.Layout),
			layout.Rigid(v.label("People and messages here are sample data.", 12, v.Muted)),
		)
	})
}
