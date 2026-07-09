package kumi

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) downloadAndExtract(url, destination, explicitName string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("no download URL configured")
	}

	tmp, err := os.CreateTemp("", "polyforge-*.zip")
	if err != nil {
		return err
	}

	tempPath := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tempPath)
	}()

	if err := s.downloadFile(url, tmp); err != nil {
		return err
	}

	if err := tmp.Sync(); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if explicitName != "" {
		namedPath := filepath.Join(filepath.Dir(tempPath), explicitName)
		if err := os.Rename(tempPath, namedPath); err == nil {
			tempPath = namedPath
		}
	}

	return extractZip(tempPath, destination)
}

// resolveDownloadURL turns a root-relative pack URL ("/packs/foo.polypack")
// into an absolute one against the website base (downloadGatewayBase), leaving
// already-absolute URLs — the launcher CDN links, or a pack URL the server
// already absolutized — untouched. This is the safety net for hosted-pack URLs
// that reach the app relative: the app fetches them directly and would 404
// (or fail with no host) without a base to resolve against.
func resolveDownloadURL(raw string) string {
	u := strings.TrimSpace(raw)
	if strings.HasPrefix(u, "/") {
		return strings.TrimRight(downloadGatewayBase, "/") + u
	}
	return u
}

func (s *Service) downloadFile(url string, file *os.File) error {
	req, err := http.NewRequest(http.MethodGet, resolveDownloadURL(url), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent())

	if s.ctx != nil {
		req = req.WithContext(s.ctx)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("download failed with status %s", resp.Status)
	}

	if _, err := io.Copy(file, resp.Body); err != nil {
		return err
	}
	return nil
}

// downloadToTemp streams a file to a temp path, reporting byte progress live
// through the service emitter as it goes. Unlike downloadAndExtract it does not
// unzip — hosted packs are .polypack (slime-obfuscated) containers, so the
// caller installs the downloaded file via installLocalPack, which handles the
// unwrap. The caller owns the returned path and must remove it.
func (s *Service) downloadToTemp(rawURL, label string) (string, error) {
	url := resolveDownloadURL(rawURL)
	if strings.TrimSpace(url) == "" {
		return "", errors.New("no download URL configured")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent())
	if s.ctx != nil {
		req = req.WithContext(s.ctx)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("download failed with status %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "polyforge-pack-*.polypack")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	counter := &countingWriter{total: resp.ContentLength, label: label, emit: s.emitProgress}
	if _, err := io.Copy(io.MultiWriter(tmp, counter), resp.Body); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}

// countingWriter tallies bytes written and emits a progress event whenever the
// integer percentage advances (so at most ~100 events per download). When the
// content length is unknown it reports downloaded bytes with an indeterminate
// bar via emitProgress's negative-percent contract.
type countingWriter struct {
	done, total int64
	label       string
	lastPercent int
	emit        func(percent int, label string)
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.done += int64(n)
	if w.emit == nil {
		return n, nil
	}
	if w.total > 0 {
		pct := int(w.done * 100 / w.total)
		if pct != w.lastPercent {
			w.lastPercent = pct
			w.emit(pct, fmt.Sprintf("%s  %s / %s", w.label, humanBytes(w.done), humanBytes(w.total)))
		}
		return n, nil
	}
	// Unknown length: report bytes so far with an indeterminate bar.
	w.emit(-1, fmt.Sprintf("%s  %s", w.label, humanBytes(w.done)))
	return n, nil
}

// humanBytes formats a byte count as a compact human-readable string.
func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for size := n / unit; size >= unit; size /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func extractZip(zipPath, destination string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		targetPath := filepath.Join(destination, file.Name)
		if !strings.HasPrefix(targetPath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("zip entry %s is outside the destination", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := ensureDir(targetPath); err != nil {
				return err
			}
			continue
		}

		if err := ensureDir(filepath.Dir(targetPath)); err != nil {
			return err
		}

		src, err := file.Open()
		if err != nil {
			return err
		}

		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.Mode())
		if err != nil {
			src.Close()
			return err
		}

		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			dst.Close()
			return err
		}
		src.Close()
		dst.Close()
	}
	return nil
}
