package ui

import (
	"image"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) chat(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			visible := a.visibleMessages()
			if len(visible) == 0 {
				message := "No messages in this channel yet."
				if a.pinsOnly {
					message = "No pinned messages in this channel."
				} else if query := strings.TrimSpace(a.searchEditor.Text()); query != "" {
					message = "No messages match “" + query + "”."
				}
				return layout.Inset{Top: 24, Left: 22, Right: 22}.Layout(gtx, a.text(message, 14, a.palette.Subtle))
			}
			return layout.Inset{Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return a.messageList.Layout(gtx, len(visible), func(gtx layout.Context, i int) layout.Dimensions {
					grouped := i > 0 && messagesGroup(a.messages[visible[i-1]], a.messages[visible[i]])
					bottom := float32(13)
					if a.compactMessages {
						bottom = 7
					}
					if i+1 < len(visible) && messagesGroup(a.messages[visible[i]], a.messages[visible[i+1]]) {
						bottom = 3
					}
					return layout.Inset{Bottom: unit.Dp(bottom)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.message(gtx, visible[i], a.messages[visible[i]], grouped)
					})
				})
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			progress := a.transition(gtx, &a.emojiSlide, boolProgress(a.emojiOpen), 220*time.Millisecond)
			return a.animatedEmojiPicker(gtx, progress)
		}),
		layout.Rigid(a.composerContextPreview),
		layout.Rigid(a.attachmentDraft),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Top: 8, Bottom: 14, Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 24, padding(layout.Inset{Left: 18, Right: 6, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.editTarget >= 0 {
								return layout.Dimensions{}
							}
							return a.iconControl(gtx, &a.attachButton, a.attachIcon, 32, 20, "Attach file")
						}),
						hspace(6),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							hint := "Write a message…"
							if a.editTarget >= 0 {
								hint = "Edit your message…"
							}
							input := material.Editor(a.theme, &a.composer, hint)
							input.TextSize = 14
							input.Color = a.palette.Text
							return input.Layout(gtx)
						}),
						hspace(8),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.editTarget >= 0 {
								return layout.Dimensions{}
							}
							return a.mediaTextControl(gtx, &a.gifButton, 1, "GIF", "Choose a GIF")
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.editTarget >= 0 {
								return layout.Dimensions{}
							}
							return a.mediaTextControl(gtx, &a.stickerButton, 2, "Sticker", "Choose a sticker")
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.editTarget >= 0 {
								return layout.Dimensions{}
							}
							return a.emojiControl(gtx)
						}),
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

func (a *App) message(gtx layout.Context, index int, m mock.Message, grouped bool) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	w := a.reactionWidget(index)
	dim := a.messageContentGrouped(gtx, index, m, grouped)
	if w.hovered || a.reactionTarget == index {
		toolbar := gtx
		toolbar.Constraints.Min = image.Point{}
		macro := op.Record(gtx.Ops)
		toolbarSize := a.reactionToolbar(toolbar, index).Size
		call := macro.Stop()
		offset := op.Offset(image.Pt(max(0, dim.Size.X-toolbarSize.X), 0)).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
	}
	a.messageHover(gtx, index, dim.Size)
	if a.contextShown == index {
		progress := a.transition(gtx, &a.contextSlide, boolProgress(a.contextTarget == index), 220*time.Millisecond)
		if progress == 0 && a.contextTarget != index {
			a.contextShown = -1
		}
		if progress <= 0 {
			return dim
		}
		position := a.contextPosition
		position.X = max(0, min(position.X, dim.Size.X-gtx.Dp(190)))
		menuHeight := unit.Dp(a.contextMenuHeight())
		position.Y = max(0, min(position.Y, dim.Size.Y-gtx.Dp(menuHeight)))
		macro := op.Record(gtx.Ops)
		offset := op.Offset(image.Pt(position.X, position.Y)).Push(gtx.Ops)
		menuContext := gtx
		menuContext.Constraints = layout.Constraints{Max: image.Pt(gtx.Dp(190), gtx.Dp(menuHeight))}
		animatedReveal(menuContext, progress, menuHeight, fixedHeight(menuHeight, a.contextMenu))
		offset.Pop()
		op.Defer(gtx.Ops, macro.Stop())
	}
	return dim
}

func (a *App) messageContent(gtx layout.Context, index int, m mock.Message) layout.Dimensions {
	return a.messageContentGrouped(gtx, index, m, false)
}

func messagesGroup(previous, current mock.Message) bool {
	if previous.Author != current.Author || previous.Direct != current.Direct || (!previous.Direct && (previous.Server != current.Server || previous.Channel != current.Channel)) || previous.SentAt.IsZero() || current.SentAt.IsZero() || current.SentAt.Before(previous.SentAt) {
		return false
	}
	return previous.SentAt.Truncate(time.Minute).Equal(current.SentAt.Truncate(time.Minute))
}

func (a *App) messageContentGrouped(gtx layout.Context, index int, m mock.Message, grouped bool) layout.Dimensions {
	return layout.Flex{Alignment: layout.Start}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if grouped {
				return layout.Dimensions{Size: image.Pt(gtx.Dp(38), 0)}
			}
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
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.messageReply(gtx, index, m) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if grouped {
						return layout.Dimensions{}
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(a.text(m.Author, 14, a.palette.Text)), hspace(10),
						layout.Rigid(a.text(m.Time, 11, a.palette.Subtle)),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if !m.Edited {
								return layout.Dimensions{}
							}
							return layout.Inset{Left: 6}.Layout(gtx, a.text("edited", 10, a.palette.Subtle))
						}),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if grouped {
						return layout.Dimensions{}
					}
					return layout.Spacer{Height: 4}.Layout(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if m.Body == "" {
						return layout.Dimensions{}
					}
					return a.messageBody(gtx, index, m.Body)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.attachmentCard(gtx, m.Attachment) }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.mediaCard(gtx, m.Media) }),
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
