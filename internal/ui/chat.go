package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"
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
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	w := a.reactionWidget(index)
	dim := layout.Stack{Alignment: layout.NE}.Layout(gtx,
		layout.Stacked(func(gtx layout.Context) layout.Dimensions { return a.messageContent(gtx, index, m) }),
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			if !w.hovered && a.reactionTarget != index {
				return layout.Dimensions{Size: gtx.Constraints.Min}
			}
			return layout.NE.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X, gtx.Constraints.Min.Y = 0, 0
				return a.reactionToolbar(gtx, index)
			})
		}),
	)
	a.messageHover(gtx, index, dim.Size)
	if a.contextTarget == index {
		position := a.contextPosition
		position.X = max(0, min(position.X, dim.Size.X-gtx.Dp(190)))
		position.Y = max(0, position.Y)
		macro := op.Record(gtx.Ops)
		offset := op.Offset(image.Pt(position.X, position.Y)).Push(gtx.Ops)
		menuContext := gtx
		menuContext.Constraints = layout.Constraints{Max: image.Pt(gtx.Dp(190), gtx.Dp(100))}
		a.contextMenu(menuContext)
		offset.Pop()
		op.Defer(gtx.Ops, macro.Stop())
	}
	return dim
}

func (a *App) messageContent(gtx layout.Context, index int, m mock.Message) layout.Dimensions {
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
					)
				}),
				vspace(4),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := material.Label(a.theme, 15, m.Body)
					label.Color = a.palette.Muted
					label.State = &a.reactionWidget(index).body
					return label.Layout(gtx)
				}),
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
