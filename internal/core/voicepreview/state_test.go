package voicepreview

import "testing"

func TestVoiceActionsRequireConnectionAndResetOnLeave(t *testing.T) {
	var voice State
	voice.ToggleScreen()
	voice.ToggleCamera()
	if voice.Screen || voice.Camera {
		t.Fatal("started camera or screen while disconnected")
	}
	voice.Join()
	voice.ToggleScreen()
	voice.ToggleCamera()
	if !voice.Screen || !voice.Camera {
		t.Fatal("connected controls did not change")
	}
	voice.Leave()
	if voice.Connected || voice.Screen || voice.Camera {
		t.Fatalf("leave did not stop active sessions: %+v", voice)
	}
}

func TestDeafenAlsoMutes(t *testing.T) {
	var voice State
	voice.ToggleDeafen()
	if !voice.Deafened || !voice.Muted {
		t.Fatal("deafen must also mute")
	}
	voice.ToggleMute()
	if voice.Deafened || voice.Muted {
		t.Fatal("unmuting must clear deafen")
	}
}
