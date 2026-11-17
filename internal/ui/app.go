package ui

import (
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"

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
	sidebarButton, profileButton              widget.Clickable
	sidebar                                   sidebarState
	menuIcon                                  *widget.Icon
	sendIcon                                  *widget.Icon
	settingsIcon                              *widget.Icon
	emojiIcon                                 *widget.Icon
	emojiButton                               widget.Clickable
	emojiChoices                              [8]widget.Clickable
	emojiOpen                                 bool
	sendHover                                 float32
	sendHoverTime                             time.Time
	activeChannel                             int
	updatesView                               bool
	direct                                    bool
	profileOpen                               bool
	messageList                               layout.List
	composer                                  widget.Editor
	messages                                  []mock.Message
}

func New() *App {
	a := &App{palette: theme.Dark, activeChannel: 1, sidebar: newSidebarState(), profileOpen: true, messageList: layout.List{Axis: layout.Vertical}}
	a.theme = material.NewTheme()
	if shaper, err := theme.NewShaper(); err == nil {
		a.theme.Shaper = shaper
		a.theme.Face = font.Typeface(theme.TextFamilies)
	}
	a.menuIcon, _ = widget.NewIcon(icons.NavigationMenu)
	a.sendIcon, _ = widget.NewIcon(icons.ContentSend)
	a.settingsIcon, _ = widget.NewIcon(icons.ActionSettings)
	a.emojiIcon, _ = widget.NewIcon(icons.SocialSentimentSatisfied)
	a.theme.Palette = material.Palette{Bg: a.palette.Canvas, Fg: a.palette.Text, ContrastBg: a.palette.Accent, ContrastFg: a.palette.Rail}
	a.composer.SingleLine = true
	a.composer.Submit = true
	a.composer.MaxLen = 2000
	a.messages = append([]mock.Message(nil), mock.Messages...)
	return a
}

func (a *App) process(gtx layout.Context) {
	if a.emojiButton.Clicked(gtx) {
		a.emojiOpen = !a.emojiOpen
	}
	for i := range a.emojiChoices {
		if a.emojiChoices[i].Clicked(gtx) {
			a.composer.Insert(emojiOptions[i])
			a.emojiOpen = false
			gtx.Execute(key.FocusCmd{Tag: &a.composer})
		}
	}
	if a.sidebarButton.Clicked(gtx) {
		a.sidebar.toggle()
	}
	if a.profileButton.Clicked(gtx) {
		a.profileOpen = !a.profileOpen
	}
	a.sidebar.update(gtx)
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
	width := a.sidebar.animate(gtx)
	return rounded(gtx, a.palette.Canvas, 0, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Left: 8, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, a.palette.Sidebar, 18, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{}.Layout(gtx,
							layout.Rigid(fixedWidth(52, a.serverRail)),
							layout.Rigid(fixedWidth(width, func(gtx layout.Context) layout.Dimensions {
								if width < 120 {
									return layout.Dimensions{Size: gtx.Constraints.Min}
								}
								return a.channelRail(gtx)
							})),
						)
					})
				})
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if !a.sidebar.shown || width < 120 {
					return layout.Dimensions{}
				}
				return a.sidebar.handle(gtx, a.palette.Border)
			}),
			layout.Flexed(1, a.workspace),
		)
	})
}

func (a *App) serverRail(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.railButton(gtx, &a.directButton, "@", a.direct) }),
			vspace(10),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.railButton(gtx, &a.serverButton, "W", !a.direct) }),
		)
	})
}

func (a *App) channelRail(gtx layout.Context) layout.Dimensions {
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
				var kids []layout.FlexChild
				for i, n := range mock.Channels {
					i, n := i, n
					kids = append(kids, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.clickLabel(gtx, &a.channelButtons[i], "#  "+n, i == a.activeChannel && !a.updatesView)
					}), vspace(4))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, kids...)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 8, Right: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 8, Right: 6, Top: 7, Bottom: 7}, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(fixedWidth(32, fixedHeight(32, func(gtx layout.Context) layout.Dimensions {
							return rounded(gtx, a.palette.AccentDim, 16, func(gtx layout.Context) layout.Dimensions {
								return layout.Center.Layout(gtx, a.text("Y", 13, a.palette.Text))
							})
						}))),
						hspace(8),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(a.text("You", 13, a.palette.Text)), layout.Rigid(a.text("Offline", 11, a.palette.Subtle)))
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.updatesButton.Layout(gtx, fixedWidth(34, fixedHeight(34, func(gtx layout.Context) layout.Dimensions {
								return layout.Center.Layout(gtx, fixedWidth(18, func(gtx layout.Context) layout.Dimensions { return a.settingsIcon.Layout(gtx, a.palette.Muted) }))
							})))
						}),
					)
				}))
			})
		}),
	)
}

func (a *App) workspace(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(a.header),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if a.updatesView {
				return a.updates(gtx)
			}
			if a.direct && a.profileOpen && gtx.Constraints.Max.X >= gtx.Dp(730) {
				return layout.Flex{}.Layout(gtx, layout.Flexed(1, a.chat), layout.Rigid(fixedWidth(210, a.directProfile)))
			}
			return a.chat(gtx)
		}),
	)
}
func (a *App) header(gtx layout.Context) layout.Dimensions {
	wide := gtx.Constraints.Max.X >= gtx.Dp(730)
	return rounded(gtx, a.palette.Canvas, 0, fixedHeight(58, func(gtx layout.Context) layout.Dimensions {
		return layout.Inset{Left: 20, Right: 20, Top: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return a.sidebarButton.Layout(gtx, fixedWidth(36, fixedHeight(36, func(gtx layout.Context) layout.Dimensions {
							return layout.Center.Layout(gtx, fixedWidth(20, func(gtx layout.Context) layout.Dimensions { return a.menuIcon.Layout(gtx, a.palette.Muted) }))
						})))
					})
				}),
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
					if a.direct && !a.updatesView && wide {
						return a.pill(gtx, &a.profileButton, "Profile", a.profileOpen)
					}
					if !a.updatesView {
						return layout.Dimensions{}
					}
					return a.pill(gtx, &a.chatButton, "Back", false)
				}),
			)
		})
	}))
}
