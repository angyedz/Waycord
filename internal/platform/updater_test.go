package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestManagedUpdaterPathRequiresAnExecutableInsideInstallRoot(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, updaterBinaryName())
	if err := os.WriteFile(path, []byte("updater"), 0700); err != nil {
		t.Fatal(err)
	}
	got, err := ManagedUpdaterPath(root)
	pathMatches := got == path
	if runtime.GOOS == "windows" {
		pathMatches = strings.EqualFold(filepath.Clean(got), filepath.Clean(path))
	}
	if err != nil || !pathMatches {
		t.Fatalf("managed updater path = %q, %v; want %q", got, err, path)
	}

	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if _, err := ManagedUpdaterPath(root); err == nil {
			t.Fatal("accepted a non-executable updater")
		}
	}
}

func TestManagedUpdaterPathRejectsSymlinkOutsideInstallRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation can require elevated privileges on Windows")
	}
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), updaterBinaryName())
	if err := os.WriteFile(outside, []byte("updater"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, updaterBinaryName())); err != nil {
		t.Fatal(err)
	}
	if _, err := ManagedUpdaterPath(root); err == nil {
		t.Fatal("accepted an updater symlink that escapes the managed install root")
	}
}
