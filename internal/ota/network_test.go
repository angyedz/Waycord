package ota

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDownloadHashAndSize(t *testing.T) {
	payload := []byte("a small executable fixture")
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(payload) }))
	defer server.Close()
	// The local test TLS authority is intentionally not in the system trust
	// store; verify that the production HTTP client rejects this connection.
	p := samplePayload()
	p.URL = server.URL
	p.Size = int64(len(payload))
	h := sha256.Sum256(payload)
	p.SHA256 = hex.EncodeToString(h[:])
	if _, err := Download(context.Background(), p, t.TempDir()); err == nil {
		t.Fatal("accepted untrusted TLS certificate")
	}
	// Same security requirements without networking, test the manifest bounds.
	p.URL = "https://example.com/a"
	p.Size = 0
	if _, err := Download(context.Background(), p, t.TempDir()); err == nil {
		t.Fatal("accepted empty artifact")
	}
}

func TestDownloadVerifiesAndCleansInterruptedArtifacts(t *testing.T) {
	data := []byte("trusted executable fixture")
	digest := sha256.Sum256(data)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			_, _ = w.Write(data)
		case "/short":
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			_, _ = w.Write(data[:4])
		case "/wrong":
			_, _ = w.Write([]byte(strings.Repeat("x", len(data))))
		default:
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
		}
	}))
	defer server.Close()
	p := samplePayload()
	p.Size, p.SHA256 = int64(len(data)), hex.EncodeToString(digest[:])
	dir := t.TempDir()
	p.URL = server.URL + "/ok"
	path, err := download(context.Background(), p, dir, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(data) {
		t.Fatalf("downloaded bytes: %q, error: %v", got, err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	for _, endpoint := range []string{"/short", "/wrong", "/unavailable"} {
		p.URL = server.URL + endpoint
		if _, err := download(context.Background(), p, dir, server.Client()); err == nil {
			t.Errorf("%s: invalid download accepted", endpoint)
		}
		files, err := filepath.Glob(filepath.Join(dir, "artifact-*"))
		if err != nil || len(files) != 0 {
			t.Fatalf("%s: incomplete artifacts remain: %v (%v)", endpoint, files, err)
		}
	}
}
func TestNoHTTPOrOversizedManifest(t *testing.T) {
	p := samplePayload()
	p.URL = "http://localhost/binary"
	if err := p.Validate(); err == nil {
		t.Fatal("HTTP artifact accepted")
	}
	if _, err := Verify([]byte(strings.Repeat("x", MaxManifestSize+1)), make([]byte, 32)); err == nil {
		t.Fatal("oversized manifest accepted")
	}
	_, err := os.Stat("does-not-exist-test-file")
	if err == nil {
		t.Fatal("test environment unexpected file")
	}
}
