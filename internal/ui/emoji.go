package ui

import (
	"gioui.org/layout"
)

var emojiOptions = [...]string{"😀", "😂", "😊", "😍", "👍", "❤️", "🎉", "🔥"}

func (a *App) emojiControl(gtx layout.Context) layout.Dimensions {
	return a.emojiButton.Layout(gtx, fixedWidth(34, fixedHeight(34, func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, fixedWidth(20, func(gtx layout.Context) layout.Dimensions {
			return a.emojiIcon.Layout(gtx, a.palette.Muted)
		}))
	})))
}

func (a *App) emojiPicker(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 16, padding(layout.Inset{Left: 8, Right: 8, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, 0, len(emojiOptions))
			for i, value := range emojiOptions {
				i, value := i, value
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.emojiChoices[i].Layout(gtx, fixedWidth(38, fixedHeight(38, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, a.text(value, 21, a.palette.Text))
					})))
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		}))
	})
}
