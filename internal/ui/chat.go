package ui

import (
	"gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) chat(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 5, Left: 24, Right: 24, Bottom: 18}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.AccentDim, 10, inset(12, a.text("DESIGN PREVIEW   ·   Messages are local fixtures, not Discord data", 12, a.palette.Muted)))
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			var visible []mock.Message
			for _, m := range a.messages {
				if m.Channel == a.activeChannel {
					visible = append(visible, m)
				}
			}
			return layout.Inset{Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.messageList.Layout(gtx, len(visible), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.Inset{Bottom: 14}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return a.message(gtx, visible[i]) })
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 12, Bottom: 22, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 14, Right: 10, Top: 9, Bottom: 9}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							input := material.Editor(a.theme, &a.composer, "Message in local preview…")
							input.TextSize = 14
							input.Color = a.palette.Text
							return input.Layout(gtx)
						}),
						hspace(12),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.pill(gtx, &a.sendButton, "Send preview", true)
						}),
					)
				}))
			})
		}),
	)
}

func (a *App) message(gtx layout.Context, m mock.Message) layout.Dimensions {
	return layout.Flex{Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			avatar := a.palette.Surface2
			if m.Self {
				avatar = a.palette.AccentDim
			}
			return fixedWidth(45, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, avatar, 12, inset(10, a.text(firstLetter(m.Author), 16, a.palette.Text)))
			})(gtx)
		}),
		hspace(10),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(a.text(m.Author, 14, a.palette.Text)), hspace(10),
						layout.Rigid(a.text(m.Time, 11, a.palette.Subtle)),
					)
				}),
				vspace(7),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Surface, 9, padding(layout.Inset{Left: 13, Right: 13, Top: 11, Bottom: 11},
						a.text(m.Body, 14, a.palette.Muted)))
				}),
			)
		}),
	)
}
func firstLetter(s string) string {
	for _, r := range s {
		return string(r)
	}
	return "?"
}
