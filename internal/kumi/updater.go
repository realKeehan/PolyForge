package kumi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

// ══════════════════════════════════════════════════
// Self-updater scaffolding
// Separates binary updates from content (modpack) updates.
// ══════════════════════════════════════════════════

// AppVersion is set at build time via -ldflags.
var AppVersion = "dev"

// ── Version manifest ─────────────────────────────
// The app fetches this manifest on startup (or on demand) to check for
// binary updates. Host it on your website or GitHub releases.
//
// Example manifest URL:
//   https://polyforge.dev/update/manifest.json
//   https://github.com/realKeehan/PolyForge/releases/latest/download/manifest.json

// UpdateManifest describes the latest available binary release.
type UpdateManifest struct {
	Version   string        `json:"version"` // semver, e.g. "5.6.0"
	Published time.Time     `json:"published"`
	Assets    []UpdateAsset `json:"assets"`
	Notes     string        `json:"notes"`              // changelog markdown
	Password  string        `json:"password,omitempty"` // optional: required password for private builds
}

// UpdateAsset describes a single downloadable binary for a platform.
type UpdateAsset struct {
	OS       string `json:"os"`       // "windows", "linux", "darwin"
	Arch     string `json:"arch"`     // "amd64", "arm64"
	URL      string `json:"url"`      // direct download URL
	SHA256   string `json:"sha256"`   // hex-encoded SHA256 of the file
	Size     int64  `json:"size"`     // bytes
	Filename string `json:"filename"` // e.g. "PolyForge-5.6.0-windows-amd64.exe"
}

// ── Content manifest ─────────────────────────────
// Modpack lists are fetched independently so new packs appear without
// updating the binary. The app loads a cached copy on startup and
// refreshes in the background.
//
// Example content URL:
//   https://polyforge.dev/content/packs.json

// ContentManifest lists available modpacks and options.
type ContentManifest struct {
	Version int         `json:"version"`
	Updated time.Time   `json:"updated"`
	Packs   []PackEntry `json:"packs"`
}

// PackEntry describes a modpack available for installation.
type PackEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Loader      string   `json:"loader"`       // "fabric", "forge", "neoforge", "quilt"
	GameVersion string   `json:"game_version"` // "1.20.4"
	DownloadURL string   `json:"download_url"`
	SHA256      string   `json:"sha256"`
	Tags        []string `json:"tags"`
	// Password-protected packs require authentication before download.
	RequiresAuth bool `json:"requires_auth,omitempty"`
}

// ── Update check ─────────────────────────────────

// UpdateCheckResult describes the outcome of an update check.
type UpdateCheckResult struct {
	Available    bool
	LatestVer    string
	CurrentVer   string
	Asset        *UpdateAsset // nil if no matching asset for this platform
	ReleaseNotes string
}

// CheckForUpdate fetches the manifest and compares versions.
// manifestURL should point to your hosted manifest.json.
func CheckForUpdate(manifestURL string) (*UpdateCheckResult, error) {
	manifest, err := fetchManifest(manifestURL)
	if err != nil {
		return nil, fmt.Errorf("update check failed: %w", err)
	}

	result := &UpdateCheckResult{
		LatestVer:    manifest.Version,
		CurrentVer:   AppVersion,
		ReleaseNotes: manifest.Notes,
	}

	if manifest.Version == AppVersion || AppVersion == "dev" {
		return result, nil
	}

	// Find matching asset for current OS/arch
	for i := range manifest.Assets {
		a := &manifest.Assets[i]
		if a.OS == runtime.GOOS && a.Arch == runtime.GOARCH {
			result.Available = true
			result.Asset = a
			break
		}
	}

	return result, nil
}

func fetchManifest(url string) (*UpdateManifest, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest returned %d", resp.StatusCode)
	}

	var m UpdateManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	return &m, nil
}

// ── Download + verify ────────────────────────────

// DownloadUpdate downloads the asset to a temp file and verifies its SHA256.
// Returns the path to the verified temp file.
func DownloadUpdate(asset *UpdateAsset, progressFn func(downloaded, total int64)) (string, error) {
	if asset == nil {
		return "", fmt.Errorf("no asset provided")
	}

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(asset.URL)
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "polyforge-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := writer.Write(buf[:n]); writeErr != nil {
				os.Remove(tmpFile.Name())
				return "", writeErr
			}
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, asset.Size)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			os.Remove(tmpFile.Name())
			return "", readErr
		}
	}

	// Verify SHA256
	got := hex.EncodeToString(hasher.Sum(nil))
	if asset.SHA256 != "" && got != asset.SHA256 {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("SHA256 mismatch: expected %s, got %s", asset.SHA256, got)
	}

	return tmpFile.Name(), nil
}

// ── Apply update (scaffolding) ───────────────────
// On Windows, the running exe cannot be replaced directly.
// The recommended flow:
//   1. Download to temp
//   2. Rename current exe to .old
//   3. Move temp to current exe path
//   4. Relaunch
//   5. On next launch, clean up .old
//
// This is scaffolded below - the actual relaunch logic will be
// platform-specific and integrated when the self-updater ships.

// ApplyUpdate replaces the current binary with the downloaded update.
// This is a basic scaffolding - production use needs platform-specific
// handling for locked files (Windows), permissions (Linux/macOS), etc.
func ApplyUpdate(downloadedPath string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable: %w", err)
	}

	backupPath := currentExe + ".old"

	// Remove any previous backup
	_ = os.Remove(backupPath)

	// Rename current to backup
	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("cannot backup current binary: %w", err)
	}

	// Move downloaded to current
	if err := os.Rename(downloadedPath, currentExe); err != nil {
		// Rollback
		_ = os.Rename(backupPath, currentExe)
		return fmt.Errorf("cannot install update: %w", err)
	}

	return nil
}

// CleanupOldBinary removes the .old backup from a previous update.
// Call this early in the app lifecycle.
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	_ = os.Remove(exe + ".old")
}

// ── Content manifest fetch ───────────────────────

// FetchContentManifest downloads and parses the modpack content manifest.
func FetchContentManifest(url string) (*ContentManifest, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("content manifest returned %d", resp.StatusCode)
	}

	var m ContentManifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid content manifest: %w", err)
	}
	return &m, nil
}
