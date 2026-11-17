package ota

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
)

func TestSignedUpdateStagesOnlyVerifiedArtifact(t *testing.T) {
	if (runtime.GOOS != "linux" && runtime.GOOS != "windows") || (runtime.GOARCH != "amd64" && runtime.GOARCH != "arm64") {
		t.Skip("unsupported updater platform")
	}
	artifact := []byte("test binary")
	digest := sha256.Sum256(artifact)
	var manifest []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/manifest" {
			_, _ = w.Write(manifest)
			return
		}
		_, _ = w.Write(artifact)
	}))
	defer server.Close()
	p := samplePayload()
	p.OS, p.Arch, p.URL, p.Size, p.SHA256 = runtime.GOOS, runtime.GOARCH, server.URL+"/binary", int64(len(artifact)), hex.EncodeToString(digest[:])
	manifest, key := signedFixture(t, p)
	store := Store{Root: t.TempDir()}
	wrongKey := make([]byte, len(key))
	if _, err := store.checkAndInstall(context.Background(), server.URL+"/manifest", "stable", wrongKey, server.Client()); err == nil {
		t.Fatal("accepted manifest with an untrusted key")
	}
	state, _, _, err := store.Load()
	if err != nil || state.Pending != "" {
		t.Fatalf("untrusted manifest changed state: %+v %v", state, err)
	}
	if _, err := store.checkAndInstall(context.Background(), server.URL+"/manifest", "stable", key, server.Client()); err != nil {
		t.Fatal(err)
	}
	state, _, _, err = store.Load()
	if err != nil || state.Pending != p.Version {
		t.Fatalf("version not staged: %+v %v", state, err)
	}
	path, err := store.Binary(p.Version)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(artifact) {
		t.Fatalf("staged artifact differs: %q %v", got, err)
	}
}
