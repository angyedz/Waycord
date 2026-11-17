package ota

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubSourceChecksLatestCompatibleSignedRelease(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	asset := []byte("waycord-binary")
	digest := sha256.Sum256(asset)
	payload := Payload{SchemaVersion: 1, Project: "Waycord", Channel: "stable", Version: "v1.2.0", OS: "linux", Arch: "amd64", URL: "https://github.com/angyedz/Waycord/releases/download/v1.2.0/waycord-linux-amd64", Size: int64(len(asset)), SHA256: hex.EncodeToString(digest[:])}
	body, err := payload.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{Payload: payload, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, body))}
	manifestBytes, _ := json.Marshal(manifest)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/angyedz/Waycord/releases":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"tag_name":"v9.0.0","draft":true,"assets":[]},
				{"tag_name":"v1.2.0-beta.3","prerelease":true,"assets":[]},
				{"tag_name":"v1.2.0","html_url":"https://github.com/angyedz/Waycord/releases/tag/v1.2.0","body":"Release notes","assets":[{"name":"waycord-linux-amd64-stable.json","browser_download_url":"` + serverURL(r) + `/manifest"}]}
			]`))
		case "/manifest":
			_, _ = w.Write(manifestBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	source := GitHubSource{APIBase: server.URL, Client: server.Client(), Owner: "angyedz", Repository: "Waycord"}
	got, err := source.Check(context.Background(), "stable", "linux", "amd64", "v1.0.0", public)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Available || got.Payload.Version != "v1.2.0" || got.Notes != "Release notes" {
		t.Fatalf("unexpected update: %+v", got)
	}
}

func TestGitHubSourceFailsClosedWithoutKeyBeforeNetwork(t *testing.T) {
	called := false
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer server.Close()
	source := GitHubSource{APIBase: server.URL, Client: server.Client(), Owner: "angyedz", Repository: "Waycord"}
	_, err := source.Check(context.Background(), "stable", "linux", "amd64", "v1.0.0", nil)
	if err == nil || !strings.Contains(err.Error(), "public key") || called {
		t.Fatalf("check should fail closed before network: err=%v request=%t", err, called)
	}
}

func TestGitHubSourceRejectsWrongChannelAndPlatformManifest(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	payload := Payload{SchemaVersion: 1, Project: "Waycord", Channel: "beta", Version: "v1.2.0-beta.1", OS: "windows", Arch: "amd64", URL: "https://github.com/angyedz/Waycord/releases/download/x/waycord.exe", Size: 5, SHA256: strings.Repeat("a", 64)}
	body, _ := payload.Bytes()
	manifest, _ := json.Marshal(Manifest{Payload: payload, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, body))})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/angyedz/Waycord/releases" {
			_, _ = w.Write([]byte(`[{"tag_name":"v1.2.0-beta.1","prerelease":true,"body":"notes","assets":[{"name":"waycord-linux-amd64-beta.json","browser_download_url":"` + serverURL(r) + `/manifest"}]}]`))
			return
		}
		_, _ = w.Write(manifest)
	}))
	defer server.Close()
	source := GitHubSource{APIBase: server.URL, Client: server.Client(), Owner: "angyedz", Repository: "Waycord"}
	if _, err := source.Check(context.Background(), "stable", "linux", "amd64", "v1.0.0", public); err == nil {
		t.Fatal("beta release was accepted for stable channel")
	}
	if _, err := source.Check(context.Background(), "beta", "linux", "amd64", "v1.0.0-beta.1", public); err == nil {
		t.Fatal("Windows manifest was accepted for Linux")
	}
}

func serverURL(r *http.Request) string { return "https://" + r.Host }
