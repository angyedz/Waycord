package ui

import (
	"fmt"
	"strconv"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/core/attachments"
)

type attachmentResult struct {
	conversation string
	attachment   *attachments.Attachment
	err          error
}

type attachmentPreviewCache struct {
	items map[*attachments.Attachment]paint.ImageOp
	order []*attachments.Attachment
}

func (c *attachmentPreviewCache) image(file *attachments.Attachment) paint.ImageOp {
	if c.items == nil {
		c.items = make(map[*attachments.Attachment]paint.ImageOp)
	}
	if image, ok := c.items[file]; ok {
		return image
	}
	if len(c.order) == 16 {
		delete(c.items, c.order[0])
		c.order = c.order[1:]
	}
	image := paint.NewImageOp(file.Preview)
	c.items[file] = image
	c.order = append(c.order, file)
	return image
}

// ConfigureAttachments provides a native chooser without coupling UI to a window.
func (a *App) ConfigureAttachments(choose func() (*attachments.Attachment, error)) {
	a.chooseAttachment = choose
}

func (a *App) conversationKey() string {
	if a.direct {
		return "direct:mira"
	}
	return "server:" + strconv.Itoa(a.activeServer) + ":" + strconv.Itoa(a.activeChannel)
}

func (a *App) beginAttachmentSelection() {
	if a.attachmentBusy || a.chooseAttachment == nil {
		return
	}
	a.attachmentBusy = true
	a.attachmentError = ""
	key, choose, invalidate := a.conversationKey(), a.chooseAttachment, a.updateInvalidate
	go func() {
		file, err := choose()
		a.attachmentResults <- attachmentResult{conversation: key, attachment: file, err: err}
		if invalidate != nil {
			invalidate()
		}
	}()
}

func (a *App) processAttachmentResults() {
	select {
	case result := <-a.attachmentResults:
		a.attachmentBusy = false
		if result.err != nil {
			a.attachmentError = "Couldn't attach file: " + result.err.Error()
		} else if result.attachment != nil {
			a.attachmentDrafts[result.conversation] = result.attachment
		}
	default:
	}
}

func (a *App) attachmentCard(gtx layout.Context, file *attachments.Attachment) layout.Dimensions {
	if file == nil {
		return layout.Dimensions{}
	}
	return layout.Inset{Top: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 12, inset(10, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(a.text(file.Name, 14, a.palette.Text)),
				layout.Rigid(a.text(fmt.Sprintf("%.1f KB", float64(len(file.Data))/1024), 12, a.palette.Subtle)),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if file.Preview == nil {
						return layout.Dimensions{}
					}
					gtx.Constraints.Min.X = 0
					gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(280))
					gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(180))
					preview := widget.Image{Src: a.attachmentPreviews.image(file), Fit: widget.Contain}
					return preview.Layout(gtx)
				}),
			)
		}))
	})
}

func (a *App) attachmentDraft(gtx layout.Context) layout.Dimensions {
	file := a.attachmentDrafts[a.conversationKey()]
	if file == nil && a.attachmentError == "" && !a.attachmentBusy {
		return layout.Dimensions{}
	}
	return layout.Inset{Left: 20, Right: 20, Top: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if a.attachmentError != "" {
			return a.text(a.attachmentError, 12, a.palette.Warning)(gtx)
		}
		if file == nil {
			return a.text("Choosing file…", 12, a.palette.Muted)(gtx)
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, a.text(file.Name, 13, a.palette.Text)),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.pill(gtx, &a.removeAttachmentButton, "Remove", false)
			}),
		)
	})
}
