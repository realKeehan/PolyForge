package kumi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ══════════════════════════════════════════════════
// Remote content system
//
// The app fetches a single manifest from the website at startup so that
// modpacks, launcher option text, and availability can be updated remotely
// without shipping a new binary. Binary updates are only signalled through
// the manifest's app block; when the running version is older than
// latestVersion the UI shows an update prompt (mandatory when older than
// minSupportedVersion).
//
// Hosted at: https://polyforge.dev/api/manifest.json
// The manifest is cached on disk so the app keeps working offline.
// ══════════════════════════════════════════════════

const (
	remoteManifestURL = "https://polyforge.dev/api/manifest.json"
	packAccessURL     = "https://polyforge.dev/api/pack-access"
)

// RemoteManifest is the combined content + version manifest hosted on the website.
type RemoteManifest struct {
	SchemaVersion   int                    `json:"schemaVersion"`
	Updated         string                 `json:"updated,omitempty"`
	App             RemoteAppInfo          `json:"app"`
	Modpacks        []RemotePack           `json:"modpacks,omitempty"`
	OptionOverrides []RemoteOptionOverride `json:"optionOverrides,omitempty"`
	DisabledOptions []string               `json:"disabledOptions,omitempty"`
}

// RemoteAppInfo describes the latest binary release for update prompts.
type RemoteAppInfo struct {
	LatestVersion       string `json:"latestVersion"`
	MinSupportedVersion string `json:"minSupportedVersion,omitempty"`
	DownloadURL         string `json:"downloadUrl,omitempty"`
	Notes               string `json:"notes,omitempty"`
}

// RemotePack describes a modpack entry the UI should offer.
type RemotePack struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	RequiresPassword bool   `json:"requiresPassword,omitempty"`
	PasswordHash     string `json:"passwordHash,omitempty"`
}

// RemoteOptionOverride patches the title/description of a built-in option.
type RemoteOptionOverride struct {
	ID          string `json:"id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// RemoteContentResult is what the frontend receives at startup.
type RemoteContentResult struct {
	Manifest        *RemoteManifest `json:"manifest,omitempty"`
	FromCache       bool            `json:"fromCache"`
	UpdateAvailable bool            `json:"updateAvailable"`
	Mandatory       bool            `json:"mandatory"`
	CurrentVersion  string          `json:"currentVersion"`
	Error           string          `json:"error,omitempty"`
}

// RemoteContent fetches the manifest, falling back to the on-disk cache when
// the network is unavailable. It never returns an error — the app must always
// be able to start offline with built-in defaults.
func (s *Service) RemoteContent() RemoteContentResult {
	result := RemoteContentResult{CurrentVersion: currentAppVersion()}

	manifest, err := fetchRemoteManifest(s.client)
	if err != nil {
		result.Error = err.Error()
		if cached, cacheErr := readCachedManifest(); cacheErr == nil {
			manifest = cached
			result.FromCache = true
		}
	} else {
		writeCachedManifest(manifest)
	}

	if manifest == nil {
		return result
	}

	result.Manifest = manifest
	current := result.CurrentVersion
	if current != "dev" {
		if compareVersions(manifest.App.LatestVersion, current) > 0 {
			result.UpdateAvailable = true
		}
		if manifest.App.MinSupportedVersion != "" && compareVersions(manifest.App.MinSupportedVersion, current) > 0 {
			result.Mandatory = true
		}
	}

	return result
}

func currentAppVersion() string {
	if AppVersion != "" && AppVersion != "dev" {
		return AppVersion
	}
	return version
}

func fetchRemoteManifest(client *http.Client) (*RemoteManifest, error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequest(http.MethodGet, remoteManifestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())

	fetcher := *client
	fetcher.Timeout = 15 * time.Second

	resp, err := fetcher.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest returned %d", resp.StatusCode)
	}

	var m RemoteManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	return &m, nil
}

// ── Pack password verification ───────────────────

// PackAccessResult is the outcome of a server-side pack password check.
type PackAccessResult struct {
	Granted bool   `json:"granted"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
	// Offline is true when the server could not be reached at all —
	// the caller may fall back to a locally cached hash if it has one.
	Offline bool `json:"offline"`
}

// VerifyPackAccess checks a pack password against the website endpoint.
// The password hash never ships with the app; only the server knows it.
func (s *Service) VerifyPackAccess(packID, password string) PackAccessResult {
	payload, err := json.Marshal(map[string]string{"packId": packID, "password": password})
	if err != nil {
		return PackAccessResult{Error: err.Error()}
	}

	req, err := http.NewRequest(http.MethodPost, packAccessURL, bytes.NewReader(payload))
	if err != nil {
		return PackAccessResult{Error: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent())

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}
	fetcher := *client
	fetcher.Timeout = 15 * time.Second

	resp, err := fetcher.Do(req)
	if err != nil {
		return PackAccessResult{Offline: true, Error: "could not reach the verification server"}
	}
	defer resp.Body.Close()

	var result struct {
		Granted bool   `json:"granted"`
		URL     string `json:"url"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PackAccessResult{Error: "invalid response from verification server"}
	}

	return PackAccessResult{Granted: result.Granted, URL: result.URL, Error: result.Error}
}

// ── Disk cache ───────────────────────────────────

func manifestCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "PolyForge", "content-manifest.json"), nil
}

func readCachedManifest() (*RemoteManifest, error) {
	path, err := manifestCachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m RemoteManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func writeCachedManifest(m *RemoteManifest) {
	path, err := manifestCachePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

// ── Version comparison ───────────────────────────

// compareVersions compares dotted version strings ("5.6.0" style, optional
// leading "v"). Returns >0 if a is newer, <0 if b is newer, 0 if equal.
func compareVersions(a, b string) int {
	pa := versionParts(a)
	pb := versionParts(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var va, vb int
		if i < len(pa) {
			va = pa[i]
		}
		if i < len(pb) {
			vb = pb[i]
		}
		if va != vb {
			return va - vb
		}
	}
	return 0
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return nil
	}
	segments := strings.Split(v, ".")
	parts := make([]int, 0, len(segments))
	for _, seg := range segments {
		// Tolerate suffixes like "1-beta" by reading leading digits only.
		digits := seg
		for j, r := range seg {
			if r < '0' || r > '9' {
				digits = seg[:j]
				break
			}
		}
		n, err := strconv.Atoi(digits)
		if err != nil {
			n = 0
		}
		parts = append(parts, n)
	}
	return parts
}
