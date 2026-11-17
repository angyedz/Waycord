package ota

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// State is a small version pointer, not the executable itself.
type State struct {
	Current  string `json:"current"`
	Previous string `json:"previous,omitempty"`
	Pending  string `json:"pending,omitempty"`
}

type stateEnvelope struct {
	Generation uint64 `json:"generation"`
	State      State  `json:"state"`
	Digest     string `json:"digest"`
}

type Store struct{ Root string }

func (s Store) statePath(index int) string {
	return filepath.Join(s.Root, fmt.Sprintf("state-%d.json", index))
}

func checkState(st State) error {
	for _, v := range []string{st.Current, st.Previous, st.Pending} {
		if v != "" && !validVersion.MatchString(v) {
			return fmt.Errorf("unsafe version in state: %q", v)
		}
	}
	return nil
}

func stateDigest(generation uint64, state State) string {
	b, _ := json.Marshal(struct {
		Generation uint64 `json:"generation"`
		State      State  `json:"state"`
	}{generation, state})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (s Store) Load() (State, uint64, int, error) {
	var selected State
	var generation uint64
	slot := -1
	seen := false
	for i := 0; i < 2; i++ {
		raw, err := os.ReadFile(s.statePath(i))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return State{}, 0, -1, err
		}
		seen = true
		var env stateEnvelope
		if json.Unmarshal(raw, &env) != nil || env.Generation == 0 || env.Digest != stateDigest(env.Generation, env.State) || checkState(env.State) != nil {
			continue
		}
		if env.Generation > generation {
			selected = env.State
			generation = env.Generation
			slot = i
		}
	}
	if seen && slot < 0 {
		return State{}, 0, -1, errors.New("both updater state files are invalid")
	}
	return selected, generation, slot, nil
}

// Save always writes to the inactive slot, leaving the previous valid one intact.
// This avoids depending on replacement-rename semantics on Windows.
func (s Store) Save(next State) error {
	if err := checkState(next); err != nil {
		return err
	}
	if err := os.MkdirAll(s.Root, 0700); err != nil {
		return err
	}
	_, generation, active, err := s.Load()
	if err != nil {
		return err
	}
	slot := 0
	if active == 0 {
		slot = 1
	}
	generation++
	env := stateEnvelope{Generation: generation, State: next, Digest: stateDigest(generation, next)}
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.Root, ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	// Only the inactive slot is removed; the last valid state remains readable.
	if err = os.Remove(s.statePath(slot)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return os.Rename(tmp.Name(), s.statePath(slot))
}

func (s Store) Stage(version string) error {
	if !validVersion.MatchString(version) {
		return errors.New("invalid version")
	}
	st, _, _, err := s.Load()
	if err != nil {
		return err
	}
	if st.Pending != "" {
		return errors.New("another update is already pending")
	}
	if st.Current == version {
		return errors.New("version is already active")
	}
	st.Pending = version
	return s.Save(st)
}

func (s Store) Promote() error {
	st, _, _, err := s.Load()
	if err != nil {
		return err
	}
	if st.Pending == "" {
		return errors.New("no pending update")
	}
	st.Previous, st.Current, st.Pending = st.Current, st.Pending, ""
	return s.Save(st)
}

func (s Store) Rollback() error {
	st, _, _, err := s.Load()
	if err != nil {
		return err
	}
	if st.Pending == "" {
		return errors.New("no pending update")
	}
	st.Pending = ""
	return s.Save(st)
}
