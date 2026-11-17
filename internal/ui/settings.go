package ui

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	uiMarkdown "github.com/angyedz/Waycord/internal/ui/markdown"
	"github.com/angyedz/Waycord/internal/ui/theme"
)

var settingsNames = [...]string{"Appearance", "Performance", "Audio", "Video", "Updates"}
var themeNames = [...]string{"Dark", "Midnight", "Light"}
var themeKeys = [...]string{"dark", "midnight", "light"}

func (a *App) ConfigureUIPreferences(themeName string, compactMessages, reduceMotion bool, save func(string, bool, bool) error) {
	if _, ok := theme.ForName(themeName); !ok {
		themeName = "dark"
	}
	a.themeName, a.compactMessages, a.reduceMotion, a.saveUIPreferences = themeName, compactMessages, reduceMotion, save
	a.applyTheme(themeName)
}

func (a *App) setUIPreferences(themeName string, compactMessages, reduceMotion bool) {
	if _, ok := theme.ForName(themeName); !ok {
		return
	}
	if a.saveUIPreferences != nil {
		if err := a.saveUIPreferences(themeName, compactMessages, reduceMotion); err != nil {
			a.settingsStatus = "Couldn't save these settings. Try again."
			return
		}
	}
	a.settingsStatus = ""
	a.themeName, a.compactMessages, a.reduceMotion = themeName, compactMessages, reduceMotion
	a.applyTheme(themeName)
}

func (a *App) applyTheme(name string) {
	palette, ok := theme.ForName(name)
	if !ok {
		return
	}
	a.palette = palette
	a.theme.Palette = materialPalette(palette)
	a.messageMarkdown = uiMarkdown.NewCache(uiMarkdown.New(theme.TextFamilies, palette.Muted, palette.Accent), 64)
}

func materialPalette(palette theme.Palette) material.Palette {
	return material.Palette{Bg: palette.Canvas, Fg: palette.Text, ContrastBg: palette.Accent, ContrastFg: palette.Rail}
}

func (a *App) settings(gtx layout.Context) layout.Dimensions {
	return layout.Flex{}.Layout(gtx,
		layout.Rigid(fixedWidth(160, func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: 12, Top: 16, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, 0, len(settingsNames)*2)
				for i, title := range settingsNames {
					i, title := i, title
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return a.clickLabel(gtx, &a.settingsSections[i], title, a.settingsSection == i)
					}), vspace(3))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
		})),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if a.settingsSection == 4 {
				return a.updates(gtx)
			}
			return layout.Inset{Top: 16, Left: 24, Right: 24}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := a.settingsSectionContent()
				if a.settingsStatus != "" {
					children = append(children, vspace(14), layout.Rigid(a.text(a.settingsStatus, 13, a.palette.Warning)))
				}
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
			})
		}),
	)
}

func (a *App) settingsSectionContent() []layout.FlexChild {
	title := settingsNames[a.settingsSection]
	description := [...]string{
		"Make Waycord feel right for you.",
		"Choose how motion and message density behave.",
		"Control your microphone and headphones.",
		"Manage camera and screen sharing while in voice.",
	}[a.settingsSection]
	children := []layout.FlexChild{
		layout.Rigid(a.text(title, 24, a.palette.Text)),
		vspace(6),
		layout.Rigid(a.text(description, 14, a.palette.Muted)),
		vspace(18),
	}
	switch a.settingsSection {
	case 0:
		children = append(children,
			layout.Rigid(a.text("Color theme", 15, a.palette.Text)),
			vspace(8),
			layout.Rigid(a.themeChoices),
			vspace(18),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.preferenceRow(gtx, "Compact messages", "Tighter spacing in busy conversations", &a.compactButton, a.compactMessages)
			}),
		)
	case 1:
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.preferenceRow(gtx, "Reduce motion", "Skip movement and hover transitions", &a.reduceMotionButton, a.reduceMotion)
		}))
	case 2:
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Microphone is unmuted"
				if a.voice.Muted {
					label = "Microphone is muted"
				}
				return a.preferenceRow(gtx, "Microphone", label, &a.voiceMute, a.voice.Muted)
			}),
			vspace(8),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Headphones are on"
				if a.voice.Deafened {
					label = "Headphones are muted"
				}
				return a.preferenceRow(gtx, "Deafen", label, &a.voiceDeafen, a.voice.Deafened)
			}),
			vspace(16),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.text(a.audioStateLabel(), 12, a.palette.Subtle)(gtx)
			}),
		)
	case 3:
		if !a.voice.Connected {
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 16, Right: 16, Top: 14, Bottom: 14}, func(gtx layout.Context) layout.Dimensions {
					return a.text("Join a voice channel to use camera and screen sharing.", 14, a.palette.Muted)(gtx)
				}))
			}))
		} else {
			children = append(children,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "Camera is off"
					if a.voice.Camera {
						label = "Camera is on"
					}
					return a.preferenceRow(gtx, "Camera", label, &a.voiceCamera, a.voice.Camera)
				}),
				vspace(8),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "Screen sharing is off"
					if a.voice.Screen {
						label = "Screen sharing is on"
					}
					return a.preferenceRow(gtx, "Screen sharing", label, &a.voiceScreen, a.voice.Screen)
				}),
			)
		}
	}
	return children
}

func (a *App) themeChoices(gtx layout.Context) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(themeNames)*2)
	for i, name := range themeNames {
		i, name := i, name
		children = append(children,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return a.pill(gtx, &a.themeButtons[i], name, a.themeName == themeKeys[i])
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if i == len(themeNames)-1 {
					return layout.Dimensions{}
				}
				return layout.Spacer{Width: 8}.Layout(gtx)
			}),
		)
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
}

func (a *App) preferenceRow(gtx layout.Context, title, description string, button *widget.Clickable, active bool) layout.Dimensions {
	return rounded(gtx, a.palette.Surface, 12, padding(layout.Inset{Left: 14, Right: 10, Top: 10, Bottom: 10}, func(gtx layout.Context) layout.Dimensions {
		label := "Off"
		if active {
			label = "On"
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(a.text(title, 14, a.palette.Text)),
					layout.Rigid(a.text(description, 12, a.palette.Subtle)),
				)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions { return a.pill(gtx, button, label, active) }),
		)
	}))
}

func (a *App) audioStateLabel() string {
	if a.voice.Connected {
		return "In " + a.voiceChannel + " · " + a.voiceServer
	}
	return "Not in a voice channel"
}
