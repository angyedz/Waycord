package ui

import (
	"testing"
	"time"
)

func TestInteractionTransitionStopsAtTarget(t *testing.T) {
	value := advanceInteraction(0, 1, 50*time.Millisecond)
	if value <= 0 || value >= 1 {
		t.Fatalf("expected intermediate value, got %v", value)
	}
	value = advanceInteraction(value, 1, time.Second)
	if value != 1 {
		t.Fatalf("did not settle: %v", value)
	}
	value = advanceInteraction(value, 0, time.Second)
	if value != 0 {
		t.Fatalf("did not return: %v", value)
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
	if closing >= opening || closing < 0 {
		t.Fatalf("expected smooth reversal: %v -> %v", opening, closing)
	}
	if end := state.advance(start.Add(time.Second), false); end != 0 {
		t.Fatalf("did not close: %v", end)
	}
}
