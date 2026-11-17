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

	"github.com/angyedz/Waycord/internal/ui/theme"
)

func TestUIPreferencesPersistBeforeApplying(t *testing.T) {
	app := New()
	app.ConfigureUIPreferences("dark", false, false, func(string, bool, bool) error { return errors.New("disk full") })
	app.setUIPreferences("light", true, true)
	if app.themeName != "dark" || app.compactMessages || app.reduceMotion || app.palette != theme.Dark {
		t.Fatal("settings were applied despite a failed save")
	}
	if app.settingsStatus == "" {
		t.Fatal("save failure was not visible to the user")
	}
	app.ConfigureUIPreferences("midnight", true, false, nil)
	if app.themeName != "midnight" || !app.compactMessages || app.palette != theme.Midnight {
		t.Fatal("loaded UI preferences were not applied")
	}
}

func TestReduceMotionSkipsTransitionFrames(t *testing.T) {
	app := New()
	app.reduceMotion = true
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Now: time.Now()}
	var transition panelTransition
	if got := app.transition(gtx, &transition, 1, 220*time.Millisecond); got != 1 {
		t.Fatalf("reduced motion did not finish immediately: %f", got)
	}
	if transition.value != 1 || transition.duration != 0 {
		t.Fatalf("transition state was not snapped to target: %+v", transition)
	}
}

func TestAudioSettingsControlsUpdateVoicePreview(t *testing.T) {
	app := New()
	clickPreference(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.preferenceRow(gtx, "Microphone", "Microphone is unmuted", &app.voiceMute, false)
	})
	if !app.voice.Muted {
		t.Fatal("microphone control did not mute the local voice state")
	}
	clickPreference(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.preferenceRow(gtx, "Deafen", "Headphones are on", &app.voiceDeafen, false)
	})
	if !app.voice.Deafened || !app.voice.Muted {
		t.Fatal("deafen control did not mute the local voice state")
	}
}

func TestVideoSettingsControlsRespectVoiceConnection(t *testing.T) {
	app := New()
	clickPreference(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.preferenceRow(gtx, "Camera", "Camera is off", &app.voiceCamera, false)
	})
	if app.voice.Camera {
		t.Fatal("camera turned on while disconnected")
	}
	app.voice.Join()
	clickPreference(t, app, func(gtx layout.Context) layout.Dimensions {
		return app.preferenceRow(gtx, "Camera", "Camera is off", &app.voiceCamera, false)
	})
	if !app.voice.Camera {
		t.Fatal("camera control did not change the connected voice state")
	}
}

func clickPreference(t *testing.T, app *App, draw layout.Widget) {
	t.Helper()
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Constraints{Max: image.Pt(500, 80)}}
	draw(gtx)
	router.Frame(&ops)
	router.Queue(
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Press, Buttons: pointer.ButtonPrimary, Position: f32.Pt(460, 22)},
		pointer.Event{Source: pointer.Mouse, Kind: pointer.Release, Position: f32.Pt(460, 22)},
	)
	app.process(gtx)
}
