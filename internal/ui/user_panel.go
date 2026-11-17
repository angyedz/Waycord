package ui

import "gioui.org/layout"

func (a *App) userPanel(gtx layout.Context, expanded bool) layout.Dimensions {
	if !expanded {
		return layout.Inset{Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return layout.Center.Layout(gtx, a.settingsControl) })
	}
	return layout.Inset{Left: 8, Right: 8, Bottom: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 8, Right: 6, Top: 7, Bottom: 7}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					background := a.interactionColor(gtx, &a.selfProfileButton, a.palette.Surface, a.palette.Surface2)
					return a.selfProfileButton.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return rounded(gtx, background, 10, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(a.profileAvatar(32, 13)), hspace(8),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									status := selfPresenceNames[a.selfPresence]
									if a.voice.Connected {
										status = "In " + a.voiceChannel
									}
									return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(a.text(a.selfProfileName, 13, a.palette.Text)), layout.Rigid(a.text(status, 11, a.palette.Subtle)))
								}),
							)
						})
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					icon, label := a.micIcon, "Mute microphone"
					if a.voice.Muted {
						icon, label = a.micOffIcon, "Unmute microphone"
					}
					return a.iconControl(gtx, &a.voiceMute, icon, 30, 17, label)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					label := "Deafen"
					if a.voice.Deafened {
						label = "Undeafen"
					}
					return a.iconControl(gtx, &a.voiceDeafen, a.headsetIcon, 30, 17, label)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return a.settingsControl(gtx)
				}),
			)
		}))
	})
}

func (a *App) settingsControl(gtx layout.Context) layout.Dimensions {
	return a.iconControl(gtx, &a.updatesButton, a.settingsIcon, 34, 18, "Settings")
}
