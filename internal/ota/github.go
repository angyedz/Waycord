package ota

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxGitHubReleases = 2 << 20

// Update describes a verified release. Notes and the release page are display
// metadata; only Payload, verified from the signed manifest, is trusted for install.
type Update struct {
	Available  bool
	Payload    Payload
	Notes      string
	ReleaseURL string
}

// GitHubSource resolves a channel's latest release to its platform manifest.
// Owner and Repository are configuration, not values from the API response.
type GitHubSource struct {
	APIBase    string
	Owner      string
	Repository string
	Client     *http.Client
}

type releaseRecord struct {
	TagName    string `json:"tag_name"`
	Body       string `json:"body"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Check requires a configured trusted key, checks the GitHub release list over
// HTTPS, then verifies the release's signed manifest before returning it.
func (s GitHubSource) Check(ctx context.Context, channel, goos, goarch, installed string, key ed25519.PublicKey) (Update, error) {
	if len(key) != ed25519.PublicKeySize {
		return Update{}, errors.New("no trusted Ed25519 public key configured")
	}
	if !IsChannel(channel) {
		return Update{}, errors.New("invalid OTA channel")
	}
	if (goos != "linux" && goos != "windows") || (goarch != "amd64" && goarch != "arm64") {
		return Update{}, errors.New("unsupported platform")
	}
	if s.Owner == "" || s.Repository == "" || strings.ContainsAny(s.Owner+s.Repository, "/?#") {
		return Update{}, errors.New("invalid GitHub repository configuration")
	}
	base := strings.TrimRight(s.APIBase, "/")
	if base == "" {
		base = "https://api.github.com"
	}
	if err := secureURL(base); err != nil {
		return Update{}, err
	}
	client := s.Client
	if client == nil {
		client = secureGitHubClient(base)
	}
	endpoint := base + "/repos/" + url.PathEscape(s.Owner) + "/" + url.PathEscape(s.Repository) + "/releases?per_page=100"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Update{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := client.Do(req)
	if err != nil {
		return Update{}, fmt.Errorf("fetch GitHub releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Update{}, fmt.Errorf("GitHub releases HTTP status: %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxGitHubReleases+1))
	if err != nil {
		return Update{}, err
	}
	if len(raw) == 0 || len(raw) > maxGitHubReleases {
		return Update{}, errors.New("GitHub releases response exceeds limit")
	}
	var releases []releaseRecord
	if err := json.Unmarshal(raw, &releases); err != nil {
		return Update{}, fmt.Errorf("invalid GitHub releases response: %w", err)
	}
	var latest *releaseRecord
	for i := range releases {
		release := &releases[i]
		if release.Draft || !versionForChannel(release.TagName, channel) || release.Prerelease != (channel != "stable") {
			continue
		}
		if latest == nil {
			latest = release
			continue
		}
		newer, err := IsNewer(release.TagName, latest.TagName)
		if err != nil {
			return Update{}, err
		}
		if newer {
			latest = release
		}
	}
	if latest == nil {
		return Update{}, fmt.Errorf("no %s release found", channel)
	}
	newer, err := IsNewer(latest.TagName, installed)
	if err != nil {
		return Update{}, err
	}
	if !newer {
		return Update{Available: false}, nil
	}
	manifestName := fmt.Sprintf("waycord-%s-%s-%s.json", goos, goarch, channel)
	manifestURL := ""
	for _, asset := range latest.Assets {
		if asset.Name == manifestName {
			manifestURL = asset.BrowserDownloadURL
			break
		}
	}
	if manifestURL == "" {
		return Update{}, fmt.Errorf("release %s has no %s manifest", latest.TagName, manifestName)
	}
	if err := secureURL(manifestURL); err != nil {
		return Update{}, fmt.Errorf("unsafe release manifest URL: %w", err)
	}
	manifestAddress, _ := url.Parse(manifestURL)
	if s.APIBase == "" && (manifestAddress.Hostname() != "github.com" || !strings.HasPrefix(manifestAddress.Path, "/"+s.Owner+"/"+s.Repository+"/releases/download/")) {
		return Update{}, errors.New("release manifest is not hosted in the configured GitHub repository")
	}
	manifestReq, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return Update{}, err
	}
	manifestResp, err := client.Do(manifestReq)
	if err != nil {
		return Update{}, fmt.Errorf("fetch release manifest: %w", err)
	}
	defer manifestResp.Body.Close()
	if manifestResp.StatusCode != http.StatusOK {
		return Update{}, fmt.Errorf("release manifest HTTP status: %d", manifestResp.StatusCode)
	}
	manifest, err := io.ReadAll(io.LimitReader(manifestResp.Body, MaxManifestSize+1))
	if err != nil {
		return Update{}, err
	}
	payload, err := Verify(manifest, key)
	if err != nil {
		return Update{}, err
	}
	if payload.Channel != channel || payload.Version != latest.TagName || payload.OS != goos || payload.Arch != goarch {
		return Update{}, errors.New("signed release manifest does not match release, channel or platform")
	}
	notes := latest.Body
	if len(notes) > 16<<10 {
		notes = strings.ToValidUTF8(notes[:16<<10], "�") + "\n…"
	}
	releaseURL := "https://github.com/" + url.PathEscape(s.Owner) + "/" + url.PathEscape(s.Repository) + "/releases/tag/" + url.PathEscape(latest.TagName)
	return Update{Available: true, Payload: payload, Notes: notes, ReleaseURL: releaseURL}, nil
}

func secureGitHubClient(apiBase string) *http.Client {
	base, _ := url.Parse(apiBase)
	allowed := map[string]bool{
		"api.github.com": true, "github.com": true,
		"release-assets.githubusercontent.com": true, "objects.githubusercontent.com": true,
	}
	allowed[strings.ToLower(base.Hostname())] = true
	return &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" || req.URL.User != nil || len(via) >= 5 || !allowed[strings.ToLower(req.URL.Hostname())] {
				return errors.New("unsafe GitHub release redirect")
			}
			return nil
		},
	}
}

func versionForChannel(version, channel string) bool {
	if !validVersion.MatchString(version) {
		return false
	}
	switch channel {
	case "stable":
		return !strings.Contains(version, "-")
	case "beta":
		return strings.Contains(version, "-beta.")
	case "nightly":
		return strings.Contains(version, "-nightly.")
	default:
		return false
	}
}
