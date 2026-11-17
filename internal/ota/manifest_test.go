package ota

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func samplePayload() Payload {
	return Payload{SchemaVersion: 1, Project: "Waycord", Channel: "stable", Version: "v0.1.0", OS: "linux", Arch: "amd64", URL: "https://github.com/angyedz/Waycord/releases/download/v0.1.0/waycord-linux-amd64", Size: 4, SHA256: strings.Repeat("0", 64)}
}
func signedFixture(t *testing.T, p Payload) ([]byte, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	b, err := p.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(Manifest{Payload: p, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(priv, b))})
	if err != nil {
		t.Fatal(err)
	}
	return raw, pub
}
func TestSignedManifest(t *testing.T) {
	p := samplePayload()
	raw, pub := signedFixture(t, p)
	got, err := Verify(raw, pub)
	if err != nil || got.Version != p.Version {
		t.Fatalf("verified payload: %+v %v", got, err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	m.Payload.Size++
	tampered, _ := json.Marshal(m)
	if _, err := Verify(tampered, pub); err == nil {
		t.Fatal("accepted tampered payload")
	}
	another, _, _ := ed25519.GenerateKey(rand.Reader)
	if _, err := Verify(raw, another); err == nil {
		t.Fatal("accepted a different public key")
	}
	if _, err := Verify(append(raw, []byte(` {}`)...), pub); err == nil {
		t.Fatal("accepted extra JSON")
	}
	if _, err := Verify([]byte(`{"payload":{},"signature":"","extra":1}`), pub); err == nil {
		t.Fatal("accepted unknown field")
	}
}
func TestPayloadValidation(t *testing.T) {
	p := samplePayload()
	cases := []struct {
		name   string
		mutate func(*Payload)
	}{
		{"http", func(p *Payload) { p.URL = "http://example.com/app" }},
		{"credentials", func(p *Payload) { p.URL = "https://user:pass@example.com/app" }},
		{"traversal-version", func(p *Payload) { p.Version = "v0.1.0/../../" }},
		{"huge", func(p *Payload) { p.Size = MaxArtifactSize + 1 }},
		{"wrong-platform", func(p *Payload) { p.OS = "osx" }},
		{"bad-channel", func(p *Payload) { p.Channel = "beta" }},
		{"overflow-major", func(p *Payload) { p.Version = "v18446744073709551616.0.0" }},
		{"overflow-serial", func(p *Payload) { p.Channel = "beta"; p.Version = "v1.0.0-beta.18446744073709551616" }},
		{"bad-digest", func(p *Payload) { p.SHA256 = "bad" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			copy := p
			tc.mutate(&copy)
			if err := copy.Validate(); err == nil {
				t.Fatal("invalid release accepted")
			}
		})
	}
	sum := sha256.Sum256([]byte("test"))
	p.SHA256 = hex.EncodeToString(sum[:])
	if err := p.Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestVersionOrder(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"v1.10.0", "v1.9.99", true}, {"v1.2.0", "v1.10.0", false},
		{"v1.0.0", "v1.0.0-beta.2", true}, {"v1.0.0-beta.2", "v1.0.0-nightly.7", true},
		{"v1.0.0-beta.3", "v1.0.0-beta.2", true}, {"v1.0.0-beta.1", "v1.0.0", false},
	} {
		got, err := IsNewer(tc.a, tc.b)
		if err != nil || got != tc.want {
			t.Errorf("newer(%s,%s)=%v, %v", tc.a, tc.b, got, err)
		}
	}
	if _, err := IsNewer("broken", "v1.0.0"); err == nil {
		t.Fatal("invalid version accepted")
	}
	if _, err := IsNewer("v1.0.0-beta.18446744073709551616", "v1.0.0-beta.1"); err == nil {
		t.Fatal("overflowing prerelease serial accepted")
	}
}
