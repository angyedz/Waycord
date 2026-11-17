package ota

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type ProgressFunc func(downloaded, total int64)

func secureClient() *http.Client {
	return &http.Client{
		Timeout: 3 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" || len(via) >= 5 {
				return errors.New("unsafe OTA redirect")
			}
			return nil
		},
	}
}

func secureURL(address string) error {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
		return errors.New("updates require an absolute HTTPS URL")
	}
	return nil
}

func FetchManifest(ctx context.Context, address string, key ed25519.PublicKey) (Payload, error) {
	return fetchManifest(ctx, address, key, secureClient())
}

func fetchManifest(ctx context.Context, address string, key ed25519.PublicKey, client *http.Client) (Payload, error) {
	if len(key) != ed25519.PublicKeySize {
		return Payload{}, errors.New("no trusted Ed25519 public key configured")
	}
	if err := secureURL(address); err != nil {
		return Payload{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return Payload{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Payload{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Payload{}, fmt.Errorf("manifest HTTP status: %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, MaxManifestSize+1))
	if err != nil {
		return Payload{}, err
	}
	return Verify(raw, key)
}

// Download verifies size and SHA-256 before returning a temporary artifact.
// The caller owns the returned path and must remove it after installing.
func Download(ctx context.Context, p Payload, tempDir string) (string, error) {
	return download(ctx, p, tempDir, secureClient())
}

func download(ctx context.Context, p Payload, tempDir string, client *http.Client) (string, error) {
	return downloadWithProgress(ctx, p, tempDir, client, nil)
}

func downloadWithProgress(ctx context.Context, p Payload, tempDir string, client *http.Client, progress ProgressFunc) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	if err := secureURL(p.URL); err != nil {
		return "", err
	}
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("artifact HTTP status: %d", resp.StatusCode)
	}
	f, err := os.CreateTemp(tempDir, "artifact-*")
	if err != nil {
		return "", err
	}
	path := f.Name()
	success := false
	defer func() {
		f.Close()
		if !success {
			os.Remove(path)
		}
	}()
	if err = f.Chmod(0600); err != nil {
		return "", err
	}
	h := sha256.New()
	var n int64
	reader := progressReader{reader: io.LimitReader(resp.Body, p.Size+1), total: p.Size, progress: progress}
	n, err = io.Copy(io.MultiWriter(f, h), &reader)
	if err != nil {
		return "", err
	}
	if n != p.Size {
		return "", fmt.Errorf("artifact length mismatch: got %d want %d", n, p.Size)
	}
	if hex.EncodeToString(h.Sum(nil)) != p.SHA256 {
		return "", errors.New("artifact SHA-256 mismatch")
	}
	if err = f.Sync(); err != nil {
		return "", err
	}
	if err = f.Close(); err != nil {
		return "", err
	}
	success = true
	return filepath.Clean(path), nil
}

type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	progress   ProgressFunc
}

func (r *progressReader) Read(buffer []byte) (int, error) {
	n, err := r.reader.Read(buffer)
	r.downloaded += int64(n)
	if n > 0 && r.progress != nil {
		r.progress(min(r.downloaded, r.total), r.total)
	}
	return n, err
}
