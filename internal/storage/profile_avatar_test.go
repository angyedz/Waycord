package storage

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalAvatarStoreRoundTrip(t *testing.T) {
	store := LocalAvatarStore{Root: t.TempDir()}
	data := testAvatarPNG(t)
	digest, err := store.Save(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(digest) != 64 {
		t.Fatalf("unexpected digest length %d", len(digest))
	}
	loaded, err := store.Load(digest)
	if err != nil {
		t.Fatal(err)
	}
	if loaded == nil || loaded.Bounds() != image.Rect(0, 0, 4, 4) {
		t.Fatalf("unexpected decoded avatar: %v", loaded)
	}
}

func TestProfilePreferencesResolveTheirSavedAvatar(t *testing.T) {
	root := t.TempDir()
	avatars := LocalAvatarStore{Root: root}
	preferences := UIPreferencesStore{Root: root}
	digest, err := avatars.Save(testAvatarPNG(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := preferences.SaveProfile("Mira", "", 0); err != nil {
		t.Fatal(err)
	}
	if err := preferences.SaveAvatarDigest(digest); err != nil {
		t.Fatal(err)
	}
	loadedPreferences, err := preferences.Load()
	if err != nil {
		t.Fatal(err)
	}
	loadedAvatar, err := avatars.Load(loadedPreferences.AvatarDigest)
	if err != nil {
		t.Fatal(err)
	}
	if loadedAvatar.Bounds() != image.Rect(0, 0, 4, 4) {
		t.Fatalf("unexpected avatar restored from settings: %v", loadedAvatar.Bounds())
	}
}

func TestLocalAvatarStoreRejectsInvalidDataAndDigest(t *testing.T) {
	store := LocalAvatarStore{Root: t.TempDir()}
	if _, err := store.Save([]byte("not an image")); err == nil {
		t.Fatal("invalid image was accepted")
	}
	for _, digest := range []string{"", "../secret", strings.Repeat("a", 63), strings.Repeat("g", 64)} {
		if _, err := store.Load(digest); err == nil {
			t.Errorf("accepted invalid digest %q", digest)
		}
	}
}

func TestLocalAvatarStoreRejectsOversizedDecodedImage(t *testing.T) {
	store := LocalAvatarStore{Root: t.TempDir()}
	large := image.NewNRGBA(image.Rect(0, 0, 2001, 2000))
	var data bytes.Buffer
	if err := png.Encode(&data, large); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(data.Bytes()); err == nil {
		t.Fatal("avatar above the pixel limit was accepted")
	}
}

func TestLocalAvatarStoreDetectsCorruptStoredData(t *testing.T) {
	store := LocalAvatarStore{Root: t.TempDir()}
	data := testAvatarPNG(t)
	digest, err := store.Save(data)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "avatars", digest+".image")
	if err := os.WriteFile(path, []byte("corrupted"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Load(digest); err == nil {
		t.Fatal("corrupted avatar was returned")
	}
	if repairedDigest, err := store.Save(data); err != nil || repairedDigest != digest {
		t.Fatalf("valid avatar could not repair its corrupt cache entry: digest=%q err=%v", repairedDigest, err)
	}
	if _, err := store.Load(digest); err != nil {
		t.Fatalf("repaired avatar could not be loaded: %v", err)
	}
}

func testAvatarPNG(t *testing.T) []byte {
	t.Helper()
	avatar := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			avatar.SetRGBA(x, y, color.RGBA{R: uint8(30 + x*20), G: uint8(80 + y*20), B: 160, A: 255})
		}
	}
	var data bytes.Buffer
	if err := png.Encode(&data, avatar); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
