package ota

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

// Install verifies the manifest, downloads and hashes the executable, then stages
// it. The launcher will promote it only after a successful health handshake.
func (s Store) Install(ctx context.Context, p Payload) error {
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
	temp, err := Download(ctx, p, filepath.Join(s.Root, "downloads"))
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
	if !IsChannel(channel) {
		return Payload{}, errors.New("invalid OTA channel")
	}
	p, err := FetchManifest(ctx, manifestURL, pub)
	if err != nil {
		return Payload{}, err
	}
	if p.Channel != channel {
		return Payload{}, errors.New("manifest channel mismatch")
	}
	if err = s.Install(ctx, p); err != nil {
		return Payload{}, err
	}
	return p, nil
}
