package platform

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ManagedUpdaterPath returns the updater bundled in a managed install root.
// It rejects missing files, non-regular files, and links that escape the root.
func ManagedUpdaterPath(root string) (string, error) {
	if root == "" {
		return "", errors.New("empty updater root")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("resolve updater root: %w", err)
	}
	path := filepath.Join(absoluteRoot, updaterBinaryName())
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve managed updater: %w", err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedPath)
	if err != nil || relative == ".." || filepath.IsAbs(relative) || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator) {
		return "", errors.New("managed updater resolves outside its install root")
	}
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errors.New("managed updater is not a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		return "", errors.New("managed updater is not executable")
	}
	return resolvedPath, nil
}

// StartManagedUpdater starts the launcher for a per-user install. The caller
// should close its window only after this function returns successfully.
func StartManagedUpdater(root string) error {
	if runtime.GOOS == "linux" && os.Geteuid() == 0 {
		return errors.New("refusing to self-update a root-owned application")
	}
	path, err := ManagedUpdaterPath(root)
	if err != nil {
		return err
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	command := exec.Command(path, "launch", "-dir", absoluteRoot)
	command.Stdin = nil
	command.Stdout = nil
	command.Stderr = nil
	if err := command.Start(); err != nil {
		return fmt.Errorf("start update launcher: %w", err)
	}
	return command.Process.Release()
}

func updaterBinaryName() string {
	if runtime.GOOS == "windows" {
		return "waycord-updater.exe"
	}
	return "waycord-updater"
}
