package ui

import (
	"gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) chat(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 2, Left: 22, Right: 22, Bottom: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.text("Preview · Offline", 11, a.palette.Subtle)(gtx)
			})
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			var visible []int
			for index, m := range a.messages {
				if m.Direct == a.direct && (a.direct || m.Channel == a.activeChannel) {
					visible = append(visible, index)
				}
			}
			return layout.Inset{Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.messageList.Layout(gtx, len(visible), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.Inset{Bottom: 13}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return a.message(gtx, visible[i], a.messages[visible[i]]) })
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !a.emojiOpen {
				return layout.Dimensions{}
			}
			return a.emojiPicker(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 14, Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 24, padding(layout.Inset{Left: 18, Right: 6, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							input := material.Editor(a.theme, &a.composer, "Write a message…")
							input.TextSize = 14
							input.Color = a.palette.Text
							return input.Layout(gtx)
						}),
						hspace(8),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.emojiControl(gtx) }),
						hspace(4),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.sendControl(gtx)
						}),
					)
				}))
			})
		}),
	)
}

func (a *App) message(gtx layout.Context, index int, m mock.Message) layout.Dimensions {
	return layout.Flex{Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			avatar := a.palette.Surface2
			if m.Self {
				avatar = a.palette.AccentDim
			}
			return fixedWidth(38, fixedHeight(38, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, avatar, 19, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, a.text(firstLetter(m.Author), 15, a.palette.Text))
				})
			}))(gtx)
		}),
		hspace(14),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(a.text(m.Author, 14, a.palette.Text)), hspace(10),
						layout.Rigid(a.text(m.Time, 11, a.palette.Subtle)),
						hspace(8), layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.reactionAdd(gtx, index) }),
					)
				}),
				vspace(4),
				layout.Rigid(a.text(m.Body, 14, a.palette.Muted)),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.reactionRow(gtx, index) }),
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
