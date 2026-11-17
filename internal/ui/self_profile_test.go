package ui

import (
	"errors"
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

func TestSelfProfileChangesPersistAndSaveErrorsKeepEditorOpen(t *testing.T) {
	app := New()
	var savedName, savedStatus string
	var savedPresence int
	app.ConfigureLocalProfile("angyedz", "On a break", 3, func(name, status string, presence int) error {
		savedName, savedStatus, savedPresence = name, status, presence
		return nil
	})
	app.selfProfileEditing = true
	app.selfProfileNameEditor.SetText("  Mira  ")
	app.selfProfileStatusEditor.SetText("  At lunch  ")
	app.saveSelfProfile()
	if savedName != "Mira" || savedStatus != "At lunch" || savedPresence != 3 {
		t.Fatalf("profile save callback got %q, %q, %d", savedName, savedStatus, savedPresence)
	}
	if app.selfProfileName != "Mira" || app.selfProfileStatus != "At lunch" || app.selfProfileEditing {
		t.Fatal("saved profile was not applied to the UI")
	}
	app.saveLocalProfile = func(string, string, int) error { return errors.New("disk full") }
	app.selfProfileEditing = true
	app.selfProfileNameEditor.SetText("Not saved")
	app.saveSelfProfile()
	if app.selfProfileName != "Mira" || !app.selfProfileEditing || app.selfProfileError == "" {
		t.Fatal("failed profile save should preserve the current identity and keep the editor open")
	}
}

func TestSelfProfilePresencePersistsBeforeClosing(t *testing.T) {
	app := New()
	app.selfProfileOpen = true
	app.ConfigureLocalProfile("Mira", "At lunch", 1, func(name, status string, presence int) error {
		if name != "Mira" || status != "At lunch" || presence != 2 {
			t.Fatalf("unexpected persisted presence: %q, %q, %d", name, status, presence)
		}
		return nil
	})
	app.setLocalPresence(2)
	if app.selfPresence != 2 || app.selfProfileOpen {
		t.Fatal("presence change was not applied and dismissed")
	}
	app.selfProfileOpen = true
	app.saveLocalProfile = func(string, string, int) error { return errors.New("disk full") }
	app.setLocalPresence(0)
	if app.selfPresence != 2 || !app.selfProfileOpen || app.selfProfileError == "" {
		t.Fatal("failed presence save should leave the panel open and preserve the previous status")
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
