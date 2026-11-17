package ui

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/widget"
)

type reactionWidget struct {
	add        widget.Clickable
	body       widget.Selectable
	contextTag int
	hover      gesture.Hover
	hovered    bool
	quick      [3]widget.Clickable
	choices    [len(emojiOptions)]widget.Clickable
	existing   [len(emojiOptions)]widget.Clickable
	picker     panelTransition
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
		w.hovered = w.hover.Update(gtx.Source)
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
		for i, option := range quickReactions {
			if w.quick[i].Clicked(gtx) {
				a.reactions.Toggle(index, emojiOptions[option])
				a.reactionTarget = -1
			}
		}
	}
}

var quickReactions = [...]int{5, 7, 4}

// messageHover registers the entire message row without intercepting its buttons.
func (a *App) messageHover(gtx layout.Context, index int, size image.Point) {
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := (pointer.PassOp{}).Push(gtx.Ops)
	a.reactionWidget(index).hover.Add(gtx.Ops)
	event.Op(gtx.Ops, &a.reactionWidget(index).contextTag)
	pass.Pop()
	area.Pop()
}

func (a *App) reactionToolbar(gtx layout.Context, index int) layout.Dimensions {
	w := a.reactionWidget(index)
	return rounded(gtx, a.palette.Surface2, 10, padding(layout.Inset{Left: 3, Right: 3, Top: 2, Bottom: 2}, func(gtx layout.Context) layout.Dimensions {
		var buttons []layout.FlexChild
		for i, option := range quickReactions {
			i, value := i, emojiOptions[option]
			buttons = append(buttons, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.emojiChoice(gtx, &w.quick[i], value, 29, 18)
			}))
		}
		buttons = append(buttons, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.reactionAdd(gtx, index) }))
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, buttons...)
	}))
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
	progress := w.picker.layoutProgress(gtx, a.reactionTarget == index)
	if !hasReaction && progress <= 0 {
		return layout.Dimensions{}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			var chips []layout.FlexChild
			for i, value := range emojiOptions {
				if !a.reactions.Has(index, value) {
					continue
				}
				i, value := i, value
				chips = append(chips, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.reactionChip(gtx, &w.existing[i], value)
				}), hspace(4))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, chips...)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return animatedReveal(gtx, progress, 42, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 4, Right: 4, Top: 2, Bottom: 2}, func(gtx layout.Context) layout.Dimensions {
						var choices []layout.FlexChild
						for i, value := range emojiOptions {
							i, value := i, value
							choices = append(choices, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return a.emojiChoice(gtx, &w.choices[i], value, 34, 19)
							}))
						}
						return layout.Flex{Alignment: layout.Middle}.Layout(gtx, choices...)
					}))
				})
			})
		}),
	)
}

func (a *App) reactionChip(gtx layout.Context, button *widget.Clickable, value string) layout.Dimensions {
	gtx.Constraints.Min = image.Point{}
	background := a.interactionColor(gtx, button, a.palette.AccentDim, a.palette.Surface2)
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, background, 12, padding(layout.Inset{Left: 7, Right: 8, Top: 3, Bottom: 3}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(a.text(value, 16, a.palette.Text)),
				hspace(4),
				layout.Rigid(a.text("1", 12, a.palette.Text)),
			)
		}))
	})
}

func (a *App) reactionAdd(gtx layout.Context, index int) layout.Dimensions {
	w := a.reactionWidget(index)
	return a.iconControl(gtx, &w.add, a.emojiIcon, 24, 16, "Add reaction")
}
