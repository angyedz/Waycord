package ui

import (
	"strings"

	"gioui.org/layout"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) replyPreview(gtx layout.Context) layout.Dimensions {
	if a.replyTarget < 0 || a.replyTarget >= len(a.messages) {
		return layout.Dimensions{}
	}
	replied := a.messages[a.replyTarget]
	return layout.Inset{Left: 20, Right: 20, Bottom: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 12, Right: 6, Top: 6, Bottom: 6}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(fixedWidth(2, fixedHeight(32, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Accent, 1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
				}))),
				hspace(9),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("Replying to "+replied.Author, 11, a.palette.Accent)),
						layout.Rigid(a.text(messageExcerpt(replied.Body, 72), 13, a.palette.Muted)),
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.pill(gtx, &a.replyCancelButton, "×", false) }),
			)
		}))
	})
}

func (a *App) messageReply(gtx layout.Context, messageIndex int, message mock.Message) layout.Dimensions {
	if !message.HasReply || message.ReplyTo < 0 || message.ReplyTo >= messageIndex || message.ReplyTo >= len(a.messages) {
		return layout.Dimensions{}
	}
	replied := a.messages[message.ReplyTo]
	return layout.Inset{Bottom: 5}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(fixedWidth(2, fixedHeight(26, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Border, 1, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{} })
			}))),
			hspace(8),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(a.text(replied.Author, 11, a.palette.Subtle)),
					layout.Rigid(a.text(messageExcerpt(replied.Body, 96), 12, a.palette.Muted)),
				)
			}),
		)
	})
}

func messageExcerpt(value string, limit int) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "Attachment"
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit-1]) + "…"
}
