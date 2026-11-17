package ui

import (
	"image"
	"strings"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget/material"
)

func (a *App) processSearch(gtx layout.Context) {
	if a.searchButton.Clicked(gtx) {
		a.searchOpen = true
		gtx.Execute(key.FocusCmd{Tag: &a.searchEditor})
	}
	if a.searchCloseButton.Clicked(gtx) {
		a.searchOpen = false
		a.searchEditor.SetText("")
		a.messageList.Position = layout.Position{}
	}
	if a.pinnedButton.Clicked(gtx) {
		a.pinsOnly = !a.pinsOnly
		a.messageList.Position = layout.Position{}
	}
	if a.searchOpen {
		previous := a.searchEditor.Text()
		for {
			if _, ok := a.searchEditor.Update(gtx); !ok {
				break
			}
		}
		if previous != a.searchEditor.Text() {
			a.messageList.Position = layout.Position{}
			a.messageList.ScrollToEnd = false
		}
		for {
			e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
			if !ok {
				break
			}
			if event, ok := e.(key.Event); ok && event.State == key.Press {
				a.searchOpen = false
				a.searchEditor.SetText("")
				a.messageList.Position = layout.Position{}
			}
		}
	}
}

func (a *App) searchField(gtx layout.Context, progress float32) layout.Dimensions {
	fullWidth := gtx.Constraints.Max.X
	visibleWidth := int(float32(fullWidth) * progress)
	if visibleWidth < 1 {
		visibleWidth = 1
	}
	area := clip.Rect{Max: image.Pt(visibleWidth, gtx.Constraints.Max.Y)}.Push(gtx.Ops)
	offset := op.Offset(image.Pt(visibleWidth-fullWidth, 0)).Push(gtx.Ops)
	child := gtx
	child.Constraints.Min.X, child.Constraints.Max.X = fullWidth, fullWidth
	content := func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 18, padding(layout.Inset{Left: 12, Right: 12, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
			input := material.Editor(a.theme, &a.searchEditor, "Search this channel…")
			input.TextSize = 14
			input.Color = a.palette.Text
			return input.Layout(gtx)
		}))
	}
	content(child)
	offset.Pop()
	area.Pop()
	return layout.Dimensions{Size: image.Pt(visibleWidth, gtx.Constraints.Max.Y)}
}

func (a *App) visibleMessages() []int {
	query := strings.ToLower(strings.TrimSpace(a.searchEditor.Text()))
	visible := make([]int, 0, len(a.messages))
	for index, message := range a.messages {
		if message.Direct != a.direct || (!a.direct && (message.Server != a.activeServer || message.Channel != a.activeChannel)) {
			continue
		}
		if a.pinsOnly && !message.Pinned {
			continue
		}
		if query != "" {
			searchable := strings.ToLower(message.Author + " " + message.Body)
			if message.Attachment != nil {
				searchable += " " + strings.ToLower(message.Attachment.Name)
			}
			if !strings.Contains(searchable, query) {
				continue
			}
		}
		visible = append(visible, index)
	}
	return visible
}
