// Package voicepreview models local voice controls without connecting devices.
package voicepreview

type State struct {
	Connected bool
	Muted     bool
	Deafened  bool
	Camera    bool
	Screen    bool
}

func (s *State) Join() { s.Connected = true }

func (s *State) Leave() {
	s.Connected, s.Camera, s.Screen = false, false, false
}

func (s *State) ToggleCamera() {
	if s.Connected {
		s.Camera = !s.Camera
	}
}

func (s *State) ToggleScreen() {
	if s.Connected {
		s.Screen = !s.Screen
	}
}

func (s *State) ToggleMute() {
	if s.Deafened {
		s.Deafened = false
	}
	s.Muted = !s.Muted
}

func (s *State) ToggleDeafen() {
	s.Deafened = !s.Deafened
	if s.Deafened {
		s.Muted = true
	}
}
