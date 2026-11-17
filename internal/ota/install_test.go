package ota

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestStoreManagesOnlyItsCurrentVersionExecutable(t *testing.T) {
	store := Store{Root: t.TempDir()}
	if err := store.Save(State{Current: "v1.0.0"}); err != nil {
		t.Fatal(err)
	}
	managed, err := store.Binary("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(managed), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managed, []byte("current app"), 0700); err != nil {
		t.Fatal(err)
	}
	if !store.ManagesExecutable(managed, "v1.0.0") {
		t.Fatal("current binary in the user-scoped version slot was not recognized")
	}
	if store.ManagesExecutable(managed, "v2.0.0") {
		t.Fatal("binary was treated as managed by a different version")
	}
	packageManaged := filepath.Join(t.TempDir(), BinaryName())
	if err := os.WriteFile(packageManaged, []byte("system-managed app"), 0700); err != nil {
		t.Fatal(err)
	}
	if store.ManagesExecutable(packageManaged, "v1.0.0") {
		t.Fatal("binary outside the user-scoped store was eligible for self-update")
	}
}
