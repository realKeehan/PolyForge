package kumi

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// downloadGatewayBase is the website root serving the counting download
// gateway (/api/download) and the raw release folders (/releases/<type>/).
// It is a var (not const) only so tests can point it at a local server.
var downloadGatewayBase = "https://polyforge.dev"

// ══════════════════════════════════════════════════
// Self-updater scaffolding
// Separates binary updates from content (modpack) updates.
// ══════════════════════════════════════════════════

// AppVersion is set at build time via -ldflags.
var AppVersion = "dev"

// UpdateAsset describes a single downloadable binary for a platform.
type UpdateAsset struct {
	OS       string `json:"os"`       // "windows", "linux", "darwin"
	Arch     string `json:"arch"`     // "amd64", "arm64"
	URL      string `json:"url"`      // direct download URL
	SHA256   string `json:"sha256"`   // hex-encoded SHA256 of the file
	Size     int64  `json:"size"`     // bytes
	Filename string `json:"filename"` // e.g. "PolyForge-5.6.0-windows-amd64.exe"
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

	// Stage the download in the SAME directory as the running executable so
	// the final ApplyUpdate rename is an intra-volume move (a cross-device
	// rename from the system temp dir would fail with EXDEV on Windows).
	stageDir := os.TempDir()
	if exe, exeErr := os.Executable(); exeErr == nil {
		stageDir = filepath.Dir(exe)
	}
	tmpFile, err := os.CreateTemp(stageDir, "polyforge-update-*.tmp")
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
	if asset.SHA256 != "" && !strings.EqualFold(got, asset.SHA256) {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("SHA256 mismatch: expected %s, got %s", asset.SHA256, got)
	}

	return tmpFile.Name(), nil
}

// ── Self-update against the website download gateway ─────
//
// The website already exposes everything a self-updater needs:
//   • /api/download?type=<platform> 302-redirects to the newest build and
//     counts the download, so we learn the exact filename from the redirect.
//   • /releases/<type>/SHA256SUMS.txt (auto-refreshed on every upload) gives
//     the checksum to verify against.
// So no extra per-asset manifest is required — PerformSelfUpdate resolves the
// latest asset, downloads + verifies it, and swaps the running binary.

// UpdateSelfResult is returned to the UI after a self-update attempt.
type UpdateSelfResult struct {
	Applied bool   `json:"applied"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PerformSelfUpdate downloads the latest build for this platform, verifies it
// against the published checksum, and replaces the running binary in place. It
// does not relaunch — the caller (app layer) relaunches and quits.
func (s *Service) PerformSelfUpdate() UpdateSelfResult {
	asset, err := s.resolveLatestAsset()
	if err != nil {
		return UpdateSelfResult{Error: err.Error()}
	}
	tmp, err := DownloadUpdate(asset, nil)
	if err != nil {
		return UpdateSelfResult{Error: err.Error()}
	}
	if err := ApplyUpdate(tmp); err != nil {
		_ = os.Remove(tmp)
		return UpdateSelfResult{Error: err.Error()}
	}
	return UpdateSelfResult{Applied: true, Version: versionFromFilename(asset.Filename)}
}

// resolveLatestAsset asks the download gateway which file is newest for this
// platform (via the 302 Location) and pairs it with its published checksum.
func (s *Service) resolveLatestAsset() (*UpdateAsset, error) {
	dlType, err := downloadTypeForPlatform()
	if err != nil {
		return nil, err
	}

	// Resolve the redirect without pulling the body, to learn the filename.
	resolver := s.updateHTTPClient(20 * time.Second)
	resolver.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, err := http.NewRequest(http.MethodGet, downloadGatewayBase+"/api/download?type="+dlType, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())
	resp, err := resolver.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach the update server: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("no %s build is available to update to (server returned %d)", dlType, resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if loc == "" {
		return nil, errors.New("update server did not return a download location")
	}

	fileURL := loc
	if strings.HasPrefix(loc, "/") {
		fileURL = downloadGatewayBase + loc
	}
	filename := path.Base(loc)
	if decoded, decErr := url.PathUnescape(filename); decErr == nil {
		filename = decoded
	}

	sum, err := s.fetchChecksum(downloadGatewayBase+"/releases/"+dlType+"/SHA256SUMS.txt", filename)
	if err != nil {
		return nil, err
	}
	return &UpdateAsset{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		URL:      fileURL,
		SHA256:   sum,
		Filename: filename,
	}, nil
}

// fetchChecksum reads a coreutils-style SHA256SUMS.txt and returns the hash
// listed for filename (lines are "<hex>␠␠<filename>"; names may contain spaces).
func (s *Service) fetchChecksum(sumsURL, filename string) (string, error) {
	client := s.updateHTTPClient(20 * time.Second)
	req, err := http.NewRequest(http.MethodGet, sumsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent())
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not fetch checksums: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums are unavailable yet (server returned %d)", resp.StatusCode)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		parts := strings.SplitN(strings.TrimRight(scanner.Text(), "\r"), "  ", 2)
		if len(parts) == 2 && parts[1] == filename {
			return strings.ToLower(strings.TrimSpace(parts[0])), nil
		}
	}
	return "", fmt.Errorf("no checksum has been published for %s yet", filename)
}

// updateHTTPClient clones the service's TLS-configured client (falling back to
// the default) with the given timeout, so update traffic reuses the same
// transport as the rest of the app.
func (s *Service) updateHTTPClient(timeout time.Duration) *http.Client {
	base := s.client
	if base == nil {
		base = http.DefaultClient
	}
	clone := *base
	clone.Timeout = timeout
	clone.CheckRedirect = nil
	return &clone
}

// downloadTypeForPlatform maps the running OS/arch to a website release-type
// folder that holds a single swappable executable.
func downloadTypeForPlatform() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "arm64" {
			return "windows-arm64", nil
		}
		return "windows", nil
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "linux-arm64", nil
		}
		return "linux", nil
	case "darwin":
		return "macos", nil
	}
	return "", fmt.Errorf("self-update is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}

var versionInFilename = regexp.MustCompile(`-(\d+\.\d+\.\d+)-`)

// versionFromFilename best-effort extracts "6.1.0" from
// "PolyForge-6.1.0-windows-amd64.exe"; empty string if it can't.
func versionFromFilename(name string) string {
	if m := versionInFilename.FindStringSubmatch(name); len(m) == 2 {
		return m[1]
	}
	return ""
}

// RelaunchSelf starts a fresh copy of the (now-updated) executable as an
// independent process. The caller quits the current instance afterwards.
func RelaunchSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	return cmd.Start()
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

