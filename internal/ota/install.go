package ota

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "waycord.exe"
	}
	return "waycord"
}

func (s Store) Binary(version string) (string, error) {
	if !validVersion.MatchString(version) {
		return "", errors.New("invalid version path")
	}
	return filepath.Join(s.Root, "versions", version, BinaryName()), nil
}

// ManagesExecutable reports whether executable is the active binary in this
// user-scoped version store. System and package-manager installs stay outside
// the self-update path.
func (s Store) ManagesExecutable(executable, version string) bool {
	if !validVersion.MatchString(version) || executable == "" {
		return false
	}
	state, _, _, err := s.Load()
	if err != nil || state.Current != version {
		return false
	}
	expected, err := s.Binary(version)
	if err != nil {
		return false
	}
	actualResolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return false
	}
	expectedResolved, err := filepath.EvalSymlinks(expected)
	if err != nil {
		return false
	}
	actual, err := filepath.Abs(actualResolved)
	if err != nil {
		return false
	}
	expected, err = filepath.Abs(expectedResolved)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		if !strings.EqualFold(actual, expected) {
			return false
		}
	} else if actual != expected {
		return false
	}
	actualInfo, actualErr := os.Stat(actual)
	expectedInfo, expectedErr := os.Stat(expected)
	return actualErr == nil && expectedErr == nil && actualInfo.Mode().IsRegular() && os.SameFile(actualInfo, expectedInfo)
}

// install stages a payload whose signature was checked by the caller in this package.
func (s Store) install(ctx context.Context, p Payload, client *http.Client) error {
	return s.installWithProgress(ctx, p, client, nil)
}

func (s Store) installWithProgress(ctx context.Context, p Payload, client *http.Client, progress ProgressFunc) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.PlatformMatches() {
		return fmt.Errorf("release targets %s/%s, running %s/%s", p.OS, p.Arch, runtime.GOOS, runtime.GOARCH)
	}
	st, _, _, err := s.Load()
	if err != nil {
		return err
	}
	if st.Pending != "" {
		return errors.New("finish pending update first")
	}
	if st.Current != "" {
		newer, err := IsNewer(p.Version, st.Current)
		if err != nil {
			return err
		}
		if !newer {
			return errors.New("downgrade/reinstall is not allowed")
		}
	}
	binary, err := s.Binary(p.Version)
	if err != nil {
		return err
	}
	if _, err = os.Stat(binary); err == nil {
		return errors.New("version already exists on disk")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	dir := filepath.Dir(binary)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	temp, err := downloadWithProgress(ctx, p, filepath.Join(s.Root, "downloads"), client, progress)
	if err != nil {
		return err
	}
	defer os.Remove(temp)
	if err = os.Chmod(temp, 0700); err != nil {
		return err
	}
	// Rename across volumes may fail. Create a staged file on the same volume.
	dst, err := os.CreateTemp(dir, ".waycord-*")
	if err != nil {
		return err
	}
	staged := dst.Name()
	dst.Close()
	defer os.Remove(staged)
	if err = copyFile(staged, temp); err != nil {
		return err
	}
	if err = os.Chmod(staged, 0700); err != nil {
		return err
	}
	if err = os.Rename(staged, binary); err != nil {
		return err
	}
	if err = s.Stage(p.Version); err != nil {
		os.Remove(binary)
		return err
	}
	return nil
}

// InstallUpdate stages an update returned by GitHubSource.Check. The signed
// manifest is retained privately and verified again here so callers cannot
// alter a payload between release discovery and installation.
func (s Store) InstallUpdate(ctx context.Context, update Update, channel string, key ed25519.PublicKey, progress ProgressFunc) error {
	return s.installUpdateWithClient(ctx, update, channel, key, secureClient(), progress)
}

func (s Store) installUpdateWithClient(ctx context.Context, update Update, channel string, key ed25519.PublicKey, client *http.Client, progress ProgressFunc) error {
	if !update.Available || len(update.manifest) == 0 {
		return errors.New("update has no verified release manifest")
	}
	if !IsChannel(channel) {
		return errors.New("invalid OTA channel")
	}
	payload, err := Verify(update.manifest, key)
	if err != nil {
		return err
	}
	if payload != update.Payload {
		return errors.New("update payload changed after signature verification")
	}
	if payload.Channel != channel {
		return errors.New("signed update does not match the selected channel")
	}
	return s.installWithProgress(ctx, payload, client, progress)
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err = out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// CheckAndInstall can only be called with an explicitly trusted key.
func (s Store) CheckAndInstall(ctx context.Context, manifestURL, channel string, pub ed25519.PublicKey) (Payload, error) {
	return s.checkAndInstall(ctx, manifestURL, channel, pub, secureClient())
}

func (s Store) checkAndInstall(ctx context.Context, manifestURL, channel string, pub ed25519.PublicKey, client *http.Client) (Payload, error) {
	if !IsChannel(channel) {
		return Payload{}, errors.New("invalid OTA channel")
	}
	p, err := fetchManifest(ctx, manifestURL, pub, client)
	if err != nil {
		return Payload{}, err
	}
	if p.Channel != channel {
		return Payload{}, errors.New("manifest channel mismatch")
	}
	if err = s.install(ctx, p, client); err != nil {
		return Payload{}, err
	}
	return p, nil
}
