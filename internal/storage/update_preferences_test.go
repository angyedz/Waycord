package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdatePreferencesDefaultAndPersistence(t *testing.T) {
	store := UpdatePreferencesStore{Root: t.TempDir()}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Channel != "stable" || !got.AutoCheck {
		t.Fatalf("unexpected update defaults: %+v", got)
	}
	want := UpdatePreferences{Channel: "beta", AutoCheck: true}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil || got != want {
		t.Fatalf("saved preferences not restored: got=%+v err=%v", got, err)
	}
}

func TestUpdatePreferencesRecoverPreviousValidSlot(t *testing.T) {
	store := UpdatePreferencesStore{Root: t.TempDir()}
	if err := store.Save(UpdatePreferences{Channel: "stable"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(UpdatePreferences{Channel: "nightly", AutoCheck: true}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.Root, "update-preferences-1.json"), []byte("torn write"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil || got != (UpdatePreferences{Channel: "stable"}) {
		t.Fatalf("previous valid preferences not recovered: got=%+v err=%v", got, err)
	}
}

func TestUpdatePreferencesRejectInvalidChannel(t *testing.T) {
	store := UpdatePreferencesStore{Root: t.TempDir()}
	if err := store.Save(UpdatePreferences{Channel: "preview"}); err == nil {
		t.Fatal("invalid release channel was accepted")
	}
}
