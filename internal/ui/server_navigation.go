package ui

import (
	"strconv"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/angyedz/Waycord/internal/ui/mock"
)

func (a *App) categoryKey(index int) string {
	return strconv.Itoa(a.activeServer) + ":" + strconv.Itoa(index)
}

func (a *App) categoryButton(key string) *widget.Clickable {
	if a.categoryButtons[key] == nil {
		a.categoryButtons[key] = new(widget.Clickable)
	}
	return a.categoryButtons[key]
}

func (a *App) categoryTransition(key string, open bool) *panelTransition {
	if a.categoryTransitions[key] == nil {
		value := float32(0)
		if open {
			value = 1
		}
		a.categoryTransitions[key] = &panelTransition{value: value, from: value, target: value}
	}
	return a.categoryTransitions[key]
}

func (a *App) serverChannelsLayout(gtx layout.Context) layout.Dimensions {
	a.serverChannelList.Axis = layout.Vertical
	return a.serverChannelList.Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
		serverIndex := a.activeServer
		server := mock.Servers[serverIndex]
		children := a.serverFeatureRows()
		if len(children) > 0 {
			children = append(children,
				vspace(5),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions { return horizontalDivider(gtx, a.palette.Border) }),
				vspace(8),
			)
		}
		addChannel := func(index int) {
			if index < 0 || index >= len(server.Channels) || index >= len(a.channelButtons[serverIndex]) {
				return
			}
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.clickLabel(gtx, &a.channelButtons[serverIndex][index], "#  "+server.Channels[index], index == a.activeChannel && !a.updatesView && a.activeServerFeature == serverFeatureNone)
			}), vspace(4))
		}
		for _, index := range server.Uncategorized {
			addChannel(index)
		}
		for index, category := range server.Categories {
			key, category := a.categoryKey(index), category
			button := a.categoryButton(key)
			collapsed := a.collapsedCategories[key]
			arrow := "⌄  "
			if collapsed {
				arrow = "›  "
			}
			children = append(children, vspace(5), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				background := a.interactionColor(gtx, button, a.palette.Sidebar, a.palette.Surface2)
				return button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return rounded(gtx, background, 8, padding(layout.Inset{Left: 8, Right: 8, Top: 6, Bottom: 6}, a.text(arrow+category.Name, 12, a.palette.Muted)))
				})
			}))
			progress := a.transition(gtx, a.categoryTransition(key, !collapsed), boolProgress(!collapsed), 180*time.Millisecond)
			if progress > 0 && len(category.Channels) > 0 {
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return animatedReveal(gtx, progress, unit.Dp(len(category.Channels)*40), func(gtx layout.Context) layout.Dimensions {
						var channels []layout.FlexChild
						for _, index := range category.Channels {
							if index < 0 || index >= len(server.Channels) || index >= len(a.channelButtons[serverIndex]) {
								continue
							}
							index := index
							channels = append(channels, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return a.clickLabel(gtx, &a.channelButtons[serverIndex][index], "#  "+server.Channels[index], index == a.activeChannel && !a.updatesView && a.activeServerFeature == serverFeatureNone)
							}), vspace(4))
						}
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx, channels...)
					})
				}))
			}
		}
		if server.VoiceChannel != "" {
			connected := a.voice.Connected && a.voiceServer == server.Name && a.voiceChannel == server.VoiceChannel
			children = append(children, vspace(8), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.clickLabel(gtx, &a.voiceJoin, "◖  "+server.VoiceChannel, connected)
			}))
		}
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
	})
}
