package ui

import (
	"gioui.org/layout"
)

var emojiOptions = [...]string{"😀", "😂", "😊", "😍", "👍", "❤️", "🎉", "🔥"}

func (a *App) emojiControl(gtx layout.Context) layout.Dimensions {
	return a.iconControl(gtx, &a.emojiButton, a.emojiIcon, 34, 20, "Emoji")
}

func (a *App) emojiPicker(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 16, padding(layout.Inset{Left: 8, Right: 8, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, 0, len(emojiOptions))
			for i, value := range emojiOptions {
				i, value := i, value
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.emojiChoice(gtx, &a.emojiChoices[i], value, 38, 21)
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		}))
	})
}
