package ui

import (
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op"
)

// sendControl eases its hover state without scheduling frames while idle.
func (a *App) sendControl(gtx layout.Context) layout.Dimensions {
	target := float32(0)
	if a.sendButton.Pressed() {
		target = 1
	} else if a.sendButton.Hovered() {
		target = 0.6
	}
	if a.sendHoverTime.IsZero() || gtx.Now.Before(a.sendHoverTime) {
		a.sendHoverTime = gtx.Now
	}
	elapsed := min(float32(gtx.Now.Sub(a.sendHoverTime).Seconds()), 0.05)
	a.sendHoverTime = gtx.Now
	diff := target - a.sendHover
	step := elapsed * 8
	if float32(math.Abs(float64(diff))) <= step || diff == 0 {
		a.sendHover = target
	} else {
		a.sendHover += float32(math.Copysign(float64(step), float64(diff)))
		gtx.Execute(op.InvalidateCmd{})
	}
	mix := func(x, y uint8) uint8 { return uint8(float32(x) + (float32(y)-float32(x))*a.sendHover) }
	from, to := a.palette.AccentDim, a.palette.Accent
	background := color.NRGBA{R: mix(from.R, to.R), G: mix(from.G, to.G), B: mix(from.B, to.B), A: 255}
	foreground := color.NRGBA{R: mix(a.palette.Text.R, a.palette.Rail.R), G: mix(a.palette.Text.G, a.palette.Rail.G), B: mix(a.palette.Text.B, a.palette.Rail.B), A: 255}
	return a.sendButton.Layout(gtx, fixedWidth(36, fixedHeight(36, func(gtx layout.Context) layout.Dimensions {
		return rounded(gtx, background, 18, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, fixedWidth(18, func(gtx layout.Context) layout.Dimensions { return a.sendIcon.Layout(gtx, foreground) }))
		})
	})))
}
