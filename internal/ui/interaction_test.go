package ui

import (
	"testing"
	"time"
)

func TestPanelTransitionEasesAndSettles(t *testing.T) {
	var state panelTransition
	start := time.Unix(1, 0)
	state.advanceTo(start, 0.6, 130*time.Millisecond)
	early := state.advanceTo(start.Add(13*time.Millisecond), 0.6, 130*time.Millisecond)
	if early <= 0 || early >= 0.1 {
		t.Fatalf("expected a soft easing start, got %v", early)
	}
	middle := state.advanceTo(start.Add(65*time.Millisecond), 0.6, 130*time.Millisecond)
	if middle <= early || middle >= 0.6 {
		t.Fatalf("expected intermediate progress, got %v", middle)
	}
	if end := state.advanceTo(start.Add(time.Second), 0.6, 130*time.Millisecond); end != 0.6 {
		t.Fatalf("did not settle at the target: %v", end)
	}
}

func TestProfileTransitionOpensClosesAndReverses(t *testing.T) {
	var state panelTransition
	start := time.Unix(1, 0)
	if v := state.advance(start, true); v != 0 {
		t.Fatalf("initial progress: %v", v)
	}
	opening := state.advance(start.Add(50*time.Millisecond), true)
	if opening <= 0 || opening >= 1 {
		t.Fatalf("expected an intermediate opening, got %v", opening)
	}
	closing := state.advance(start.Add(75*time.Millisecond), false)
	if closing != opening {
		t.Fatalf("reversal jumped from %v to %v", opening, closing)
	}
	closing = state.advance(start.Add(100*time.Millisecond), false)
	if closing >= opening || closing < 0 {
		t.Fatalf("expected smooth reverse movement: %v -> %v", opening, closing)
	}
	if end := state.advance(start.Add(time.Second), false); end != 0 {
		t.Fatalf("did not close: %v", end)
	}
}
