//go:build !windows

package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalAvatarStoreCreatesOwnerOnlyFile(t *testing.T) {
	store := LocalAvatarStore{Root: t.TempDir()}
	digest, err := store.Save(testAvatarPNG(t))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Root, "avatars", digest+".image")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("avatar file permissions = %o, want 600", info.Mode().Perm())
	}
	if _, err := store.Load(digest); err != nil {
		t.Fatalf("owner-only file could not be read: %v", err)
	}
}
