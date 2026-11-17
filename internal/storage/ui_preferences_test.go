package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUIPreferencesStoreRoundTripAndRecovery(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	defaults, err := store.Load()
	if err != nil || defaults != (UIPreferences{Theme: "dark"}) {
		t.Fatalf("unexpected defaults: %+v, %v", defaults, err)
	}
	first := UIPreferences{Theme: "midnight", CompactMessages: true}
	if err := store.Save(first); err != nil {
		t.Fatal(err)
	}
	second := UIPreferences{Theme: "light", ReduceMotion: true}
	if err := store.Save(second); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Load(); err != nil || got != second {
		t.Fatalf("latest preferences did not load: %+v, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(store.Root, "ui-preferences-1.json"), []byte("truncated"), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := store.Load(); err != nil || got != first {
		t.Fatalf("previous generation was not recovered: %+v, %v", got, err)
	}
}

func TestUIPreferencesStoreRejectsInvalidAndUnknownFields(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	if err := store.Save(UIPreferences{Theme: "neon"}); err == nil {
		t.Fatal("invalid theme was accepted")
	}
	if err := os.MkdirAll(store.Root, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.Root, "ui-preferences-0.json"), []byte(`{"generation":1,"preferences":{"theme":"dark","surprise":true},"digest":"bad"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(); err == nil {
		t.Fatal("unknown fields in persisted preferences were accepted")
	}
	if err := store.Save(UIPreferences{Theme: "dark", CompactMessages: true}); err != nil {
		t.Fatalf("could not recover from corrupt preferences: %v", err)
	}
	if got, err := store.Load(); err != nil || !got.CompactMessages {
		t.Fatalf("recovered preferences did not load: %+v, %v", got, err)
	}
}

func TestUIPreferencesStoreMergesAppearanceAndLocalProfileUpdates(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	initial := UIPreferences{Theme: "midnight", CompactMessages: true, DisplayName: "mira", CustomStatus: "At lunch", Presence: 1}
	if err := store.Save(initial); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveProfile("angyedz", "Building Waycord", 3); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAppearance(UIPreferences{Theme: "light", ReduceMotion: true}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := UIPreferences{Theme: "light", ReduceMotion: true, DisplayName: "angyedz", CustomStatus: "Building Waycord", Presence: 3}
	if got != want {
		t.Fatalf("profile and appearance updates did not merge: got %+v, want %+v", got, want)
	}
}

func TestUIPreferencesStoreRejectsInvalidLocalProfile(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	for _, profile := range []struct {
		name, status string
		presence     int
	}{{" ", "", 0}, {"Mira", strings.Repeat("x", 81), 0}, {"Mira", "", 4}} {
		if err := store.SaveProfile(profile.name, profile.status, profile.presence); err == nil {
			t.Fatalf("accepted invalid profile %+v", profile)
		}
	}
}

func TestUIPreferencesStorePersistsAvatarDigestAlongsideProfileAndAppearance(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	if err := store.SaveProfile("Mira", "At lunch", 1); err != nil {
		t.Fatal(err)
	}
	digest := strings.Repeat("a", 64)
	if err := store.SaveAvatarDigest(digest); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveAppearance(UIPreferences{Theme: "midnight", CompactMessages: true}); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := UIPreferences{Theme: "midnight", CompactMessages: true, DisplayName: "Mira", CustomStatus: "At lunch", Presence: 1, AvatarDigest: digest}
	if got != want {
		t.Fatalf("avatar digest was not merged with other preferences: got %+v, want %+v", got, want)
	}
}

func TestUIPreferencesStoreRejectsInvalidAvatarDigest(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	for _, digest := range []string{"../avatar.png", "not-a-digest", strings.Repeat("a", 63), strings.Repeat("z", 64)} {
		if err := store.SaveAvatarDigest(digest); err == nil {
			t.Errorf("accepted invalid avatar digest %q", digest)
		}
	}
}

func TestUIPreferencesStoreLoadsPreferencesWrittenBeforeLocalProfileFields(t *testing.T) {
	store := UIPreferencesStore{Root: t.TempDir()}
	legacy := UIPreferences{Theme: "dark", CompactMessages: true}
	envelope := uiPreferencesEnvelope{Generation: 1, Preferences: legacy}
	envelope.Digest = uiPreferencesDigest(envelope.Generation, envelope.Preferences)
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store.path(0), data, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got != legacy {
		t.Fatalf("legacy settings did not load: got %+v, err %v", got, err)
	}
}
