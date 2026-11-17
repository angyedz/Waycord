package ui

import (
	"gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) chat(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 3, Left: 26, Right: 26, Bottom: 22}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.text("OFFLINE PREVIEW   /   Messages are temporary and local", 11, a.palette.Subtle)(gtx)
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
					return layout.Inset{Bottom: 22}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return a.message(gtx, visible[i]) })
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 12, Bottom: 22, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 8, padding(layout.Inset{Left: 16, Right: 10, Top: 9, Bottom: 9}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							input := material.Editor(a.theme, &a.composer, "Message in local preview…")
							input.TextSize = 14
							input.Color = a.palette.Text
							return input.Layout(gtx)
						}),
						hspace(12),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.pill(gtx, &a.sendButton, "Send", true)
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
			return fixedWidth(38, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, avatar, 19, inset(9, a.text(firstLetter(m.Author), 15, a.palette.Text)))
			})(gtx)
		}),
		hspace(14),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(a.text(m.Author, 14, a.palette.Text)), hspace(10),
						layout.Rigid(a.text(m.Time, 11, a.palette.Subtle)),
					)
				}),
				vspace(4),
				layout.Rigid(a.text(m.Body, 14, a.palette.Muted)),
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
