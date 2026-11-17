package ui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
)

func (a *App) voicePanel(gtx layout.Context, expanded bool) layout.Dimensions {
	progress := a.transition(gtx, &a.voicePanelSlide, boolProgress(a.voice.Connected && expanded), 220*time.Millisecond)
	return animatedReveal(gtx, progress, unit.Dp(110), a.voicePanelContent)
}

func (a *App) voicePanelContent(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Left: 8, Right: 8, Bottom: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, a.palette.Surface, 14, padding(layout.Inset{Left: 10, Right: 8, Top: 9, Bottom: 8}, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(a.text("Voice connected", 13, a.palette.Positive)),
								layout.Rigid(a.text(a.voiceChannel+" · "+a.voiceServer, 11, a.palette.Muted)),
							)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.iconControl(gtx, &a.voiceLeave, a.hangupIcon, 32, 18, "Disconnect")
						}),
					)
				}),
				vspace(5),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					camera, cameraLabel := a.videoIcon, "Turn on camera"
					if a.voice.Camera {
						camera, cameraLabel = a.videoOffIcon, "Turn off camera"
					}
					screen, screenLabel := a.screenIcon, "Share screen"
					if a.voice.Screen {
						screen, screenLabel = a.screenOffIcon, "Stop sharing"
					}
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.iconControl(gtx, &a.voiceCamera, camera, 34, 19, cameraLabel)
						}),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return a.iconControl(gtx, &a.voiceScreen, screen, 34, 19, screenLabel)
						}),
					)
				}),
			)
		}))
	})
}
