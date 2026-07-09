package kumi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"
)

func TestVersionFromFilename(t *testing.T) {
	cases := map[string]string{
		"PolyForge-6.1.0-windows-amd64.exe": "6.1.0",
		"PolyForge-10.2.33-linux-amd64":     "10.2.33",
		"no-version-here.exe":               "",
	}
	for in, want := range cases {
		if got := versionFromFilename(in); got != want {
			t.Errorf("versionFromFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDownloadTypeForPlatform(t *testing.T) {
	got, err := downloadTypeForPlatform()
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if err != nil {
			t.Fatalf("unexpected error on %s: %v", runtime.GOOS, err)
		}
		if got == "" {
			t.Fatalf("empty type on %s", runtime.GOOS)
		}
	}
}

// newFakeGateway stands in for the website: the counting redirect gateway, the
// raw release file, and the auto-generated SHA256SUMS.txt.
func newFakeGateway(t *testing.T, filename string, payload []byte, corruptSum bool) *httptest.Server {
	t.Helper()
	sum := sha256.Sum256(payload)
	hexSum := hex.EncodeToString(sum[:])
	if corruptSum {
		hexSum = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	mux := http.NewServeMux()
	// Gateway resolves the "latest" file via a 302 (like download.php).
	mux.HandleFunc("/api/download", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/releases/"+r.URL.Query().Get("type")+"/"+filename, http.StatusFound)
	})
	// SHA256SUMS.txt in coreutils format (hash + two spaces + name).
	mux.HandleFunc("/releases/windows/SHA256SUMS.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  %s\n", hexSum, filename)
	})
	// The raw release binary.
	mux.HandleFunc("/releases/windows/"+filename, func(w http.ResponseWriter, r *http.Request) {
		w.Write(payload)
	})
	return httptest.NewServer(mux)
}

func TestResolveLatestAssetAndDownload(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("test fixtures use the windows release type")
	}
	filename := "PolyForge-9.9.9-windows-amd64.exe"
	payload := []byte("pretend this is a freshly built PolyForge binary")
	srv := newFakeGateway(t, filename, payload, false)
	defer srv.Close()

	prev := downloadGatewayBase
	downloadGatewayBase = srv.URL
	defer func() { downloadGatewayBase = prev }()

	s := &Service{client: srv.Client()}
	asset, err := s.resolveLatestAsset()
	if err != nil {
		t.Fatalf("resolveLatestAsset: %v", err)
	}
	if asset.Filename != filename {
		t.Errorf("filename = %q, want %q", asset.Filename, filename)
	}
	wantSum := sha256.Sum256(payload)
	if asset.SHA256 != hex.EncodeToString(wantSum[:]) {
		t.Errorf("sha256 = %q, want %q", asset.SHA256, hex.EncodeToString(wantSum[:]))
	}

	// DownloadUpdate verifies the checksum and returns the staged temp path.
	tmp, err := DownloadUpdate(asset, nil)
	if err != nil {
		t.Fatalf("DownloadUpdate: %v", err)
	}
	defer os.Remove(tmp)
	got, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatalf("read staged file: %v", err)
	}
	if string(got) != string(payload) {
		t.Errorf("staged bytes mismatch")
	}
}

func TestDownloadUpdateRejectsChecksumMismatch(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("test fixtures use the windows release type")
	}
	filename := "PolyForge-9.9.9-windows-amd64.exe"
	payload := []byte("payload that will not match the published hash")
	srv := newFakeGateway(t, filename, payload, true) // corrupt sum
	defer srv.Close()

	prev := downloadGatewayBase
	downloadGatewayBase = srv.URL
	defer func() { downloadGatewayBase = prev }()

	s := &Service{client: srv.Client()}
	asset, err := s.resolveLatestAsset()
	if err != nil {
		t.Fatalf("resolveLatestAsset: %v", err)
	}
	tmp, err := DownloadUpdate(asset, nil)
	if err == nil {
		os.Remove(tmp)
		t.Fatal("expected a checksum-mismatch error, got nil")
	}
}
