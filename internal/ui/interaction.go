package ui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
)

type interactionState struct {
	transition panelTransition
}

type panelTransition struct {
	value    float32
	from     float32
	target   float32
	started  time.Time
	duration time.Duration
}

func (p *panelTransition) advance(now time.Time, open bool) float32 {
	target := float32(0)
	if open {
		target = 1
	}
	return p.advanceTo(now, target, 220*time.Millisecond)
}

func (p *panelTransition) advanceTo(now time.Time, target float32, duration time.Duration) float32 {
	if p.started.IsZero() || target != p.target || now.Before(p.started) {
		p.from, p.target, p.started, p.duration = p.value, target, now, duration
	}
	if p.duration <= 0 {
		p.value = p.target
		return p.value
	}
	progress := float32(now.Sub(p.started)) / float32(p.duration)
	if progress <= 0 {
		p.value = p.from
		return p.value
	}
	if progress >= 1 {
		p.value = p.target
		return p.value
	}
	// Smoothstep has zero velocity at both ends, avoiding the abrupt start
	// and stop that made panels feel like they snapped into place.
	eased := progress * progress * (3 - 2*progress)
	p.value = p.from + (p.target-p.from)*eased
	return p.value
}

func (p *panelTransition) layoutProgress(gtx layout.Context, open bool) float32 {
	value := p.advance(gtx.Now, open)
	if value != p.target {
		gtx.Execute(op.InvalidateCmd{})
	}
	return value
}

func boolProgress(open bool) float32 {
	if open {
		return 1
	}
	return 0
}

func (a *App) transition(gtx layout.Context, state *panelTransition, target float32, duration time.Duration) float32 {
	if a.reduceMotion {
		state.value, state.from, state.target, state.duration = target, target, target, 0
		state.started = gtx.Now
		return target
	}
	value := state.advanceTo(gtx.Now, target, duration)
	if value != target {
		gtx.Execute(op.InvalidateCmd{})
	}
	return value
}

func animatedReveal(gtx layout.Context, progress float32, height unit.Dp, child layout.Widget) layout.Dimensions {
	if progress <= 0 {
		return layout.Dimensions{}
	}
	fullHeight := gtx.Dp(height)
	visibleHeight := int(float32(fullHeight) * progress)
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	area := clip.Rect{Max: image.Pt(gtx.Constraints.Max.X, visibleHeight)}.Push(gtx.Ops)
	offset := op.Offset(image.Pt(0, visibleHeight-fullHeight)).Push(gtx.Ops)
	childContext := gtx
	childContext.Constraints.Min.Y, childContext.Constraints.Max.Y = fullHeight, fullHeight
	child(childContext)
	offset.Pop()
	area.Pop()
	return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, visibleHeight)}
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
		state = &interactionState{}
		a.interactions[button] = state
	}
	value := a.transition(gtx, &state.transition, target, 130*time.Millisecond)
	if value == 0 && target == 0 {
		delete(a.interactions, button)
	}
	return mixColor(normal, hover, value)
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

func (a *App) iconToggle(gtx layout.Context, button *widget.Clickable, icon *widget.Icon, size, glyph unit.Dp, label string, active bool) layout.Dimensions {
	background := color.NRGBA{}
	if active {
		background = a.palette.AccentDim
	}
	background = a.interactionColor(gtx, button, background, a.palette.Surface2)
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
