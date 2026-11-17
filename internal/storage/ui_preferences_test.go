package storage

import (
	"os"
	"path/filepath"
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
