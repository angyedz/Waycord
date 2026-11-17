package ui

import (
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"
)

type reactionWidget struct {
	add      widget.Clickable
	choices  [len(emojiOptions)]widget.Clickable
	existing [len(emojiOptions)]widget.Clickable
}

func (a *App) reactionWidget(index int) *reactionWidget {
	w := a.reactionWidgets[index]
	if w == nil {
		w = &reactionWidget{}
		a.reactionWidgets[index] = w
	}
	return w
}

func (a *App) processReactions(gtx layout.Context) {
	for index, w := range a.reactionWidgets {
		if w.add.Clicked(gtx) {
			if a.reactionTarget == index {
				a.reactionTarget = -1
			} else {
				a.reactionTarget = index
			}
		}
		for i, value := range emojiOptions {
			choice := w.choices[i].Clicked(gtx)
			existing := w.existing[i].Clicked(gtx)
			if choice || existing {
				a.reactions.Toggle(index, value)
				a.reactionTarget = -1
			}
		}
	}
}

func (a *App) reactionRow(gtx layout.Context, index int) layout.Dimensions {
	w := a.reactionWidget(index)
	hasReaction := false
	for _, value := range emojiOptions {
		if a.reactions.Has(index, value) {
			hasReaction = true
			break
		}
	}
	if !hasReaction && a.reactionTarget != index {
		return layout.Dimensions{}
	}
	return layout.Inset{Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				var chips []layout.FlexChild
				for i, value := range emojiOptions {
					if !a.reactions.Has(index, value) {
						continue
					}
					i, value := i, value
					chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &w.existing[i], value+"  1", true)
					}), hspace(4))
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, chips...)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if a.reactionTarget != index {
					return layout.Dimensions{}
				}
				return layout.Inset{Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 4, Right: 4, Top: 2, Bottom: 2}, func(gtx layout.Context) layout.Dimensions {
						var choices []layout.FlexChild
						for i, value := range emojiOptions {
							i, value := i, value
							choices = append(choices, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return w.choices[i].Layout(gtx, fixedWidth(34, fixedHeight(34, func(gtx layout.Context) layout.Dimensions {
									return layout.Center.Layout(gtx, a.text(value, 19, a.palette.Text))
								})))
							}))
						}
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx, choices...)
					}))
				})
			}),
		)
	})
}

func (a *App) reactionAdd(gtx layout.Context, index int) layout.Dimensions {
	w := a.reactionWidget(index)
	return w.add.Layout(gtx, fixedWidth(24, fixedHeight(24, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp("Add reaction").Add(gtx.Ops)
		return layout.Center.Layout(gtx, fixedWidth(16, func(gtx layout.Context) layout.Dimensions { return a.emojiIcon.Layout(gtx, a.palette.Subtle) }))
	})))
}
