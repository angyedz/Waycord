// Package storage contains small user-scoped persistence adapters.
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
)

type UpdatePreferences struct {
	Channel   string `json:"channel"`
	AutoCheck bool   `json:"auto_check"`
}

type preferencesEnvelope struct {
	Generation  uint64            `json:"generation"`
	Preferences UpdatePreferences `json:"preferences"`
	Digest      string            `json:"digest"`
}

type UpdatePreferencesStore struct{ Root string }

const maxPreferencesSize = 64 << 10

func (s UpdatePreferencesStore) path(slot int) string {
	return filepath.Join(s.Root, fmt.Sprintf("update-preferences-%d.json", slot))
}

func validUpdatePreferences(p UpdatePreferences) bool {
	return p.Channel == "stable" || p.Channel == "beta" || p.Channel == "nightly"
}

func preferencesDigest(generation uint64, p UpdatePreferences) string {
	b, _ := json.Marshal(struct {
		Generation  uint64            `json:"generation"`
		Preferences UpdatePreferences `json:"preferences"`
	}{generation, p})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// Load returns safe defaults when no preference file exists. If one
// slot is torn, the other valid generation remains available for recovery.
func (s UpdatePreferencesStore) Load() (UpdatePreferences, error) {
	preferences, _, _, err := s.load()
	return preferences, err
}

func (s UpdatePreferencesStore) load() (UpdatePreferences, uint64, int, error) {
	selected := UpdatePreferences{Channel: "stable", AutoCheck: true}
	var generation uint64
	active := -1
	seen := false
	for slot := range 2 {
		file, err := os.Open(s.path(slot))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return UpdatePreferences{}, 0, -1, err
		}
		seen = true
		raw, readErr := io.ReadAll(io.LimitReader(file, maxPreferencesSize+1))
		closeErr := file.Close()
		if readErr != nil {
			return UpdatePreferences{}, 0, -1, readErr
		}
		if closeErr != nil {
			return UpdatePreferences{}, 0, -1, closeErr
		}
		if len(raw) > maxPreferencesSize {
			continue
		}
		var envelope preferencesEnvelope
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&envelope) != nil || decoder.Decode(new(any)) != io.EOF ||
			envelope.Generation == 0 || !validUpdatePreferences(envelope.Preferences) ||
			envelope.Digest != preferencesDigest(envelope.Generation, envelope.Preferences) {
			continue
		}
		if envelope.Generation > generation {
			selected, generation, active = envelope.Preferences, envelope.Generation, slot
		}
	}
	if seen && active < 0 {
		return UpdatePreferences{}, 0, -1, errors.New("both update preference files are invalid")
	}
	return selected, generation, active, nil
}

func (s UpdatePreferencesStore) Save(p UpdatePreferences) error {
	if !validUpdatePreferences(p) {
		return errors.New("invalid update channel")
	}
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		return err
	}
	_, generation, active, err := s.load()
	if err != nil {
		return err
	}
	slot := 0
	if active == 0 {
		slot = 1
	}
	generation++
	envelope := preferencesEnvelope{Generation: generation, Preferences: p, Digest: preferencesDigest(generation, p)}
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Root, ".update-preferences-*")
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
