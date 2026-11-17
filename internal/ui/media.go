package ui

import (
	"gioui.org/layout"
	"gioui.org/unit"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) mediaCard(gtx layout.Context, media *mock.Media) layout.Dimensions {
	if media == nil {
		return layout.Dimensions{}
	}
	if media.Kind == mock.MediaSticker {
		return layout.Inset{Top: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return fixedWidth(126, fixedHeight(106, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 14, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Center.Layout(gtx, a.text(media.Emoji, 43, a.palette.Text))
						}),
						layout.Rigid(a.text(media.Name, 11, a.palette.Muted)),
					)
				})
			}))(gtx)
		})
	}
	if media.Kind != mock.MediaGIF {
		return layout.Dimensions{}
	}
	return layout.Inset{Top: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return fixedWidth(238, func(gtx layout.Context) layout.Dimensions {
			return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 8, Right: 8, Top: 7, Bottom: 7}, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return fixedHeight(unit.Dp(76), func(gtx layout.Context) layout.Dimensions {
							return rounded(gtx, a.palette.AccentDim, 8, func(gtx layout.Context) layout.Dimensions {
								return layout.Center.Layout(gtx, a.text(media.Emoji, 42, a.palette.Text))
							})
						})(gtx)
					}),
					vspace(5),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return rounded(gtx, a.palette.Accent, 4, padding(layout.Inset{Left: 5, Right: 5, Top: 2, Bottom: 2}, a.text("GIF", 9, a.palette.Text)))
							}),
							hspace(7),
							layout.Rigid(a.text(media.Name, 12, a.palette.Muted)),
						)
					}),
				)
			}))
		})(gtx)
	})
}
