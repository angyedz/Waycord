package ui

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) processContextMenu(gtx layout.Context) {
	if a.deleteDialogOpen {
		return
	}
	if a.contextTarget >= 0 {
		if a.contextEdit.Clicked(gtx) {
			a.beginEdit(gtx, a.contextTarget)
		}
		if a.contextDelete.Clicked(gtx) {
			a.requestDelete(a.contextTarget)
		}
		if a.contextPin.Clicked(gtx) {
			if a.contextTarget < len(a.messages) {
				a.messages[a.contextTarget].Pinned = !a.messages[a.contextTarget].Pinned
			}
			a.contextTarget = -1
		}
		if a.contextCopy.Clicked(gtx) {
			if a.contextTarget < len(a.messages) {
				gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(a.messages[a.contextTarget].Body))})
			}
			a.contextTarget = -1
		}
		if a.contextReact.Clicked(gtx) {
			a.reactionTarget = a.contextTarget
			a.contextTarget = -1
		}
		if a.contextReply.Clicked(gtx) {
			a.beginReply(gtx, a.contextTarget)
		}
		if a.contextBackdrop.Clicked(gtx) {
			a.contextTarget = -1
		}
		for {
			e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
			if !ok {
				break
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				a.contextTarget = -1
			}
		}
	}
	for index, w := range a.reactionWidgets {
		for {
			raw, ok := gtx.Event(pointer.Filter{Target: &w.contextTag, Kinds: pointer.Press})
			if !ok {
				break
			}
			e, ok := raw.(pointer.Event)
			if ok && e.Source == pointer.Mouse && e.Buttons.Contain(pointer.ButtonSecondary) {
				a.contextTarget = index
				a.contextShown = index
				a.contextPosition = e.Position.Round()
				a.reactionTarget = -1
			}
		}
	}
}

func (a *App) contextMenu(gtx layout.Context) layout.Dimensions {
	pinLabel := "Pin message"
	message := a.contextMessage()
	if message != nil && message.Pinned {
		pinLabel = "Remove pin"
	}
	return fixedWidth(190, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface2, 12, padding(layout.Inset{Left: 4, Right: 4, Top: 4, Bottom: 4}, func(gtx layout.Context) layout.Dimensions {
			children := []layout.FlexChild{
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.contextItem(gtx, &a.contextReply, "Reply") }),
			}
			if message != nil && canEditMessage(*message) {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.contextItem(gtx, &a.contextEdit, "Edit message") }))
			}
			children = append(children,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.contextItem(gtx, &a.contextCopy, "Copy message") }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.contextItem(gtx, &a.contextReact, "Add reaction") }),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.contextItem(gtx, &a.contextPin, pinLabel) }),
			)
			if message != nil && message.Self {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.contextItem(gtx, &a.contextDelete, "Delete message")
				}))
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
		}))
	})(gtx)
}

func (a *App) contextMessage() *mock.Message {
	index := a.contextTarget
	if index < 0 {
		index = a.contextShown
	}
	if index < 0 || index >= len(a.messages) {
		return nil
	}
	return &a.messages[index]
}

func canEditMessage(message mock.Message) bool {
	return message.Self && strings.TrimSpace(message.Body) != "" && message.Attachment == nil && message.Media == nil
}

func (a *App) contextMenuHeight() int {
	message := a.contextMessage()
	height := 144
	if message != nil && canEditMessage(*message) {
		height += 36
	}
	if message != nil && message.Self {
		height += 36
	}
	return height
}

func (a *App) contextItem(gtx layout.Context, button *widget.Clickable, title string) layout.Dimensions {
	background := a.interactionColor(gtx, button, a.palette.Surface2, a.palette.AccentDim)
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, background, 8, padding(layout.Inset{Left: 12, Right: 12, Top: 8, Bottom: 8}, a.text(title, 14, a.palette.Text)))
	})
}
