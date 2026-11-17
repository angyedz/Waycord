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
	palette                               theme.Palette
	theme                                 *material.Theme
	serverButtons                         [4]widget.Clickable
	channelButtons                        [4]widget.Clickable
	chatButton, updatesButton, sendButton widget.Clickable
	profileButton                         widget.Clickable
	channelOptions                        [3]widget.Clickable
	activeChannel                         int
	activeServer                          int
	updatesView                           bool
	showProfile                           bool
	updateChannel                         int
	messageList                           layout.List
	composer                              widget.Editor
	messages                              []mock.Message
}

func New() *App {
	a := &App{palette: theme.Dark, activeChannel: 1, showProfile: true, messageList: layout.List{Axis: layout.Vertical}}
	a.theme = material.NewTheme()
	a.theme.Palette = material.Palette{Bg: a.palette.Canvas, Fg: a.palette.Text, ContrastBg: a.palette.Accent, ContrastFg: a.palette.Rail}
	a.composer.SingleLine = true
	a.composer.Submit = true
	a.composer.MaxLen = 2000
	a.messages = append([]mock.Message(nil), mock.Messages...)
	return a
}

func (a *App) process(gtx layout.Context) {
	for i := range a.serverButtons {
		if a.serverButtons[i].Clicked(gtx) {
			a.activeServer = i
		}
	}
	for i := range a.channelButtons {
		if a.channelButtons[i].Clicked(gtx) {
			a.activeChannel = i
			a.updatesView = false
			a.messageList.Position = layout.Position{}
		}
	}
	if a.chatButton.Clicked(gtx) {
		a.updatesView = false
	}
	if a.updatesButton.Clicked(gtx) {
		a.updatesView = true
	}
	if a.profileButton.Clicked(gtx) {
		a.showProfile = !a.showProfile
	}
	for i := range a.channelOptions {
		if a.channelOptions[i].Clicked(gtx) {
			a.updateChannel = i
		}
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
	a.messages = append(a.messages, mock.Message{Channel: a.activeChannel, Author: "you (preview)", Time: "now", Body: body, Self: true})
	a.composer.SetText("")
	a.messageList.ScrollToEnd = true
}

func (a *App) Layout(gtx layout.Context) layout.Dimensions {
	a.process(gtx)
	return rounded(gtx, a.palette.Canvas, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(fixedWidth(64, a.serverRail)),
			layout.Rigid(fixedWidth(216, a.channelRail)),
			layout.Flexed(1, a.workspace),
		)
	})
}

func (a *App) serverRail(gtx layout.Context) layout.Dimensions {
	return rounded(gtx, a.palette.Rail, 0, inset(9, func(gtx layout.Context) layout.Dimensions {
		children := []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Accent, 15, inset(10, a.text("W", 21, a.palette.Rail)))
			}), vspace(18),
		}
		for i, name := range mock.Servers {
			i, name := i, name
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				bg := a.palette.Surface
				fg := a.palette.Muted
				if i == a.activeServer {
					bg = a.palette.AccentDim
					fg = a.palette.Text
				}
				return a.serverButtons[i].Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, bg, 14, inset(12, a.text(name, 17, fg)))
				})
			}), vspace(10))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	}))
}

func (a *App) channelRail(gtx layout.Context) layout.Dimensions {
	return rounded(gtx, a.palette.Sidebar, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return fixedHeight(70, inset(20, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("Waycord", 18, a.palette.Text)),
						vspace(3), layout.Rigid(a.text("YOUR SPACE", 10, a.palette.Subtle)),
					)
				}))(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 10, Right: 10, Top: 22}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					kids := []layout.FlexChild{
						layout.Rigid(inset(9, a.text("CHANNELS", 11, a.palette.Subtle))), vspace(5),
					}
					for i, n := range mock.Channels {
						i, n := i, n
						kids = append(kids, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.clickLabel(gtx, &a.channelButtons[i], "#  "+n, i == a.activeChannel && !a.updatesView)
						}), vspace(4))
					}
					kids = append(kids, vspace(22), layout.Rigid(inset(9, a.text("VOICE CHANNELS", 11, a.palette.Subtle))), vspace(8), layout.Rigid(inset(12, a.text("Lounge  ·  coming soon", 13, a.palette.Muted))))
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, kids...)
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 0, inset(14, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(a.text("●  Offline preview", 13, a.palette.Positive)),
						vspace(3), layout.Rigid(a.text("No account connected", 11, a.palette.Muted)),
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
			if !a.showProfile || gtx.Constraints.Max.X < gtx.Dp(730) {
				return a.chat(gtx)
			}
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, a.chat),
				layout.Rigid(fixedWidth(236, a.profile)),
			)
		}),
	)
}
func (a *App) header(gtx layout.Context) layout.Dimensions {
	wide := gtx.Constraints.Max.X >= gtx.Dp(730)
	return rounded(gtx, a.palette.Canvas, 0, fixedHeight(70, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: 24, Right: 24, Top: 13, Bottom: 13}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					title := "# " + mock.Channels[a.activeChannel]
					if a.updatesView {
						title = "Settings / Updates"
					}
					return a.text(title, 19, a.palette.Text)(gtx)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.pill(gtx, &a.chatButton, "Chat", !a.updatesView) }),
				hspace(10),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.pill(gtx, &a.updatesButton, "Updates", a.updatesView)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if a.updatesView || !wide {
						return layout.Dimensions{}
					}
					return layout.Inset{Left: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.pill(gtx, &a.profileButton, "Profile", a.showProfile)
					})
				}),
			)
		})
	}))
}
