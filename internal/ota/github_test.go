package ota

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
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

func TestVerifiedGitHubUpdateStagesWithMonotonicDownloadProgress(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	artifact := []byte(strings.Repeat("waycord", 1<<15))
	digest := sha256.Sum256(artifact)
	var manifest []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/acme/Waycord/releases":
			_, _ = fmt.Fprintf(w, `[{"tag_name":"v1.0.1","body":"A small update","assets":[{"name":"waycord-%s-%s-stable.json","browser_download_url":"%s/manifest"}]}]`, runtime.GOOS, runtime.GOARCH, "https://"+r.Host)
		case "/manifest":
			_, _ = w.Write(manifest)
		case "/artifact":
			_, _ = w.Write(artifact)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	payload := Payload{
		SchemaVersion: 1, Project: "Waycord", Channel: "stable", Version: "v1.0.1",
		OS: runtime.GOOS, Arch: runtime.GOARCH, URL: server.URL + "/artifact",
		Size: int64(len(artifact)), SHA256: hex.EncodeToString(digest[:]),
	}
	body, err := payload.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err = json.Marshal(Manifest{Payload: payload, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, body))})
	if err != nil {
		t.Fatal(err)
	}
	source := GitHubSource{APIBase: server.URL, Client: server.Client(), Owner: "acme", Repository: "Waycord"}
	update, err := source.Check(context.Background(), "stable", runtime.GOOS, runtime.GOARCH, "v1.0.0", public)
	if err != nil || !update.Available {
		t.Fatalf("signed update check failed: %+v %v", update, err)
	}

	store := Store{Root: t.TempDir()}
	if err := store.Save(State{Current: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	wrongKey := make(ed25519.PublicKey, ed25519.PublicKeySize)
	if err := store.installUpdateWithClient(context.Background(), update, "stable", wrongKey, server.Client(), nil); err == nil {
		t.Fatal("installer accepted an update verified with a different key")
	}
	state, _, _, err := store.Load()
	if err != nil || state.Pending != "" {
		t.Fatalf("untrusted update changed staged state: %+v %v", state, err)
	}

	changed := update
	changed.Payload.Version = "v1.0.2"
	if err := store.installUpdateWithClient(context.Background(), changed, "stable", public, server.Client(), nil); err == nil {
		t.Fatal("installer accepted payload fields changed after signature verification")
	}

	var progress []int64
	err = store.installUpdateWithClient(context.Background(), update, "stable", public, server.Client(), func(downloaded, total int64) {
		if total != int64(len(artifact)) {
			t.Errorf("progress total = %d, want %d", total, len(artifact))
		}
		progress = append(progress, downloaded)
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(progress) < 2 || progress[len(progress)-1] != int64(len(artifact)) {
		t.Fatalf("download progress did not reach total in multiple steps: %v", progress)
	}
	for i, n := range progress {
		if n <= 0 || (i > 0 && n < progress[i-1]) || n > int64(len(artifact)) {
			t.Fatalf("download progress was not monotonic and bounded: %v", progress)
		}
	}
	state, _, _, err = store.Load()
	if err != nil || state.Pending != payload.Version {
		t.Fatalf("verified update was not staged: %+v %v", state, err)
	}
	binary, err := store.Binary(payload.Version)
	if err != nil {
		t.Fatal(err)
	}
	installed, err := os.ReadFile(binary)
	if err != nil || string(installed) != string(artifact) {
		t.Fatalf("staged artifact differs from verified download: %v", err)
	}
}

func serverURL(r *http.Request) string { return "https://" + r.Host }
