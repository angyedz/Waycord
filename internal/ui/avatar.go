package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/core/attachments"
)

type selfAvatarResult struct {
	avatar *attachments.Attachment
	err    error
}

// ConfigureLocalAvatar supplies the local file chooser and durable save hook.
// Image data stays in the user's configuration directory and never leaves the app.
func (a *App) ConfigureLocalAvatar(current image.Image, choose func() (*attachments.Attachment, error), save func([]byte) error) {
	a.selfAvatar = current
	if current != nil {
		a.selfAvatarOp = paint.NewImageOp(current)
	} else {
		a.selfAvatarOp = paint.ImageOp{}
	}
	a.chooseSelfAvatar = choose
	a.saveSelfAvatar = save
}

func (a *App) beginSelfAvatarSelection() {
	if a.selfAvatarBusy || a.attachmentBusy || a.chooseSelfAvatar == nil {
		return
	}
	a.selfAvatarBusy = true
	a.selfAvatarError = ""
	choose, invalidate := a.chooseSelfAvatar, a.updateInvalidate
	go func() {
		avatar, err := choose()
		a.selfAvatarResults <- selfAvatarResult{avatar: avatar, err: err}
		if invalidate != nil {
			invalidate()
		}
	}()
}

func (a *App) processSelfAvatarResult() {
	select {
	case result := <-a.selfAvatarResults:
		a.applySelfAvatarResult(result)
	default:
	}
}

func (a *App) applySelfAvatarResult(result selfAvatarResult) {
	a.selfAvatarBusy = false
	if result.err != nil {
		a.selfAvatarError = "Couldn't open that picture. Try PNG, JPG, or GIF under 8 MB."
		return
	}
	if result.avatar == nil {
		return
	}
	if result.avatar.Preview == nil || len(result.avatar.Data) == 0 {
		a.selfAvatarError = "Choose a PNG, JPG, or GIF image up to 8 MB."
		return
	}
	if a.saveSelfAvatar != nil {
		if err := a.saveSelfAvatar(result.avatar.Data); err != nil {
			a.selfAvatarError = "Couldn't save the avatar. Try again."
			return
		}
	}
	a.selfAvatar = result.avatar.Preview
	a.selfAvatarOp = paint.NewImageOp(a.selfAvatar)
	a.selfAvatarError = ""
}

func (a *App) profileAvatar(size unit.Dp, fallbackSize unit.Sp) layout.Widget {
	return fixedWidth(size, fixedHeight(size, func(gtx layout.Context) layout.Dimensions {
		if a.selfAvatar == nil {
			return rounded(gtx, a.palette.AccentDim, size/2, func(gtx layout.Context) layout.Dimensions {
				return layout.Center.Layout(gtx, a.text(firstLetter(a.selfProfileName), fallbackSize, a.palette.Text))
			})
		}
		bounds := gtx.Constraints.Max
		area := clip.Ellipse{Max: bounds}.Push(gtx.Ops)
		dimensions := (widget.Image{Src: a.selfAvatarOp, Fit: widget.Cover}).Layout(gtx)
		area.Pop()
		return dimensions
	}))
}
