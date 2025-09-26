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

func (s *Service) downloadFile(url string, file *os.File) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

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
