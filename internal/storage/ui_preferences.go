package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

type UIPreferences struct {
	Theme           string `json:"theme"`
	CompactMessages bool   `json:"compact_messages"`
	ReduceMotion    bool   `json:"reduce_motion"`
	DisplayName     string `json:"display_name,omitempty"`
	CustomStatus    string `json:"custom_status,omitempty"`
	Presence        int    `json:"presence,omitempty"`
	AvatarDigest    string `json:"avatar_digest,omitempty"`
}

type uiPreferencesEnvelope struct {
	Generation  uint64        `json:"generation"`
	Preferences UIPreferences `json:"preferences"`
	Digest      string        `json:"digest"`
}

// UIPreferencesStore keeps two generations so a torn write can fall back to
// the previous file on Linux and Windows.
type UIPreferencesStore struct{ Root string }

const maxUIPreferencesSize = 4 << 10

var errUIPreferencesCorrupt = errors.New("both UI preference files are invalid")

func (s UIPreferencesStore) path(slot int) string {
	return filepath.Join(s.Root, fmt.Sprintf("ui-preferences-%d.json", slot))
}

func validUIPreferences(p UIPreferences) bool {
	validTheme := p.Theme == "dark" || p.Theme == "midnight" || p.Theme == "light"
	profileUnset := p.DisplayName == "" && p.CustomStatus == "" && p.Presence == 0
	return validTheme && validAvatarDigest(p.AvatarDigest) && (profileUnset || validLocalProfile(p.DisplayName, p.CustomStatus, p.Presence))
}

func uiPreferencesDigest(generation uint64, p UIPreferences) string {
	data, _ := json.Marshal(struct {
		Generation  uint64        `json:"generation"`
		Preferences UIPreferences `json:"preferences"`
	}{generation, p})
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func (s UIPreferencesStore) Load() (UIPreferences, error) {
	preferences, _, _, err := s.load()
	return preferences, err
}

func (s UIPreferencesStore) load() (UIPreferences, uint64, int, error) {
	preferences := UIPreferences{Theme: "dark"}
	var generation uint64
	active, seen := -1, false
	for slot := range 2 {
		file, err := os.Open(s.path(slot))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return UIPreferences{}, 0, -1, err
		}
		seen = true
		data, readErr := io.ReadAll(io.LimitReader(file, maxUIPreferencesSize+1))
		closeErr := file.Close()
		if readErr != nil {
			return UIPreferences{}, 0, -1, readErr
		}
		if closeErr != nil {
			return UIPreferences{}, 0, -1, closeErr
		}
		if len(data) > maxUIPreferencesSize {
			continue
		}
		var envelope uiPreferencesEnvelope
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&envelope) != nil || decoder.Decode(new(any)) != io.EOF || envelope.Generation == 0 ||
			!validUIPreferences(envelope.Preferences) || envelope.Digest != uiPreferencesDigest(envelope.Generation, envelope.Preferences) {
			continue
		}
		if envelope.Generation > generation {
			preferences, generation, active = envelope.Preferences, envelope.Generation, slot
		}
	}
	if seen && active < 0 {
		return UIPreferences{}, 0, -1, errUIPreferencesCorrupt
	}
	return preferences, generation, active, nil
}

func (s UIPreferencesStore) Save(preferences UIPreferences) error {
	if !validUIPreferences(preferences) {
		return errors.New("invalid UI preferences")
	}
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		return err
	}
	_, generation, active, err := s.load()
	if errors.Is(err, errUIPreferencesCorrupt) {
		generation, active, err = 0, -1, nil
	}
	if err != nil {
		return err
	}
	slot := 0
	if active == 0 {
		slot = 1
	}
	generation++
	envelope := uiPreferencesEnvelope{Generation: generation, Preferences: preferences}
	envelope.Digest = uiPreferencesDigest(generation, preferences)
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Root, ".ui-preferences-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.path(slot)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmpPath, s.path(slot))
}

// SaveAppearance merges the appearance section into the current preferences so
// changing a theme cannot erase the locally stored profile.
func (s UIPreferencesStore) SaveAppearance(appearance UIPreferences) error {
	if !validUIPreferences(appearance) {
		return errors.New("invalid UI preferences")
	}
	current, err := s.loadForUpdate()
	if err != nil {
		return err
	}
	current.Theme = appearance.Theme
	current.CompactMessages = appearance.CompactMessages
	current.ReduceMotion = appearance.ReduceMotion
	return s.Save(current)
}

// SaveProfile merges the local mock identity into the current preferences.
func (s UIPreferencesStore) SaveProfile(displayName, customStatus string, presence int) error {
	if !validLocalProfile(displayName, customStatus, presence) {
		return errors.New("invalid local profile")
	}
	current, err := s.loadForUpdate()
	if err != nil {
		return err
	}
	current.DisplayName = displayName
	current.CustomStatus = customStatus
	current.Presence = presence
	return s.Save(current)
}

// SaveAvatarDigest merges the content-addressed profile image into the existing
// preferences without overwriting appearance or profile fields.
func (s UIPreferencesStore) SaveAvatarDigest(digest string) error {
	if !validAvatarDigest(digest) {
		return errors.New("invalid avatar digest")
	}
	current, err := s.loadForUpdate()
	if err != nil {
		return err
	}
	current.AvatarDigest = digest
	return s.Save(current)
}

func validAvatarDigest(digest string) bool {
	if digest == "" {
		return true
	}
	if len(digest) != 64 {
		return false
	}
	for _, char := range digest {
		if !(char >= '0' && char <= '9' || char >= 'a' && char <= 'f') {
			return false
		}
	}
	return true
}

func (s UIPreferencesStore) loadForUpdate() (UIPreferences, error) {
	current, err := s.Load()
	if errors.Is(err, errUIPreferencesCorrupt) {
		return UIPreferences{Theme: "dark"}, nil
	}
	return current, err
}

func validLocalProfile(displayName, customStatus string, presence int) bool {
	displayName = strings.TrimSpace(displayName)
	return displayName != "" && utf8.RuneCountInString(displayName) <= 32 &&
		utf8.RuneCountInString(customStatus) <= 80 && presence >= 0 && presence < 4
}
