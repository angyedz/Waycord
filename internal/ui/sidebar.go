package ui

import (
	"image"
	"image/color"
	"time"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

const minSidebarWidth = 160
const maxSidebarWidth = 360

// sidebarState retains the chosen width while the panel is collapsed.
type sidebarState struct {
	shown          bool
	width, current float32
	frameWidth     float32
	transition     panelTransition
	drag           gesture.Drag
	pressX         float32
}

func newSidebarState() sidebarState {
	return sidebarState{shown: true, width: 200, current: 200, frameWidth: 200, transition: panelTransition{value: 200, target: 200}}
}
func (s *sidebarState) toggle() { s.shown = !s.shown }
func clampSidebarWidth(width float32) float32 {
	return max(minSidebarWidth, min(maxSidebarWidth, width))
}

func (s *sidebarState) update(gtx layout.Context) {
	for {
		e, ok := s.drag.Update(gtx.Metric, gtx.Source, gesture.Horizontal)
		if !ok {
			break
		}
		switch e.Kind {
		case pointer.Press:
			s.pressX = e.Position.X
		case pointer.Drag:
			if s.shown && gtx.Metric.PxPerDp > 0 {
				s.dragTo(e.Position.X, gtx.Metric.PxPerDp)
			}
		}
	}
}

func (s *sidebarState) dragTo(position, pixelsPerDP float32) {
	s.width = clampSidebarWidth(s.frameWidth + (position-s.pressX)/pixelsPerDP)
	s.current = s.width
	s.transition.value, s.transition.from, s.transition.target = s.current, s.current, s.current
	s.transition.started = time.Time{}
}

func (s *sidebarState) animate(gtx layout.Context, reduceMotion bool) unit.Dp {
	target := float32(0)
	if s.shown {
		target = s.width
	}
	if reduceMotion {
		s.current = target
		s.transition.value, s.transition.from, s.transition.target, s.transition.duration = target, target, target, 0
		s.transition.started = gtx.Now
		return unit.Dp(s.current)
	}
	s.current = s.transition.advanceTo(gtx.Now, target, 240*time.Millisecond)
	if s.current != target {
		gtx.Execute(op.InvalidateCmd{})
	}
	return unit.Dp(s.current)
}

func (s *sidebarState) handle(gtx layout.Context, border color.NRGBA) layout.Dimensions {
	if gtx.Metric.PxPerDp > 0 {
		s.frameWidth = float32(gtx.Dp(unit.Dp(s.current))) / gtx.Metric.PxPerDp
	}
	size := image.Pt(gtx.Dp(6), gtx.Constraints.Max.Y)
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	paint.FillShape(gtx.Ops, border, clip.Rect{Min: image.Pt(size.X/2, 0), Max: image.Pt(size.X/2+gtx.Dp(1), size.Y)}.Op())
	pointer.CursorColResize.Add(gtx.Ops)
	s.drag.Add(gtx.Ops)
	area.Pop()
	return layout.Dimensions{Size: size}
}

func (a *App) navigationDivider(gtx layout.Context) layout.Dimensions {
	width := max(1, gtx.Dp(1))
	height := gtx.Constraints.Max.Y
	inset := min(gtx.Dp(12), height/2)
	paint.FillShape(gtx.Ops, a.palette.Border, clip.Rect{Min: image.Pt(0, inset), Max: image.Pt(width, height-inset)}.Op())
	return layout.Dimensions{Size: image.Pt(width, height)}
}
