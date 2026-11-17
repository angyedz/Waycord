package ui

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget/material"
)

func (a *App) beginEdit(gtx layout.Context, index int) bool {
	if index < 0 || index >= len(a.messages) || !canEditMessage(a.messages[index]) {
		return false
	}
	if a.editTarget >= 0 {
		a.cancelEdit()
	}
	draftKey := a.conversationKey()
	a.composerDrafts[draftKey] = a.composer.Text()
	a.editTarget, a.editDraftKey = index, draftKey
	a.composer.SetText(a.messages[index].Body)
	a.contextTarget, a.reactionTarget = -1, -1
	a.emojiOpen = false
	gtx.Execute(key.FocusCmd{Tag: &a.composer})
	return true
}

func (a *App) cancelEdit() {
	if a.editTarget < 0 {
		return
	}
	draft := a.composerDrafts[a.editDraftKey]
	a.composer.SetText(draft)
	if draft == "" {
		delete(a.composerDrafts, a.editDraftKey)
	}
	a.editTarget, a.editDraftKey = -1, ""
}

func (a *App) requestDelete(index int) bool {
	if index < 0 || index >= len(a.messages) || !a.messages[index].Self {
		return false
	}
	a.deleteTarget = index
	a.deleteDialogOpen = true
	a.contextTarget, a.reactionTarget = -1, -1
	a.emojiOpen = false
	return true
}

func (a *App) cancelDelete() {
	a.deleteDialogOpen = false
	a.deleteTarget = -1
}

func (a *App) confirmDelete() {
	if a.deleteDialogOpen {
		a.deleteMessage(a.deleteTarget)
	}
	a.cancelDelete()
}

func (a *App) deleteMessage(index int) bool {
	if index < 0 || index >= len(a.messages) || !a.messages[index].Self {
		return false
	}
	if a.editTarget == index {
		a.cancelEdit()
	} else if a.editTarget > index {
		a.editTarget--
	}
	if a.replyTarget == index {
		a.replyTarget = -1
	} else if a.replyTarget > index {
		a.replyTarget--
	}
	a.messages = append(a.messages[:index], a.messages[index+1:]...)
	for message := range a.messages {
		if !a.messages[message].HasReply {
			continue
		}
		switch {
		case a.messages[message].ReplyTo == index:
			a.messages[message].HasReply = false
			a.messages[message].ReplyTo = 0
		case a.messages[message].ReplyTo > index:
			a.messages[message].ReplyTo--
		}
	}
	a.reactions.RemoveAndShift(index)
	a.reactionWidgets = make(map[int]*reactionWidget)
	a.contextTarget, a.contextShown, a.reactionTarget = -1, -1, -1
	return true
}

// processDeleteConfirmation blocks other UI input while the destructive
// action is pending and returns true for the entire modal frame.
func (a *App) processDeleteConfirmation(gtx layout.Context) bool {
	if !a.deleteDialogOpen {
		return false
	}
	if a.deleteConfirm.Clicked(gtx) {
		a.confirmDelete()
	} else if a.deleteCancel.Clicked(gtx) || a.deleteBackdrop.Clicked(gtx) {
		a.cancelDelete()
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if event, ok := e.(key.Event); ok && event.State == key.Press {
			a.cancelDelete()
		}
	}
	return true
}

func (a *App) deleteConfirmation(gtx layout.Context, size image.Point) {
	backdrop := gtx
	backdrop.Constraints = layout.Exact(size)
	a.deleteBackdrop.Layout(backdrop, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{A: 150}, clip.Rect{Max: size}.Op())
		return layout.Dimensions{Size: size}
	})
	modal := gtx
	modal.Constraints = layout.Constraints{Max: size}
	layout.Center.Layout(modal, func(gtx layout.Context) layout.Dimensions {
		return fixedWidth(360, func(gtx layout.Context) layout.Dimensions {
			return rounded(gtx, a.palette.Surface, 16, padding(layout.Inset{Left: 18, Right: 18, Top: 16, Bottom: 14}, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(a.text("Delete message?", 18, a.palette.Text)),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						label := "This message will be deleted."
						if a.deleteTarget >= 0 && a.deleteTarget < len(a.messages) {
							label = "“" + messageExcerpt(a.messages[a.deleteTarget].Body, 88) + "” will be deleted."
						}
						return a.text(label, 13, a.palette.Muted)(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Inset{Top: 18}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return a.pill(gtx, &a.deleteCancel, "Cancel", false)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									button := material.Button(a.theme, &a.deleteConfirm, "Delete")
									button.Background, button.Color = a.palette.Warning, a.palette.Rail
									return button.Layout(gtx)
								}),
							)
						})
					}),
				)
			}))
		})(gtx)
	})
}

func (a *App) composerContextPreview(gtx layout.Context) layout.Dimensions {
	if a.editTarget < 0 || a.editTarget >= len(a.messages) {
		return a.replyPreview(gtx)
	}
	message := a.messages[a.editTarget]
	return layout.Inset{Left: 20, Right: 20, Bottom: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 10, padding(layout.Inset{Left: 12, Right: 6, Top: 6, Bottom: 6}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.text("Editing · "+strings.TrimSpace(messageExcerpt(message.Body, 72)), 12, a.palette.Muted)(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.pill(gtx, &a.editCancelButton, "Cancel", false)
				}),
			)
		}))
	})
}
