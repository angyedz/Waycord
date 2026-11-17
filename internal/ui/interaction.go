package ui

import (
	"image/color"
	"math"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
)

type interactionState struct {
	value float32
	last  time.Time
}

type panelTransition struct {
	value float32
	last  time.Time
}

func (p *panelTransition) advance(now time.Time, open bool) float32 {
	if p.last.IsZero() || now.Before(p.last) {
		p.last = now
	}
	elapsed := now.Sub(p.last)
	if elapsed > 50*time.Millisecond {
		elapsed = 50 * time.Millisecond
	}
	p.last = now
	target := float32(0)
	if open {
		target = 1
	}
	p.value = advanceInteraction(p.value, target, elapsed)
	return p.value
}

func (p *panelTransition) layoutProgress(gtx layout.Context, open bool) float32 {
	value := p.advance(gtx.Now, open)
	if (open && value < 1) || (!open && value > 0) {
		gtx.Execute(op.InvalidateCmd{})
	}
	return value
}

// advanceInteraction is independent of frame rate and stops exactly at the target.
func advanceInteraction(value, target float32, elapsed time.Duration) float32 {
	step := float32(elapsed.Seconds()) * 8
	if step < 0 {
		step = 0
	}
	if step > 1 {
		step = 1
	}
	if float32(math.Abs(float64(target-value))) <= step {
		return target
	}
	return value + float32(math.Copysign(float64(step), float64(target-value)))
}

func mixColor(from, to color.NRGBA, value float32) color.NRGBA {
	mix := func(a, b uint8) uint8 { return uint8(float32(a) + (float32(b)-float32(a))*value) }
	return color.NRGBA{R: mix(from.R, to.R), G: mix(from.G, to.G), B: mix(from.B, to.B), A: mix(from.A, to.A)}
}

// interactionColor schedules frames only while a visible control is transitioning.
func (a *App) interactionColor(gtx layout.Context, button *widget.Clickable, normal, hover color.NRGBA) color.NRGBA {
	if a.interactions == nil {
		a.interactions = make(map[*widget.Clickable]*interactionState)
	}
	state := a.interactions[button]
	target := float32(0)
	if button.Pressed() {
		target = 1
	} else if button.Hovered() {
		target = 0.6
	}
	if state == nil {
		if target == 0 {
			return normal
		}
		state = &interactionState{last: gtx.Now}
		a.interactions[button] = state
	}
	elapsed := gtx.Now.Sub(state.last)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > 50*time.Millisecond {
		elapsed = 50 * time.Millisecond
	}
	state.last = gtx.Now
	state.value = advanceInteraction(state.value, target, elapsed)
	if state.value != target {
		gtx.Execute(op.InvalidateCmd{})
	}
	if state.value == 0 && target == 0 {
		delete(a.interactions, button)
	}
	return mixColor(normal, hover, state.value)
}

// iconControl provides a consistent animated surface and a labeled hit target.
func (a *App) iconControl(gtx layout.Context, button *widget.Clickable, icon *widget.Icon, size, glyph unit.Dp, label string) layout.Dimensions {
	background := a.interactionColor(gtx, button, color.NRGBA{}, a.palette.AccentDim)
	return button.Layout(gtx, fixedWidth(size, fixedHeight(size, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(label).Add(gtx.Ops)
		return rounded(gtx, background, size/2, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, fixedWidth(glyph, func(gtx layout.Context) layout.Dimensions { return icon.Layout(gtx, a.palette.Text) }))
		})
	})))
}

func (a *App) emojiChoice(gtx layout.Context, button *widget.Clickable, value string, size unit.Dp, textSize unit.Sp) layout.Dimensions {
	background := a.interactionColor(gtx, button, color.NRGBA{}, a.palette.AccentDim)
	return button.Layout(gtx, fixedWidth(size, fixedHeight(size, func(gtx layout.Context) layout.Dimensions {
		semantic.LabelOp(value).Add(gtx.Ops)
		return rounded(gtx, background, size/2, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, a.text(value, textSize, a.palette.Text))
		})
	})))
}
