package ui

import "testing"

func TestSidebarDragUsesLastDrawnPositionForAllEvents(t *testing.T) {
	s := newSidebarState()
	s.frameWidth = 200
	s.pressX = 3
	// Both pointer positions come from the same drawn frame, even if delivered together.
	s.dragTo(20, 1)
	s.dragTo(30, 1)
	if s.width != 227 {
		t.Fatalf("width after drag batch = %v, want 227", s.width)
	}
	// Once the handle has been redrawn, coordinates are relative to its new position.
	s.frameWidth = 227
	s.dragTo(8, 1)
	if s.width != 232 {
		t.Fatalf("width after redraw = %v, want 232", s.width)
	}
}
