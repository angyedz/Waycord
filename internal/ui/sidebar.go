package ui

import (
	"image"
	"image/color"
	"math"
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
	last           time.Time
	drag           gesture.Drag
	pressX         float32
}

func newSidebarState() sidebarState {
	return sidebarState{shown: true, width: 200, current: 200, frameWidth: 200}
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
}

func (s *sidebarState) animate(gtx layout.Context) unit.Dp {
	target := float32(0)
	if s.shown {
		target = s.width
	}
	if s.last.IsZero() || gtx.Now.Before(s.last) {
		s.last = gtx.Now
	}
	elapsed := float32(gtx.Now.Sub(s.last).Seconds())
	s.last = gtx.Now
	// Clamp elapsed time so returning from a suspended window remains predictable.
	if elapsed > 0.05 {
		elapsed = 0.05
	}
	distance := target - s.current
	step := float32(1400) * elapsed
	if float32(math.Abs(float64(distance))) <= step || distance == 0 {
		s.current = target
	} else {
		s.current += float32(math.Copysign(float64(step), float64(distance)))
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
