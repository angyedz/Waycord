package ui

import (
	"image/color"
	"time"

	"gioui.org/io/key"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

var emojiOptions = [...]string{"😀", "😂", "😊", "😍", "👍", "❤️", "🎉", "🔥"}

var emojiPickerOptions = [...]string{
	"😀", "😂", "🙂", "🙃", "😉", "😊", "🥰", "😍",
	"🤔", "😎", "😭", "😅", "😴", "🤗", "🥳", "😡",
	"😱", "🫡", "👋", "🙏", "💯", "✨", "🔥", "🎉",
}

var gifOptions = [...]mock.Media{
	{Kind: mock.MediaGIF, Name: "Dancing cat", Emoji: "🐈"},
	{Kind: mock.MediaGIF, Name: "Tiny celebration", Emoji: "🎉"},
	{Kind: mock.MediaGIF, Name: "That was close", Emoji: "😅"},
	{Kind: mock.MediaGIF, Name: "Big applause", Emoji: "👏"},
	{Kind: mock.MediaGIF, Name: "Thinking", Emoji: "🤔"},
	{Kind: mock.MediaGIF, Name: "Good morning", Emoji: "☀️"},
}

var stickerOptions = [...]mock.Media{
	{Kind: mock.MediaSticker, Name: "Happy blob", Emoji: "😊"},
	{Kind: mock.MediaSticker, Name: "Love you", Emoji: "🥰"},
	{Kind: mock.MediaSticker, Name: "Little wave", Emoji: "👋"},
	{Kind: mock.MediaSticker, Name: "Let's go", Emoji: "🚀"},
	{Kind: mock.MediaSticker, Name: "Very proud", Emoji: "🏆"},
	{Kind: mock.MediaSticker, Name: "Good vibes", Emoji: "✨"},
}

func (a *App) openMediaPicker(tab int) {
	if tab < 0 || tab > 2 {
		return
	}
	if a.emojiOpen && a.mediaPickerTab == tab {
		a.emojiOpen = false
		return
	}
	a.mediaPickerTab = tab
	a.emojiOpen = true
}

func (a *App) processMediaPicker(gtx layout.Context) {
	if !a.emojiOpen {
		return
	}
	for {
		e, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if event, ok := e.(key.Event); ok && event.State == key.Press {
			a.emojiOpen = false
		}
	}
}

func (a *App) emojiControl(gtx layout.Context) layout.Dimensions {
	return a.iconControl(gtx, &a.emojiButton, a.emojiIcon, 34, 20, "Emoji")
}

func (a *App) mediaTextControl(gtx layout.Context, button *widget.Clickable, tab int, label, tooltip string) layout.Dimensions {
	background := color.NRGBA{}
	if a.emojiOpen && a.mediaPickerTab == tab {
		background = a.palette.AccentDim
	}
	background = a.interactionColor(gtx, button, background, a.palette.Surface2)
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(tooltip).Add(gtx.Ops)
		return rounded(gtx, background, 8, padding(layout.Inset{Left: 7, Right: 7, Top: 6, Bottom: 6}, a.text(label, 11, a.palette.Muted)))
	})
}

func (a *App) emojiPicker(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return layout.Inset{Left: 20, Right: 20}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 8, Right: 8, Top: 5, Bottom: 5}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.mediaPickerTabs(gtx) }),
				vspace(5),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.mediaPickerContent(gtx) }),
			)
		}))
	})
}

func (a *App) mediaPickerTabs(gtx layout.Context) layout.Dimensions {
	labels := [...]string{"Emoji", "GIFs", "Stickers"}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.pill(gtx, &a.mediaTabButtons[0], labels[0], a.mediaPickerTab == 0)
		}),
		hspace(5),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.pill(gtx, &a.mediaTabButtons[1], labels[1], a.mediaPickerTab == 1)
		}),
		hspace(5),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.pill(gtx, &a.mediaTabButtons[2], labels[2], a.mediaPickerTab == 2)
		}),
	)
}

func (a *App) mediaPickerContent(gtx layout.Context) layout.Dimensions {
	switch a.mediaPickerTab {
	case 1:
		return a.mediaGrid(gtx, gifOptions[:], 0)
	case 2:
		return a.mediaGrid(gtx, stickerOptions[:], len(gifOptions))
	default:
		return layout.Center.Layout(gtx, a.emojiGrid)
	}
}

func (a *App) emojiGrid(gtx layout.Context) layout.Dimensions {
	rows := make([]layout.FlexChild, 0, 5)
	for start := 0; start < len(emojiPickerOptions); start += 8 {
		start := start
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, 0, 16)
			for index := start; index < min(start+8, len(emojiPickerOptions)); index++ {
				index := index
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.emojiChoice(gtx, &a.emojiPickerChoices[index], emojiPickerOptions[index], 34, 20)
				}), hspace(3))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		}))
		if start+8 < len(emojiPickerOptions) {
			rows = append(rows, vspace(2))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (a *App) mediaGrid(gtx layout.Context, options []mock.Media, offset int) layout.Dimensions {
	rows := make([]layout.FlexChild, 0, 3)
	for start := 0; start < len(options); start += 3 {
		start := start
		rows = append(rows, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			children := make([]layout.FlexChild, 0, 5)
			for index := start; index < min(start+3, len(options)); index++ {
				index := index
				children = append(children, layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					return a.mediaChoice(gtx, &a.mediaChoiceButtons[offset+index], options[index])
				}))
				if index+1 < min(start+3, len(options)) {
					children = append(children, hspace(6))
				}
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
		}))
		if start+3 < len(options) {
			rows = append(rows, vspace(4))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (a *App) mediaChoice(gtx layout.Context, button *widget.Clickable, media mock.Media) layout.Dimensions {
	background := a.interactionColor(gtx, button, a.palette.Surface2, a.palette.AccentDim)
	return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, background, 10, padding(layout.Inset{Left: 7, Right: 7, Top: 4, Bottom: 4}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(a.text(media.Emoji, 19, a.palette.Text)),
				hspace(6),
				layout.Flexed(1, a.text(media.Name, 11, a.palette.Muted)),
			)
		}))
	})
}

// animatedEmojiPicker expands from the composer while keeping its controls
// clipped to the visible portion during both opening and closing.
func (a *App) animatedEmojiPicker(gtx layout.Context, progress float32) layout.Dimensions {
	if progress <= 0 {
		return layout.Dimensions{}
	}
	targetHeight := float32(118)
	if a.mediaPickerTab == 0 {
		targetHeight = 156
	}
	height := a.transition(gtx, &a.emojiPickerHeight, targetHeight, 180*time.Millisecond)
	return animatedRevealSized(gtx, progress, height, a.emojiPicker)
}
