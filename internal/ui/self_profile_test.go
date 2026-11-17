package ui

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func TestUserPanelClickOpensOwnProfile(t *testing.T) {
	app := New()
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(260, 100)}}
	app.userPanel(gtx, true)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(24, 20)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(24, 20)})
	app.processSelfProfile(gtx)
	if !app.selfProfileOpen {
		t.Fatal("clicking the user avatar did not open the profile")
	}
}

func TestSelfProfilePanelSlidesOpen(t *testing.T) {
	app := New()
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Now: time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)}
	app.selfProfileOpen = true
	if got := app.selfProfileSlide.layoutProgress(gtx, true); got != 0 {
		t.Fatalf("first animation frame should start at zero, got %f", got)
	}
	gtx.Now = gtx.Now.Add(110 * time.Millisecond)
	if got := app.selfProfileSlide.layoutProgress(gtx, true); got <= 0 || got >= 1 {
		t.Fatalf("profile should be partway through its entrance, got %f", got)
	}
	gtx.Now = gtx.Now.Add(110 * time.Millisecond)
	if got := app.selfProfileSlide.layoutProgress(gtx, true); got != 1 {
		t.Fatalf("profile entrance did not finish, got %f", got)
	}
}

func TestSelfProfileSaveUpdatesLocalIdentity(t *testing.T) {
	app := New()
	app.selfProfileEditing = true
	app.selfProfileNameEditor.SetText("  Mira  ")
	app.selfProfileStatusEditor.SetText("  At lunch  ")
	app.saveSelfProfile()
	if app.selfProfileName != "Mira" || app.selfProfileStatus != "At lunch" || app.selfProfileEditing {
		t.Fatalf("profile changes were not applied: name=%q status=%q editing=%t", app.selfProfileName, app.selfProfileStatus, app.selfProfileEditing)
	}
	app.selfProfileEditing = true
	app.selfProfileNameEditor.SetText("  ")
	app.selfProfileStatusEditor.SetText("")
	app.saveSelfProfile()
	if app.selfProfileName != "Mira" || app.selfProfileStatus != "" {
		t.Fatalf("empty name should be ignored and empty status should clear: name=%q status=%q", app.selfProfileName, app.selfProfileStatus)
	}
}

func TestSelfProfileClickingOutsideClosesCard(t *testing.T) {
	app := New()
	app.selfProfileOpen = true
	app.selfProfileSlide.value = 1
	app.selfProfileSlide.target = 1
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(600, 700)}}
	app.selfProfileOverlay(gtx, gtx.Constraints.Max, 1)
	router.Frame(&ops)
	router.Queue(pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(500, 100)}, pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(500, 100)})
	app.processSelfProfile(gtx)
	if app.selfProfileOpen {
		t.Fatal("clicking outside the profile did not close it")
	}
}
