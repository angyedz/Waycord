package ui

import (
	"strings"

	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/angyedz/Waycord/internal/ui/mock"
	"github.com/angyedz/Waycord/internal/ui/theme"
)

// App contains UI state only. Future protocol/media services stay outside it.
type App struct {
	palette                                   theme.Palette
	theme                                     *material.Theme
	channelButtons                            [4]widget.Clickable
	chatButton, updatesButton, sendButton     widget.Clickable
	directButton, serverButton, contactButton widget.Clickable
	activeChannel                             int
	updatesView                               bool
	direct                                    bool
	messageList                               layout.List
	composer                                  widget.Editor
	messages                                  []mock.Message
}

func New() *App {
	a := &App{palette: theme.Dark, activeChannel: 1, messageList: layout.List{Axis: layout.Vertical}}
	a.theme = material.NewTheme()
	a.theme.Palette = material.Palette{Bg: a.palette.Canvas, Fg: a.palette.Text, ContrastBg: a.palette.Accent, ContrastFg: a.palette.Rail}
	a.composer.SingleLine = true
	a.composer.Submit = true
	a.composer.MaxLen = 2000
	a.messages = append([]mock.Message(nil), mock.Messages...)
	return a
}

func (a *App) process(gtx layout.Context) {
	if a.directButton.Clicked(gtx) || a.contactButton.Clicked(gtx) {
		a.direct, a.updatesView = true, false
		a.messageList.Position = layout.Position{}
		a.messageList.ScrollToEnd = false
	}
	if a.serverButton.Clicked(gtx) {
		a.direct, a.updatesView = false, false
		a.messageList.Position = layout.Position{}
		a.messageList.ScrollToEnd = false
	}
	for i := range a.channelButtons {
		if a.channelButtons[i].Clicked(gtx) {
			a.activeChannel = i
			a.updatesView = false
			a.direct = false
			a.messageList.Position = layout.Position{}
			a.messageList.ScrollToEnd = false
		}
	}
	if a.chatButton.Clicked(gtx) {
		a.updatesView = false
	}
	if a.updatesButton.Clicked(gtx) {
		a.updatesView = true
	}
	if a.sendButton.Clicked(gtx) && !a.updatesView {
		a.sendPreview()
	}
	for {
		event, ok := a.composer.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.SubmitEvent); ok && !a.updatesView {
			a.sendPreview()
		}
	}
}
func (a *App) sendPreview() {
	body := strings.TrimSpace(a.composer.Text())
	if body == "" {
		return
	}
	a.messages = append(a.messages, mock.Message{Channel: a.activeChannel, Direct: a.direct, Author: "You", Time: "now", Body: body, Self: true})
	a.composer.SetText("")
	a.messageList.ScrollToEnd = true
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	a.process(gtx)
	return rounded(gtx, a.palette.Canvas, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(fixedWidth(60, a.serverRail)),
			layout.Rigid(fixedWidth(200, a.channelRail)),
			layout.Flexed(1, a.workspace),
		)
	})
}

func (a *App) serverRail(gtx layout.Context) layout.Dimensions {
	return rounded(gtx, a.palette.Rail, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.railButton(gtx, &a.directButton, "@", a.direct) }),
				vspace(10),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.railButton(gtx, &a.serverButton, "W", !a.direct) }),
			)
		})
	})
}

func (a *App) channelRail(gtx layout.Context) layout.Dimensions {
	return rounded(gtx, a.palette.Sidebar, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return fixedHeight(58, inset(14, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							if a.direct {
								return a.text("Messages", 18, a.palette.Text)(gtx)
							}
							return a.text("Waycord", 18, a.palette.Text)(gtx)
						}),
					)
				}))(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 8, Right: 8, Top: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					if a.direct {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(inset(9, a.text("DIRECT MESSAGES", 11, a.palette.Subtle))),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return a.clickLabel(gtx, &a.contactButton, "mira", !a.updatesView)
							}),
						)
					}
					kids := []layout.FlexChild{
						layout.Rigid(inset(9, a.text("CHANNELS", 11, a.palette.Subtle))), vspace(5),
					}
					for i, n := range mock.Channels {
						i, n := i, n
						kids = append(kids, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.clickLabel(gtx, &a.channelButtons[i], "#  "+n, i == a.activeChannel && !a.updatesView)
						}), vspace(4))
					}
					kids = append(kids, vspace(12), layout.Rigid(inset(9, a.text("VOICE CHANNELS", 11, a.palette.Subtle))), vspace(4), layout.Rigid(inset(10, a.text("Lounge  ·  coming soon", 13, a.palette.Muted))))
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, kids...)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 0, inset(10, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("●  Offline", 13, a.palette.Muted)),
						vspace(6), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.pill(gtx, &a.updatesButton, "Settings", a.updatesView)
						}),
					)
				}))
			}),
		)
	})
}

func (a *App) workspace(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.header),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if a.updatesView {
				return a.updates(gtx)
			}
			if a.direct && gtx.Constraints.Max.X >= gtx.Dp(730) {
				return layout.Flex{}.Layout(gtx, layout.Flexed(1, a.chat), layout.Rigid(fixedWidth(210, a.directProfile)))
			}
			return a.chat(gtx)
		}),
	)
}
func (a *App) header(gtx layout.Context) layout.Dimensions {
	return rounded(gtx, a.palette.Canvas, 0, fixedHeight(58, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: 20, Right: 20, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					title := "# " + mock.Channels[a.activeChannel]
					if a.direct {
						title = "@ mira"
					}
					if a.updatesView {
						title = "Settings / Updates"
					}
					return a.text(title, 19, a.palette.Text)(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if !a.updatesView {
						return layout.Dimensions{}
					}
					return a.pill(gtx, &a.chatButton, "Back", false)
				}),
			)
		})
	}))
}
