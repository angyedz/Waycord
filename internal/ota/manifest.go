// Package ota implements signed, user-scoped application updates.
// The updater intentionally has no dependency on a Discord account or GUI.
package ota

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"runtime"
	"strings"
)

const (
	MaxManifestSize = 128 << 10
	MaxArtifactSize = 300 << 20
)

var validVersion = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-(?:beta|nightly)\.(0|[1-9][0-9]*))?$`)

// Payload's field order is part of the v1 signing format. Do not reorder fields
// or add fields without increasing SchemaVersion and updating verification.
type Payload struct {
	SchemaVersion int    `json:"schema_version"`
	Project       string `json:"project"`
	Channel       string `json:"channel"`
	Version       string `json:"version"`
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	URL           string `json:"url"`
	Size          int64  `json:"size"`
	SHA256        string `json:"sha256"`
}

type Manifest struct {
	Payload   Payload `json:"payload"`
	Signature string  `json:"signature"`
}

func IsChannel(channel string) bool {
	return channel == "stable" || channel == "beta" || channel == "nightly"
}

func (p Payload) Validate() error {
	if p.SchemaVersion != 1 || p.Project != "Waycord" {
		return errors.New("unknown OTA manifest schema/project")
	}
	if !IsChannel(p.Channel) || !validVersion.MatchString(p.Version) {
		return errors.New("invalid channel or version")
	}
	if _, err := parseVersion(p.Version); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	if (p.Channel == "stable" && strings.Contains(p.Version, "-")) ||
		(p.Channel == "beta" && !strings.Contains(p.Version, "-beta.")) ||
		(p.Channel == "nightly" && !strings.Contains(p.Version, "-nightly.")) {
		return errors.New("version does not match channel")
	}
	if (p.OS != "linux" && p.OS != "windows") || (p.Arch != "amd64" && p.Arch != "arm64") {
		return errors.New("unsupported OS or architecture")
	}
	u, err := url.Parse(p.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return errors.New("artifact URL must be an absolute HTTPS URL without credentials")
	}
	if p.Size <= 0 || p.Size > MaxArtifactSize {
		return errors.New("artifact size outside allowed bounds")
	}
	sha, err := hex.DecodeString(p.SHA256)
	if err != nil || len(sha) != sha256.Size || p.SHA256 != strings.ToLower(p.SHA256) {
		return errors.New("invalid lowercase sha256 digest")
	}
	return nil
}

func (p Payload) Bytes() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// Verify checks a pinned signing key before interpreting the release payload.
func Verify(raw []byte, publicKey ed25519.PublicKey) (Payload, error) {
	if len(publicKey) != ed25519.PublicKeySize {
		return Payload{}, errors.New("no trusted Ed25519 public key configured")
	}
	if len(raw) == 0 || len(raw) > MaxManifestSize {
		return Payload{}, errors.New("manifest size exceeds limit")
	}
	var m Manifest
	if err := strictJSON(raw, &m); err != nil {
		return Payload{}, err
	}
	body, err := m.Payload.Bytes()
	if err != nil {
		return Payload{}, err
	}
	sig, err := base64.StdEncoding.DecodeString(m.Signature)
	if err != nil || !ed25519.Verify(publicKey, body, sig) {
		return Payload{}, errors.New("OTA signature verification failed")
	}
	return m.Payload, nil
}

func strictJSON(raw []byte, target any) error {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return fmt.Errorf("invalid OTA JSON: %w", err)
	}
	if d.Decode(new(any)) != io.EOF { // reject trailing JSON documents
		return errors.New("trailing data in OTA JSON")
	}
	return nil
}

// PlatformMatches limits signed releases to the running platform.
func (p Payload) PlatformMatches() bool { return p.OS == runtime.GOOS && p.Arch == runtime.GOARCH }
